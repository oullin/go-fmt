package spacing

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"strings"

	"github.com/oullin/go-fmt/packages/formatter/rules"
)

// Rule enforces blank-line and type-order spacing rules.
type Rule struct{}

type importAliases map[string]string

var stdlibSpacingImports = map[string]string{
	"sort":         "sort",
	"slices":       "slices",
	"math/rand":    "rand",
	"math/rand/v2": "rand",
}

// New returns the built-in spacing rule.
func New() Rule {
	return Rule{}
}

// Name returns the rule identifier used in reports.
func (Rule) Name() string {
	return "spacing"
}

func (r Rule) Apply(path string, src []byte) ([]rules.Violation, []byte, error) {
	return analyse(path, src)
}

func analyse(filename string, src []byte) ([]rules.Violation, []byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)

	if err != nil {
		return nil, nil, err
	}

	tokenFile := fset.File(file.Pos())

	if tokenFile == nil {
		return nil, nil, fmt.Errorf("missing token file for %s", filename)
	}

	lineStarts := buildLineStarts(src)
	insertions := map[int]struct{}{}
	aliases := buildImportAliases(file)

	var violations []rules.Violation

	inspectStmtLists(file, func(list []ast.Stmt) {
		for i := 0; i < len(list)-1; i++ {
			current := list[i]
			next := list[i+1]
			endLine := fset.Position(current.End()).Line
			nextLine := fset.Position(next.Pos()).Line

			if currentLine, ok := setupSpacingLine(list, i, current, next, aliases, fset); ok {
				violations = append(violations, rules.Violation{
					Rule:    "spacing",
					File:    filename,
					Line:    currentLine,
					Message: "missing blank line before selector call setup",
				})

				offset := lineStartOffset(lineStarts, currentLine)
				insertions[offset] = struct{}{}
			}

			if endLine == nextLine {
				continue
			}

			if message, ok := statementGapRule(current, next, aliases, fset); ok {
				if nextLine < endLine+2 {
					violations = append(violations, rules.Violation{
						Rule:    "spacing",
						File:    filename,
						Line:    nextLine,
						Message: message,
					})

					offset := lineStartOffset(lineStarts, nextLine)
					insertions[offset] = struct{}{}
				}
			}
		}
	})

	for i := 0; i < len(file.Decls)-1; i++ {
		current := file.Decls[i]
		next := file.Decls[i+1]

		if !requiresTypeDeclSpacing(current, next) {
			continue
		}

		endLine := fset.Position(current.End()).Line
		nextLine := fset.Position(next.Pos()).Line

		if nextLine >= endLine+2 {
			continue
		}

		violations = append(violations, rules.Violation{
			Rule:    "spacing",
			File:    filename,
			Line:    nextLine,
			Message: "missing blank line around type definition",
		})

		offset := lineStartOffset(lineStarts, nextLine)
		insertions[offset] = struct{}{}
	}

	violations = append(violations, typeOrderViolations(file, fset, filename)...)

	formatted := src

	if len(insertions) > 0 {
		formatted = applyInsertions(formatted, insertions)
	}

	if len(violations) > 0 {
		reordered, changed, err := reorderTypeDecls(filename, formatted)

		if err != nil {
			return nil, nil, err
		}

		if changed {
			formatted = reordered
		}
	}

	return violations, formatted, nil
}

func setupSpacingLine(list []ast.Stmt, index int, current ast.Stmt, next ast.Stmt, aliases importAliases, fset *token.FileSet) (int, bool) {
	if index == 0 {
		return 0, false
	}

	receiverName, ok := selectorReceiverName(next, aliases)

	if !ok {
		return 0, false
	}

	assignedName, ok := assignedIdentifier(current)

	if !ok || assignedName != receiverName {
		return 0, false
	}

	currentLine := fset.Position(current.Pos()).Line
	prevEndLine := fset.Position(list[index-1].End()).Line

	if currentLine >= prevEndLine+2 {
		return 0, false
	}

	return currentLine, true
}

