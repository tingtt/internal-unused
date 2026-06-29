package svc

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

// NotUsedAtAll is a concrete type that is never assigned to Runner.
type NotUsedAtAll struct{}

func (NotUsedAtAll) Run() error  { return nil }
func (NotUsedAtAll) Unused() string { return "" }
