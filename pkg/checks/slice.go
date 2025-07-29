package checks

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"github.com/devglyph1/panicscan/pkg/report"
)

// CheckSliceBounds checks for obvious out-of-bounds slice or array access.
func CheckSliceBounds(fset *token.FileSet, node ast.Node, info *types.Info) []report.PanicInfo {
	var panics []report.PanicInfo

	// Check for arr[index]
	if indexExpr, ok := node.(*ast.IndexExpr); ok {
		tv := info.Types[indexExpr.X]
		if tv.Type == nil {
			return nil
		}

		var length int64 = -1

		// Determine the length if it's a fixed-size array
		if arr, ok := tv.Type.Underlying().(*types.Array); ok {
			length = arr.Len()
		}

		// If we have a known length and the index is a literal
		if length != -1 {
			if lit, ok := indexExpr.Index.(*ast.BasicLit); ok && lit.Kind == token.INT {
				if index, err := strconv.ParseInt(lit.Value, 10, 64); err == nil {
					if index >= length || index < 0 {
						panics = append(panics, report.PanicInfo{
							Pos:     fset.Position(indexExpr.Pos()),
							Message: "Index out of range: index is " + lit.Value + " but array length is " + strconv.FormatInt(length, 10),
						})
					}
				}
			}
		}
	}

	// Check for slice[low:high]
	if sliceExpr, ok := node.(*ast.SliceExpr); ok {
		tv := info.Types[sliceExpr.X]
		if tv.Type == nil {
			return nil
		}

		if arr, ok := tv.Type.Underlying().(*types.Array); ok {
			length := arr.Len()
			// Check high index if it's a literal
			if high, ok := sliceExpr.High.(*ast.BasicLit); ok && high.Kind == token.INT {
				if highIndex, err := strconv.ParseInt(high.Value, 10, 64); err == nil {
					if highIndex > length {
						panics = append(panics, report.PanicInfo{
							Pos:     fset.Position(sliceExpr.Pos()),
							Message: "Slice bound out of range: high index is " + high.Value + " but array length is " + strconv.FormatInt(length, 10),
						})
					}
				}
			}
		}
	}

	return panics
}
