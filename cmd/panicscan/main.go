package main

import (
	"fmt"
	"log"
	"os"

	"github.com/devglyph1/panicscan/pkg/checker"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: panicscan <directory>")
		fmt.Println("Example: panicscan ./...")
		os.Exit(1)
	}
	dir := os.Args[1]

	// If the user provides the `./...` pattern, interpret it as the current directory.
	if dir == "./..." {
		dir = "."
	}

	// Create a new checker
	c := checker.NewChecker()

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
