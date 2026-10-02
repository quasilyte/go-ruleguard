package xtypes_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/quasilyte/go-ruleguard/internal/xtypes"
)

func TestGenericInterfaceMatching(t *testing.T) {
	src := `package testpkg

type CustomCompare[T any] interface {
	comparable
	GetT() T
}

type MultiParam[K comparable, V any] interface {
	Get(key K) V
	Set(key K, val V)
}

type impl struct{}
func (impl) GetT() int { return 0 }

type notComp struct{ s []int }
func (notComp) GetT() int { return 0 }

type wrongMethod struct{}
func (wrongMethod) GetT() string { return "" }

type store struct{}
func (store) Get(key string) int { return 10 }
func (store) Set(key string, val int) {}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("testpkg", fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatal(err)
	}

	implType := pkg.Scope().Lookup("impl").Type()
	notCompType := pkg.Scope().Lookup("notComp").Type()
	wrongMethodType := pkg.Scope().Lookup("wrongMethod").Type()
	storeType := pkg.Scope().Lookup("store").Type()

	customCompareObj := pkg.Scope().Lookup("CustomCompare")
	ccIface := customCompareObj.Type().Underlying().(*types.Interface)

	multiParamObj := pkg.Scope().Lookup("MultiParam")
	mpIface := multiParamObj.Type().Underlying().(*types.Interface)

	t.Run("CustomCompare uninstantiated", func(t *testing.T) {
		if !xtypes.Implements(implType, ccIface) {
			t.Errorf("implType should implement CustomCompare")
		}
		if xtypes.Implements(notCompType, ccIface) {
			t.Errorf("notCompType should NOT implement CustomCompare")
		}
		if !xtypes.Implements(wrongMethodType, ccIface) {
			t.Errorf("wrongMethodType should implement CustomCompare[string]")
		}
	})

	t.Run("MultiParam uninstantiated", func(t *testing.T) {
		if !xtypes.Implements(storeType, mpIface) {
			t.Errorf("storeType should implement MultiParam")
		}
		if xtypes.Implements(implType, mpIface) {
			t.Errorf("implType should NOT implement MultiParam")
		}
	})
}
