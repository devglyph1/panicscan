package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// env tracks, for each variable, whether it is definitely non-nil in the current context.
type env map[string]bool

// copy returns a shallow copy of an env map.
func (e env) copy() env {
	out := env{}
	for k, v := range e {
		out[k] = v
	}
	return out
}

// checkNilDereference traverses expressions and tracks nil-ness context.
func checkNilDereference(fset *token.FileSet, root ast.Node, ti *types.Info, file string, rep *report.Report) {
	// entry: unknown nil-ness for everything
	analyzeNode(root, env{}, fset, ti, file, rep)
}

// analyzeNode recursively analyzes AST nodes with an environment specifying nil facts.
func analyzeNode(node ast.Node, e env, fset *token.FileSet, ti *types.Info, file string, rep *report.Report) {
	switch n := node.(type) {

	case *ast.IfStmt:
		// Analyze the condition for new info, then body with updated env.
		postEnv := analyzeCond(n.Cond, e)
		// then-branch: guaranteed facts
		analyzeNode(n.Body, postEnv, fset, ti, file, rep)
		if n.Else != nil {
			// else-branch: use old env for now (could be improved)
			analyzeNode(n.Else, e, fset, ti, file, rep)
		}
		analyzeNode(n.Cond, e, fset, ti, file, rep)
		return

	case *ast.BinaryExpr:
		// Handle short-circuit for logic op
		if n.Op == token.LAND || n.Op == token.LOR {
			handleShortCircuitBool(n, e, fset, ti, file, rep)
			return
		}
	}

	// For general traversal.
	ast.Inspect(node, func(child ast.Node) bool {
		switch x := child.(type) {
		case *ast.BinaryExpr:
			if x.Op == token.LAND || x.Op == token.LOR {
				handleShortCircuitBool(x, e, fset, ti, file, rep)
				return false
			}
		case *ast.StarExpr:
			if id, ok := x.X.(*ast.Ident); ok {
				v := e[id.Name]
				pos := fset.Position(x.Star)
				if v {
					return true // definitely non-nil
				}
				// Unknown; could be nil unless variable is always initialized
				rep.Warning(file, pos.Line, pos.Column, "possible nil pointer dereference")
			}

		case *ast.SelectorExpr:
			// method/field: ptr.Method() or ptr.Field
			if id, ok := x.X.(*ast.Ident); ok {
				if _, ok := ti.TypeOf(x.X).(*types.Pointer); ok {
					v := e[id.Name]
					pos := fset.Position(x.Pos())
					if v {
						return true // non-nil
					}
					rep.Warning(file, pos.Line, pos.Column, "possible nil pointer dereference in field/method access")
				}
			}

		case *ast.IndexExpr:
			if id, ok := x.X.(*ast.Ident); ok {
				if _, ok := ti.TypeOf(x.X).(*types.Slice); ok {
					v := e[id.Name]
					pos := fset.Position(x.Lbrack)
					if v {
						return true
					}
					rep.Warning(file, pos.Line, pos.Column, "possible nil slice access")
				}
			}
		}
		return true
	})
}

// handleShortCircuitBool analyzes short-circuit boolean logic with context propagation.
func handleShortCircuitBool(be *ast.BinaryExpr, e env, fset *token.FileSet, ti *types.Info, file string, rep *report.Report) {
	switch be.Op {
	case token.LAND:
		// For a && b:
		// - left: in parent env
		// - right: env updated with facts from left
		analyzeNode(be.X, e, fset, ti, file, rep)
		newEnv := analyzeCond(be.X, e)
		analyzeNode(be.Y, newEnv, fset, ti, file, rep)
	case token.LOR:
		// For a || b:
		// - left: in parent env
		// - right: only relevant if left is false, so assume negation of left
		analyzeNode(be.X, e, fset, ti, file, rep)
		newEnv := analyzeCondNegated(be.X, e)
		analyzeNode(be.Y, newEnv, fset, ti, file, rep)
	}
}

// analyzeCond returns a new environment with facts inferred from cond under "true".
func analyzeCond(expr ast.Expr, e env) env {
	switch c := expr.(type) {
	case *ast.BinaryExpr:
		switch c.Op {
		case token.NEQ:
			if id, ok := c.X.(*ast.Ident); ok && isNilIdent(c.Y) {
				ne := e.copy()
				ne[id.Name] = true // definitely non-nil
				return ne
			}
			if id, ok := c.Y.(*ast.Ident); ok && isNilIdent(c.X) {
				ne := e.copy()
				ne[id.Name] = true
				return ne
			}
		case token.EQL:
			if _, ok := c.X.(*ast.Ident); ok && isNilIdent(c.Y) {
				// could add ne[id.Name] = false but we only track non-nil state for now
				return e.copy()
			}
			if _, ok := c.Y.(*ast.Ident); ok && isNilIdent(c.X) {
				return e.copy()
			}
		}
	}
	return e.copy()
}

// analyzeCondNegated returns a new env with facts inferred from cond under "false".
func analyzeCondNegated(expr ast.Expr, e env) env {
	switch c := expr.(type) {
	case *ast.BinaryExpr:
		switch c.Op {
		case token.NEQ:
			if _, ok := c.X.(*ast.Ident); ok && isNilIdent(c.Y) {
				return e.copy()
			}
			if _, ok := c.Y.(*ast.Ident); ok && isNilIdent(c.X) {
				return e.copy()
			}
		case token.EQL:
			if id, ok := c.X.(*ast.Ident); ok && isNilIdent(c.Y) {
				ne := e.copy()
				ne[id.Name] = true // under !(x == nil), x is non-nil
				return ne
			}
			if id, ok := c.Y.(*ast.Ident); ok && isNilIdent(c.X) {
				ne := e.copy()
				ne[id.Name] = true
				return ne
			}
		}
	}
	return e.copy()
}
