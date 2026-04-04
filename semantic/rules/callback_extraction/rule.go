package callback_extraction

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/oullin/go-fmt/semantic/rules"
)

type Rule struct{}

func New() Rule {
	return Rule{}
}

func (Rule) Name() string {
	return "callback_extraction"
}

func (Rule) Apply(path string, src []byte) ([]rules.Violation, []byte, error) {
	return analyse(path, src)
}

func analyse(filename string, src []byte) ([]rules.Violation, []byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)

	if err != nil {
		return nil, nil, err
	}

	var violations []rules.Violation
	changed := false

	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.FuncDecl:
			if typed.Body != nil && processBlock(typed.Body, newNameRegistry(typed.Type, typed.Body), fset, filename, &violations) {
				changed = true
			}
		case *ast.FuncLit:
			if processBlock(typed.Body, newNameRegistry(typed.Type, typed.Body), fset, filename, &violations) {
				changed = true
			}
		}

		return true
	})

	if !changed {
		return nil, src, nil
	}

	var out bytes.Buffer

	if err := format.Node(&out, fset, file); err != nil {
		return nil, nil, fmt.Errorf("format callback extraction: %w", err)
	}

	return violations, out.Bytes(), nil
}

func processBlock(block *ast.BlockStmt, registry *nameRegistry, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	if block == nil {
		return false
	}

	changed := false
	rewritten := make([]ast.Stmt, 0, len(block.List))

	for _, stmt := range block.List {
		decls, stmtChanged := extractCallbacks(stmt, registry, fset, filename, violations)

		if stmtChanged {
			changed = true
			rewritten = append(rewritten, decls...)
		}

		if processNestedStmtLists(stmt, registry, fset, filename, violations) {
			changed = true
		}

		rewritten = append(rewritten, stmt)
	}

	if changed {
		block.List = rewritten
	}

	return changed
}

func extractCallbacks(stmt ast.Stmt, registry *nameRegistry, fset *token.FileSet, filename string, violations *[]rules.Violation) ([]ast.Stmt, bool) {
	var decls []ast.Stmt
	changed := false

	astutil.Apply(stmt, func(cursor *astutil.Cursor) bool {
		node := cursor.Node()

		switch typed := node.(type) {
		case *ast.BlockStmt, *ast.CaseClause, *ast.CommClause:
			return node == stmt
		case *ast.FuncLit:
			return false
		case *ast.KeyValueExpr:
			if !isStructFieldFuncLiteral(typed) {
				return true
			}

			funcLit := typed.Value.(*ast.FuncLit)
			name := registry.unique(fieldVariableName(typed.Key))
			decls = append(decls, &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent(name)},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{funcLit},
			})
			typed.Value = ast.NewIdent(name)
			*violations = append(*violations, rules.Violation{
				Rule:    "callback_extraction",
				File:    filename,
				Line:    fset.Position(funcLit.Pos()).Line,
				Message: fmt.Sprintf("extracted inline callback into %s", name),
			})
			changed = true

			return false
		default:
			return true
		}
	}, nil)

	return decls, changed
}

func processNestedStmtLists(stmt ast.Stmt, registry *nameRegistry, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	switch typed := stmt.(type) {
	case *ast.BlockStmt:
		return processBlock(typed, registry, fset, filename, violations)
	case *ast.IfStmt:
		changed := processBlock(typed.Body, registry, fset, filename, violations)

		switch elseStmt := typed.Else.(type) {
		case *ast.BlockStmt:
			return processBlock(elseStmt, registry, fset, filename, violations) || changed
		case *ast.IfStmt:
			return processNestedStmtLists(elseStmt, registry, fset, filename, violations) || changed
		default:
			return changed
		}
	case *ast.ForStmt:
		return processBlock(typed.Body, registry, fset, filename, violations)
	case *ast.RangeStmt:
		return processBlock(typed.Body, registry, fset, filename, violations)
	case *ast.SwitchStmt:
		return processCaseClauses(typed.Body.List, registry, fset, filename, violations)
	case *ast.TypeSwitchStmt:
		return processCaseClauses(typed.Body.List, registry, fset, filename, violations)
	case *ast.SelectStmt:
		return processCommClauses(typed.Body.List, registry, fset, filename, violations)
	case *ast.LabeledStmt:
		return processNestedStmtLists(typed.Stmt, registry, fset, filename, violations)
	default:
		return false
	}
}

func processCaseClauses(clauses []ast.Stmt, registry *nameRegistry, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	changed := false

	for _, stmt := range clauses {
		clause, ok := stmt.(*ast.CaseClause)

		if !ok {
			continue
		}

		if processStmtSlice(&clause.Body, registry, fset, filename, violations) {
			changed = true
		}
	}

	return changed
}

func processCommClauses(clauses []ast.Stmt, registry *nameRegistry, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	changed := false

	for _, stmt := range clauses {
		clause, ok := stmt.(*ast.CommClause)

		if !ok {
			continue
		}

		if processStmtSlice(&clause.Body, registry, fset, filename, violations) {
			changed = true
		}
	}

	return changed
}

func processStmtSlice(list *[]ast.Stmt, registry *nameRegistry, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	block := &ast.BlockStmt{List: *list}

	if !processBlock(block, registry, fset, filename, violations) {
		return false
	}

	*list = block.List

	return true
}

func isStructFieldFuncLiteral(kv *ast.KeyValueExpr) bool {
	_, ok := kv.Value.(*ast.FuncLit)

	if !ok {
		return false
	}

	switch kv.Key.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		return true
	default:
		return false
	}
}

func fieldVariableName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return lowerFirst(typed.Name) + "Fn"
	case *ast.SelectorExpr:
		return lowerFirst(typed.Sel.Name) + "Fn"
	default:
		return "callbackFn"
	}
}

func lowerFirst(value string) string {
	if value == "" {
		return "callback"
	}

	runes := []rune(value)
	runes[0] = []rune(strings.ToLower(string(runes[0])))[0]

	return string(runes)
}

type nameRegistry struct {
	used map[string]struct{}
}

func newNameRegistry(funcType *ast.FuncType, body *ast.BlockStmt) *nameRegistry {
	used := map[string]struct{}{}

	recordFieldList(used, funcType.Params)
	recordFieldList(used, funcType.Results)

	ast.Inspect(body, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)

		if ok && ident.Name != "_" {
			used[ident.Name] = struct{}{}
		}

		return true
	})

	return &nameRegistry{used: used}
}

func recordFieldList(used map[string]struct{}, fields *ast.FieldList) {
	if fields == nil {
		return
	}

	for _, field := range fields.List {
		for _, name := range field.Names {
			if name.Name == "_" {
				continue
			}

			used[name.Name] = struct{}{}
		}
	}
}

func (n *nameRegistry) unique(base string) string {
	if base == "" {
		base = "callbackFn"
	}

	if _, ok := n.used[base]; !ok {
		n.used[base] = struct{}{}

		return base
	}

	for i := 1; ; i++ {
		candidate := base + strconv.Itoa(i)

		if _, ok := n.used[candidate]; ok {
			continue
		}

		n.used[candidate] = struct{}{}

		return candidate
	}
}
