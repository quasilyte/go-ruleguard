package ruleguard

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"runtime"
	"strings"
	"testing"
)

func TestGenericInterfaceMatching(t *testing.T) {
	rules := `
package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

type CustomCompare[T any] interface {
	comparable
	GetT() T
}

type MultiParam[K comparable, V any] interface {
	Get(key K) V
	Set(key K, val V)
}

func testRules(m dsl.Matcher) {
	m.Match("matchUninst($x)").
		Where(m["x"].Type.Implements("CustomCompare")).
		Report("matched CustomCompare uninstantiated")

	m.Match("matchInt($x)").
		Where(m["x"].Type.Implements("CustomCompare[int]")).
		Report("matched CustomCompare[int]")

	m.Match("matchString($x)").
		Where(m["x"].Type.Implements("CustomCompare[string]")).
		Report("matched CustomCompare[string]")

	m.Match("matchMulti($x)").
		Where(m["x"].Type.Implements("MultiParam")).
		Report("matched MultiParam uninstantiated")

	m.Match("matchMultiInst($x)").
		Where(m["x"].Type.Implements("MultiParam[string, int]")).
		Report("matched MultiParam[string, int]")

	m.Match("matchMultiWrong($x)").
		Where(m["x"].Type.Implements("MultiParam[int, string]")).
		Report("matched MultiParam[int, string]")
}
`

	targetSrc := `
package testpkg

type impl struct{}
func (impl) GetT() int { return 0 }

type stringImpl struct{}
func (stringImpl) GetT() string { return "" }

type notComp struct{ s []int }
func (notComp) GetT() int { return 0 }

type store struct{}
func (store) Get(key string) int { return 0 }
func (store) Set(key string, val int) {}

type other struct{}

func matchUninst(interface{}) {}
func matchInt(interface{}) {}
func matchString(interface{}) {}
func matchMulti(interface{}) {}
func matchMultiInst(interface{}) {}
func matchMultiWrong(interface{}) {}

func testFunc() {
	i := impl{}
	s := stringImpl{}
	nc := notComp{}
	st := store{}
	o := other{}

	matchUninst(i)      // want matched CustomCompare uninstantiated
	matchUninst(s)      // want matched CustomCompare uninstantiated
	matchUninst(nc)     // no match (not comparable)
	matchUninst(o)      // no match

	matchInt(i)         // want matched CustomCompare[int]
	matchInt(s)         // no match (returns string, not int)
	matchInt(nc)        // no match (not comparable)

	matchString(s)      // want matched CustomCompare[string]
	matchString(i)      // no match (returns int, not string)

	matchMulti(st)      // want matched MultiParam uninstantiated
	matchMulti(i)       // no match
	matchMultiInst(st)  // want matched MultiParam[string, int]
	matchMultiWrong(st) // no match (key is string, not int)
}
`

	e := NewEngine()
	ctx := &LoadContext{
		Fset: token.NewFileSet(),
	}
	if err := e.Load(ctx, "rules.go", strings.NewReader(rules)); err != nil {
		t.Fatalf("load rules: %v", err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "target.go", targetSrc, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse target: %v", err)
	}

	var typechecker types.Config
	info := types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
		Uses:  map[*ast.Ident]types.Object{},
	}
	pkg, err := typechecker.Check("testpkg", fset, []*ast.File{f}, &info)
	if err != nil {
		t.Fatalf("typecheck target: %v", err)
	}

	var reports []string
	runCtx := &RunContext{
		Pkg:   pkg,
		Types: &info,
		Sizes: types.SizesFor("gc", runtime.GOARCH),
		Fset:  fset,
		Report: func(data *ReportData) {
			reports = append(reports, fmt.Sprintf("%d: %s", fset.Position(data.Node.Pos()).Line, data.Message))
		},
	}

	if err := e.Run(runCtx, f); err != nil {
		t.Fatalf("run rules: %v", err)
	}

	t.Logf("reports:\n%s", strings.Join(reports, "\n"))

	expected := []struct {
		line    int
		message string
	}{
		{33, "matched CustomCompare uninstantiated"},
		{34, "matched CustomCompare uninstantiated"},
		{38, "matched CustomCompare[int]"},
		{42, "matched CustomCompare[string]"},
		{45, "matched MultiParam uninstantiated"},
		{47, "matched MultiParam[string, int]"},
	}

	if len(reports) != len(expected) {
		t.Fatalf("got %d reports, want %d:\ngot:\n%s", len(reports), len(expected), strings.Join(reports, "\n"))
	}

	for i, exp := range expected {
		expectedStr := fmt.Sprintf("%d: %s", exp.line, exp.message)
		if reports[i] != expectedStr {
			t.Errorf("report[%d] = %q, want %q", i, reports[i], expectedStr)
		}
	}
}
