package checks

import (
	"go/ast"
	"go/types"
)

type VariableState int

const (
	Unknown VariableState = iota
	DefinitelyNil
	DefinitelyNonNil
)

type StateTracker struct {
	varStates map[string]VariableState
}

func NewStateTracker() *StateTracker {
	return &StateTracker{varStates: map[string]VariableState{}}
}

func (st *StateTracker) VisitAssign(name string, state VariableState) {
	st.varStates[name] = state
}

func (st *StateTracker) State(name string) VariableState {
	return st.varStates[name]
}

// Example walk showing how to hook up state tracker to AST
func TrackDataFlow(f ast.Node, typesInfo *types.Info) *StateTracker {
	st := NewStateTracker()
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					rhs := node.Rhs[0]
					if val, ok := rhs.(*ast.Ident); ok && val.Name == "nil" {
						st.VisitAssign(ident.Name, DefinitelyNil)
					} else {
						st.VisitAssign(ident.Name, DefinitelyNonNil)
					}
				}
			}
		}
		return true
	})
	return st
}
