package spacing

import (
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/oullin/go-fmt/semantic/rules"
)

type Rule struct{}

func New() Rule {
	return Rule{}
}

func (Rule) Name() string {
	return "spacing"
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

	lineStarts := buildLineStarts(src)
	insertions := map[int]struct{}{}
	aliases := buildImportAliases(file)

	var violations []rules.Violation

	inspectStmtLists(file, func(list []ast.Stmt) {
		for i := 0; i < len(list)-1; i++ {
			current := list[i]
			next := list[i+1]

			if message, ok := statementGapRule(current, next, aliases); ok {
				recordInsertion(fset, filename, lineStarts, insertions, &violations, current.End(), next.Pos(), message)
			}
		}
	})

	for i := 0; i < len(file.Decls)-1; i++ {
		current := file.Decls[i]
		next := file.Decls[i+1]

		if message, ok := declGapRule(current, next); ok {
			recordInsertion(fset, filename, lineStarts, insertions, &violations, current.End(), next.Pos(), message)
		}
	}

	if len(insertions) == 0 {
		return violations, src, nil
	}

	return violations, applyInsertions(src, insertions), nil
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

func recordInsertion(fset *token.FileSet, filename string, lineStarts []int, insertions map[int]struct{}, violations *[]rules.Violation, currentEnd, nextPos token.Pos, message string) {
	endLine := fset.Position(currentEnd).Line
	nextLine := fset.Position(nextPos).Line

	if nextLine >= endLine+2 {
		return
	}

	*violations = append(*violations, rules.Violation{
		Rule:    "spacing",
		File:    filename,
		Line:    nextLine,
		Message: message,
	})

	insertions[lineStartOffset(lineStarts, nextLine)] = struct{}{}
}
