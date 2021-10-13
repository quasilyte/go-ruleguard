package gogrep

import (
	"go/token"
	"strings"
	"testing"
)

func BenchmarkMatch(b *testing.B) {
	tests := []struct {
		name  string
		pat   string
		input string
	}{
		{
			name:  `failFast`,
			pat:   `f()`,
			input: `[]int{}`,
		},
		{
			name:  `failCall`,
			pat:   `f(1, 2, 3, 4)`,
			input: `f(1, 2, 3, _)`,
		},
		{
			name:  `failCallFast`,
			pat:   `f(1, 2, 3)`,
			input: `f()`,
		},
		{
			name:  `assign`,
			pat:   `x := 1`,
			input: `x := 1`,
		},
		{
			name:  `assignMulti`,
			pat:   `$*_ = f()`,
			input: `x, y = f()`,
		},
		{
			name:  `simpleLit`,
			pat:   `true`,
			input: `true`,
		},
		{
			name:  `simpleBinaryOp`,
			pat:   `1 + 2`,
			input: `1 + 2`,
		},
		{
			name:  `simpleSelectorExpr`,
			pat:   `x.y.z`,
			input: `x.y.z`,
		},
		{
			name:  `simpleCall`,
			pat:   `f(1, 2)`,
			input: `f(1, 2)`,
		},
		{
			name:  `selectorExpr`,
			pat:   `a.$x`,
			input: `a.b.c`,
		},
		{
			name:  `sliceExpr`,
			pat:   `s[$x:]`,
			input: `s[1:]`,
		},
		{
			name:  `any`,
			pat:   `$_`,
			input: `x + y`,
		},
		{
			name:  `anyCall`,
			pat:   `$_($*_)`,
			input: `f(1, "2", '3',)`,
		},
		{
			name:  `ifStmt`,
			pat:   `if cond { $y }`,
			input: `if cond { return nil }`,
		},
		{
			name:  `optStmt1`,
			pat:   `if $*_ {}`,
			input: `if init; f() {}`,
		},
		{
			name:  `optStmt2`,
			pat:   `if $*_; cond {}`,
			input: `if cond {}`,
		},
		{
			name:  `namedOptStmt1`,
			pat:   `if $*x {}; if $*x {}`,
			input: `{ if init; cond {}; if init; cond {} }`,
		},
		{
			name:  `namedOptStmt2`,
			pat:   `if $*x; cond {}; if $*x; cond {}`,
			input: `{ if init; cond {}; if init; cond {} }`,
		},
		{
			name:  `branchStmt`,
			pat:   `break foo`,
			input: `break foo`,
		},
		{
			name:  `multiStmt`,
			pat:   `x; y`,
			input: `{ f(); x; y; z }`,
		},
		{
			name:  `multiExpr`,
			pat:   `x, y`,
			input: `f(_, x, _, _, x, y, _)`,
		},
		{
			name:  `variadicCall`,
			pat:   `f(xs...)`,
			input: `f(xs...)`,
		},
		{
			name:  `capture1`,
			pat:   `+$x`,
			input: `+50`,
		},
		{
			name:  `capture2`,
			pat:   `$x + $y`,
			input: `x + 4`,
		},
		{
			name:  `capture8`,
			pat:   `f($x1, $x2, $x3, $x4, $x5, $x6, $x7, $x8)`,
			input: `f(1, 2, 3, 4, 5, 6, 7, 8)`,
		},
		{
			name:  `capture2same`,
			pat:   `$x + $x`,
			input: `a + a`,
		},
		{
			name:  `capture8same`,
			pat:   `f($x, $x, $x, $x, $x, $x, $x, $x)`,
			input: `f(1, 1, 1, 1, 1, 1, 1, 1)`,
		},
		{
			name:  `captureBacktrackLeft`,
			pat:   `f($*xs, $y)`,
			input: `f(1, 2, 3, 4, 5, 6)`,
		},
		{
			name:  `captureBacktrackRight`,
			pat:   `f($x, $*ys)`,
			input: `f(1, 2, 3, 4, 5, 6)`,
		},
	}

	for i := range tests {
		test := tests[i]
		b.Run(test.name, func(b *testing.B) {
			fset := token.NewFileSet()
			pat, err := Compile(fset, test.pat, true)
			if err != nil {
				b.Errorf("parse `%s`: %v", test.pat, err)
				return
			}
			target := testParseNode(b, token.NewFileSet(), test.input)
			if err != nil {
				b.Errorf("parse target `%s`: %v", test.input, err)
				return
			}
			if !strings.HasPrefix(test.name, "fail") {
				matches := 0
				testAllMatches(pat, target, func(m MatchData) {
					matches++
				})
				if matches == 0 {
					b.Fatal("matching failed")
				}
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				testAllMatches(pat, target, func(m MatchData) {})
			}
		})
	}
}
