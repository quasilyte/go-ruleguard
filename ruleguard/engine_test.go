package ruleguard

import (
	"fmt"
	"go/types"
	"testing"
)

func TestEngineAddCachedPackage(t *testing.T) {
	state := newEngineState()

	// 1. Handling nil package safely.
	state.AddCachedPackage("nilpkg", nil)
	if pkg := state.GetCachedPackage("nilpkg"); pkg != nil {
		t.Fatalf("expected nil package for nilpkg, got %v", pkg)
	}

	// 2. Diamond dependency graph:
	//     pkgA
	//    /    \
	//  pkgB   pkgC
	//    \    /
	//     pkgD
	//      |
	//     pkgE
	pkgE := types.NewPackage("pkgE", "pkgE")
	pkgE.MarkComplete()

	pkgD := types.NewPackage("pkgD", "pkgD")
	pkgD.SetImports([]*types.Package{pkgE, nil}) // contains nil import defensively
	pkgD.MarkComplete()

	pkgB := types.NewPackage("pkgB", "pkgB")
	pkgB.SetImports([]*types.Package{pkgD})
	pkgB.MarkComplete()

	pkgC := types.NewPackage("pkgC", "pkgC")
	pkgC.SetImports([]*types.Package{pkgD})
	pkgC.MarkComplete()

	pkgA := types.NewPackage("pkgA", "pkgA")
	pkgA.SetImports([]*types.Package{pkgB, pkgC})
	pkgA.MarkComplete()

	state.AddCachedPackage("pkgA", pkgA)

	for _, name := range []string{"pkgA", "pkgB", "pkgC", "pkgD", "pkgE"} {
		if pkg := state.GetCachedPackage(name); pkg == nil {
			t.Errorf("expected %s to be in cache", name)
		}
	}

	// 3. Adding already-cached package again should be a no-op.
	state.AddCachedPackage("pkgA", pkgA)

	// 4. Incomplete dependencies should not be cached.
	pkgIncompleteDep := types.NewPackage("pkgIncompleteDep", "pkgIncompleteDep")
	// not complete

	pkgWithIncomplete := types.NewPackage("pkgWithIncomplete", "pkgWithIncomplete")
	pkgWithIncomplete.SetImports([]*types.Package{pkgIncompleteDep})
	pkgWithIncomplete.MarkComplete()

	state.AddCachedPackage("pkgWithIncomplete", pkgWithIncomplete)
	if pkg := state.GetCachedPackage("pkgWithIncomplete"); pkg == nil {
		t.Fatalf("expected pkgWithIncomplete to be cached")
	}
	if pkg := state.GetCachedPackage("pkgIncompleteDep"); pkg != nil {
		t.Fatalf("expected incomplete dependency pkgIncompleteDep NOT to be cached")
	}
}

func TestEngineAddCachedPackageRedundantPerformance(t *testing.T) {
	// Construct a DAG with 15 layers where each node connects to both nodes in the next layer.
	// Total paths = 2^14 = 16384 paths.
	// Without deduplication, this takes exponential visits; with memoization, it completes instantly.
	const layers = 15
	lastPkg := types.NewPackage("node_leaf", "node_leaf")
	lastPkg.MarkComplete()
	prevLayer := []*types.Package{lastPkg, lastPkg}

	for i := layers - 1; i >= 0; i-- {
		p1 := types.NewPackage(fmt.Sprintf("node_%d_1", i), "p")
		p1.SetImports(prevLayer)
		p1.MarkComplete()

		p2 := types.NewPackage(fmt.Sprintf("node_%d_2", i), "p")
		p2.SetImports(prevLayer)
		p2.MarkComplete()

		prevLayer = []*types.Package{p1, p2}
	}

	root := types.NewPackage("root", "root")
	root.SetImports(prevLayer)
	root.MarkComplete()

	state := newEngineState()
	state.AddCachedPackage("root", root)

	if pkg := state.GetCachedPackage("node_leaf"); pkg == nil {
		t.Fatalf("expected leaf to be cached")
	}
}
