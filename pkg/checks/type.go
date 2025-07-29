package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// checkUnsafeTypeAssertion flags type assertions not protected by the ", ok" idiom.
func checkUnsafeTypeAssertion(fset *token.FileSet, f ast.Node, typesInfo *types.Info, file string, r *report.Report) {
	// Collect all type assertions used in two-value assignments
	twoValue := make(map[*ast.TypeAssertExpr]struct{})

	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
			return true
		}
		if tae, ok := assign.Rhs[0].(*ast.TypeAssertExpr); ok {
			twoValue[tae] = struct{}{}
		}
		return true
	})

	ast.Inspect(f, func(n ast.Node) bool {
		tae, ok := n.(*ast.TypeAssertExpr)
		if !ok {
			return true
		}
		// Ignore if used in a two-value assignment
		if _, found := twoValue[tae]; found {
			return true
		}
		pos := fset.Position(tae.Pos())
		r.Add(file, pos.Line, pos.Column, "Potential unsafe type assertion: missing 'ok' check")
		return true
	})
}
