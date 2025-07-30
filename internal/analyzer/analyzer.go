// panicscan/internal/analyzer/analyzer.go

package analyzer

import (
	"go/ast"

	"github.com/devglyph1/panicscan/internal/checks"
	"github.com/devglyph1/panicscan/internal/report" // <-- Updated import
	"github.com/devglyph1/panicscan/internal/state"  // <-- Updated import
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
		visitor := &fileVisitor{
			pkg:          pkg,
			stateTracker: state.NewStateTracker(),
			reporter:     a.reporter,
		}
		ast.Walk(visitor, file)
	}
}

// fileVisitor implements the ast.Visitor interface.
type fileVisitor struct {
	pkg          *packages.Package
	stateTracker *state.StateTracker
	reporter     *report.Reporter
}

// Visit is called for each node encountered by ast.Walk.
func (v *fileVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return v
	}

	// Dispatch to various check functions based on node type

	checks.CheckExplicitPanic(v.reporter, v.pkg.TypesInfo, node)
	checks.CheckNilDereference(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	checks.CheckNilMapWrite(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	checks.CheckNilChannelOps(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	checks.CheckClosedChannel(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)
	checks.CheckDivisionByZero(v.reporter, v.pkg.TypesInfo, node)
	checks.CheckOutOfBounds(v.reporter, v.pkg.TypesInfo, node)
	checks.CheckTypeAssertion(v.reporter, node)
	checks.CheckNilFuncCall(v.reporter, v.pkg.TypesInfo, v.stateTracker, node)

	// --- Handle state changes based on control flow ---
	switch n := node.(type) {
	case *ast.FuncDecl:
		v.stateTracker = state.NewStateTracker()
		ast.Walk(v, n.Body)
		return nil

	case *ast.AssignStmt:
		v.handleAssignment(n)

	case *ast.IfStmt:
		v.stateTracker.EnterScope()
		v.stateTracker.UpdateStateFromCondition(n.Cond, v.pkg.TypesInfo, true)
		ast.Walk(v, n.Body)
		v.stateTracker.LeaveScope()

		if n.Else != nil {
			v.stateTracker.EnterScope()
			v.stateTracker.UpdateStateFromCondition(n.Cond, v.pkg.TypesInfo, false)
			ast.Walk(v, n.Else)
			v.stateTracker.LeaveScope()
		}
		return nil
	}

	return v
}

func (v *fileVisitor) handleAssignment(assign *ast.AssignStmt) {
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return
	}

	lhs := assign.Lhs[0]
	rhs := assign.Rhs[0]

	ident, ok := lhs.(*ast.Ident)
	if !ok {
		return
	}
	obj := v.pkg.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return
	}

	if state.IsNil(rhs, v.pkg.TypesInfo) {
		v.stateTracker.SetState(obj, state.StateNil)
		return
	}

	if _, ok := rhs.(*ast.CompositeLit); ok {
		v.stateTracker.SetState(obj, state.StateNonNil)
	} else {
		v.stateTracker.SetState(obj, state.StateUnknown)
	}
}
