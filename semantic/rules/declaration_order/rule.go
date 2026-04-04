package declaration_order

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"github.com/oullin/go-fmt/semantic/rules"
)

type Rule struct{}

func New() Rule {
	return Rule{}
}

func (Rule) Name() string {
	return "declaration_order"
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

	violations := orderViolations(file, fset, filename)

	if len(violations) == 0 {
		return nil, src, nil
	}

	reordered := reorderDecls(file)
	file.Decls = reordered

	var out bytes.Buffer

	if err := format.Node(&out, fset, file); err != nil {
		return nil, nil, fmt.Errorf("format reordered declarations: %w", err)
	}

	return violations, collapseEmbedSpacing(out.Bytes()), nil
}

func orderViolations(file *ast.File, fset *token.FileSet, filename string) []rules.Violation {
	var violations []rules.Violation

	highestCategory := declCategoryImport

	for _, decl := range file.Decls {
		category := declCategory(decl)

		if category == declCategoryImport {
			continue
		}

		if category < highestCategory {
			violations = append(violations, rules.Violation{
				Rule:    "declaration_order",
				File:    filename,
				Line:    fset.Position(decl.Pos()).Line,
				Message: violationMessage(category),
			})
		}

		if category > highestCategory {
			highestCategory = category
		}
	}

	return violations
}

func reorderDecls(file *ast.File) []ast.Decl {
	imports := make([]ast.Decl, 0, len(file.Decls))
	vars := make([]ast.Decl, 0, len(file.Decls))
	types := make([]ast.Decl, 0, len(file.Decls))
	others := make([]ast.Decl, 0, len(file.Decls))

	for _, decl := range file.Decls {
		switch declCategory(decl) {
		case declCategoryImport:
			imports = append(imports, decl)
		case declCategoryVar:
			vars = append(vars, decl)
		case declCategoryType:
			types = append(types, decl)
		default:
			others = append(others, decl)
		}
	}

	reordered := make([]ast.Decl, 0, len(file.Decls))
	reordered = append(reordered, imports...)
	reordered = append(reordered, vars...)
	reordered = append(reordered, types...)
	reordered = append(reordered, others...)

	return reordered
}

const (
	declCategoryImport = iota
	declCategoryVar
	declCategoryType
	declCategoryOther
)

func declCategory(decl ast.Decl) int {
	genDecl, ok := decl.(*ast.GenDecl)

	if !ok {
		return declCategoryOther
	}

	switch genDecl.Tok {
	case token.IMPORT:
		return declCategoryImport
	case token.VAR:
		return declCategoryVar
	case token.TYPE:
		return declCategoryType
	default:
		return declCategoryOther
	}
}

func violationMessage(category int) string {
	switch category {
	case declCategoryVar:
		return "file-scope var declarations must appear before type declarations and other declarations"
	case declCategoryType:
		return "file-scope type declarations must appear after vars and before other declarations"
	default:
		return "file-scope declarations are out of order"
	}
}

func collapseEmbedSpacing(src []byte) []byte {
	lines := bytes.Split(src, []byte{'\n'})
	out := make([][]byte, 0, len(lines))

	for i := 0; i < len(lines); i++ {
		out = append(out, lines[i])

		if i+2 >= len(lines) {
			continue
		}

		if !bytes.HasPrefix(bytes.TrimSpace(lines[i]), []byte("//go:embed ")) {
			continue
		}

		if len(bytes.TrimSpace(lines[i+1])) != 0 {
			continue
		}

		next := bytes.TrimSpace(lines[i+2])

		if bytes.HasPrefix(next, []byte("var ")) || bytes.Equal(next, []byte("var (")) {
			i++
		}
	}

	return bytes.Join(out, []byte{'\n'})
}
