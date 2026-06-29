package analysis

import "fmt"

// FormatDiagnostic formats a diagnostic in the standard text format:
//
//	<file>:<line>:<col>: <message>
func FormatDiagnostic(d Diagnostic) string {
	pos := d.Pos
	return fmt.Sprintf("%s:%d:%d: %s", pos.Filename, pos.Line, pos.Column, d.Message())
}
