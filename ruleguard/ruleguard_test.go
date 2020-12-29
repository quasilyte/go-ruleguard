package ruleguard

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseRuleError(t *testing.T) {
	tests := []struct {
		expr string
		err  string
	}{
		{
			`m.Where(m.File().Imports("strings")).Report("no match call")`,
			`missing Match() call`,
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
			`m.Match().Report("$$")`,
			`too few arguments in call to m.Match`,
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
		ctx := &ParseContext{Fset: token.NewFileSet()}
		_, err := ParseRules(ctx, "rules.go", strings.NewReader(file))
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
			"only `error` unqualified type is recognized",
		},

		{
			`m["x"].Type.Implements("func()")`,
			"only qualified names (and `error`) are supported",
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
		ctx := &ParseContext{Fset: token.NewFileSet()}
		_, err := ParseRules(ctx, "rules.go", strings.NewReader(file))
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

func TestRenderMessage(t *testing.T) {
	tests := []struct {
		msg  string
		want string
		vars []string
	}{
		{
			`$x`,
			`xvar`,
			[]string{`x`},
		},
		{
			`$x$x`,
			`xvarxvar`,
			[]string{`x`},
		},
		{
			`$f $foo`,
			`fvar foovar`,
			[]string{`f`, `foo`},
		},
		{
			`$foo $f`,
			`foovar fvar`,
			[]string{`f`, `foo`},
		},
		{
			`$foo $f`,
			`foovar fvar`,
			[]string{`foo`, `f`},
		},
		{
			`$foo $foo $f`,
			`foovar foovar fvar`,
			[]string{`foo`, `f`},
		},
		{
			`$foo$f`,
			`foovarfvar`,
			[]string{`foo`, `f`},
		},
		{
			`$foo($f) + $f.$foo`,
			`foovar(fvar) + fvar.foovar`,
			[]string{`foo`, `f`},
		},

		// Do we care about finding a proper variable border?
		{
			`$fooo`,
			`foovaro`,
			[]string{`foo`},
		},

		// Unknown $-expressions are not interpolated.
		{
			`$nonexisting`,
			`$nonexisting`,
			[]string{`x`},
		},

		// Double dollar interpolation.
		{
			`$$`,
			`dd`,
			[]string{`x`},
		},
		{
			`$$[$x]`,
			`dd[xvar]`,
			[]string{`x`},
		},
	}

	var rr rulesRunner
	rr.ctx = &Context{
		Fset: token.NewFileSet(),
	}
	for _, test := range tests {
		nodes := make(map[string]ast.Node, len(test.vars))
		for _, v := range test.vars {
			nodes[v] = &ast.Ident{Name: v + "var"}
		}

		have := rr.renderMessage(test.msg, &ast.Ident{Name: "dd"}, nodes, false)
		if diff := cmp.Diff(have, test.want); diff != "" {
			t.Errorf("render %s %v:\n(+want -have)\n%s", test.msg, test.vars, diff)
		}
	}
}
