package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// env tracks, for each variable, whether it is definitely non-nil in the current context.
type env map[string]bool

func (e env) copy() env {
	out := env{}
	for k, v := range e {
		out[k] = v
	}
	return out
}

func checkNilDereference(fset *token.FileSet, root ast.Node, ti *types.Info, file string, rep *report.Report) {
	track := make(env)
	ast.Inspect(root, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			for i, lhs := range node.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}
				if i < len(node.Rhs) {
					// Assignment via address-of: x := &T{...} or x = &T{...}
					switch rhs := node.Rhs[i].(type) {
					case *ast.UnaryExpr:
						if rhs.Op == token.AND {
							track[id.Name] = true
							continue
						}
					}
					if isNilIdent(node.Rhs[i]) {
						track[id.Name] = false
					}
				}
			}
		case *ast.DeclStmt:
			// var x = &T{...}
			decl, ok := node.Decl.(*ast.GenDecl)
			if !ok {
				return true
			}
			for _, spec := range decl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vs.Names {
					if vs.Values == nil || i >= len(vs.Values) {
						continue
					}
					switch rhs := vs.Values[i].(type) {
					case *ast.UnaryExpr:
						if rhs.Op == token.AND {
							track[name.Name] = true
							continue
						}
					}
					if isNilIdent(vs.Values[i]) {
						track[name.Name] = false
					}
				}
			}
		case *ast.StarExpr:
			// Deref: *x
			if id, ok := node.X.(*ast.Ident); ok {
				pos := fset.Position(node.Star)
				if nonil, ok := track[id.Name]; ok && nonil {
					// safe; skip
				} else {
					rep.Warning(file, pos.Line, pos.Column, "possible nil pointer dereference")
				}
			}
		case *ast.SelectorExpr:
			// deref via field/method: x.Method()
			if id, ok := node.X.(*ast.Ident); ok {
				if _, ok := ti.TypeOf(node.X).(*types.Pointer); ok {
					pos := fset.Position(node.Pos())
					if nonil, ok := track[id.Name]; ok && nonil {
						// safe; skip
					} else {
						rep.Warning(file, pos.Line, pos.Column, "possible nil pointer dereference in field/method access")
					}
				}
			}
		}
		return true
	})
}
