package analysis_test

import (
	"testing"

	"github.com/tingtt/internal-unused/analysis"
)

// TestRefCollection_DirectFunctionCall verifies that a called function is not
// reported as unused.
func TestRefCollection_DirectFunctionCall(t *testing.T) {
	dir := testdataDir(t, "simple")
	diags := runAnalysis(t, dir, analysis.ModeAll)
	if containsName(diags, "example.com/simple/internal/pkg.UsedFunc") {
		t.Error("UsedFunc is called; it must not be in diagnostics")
	}
}

// TestRefCollection_StructFieldUsed verifies that a field used in a keyed
// composite literal is not reported.
func TestRefCollection_StructFieldUsed(t *testing.T) {
	dir := testdataDir(t, "simple")
	diags := runAnalysis(t, dir, analysis.ModeAll)
	if containsName(diags, "example.com/simple/internal/pkg.UsedType.UsedField") {
		t.Error("UsedField is used in a composite literal; it must not be in diagnostics")
	}
}

// TestRefCollection_UnusedFieldReported verifies that a field never referenced
// produces a diagnostic.
func TestRefCollection_UnusedFieldReported(t *testing.T) {
	dir := testdataDir(t, "simple")
	diags := runAnalysis(t, dir, analysis.ModeAll)
	if !containsName(diags, "example.com/simple/internal/pkg.UsedType.UnusedField") {
		t.Errorf("UnusedField is never referenced; expected a diagnostic; got %v", diagNames(diags))
	}
}

// TestRefCollection_InterfaceDispatch verifies that a concrete method is marked
// used when it is invoked via interface dispatch (Impl assigned to Runner,
// Run() called on Runner).
func TestRefCollection_InterfaceDispatch(t *testing.T) {
	dir := testdataDir(t, "interface")
	diags := runAnalysis(t, dir, analysis.ModeAll)

	// Impl.Run() is dispatched through Runner.Run(); must not be reported.
	if containsName(diags, "example.com/iface/internal/svc.Impl.Run") {
		t.Error("Impl.Run is used via interface dispatch; must not be in diagnostics")
	}
	// Runner.Run() itself is called directly; must not be reported.
	if containsName(diags, "example.com/iface/internal/svc.Runner.Run") {
		t.Error("Runner.Run is called; must not be in diagnostics")
	}
}

// TestRefCollection_NoInterfaceDispatchWithoutAssignment verifies that a type
// whose methods are not invoked through an interface is not treated as used
// via interface dispatch.
func TestRefCollection_NoInterfaceDispatchWithoutAssignment(t *testing.T) {
	dir := testdataDir(t, "interface")
	diags := runAnalysis(t, dir, analysis.ModeAll)

	// NotUsedAtAll is never assigned to Runner and never directly referenced.
	// Its Run and Unused methods should be reported (or the type itself if
	// parent suppression applies; either way the type must appear).
	typeOrMethod := containsName(diags, "example.com/iface/internal/svc.NotUsedAtAll") ||
		containsName(diags, "example.com/iface/internal/svc.NotUsedAtAll.Run")
	if !typeOrMethod {
		t.Errorf("NotUsedAtAll and its methods are not used; expected a diagnostic; got %v", diagNames(diags))
	}
}
