// panicscan/internal/report/report.go

package report

import (
	"fmt"
	"go/token"
	"sort"
)

// Finding represents a single potential panic detected by the tool.
type Finding struct {
	Pos     token.Pos
	Message string
}

// Reporter collects and prints findings.
type Reporter struct {
	Fset     *token.FileSet
	Findings []Finding
}

// NewReporter creates a new Reporter.
func NewReporter() *Reporter {
	return &Reporter{
		Fset:     token.NewFileSet(),
		Findings: []Finding{},
	}
}

// Add appends a new finding to the report.
func (r *Reporter) Add(pos token.Pos, message string) {
	r.Findings = append(r.Findings, Finding{Pos: pos, Message: message})
}

// Print outputs all collected findings to standard output.
func (r *Reporter) Print() {
	sort.Slice(r.Findings, func(i, j int) bool {
		posI := r.Fset.Position(r.Findings[i].Pos)
		posJ := r.Fset.Position(r.Findings[j].Pos)
		if posI.Filename != posJ.Filename {
			return posI.Filename < posJ.Filename
		}
		return posI.Line < posJ.Line
	})

	for _, f := range r.Findings {
		position := r.Fset.Position(f.Pos)
		fmt.Printf("- %s:%d:%d: %s\n", position.Filename, position.Line, position.Column, f.Message)
	}
}
