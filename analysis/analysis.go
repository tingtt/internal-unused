// Package analysis provides the core logic for detecting unused exported
// declarations in internal packages of a Go module.
//
// The analysis pipeline is:
//
//  1. LoadModule   – load all module packages (including test variants) via
//     golang.org/x/tools/go/packages.
//  2. Run          – classify packages, collect declarations, collect
//     references, propagate interface dispatch, evaluate usage, and
//     return sorted Diagnostics.
//
// The CLI adapter (cmd/internalunused) and any future golangci-lint adapter
// call LoadModule and Run; they do not contain analysis logic themselves.
package analysis

import (
	"go/token"

	"golang.org/x/tools/go/packages"
)

// Config holds the settings for a single analysis run.
type Config struct {
	Mode Mode
}

// Run performs the full analysis on the already-loaded packages and returns
// the sorted list of diagnostics.  It returns no error unless the analysis
// itself encounters an unexpected state (package load errors are surfaced by
// LoadModule before Run is called).
func Run(fset *token.FileSet, mod ModuleInfo, pkgs []*packages.Package, cfg Config) ([]Diagnostic, error) {
	// 1. Classify packages.
	declPkgs, refPkgs := classifyPackages(pkgs, mod)

	// 2. Collect exported declarations from internal packages.
	declIdx := collectDeclarations(declPkgs, fset)

	// 3. Collect references from all module packages (production + test).
	usage := newUsageIndex()
	ifaceAssigns := collectReferences(refPkgs, fset, declIdx, usage)

	// 4. Propagate interface-dispatch usage to concrete methods.
	propagateInterfaceDispatch(declIdx, usage, ifaceAssigns)

	// 5. Evaluate usage state and build sorted diagnostics.
	diagnostics := evaluate(declIdx, usage, cfg.Mode)

	return diagnostics, nil
}
