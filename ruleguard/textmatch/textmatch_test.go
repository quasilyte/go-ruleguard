package textmatch

import (
	"fmt"
	"regexp"
	"testing"
)

func TestCompileAndRun(t *testing.T) {
	tests := []struct {
		re              string
		expectedMatcher string
	}{
		{`foo`, `containsLiteralMatcher`},
		{`.*foo.*`, `containsLiteralMatcher`},
		{`^foo$`, `eqLiteralMatcher`},
		{`^foo`, `prefixLiteralMatcher`},
		{`foo$`, `suffixLiteralMatcher`},

		{`^\p{Lu}`, `prefixRunePredMatcher`},
		{`^\p{Ll}`, `prefixRunePredMatcher`},
	}

	inputs := make([]string, 0, len(inputStrings))
	for _, s := range inputStrings {
		inputs = append(inputs, s)
		inputs = append(inputs, s+" "+s)
		inputs = append(inputs, s+"_"+s)
		inputs = append(inputs, " "+s)
		inputs = append(inputs, s+" ")
		inputs = append(inputs, " "+s+" ")
		inputs = append(inputs, "\n"+s+"\n")
	}

	for _, test := range tests {
		p, err := Compile(test.re)
		if err != nil {
			t.Fatal(err)
		}
		wantMatcher := `*textmatch.` + test.expectedMatcher
		if IsRegexp(p) {
			t.Errorf("`%s` is not optimized (want %s)", test.re, wantMatcher)
			continue
		}
		haveMatcher := fmt.Sprintf("%T", p)
		if haveMatcher != wantMatcher {
			t.Errorf("`%s` matcher is %s, want %s", test.re, haveMatcher, wantMatcher)
			continue
		}
		re, err := regexp.Compile(test.re)
		if err != nil {
			t.Fatal(err)
		}
		for _, input := range inputs {
			have := p.MatchString(input)
			want := re.MatchString(input)
			if have != want {
				t.Errorf("`%s` invalid MatchString() result on %q (want %v)", test.re, input, want)
				break
			}
			have = p.Match([]byte(input))
			want = re.Match([]byte(input))
			if have != want {
				t.Errorf("`%s` invalid Match() result on %q (want %v)", test.re, input, want)
				break
			}
		}
	}
}

func BenchmarkMatch(b *testing.B) {
	tests := []struct {
		re     string
		inputs []string
	}{
		{
			`^\p{Lu}`,
			[]string{
				`Foo`,
				`foo`,
			},
		},

		{
			`^\p{Ll}`,
			[]string{
				`foo`,
				`Foo`,
			},
		},

		{
			`foo$`,
			[]string{
				`   foo`,
				`bar`,
			},
		},

		{
			`^foo`,
			[]string{
				`foo`,
				`   bar`,
			},
		},

		{
			`.*simpleIdent.*`,
			[]string{
				`text simpleIdent other text`,
				`text without matching ident`,
			},
		},

		{
			`simpleIdent`,
			[]string{
				`simpleIdent`,
				`text without simpleIdent`,
			},
		},
	}

	for _, test := range tests {
		re, err := regexp.Compile(test.re)
		if err != nil {
			b.Fatal(err)
		}
		pat, err := Compile(test.re)
		if err != nil {
			b.Fatal(err)
		}
		if IsRegexp(pat) {
			b.Fatalf("`%s` is not optimized", test.re)
		}
		for i, input := range test.inputs {
			b.Run(fmt.Sprintf("%s_%d_re", test.re, i), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = re.MatchString(input)
				}
			})
			b.Run(fmt.Sprintf("%s_%d_opt", test.re, i), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = pat.MatchString(input)
				}
			})
		}
	}
}

var inputStrings = []string{
	``,
	"\x00",
	`foo`,
	`foo2`,
	`_foo`,
	`foobarfoo`,
	`Foo`,
	`FOO`,
	`bar_baz`,
	`2493`,
	"some longer text fragment (foo)",
	"multi\nline\ntext\fragment",
	"foo\nbar\n(foo)\n\n",
	"ƇƉ",
}
