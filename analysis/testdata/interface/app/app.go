package app

import (
	"io"

	"example.com/iface/internal/svc"
)

// runWith uses the Runner interface, triggering dispatch propagation.
func runWith(r svc.Runner) {
	_ = r.Run()
}

func main() {
	runWith(svc.Impl{}) // Impl is assigned to Runner → Run() is propagated
	_ = svc.NewRunner().Run()

	var w io.Writer = &svc.Writer{}
	_, _ = w.Write(nil)

	var err error = svc.CustomError{}
	_ = err.Error()

	converted := svc.Runner(svc.ConvertedImpl{})
	_ = converted.Run()
}
