package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

// RunAllChecks executes every individual check against a single file.
func RunAllChecks(fset *token.FileSet, node ast.Node, ti *types.Info, path string, rep *report.Report) {
	checkExplicitPanic(fset, node, ti, path, rep)
	checkNilDereference(fset, node, ti, path, rep)
	checkNilMapWrite(fset, node, ti, path, rep)
	checkChannelOps(fset, node, ti, path, rep)
	checkSliceBounds(fset, node, ti, path, rep)
	checkUnsafeTypeAssertion(fset, node, ti, path, rep)
	checkNilFuncCall(fset, node, ti, path, rep)
}
