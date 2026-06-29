package analysis_test

import (
	"testing"

	"github.com/tingtt/internal-unused/analysis"
)

// TestEvaluate_AllMode_TestOnlyCountsAsUsed verifies that in ModeAll a
// declaration referenced only from test code is treated as used (no diagnostic).
func TestEvaluate_AllMode_TestOnlyCountsAsUsed(t *testing.T) {
	dir := testdataDir(t, "multimode")
	diags := runAnalysis(t, dir, analysis.ModeAll)
	if containsName(diags, "example.com/multimode/internal/api.TestOnlyFunc") {
		t.Error("in ModeAll, TestOnlyFunc (used by tests) must not be reported")
	}
}

// TestEvaluate_AllMode_CompletelyUnusedReported verifies that in ModeAll a
// declaration unused in both environments is reported with code "unused".
func TestEvaluate_AllMode_CompletelyUnusedReported(t *testing.T) {
	dir := testdataDir(t, "multimode")
	diags := runAnalysis(t, dir, analysis.ModeAll)
	codes := diagCodes(diags)
	code, found := codes["example.com/multimode/internal/api.CompletelyUnused"]
	if !found {
		t.Errorf("CompletelyUnused must be reported in ModeAll; got %v", diagNames(diags))
		return
	}
	if code != analysis.CodeUnused {
		t.Errorf("CompletelyUnused code = %q, want %q", code, analysis.CodeUnused)
	}
}

// TestEvaluate_ProductionMode_TestOnlyReportedAsTestOnly verifies that in
// ModeProduction, a declaration used only by tests gets code "test-only".
func TestEvaluate_ProductionMode_TestOnlyReportedAsTestOnly(t *testing.T) {
	dir := testdataDir(t, "multimode")
	diags := runAnalysis(t, dir, analysis.ModeProduction)
	codes := diagCodes(diags)
	code, found := codes["example.com/multimode/internal/api.TestOnlyFunc"]
	if !found {
		t.Errorf("TestOnlyFunc must be reported in ModeProduction; got %v", diagNames(diags))
		return
	}
	if code != analysis.CodeTestOnly {
		t.Errorf("TestOnlyFunc code = %q, want %q", code, analysis.CodeTestOnly)
	}
}

// TestEvaluate_ProductionMode_CompletelyUnusedStillUnused verifies that in
// ModeProduction a completely unused declaration keeps code "unused".
func TestEvaluate_ProductionMode_CompletelyUnusedStillUnused(t *testing.T) {
	dir := testdataDir(t, "multimode")
	diags := runAnalysis(t, dir, analysis.ModeProduction)
	codes := diagCodes(diags)
	code, found := codes["example.com/multimode/internal/api.CompletelyUnused"]
	if !found {
		t.Errorf("CompletelyUnused must be reported in ModeProduction; got %v", diagNames(diags))
		return
	}
	if code != analysis.CodeUnused {
		t.Errorf("CompletelyUnused code = %q, want %q", code, analysis.CodeUnused)
	}
}

// TestEvaluate_ProductionMode_ProdUsedNotReported verifies that a declaration
// used in production code is not reported in either mode.
func TestEvaluate_ProductionMode_ProdUsedNotReported(t *testing.T) {
	for _, mode := range []analysis.Mode{analysis.ModeAll, analysis.ModeProduction} {
		dir := testdataDir(t, "multimode")
		diags := runAnalysis(t, dir, mode)
		if containsName(diags, "example.com/multimode/internal/api.ProdUsed") {
			t.Errorf("mode=%v: ProdUsed is used in production; must not be reported", mode)
		}
	}
}

// TestEvaluate_NoDiagnosticsWhenAllUsed verifies that the noerror fixture
// produces zero diagnostics.
func TestEvaluate_NoDiagnosticsWhenAllUsed(t *testing.T) {
	dir := testdataDir(t, "noerror")
	for _, mode := range []analysis.Mode{analysis.ModeAll, analysis.ModeProduction} {
		diags := runAnalysis(t, dir, mode)
		if len(diags) != 0 {
			t.Errorf("mode=%v: expected no diagnostics, got %v", mode, diagNames(diags))
		}
	}
}

// TestEvaluate_DiagnosticOrder verifies that diagnostics are returned in
// stable order regardless of map iteration.
func TestEvaluate_DiagnosticOrder(t *testing.T) {
	dir := testdataDir(t, "simple")
	d1 := runAnalysis(t, dir, analysis.ModeAll)
	d2 := runAnalysis(t, dir, analysis.ModeAll)

	if len(d1) != len(d2) {
		t.Fatalf("inconsistent diagnostic counts between runs: %d vs %d", len(d1), len(d2))
	}
	for i := range d1 {
		if d1[i].Name != d2[i].Name || d1[i].Pos.Line != d2[i].Pos.Line {
			t.Errorf("diagnostic[%d] differs between runs: %+v vs %+v", i, d1[i], d2[i])
		}
	}
}

// TestEvaluate_UnusedTypeSuppressesMembers verifies parent-type suppression:
// members of an unused type do not generate independent diagnostics.
func TestEvaluate_UnusedTypeSuppressesMembers(t *testing.T) {
	dir := testdataDir(t, "multimode")
	diags := runAnalysis(t, dir, analysis.ModeAll)

	// UnusedMemberType is used (instantiated in consumer), so its unused
	// members should each generate their own diagnostics.
	if !containsName(diags, "example.com/multimode/internal/api.UnusedMemberType.UnusedMemberField") {
		t.Errorf("UnusedMemberField on a used type must be reported; got %v", diagNames(diags))
	}
	if !containsName(diags, "example.com/multimode/internal/api.UnusedMemberType.UnusedMemberMethod") {
		t.Errorf("UnusedMemberMethod on a used type must be reported; got %v", diagNames(diags))
	}
}
