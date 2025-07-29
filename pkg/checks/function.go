package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// CheckNilFunctionCall detects calls to nil function variables.
func CheckNilFunctionCall(fset *token.FileSet, node ast.Node, info *types.Info, state *StateTracker) []report.PanicInfo {
	var panics []report.PanicInfo

	if callExpr, ok := node.(*ast.CallExpr); ok {
		if funIdent, ok := callExpr.Fun.(*ast.Ident); ok {
			// Check if the identifier is a function variable
			if tv, ok := info.Types[funIdent]; ok {
				if _, isSig := tv.Type.(*types.Signature); isSig {
					// It's a function call. Now, check if the variable's state is nil.
					if state.GetState(funIdent) == StateNil {
						panics = append(panics, report.PanicInfo{
							Pos:     fset.Position(callExpr.Pos()),
							Message: "Call to a nil function variable.",
						})
					}
				}
			}
		}
	}

	return panics
}
