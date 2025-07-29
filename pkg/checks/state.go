package checks

import (
	"go/ast"
	"go/types"
)

// VarState represents the tracked state of a variable.
type VarState int

const (
	StateUnknown VarState = iota
	StateNil
	StateNonNil
	StateClosed
)

// StateTracker tracks the state of variables within a scope.
// NOTE: This is a simplified implementation. A production-grade tool would
// require a full control-flow graph (CFG) and data-flow analysis for accuracy.
type StateTracker struct {
	Info   *types.Info
	States map[types.Object]VarState
}

// NewStateTracker creates a new state tracker.
func NewStateTracker(info *types.Info) *StateTracker {
	return &StateTracker{
		Info:   info,
		States: make(map[types.Object]VarState),
	}
}

// Track analyzes a node and updates the state of variables.
func (st *StateTracker) Track(node ast.Node) {
	switch n := node.(type) {
	// Track assignments
	case *ast.AssignStmt:
		for i, lhsExpr := range n.Lhs {
			if i >= len(n.Rhs) {
				continue
			}
			rhsExpr := n.Rhs[i]
			lhsIdent, ok := lhsExpr.(*ast.Ident)
			if !ok {
				continue
			}
			obj := st.Info.ObjectOf(lhsIdent)
			if obj == nil {
				continue
			}

			// Track assignments to nil
			if rhsIdent, ok := rhsExpr.(*ast.Ident); ok && rhsIdent.Name == "nil" {
				st.States[obj] = StateNil
				continue
			}

			// Track `make` calls for maps and channels, marking them as non-nil
			if call, ok := rhsExpr.(*ast.CallExpr); ok {
				if fun, ok := call.Fun.(*ast.Ident); ok && fun.Name == "make" {
					st.States[obj] = StateNonNil
				}
			}
		}
	// Track `close` calls
	case *ast.ExprStmt:
		if call, ok := n.X.(*ast.CallExpr); ok {
			if fun, ok := call.Fun.(*ast.Ident); ok && fun.Name == "close" && len(call.Args) == 1 {
				if argIdent, ok := call.Args[0].(*ast.Ident); ok {
					if obj := st.Info.ObjectOf(argIdent); obj != nil {
						st.States[obj] = StateClosed
					}
				}
			}
		}
	}
}

// GetState returns the tracked state of an identifier.
func (st *StateTracker) GetState(ident *ast.Ident) VarState {
	if ident.Name == "nil" {
		return StateNil
	}
	if obj := st.Info.ObjectOf(ident); obj != nil {
		if state, found := st.States[obj]; found {
			return state
		}
	}
	return StateUnknown
}
