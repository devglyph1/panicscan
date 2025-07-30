// panicscan/internal/checks/controlflow.go

package checks

import (
	"go/ast"
	"go/types"

	"github.com/devglyph1/panicscan/internal/report"
	"github.com/devglyph1/panicscan/internal/state"
)

// CheckExplicitPanic flags direct calls to the built-in panic function.
func CheckExplicitPanic(r *report.Reporter, info *types.Info, node ast.Node) {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return
	}

	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return
	}

	if ident.Name == "panic" {
		if IsBuiltin(info, ident) {
			r.Add(call.Pos(), "explicit call to panic")
		}
	}
}

// CheckNilFuncCall detects calls to function variables that could be nil.
func CheckNilFuncCall(r *report.Reporter, info *types.Info, st *state.StateTracker, node ast.Node) {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return
	}

	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return
	}

	obj := info.ObjectOf(ident)
	if _, isVar := obj.(*types.Var); !isVar {
		return
	}

	if _, isSig := obj.Type().Underlying().(*types.Signature); !isSig {
		return
	}

	if st.GetState(obj) == state.StateNil {
		r.Add(call.Pos(), "call to a nil function variable")
	}
}
