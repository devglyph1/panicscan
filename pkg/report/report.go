package report

import "go/token"

// PanicInfo holds information about a potential panic.
// It is in its own package to avoid import cycles between the checker and checks packages.
type PanicInfo struct {
	Pos     token.Position
	Message string
}
