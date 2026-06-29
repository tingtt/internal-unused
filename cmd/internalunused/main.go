// Command internalunused detects unused exported declarations in the internal
// packages of a Go module.
//
// Usage:
//
//	internalunused [flags] [package patterns]
//
// When no package patterns are given, "./..." (the entire module) is used.
//
// Flags:
//
//	-dir string    Root directory of the module to analyse (default: current directory).
//	-mode string   Detection mode: "all" (default) or "production".
//	               all:        report declarations unused by both production
//	                           and test code.
//	               production: additionally report declarations that are only
//	                           used by tests.
//
// Exit codes:
//
//	0   No unused declarations found.
//	1   One or more unused declarations were detected (lint failure).
//	2   The analysis could not be completed (execution failure).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tingtt/internal-unused/analysis"
)

func main() {
	os.Exit(run())
}

func run() int {
	dirFlag := flag.String("dir", "", "module root directory to analyse (default: current directory)")
	modeFlag := flag.String("mode", "all", `detection mode: "all" or "production"`)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: internalunused [flags] [package patterns]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	mode, ok := analysis.ParseMode(*modeFlag)
	if !ok {
		fmt.Fprintf(os.Stderr, "internalunused: invalid -mode %q (want: all, production)\n", *modeFlag)
		return 2
	}

	patterns := flag.Args()

	fset, mod, pkgs, err := analysis.LoadModule(analysis.LoadConfig{
		Dir:      *dirFlag,
		Patterns: patterns,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "internalunused: %v\n", err)
		return 2
	}

	diagnostics, err := analysis.Run(fset, mod, pkgs, analysis.Config{Mode: mode})
	if err != nil {
		fmt.Fprintf(os.Stderr, "internalunused: analysis failed: %v\n", err)
		return 2
	}

	for _, d := range diagnostics {
		fmt.Println(analysis.FormatDiagnostic(d))
	}

	if len(diagnostics) > 0 {
		return 1
	}
	return 0
}
