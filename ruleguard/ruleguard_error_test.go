package ruleguard

import (
	"errors"
	"fmt"
	"go/token"
	"strings"
	"testing"
)

func TestImportError(t *testing.T) {
	src := `
	package gorules
	import "github.com/quasilyte/go-ruleguard/dsl"
	func badLock(m dsl.Matcher) {
		m.Import("foo/nonexisting")
		m.Match("$x").Where(m["x"].Type.Implements("nonexisting.Iface")).Report("ok")
	}
	`
	e := NewEngine()
	ctx := &ParseContext{
		Fset: token.NewFileSet(),
	}
	err := e.Load(ctx, "rules.go", strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error, got none")
	}
	var importError *ImportError
	if !errors.As(err, &importError) {
		t.Fatal("got import that is not ImportError")
	}
}

func TestParseFilterFuncError(t *testing.T) {
	type testCase struct {
		src string
		err string
	}

	simpleTests := []testCase{
		// Unsupported features.
		// Some of them might be implemented later, but for now
		// we want to ensure that the user gets understandable error messages.
		{
			`b := true; switch true {}; return b`,
			`can't compile *ast.SwitchStmt yet`,
		},
		{
			`b := 0; return &b != nil`,
			`can't compile unary & yet`,
		},
		{
			`b := 0; return (b << 1) != 0`,
			`can't compile binary << yet`,
		},
		{
			`return g(ctx)`,
			`can't compile a call to gorules.g func`,
		},
		{
			`return new(int) != nil`,
			`can't compile new() builtin function call yet`,
		},
		{
			`x := 5.6; return x != 0`,
			`can't compile float constants yet`,
		},
		{
			`s := ""; return s >= "a"`,
			`>= is not implemented for string operands`,
		},
		{
			`s := "foo"; b := s[0]; return b == 0`,
			`can't compile *ast.IndexExpr yet`,
		},
		{
			`s := Foo{}; return s.X == 0`,
			`can't compile *ast.CompositeLit yet`,
		},

		// Assignment errors.
		{
			`x, y := 1, 2; return x == y`,
			`only single left operand is allowed in assignments`,
		},
		{
			`x := 0; { x := 1; return x == 1 }; return x == 0`,
			`x variable shadowing is not allowed`,
		},
		{
			`ctx = ctx; return true`,
			`can't assign to ctx, params are readonly`,
		},
		{
			`ctx.Type = nil; return true`,
			`can assign only to simple variables`,
		},
		{
			`i++; return true`,
			`can't assign to i, params are readonly`,
		},

		// Unsupported type errors.
		{
			`x := int32(0); return x == 0`,
			`x local variable type: int32 is not supported, try something simpler`,
		},

		// Implementation limits.
		{
			`x1:=1; x2:=x1; x3:=x2; x4:=x3; x5:=x4; x6:=x5; x7:=x6; x8:=x7; x9:=x8; return x9 == 1`,
			`can't define x9: too many locals`,
		},
	}

	tests := []testCase{
		{
			`func f() int32 { return 0 }`,
			`function result type: int32 is not supported, try something simpler`,
		},
		{
			`func f() []int { return nil }`,
			`function result type: []int is not supported, try something simpler`,
		},
		{
			`func f(s *string) int { return 0 }`,
			`s param type: *string is not supported, try something simpler`,
		},

		{
			`func f(foo *Foo) int { return foo.X }`,
			`can't compile X field access`,
		},
		{
			`func f(foo *Foo) string { return foo.String() }`,
			`can't compile a call to *gorules.Foo.String func`,
		},

		{
			`func f() {}`,
			`only functions with a single non-void results are supported`,
		},
		{
			`func f() (int, int) { return 0, 0 }`,
			`only functions with a single non-void results are supported`,
		},

		{
			`func f() (b bool) { return }`,
			`'naked' return statements are not allowed`,
		},
	}

	for _, test := range simpleTests {
		test.src = `func f(ctx *dsl.VarFilterContext, i int) bool { ` + test.src + ` }`
		tests = append(tests, test)
	}

	for _, test := range tests {
		file := fmt.Sprintf(`
			package gorules
			import "github.com/quasilyte/go-ruleguard/dsl"
			type Foo struct { X int }
			func (foo *Foo) String() string { return "" }
			func g(ctx *dsl.VarFilterContext) bool { return false }
			` + test.src)
		e := NewEngine()
		ctx := &ParseContext{
			Fset: token.NewFileSet(),
		}
		err := e.Load(ctx, "rules.go", strings.NewReader(file))
		if err == nil {
			t.Errorf("parse %s: expected %s error, got none", test.src, test.err)
			continue
		}
		have := err.Error()
		want := test.err
		if !strings.Contains(have, want) {
			t.Errorf("parse %s: errors mismatch:\nhave: %s\nwant: %s", test.src, have, want)
			continue
		}
	}
}

