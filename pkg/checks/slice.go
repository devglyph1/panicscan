package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// detect obviously out-of-range constant index/slice ops at compile time.
func checkSliceBounds(fset *token.FileSet, root ast.Node, ti *types.Info, file string, rep *report.Report) {
	ast.Inspect(root, func(n ast.Node) bool {
		switch expr := n.(type) {
		case *ast.IndexExpr:
			base := expr.X
			idx := expr.Index
			// only handle constant indexes for now (fast + no FP)
			if k, ok := constInt(ti, idx); ok {
				switch t := ti.TypeOf(base).(type) {
				case *types.Array:
					if k < 0 || k >= t.Len() {
						pos := fset.Position(idx.Pos())
						rep.Error(file, pos.Line, pos.Column, "constant index out of range")
					}
				}
			}

		case *ast.SliceExpr:
			if expr.High == nil {
				return true
			}
			hi, ok1 := constInt(ti, expr.High)
			lo, ok2 := constInt(ti, expr.Low)
			if !ok1 || (expr.Low != nil && !ok2) {
				return true
			}
			if arr, ok := ti.TypeOf(expr.X).(*types.Array); ok {
				if hi > arr.Len() || lo < 0 || lo > hi {
					pos := fset.Position(expr.High.Pos())
					rep.Error(file, pos.Line, pos.Column, "constant slice bounds out of range")
				}
			}
		}
		return true
	})
}
