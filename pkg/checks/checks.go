package checks

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/devglyph1/panicscan/pkg/report"
)

func RunAllChecks(fset *token.FileSet, f ast.Node, typesInfo *types.Info, file string, r *report.Report) {
	checkExplicitPanic(fset, f, typesInfo, file, r)
	checkNilDereference(fset, f, typesInfo, file, r)
	checkNilMapWrite(fset, f, typesInfo, file, r)
	checkChannelOps(fset, f, typesInfo, file, r)
	checkSliceBounds(fset, f, typesInfo, file, r)
	checkUnsafeTypeAssertion(fset, f, typesInfo, file, r)
	checkNilFuncCall(fset, f, typesInfo, file, r)
	// Future: Inter-procedural, concurrency pattern checks, etc.
}
