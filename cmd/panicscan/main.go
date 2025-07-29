package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/devglyph1/panicscan/pkg/checker"
)

func main() {
	// Define the exclude-dirs flag
	excludeDirs := flag.String("exclude-dirs", "", "Comma-separated list of directories to exclude from scanning.")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: panicscan --exclude-dirs=\"dir1,dir2\" <directory>")
		fmt.Println("Example: panicscan ./...")
		os.Exit(1)
	}
	dir := flag.Arg(0)

	// If the user provides the `./...` pattern, interpret it as the current directory.
	// The go/packages library will handle the recursive pattern correctly.
	if dir == "./..." {
		dir = "."
	}

	// Create a new checker, passing the excluded directories
	c := checker.NewChecker(excludeDirs)

	// Run the checker on the specified directory
	panics, err := c.CheckDir(dir)
	if err != nil {
		log.Fatalf("Error checking directory: %v", err)
	}

	// Print any found potential panics
	if len(panics) > 0 {
		fmt.Printf("Found %d potential panic(s):\n", len(panics))
		for _, p := range panics {
			fmt.Printf("  - %s:%d:%d: %s\n", p.Pos.Filename, p.Pos.Line, p.Pos.Column, p.Message)
		}
		os.Exit(1) // Exit with a non-zero code to indicate issues were found
	}

	fmt.Println("✅ No potential panics found.")
}
