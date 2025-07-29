package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// checkNilDereference performs basic nil pointer dereference detection
// using simple pattern matching and basic data flow analysis.
func checkNilDereference(fset *token.FileSet, root ast.Node, ti *types.Info, file string, rep *report.Report) {
	// Track variables that are known to be nil or non-nil
	nilVars := make(map[string]bool)    // true = definitely nil
	nonNilVars := make(map[string]bool) // true = definitely non-nil

	// First pass: collect nil assignments and checks
	ast.Inspect(root, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			// Look for explicit nil assignments: var x *T = nil
			for i, lhs := range node.Lhs {
				if i < len(node.Rhs) {
					if ident, ok := lhs.(*ast.Ident); ok {
						if isNilIdent(node.Rhs[i]) {
							nilVars[ident.Name] = true
							delete(nonNilVars, ident.Name)
						} else if !isNilIdent(node.Rhs[i]) {
							// Non-nil assignment
							nonNilVars[ident.Name] = true
							delete(nilVars, ident.Name)
						}
					}
				}
			}
		case *ast.IfStmt:
			// Look for nil checks: if x != nil { ... }
			if bin, ok := node.Cond.(*ast.BinaryExpr); ok {
				if ident, ok := bin.X.(*ast.Ident); ok && isNilIdent(bin.Y) {
					switch bin.Op {
					case token.NEQ: // x != nil
						// Inside the if block, x is non-nil
						ast.Inspect(node.Body, func(inner ast.Node) bool {
							nonNilVars[ident.Name] = true
							return true
						})
					case token.EQL: // x == nil
						// Inside the if block, x is nil
						ast.Inspect(node.Body, func(inner ast.Node) bool {
							nilVars[ident.Name] = true
							return true
						})
					}
				}
			}
		}
		return true
	})

	// Second pass: check for potential nil dereferences
	ast.Inspect(root, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.StarExpr:
			// Direct dereference: *ptr
			if ident, ok := node.X.(*ast.Ident); ok {
				pos := fset.Position(node.Star)
				if nilVars[ident.Name] {
					rep.Error(file, pos.Line, pos.Column, "nil pointer dereference")
				} else if !nonNilVars[ident.Name] {
					// Unknown state - could be nil
					rep.Warning(file, pos.Line, pos.Column, "possible nil pointer dereference")
				}
			}
		case *ast.SelectorExpr:
			// Method call or field access: ptr.Method() or ptr.Field
			if ident, ok := node.X.(*ast.Ident); ok {
				// Check if this is a pointer type
				if typ := ti.TypeOf(node.X); typ != nil {
					if _, ok := typ.(*types.Pointer); ok {
						pos := fset.Position(node.Pos())
						if nilVars[ident.Name] {
							rep.Error(file, pos.Line, pos.Column, "nil pointer dereference in field/method access")
						} else if !nonNilVars[ident.Name] {
							rep.Warning(file, pos.Line, pos.Column, "possible nil pointer dereference in field/method access")
						}
					}
				}
			}
		case *ast.IndexExpr:
			// Array/slice indexing: arr[i] (could be nil slice)
			if ident, ok := node.X.(*ast.Ident); ok {
				if typ := ti.TypeOf(node.X); typ != nil {
					if _, ok := typ.(*types.Slice); ok {
						pos := fset.Position(node.Lbrack)
						if nilVars[ident.Name] {
							rep.Error(file, pos.Line, pos.Column, "nil slice access")
						} else if !nonNilVars[ident.Name] {
							rep.Warning(file, pos.Line, pos.Column, "possible nil slice access")
						}
					}
				}
			}
		}
		return true
	})
}
