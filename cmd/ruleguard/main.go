package main

import (
	"go/build"
	"log"
	"os/exec"
	"strings"

	"github.com/quasilyte/go-ruleguard/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	// If we don't do this, release binaries will have GOROOT set
	// to the `go env GOROOT` of the machine that built them.
	//
	// Usually, it doesn't matter, but since we're using "source"
	// importers, it *will* use build.Default.GOROOT to locate packages.
	//
	// Example: release binary was built with GOROOT="/foo/bar/go",
	// user has GOROOT at "/usr/local/go"; if we don't adjust GOROOT
	// field here, it'll be "/foo/bar/go".
	build.Default.GOROOT = hostGOROOT()

	singlechecker.Main(analyzer.Analyzer)
}

func hostGOROOT() string {
	// `go env GOROOT` should return the correct value even
	// if it was overwritten by explicit GOROOT env var.
	out, err := exec.Command("go", "env", "GOROOT").CombinedOutput()
	if err != nil {
		log.Fatalf("infer GOROOT: %v: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}
