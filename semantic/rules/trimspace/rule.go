package trimspace

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/oullin/go-fmt/semantic/rules"
)

const (
	stringsImportMissing stringsImportKind = iota
	stringsImportDefault
	stringsImportAlias
	stringsImportDot
	stringsImportBlank
)

type Rule struct{}

type stringsImportKind int

type stringsImportInfo struct {
	kind stringsImportKind
	name string
}

type shadowState struct {
	fileScopeShadows bool
	scopes           []bool
}

func New() Rule {
	return Rule{}
}

func (Rule) Name() string {
	return "trimspace"
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

	importInfo := stringsImport(file)
	state := shadowState{
		fileScopeShadows: fileScopeShadowsStrings(file, importInfo),
	}

	var violations []rules.Violation

	changed := false

	for _, decl := range file.Decls {
		if rewriteDecl(decl, importInfo, &state, fset, filename, &violations) {
			changed = true
		}
	}

	if !changed {
		return nil, src, nil
	}

	if importInfo.kind == stringsImportMissing {
		astutil.AddImport(fset, file, "strings")
	}

	var out bytes.Buffer

	if err := format.Node(&out, fset, file); err != nil {
		return nil, nil, fmt.Errorf("format trimspace rule: %w", err)
	}

	return violations, out.Bytes(), nil
}

func stringsImport(file *ast.File) stringsImportInfo {
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)

		if err != nil || path != "strings" {
			continue
		}

		if spec.Name == nil {
			return stringsImportInfo{
				kind: stringsImportDefault,
				name: "strings",
			}
		}

		switch spec.Name.Name {
		case ".":
			return stringsImportInfo{
				kind: stringsImportDot,
				name: ".",
			}
		case "_":
			return stringsImportInfo{
				kind: stringsImportBlank,
				name: "_",
			}
		default:
			return stringsImportInfo{
				kind: stringsImportAlias,
				name: spec.Name.Name,
			}
		}
	}

	return stringsImportInfo{
		kind: stringsImportMissing,
		name: "strings",
	}
}

func nonEmptyStringOperand(binary *ast.BinaryExpr) *ast.Expr {
	switch {
	case isEmptyStringLiteral(binary.X) && !isEmptyStringLiteral(binary.Y):
		return &binary.Y
	case isEmptyStringLiteral(binary.Y) && !isEmptyStringLiteral(binary.X):
		return &binary.X
	default:
		return nil
	}
}

func isEmptyStringLiteral(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)

	return ok && lit.Kind == token.STRING && lit.Value == `""`
}

func rewriteDecl(decl ast.Decl, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		changed := rewriteFieldList(typed.Recv, info, state, fset, filename, violations)
		changed = rewriteFieldList(typed.Type.Params, info, state, fset, filename, violations) || changed
		changed = rewriteFieldList(typed.Type.Results, info, state, fset, filename, violations) || changed

		if typed.Body == nil {
			return changed
		}

		state.push(fieldListShadowsStrings(typed.Recv) || fieldListShadowsStrings(typed.Type.Params) || fieldListShadowsStrings(typed.Type.Results))

		defer state.pop()

		return rewriteBlock(typed.Body, info, state, fset, filename, violations) || changed
	case *ast.GenDecl:
		return rewriteGenDecl(typed, info, state, fset, filename, violations, false)
	default:
		return false
	}
}

func rewriteBlock(block *ast.BlockStmt, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	if block == nil {
		return false
	}

	state.push(false)

	defer state.pop()

	changed := false

	for _, stmt := range block.List {
		if rewriteStmt(stmt, info, state, fset, filename, violations) {
			changed = true
		}
	}

	return changed
}

