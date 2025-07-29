package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

func checkNilFuncCall(fset *token.FileSet, f ast.Node, typesInfo *types.Info, file string, r *report.Report) {
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// If call.Fun is an identifier and refers to a function variable, check assignment
		// TODO: track variable initialization
		pos := fset.Position(call.Lparen)
		r.Add(file, pos.Line, pos.Column, "Potential call to nil function variable")
		return true
	})
}
