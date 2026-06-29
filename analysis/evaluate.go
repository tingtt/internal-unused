package analysis

import (
	"go/types"
	"sort"
)

// evaluate applies the detection policy to every collected declaration and
// returns the sorted list of diagnostics.
//
// Parent diagnostic suppression: when a named type is itself unused, its
// methods, struct fields, and interface methods are suppressed — removing the
// type eliminates them implicitly.  The internal usage state for children is
// still retained for future detailed-output modes.
func evaluate(declIdx *DeclIndex, usage *UsageIndex, mode Mode) []Diagnostic {
	// First pass: identify unused named types (potential parents to suppress).
	unusedTypeDecls := make(map[*types.TypeName]bool)
	for _, decl := range declIdx.all() {
		if decl.Kind != KindType {
			continue
		}
		state := usage.Get(decl.Object)
		if isDeclUnused(state, mode) {
			unusedTypeDecls[decl.Object.(*types.TypeName)] = true
		}
	}

	// Second pass: build diagnostics, applying parent suppression.
	var diagnostics []Diagnostic
	for _, decl := range declIdx.all() {
		// Suppress member diagnostics when the owning type is already unused.
		if decl.Owner != nil && unusedTypeDecls[decl.Owner] {
			continue
		}

		state := usage.Get(decl.Object)

		switch mode {
		case ModeAll:
			if !state.Production && !state.Test {
				diagnostics = append(diagnostics, Diagnostic{
					Pos:  decl.Pos,
					Kind: decl.Kind,
					Name: decl.QualifiedName(),
					Code: CodeUnused,
				})
			}
		case ModeProduction:
			if !state.Production {
				code := CodeUnused
				if state.Test {
					code = CodeTestOnly
				}
				diagnostics = append(diagnostics, Diagnostic{
					Pos:  decl.Pos,
					Kind: decl.Kind,
					Name: decl.QualifiedName(),
					Code: code,
				})
			}
		}
	}

	// Sort for stable output regardless of map iteration order or load order.
	sort.Slice(diagnostics, func(i, j int) bool {
		a, b := diagnostics[i], diagnostics[j]
		if a.Pos.Filename != b.Pos.Filename {
			return a.Pos.Filename < b.Pos.Filename
		}
		if a.Pos.Line != b.Pos.Line {
			return a.Pos.Line < b.Pos.Line
		}
		if a.Pos.Column != b.Pos.Column {
			return a.Pos.Column < b.Pos.Column
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Name < b.Name
	})
	return diagnostics
}

// isDeclUnused returns whether the declaration is considered unused for the
// given mode — used to determine parent-suppression eligibility for types.
func isDeclUnused(state UsageState, mode Mode) bool {
	switch mode {
	case ModeAll:
		return !state.Production && !state.Test
	case ModeProduction:
		return !state.Production
	}
	return false
}
