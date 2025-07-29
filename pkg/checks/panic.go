package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// explicit panic calls are definite runtime panics ⇒ severity = error.
func checkExplicitPanic(fset *token.FileSet, n ast.Node, _ *types.Info, file string, rep *report.Report) {
	ast.Inspect(n, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "panic" {
			pos := fset.Position(call.Lparen)
			rep.Error(file, pos.Line, pos.Column, "explicit panic() call")
		}
		return true
	})
}
