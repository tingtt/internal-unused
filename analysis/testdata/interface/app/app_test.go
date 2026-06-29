package app

import (
	"testing"

	"example.com/iface/internal/svc"
)

func TestDefaultRunner(t *testing.T) {
	if err := svc.DefaultRunner.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestVariantInterfaceDispatch(t *testing.T) {
	var r svc.Runner = svc.TestVariantImpl{}
	if err := r.Run(); err != nil {
		t.Fatal(err)
	}
}
