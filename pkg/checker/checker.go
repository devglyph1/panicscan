package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/devglyph1/panicscan/pkg/checks"
	"github.com/devglyph1/panicscan/pkg/report"
	"golang.org/x/tools/go/packages"
)

// Checker holds the state for the analysis.
type Checker struct {
	Panics       []report.PanicInfo
	ExcludedDirs map[string]bool
}

// NewChecker creates a new Checker.
func NewChecker(excludeDirsStr string) *Checker {
	excluded := make(map[string]bool)
	if excludeDirsStr != "" {
		// Normalize and store the excluded directories for easy lookup.
		dirs := strings.Split(excludeDirsStr, ",")
		for _, d := range dirs {
			cleanDir := filepath.Clean(strings.TrimSpace(d))
			if cleanDir != "" {
				excluded[cleanDir] = true
			}
		}
	}
	return &Checker{
		ExcludedDirs: excluded,
	}
}

// CheckDir analyzes all .go files matching the path pattern (e.g., "./...").
func (c *Checker) CheckDir(path string) ([]report.PanicInfo, error) {
	// Use go/packages to correctly load and parse all packages.
	// This correctly handles multiple packages and the "./..." pattern.
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  path, // Set the working directory for the load operation
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Iterate over each loaded package.
	for _, pkg := range pkgs {
		// Report errors but continue analysis if possible.
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				// This provides much cleaner error reporting than before.
				fmt.Fprintf(os.Stderr, "Error loading package %s: %v\n", pkg.ID, e)
			}
			// If a package has errors, its type info might be incomplete, so we skip it.
			continue
		}

		// Check if the package should be excluded.
		if len(pkg.GoFiles) > 0 {
			pkgDir := filepath.Dir(pkg.GoFiles[0])
			isExcluded := false
			for excludedDir := range c.ExcludedDirs {
				// Check if the package directory is a subdirectory of an excluded directory.
				if strings.HasPrefix(pkgDir, excludedDir) {
					isExcluded = true
					break
				}
			}
			if isExcluded {
				continue // Skip this package
			}
		}

		// For each file in the now correctly-loaded package, run the checks.
		for _, file := range pkg.Syntax {
			state := checks.NewStateTracker(pkg.TypesInfo)
			ast.Inspect(file, func(node ast.Node) bool {
				if node == nil {
					return false
				}
				// Run stateful analysis first to update variable states
				state.Track(node)
				// Run all individual checks on the node
				c.runChecks(pkg.Fset, pkg.TypesInfo, file, node, state)
				return true
			})
		}
	}

	return c.Panics, nil
}

// runChecks runs all registered checks on a given AST node.
func (c *Checker) runChecks(fset *token.FileSet, info *types.Info, file *ast.File, node ast.Node, state *checks.StateTracker) {
	// c.Panics = append(c.Panics, checks.CheckExplicitPanic(fset, node, info)...)
	c.Panics = append(c.Panics, checks.CheckDivisionByZero(fset, node)...)
	c.Panics = append(c.Panics, checks.CheckNilDereference(fset, node, info, state)...)
	c.Panics = append(c.Panics, checks.CheckSliceBounds(fset, node, info)...)
	c.Panics = append(c.Panics, checks.CheckChannelPanics(fset, node, state)...)
	c.Panics = append(c.Panics, checks.CheckMapPanics(fset, node, state)...)
	c.Panics = append(c.Panics, checks.CheckTypeAssertion(fset, node, info)...)
	c.Panics = append(c.Panics, checks.CheckNilFunctionCall(fset, node, info, state)...)
}
