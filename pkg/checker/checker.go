package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
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

// visitor helps traverse the AST and keeps track of the parent nodes.
type visitor struct {
	checker *Checker
	fset    *token.FileSet
	info    *types.Info
	state   *checks.StateTracker
	path    []ast.Node // A stack of parent nodes
}

// Visit implements the ast.Visitor interface. It's called for each node in the AST.
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		v.path = v.path[:len(v.path)-1] // Pop from stack when leaving a node
		return nil
	}

	// Run checks with the current node and its parent stack for context.
	v.checker.runChecks(v.fset, v.info, node, v.state, v.path)

	v.path = append(v.path, node) // Push to stack when entering a node
	return v
}

// NewChecker creates a new Checker.
func NewChecker(excludeDirsStr string) *Checker {
	excluded := make(map[string]bool)
	if excludeDirsStr != "" {
		dirs := strings.Split(excludeDirsStr, ",")
		for _, d := range dirs {
			absPath, err := filepath.Abs(strings.TrimSpace(d))
			if err != nil {
				log.Printf("Warning: could not resolve exclude path %s: %v", d, err)
				continue
			}
			if absPath != "" {
				excluded[absPath] = true
			}
		}
	}
	return &Checker{
		ExcludedDirs: excluded,
	}
}

// CheckDir analyzes all .go files matching the path pattern (e.g., "./...").
func (c *Checker) CheckDir(path string) ([]report.PanicInfo, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  path,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				fmt.Fprintf(os.Stderr, "Error loading package %s: %v\n", pkg.ID, e)
			}
			continue
		}

		if len(pkg.GoFiles) > 0 {
			pkgDir := filepath.Dir(pkg.GoFiles[0])
			isExcluded := false
			for excludedDir := range c.ExcludedDirs {
				if strings.HasPrefix(pkgDir, excludedDir) {
					isExcluded = true
					break
				}
			}
			if isExcluded {
				continue
			}
		}

		for _, file := range pkg.Syntax {
			state := checks.NewStateTracker(pkg.TypesInfo)
			// Use the custom visitor to walk the AST, providing parent context to checks.
			vis := &visitor{
				checker: c,
				fset:    pkg.Fset,
				info:    pkg.TypesInfo,
				state:   state,
			}
			ast.Walk(vis, file)
		}
	}

	return c.Panics, nil
}

// runChecks runs all registered checks on a given AST node.
func (c *Checker) runChecks(fset *token.FileSet, info *types.Info, node ast.Node, state *checks.StateTracker, path []ast.Node) {
	state.Track(node) // Update state first

	c.Panics = append(c.Panics, checks.CheckExplicitPanic(fset, node, info)...)
	c.Panics = append(c.Panics, checks.CheckDivisionByZero(fset, node)...)
	c.Panics = append(c.Panics, checks.CheckNilDereference(fset, node, info, state)...)
	c.Panics = append(c.Panics, checks.CheckSliceBounds(fset, node, info)...)
	c.Panics = append(c.Panics, checks.CheckChannelPanics(fset, node, state, path)...) // Pass path for context
	c.Panics = append(c.Panics, checks.CheckMapPanics(fset, node, state)...)
	c.Panics = append(c.Panics, checks.CheckTypeAssertion(fset, node, info)...)
	c.Panics = append(c.Panics, checks.CheckNilFunctionCall(fset, node, info, state)...)
}