func rewriteStmt(stmt ast.Stmt, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	switch typed := stmt.(type) {
	case *ast.AssignStmt:
		changed := rewriteExprList(typed.Lhs, info, state, fset, filename, violations)
		changed = rewriteExprList(typed.Rhs, info, state, fset, filename, violations) || changed

		if typed.Tok == token.DEFINE {
			declareAssignNames(state, typed.Lhs)
		}

		return changed
	case *ast.BlockStmt:
		return rewriteBlock(typed, info, state, fset, filename, violations)
	case *ast.BranchStmt:
		return false
	case *ast.CaseClause:
		return rewriteCaseClause(typed, info, state, fset, filename, violations)
	case *ast.CommClause:
		return rewriteCommClause(typed, info, state, fset, filename, violations)
	case *ast.DeclStmt:
		gen, ok := typed.Decl.(*ast.GenDecl)

		if !ok {
			return false
		}

		return rewriteGenDecl(gen, info, state, fset, filename, violations, true)
	case *ast.DeferStmt:
		return rewriteExpr(typed.Call, info, state, fset, filename, violations)
	case *ast.EmptyStmt:
		return false
	case *ast.ExprStmt:
		return rewriteExpr(typed.X, info, state, fset, filename, violations)
	case *ast.ForStmt:
		state.push(false)

		defer state.pop()

		changed := rewriteStmt(typed.Init, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Cond, info, state, fset, filename, violations) || changed
		changed = rewriteBlock(typed.Body, info, state, fset, filename, violations) || changed
		changed = rewriteStmt(typed.Post, info, state, fset, filename, violations) || changed

		return changed
	case *ast.GoStmt:
		return rewriteExpr(typed.Call, info, state, fset, filename, violations)
	case *ast.IfStmt:
		state.push(false)

		defer state.pop()

		changed := rewriteStmt(typed.Init, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Cond, info, state, fset, filename, violations) || changed
		changed = rewriteBlock(typed.Body, info, state, fset, filename, violations) || changed
		changed = rewriteStmt(typed.Else, info, state, fset, filename, violations) || changed

		return changed
	case *ast.IncDecStmt:
		return rewriteExpr(typed.X, info, state, fset, filename, violations)
	case *ast.LabeledStmt:
		return rewriteStmt(typed.Stmt, info, state, fset, filename, violations)
	case *ast.RangeStmt:
		changed := rewriteExpr(typed.Key, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Value, info, state, fset, filename, violations) || changed
		changed = rewriteExpr(typed.X, info, state, fset, filename, violations) || changed

		if typed.Tok == token.DEFINE {
			state.push(false)

			declareRangeNames(state, typed.Key, typed.Value)

			changed = rewriteBlock(typed.Body, info, state, fset, filename, violations) || changed

			state.pop()

			return changed
		}

		return rewriteBlock(typed.Body, info, state, fset, filename, violations) || changed
	case *ast.ReturnStmt:
		return rewriteExprList(typed.Results, info, state, fset, filename, violations)
	case *ast.SelectStmt:
		if typed.Body == nil {
			return false
		}

		changed := false

		for _, stmt := range typed.Body.List {
			if rewriteStmt(stmt, info, state, fset, filename, violations) {
				changed = true
			}
		}

		return changed
	case *ast.SendStmt:
		changed := rewriteExpr(typed.Chan, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Value, info, state, fset, filename, violations) || changed

		return changed
	case *ast.SwitchStmt:
		state.push(false)

		defer state.pop()

		changed := rewriteStmt(typed.Init, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Tag, info, state, fset, filename, violations) || changed

		if typed.Body == nil {
			return changed
		}

		for _, stmt := range typed.Body.List {
			if rewriteStmt(stmt, info, state, fset, filename, violations) {
				changed = true
			}
		}

		return changed
	case *ast.TypeSwitchStmt:
		state.push(false)

		defer state.pop()

		changed := rewriteStmt(typed.Init, info, state, fset, filename, violations)
		changed = rewriteTypeSwitchAssign(typed.Assign, info, state, fset, filename, violations) || changed

		if typed.Body == nil {
			return changed
		}

		for _, stmt := range typed.Body.List {
			if rewriteStmt(stmt, info, state, fset, filename, violations) {
				changed = true
			}
		}

		return changed
	default:
		return false
	}
}

