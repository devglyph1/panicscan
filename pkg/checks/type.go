package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// flags unsafe single-value type assertions.
// The two-value form (x, ok := v.(T)) is considered safe.
func checkUnsafeTypeAssertion(fset *token.FileSet, root ast.Node, _ *types.Info, file string, rep *report.Report) {
	// gather assertions used in two-value assignments
	okProtected := map[*ast.TypeAssertExpr]struct{}{}
	ast.Inspect(root, func(n ast.Node) bool {
		if as, ok := n.(*ast.AssignStmt); ok && len(as.Lhs) == 2 && len(as.Rhs) == 1 {
			if tae, ok := as.Rhs[0].(*ast.TypeAssertExpr); ok {
				okProtected[tae] = struct{}{}
			}
		}
		return true
	})
	ast.Inspect(root, func(n ast.Node) bool {
		tae, ok := n.(*ast.TypeAssertExpr)
		if !ok || tae.Type == nil {
			return true
		}
		if _, protected := okProtected[tae]; protected {
			return true
		}
		pos := fset.Position(tae.Pos())
		rep.Error(file, pos.Line, pos.Column, "unsafe type assertion: use the \", ok\" form to avoid panics")
		return true
	})
}
