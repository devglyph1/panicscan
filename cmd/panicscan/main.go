// panicscan/cmd/panicscan/main.go

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devglyph1/panicscan/internal/analyzer"
	"golang.org/x/tools/go/packages"
)

const usage = `
panicscan: A static analysis tool to find potential panics in Go code.

Usage:
    panicscan [flags] [path ...]

Examples:
    panicscan .
    panicscan ./...
    panicscan --exclude-dirs="vendor,tests" ./...

Flags:
`

func main() {
	// 1. Define and parse command-line flags
	var excludeDirsFlag string
	flag.StringVar(&excludeDirsFlag, "exclude-dirs", "", "Comma-separated list of relative directory paths to exclude.")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	patterns := flag.Args()
	if len(patterns) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Path argument is mandatory. Please specify a directory to scan (e.g., . or ./...).")
		flag.Usage()
		os.Exit(1)
	}

	// 2. Load and type-check the packages
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule,
		// It's often useful to set build flags for tools, e.g., for different OS/arch
		// BuildFlags: []string{"-tags", "some_tag"},
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading packages: %v\n", err)
		os.Exit(1)
	}

	// 3. Filter excluded directories
	excludedDirs := parseExcludedDirs(excludeDirsFlag)
	filteredPkgs := filterPackages(pkgs, excludedDirs)

	if len(filteredPkgs) == 0 {
		fmt.Println("No packages to analyze after filtering.")
		os.Exit(0)
	}

	// 4. Run the analysis
	coreAnalyzer := analyzer.New()
	for _, pkg := range filteredPkgs {
		// Print errors for a package but continue analysis on others
		if len(pkg.Errors) > 0 {
			for _, e := range pkg.Errors {
				fmt.Fprintf(os.Stderr, "Error in package %s: %v\n", pkg.ID, e)
			}
			continue
		}
		coreAnalyzer.Run(pkg)
	}

	// 5. Report results
	reporter := coreAnalyzer.Reporter()
	if len(reporter.Findings) == 0 {
		fmt.Println("✅ No potential panics found.")
		os.Exit(0)
	}

	fmt.Println("🚨 Potential panics found:")
	reporter.Print()
	os.Exit(1)
}

func parseExcludedDirs(flagValue string) map[string]struct{} {
	excluded := make(map[string]struct{})
	if flagValue == "" {
		return excluded
	}
	dirs := strings.Split(flagValue, ",")
	for _, dir := range dirs {
		absPath, err := filepath.Abs(strings.TrimSpace(dir))
		if err == nil {
			excluded[absPath] = struct{}{}
		}
	}
	return excluded
}

func filterPackages(pkgs []*packages.Package, excludedDirs map[string]struct{}) []*packages.Package {
	var filtered []*packages.Package
	for _, pkg := range pkgs {
		// Get the directory of the first Go file as a proxy for the package's directory
		if len(pkg.GoFiles) == 0 {
			continue
		}
		pkgDir := filepath.Dir(pkg.GoFiles[0])
		if _, isExcluded := excludedDirs[pkgDir]; !isExcluded {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}