func rewriteCaseClause(clause *ast.CaseClause, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	state.push(false)

	defer state.pop()

	changed := rewriteExprList(clause.List, info, state, fset, filename, violations)

	for _, stmt := range clause.Body {
		if rewriteStmt(stmt, info, state, fset, filename, violations) {
			changed = true
		}
	}

	return changed
}

func rewriteCommClause(clause *ast.CommClause, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	state.push(false)

	defer state.pop()

	changed := rewriteSelectComm(clause.Comm, info, state, fset, filename, violations)

	for _, stmt := range clause.Body {
		if rewriteStmt(stmt, info, state, fset, filename, violations) {
			changed = true
		}
	}

	return changed
}

func rewriteSelectComm(stmt ast.Stmt, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	assign, ok := stmt.(*ast.AssignStmt)

	if !ok {
		return rewriteStmt(stmt, info, state, fset, filename, violations)
	}

	changed := rewriteExprList(assign.Lhs, info, state, fset, filename, violations)
	changed = rewriteExprList(assign.Rhs, info, state, fset, filename, violations) || changed

	if assign.Tok == token.DEFINE {
		declareAssignNames(state, assign.Lhs)
	}

	return changed
}

func rewriteTypeSwitchAssign(stmt ast.Stmt, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	switch typed := stmt.(type) {
	case nil:
		return false
	case *ast.AssignStmt:
		changed := rewriteExprList(typed.Lhs, info, state, fset, filename, violations)
		changed = rewriteExprList(typed.Rhs, info, state, fset, filename, violations) || changed

		if typed.Tok == token.DEFINE {
			declareAssignNames(state, typed.Lhs)
		}

		return changed
	case *ast.ExprStmt:
		return rewriteExpr(typed.X, info, state, fset, filename, violations)
	default:
		return rewriteStmt(stmt, info, state, fset, filename, violations)
	}
}

func rewriteGenDecl(decl *ast.GenDecl, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation, declareLocalNames bool) bool {
	changed := false

	for _, spec := range decl.Specs {
		switch typed := spec.(type) {
		case *ast.ValueSpec:
			changed = rewriteExprList(typed.Values, info, state, fset, filename, violations) || changed

			if declareLocalNames {
				declareValueSpecNames(state, typed)
			}
		case *ast.TypeSpec:
			if typed.TypeParams != nil {
				changed = rewriteFieldList(typed.TypeParams, info, state, fset, filename, violations) || changed
			}

			if declareLocalNames {
				state.declare(typed.Name.Name)
			}
		}
	}

	return changed
}

func rewriteFieldList(fields *ast.FieldList, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	if fields == nil {
		return false
	}

	changed := false

	for _, field := range fields.List {
		changed = rewriteExpr(field.Type, info, state, fset, filename, violations) || changed

		for _, name := range field.Names {
			changed = rewriteExpr(name, info, state, fset, filename, violations) || changed
		}
	}

	return changed
}

func rewriteExprList(list []ast.Expr, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	changed := false

	for _, expr := range list {
		if rewriteExpr(expr, info, state, fset, filename, violations) {
			changed = true
		}
	}

	return changed
}

