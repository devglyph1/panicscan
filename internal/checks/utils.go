// panicscan/internal/checks/utils.go

package checks

import (
	"go/ast"
	"go/types"
)

// IsBuiltin checks if an identifier refers to a built-in function like panic or close.
// It correctly returns a boolean value.
func IsBuiltin(info *types.Info, ident *ast.Ident) bool {
	obj := info.ObjectOf(ident)
	if obj == nil {
		return false
	}
	// Built-in functions are part of the "universe" scope.
	return types.Universe.Lookup(ident.Name) == obj
}
