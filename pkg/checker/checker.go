package checker

import (
	"fmt"
	"os"
	"strings"

	"github.com/devglyph1/panicscan/pkg/checks"
	"github.com/devglyph1/panicscan/pkg/report"
	"golang.org/x/tools/go/packages"
)

func Run(pattern string, excludeDirs map[string]struct{}) (int, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles,
		Dir:  ".",
		Env:  os.Environ(),
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return 2, err
	}

	rep := report.New()

	for _, pkg := range pkgs {
		for i, f := range pkg.Syntax {
			filepath := pkg.GoFiles[i]
			skip := false
			for d := range excludeDirs {
				if strings.Contains(filepath, d) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			checks.RunAllChecks(pkg.Fset, f, pkg.TypesInfo, filepath, rep)
		}
	}

	if rep.Count() == 0 {
		fmt.Println("✅ No potential panics found.")
		return 0, nil
	}

	rep.PrintAll()
	return 1, nil
}
