package gogrep

import (
	"go/token"
	"testing"
)

func BenchmarkMatch(b *testing.B) {
	tests := []struct {
		name  string
		pat   string
		input string
	}{
		{
			name:  `simpleLit`,
			pat:   `true`,
			input: `true`,
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
			pat, err := Parse(fset, test.pat, true)
			if err != nil {
				b.Errorf("parse `%s`: %v", test.pat, err)
				return
			}
			target := testParseNode(b, test.input)
			if err != nil {
				b.Errorf("parse target `%s`: %v", test.input, err)
				return
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				testAllMatches(pat, target, func(m MatchData) {})
			}
		})
	}
}
