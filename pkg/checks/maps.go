package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// writing to a nil map panics.  If the map expression is obviously
// nil (var m map[T]U declared but never make()'d), flag as ERROR.
// Otherwise the map may or may not be nil at runtime → WARNING.
func checkNilMapWrite(fset *token.FileSet, root ast.Node, ti *types.Info, file string, rep *report.Report) {
	ast.Inspect(root, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || as.Tok != token.ASSIGN {
			return true
		}

		for _, lhs := range as.Lhs {
			ix, ok := lhs.(*ast.IndexExpr)
			if !ok {
				continue
			}
			if _, ok := ti.TypeOf(ix.X).(*types.Map); !ok {
				continue
			}

			// try constant nil detection
			if id, ok := ix.X.(*ast.Ident); ok && id.Obj != nil && id.Obj.Kind == ast.Var {
				if ti.Defs[id] == nil && ti.Uses[id] == nil {
					// can't be sure → warning
					pos := fset.Position(ix.X.Pos())
					rep.Warning(file, pos.Line, pos.Column, "possible assignment to entry in nil map")
					continue
				}
			}

			pos := fset.Position(ix.X.Pos())
			rep.Warning(file, pos.Line, pos.Column,
				"map write may panic if the map is nil (ensure map is initialised)")
		}
		return true
	})
}
