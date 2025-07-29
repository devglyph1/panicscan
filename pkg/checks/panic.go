package checks

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"github.com/devglyph1/panicscan/pkg/report"
)

// CheckExplicitPanic looks for direct calls to the built-in panic function.
func CheckExplicitPanic(fset *token.FileSet, node ast.Node, info *types.Info) []report.PanicInfo {
	var panics []report.PanicInfo
	if call, ok := node.(*ast.CallExpr); ok {
		// Check if the function being called is the built-in `panic`
		if ident, ok := call.Fun.(*ast.Ident); ok {
			if obj := info.Uses[ident]; obj != nil {
				// Correctly check if the object is the built-in panic function.
				if b, ok := obj.(*types.Builtin); ok && b.Name() == "panic" {
					panics = append(panics, report.PanicInfo{
						Pos:     fset.Position(node.Pos()),
						Message: "Explicit call to panic()",
					})
				}
			}
		}
	}
	return panics
}

// CheckDivisionByZero looks for integer division by a literal zero.
func CheckDivisionByZero(fset *token.FileSet, node ast.Node) []report.PanicInfo {
	var panics []report.PanicInfo
	if binaryExpr, ok := node.(*ast.BinaryExpr); ok && (binaryExpr.Op == token.QUO || binaryExpr.Op == token.REM) {
		if lit, ok := binaryExpr.Y.(*ast.BasicLit); ok && lit.Kind == token.INT {
			if val, err := strconv.Atoi(lit.Value); err == nil && val == 0 {
				panics = append(panics, report.PanicInfo{
					Pos:     fset.Position(node.Pos()),
					Message: "Potential division by zero.",
				})
			}
		}
	}
	return panics
}