func inspectStmtLists(file *ast.File, visit func([]ast.Stmt)) {
	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.BlockStmt:
			visit(typed.List)
		case *ast.CaseClause:
			visit(typed.Body)
		case *ast.CommClause:
			visit(typed.Body)
		}

		return true
	})
}

func statementGapRule(current ast.Stmt, next ast.Stmt, aliases importAliases, fset *token.FileSet) (string, bool) {
	if label, ok := requiresLeadingBlankLine(next, aliases); ok {
		return fmt.Sprintf("missing blank line before %s", label), true
	}

	if label, ok := requiresTrailingBlankLine(current, next, aliases, fset); ok {
		return fmt.Sprintf("missing blank line after %s", label), true
	}

	return "", false
}

func requiresTrailingBlankLine(current ast.Stmt, next ast.Stmt, aliases importAliases, fset *token.FileSet) (string, bool) {
	if isAnonymousFuncAssignmentStmt(current, fset) {
		return "anonymous function assignment", true
	}

	if label, ok := stdlibSpacedCallLabel(current, aliases); ok {
		return label, true
	}

	switch current.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt, *ast.DeferStmt, *ast.BranchStmt:
		return statementLabel(current, aliases), true
	case *ast.DeclStmt:
		if isTypeDeclStmt(current) {
			return statementLabel(current, aliases), true
		}

		if isVarDeclStmt(current) {
			if !isShortAssignStmt(next) && !isVarDeclStmt(next) {
				return statementLabel(current, aliases), true
			}
		}
	}

	return "", false
}

func requiresLeadingBlankLine(stmt ast.Stmt, aliases importAliases) (string, bool) {
	if label, ok := routeRegistryCallLabel(stmt); ok {
		return label, true
	}

	if label, ok := stdlibSpacedCallLabel(stmt, aliases); ok {
		return label, true
	}

	switch stmt.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt, *ast.DeferStmt, *ast.ReturnStmt, *ast.BranchStmt:
		return statementLabel(stmt, aliases), true
	case *ast.DeclStmt:
		if isTypeDeclStmt(stmt) || isVarDeclStmt(stmt) {
			return statementLabel(stmt, aliases), true
		}
	}

	return "", false
}

func routeRegistryCallLabel(stmt ast.Stmt) (string, bool) {
	selector, ok := selectorCall(stmt)

	if !ok {
		return "", false
	}

	receiver, ok := selector.X.(*ast.Ident)

	if !ok || receiver.Name != "routes" {
		return "", false
	}

	switch selector.Sel.Name {
	case "Add", "Group":
		return "routes call", true
	default:
		return "", false
	}
}

func statementLabel(stmt ast.Stmt, aliases importAliases) string {
	if label, ok := stdlibSpacedCallLabel(stmt, aliases); ok {
		return label
	}

	switch typed := stmt.(type) {
	case *ast.IfStmt:
		return "if statement"
	case *ast.ForStmt:
		return "for loop"
	case *ast.RangeStmt:
		return "range loop"
	case *ast.SwitchStmt:
		return "switch statement"
	case *ast.TypeSwitchStmt:
		return "type switch"
	case *ast.SelectStmt:
		return "select statement"
	case *ast.DeferStmt:
		return "defer statement"
	case *ast.ReturnStmt:
		return "return statement"
	case *ast.BranchStmt:
		return fmt.Sprintf("%s statement", typed.Tok)
	case *ast.DeclStmt:
		if isTypeDeclStmt(stmt) {
			return "type definition"
		}

		if isVarDeclStmt(stmt) {
			return "var declaration"
		}
	}

	return "statement"
}

func buildImportAliases(file *ast.File) importAliases {
	aliases := make(importAliases)

	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)

		if err != nil {
			continue
		}

		defaultName, ok := stdlibSpacingImports[path]

		if !ok {
			continue
		}

		name := defaultName

		if spec.Name != nil {
			name = spec.Name.Name
		}

		if name == "_" || name == "." || strings.TrimSpace(name) == "" {
			continue
		}

		aliases[name] = path
	}

	return aliases
}

