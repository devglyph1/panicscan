package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/devglyph1/panicscan/pkg/checker"
)

func main() {
	excludeFlag := flag.String("exclude-dirs", "",
		"comma-separated list of relative directories to exclude (e.g. \"vendor,tests,mocks\")")
	workers := flag.Int("workers", runtime.NumCPU(),
		"maximum number of parallel files analysed (defaults to GOMAXPROCS)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: panicscan <dir|pattern> [--exclude-dirs=d1,d2] [--workers=N]")
		os.Exit(2)
	}

	target := flag.Arg(0)
	exclude := map[string]struct{}{}
	if *excludeFlag != "" {
		for _, d := range strings.Split(*excludeFlag, ",") {
			exclude[strings.TrimSpace(d)] = struct{}{}
		}
	}

	exit, err := checker.Run(target, exclude, *workers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "panicscan: %v\n", err)
		os.Exit(2)
	}
	os.Exit(exit)
}
