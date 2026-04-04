package spacing

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"strconv"
	"strings"
)

type importAliases map[string]string

type boundary struct {
	family   string
	label    string
	leading  bool
	trailing bool
}

var stdlibSpacingImports = map[string]string{
	"sort":         "sort",
	"slices":       "slices",
	"math/rand":    "rand",
	"math/rand/v2": "rand",
}

func statementGapRule(current ast.Stmt, next ast.Stmt, aliases importAliases) (string, bool) {
	currentBoundary := classifyStatement(current, aliases)
	nextBoundary := classifyStatement(next, aliases)

	if nextBoundary.leading && preferLeadingMessage(nextBoundary) && !sameFamily(currentBoundary, nextBoundary) {
		return fmt.Sprintf("missing blank line before %s", nextBoundary.label), true
	}

	if currentBoundary.trailing && !sameFamily(currentBoundary, nextBoundary) {
		return fmt.Sprintf("missing blank line after %s", currentBoundary.label), true
	}

	if nextBoundary.leading && !sameFamily(currentBoundary, nextBoundary) {
		return fmt.Sprintf("missing blank line before %s", nextBoundary.label), true
	}

	return "", false
}

func declGapRule(current ast.Decl, next ast.Decl) (string, bool) {
	currentBoundary := classifyDecl(current)
	nextBoundary := classifyDecl(next)

	if currentBoundary.trailing && !sameFamily(currentBoundary, nextBoundary) {
		return fmt.Sprintf("missing blank line after %s", currentBoundary.label), true
	}

	if nextBoundary.leading && !sameFamily(currentBoundary, nextBoundary) {
		return fmt.Sprintf("missing blank line before %s", nextBoundary.label), true
	}

	return "", false
}

func preferLeadingMessage(next boundary) bool {
	switch next.label {
	case "sort call",
		"rand call",
		"return statement",
		"continue statement",
		"break statement",
		"goto statement",
		"fallthrough statement",
		"next middleware call",
		"route group call",
		"mutex read unlock":
		return true
	default:
		return false
	}
}

func classifyStatement(stmt ast.Stmt, aliases importAliases) boundary {
	if callBoundary, ok := classifyCallBoundary(stmt, aliases); ok {
		return callBoundary
	}

	switch typed := stmt.(type) {
	case *ast.IfStmt:
		return boundary{label: "if statement", leading: true, trailing: true}
	case *ast.ForStmt:
		return boundary{label: "for loop", leading: true, trailing: true}
	case *ast.RangeStmt:
		return boundary{label: "range loop", leading: true, trailing: true}
	case *ast.SwitchStmt:
		return boundary{label: "switch statement", leading: true, trailing: true}
	case *ast.TypeSwitchStmt:
		return boundary{label: "type switch", leading: true, trailing: true}
	case *ast.SelectStmt:
		return boundary{label: "select statement", leading: true, trailing: true}
	case *ast.DeferStmt:
		return boundary{label: "defer statement", leading: true, trailing: true}
	case *ast.ReturnStmt:
		return boundary{label: "return statement", leading: true}
	case *ast.BranchStmt:
		return boundary{label: fmt.Sprintf("%s statement", typed.Tok), leading: true, trailing: true}
	case *ast.DeclStmt:
		return classifyGenDecl(typed.Decl)
	case *ast.AssignStmt:
		if isFuncLiteralShortAssign(typed) {
			return boundary{
				family:   "function assignment",
				label:    "function assignment",
				leading:  true,
				trailing: true,
			}
		}
	}

	return boundary{}
}

func classifyDecl(decl ast.Decl) boundary {
	return classifyGenDecl(decl)
}

func classifyGenDecl(decl ast.Decl) boundary {
	genDecl, ok := decl.(*ast.GenDecl)

	if !ok {
		return boundary{}
	}

	switch genDecl.Tok {
	case token.CONST:
		return boundary{
			family:   "const declaration",
			label:    "const declaration",
			leading:  true,
			trailing: true,
		}
	case token.VAR:
		return boundary{
			family:   "var declaration",
			label:    "var declaration",
			leading:  true,
			trailing: true,
		}
	case token.TYPE:
		return boundary{
			label:    "type definition",
			leading:  true,
			trailing: true,
		}
	default:
		return boundary{}
	}
}

