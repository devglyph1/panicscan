package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// CheckNilDereference identifies potential nil pointer dereferences.
func CheckNilDereference(fset *token.FileSet, node ast.Node, info *types.Info, state *StateTracker) []report.PanicInfo {
	var panics []report.PanicInfo

	// Check for *x where x is known to be nil
	if starExpr, ok := node.(*ast.StarExpr); ok {
		if ident, ok := starExpr.X.(*ast.Ident); ok {
			if state.GetState(ident) == StateNil {
				panics = append(panics, report.PanicInfo{
					Pos:     fset.Position(starExpr.Pos()),
					Message: "Potential nil pointer dereference.",
				})
			}
		}
	}

	// Check for method calls on a nil receiver: x.Method() where x is nil
	if callExpr, ok := node.(*ast.CallExpr); ok {
		if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := selExpr.X.(*ast.Ident); ok {
				if state.GetState(ident) == StateNil {
					panics = append(panics, report.PanicInfo{
						Pos:     fset.Position(callExpr.Pos()),
						Message: "Method call on a nil receiver.",
					})
				}
			}
		}
	}
	return panics
}
