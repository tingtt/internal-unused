package analysis

import (
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// classifyPackages separates packages into two groups:
//
//   - declPkgs: non-test packages under an internal directory, used for
//     declaration collection.
//   - refPkgs: all module packages (including test variants), used for
//     reference collection.
//
// Synthetic test-binary packages (e.g. "github.com/foo/bar.test") are skipped.
func classifyPackages(pkgs []*packages.Package, mod ModuleInfo) (declPkgs, refPkgs []*packages.Package) {
	seenDecl := make(map[string]bool)
	seenRef := make(map[string]bool)

	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		if isSyntheticTestBinary(pkg) {
			return false
		}
		if !isModulePackage(pkg, mod) {
			return true // continue visiting dependencies but don't include them
		}

		if !seenRef[pkg.ID] {
			seenRef[pkg.ID] = true
			refPkgs = append(refPkgs, pkg)
		}

		if pkg.ForTest == "" && isInternalPackage(pkg, mod) {
			if !seenDecl[pkg.ID] {
				seenDecl[pkg.ID] = true
				declPkgs = append(declPkgs, pkg)
			}
		}
		return true
	}, nil)

	return
}

// isSyntheticTestBinary returns true for the synthetic test-binary main
// package that the Go toolchain generates internally (e.g. "foo.test").
// These packages contain no user-declared source and must not be analyzed.
func isSyntheticTestBinary(pkg *packages.Package) bool {
	return strings.HasSuffix(pkg.PkgPath, ".test") && pkg.Name == "main"
}

// isModulePackage returns true if pkg belongs to the target module.
func isModulePackage(pkg *packages.Package, mod ModuleInfo) bool {
	if pkg.Module != nil {
		return pkg.Module.Path == mod.Path
	}
	// Fallback for cases where Module metadata is unavailable.
	return pkg.PkgPath == mod.Path ||
		strings.HasPrefix(pkg.PkgPath, mod.Path+"/")
}

// isInternalPackage returns true if the package resides under a directory
// component named exactly "internal" within the module root.
//
// This checks the actual directory structure, not just the import path string,
// so "internalized" directories are not mistaken for "internal".
func isInternalPackage(pkg *packages.Package, mod ModuleInfo) bool {
	if len(pkg.GoFiles) == 0 {
		return false
	}
	dir := filepath.Dir(pkg.GoFiles[0])
	rel, err := filepath.Rel(mod.Root, dir)
	if err != nil {
		return false
	}
	// filepath.Rel may return paths starting with ".." for paths outside root.
	if strings.HasPrefix(rel, "..") {
		return false
	}
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if part == "internal" {
			return true
		}
	}
	return false
}

// fileEnv classifies a source file as production or test based on its name.
func fileEnv(filename string) Env {
	if strings.HasSuffix(filename, "_test.go") {
		return EnvTest
	}
	return EnvProduction
}
