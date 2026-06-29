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