func classifyCallBoundary(stmt ast.Stmt, aliases importAliases) (boundary, bool) {
	call, ok := standaloneCall(stmt)

	if !ok {
		return boundary{}, false
	}

	if label, ok := stdlibSpacedCallLabel(call, aliases); ok {
		return boundary{
			family:   renderNode(call.Fun),
			label:    label,
			leading:  true,
			trailing: true,
		}, true
	}

	if isMutexReadLockCall(call) {
		return boundary{
			family:   renderNode(call.Fun),
			label:    "mutex read lock",
			trailing: true,
		}, true
	}

	if isMutexReadUnlockCall(call) {
		return boundary{
			family:  renderNode(call.Fun),
			label:   "mutex read unlock",
			leading: true,
		}, true
	}

	if isNextServeHTTPCall(call) {
		return boundary{
			family:  renderNode(call.Fun),
			label:   "next middleware call",
			leading: true,
		}, true
	}

	if isRouteGroupCall(call) {
		return boundary{
			family:  renderNode(call.Fun),
			label:   "route group call",
			leading: true,
		}, true
	}

	family := renderNode(call.Fun)

	return boundary{
		family:   family,
		label:    fmt.Sprintf("%s call", family),
		leading:  true,
		trailing: true,
	}, true
}

func sameFamily(current boundary, next boundary) bool {
	return current.family != "" && current.family == next.family
}

func standaloneCall(stmt ast.Stmt) (*ast.CallExpr, bool) {
	exprStmt, ok := stmt.(*ast.ExprStmt)

	if !ok {
		return nil, false
	}

	call, ok := exprStmt.X.(*ast.CallExpr)

	return call, ok
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

func stdlibSpacedCallLabel(call *ast.CallExpr, aliases importAliases) (string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)

	if !ok {
		return "", false
	}

	pkgIdent, ok := selector.X.(*ast.Ident)

	if !ok {
		return "", false
	}

	switch aliases[pkgIdent.Name] {
	case "sort":
		return "sort call", true
	case "slices":
		return "sort call", strings.HasPrefix(selector.Sel.Name, "Sort")
	case "math/rand", "math/rand/v2":
		return "rand call", true
	default:
		return "", false
	}
}

func isFuncLiteralShortAssign(assign *ast.AssignStmt) bool {
	return assign.Tok == token.DEFINE && len(assign.Rhs) == 1 && isFuncLiteral(assign.Rhs[0])
}

func isFuncLiteral(expr ast.Expr) bool {
	_, ok := expr.(*ast.FuncLit)

	return ok
}

func isMutexReadLockCall(call *ast.CallExpr) bool {
	return matchesSelectorChain(call.Fun, "mu", "RLock")
}

func isMutexReadUnlockCall(call *ast.CallExpr) bool {
	return matchesSelectorChain(call.Fun, "mu", "RUnlock")
}

func isNextServeHTTPCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)

	if !ok || selector.Sel.Name != "ServeHTTP" {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)

	return ok && ident.Name == "next"
}

func isRouteGroupCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)

	if !ok || selector.Sel.Name != "Group" {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)

	if !ok {
		return false
	}

	return ident.Name == "route" || ident.Name == "routes"
}

func matchesSelectorChain(expr ast.Expr, receiverName, methodName string) bool {
	selector, ok := expr.(*ast.SelectorExpr)

	if !ok || selector.Sel.Name != methodName {
		return false
	}

	receiver, ok := selector.X.(*ast.SelectorExpr)

	return ok && receiver.Sel.Name == receiverName
}

func renderNode(node ast.Node) string {
	var out bytes.Buffer

	if err := format.Node(&out, token.NewFileSet(), node); err != nil {
		return "call"
	}

	return out.String()
}
