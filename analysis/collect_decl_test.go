package analysis_test

import (
	"testing"

	"github.com/tingtt/internal-unused/analysis"
)

// TestDeclCollection_AllKinds verifies that every supported declaration kind is
// collected from the simple testdata module.
func TestDeclCollection_AllKinds(t *testing.T) {
	dir := testdataDir(t, "simple")
	diags := runAnalysis(t, dir, analysis.ModeAll)

	// These should be unused in ModeAll.
	wantUnused := []string{
		"example.com/simple/internal/pkg.UnusedFunc",
		"example.com/simple/internal/pkg.UnusedVar",
		"example.com/simple/internal/pkg.UnusedConst",
		"example.com/simple/internal/pkg.UnusedType",
		"example.com/simple/internal/pkg.UsedType.UnusedField",
	}
	for _, name := range wantUnused {
		if !containsName(diags, name) {
			t.Errorf("expected diagnostic for %q but not found; got %v", name, diagNames(diags))
		}
	}

	// UnusedType's members must be suppressed (parent is unused).
	suppressedMembers := []string{
		"example.com/simple/internal/pkg.UnusedType.SomeField",
		"example.com/simple/internal/pkg.UnusedType.SomeMethod",
	}
	for _, name := range suppressedMembers {
		if containsName(diags, name) {
			t.Errorf("member diagnostic %q should be suppressed because parent type is unused", name)
		}
	}

	// These should NOT appear as unused.
	wantUsed := []string{
		"example.com/simple/internal/pkg.UsedFunc",
		"example.com/simple/internal/pkg.UsedType",
		"example.com/simple/internal/pkg.UsedVar",
		"example.com/simple/internal/pkg.UsedConst",
		"example.com/simple/internal/pkg.UsedType.UsedField",
	}
	for _, name := range wantUsed {
		if containsName(diags, name) {
			t.Errorf("unexpected diagnostic for used declaration %q", name)
		}
	}
}

// TestDeclCollection_NonExportedSkipped verifies unexported declarations are
// not collected.
func TestDeclCollection_NonExportedSkipped(t *testing.T) {
	dir := testdataDir(t, "simple")
	diags := runAnalysis(t, dir, analysis.ModeAll)

	for _, d := range diags {
		if len(d.Name) > 0 {
			// Extract just the bare declaration name (last segment after ".")
			name := d.Name
			for i := len(name) - 1; i >= 0; i-- {
				if name[i] == '.' {
					name = name[i+1:]
					break
				}
			}
			if len(name) > 0 && name[0] >= 'a' && name[0] <= 'z' {
				t.Errorf("unexported declaration %q should not produce a diagnostic", d.Name)
			}
		}
	}
}

// TestDeclCollection_ParentSuppression verifies that members of an unused type
// do not generate their own diagnostics.
func TestDeclCollection_ParentSuppression(t *testing.T) {
	dir := testdataDir(t, "simple")
	diags := runAnalysis(t, dir, analysis.ModeAll)

	// UnusedType should appear exactly once; its SomeField and SomeMethod must not.
	count := 0
	for _, d := range diags {
		if d.Name == "example.com/simple/internal/pkg.UnusedType" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 diagnostic for UnusedType, got %d", count)
	}
}
