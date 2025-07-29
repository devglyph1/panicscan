package checker

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/devglyph1/panicscan/pkg/checks"
	"github.com/devglyph1/panicscan/pkg/report"
	"golang.org/x/tools/go/packages"
)

// Run loads all packages matching pattern, scans their files in parallel
// (max workers goroutines), and prints a consolidated report.
// It returns the intended process-exit status (0 = success, 1 = definite panic found).
func Run(pattern string, exclude map[string]struct{}, workers int) (int, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedFiles,
		Env: os.Environ(),
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return 2, err
	}

	rep := report.New()
	sem := make(chan struct{}, workers) // semaphore to cap concurrency
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, pkg := range pkgs {
		for fi, file := range pkg.Syntax {
			srcPath := pkg.GoFiles[fi]

			// fast path: honour --exclude-dirs
			if skip(srcPath, exclude) {
				continue
			}

			wg.Add(1)
			go func(fset *token.FileSet, fPath string, node ast.Node) {
				defer wg.Done()
				select {
				case sem <- struct{}{}:
				case <-ctx.Done():
					return
				}
				defer func() { <-sem }()
				checks.RunAllChecks(fset, node, pkg.TypesInfo, fPath, rep)
			}(pkg.Fset, srcPath, file)
		}
	}
	wg.Wait()

	rep.PrintAll()

	if rep.Errors() == 0 {
		fmt.Println("✅  No definite panics found.")
		return 0, nil
	}
	return 1, nil
}

func skip(path string, excluded map[string]struct{}) bool {
	for d := range excluded {
		if strings.Contains(filepath.ToSlash(path), "/"+d+"/") {
			return true
		}
	}
	return false
}
