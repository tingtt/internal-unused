// Package pkg is an internal package with various exported declarations.
package pkg

// UsedFunc is called by the consumer.
func UsedFunc() {}

// UnusedFunc is never called.
func UnusedFunc() {}

// UsedType is referenced by the consumer.
type UsedType struct {
	UsedField   string
	UnusedField int
}

// ExportedBase is embedded by other exported types.
type ExportedBase struct{}

// UnusedEmbeddedFieldOwner declares an exported embedded field that is never
// selected by consumers.
type UnusedEmbeddedFieldOwner struct {
	ExportedBase
}

// PromotionA starts a multi-level promoted method path.
type PromotionA struct {
	PromotionB
}

// PromotionB is the intermediate embedded field in a promoted method path.
type PromotionB struct {
	PromotionC
}

// PromotionC provides a promoted method.
type PromotionC struct{}

// Run is selected through PromotionA.
func (PromotionC) Run() {}

// UnusedType and its members should produce a single diagnostic for the type.
type UnusedType struct {
	SomeField string
}

func (UnusedType) SomeMethod() {}

// UsedVar is read by the consumer.
var UsedVar = 42

// UnusedVar is never referenced.
var UnusedVar = 99

// UsedConst is referenced by the consumer.
const UsedConst = "hello"

// UnusedConst is never referenced.
const UnusedConst = "world"
