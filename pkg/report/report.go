package report

import (
	"fmt"
	"sort"
	"sync"
)

// severity encodes diagnostic gravity.
type severity int

const (
	warning severity = iota
	error
)

type finding struct {
	file string
	line int
	col  int
	msg  string
	sev  severity
}

type Report struct {
	mu        sync.Mutex
	findings  []finding
	errCount  int
	warnCount int
}

func New() *Report { return &Report{} }

func (r *Report) add(file string, line, col int, msg string, sev severity) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.findings = append(r.findings, finding{file, line, col, msg, sev})
	if sev == error {
		r.errCount++
	} else {
		r.warnCount++
	}
}

func (r *Report) Error(file string, line, col int, msg string) {
	r.add(file, line, col, msg, error)
}
func (r *Report) Warning(file string, line, col int, msg string) {
	r.add(file, line, col, msg, warning)
}

func (r *Report) Errors() int   { return r.errCount }
func (r *Report) Warnings() int { return r.warnCount }

func (r *Report) PrintAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	sort.Slice(r.findings, func(i, j int) bool {
		if r.findings[i].file != r.findings[j].file {
			return r.findings[i].file < r.findings[j].file
		}
		if r.findings[i].line != r.findings[j].line {
			return r.findings[i].line < r.findings[j].line
		}
		return r.findings[i].col < r.findings[j].col
	})
	for _, f := range r.findings {
		prefix := "-"
		if f.sev == warning {
			prefix = "⚠"
		}
		fmt.Printf("%s %s:%d:%d: %s\n", prefix, f.file, f.line, f.col, f.msg)
	}
}
