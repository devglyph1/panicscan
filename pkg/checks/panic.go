package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

func checkExplicitPanic(fset *token.FileSet, f ast.Node, typesInfo *types.Info, file string, r *report.Report) {
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if ok {
			if funID, ok := call.Fun.(*ast.Ident); ok && funID.Name == "panic" {
				pos := fset.Position(call.Pos())
				r.Add(file, pos.Line, pos.Column, "Explicit panic call")
			}
		}
		return true
	})
}
