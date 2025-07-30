// panicscan/internal/state/state.go

package state

import (
	"go/ast"
	"go/token"
	"go/types"
)

// VarState represents the tracked state of a variable.
type VarState int

const (
	StateUnknown VarState = iota
	StateNil
	StateNonNil
	StateClosed // For channels
)

// StateTracker manages the state of variables within different scopes.
type StateTracker struct {
	scopes []map[types.Object]VarState
}

// NewStateTracker creates a new state tracker.
func NewStateTracker() *StateTracker {
	return &StateTracker{
		scopes: []map[types.Object]VarState{make(map[types.Object]VarState)},
	}
}

// EnterScope creates a new, nested scope.
func (st *StateTracker) EnterScope() {
	st.scopes = append(st.scopes, make(map[types.Object]VarState))
}

// LeaveScope exits the current scope, discarding its state.
func (st *StateTracker) LeaveScope() {
	if len(st.scopes) > 1 {
		st.scopes = st.scopes[:len(st.scopes)-1]
	}
}

// SetState sets the state of a variable in the current scope.
func (st *StateTracker) SetState(obj types.Object, state VarState) {
	if obj == nil {
		return
	}
	st.scopes[len(st.scopes)-1][obj] = state
}

// GetState retrieves the state of a variable, searching from the innermost scope outwards.
func (st *StateTracker) GetState(obj types.Object) VarState {
	if obj == nil {
		return StateUnknown
	}
	for i := len(st.scopes) - 1; i >= 0; i-- {
		if state, ok := st.scopes[i][obj]; ok {
			return state
		}
	}
	return StateUnknown
}

// UpdateStateFromCondition intelligently updates variable states based on an if-condition.
func (st *StateTracker) UpdateStateFromCondition(cond ast.Expr, info *types.Info, then bool) {
	if binExpr, ok := cond.(*ast.BinaryExpr); ok {
		var ident *ast.Ident
		var isNilCheck bool
		var isNotEqual bool

		if id, ok := binExpr.X.(*ast.Ident); ok && IsNil(binExpr.Y, info) {
			ident = id
			isNilCheck = true
		} else if id, ok := binExpr.Y.(*ast.Ident); ok && IsNil(binExpr.X, info) {
			ident = id
			isNilCheck = true
		}

		if isNilCheck {
			if binExpr.Op == token.NEQ {
				isNotEqual = true
			} else if binExpr.Op == token.EQL {
				isNotEqual = false
			} else {
				return
			}

			obj := info.ObjectOf(ident)
			if then == isNotEqual {
				st.SetState(obj, StateNonNil)
			} else {
				st.SetState(obj, StateNil)
			}
		}
	}
}

// IsNil checks if an expression is the built-in `nil` identifier.
func IsNil(expr ast.Expr, info *types.Info) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		if obj := info.ObjectOf(ident); obj != nil {
			_, isBuiltin := obj.(*types.Nil)
			return isBuiltin
		}
	}
	return false
}
