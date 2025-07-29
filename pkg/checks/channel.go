package checks

import (
	"go/ast"
	"go/token"

	"github.com/devglyph1/panicscan/pkg/report"
)

// CheckChannelPanics detects panics related to channel operations.
func CheckChannelPanics(fset *token.FileSet, node ast.Node, state *StateTracker) []report.PanicInfo {
	var panics []report.PanicInfo

	// Check for send on closed channel: ch <- val
	if sendStmt, ok := node.(*ast.SendStmt); ok {
		if chIdent, ok := sendStmt.Chan.(*ast.Ident); ok {
			if state.GetState(chIdent) == StateClosed {
				panics = append(panics, report.PanicInfo{
					Pos:     fset.Position(sendStmt.Pos()),
					Message: "Potential send on a closed channel.",
				})
			}
		}
	}

	// Check for close of nil or closed channel
	if exprStmt, ok := node.(*ast.ExprStmt); ok {
		if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
			if fun, ok := callExpr.Fun.(*ast.Ident); ok && fun.Name == "close" && len(callExpr.Args) == 1 {
				if chIdent, ok := callExpr.Args[0].(*ast.Ident); ok {
					chState := state.GetState(chIdent)
					if chState == StateClosed {
						panics = append(panics, report.PanicInfo{
							Pos:     fset.Position(callExpr.Pos()),
							Message: "Potential close of a closed channel.",
						})
					} else if chState == StateNil {
						panics = append(panics, report.PanicInfo{
							Pos:     fset.Position(callExpr.Pos()),
							Message: "Potential close of a nil channel.",
						})
					}
				}
			}
		}
	}
	return panics
}