func TestParseRuleError(t *testing.T) {
	tests := []struct {
		expr string
		err  string
	}{
		{
			`m.Match("$x").MatchComment("").Report("")`,
			`Match() and MatchComment() can't be combined`,
		},

		{
			`m.MatchComment("").Match("$x").Report("")`,
			`Match() and MatchComment() can't be combined`,
		},

		{
			`m.Where(m.File().Imports("strings")).Report("no match call")`,
			`missing Match() or MatchComment() call`,
		},

		{
			`m.Match("$x").Where(m["x"].Pure)`,
			`missing Report() or Suggest() call`,
		},

		{
			`m.Match("$x").Match("$x")`,
			`Match() can't be repeated`,
		},

		{
			`m.MatchComment("").MatchComment("")`,
			`MatchComment() can't be repeated`,
		},

		{
			`m.Match().Report("$$")`,
			`too few arguments in call to m.Match`,
		},

		{
			`m.MatchComment().Report("$$")`,
			`too few arguments in call to m.MatchComment`,
		},

		{
			`m.MatchComment("(").Report("")`,
			`error parsing regexp: missing closing )`,
		},

		{
			`m.Match("func[]").Report("$$")`,
			`parse match pattern: cannot parse expr: 1:5: expected '(', found '['`,
		},
	}

	for _, test := range tests {
		file := fmt.Sprintf(`
			package gorules
			import "github.com/quasilyte/go-ruleguard/dsl"
			func testrule(m dsl.Matcher) {
				%s
			}`,
			test.expr)
		e := NewEngine()
		ctx := &ParseContext{
			Fset: token.NewFileSet(),
		}
		err := e.Load(ctx, "rules.go", strings.NewReader(file))
		if err == nil {
			t.Errorf("parse %s: expected %s error, got none", test.expr, test.err)
			continue
		}
		have := err.Error()
		want := test.err
		if !strings.Contains(have, want) {
			t.Errorf("parse %s: errors mismatch:\nhave: %s\nwant: %s", test.expr, have, want)
			continue
		}
	}
}

func TestParseFilterError(t *testing.T) {
	tests := []struct {
		expr string
		err  string
	}{
		{
			`true`,
			`unsupported expr: true`,
		},

		{
			`m["x"].Text == 5`,
			`cannot convert 5 (untyped int constant) to string`,
		},

		{
			`m["x"].Text.Matches("(12")`,
			`error parsing regexp: missing closing )`,
		},

		{
			`m["x"].Type.Is("%illegal")`,
			`parse type expr: 1:1: expected operand, found '%'`,
		},

		{
			`m["x"].Type.Is("interface{String() string}")`,
			`parse type expr: can't convert interface{String() string} type expression`,
		},

		{
			`m["x"].Type.ConvertibleTo("interface{String() string}")`,
			`can't convert interface{String() string} into a type constraint yet`,
		},

		{
			`m["x"].Type.AssignableTo("interface{String() string}")`,
			`can't convert interface{String() string} into a type constraint yet`,
		},

		{
			`m["x"].Type.Implements("foo")`,
			`can't resolve foo type; try a fully-qualified name`,
		},
		{
			`m["x"].Type.Implements("func()")`,
			`can't resolve func() type; try a fully-qualified name`,
		},
		{
			`m["x"].Type.Implements("bytes.Buffer")`,
			`bytes.Buffer is not an interface type`,
		},

		{
			`m["x"].Type.Implements("foo.Bar")`,
			`package foo is not imported`,
		},

		{
			`m["x"].Type.Implements("strings.Replacer3")`,
			`Replacer3 is not found in strings`,
		},

		{
			`m["x"].Node.Is("abc")`,
			`abc is not a valid go/ast type name`,
		},
	}

	for _, test := range tests {
		file := fmt.Sprintf(`
			package gorules
			import "github.com/quasilyte/go-ruleguard/dsl"
			func testrule(m dsl.Matcher) {
				m.Match("$x + $y[$key]").Where(%s).Report("$$")
			}`,
			test.expr)
		e := NewEngine()
		ctx := &ParseContext{
			Fset: token.NewFileSet(),
		}
		err := e.Load(ctx, "rules.go", strings.NewReader(file))
		if err == nil {
			t.Errorf("parse %s: expected %s error, got none", test.expr, test.err)
			continue
		}
		have := err.Error()
		want := test.err
		if !strings.Contains(have, want) {
			t.Errorf("parse %s: errors mismatch:\nhave: %s\nwant: %s", test.expr, have, want)
			continue
		}
	}
}
