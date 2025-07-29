package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// calling a nil function variable panics.  Flag only if the callee is
// a plain identifier of func type; anything more complex would need
// interprocedural data-flow to be certain.
func checkNilFuncCall(fset *token.FileSet, root ast.Node, ti *types.Info, file string, rep *report.Report) {
	ast.Inspect(root, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		if _, ok := ti.ObjectOf(id).Type().Underlying().(*types.Signature); !ok {
			return true
		}
		pos := fset.Position(call.Lparen)
		rep.Warning(file, pos.Line, pos.Column, "calling function variable that may be nil")
		return true
	})
}
