package checks

import (
	"go/ast"
	"go/token"

	"github.com/devglyph1/panicscan/pkg/report"
)

// CheckChannelPanics detects panics related to channel operations.
// It now accepts a path of parent nodes to understand the context of a close() call.
func CheckChannelPanics(fset *token.FileSet, node ast.Node, state *StateTracker, path []ast.Node) []report.PanicInfo {
	// Check for send on closed channel: ch <- val
	if sendStmt, ok := node.(*ast.SendStmt); ok {
		if chIdent, ok := sendStmt.Chan.(*ast.Ident); ok {
			if state.GetState(chIdent) == StateClosed {
				return []report.PanicInfo{{
					Pos:     fset.Position(sendStmt.Pos()),
					Message: "Potential send on a closed channel.",
				}}
			}
		}
	}

	// Check for close of nil or closed channel, with context awareness.
	callExpr, ok := isCloseCall(node)
	if !ok {
		return nil
	}

	// This is a close() call.
	chIdent, ok := callExpr.Args[0].(*ast.Ident)
	if !ok {
		return nil
	}

	// FIX: Check for the safe WaitGroup pattern to reduce false positives.
	// If the close call is inside a goroutine that waits on a WaitGroup, it's likely safe.
	if isSafeWaitGroupClose(path) {
		return nil // It's a safe pattern, don't report.
	}

	// If not the safe pattern, perform the original state-based checks.
	var panics []report.PanicInfo
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
	return panics
}

// isCloseCall checks if an AST node is a call to the built-in `close` function.
func isCloseCall(node ast.Node) (*ast.CallExpr, bool) {
	exprStmt, ok := node.(*ast.ExprStmt)
	if !ok {
		return nil, false
	}
	callExpr, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return nil, false
	}
	fun, ok := callExpr.Fun.(*ast.Ident)
	if !ok || fun.Name != "close" || len(callExpr.Args) != 1 {
		return nil, false
	}
	return callExpr, true
}

// isSafeWaitGroupClose checks if a close() call is part of the idiomatic
// go func() { wg.Wait(); close(ch) }() pattern by inspecting the parent nodes.
func isSafeWaitGroupClose(path []ast.Node) bool {
	// The path is ordered from root to current node. We need to find if the
	// direct parent is a BlockStmt inside a GoStmt's FuncLit.
	if len(path) < 3 {
		return false
	}

	// The node for close() is an ExprStmt. Its parent should be a BlockStmt.
	// The BlockStmt's parent should be a FuncLit.
	// The FuncLit's parent should be a CallExpr.
	// The CallExpr's parent should be a GoStmt.
	// This seems complex. Let's simplify by just looking for the GoStmt in the ancestry.
	for i := len(path) - 1; i >= 0; i-- {
		node := path[i]

		goStmt, ok := node.(*ast.GoStmt)
		if !ok {
			continue
		}

		// We are inside a goroutine. Now check its body for wg.Wait().
		funcLit, ok := goStmt.Call.Fun.(*ast.FuncLit)
		if !ok {
			return false // Not a literal func, too complex to analyze.
		}

		// Check the statements inside the goroutine's body.
		for _, stmt := range funcLit.Body.List {
			exprStmt, ok := stmt.(*ast.ExprStmt)
			if !ok {
				continue
			}
			call, ok := exprStmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			// If we find a `*.Wait()` call, we assume it's a sync.WaitGroup
			// and that this pattern is safe.
			if sel.Sel.Name == "Wait" {
				return true
			}
		}
	}
	return false
}
