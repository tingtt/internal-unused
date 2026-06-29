package analysis

import (
	"fmt"
	"go/token"
)

// DiagnosticCode is the machine-readable identifier for the diagnostic reason.
type DiagnosticCode string

const (
	CodeUnused   DiagnosticCode = "unused"
	CodeTestOnly DiagnosticCode = "test-only"
)

// Diagnostic is a single analysis finding.
type Diagnostic struct {
	Pos  token.Position
	Kind DeclKind
	Name string         // qualified declaration name
	Code DiagnosticCode
}

// Message returns the human-readable diagnostic message.
func (d Diagnostic) Message() string {
	switch d.Code {
	case CodeUnused:
		return fmt.Sprintf("exported %s %s is unused", d.Kind, d.Name)
	case CodeTestOnly:
		return fmt.Sprintf("exported %s %s is only used by tests", d.Kind, d.Name)
	default:
		return fmt.Sprintf("exported %s %s: %s", d.Kind, d.Name, d.Code)
	}
}
