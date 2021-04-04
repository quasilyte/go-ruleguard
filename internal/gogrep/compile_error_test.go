package gogrep

import (
	"fmt"
	"go/token"
	"strings"
	"testing"
)

func TestCompileError(t *testing.T) {
	strict := func(s string) string {
		return "STRICT " + s
	}
	isStrict := func(s string) bool {
		return strings.HasPrefix(s, "STRICT ")
	}
	unwrapPattern := func(s string) string {
		s = strings.TrimPrefix(s, "STRICT ")
		return s
	}

	intStatements := func() string {
		parts := make([]string, 260)
		for i := range parts {
			parts[i] = fmt.Sprint(i)
		}
		return strings.Join(parts, ";")
	}()

	tests := map[string]string{
		`$$`: `$ must be followed by ident, got ILLEGAL`,
		`$`:  `$ must be followed by ident, got EOF`,

		``:   `empty source code`,
		"\t": `empty source code`,

		`foo)`:   `expected statement, found ')'`,
		`$x)`:    `expected statement, found ')'`,
		`$x(`:    `expected operand, found '}'`,
		`$*x)`:   `expected statement, found ')'`,
		"a\n$x)": `expected statement, found ')'`,

		`0xabci`: `can't convert 0xabci (IMAG) value`,

		intStatements:         `implementation limitation: too many values`,
		strict(intStatements): `implementation limitation: too many string values`,
	}

	for input, want := range tests {
		fset := token.NewFileSet()
		testPattern := unwrapPattern(input)
		_, err := Compile(fset, testPattern, isStrict(input))
		if err == nil {
			t.Errorf("compile `%s`: expected error, got none", input)
			continue
		}
		if !strings.Contains(err.Error(), want) {
			t.Errorf("compile `%s`: error substring not found\nerror: %s\nsubstr: %s",
				input, err, want)
		}
	}
}
