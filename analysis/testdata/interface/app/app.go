package app

import (
	"example.com/iface/internal/svc"
)

// runWith uses the Runner interface, triggering dispatch propagation.
func runWith(r svc.Runner) {
	_ = r.Run()
}

func main() {
	runWith(svc.Impl{}) // Impl is assigned to Runner → Run() is propagated
}