func stdlibSpacedCallLabel(stmt ast.Stmt, aliases importAliases) (string, bool) {
	selector, ok := selectorCall(stmt)

	if !ok {
		return "", false
	}

	pkgIdent, ok := selector.X.(*ast.Ident)

	if !ok {
		return "", false
	}

	return selectorLabel(pkgIdent.Name, selector.Sel.Name, aliases)
}

func selectorCall(stmt ast.Stmt) (*ast.SelectorExpr, bool) {
	exprStmt, ok := stmt.(*ast.ExprStmt)

	if !ok {
		return nil, false
	}

	call, ok := exprStmt.X.(*ast.CallExpr)

	if !ok {
		return nil, false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)

	if !ok {
		return nil, false
	}

	return selector, true
}

func selectorReceiverName(stmt ast.Stmt, aliases importAliases) (string, bool) {
	selector, ok := selectorCall(stmt)

	if !ok {
		return "", false
	}

	pkgIdent, ok := selector.X.(*ast.Ident)

	if !ok {
		return "", false
	}

	if _, ok := selectorLabel(pkgIdent.Name, selector.Sel.Name, aliases); !ok {
		return "", false
	}

	return pkgIdent.Name, true
}

func selectorLabel(receiverName, selectorName string, aliases importAliases) (string, bool) {
	switch aliases[receiverName] {
	case "sort":
		return "sort call", true
	case "slices":
		return "sort call", strings.HasPrefix(selectorName, "Sort")
	case "math/rand", "math/rand/v2":
		return "rand call", true
	}

	switch receiverName {
	case "sort":
		return "sort call", true
	case "slices":
		return "sort call", strings.HasPrefix(selectorName, "Sort")
	case "rand":
		return "rand call", true
	default:
		return "", false
	}
}

func assignedIdentifier(stmt ast.Stmt) (string, bool) {
	switch typed := stmt.(type) {
	case *ast.AssignStmt:
		if len(typed.Lhs) != 1 {
			return "", false
		}

		ident, ok := typed.Lhs[0].(*ast.Ident)

		if !ok {
			return "", false
		}

		return ident.Name, true
	case *ast.DeclStmt:
		genDecl, ok := typed.Decl.(*ast.GenDecl)

		if !ok || genDecl.Tok != token.VAR || len(genDecl.Specs) != 1 {
			return "", false
		}

		valueSpec, ok := genDecl.Specs[0].(*ast.ValueSpec)

		if !ok || len(valueSpec.Names) != 1 {
			return "", false
		}

		return valueSpec.Names[0].Name, true
	default:
		return "", false
	}
}

func isTypeDeclStmt(stmt ast.Stmt) bool {
	return isTokenDeclStmt(stmt, token.TYPE)
}

func isVarDeclStmt(stmt ast.Stmt) bool {
	return isTokenDeclStmt(stmt, token.VAR)
}

func isTokenDeclStmt(stmt ast.Stmt, tok token.Token) bool {
	declStmt, ok := stmt.(*ast.DeclStmt)

	if !ok {
		return false
	}

	genDecl, ok := declStmt.Decl.(*ast.GenDecl)

	return ok && genDecl.Tok == tok
}

func isShortAssignStmt(stmt ast.Stmt) bool {
	assign, ok := stmt.(*ast.AssignStmt)

	return ok && assign.Tok == token.DEFINE
}

func isAnonymousFuncAssignmentStmt(stmt ast.Stmt, fset *token.FileSet) bool {
	switch typed := stmt.(type) {
	case *ast.AssignStmt:
		return hasAnonymousFuncInitializerExpr(typed.Rhs, fset)
	case *ast.DeclStmt:
		genDecl, ok := typed.Decl.(*ast.GenDecl)

		if !ok || genDecl.Tok != token.VAR {
			return false
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)

			if !ok {
				continue
			}

			if hasAnonymousFuncInitializerExpr(valueSpec.Values, fset) {
				return true
			}
		}
	}

	return false
}

