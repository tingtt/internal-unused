package analysis

import (
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ModuleInfo holds the identity of the target Go module.
type ModuleInfo struct {
	Root string // absolute filesystem path to the module root directory
	Path string // module import path declared in go.mod
}

// LoadConfig configures the package loading step.
type LoadConfig struct {
	// Dir is the working directory for the load. Defaults to ".".
	Dir string
	// Patterns are the package patterns to load (e.g. "./...").
	// Defaults to ["./..."] when empty.
	Patterns []string
	// Env contains extra environment variables forwarded to the go toolchain.
	Env []string
}

// LoadModule loads all packages (including test variants) in the module and
// returns a shared FileSet, module metadata, and the loaded packages.
//
// An error is returned if loading fails or if any package contains errors
// (e.g. syntax errors or unresolvable imports), because partial analysis
// would risk producing incorrect unused diagnostics.
func LoadModule(cfg LoadConfig) (*token.FileSet, ModuleInfo, []*packages.Package, error) {
	if cfg.Dir == "" {
		cfg.Dir = "."
	}
	if len(cfg.Patterns) == 0 {
		cfg.Patterns = []string{"./..."}
	}

	fset := token.NewFileSet()

	pcfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedModule,
		Dir:   cfg.Dir,
		Tests: true,
		Fset:  fset,
		Env:   cfg.Env,
	}

	pkgs, err := packages.Load(pcfg, cfg.Patterns...)
	if err != nil {
		return nil, ModuleInfo{}, nil, fmt.Errorf("loading packages: %w", err)
	}

	// Collect load errors across all packages.
	var errMsgs []string
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, e := range pkg.Errors {
			errMsgs = append(errMsgs, e.Error())
		}
	})
	if len(errMsgs) > 0 {
		return nil, ModuleInfo{}, nil, fmt.Errorf(
			"package load errors (analysis aborted to avoid incomplete results):\n%s",
			strings.Join(errMsgs, "\n"),
		)
	}

	// Resolve module info from the first package that exposes it.
	var mod ModuleInfo
	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		if mod.Root == "" && pkg.Module != nil {
			mod.Root = pkg.Module.Dir
			mod.Path = pkg.Module.Path
		}
		return mod.Root == ""
	}, nil)
	if mod.Root == "" {
		return nil, ModuleInfo{}, nil, fmt.Errorf(
			"could not determine module root; ensure the target directory is inside a Go module",
		)
	}

	return fset, mod, pkgs, nil
}
