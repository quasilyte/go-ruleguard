package xsrcimporter

import (
	"go/build"
	"go/importer"
	"go/token"
	"reflect"
	"testing"
	"unsafe"
)

func TestSize(t *testing.T) {
	fset := token.NewFileSet()
	imp := importer.ForCompiler(fset, "source", nil)
	have := unsafe.Sizeof(srcImporter{})
	want := reflect.ValueOf(imp).Elem().Type().Size()
	if have != want {
		t.Errorf("sizes mismatch: have %d want %d", have, want)
	}
}

func TestImport(t *testing.T) {
	fset := token.NewFileSet()
	imp := New(&build.Default, fset)

	packages := []string{
		"errors",
		"fmt",
		"encoding/json",
	}

	for _, pkgPath := range packages {
		pkg, err := imp.Import(pkgPath)
		if err != nil {
			t.Fatal(err)
		}
		if pkg.Path() != pkgPath {
			t.Fatalf("%s: pkg path mismatch (got %s)", pkgPath, pkg.Path())
		}
		if !pkg.Complete() {
			t.Fatalf("%s is incomplete", pkgPath)
		}
	}

}
