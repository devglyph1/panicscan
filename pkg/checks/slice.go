package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

func checkSliceBounds(fset *token.FileSet, f ast.Node, typesInfo *types.Info, file string, r *report.Report) {
	ast.Inspect(f, func(n ast.Node) bool {
		idx, ok := n.(*ast.IndexExpr)
		if !ok {
			return true
		}
		typ := typesInfo.TypeOf(idx.X)
		switch typ.(type) {
		case *types.Array:
			// TODO: If idx is constant out of range, flag
		case *types.Slice:
			// Bounds analysis (if statically known)
		}
		return true
	})
}
