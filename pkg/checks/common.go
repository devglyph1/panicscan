package checks

import (
	"go/ast"
	"go/constant"
	"go/types"
)

// isNilIdent reports whether expr is the predeclared identifier nil.
func isNilIdent(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "nil"
}

// constInt returns the integer constant value of expr, if any.
func constInt(ti *types.Info, expr ast.Expr) (int64, bool) {
	if tv, ok := ti.Types[expr]; ok && tv.Value != nil {
		return constant.Int64Val(tv.Value)
	}
	return 0, false
}

// withinCloseWaitPattern reports true when the func literal matches the
// safe pattern:  go func() { wg.Wait(); close(ch) }()
func withinCloseWaitPattern(node ast.Node) bool {
	fn, ok := node.(*ast.FuncLit)
	if !ok {
		return false
	}
	var sawWait, sawClose bool

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Selector calls (e.g. wg.Wait(), ch.Close()).
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			switch sel.Sel.Name {
			case "Wait":
				sawWait = true
			case "Close":
				sawClose = true
			}
			return true
		}

		// Built-in close(ch).
		if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "close" {
			sawClose = true
		}
		return true
	})

	return sawWait && sawClose
}