func rewriteExpr(expr ast.Expr, info stringsImportInfo, state *shadowState, fset *token.FileSet, filename string, violations *[]rules.Violation) bool {
	if expr == nil {
		return false
	}

	switch typed := expr.(type) {
	case *ast.ArrayType:
		changed := rewriteExpr(typed.Len, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Elt, info, state, fset, filename, violations) || changed

		return changed
	case *ast.BinaryExpr:
		changed := rewriteExpr(typed.X, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Y, info, state, fset, filename, violations) || changed

		if typed.Op != token.EQL && typed.Op != token.NEQ {
			return changed
		}

		operand := nonEmptyStringOperand(typed)

		if operand == nil || isTrimSpaceCall(*operand, info) || !canRewriteTrimSpace(info, state) {
			return changed
		}

		call, ok := trimSpaceCall(*operand, info)

		if !ok {
			return changed
		}

		*operand = call

		*violations = append(*violations, rules.Violation{
			Rule:    "trimspace",
			File:    filename,
			Line:    fset.Position(typed.Pos()).Line,
			Message: "wrapped empty-string comparison with strings.TrimSpace",
		})

		return true
	case *ast.CallExpr:
		changed := rewriteExpr(typed.Fun, info, state, fset, filename, violations)
		changed = rewriteExprList(typed.Args, info, state, fset, filename, violations) || changed

		return changed
	case *ast.ChanType:
		return rewriteExpr(typed.Value, info, state, fset, filename, violations)
	case *ast.CompositeLit:
		changed := rewriteExpr(typed.Type, info, state, fset, filename, violations)
		changed = rewriteExprList(typed.Elts, info, state, fset, filename, violations) || changed

		return changed
	case *ast.Ellipsis:
		return rewriteExpr(typed.Elt, info, state, fset, filename, violations)
	case *ast.FuncLit:
		changed := rewriteFieldList(typed.Type.Params, info, state, fset, filename, violations)
		changed = rewriteFieldList(typed.Type.Results, info, state, fset, filename, violations) || changed

		state.push(fieldListShadowsStrings(typed.Type.Params) || fieldListShadowsStrings(typed.Type.Results))

		defer state.pop()

		return rewriteBlock(typed.Body, info, state, fset, filename, violations) || changed
	case *ast.FuncType:
		changed := rewriteFieldList(typed.Params, info, state, fset, filename, violations)
		changed = rewriteFieldList(typed.Results, info, state, fset, filename, violations) || changed

		return changed
	case *ast.IndexExpr:
		changed := rewriteExpr(typed.X, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Index, info, state, fset, filename, violations) || changed

		return changed
	case *ast.IndexListExpr:
		changed := rewriteExpr(typed.X, info, state, fset, filename, violations)
		changed = rewriteExprList(typed.Indices, info, state, fset, filename, violations) || changed

		return changed
	case *ast.InterfaceType:
		return rewriteFieldList(typed.Methods, info, state, fset, filename, violations)
	case *ast.KeyValueExpr:
		changed := rewriteExpr(typed.Key, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Value, info, state, fset, filename, violations) || changed

		return changed
	case *ast.MapType:
		changed := rewriteExpr(typed.Key, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Value, info, state, fset, filename, violations) || changed

		return changed
	case *ast.ParenExpr:
		return rewriteExpr(typed.X, info, state, fset, filename, violations)
	case *ast.SelectorExpr:
		return rewriteExpr(typed.X, info, state, fset, filename, violations)
	case *ast.SliceExpr:
		changed := rewriteExpr(typed.X, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Low, info, state, fset, filename, violations) || changed
		changed = rewriteExpr(typed.High, info, state, fset, filename, violations) || changed
		changed = rewriteExpr(typed.Max, info, state, fset, filename, violations) || changed

		return changed
	case *ast.StarExpr:
		return rewriteExpr(typed.X, info, state, fset, filename, violations)
	case *ast.StructType:
		return rewriteFieldList(typed.Fields, info, state, fset, filename, violations)
	case *ast.TypeAssertExpr:
		changed := rewriteExpr(typed.X, info, state, fset, filename, violations)
		changed = rewriteExpr(typed.Type, info, state, fset, filename, violations) || changed

		return changed
	case *ast.UnaryExpr:
		return rewriteExpr(typed.X, info, state, fset, filename, violations)
	default:
		return false
	}
}

func canRewriteTrimSpace(info stringsImportInfo, state *shadowState) bool {
	switch info.kind {
	case stringsImportBlank:
		return false
	case stringsImportDefault, stringsImportMissing:
		return !state.shadowsStrings()
	default:
		return true
	}
}

