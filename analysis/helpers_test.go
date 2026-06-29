package analysis_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/tingtt/internal-unused/analysis"
)

// testdataDir returns the absolute path to the given testdata subdirectory.
func testdataDir(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

// runAnalysis is a helper that loads and analyses a testdata module.
func runAnalysis(t *testing.T, dir string, mode analysis.Mode) []analysis.Diagnostic {
	t.Helper()
	fset, mod, pkgs, err := analysis.LoadModule(analysis.LoadConfig{Dir: dir})
	if err != nil {
		t.Fatalf("LoadModule: %v", err)
	}
	diags, err := analysis.Run(fset, mod, pkgs, analysis.Config{Mode: mode})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return diags
}

// diagNames extracts the qualified declaration names from a diagnostic slice.
func diagNames(diags []analysis.Diagnostic) []string {
	names := make([]string, len(diags))
	for i, d := range diags {
		names[i] = d.Name
	}
	return names
}

// diagCodes extracts diagnostic codes indexed by qualified name.
func diagCodes(diags []analysis.Diagnostic) map[string]analysis.DiagnosticCode {
	m := make(map[string]analysis.DiagnosticCode, len(diags))
	for _, d := range diags {
		m[d.Name] = d.Code
	}
	return m
}

func containsName(diags []analysis.Diagnostic, name string) bool {
	for _, d := range diags {
		if d.Name == name {
			return true
		}
	}
	return false
}
