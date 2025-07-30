// panicscan/internal/checks/concurrency.go

package checks

import (
	"go/ast"
	"go/types"

	"github.com/devglyph1/panicscan/internal/report"
	"github.com/devglyph1/panicscan/internal/state"
)

// CheckClosedChannel detects sending on or closing an already closed channel.
func CheckClosedChannel(r *report.Reporter, info *types.Info, st *state.StateTracker, node ast.Node) {
	switch n := node.(type) {
	case *ast.SendStmt:
		if ident, ok := n.Chan.(*ast.Ident); ok {
			obj := info.ObjectOf(ident)
			if st.GetState(obj) == state.StateClosed {
				r.Add(n.Pos(), "send on a possibly closed channel")
			}
		}
	case *ast.CallExpr:
		if len(n.Args) != 1 {
			return
		}
		if fun, ok := n.Fun.(*ast.Ident); ok && fun.Name == "close" && IsBuiltin(info, fun) {
			handleClose(r, info, st, n)
		}
	}
}

func handleClose(r *report.Reporter, info *types.Info, st *state.StateTracker, call *ast.CallExpr) {
	arg := call.Args[0]
	ident, ok := arg.(*ast.Ident)
	if !ok {
		return
	}

	obj := info.ObjectOf(ident)
	if obj == nil {
		return
	}

	if st.GetState(obj) == state.StateClosed {
		r.Add(call.Pos(), "channel may be closed more than once")
		return
	}

	st.SetState(obj, state.StateClosed)
}
