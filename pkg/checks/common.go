package checks

import (
	"go/ast"
)

// Helper to detect if a variable is nil-protected within the scope
func isNilChecked(expr ast.Expr, stmt ast.Stmt) bool {
	// Basic implementation example; can be expanded using
	// advanced AST scanning and symbolic analysis
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok {
		return false
	}
	bin, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	// E.g.: if x != nil
	return bin.Op.String() == "!="
}
