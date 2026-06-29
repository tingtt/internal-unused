package api

// ProdUsed is used in production code.
func ProdUsed() {}

// TestOnlyFunc is referenced only from a test file.
func TestOnlyFunc() {}

// CompletelyUnused is never referenced anywhere.
func CompletelyUnused() {}

// UsedType is referenced by the consumer.
type UsedType struct {
	Active bool
}

// UnusedMemberType is used, but its field is not.
type UnusedMemberType struct {
	UnusedMemberField string
}

func (UnusedMemberType) UnusedMemberMethod() {}
