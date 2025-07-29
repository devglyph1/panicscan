package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// CheckTypeAssertion detects panics from unsafe type assertions.
func CheckTypeAssertion(fset *token.FileSet, node ast.Node, info *types.Info) []report.PanicInfo {
	var panics []report.PanicInfo

	// Look for x.(T) without the ", ok"
	if assertExpr, ok := node.(*ast.TypeAssertExpr); ok {
		// A single-value type assertion's type is the asserted type.
		// A two-value ("comma, ok") assertion's type is a tuple.
		// We check if the expression's type is NOT a tuple to find single-value assertions.
		if tv, ok := info.Types[assertExpr]; ok {
			if _, isTuple := tv.Type.(*types.Tuple); !isTuple {
				// This is a single-value assertion, which can panic.
				// First, check for the guaranteed panic: assertion on a literal nil.
				if ident, ok := assertExpr.X.(*ast.Ident); ok {
					if info.ObjectOf(ident) == types.Universe.Lookup("nil") {
						panics = append(panics, report.PanicInfo{
							Pos:     fset.Position(assertExpr.Pos()),
							Message: "Type assertion on 'nil' interface will panic.",
						})
						return panics // No need to report twice
					}
				}

				// For any other single-value assertion (on variables, function calls, etc.),
				// report it as a potential panic. This fixes the bug.
				panics = append(panics, report.PanicInfo{
					Pos:     fset.Position(assertExpr.Pos()),
					Message: "Single-value type assertion can panic if type is wrong.",
				})
			}
		}
	}

	return panics
}