func hasAnonymousFuncInitializerExpr(exprs []ast.Expr, fset *token.FileSet) bool {
	for _, expr := range exprs {
		if !isMultiLineAnonymousFuncInitializerExpr(expr, fset) {
			continue
		}

		return true
	}

	return false
}

func isMultiLineAnonymousFuncInitializerExpr(expr ast.Expr, fset *token.FileSet) bool {
	switch typed := expr.(type) {
	case *ast.FuncLit:
		return spansMultipleLines(typed, fset)
	case *ast.CallExpr:
		if _, ok := typed.Fun.(*ast.FuncLit); ok {
			return spansMultipleLines(typed, fset)
		}
	}

	return false
}

func spansMultipleLines(node ast.Node, fset *token.FileSet) bool {
	if node == nil || fset == nil {
		return false
	}

	return fset.Position(node.Pos()).Line != fset.Position(node.End()).Line
}

func requiresTypeDeclSpacing(current ast.Decl, next ast.Decl) bool {
	return isTypeDecl(current) || isTypeDecl(next)
}

func isTypeDecl(decl ast.Decl) bool {
	genDecl, ok := decl.(*ast.GenDecl)

	return ok && genDecl.Tok == token.TYPE
}

func isImportDecl(decl ast.Decl) bool {
	genDecl, ok := decl.(*ast.GenDecl)

	return ok && genDecl.Tok == token.IMPORT
}

func typeOrderViolations(file *ast.File, fset *token.FileSet, filename string) []rules.Violation {
	var violations []rules.Violation
	seenNonType := false

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)

		if ok && genDecl.Tok == token.IMPORT {
			continue
		}

		if isTypeDecl(decl) {
			if seenNonType {
				violations = append(violations, rules.Violation{
					Rule:    "spacing",
					File:    filename,
					Line:    fset.Position(decl.Pos()).Line,
					Message: "type definitions must appear at the beginning of the file",
				})
			}

			continue
		}

		seenNonType = true
	}

	return violations
}

func reorderTypeDecls(filename string, src []byte) ([]byte, bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)

	if err != nil {
		return nil, false, err
	}

	if !hasOutOfOrderTypeDecls(file) {
		return src, false, nil
	}

	importsEnd := 0

	for importsEnd < len(file.Decls) && isImportDecl(file.Decls[importsEnd]) {
		importsEnd++
	}

	reordered := make([]ast.Decl, 0, len(file.Decls))
	reordered = append(reordered, file.Decls[:importsEnd]...)

	for _, decl := range file.Decls[importsEnd:] {
		if isTypeDecl(decl) {
			reordered = append(reordered, decl)
		}
	}

	for _, decl := range file.Decls[importsEnd:] {
		if !isTypeDecl(decl) {
			reordered = append(reordered, decl)
		}
	}

	file.Decls = reordered

	var out bytes.Buffer

	if err := format.Node(&out, fset, file); err != nil {
		return nil, false, err
	}

	return out.Bytes(), true, nil
}

func hasOutOfOrderTypeDecls(file *ast.File) bool {
	seenNonType := false

	for _, decl := range file.Decls {
		if isImportDecl(decl) {
			continue
		}

		if isTypeDecl(decl) {
			if seenNonType {
				return true
			}

			continue
		}

		seenNonType = true
	}

	return false
}

func buildLineStarts(src []byte) []int {
	starts := []int{0, 0}

	for i, b := range src {
		if b == '\n' {
			starts = append(starts, i+1)
		}
	}

	return starts
}

func lineStartOffset(starts []int, line int) int {
	if line < 1 {
		return 0
	}

	if line >= len(starts) {
		return starts[len(starts)-1]
	}

	return starts[line]
}

func applyInsertions(src []byte, insertions map[int]struct{}) []byte {
	offsets := make([]int, 0, len(insertions))

	for offset := range insertions {
		offsets = append(offsets, offset)
	}

	slices.Sort(offsets)

	var out bytes.Buffer
	last := 0

	for _, offset := range offsets {
		out.Write(src[last:offset])
		out.WriteByte('\n')
		last = offset
	}

	out.Write(src[last:])

	return out.Bytes()
}
