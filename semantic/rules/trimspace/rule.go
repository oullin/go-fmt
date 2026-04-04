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

	var violations []rules.Violation

	changed := false

	ast.Inspect(file, func(node ast.Node) bool {
		binary, ok := node.(*ast.BinaryExpr)

		if !ok || (binary.Op != token.EQL && binary.Op != token.NEQ) {
			return true
		}

		operand := nonEmptyStringOperand(binary)

		if operand == nil || isTrimSpaceCall(*operand, importInfo) {
			return true
		}

		call, ok := trimSpaceCall(*operand, importInfo)

		if !ok {
			return true
		}

		*operand = call

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