func fileScopeShadowsStrings(file *ast.File, info stringsImportInfo) bool {
	if info.kind != stringsImportDefault && info.kind != stringsImportMissing {
		return false
	}

	for _, spec := range file.Imports {
		name := importBindingName(spec)

		if name == "strings" && !(info.kind == stringsImportDefault && isStringsImportSpec(spec)) {
			return true
		}
	}

	for _, decl := range file.Decls {
		switch typed := decl.(type) {
		case *ast.FuncDecl:
			if typed.Name != nil && typed.Name.Name == "strings" {
				return true
			}
		case *ast.GenDecl:
			for _, spec := range typed.Specs {
				switch inner := spec.(type) {
				case *ast.TypeSpec:
					if inner.Name.Name == "strings" {
						return true
					}
				case *ast.ValueSpec:
					for _, name := range inner.Names {
						if name.Name == "strings" {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func importBindingName(spec *ast.ImportSpec) string {
	if spec == nil {
		return ""
	}

	if spec.Name != nil {
		return spec.Name.Name
	}

	pathValue, err := strconv.Unquote(spec.Path.Value)

	if err != nil {
		return ""
	}

	return path.Base(pathValue)
}

func isStringsImportSpec(spec *ast.ImportSpec) bool {
	if spec == nil {
		return false
	}

	pathValue, err := strconv.Unquote(spec.Path.Value)

	return err == nil && pathValue == "strings"
}

func fieldListShadowsStrings(fields *ast.FieldList) bool {
	if fields == nil {
		return false
	}

	for _, field := range fields.List {
		for _, name := range field.Names {
			if name.Name == "strings" {
				return true
			}
		}
	}

	return false
}

func declareValueSpecNames(state *shadowState, spec *ast.ValueSpec) {
	for _, name := range spec.Names {
		state.declare(name.Name)
	}
}

func declareAssignNames(state *shadowState, exprs []ast.Expr) {
	for _, expr := range exprs {
		ident, ok := expr.(*ast.Ident)

		if !ok {
			continue
		}

		state.declare(ident.Name)
	}
}

func declareRangeNames(state *shadowState, key ast.Expr, value ast.Expr) {
	declareAssignNames(state, []ast.Expr{key, value})
}

func (s *shadowState) push(shadowsStrings bool) {
	s.scopes = append(s.scopes, shadowsStrings)
}

func (s *shadowState) pop() {
	if len(s.scopes) == 0 {
		return
	}

	s.scopes = s.scopes[:len(s.scopes)-1]
}

func (s *shadowState) declare(name string) {
	if name != "strings" || len(s.scopes) == 0 {
		return
	}

	s.scopes[len(s.scopes)-1] = true
}

func (s *shadowState) shadowsStrings() bool {
	if s.fileScopeShadows {
		return true
	}

	for i := len(s.scopes) - 1; i >= 0; i-- {
		if s.scopes[i] {
			return true
		}
	}

	return false
}

func trimSpaceCall(arg ast.Expr, info stringsImportInfo) (*ast.CallExpr, bool) {
	switch info.kind {
	case stringsImportBlank:
		return nil, false
	case stringsImportDot:
		return &ast.CallExpr{
			Fun:  ast.NewIdent("TrimSpace"),
			Args: []ast.Expr{arg},
		}, true
	default:
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(info.name),
				Sel: ast.NewIdent("TrimSpace"),
			},
			Args: []ast.Expr{arg},
		}, true
	}
}

func isTrimSpaceCall(expr ast.Expr, info stringsImportInfo) bool {
	call, ok := expr.(*ast.CallExpr)

	if !ok {
		return false
	}

	if info.kind == stringsImportDot {
		ident, ok := call.Fun.(*ast.Ident)

		return ok && ident.Name == "TrimSpace"
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)

	if !ok || selector.Sel.Name != "TrimSpace" {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)

	return ok && ident.Name == info.name
}
