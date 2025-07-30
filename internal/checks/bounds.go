// panicscan/internal/checks/bounds.go

package checks

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strconv"

	"github.com/devglyph1/panicscan/internal/report"
)

// CheckOutOfBounds detects index or slice bounds errors on arrays with a known size.
func CheckOutOfBounds(r *report.Reporter, info *types.Info, node ast.Node) {
	switch n := node.(type) {
	case *ast.IndexExpr: // a[i]
		checkIndex(r, info, n)
	case *ast.SliceExpr: // a[i:j:k]
		checkSlice(r, info, n)
	}
}

func checkIndex(r *report.Reporter, info *types.Info, expr *ast.IndexExpr) {
	tv := info.TypeOf(expr.X)
	if tv == nil {
		return
	}

	arr, ok := tv.Underlying().(*types.Array)
	if !ok {
		return
	}
	arrLen := arr.Len()

	if val := info.Types[expr.Index].Value; val != nil {
		if val.Kind() == constant.Int {
			if idx, ok := constant.Int64Val(val); ok {
				if idx >= arrLen {
					r.Add(expr.Lbrack, "index out of range for fixed-size array")
				}
			}
		}
	}
}

func checkSlice(r *report.Reporter, info *types.Info, expr *ast.SliceExpr) {
	tv := info.TypeOf(expr.X)
	if tv == nil {
		return
	}
	arr, ok := tv.Underlying().(*types.Array)
	if !ok {
		return
	}
	arrLen := arr.Len()

	checkBound := func(bound ast.Expr, name string) {
		if bound == nil {
			return
		}
		if val := info.Types[bound].Value; val != nil {
			if b, ok := constant.Int64Val(val); ok {
				if b > arrLen {
					r.Add(bound.Pos(), "slice bound '"+name+"' ("+strconv.FormatInt(b, 10)+") is out of range for fixed-size array of length "+strconv.FormatInt(arrLen, 10))
				}
			}
		}
	}

	checkBound(expr.Low, "low")
	checkBound(expr.High, "high")
	checkBound(expr.Max, "max")
}
