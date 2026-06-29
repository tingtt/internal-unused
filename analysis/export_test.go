// export_test.go exposes internal symbols for white-box testing.
// This file is only compiled during testing.
package analysis

import "golang.org/x/tools/go/packages"

// ClassifyPackages is the exported test alias for classifyPackages.
func ClassifyPackages(pkgs []*packages.Package, mod ModuleInfo) (declPkgs, refPkgs []*packages.Package) {
	return classifyPackages(pkgs, mod)
}

// FileEnv is the exported test alias for fileEnv.
func FileEnv(filename string) Env {
	return fileEnv(filename)
}
