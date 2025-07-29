package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// send/close on nil or closed channels → definite ERROR
// send on channel that *might* be closed/nil → WARNING
func checkChannelOps(fset *token.FileSet, root ast.Node, ti *types.Info, file string, rep *report.Report) {
	ast.Inspect(root, func(n ast.Node) bool {
		switch stmt := n.(type) {

		// close(ch)
		case *ast.CallExpr:
			if id, ok := stmt.Fun.(*ast.Ident); ok && id.Name == "close" {
				if withinCloseWaitPattern(stmt) {
					return true // recognised safe pattern
				}
				arg := stmt.Args[0]
				pos := fset.Position(arg.Pos())
				rep.Error(file, pos.Line, pos.Column, "close on channel may panic if channel already closed or nil")
			}

		// ch <- v
		case *ast.SendStmt:
			typ := ti.TypeOf(stmt.Chan)
			if _, ok := typ.(*types.Chan); !ok {
				return true
			}
			pos := fset.Position(stmt.Arrow)
			rep.Warning(file, pos.Line, pos.Column,
				"send on channel could panic if channel is closed or nil")
		}
		return true
	})
}
