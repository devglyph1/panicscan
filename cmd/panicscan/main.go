package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/devglyph1/panicscan/pkg/checker"
)

func main() {
	excludeFlag := flag.String("exclude-dirs", "", "comma separated list of dirs to exclude from scan")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("Usage: panicscan <dir or pattern> [--exclude-dirs=...]")
		os.Exit(2)
	}
	target := flag.Arg(0)
	excludeDirs := map[string]struct{}{}
	if *excludeFlag != "" {
		for _, d := range strings.Split(*excludeFlag, ",") {
			excludeDirs[strings.TrimSpace(d)] = struct{}{}
		}
	}

	exitCode, err := checker.Run(target, excludeDirs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed: %v\n", err)
		os.Exit(2)
	}
	os.Exit(exitCode)
}
