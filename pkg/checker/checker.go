package checker

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/devglyph1/panicscan/pkg/checks"
	"github.com/devglyph1/panicscan/pkg/report"
)

// Checker holds the state for the analysis.
type Checker struct {
	Fset   *token.FileSet
	Info   *types.Info
	Panics []report.PanicInfo
	Files  []*ast.File
	Pkg    *types.Package
}

// NewChecker creates a new Checker.
func NewChecker() *Checker {
	return &Checker{
		Fset: token.NewFileSet(),
		Info: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		},
	}
}

// CheckDir analyzes all .go files in a directory.
func (c *Checker) CheckDir(path string) ([]report.PanicInfo, error) {
	var filesToParse []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip vendor directories and test files
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			filesToParse = append(filesToParse, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", path, err)
	}

	for _, p := range filesToParse {
		f, err := parser.ParseFile(c.Fset, p, nil, parser.AllErrors)
		if err != nil {
			// In a real tool, you might want to collect errors and continue
			return nil, fmt.Errorf("could not parse %s: %w", p, err)
		}
		c.Files = append(c.Files, f)
	}

	if len(c.Files) == 0 {
		return nil, fmt.Errorf("no Go files found in %s", path)
	}

	// Configure the type checker.
	conf := types.Config{Importer: importer.Default(), Error: func(err error) { fmt.Printf("Type-checking error: %v\n", err) }}
	pkg, err := conf.Check(".", c.Fset, c.Files, c.Info)
	if err != nil {
		// Even with errors, the type checker populates some information, so we can often proceed.
		fmt.Printf("Warning: type-checking failed with errors, results may be incomplete: %v\n", err)
	}
	c.Pkg = pkg

	// Run all checks for each file.
	for _, file := range c.Files {
		state := checks.NewStateTracker(c.Info)
		ast.Inspect(file, func(node ast.Node) bool {
			// Run stateful analysis first to update variable states
			state.Track(node)

			// Run all individual checks on the node
			c.runChecks(file, node, state)
			return true
		})
	}

	return c.Panics, nil
}

// runChecks runs all registered checks on a given AST node.
func (c *Checker) runChecks(file *ast.File, node ast.Node, state *checks.StateTracker) {
	c.Panics = append(c.Panics, checks.CheckExplicitPanic(c.Fset, node, c.Info)...)
	c.Panics = append(c.Panics, checks.CheckDivisionByZero(c.Fset, node)...)
	c.Panics = append(c.Panics, checks.CheckNilDereference(c.Fset, node, c.Info, state)...)
	c.Panics = append(c.Panics, checks.CheckSliceBounds(c.Fset, node, c.Info)...)
	c.Panics = append(c.Panics, checks.CheckChannelPanics(c.Fset, node, state)...)
	c.Panics = append(c.Panics, checks.CheckMapPanics(c.Fset, node, state)...)
	c.Panics = append(c.Panics, checks.CheckTypeAssertion(c.Fset, node, c.Info)...)
	c.Panics = append(c.Panics, checks.CheckNilFunctionCall(c.Fset, node, c.Info, state)...)
}
