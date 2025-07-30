// panicscan/internal/checks/nil.go

package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/internal/report"
	"github.com/devglyph1/panicscan/internal/state"
)

// CheckNilDereference detects `*p` where `p` could be nil.
func CheckNilDereference(r *report.Reporter, info *types.Info, st *state.StateTracker, node ast.Node) {
	unary, ok := node.(*ast.UnaryExpr)
	if !ok || unary.Op != token.MUL {
		return
	}

	ident, ok := unary.X.(*ast.Ident)
	if !ok {
		return
	}

	if obj := info.ObjectOf(ident); obj != nil {
		if st.GetState(obj) == state.StateNil {
			r.Add(unary.Pos(), "nil pointer dereference")
		}
	}
}

// CheckNilMapWrite detects writes to a map that could be nil.
func CheckNilMapWrite(r *report.Reporter, info *types.Info, st *state.StateTracker, node ast.Node) {
	assign, ok := node.(*ast.AssignStmt)
	if !ok {
		return
	}

	for _, lhs := range assign.Lhs {
		indexExpr, ok := lhs.(*ast.IndexExpr)
		if !ok {
			continue
		}

		if _, isMap := info.TypeOf(indexExpr.X).Underlying().(*types.Map); !isMap {
			continue
		}

		ident, ok := indexExpr.X.(*ast.Ident)
		if !ok {
			continue
		}

		if obj := info.ObjectOf(ident); obj != nil {
			if st.GetState(obj) == state.StateNil {
				r.Add(indexExpr.Pos(), "write to a nil map")
			}
		}
	}
}

// CheckNilChannelOps detects sends to or closing of a nil channel.
func CheckNilChannelOps(r *report.Reporter, info *types.Info, st *state.StateTracker, node ast.Node) {
	var ch ast.Expr
	var pos token.Pos
	var action string

	switch n := node.(type) {
	case *ast.SendStmt:
		ch = n.Chan
		pos = n.Arrow
		action = "send on"
	case *ast.CallExpr:
		if len(n.Args) != 1 {
			return
		}
		ident, ok := n.Fun.(*ast.Ident)
		if !ok || ident.Name != "close" || !IsBuiltin(info, ident) {
			return
		}
		ch = n.Args[0]
		pos = n.Lparen
		action = "close of"
	default:
		return
	}

	ident, ok := ch.(*ast.Ident)
	if !ok {
		return
	}

	if _, isChan := info.TypeOf(ident).Underlying().(*types.Chan); !isChan {
		return
	}

	if obj := info.ObjectOf(ident); obj != nil {
		if st.GetState(obj) == state.StateNil {
			r.Add(pos, action+" a nil channel")
		}
	}
}
