// panicscan/internal/checks/numeric.go

package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/internal/report"
)

// CheckDivisionByZero detects integer division or modulo by a literal zero.
func CheckDivisionByZero(r *report.Reporter, info *types.Info, node ast.Node) {
	expr, ok := node.(*ast.BinaryExpr)
	if !ok {
		return
	}

	if expr.Op != token.QUO && expr.Op != token.REM {
		return
	}

	if val, ok := info.Types[expr.Y]; ok && val.Value != nil {
		lit, ok := expr.Y.(*ast.BasicLit)
		if ok && lit.Kind == token.INT && lit.Value == "0" {
			r.Add(expr.OpPos, "integer division or modulo by zero")
		}
	}
}
