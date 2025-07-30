// panicscan/internal/analyzer/analyzer.go

package analyzer

import (
	"go/ast"

	"github.com/devglyph1/panicscan/internal/checks"
	"github.com/devglyph1/panicscan/internal/report"
	"github.com/devglyph1/panicscan/internal/state"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// Analyzer holds the state for the analysis of a set of packages.
type Analyzer struct {
	reporter *report.Reporter
}

// New creates a new Analyzer.
func New() *Analyzer {
	return &Analyzer{
		reporter: report.NewReporter(),
	}
}

// Reporter returns the analyzer's reporter.
func (a *Analyzer) Reporter() *report.Reporter {
	return a.reporter
}

// Run analyzes a single package.
func (a *Analyzer) Run(pkg *packages.Package) {
	a.reporter.Fset = pkg.Fset

	for _, file := range pkg.Syntax {
		v := &visitor{
			pkg:          pkg,
			stateTracker: state.NewStateTracker(),
			reporter:     a.reporter,
		}
		// Use astutil.Apply for a more powerful traversal with parent context.
		astutil.Apply(file, v.pre, v.post)
	}
}

// visitor holds the state for a single file's AST traversal.
type visitor struct {
	pkg          *packages.Package
	stateTracker *state.StateTracker
	reporter     *report.Reporter
}

// pre is the pre-order traversal function called by astutil.Apply.
func (v *visitor) pre(c *astutil.Cursor) bool {
	node := c.Node()
	if node == nil {
		return true
	}

	// --- Dispatch to all check functions ---
	// Most checks only need the node itself.
	checks.CheckExplicitPanic(v.reporter, v.pkg.TypesInfo, node)
	checks.CheckNilDereference(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	checks.CheckNilMapWrite(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	checks.CheckNilChannelOps(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	checks.CheckClosedChannel(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	checks.CheckDivisionByZero(v.reporter, v.pkg.TypesInfo, node)
	checks.CheckOutOfBounds(v.reporter, v.pkg.TypesInfo, node)
	checks.CheckNilFuncCall(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	// The assertion check is special and needs the cursor for parent context.
	checks.CheckTypeAssertion(v.reporter, c)

	// --- Handle state changes based on control flow ---
	switch n := node.(type) {
	case *ast.FuncDecl:
		// Reset state tracker for each new top-level function.
		v.stateTracker = state.NewStateTracker()
		v.stateTracker.EnterScope() // Create scope for function body
	case *ast.IfStmt:
		// Create a new scope for the 'if' block body.
		v.stateTracker.EnterScope()
		v.stateTracker.UpdateStateFromCondition(n.Cond, v.pkg.TypesInfo, true)
	case *ast.AssignStmt:
		v.handleAssignment(n)
	}

	return true // Continue traversal.
}

// post is the post-order traversal function called by astutil.Apply.
func (v *visitor) post(c *astutil.Cursor) bool {
	node := c.Node()
	if node == nil {
		return true
	}

	// --- Manage state scopes ---
	// When we leave a scope, discard its state changes.
	switch n := node.(type) {
	case *ast.FuncDecl:
		v.stateTracker.LeaveScope() // Leave function body scope
	case *ast.IfStmt:
		v.stateTracker.LeaveScope() // Leave 'if' body scope

		// If there's an 'else' block, create a new scope for it.
		if n.Else != nil {
			v.stateTracker.EnterScope()
			// The condition is false in the 'else' branch.
			v.stateTracker.UpdateStateFromCondition(n.Cond, v.pkg.TypesInfo, false)
		}
	case *ast.BlockStmt:
		// This handles leaving the 'else' block scope.
		// Check if the parent is an IfStmt to avoid double-leaving.
		if _, ok := c.Parent().(*ast.IfStmt); !ok {
			v.stateTracker.LeaveScope()
		}
	}
	return true // Continue traversal.
}

func (v *visitor) handleAssignment(assign *ast.AssignStmt) {
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return // Ignore multi-value assignments for this state tracker
	}

	lhsIdent, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return
	}
	obj := v.pkg.TypesInfo.ObjectOf(lhsIdent)

	rhs := assign.Rhs[0]
	// Case 1: p = nil
	if state.IsNil(rhs, v.pkg.TypesInfo) {
		v.stateTracker.SetState(obj, state.StateNil)
		return
	}

	// Case 2: p = &T{} or similar literal
	if _, ok := rhs.(*ast.CompositeLit); ok {
		v.stateTracker.SetState(obj, state.StateNonNil)
		return
	}

	// Case 3 (FIXED): p = q (assignment from another variable)
	if rhsIdent, ok := rhs.(*ast.Ident); ok {
		if rhsObj := v.pkg.TypesInfo.ObjectOf(rhsIdent); rhsObj != nil {
			// Propagate the state from the right-hand side.
			v.stateTracker.SetState(obj, v.stateTracker.GetState(rhsObj))
			return
		}
	}

	// Default Case: p = someFunc(), etc. We don't know the result.
	v.stateTracker.SetState(obj, state.StateUnknown)
}
