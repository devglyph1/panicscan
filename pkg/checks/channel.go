package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

func checkChannelOps(fset *token.FileSet, f ast.Node, typesInfo *types.Info, file string, r *report.Report) {
	ast.Inspect(f, func(n ast.Node) bool {
		send, ok := n.(*ast.SendStmt)
		if ok {
			typ := typesInfo.TypeOf(send.Chan)
			if _, ok := typ.(*types.Chan); ok {
				pos := fset.Position(send.Arrow)
				// TODO: Add channel closed/nil/analysis
				r.Add(file, pos.Line, pos.Column, "Potential send on closed or nil channel detected")
			}
		}
		// Add more cases for close(), select, etc.
		return true
	})
}
