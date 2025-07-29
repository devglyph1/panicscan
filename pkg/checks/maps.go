package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// CheckMapPanics detects panics related to map operations.
func CheckMapPanics(fset *token.FileSet, node ast.Node, state *StateTracker) []report.PanicInfo {
	var panics []report.PanicInfo

	// Check for write to a nil map
	if assignStmt, ok := node.(*ast.AssignStmt); ok {
		if len(assignStmt.Lhs) == 1 {
			if indexExpr, ok := assignStmt.Lhs[0].(*ast.IndexExpr); ok {
				// Check if the expression is a map type
				if _, ok := state.Info.TypeOf(indexExpr.X).Underlying().(*types.Map); ok {
					if mapIdent, ok := indexExpr.X.(*ast.Ident); ok {
						if state.GetState(mapIdent) == StateNil {
							panics = append(panics, report.PanicInfo{
								Pos:     fset.Position(assignStmt.Pos()),
								Message: "Potential write to a nil map.",
							})
						}
					}
				}
			}
		}
	}

	return panics
}
