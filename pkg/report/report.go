package report

import (
	"fmt"
	"sync"
)

type Finding struct {
	File string
	Line int
	Col  int
	Msg  string
}

type Report struct {
	mu       sync.Mutex
	findings []Finding
}

func New() *Report {
	return &Report{findings: make([]Finding, 0)}
}

func (r *Report) Add(file string, line, col int, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findings = append(r.findings, Finding{file, line, col, msg})
}

func (r *Report) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.findings)
}

func (r *Report) PrintAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, f := range r.findings {
		fmt.Printf("- %s:%d:%d: %s\n", f.File, f.Line, f.Col, f.Msg)
	}
}
