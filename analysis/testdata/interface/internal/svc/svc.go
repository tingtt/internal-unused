package svc

import "io"

// Runner is an interface declared in an internal package.
type Runner interface {
	Run() error
	// Unused is never called through this interface.
	Unused() string
}

// Impl is a concrete implementation of Runner.
type Impl struct{}

// Run satisfies Runner.Run and should be marked used via interface dispatch.
func (Impl) Run() error { return nil }

// Unused satisfies Runner.Unused but is never called through the interface.
func (Impl) Unused() string { return "" }

// ReturnedImpl is only exposed through a Runner return value.
type ReturnedImpl struct{}

func (ReturnedImpl) Run() error     { return nil }
func (ReturnedImpl) Unused() string { return "" }

// ProductionAssignedImpl is assigned to Runner in production code and called
// through that interface from tests.
type ProductionAssignedImpl struct{}

func (ProductionAssignedImpl) Run() error     { return nil }
func (ProductionAssignedImpl) Unused() string { return "" }

// TestVariantImpl is assigned to Runner only from test code.
type TestVariantImpl struct{}

func (TestVariantImpl) Run() error     { return nil }
func (TestVariantImpl) Unused() string { return "" }

// Writer implements an external package interface.
type Writer struct{}

func (*Writer) Write(p []byte) (int, error) { return len(p), nil }

// CustomError implements the predeclared error interface.
type CustomError struct{}

func (CustomError) Error() string { return "" }

// NewRunner returns a concrete implementation as an interface.
func NewRunner() Runner {
	return ReturnedImpl{}
}

// DefaultRunner is built in production and exercised from tests.
var DefaultRunner Runner = ProductionAssignedImpl{}

var _ io.Writer = (*Writer)(nil)

// NotUsedAtAll is a concrete type that is never assigned to Runner.
type NotUsedAtAll struct{}

func (NotUsedAtAll) Run() error     { return nil }
func (NotUsedAtAll) Unused() string { return "" }
