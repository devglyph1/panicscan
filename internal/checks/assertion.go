// panicscan/internal/checks/assertion.go

package checks

import (
	"go/ast"

	"github.com/devglyph1/panicscan/internal/report"
	"golang.org/x/tools/go/ast/astutil"
)

// CheckTypeAssertion finds single-value type assertions that are not protected.
// It requires the AST cursor to check the parent node.
func CheckTypeAssertion(r *report.Reporter, c *astutil.Cursor) {
	// We are looking for x.(T)
	assertExpr, ok := c.Node().(*ast.TypeAssertExpr)
	if !ok {
		return
	}

	// The safe form `v, ok := x.(T)` appears as a TypeAssertExpr
	// as the right-hand side of an assignment statement with two
	// variables on the left-hand side.
	// We check the parent node to see if this is the case.

	parent := c.Parent()
	if assign, ok := parent.(*ast.AssignStmt); ok {
		// If the parent is an assignment, check if it's a two-value assignment.
		if len(assign.Lhs) == 2 {
			// This is the `v, ok := ...` form, which is safe.
			return
		}
	}

	// If we're here, it's not a safe, two-value assignment.
	// It's a single-value assertion that will panic on type mismatch.
	r.Add(assertExpr.Pos(), "unprotected type assertion will panic on type mismatch")
}
