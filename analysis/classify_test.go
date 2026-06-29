package analysis_test

import (
	"testing"

	"github.com/tingtt/internal-unused/analysis"
)

// TestClassify_InternalPackageDetection verifies that only packages under an
// "internal" path component are treated as declaration sources.
func TestClassify_InternalPackageDetection(t *testing.T) {
	dir := testdataDir(t, "simple")
	fset, mod, pkgs, err := analysis.LoadModule(analysis.LoadConfig{Dir: dir})
	if err != nil {
		t.Fatalf("LoadModule: %v", err)
	}
	_ = fset
	declPkgs, refPkgs := analysis.ClassifyPackages(pkgs, mod)

	if len(declPkgs) == 0 {
		t.Error("expected at least one declaration package (internal/pkg)")
	}
	for _, p := range declPkgs {
		if !containsPath(p.PkgPath, "internal") {
			t.Errorf("declaration package %q does not appear to be under 'internal'", p.PkgPath)
		}
	}
	if len(refPkgs) == 0 {
		t.Error("expected at least one reference package")
	}
}

// TestClassify_SyntheticTestBinarySkipped verifies that the synthetic test
// binary main package is not included in either set.
func TestClassify_SyntheticTestBinarySkipped(t *testing.T) {
	dir := testdataDir(t, "multimode")
	_, mod, pkgs, err := analysis.LoadModule(analysis.LoadConfig{Dir: dir})
	if err != nil {
		t.Fatalf("LoadModule: %v", err)
	}
	declPkgs, refPkgs := analysis.ClassifyPackages(pkgs, mod)
	for _, p := range append(declPkgs, refPkgs...) {
		if hasSuffix(p.PkgPath, ".test") && p.Name == "main" {
			t.Errorf("synthetic test binary package %q must not be included", p.ID)
		}
	}
}

// TestClassify_TestFileEnv verifies fileEnv classification.
func TestClassify_TestFileEnv(t *testing.T) {
	cases := []struct {
		filename string
		want     analysis.Env
	}{
		{"foo.go", analysis.EnvProduction},
		{"foo_test.go", analysis.EnvTest},
		{"/path/to/bar_test.go", analysis.EnvTest},
		{"/path/to/bar.go", analysis.EnvProduction},
	}
	for _, tc := range cases {
		got := analysis.FileEnv(tc.filename)
		if got != tc.want {
			t.Errorf("FileEnv(%q) = %v, want %v", tc.filename, got, tc.want)
		}
	}
}

// containsPath is a helper to check if s contains substring sub as a path
// segment.
func containsPath(s, sub string) bool {
	for _, part := range splitPath(s) {
		if part == sub {
			return true
		}
	}
	return false
}

func splitPath(s string) []string {
	var parts []string
	cur := ""
	for _, r := range s {
		if r == '/' || r == '\\' {
			if cur != "" {
				parts = append(parts, cur)
				cur = ""
			}
		} else {
			cur += string(r)
		}
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return parts
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
