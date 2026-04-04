package trimspace

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/oullin/go-fmt/semantic/rules"
)

type Rule struct{}

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

	alias, importSpec := stringsImport(file)
	var violations []rules.Violation
	changed := false

	ast.Inspect(file, func(node ast.Node) bool {
		binary, ok := node.(*ast.BinaryExpr)

		if !ok || (binary.Op != token.EQL && binary.Op != token.NEQ) {
			return true
		}

		operand := nonEmptyStringOperand(binary)

		if operand == nil || isTrimSpaceCall(*operand, alias) {
			return true
		}

		*operand = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(alias),
				Sel: ast.NewIdent("TrimSpace"),
			},
			Args: []ast.Expr{*operand},
		}

		violations = append(violations, rules.Violation{
			Rule:    "trimspace",
			File:    filename,
			Line:    fset.Position(binary.Pos()).Line,
			Message: "wrapped empty-string comparison with strings.TrimSpace",
		})
		changed = true

		return true
	})

	if !changed {
		return nil, src, nil
	}

	if importSpec == nil {
		astutil.AddImport(fset, file, "strings")
	} else if importSpec.Name != nil && (importSpec.Name.Name == "_" || importSpec.Name.Name == ".") {
		importSpec.Name = nil
	}

	var out bytes.Buffer

	if err := format.Node(&out, fset, file); err != nil {
		return nil, nil, fmt.Errorf("format trimspace rule: %w", err)
	}

	return violations, out.Bytes(), nil
}

func stringsImport(file *ast.File) (string, *ast.ImportSpec) {
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)

		if err != nil || path != "strings" {
			continue
		}

		if spec.Name == nil {
			return "strings", spec
		}

		switch spec.Name.Name {
		case "_", ".":
			return "strings", spec
		default:
			return spec.Name.Name, spec
		}
	}

	return "strings", nil
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

func isTrimSpaceCall(expr ast.Expr, alias string) bool {
	call, ok := expr.(*ast.CallExpr)

	if !ok {
		return false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)

	if !ok || selector.Sel.Name != "TrimSpace" {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)

	return ok && ident.Name == alias
}
