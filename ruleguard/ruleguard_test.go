package ruleguard

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quasilyte/gogrep"
)

func TestTruncateText(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello world", 60, "hello world"},
		{"hello world", 8, "h<...>ld"},
		{"hello world", 7, "h<...>d"},
		{"hello world", 6, "<...>d"},
		{"hello world", 5, "<...>"},
		{"have := truncateText(test.input, test.maxLen)", 20, "have :=<...>.maxLen)"},
		{"have := truncateText(test.input, test.maxLen)", 30, "have := trun<...> test.maxLen)"},
		{"have := truncateText(test.input, test.maxLen)", 40, "have := truncateT<...>nput, test.maxLen)"},
		{"have := truncateText(test.input, test.maxLen)", 41, "have := truncateTe<...>nput, test.maxLen)"},
		{"have := truncateText(test.input, test.maxLen)", 42, "have := truncateTe<...>input, test.maxLen)"},
		{"have := truncateText(test.input, test.maxLen)", 50, "have := truncateText(test.input, test.maxLen)"},
	}

	for _, test := range tests {
		have := string(truncateText([]byte(test.input), test.maxLen))
		if len(have) > test.maxLen {
			t.Errorf("truncateText(%q, %v): len %d exceeds max len",
				test.input, test.maxLen, len(have))
		}
		if len(test.input) > test.maxLen && len(have) != test.maxLen {
			t.Errorf("truncateText(%q, %v): truncated more than necessary",
				test.input, test.maxLen)
		}
		if have != test.want {
			t.Fatalf("truncateText(%q, %v):\nhave: %q\nwant: %q",
				test.input, test.maxLen, have, test.want)
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
		have := rr.renderMessage(test.msg, matchData{match: m}, false)
		if diff := cmp.Diff(have, test.want); diff != "" {
			t.Errorf("render %s %v:\n(+want -have)\n%s", test.msg, test.vars, diff)
		}
	}
}
