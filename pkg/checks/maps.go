package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

func checkNilMapWrite(fset *token.FileSet, f ast.Node, typesInfo *types.Info, file string, r *report.Report) {
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || assign.Tok != token.ASSIGN {
			return true
		}
		if len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
			indexExpr, ok := assign.Lhs[0].(*ast.IndexExpr)
			if !ok {
				return true
			}
			switch typesInfo.TypeOf(indexExpr.X).(type) {
			case *types.Map:
				// TODO: Track if map is nil (not trivial), just flag risk for demonstration
				pos := fset.Position(assign.Pos())
				r.Add(file, pos.Line, pos.Column, "Potential write to nil map")
			}
		}
		return true
	})
}
