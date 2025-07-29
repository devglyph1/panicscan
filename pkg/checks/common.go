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

// constInt returns the int64 value of an expression if—and only if—
// that expression is a *constant integer*. It carefully guards
// against non-integer constants so that constant.Int64Val never panics.
func constInt(ti *types.Info, expr ast.Expr) (int64, bool) {
	tv, ok := ti.Types[expr]
	if !ok || tv.Value == nil {
		return 0, false
	}
	switch tv.Value.Kind() {
	case constant.Int, constant.Unknown:
		// Safe to query.
		return constant.Int64Val(tv.Value)
	default:
		// String, Float, Complex, Bool, etc. → not an int constant.
		return 0, false
	}
}

// withinCloseWaitPattern reports true when a func literal matches the
// canonical safe pattern:  go func() { wg.Wait(); close(ch) }()
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

		// Selector-style calls (wg.Wait(), ch.Close()).
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
