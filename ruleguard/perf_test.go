package ruleguard

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"runtime"
	"strings"
	"testing"
)

func BenchmarkEngineRun(b *testing.B) {
	src := `
	package gorules
	
	import "github.com/quasilyte/go-ruleguard/dsl"

	func sloppyLen(m dsl.Matcher) {
		m.Match("len($_) >= 0").Report("$$ is always true")
		m.Match("len($_) < 0").Report("$$ is always false")
		m.Match("len($x) <= 0").Report("$$ can be len($x) == 0")
	}

	func dupSubExpr(m dsl.Matcher) {
		m.Match("$x || $x",
		"$x && $x",
		"$x | $x",
		"$x & $x",
		"$x ^ $x",
		"$x < $x",
		"$x > $x",
		"$x &^ $x",
		"$x % $s",
		"$x == $x",
		"$x != $x",
		"$x <= $x",
		"$x >= $x",
		"$x / $x",
		"$x - $x").
		Where(m["x"].Pure).
		Report("suspicious identical LHS and RHS")
	}

	func appendCombine(m dsl.Matcher) {
		m.Match("$dst = append($x, $a); $dst = append($x, $b)").
		Report("$dst=append($x,$a,$b) is faster")
	}

	func localVarDecl(m dsl.Matcher) {
		m.Match("var $x = $y").
			Where(!m["$$"].Node.Parent().Is("File")).
			Suggest("$x := $y").
			Report("use := for local variables declaration")
	
		m.Match("var $x $_ = $y").
			Where(!m["$$"].Node.Parent().Is("File")).
			Report("use := for local variables declaration")
	}
	`
	e := benchNewEngine(b, src)
	ctx, files := benchRunContext(e, b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			if err := e.Run(ctx, f); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkIssue412(b *testing.B) {
	src := `package gorules

import "github.com/quasilyte/go-ruleguard/dsl"
import "github.com/quasilyte/go-ruleguard/dsl/types"

func mutexConstructor(m dsl.Matcher) {
	m.Match(
		"$x := &$y{}",
		"$x := $y{}",
		"$x = &$y{}",
		"$x = $y{}",
	).
		Where(
			m["x"].Type.Underlying().Is("struct{$*_; sync.Mutex; $*_}") ||
				m["x"].Filter(embedsMutex),
		).
		Report("Use the lock constructor instead of struct initialization")
}

func embedsMutex(ctx *dsl.VarFilterContext) bool {
	typ := ctx.Type.Underlying()
	asPointer := types.AsPointer(typ)
	if asPointer != nil {
		typ = asPointer.Elem().Underlying()
	}
	asStruct := types.AsStruct(typ)
	if asStruct == nil {
		return false
	}
	mutexType := ctx.GetType("sync.Mutex")
	i := 0
	for i < asStruct.NumFields() {
		field := asStruct.Field(i)
		if field.Embedded() && types.Identical(field.Type(), mutexType) {
			return true
		}
		i++
	}
	return false
}
	`
	e := benchNewEngine(b, src)
	ctx, files := benchRunContext(e, b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			if err := e.Run(ctx, f); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func benchRunContext(e *Engine, b *testing.B) (*RunContext, []*ast.File) {
	b.Helper()

	fset := token.NewFileSet()
	pkgMap, err := parser.ParseDir(fset, "testdata", nil, parser.ParseComments)
	if err != nil {
		b.Fatalf("parse Go file: %v", err)
	}

	var files []*ast.File
	for _, pkg := range pkgMap {
		for _, f := range pkg.Files {
			files = append(files, f)
		}
	}

	imp := importer.ForCompiler(fset, "source", nil)
	typechecker := types.Config{Importer: imp}
	typesInfo := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Uses:  map[*ast.Ident]types.Object{},
		Defs:  map[*ast.Ident]types.Object{},
	}
	pkg, err := typechecker.Check("bench", fset, files, typesInfo)
	if err != nil {
		b.Fatalf("typecheck: %v", err)
	}

	ctx := &RunContext{
		Pkg:   pkg,
		Types: typesInfo,
		Sizes: types.SizesFor("gc", runtime.GOARCH),
		Fset:  fset,
		Report: func(data *ReportData) {
			// Do nothing.
		},
		State: NewRunnerState(e),
	}

	return ctx, files
}

func benchNewEngine(b *testing.B, src string) *Engine {
	b.Helper()

	e := NewEngine()
	ctx := &LoadContext{
		Fset: token.NewFileSet(),
	}
	err := e.Load(ctx, "rules.go", strings.NewReader(src))
	if err != nil {
		b.Fatalf("load error: %v", err)
	}
	return e
}
