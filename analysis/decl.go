package analysis

import (
	"fmt"
	"go/token"
	"go/types"
)

// DeclKind is the kind of exported declaration.
type DeclKind int

const (
	KindFunction  DeclKind = iota // package-level function
	KindMethod                    // method on a named type
	KindType                      // named type (including type aliases)
	KindVar                       // package-level variable
	KindConst                     // package-level constant
	KindField                     // exported struct field
	KindIfaceMethod               // explicit interface method
)

func (k DeclKind) String() string {
	switch k {
	case KindFunction:
		return "function"
	case KindMethod:
		return "method"
	case KindType:
		return "type"
	case KindVar:
		return "variable"
	case KindConst:
		return "constant"
	case KindField:
		return "field"
	case KindIfaceMethod:
		return "interface method"
	default:
		return "unknown"
	}
}

// Declaration represents an exported declaration from an internal package.
type Declaration struct {
	Object types.Object   // the resolved types.Object
	Kind   DeclKind
	Owner  *types.TypeName // non-nil for Method, Field, IfaceMethod
	Pos    token.Position
}

// QualifiedName returns the human-readable qualified name.
// Examples: "pkg/path.FuncName", "pkg/path.TypeName.MethodName".
func (d *Declaration) QualifiedName() string {
	pkg := d.Object.Pkg()
	name := d.Object.Name()
	switch d.Kind {
	case KindMethod, KindField, KindIfaceMethod:
		if d.Owner != nil {
			return fmt.Sprintf("%s.%s.%s", pkg.Path(), d.Owner.Name(), name)
		}
	}
	return fmt.Sprintf("%s.%s", pkg.Path(), name)
}
