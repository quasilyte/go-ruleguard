package ruleguard

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quasilyte/go-ruleguard/internal/gogrep"
)

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

	e := NewEngine()
	var rr rulesRunner
	rr.state = e.impl.state
	rr.ctx = &RunContext{
		Fset: token.NewFileSet(),
	}
	for _, test := range tests {
		capture := make([]gogrep.CapturedNode, len(test.vars))
		for i, v := range test.vars {
			capture[i] = gogrep.CapturedNode{
				Name: v,
				Node: &ast.Ident{Name: v + "var"},
			}
		}

		m := gogrep.MatchData{
			Node:    &ast.Ident{Name: "dd"},
			Capture: capture,
		}
		have := rr.renderMessage(test.msg, m, false)
		if diff := cmp.Diff(have, test.want); diff != "" {
			t.Errorf("render %s %v:\n(+want -have)\n%s", test.msg, test.vars, diff)
		}
	}
}
