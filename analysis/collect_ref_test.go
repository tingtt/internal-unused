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

// TestRefCollection_InterfaceDispatchMarksMethodsUsed verifies the maintained
// interface-dispatch boundaries that must not produce unused-method reports.
func TestRefCollection_InterfaceDispatchMarksMethodsUsed(t *testing.T) {
	dir := testdataDir(t, "interface")
	diags := runAnalysis(t, dir, analysis.ModeAll)

	cases := []struct {
		name string
		decl string
	}{
		{
			name: "parameter conversion to internal interface",
			decl: "example.com/iface/internal/svc.Impl.Run",
		},
		{
			name: "selected internal interface method",
			decl: "example.com/iface/internal/svc.Runner.Run",
		},
		{
			name: "return conversion to internal interface",
			decl: "example.com/iface/internal/svc.ReturnedImpl.Run",
		},
		{
			name: "external package interface",
			decl: "example.com/iface/internal/svc.Writer.Write",
		},
		{
			name: "predeclared interface",
			decl: "example.com/iface/internal/svc.CustomError.Error",
		},
		{
			name: "production assignment used from tests",
			decl: "example.com/iface/internal/svc.ProductionAssignedImpl.Run",
		},
		{
			name: "test package variant interface identity",
			decl: "example.com/iface/internal/svc.TestVariantImpl.Run",
		},
		{
			name: "explicit interface conversion",
			decl: "example.com/iface/internal/svc.ConvertedImpl.Run",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if containsName(diags, tc.decl) {
				t.Fatalf("%s is used via interface dispatch; must not be in diagnostics", tc.decl)
			}
		})
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
