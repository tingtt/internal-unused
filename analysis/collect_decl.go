package analysis

import (
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// DeclIndex holds all collected exported declarations from internal packages.
// It supports two lookup strategies:
//  1. By types.Object pointer (fast path, always works within the same compilation).
//  2. By token.Pos (fallback for test-compiled variants that re-compile the same
//     source files: within one packages.Load session each source file is added to
//     the FileSet exactly once, so the same declaration always has the same Pos).
type DeclIndex struct {
	byObject map[types.Object]*Declaration
	byPos    map[token.Pos]*Declaration
}

func newDeclIndex() *DeclIndex {
	return &DeclIndex{
		byObject: make(map[types.Object]*Declaration),
		byPos:    make(map[token.Pos]*Declaration),
	}
}

func (idx *DeclIndex) add(d *Declaration) {
	idx.byObject[d.Object] = d
	idx.byPos[d.Object.Pos()] = d
}

// get looks up a declaration by types.Object, falling back to position lookup.
func (idx *DeclIndex) get(obj types.Object) (*Declaration, bool) {
	if d, ok := idx.byObject[obj]; ok {
		return d, true
	}
	if d, ok := idx.byPos[obj.Pos()]; ok {
		return d, true
	}
	return nil, false
}

// all returns every collected declaration in an unspecified order.
func (idx *DeclIndex) all() []*Declaration {
	result := make([]*Declaration, 0, len(idx.byObject))
	for _, d := range idx.byObject {
		result = append(result, d)
	}
	return result
}

// collectDeclarations walks the given internal packages and collects every
// exported declaration into a DeclIndex.
func collectDeclarations(pkgs []*packages.Package, fset *token.FileSet) *DeclIndex {
	idx := newDeclIndex()
	for _, pkg := range pkgs {
		collectPackageDecls(pkg, fset, idx)
	}
	return idx
}

func collectPackageDecls(pkg *packages.Package, fset *token.FileSet, idx *DeclIndex) {
	if pkg.Types == nil {
		return
	}
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if !obj.Exported() {
			continue
		}
		pos := fset.Position(obj.Pos())
		switch o := obj.(type) {
		case *types.Func:
			sig := o.Type().(*types.Signature)
			if sig.Recv() == nil {
				idx.add(&Declaration{Object: o, Kind: KindFunction, Pos: pos})
			}
			// Methods are collected via the named type below.
		case *types.TypeName:
			idx.add(&Declaration{Object: o, Kind: KindType, Pos: pos})
			collectTypeMembers(o, fset, idx)
		case *types.Var:
			idx.add(&Declaration{Object: o, Kind: KindVar, Pos: pos})
		case *types.Const:
			idx.add(&Declaration{Object: o, Kind: KindConst, Pos: pos})
		}
	}
}

// collectTypeMembers collects methods, struct fields, and interface methods
// for a named type.  Promoted members are NOT collected here; only members
// directly declared on typeName are considered new declarations.
func collectTypeMembers(typeName *types.TypeName, fset *token.FileSet, idx *DeclIndex) {
	named, ok := typeName.Type().(*types.Named)
	if !ok {
		return
	}

	// Methods declared with this named type as receiver.
	for i := range named.NumMethods() {
		m := named.Method(i)
		if m.Exported() {
			idx.add(&Declaration{
				Object: m,
				Kind:   KindMethod,
				Owner:  typeName,
				Pos:    fset.Position(m.Pos()),
			})
		}
	}

	// Underlying type members.
	switch u := named.Underlying().(type) {
	case *types.Struct:
		for i := range u.NumFields() {
			f := u.Field(i)
			if f.Exported() {
				idx.add(&Declaration{
					Object: f,
					Kind:   KindField,
					Owner:  typeName,
					Pos:    fset.Position(f.Pos()),
				})
			}
		}
	case *types.Interface:
		// Only explicit methods, not methods promoted by embedded interfaces.
		for i := range u.NumExplicitMethods() {
			m := u.ExplicitMethod(i)
			if m.Exported() {
				idx.add(&Declaration{
					Object: m,
					Kind:   KindIfaceMethod,
					Owner:  typeName,
					Pos:    fset.Position(m.Pos()),
				})
			}
		}
	}
}
