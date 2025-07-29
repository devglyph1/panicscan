package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

func checkNilDereference(fset *token.FileSet, f ast.Node, typesInfo *types.Info, file string, r *report.Report) {
	ast.Inspect(f, func(n ast.Node) bool {
		unary, ok := n.(*ast.StarExpr)
		if !ok {
			return true
		}
		// Example: add data flow nil checking logic here
		pos := fset.Position(unary.Star)
		// TODO: Enhance by tracking variable state using assignment
		r.Add(file, pos.Line, pos.Column, "Potential nil pointer dereference")
		return true
	})
}
