// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func nolintFormat(m dsl.Matcher) {
	m.MatchComment(`// nolint(?:$|\s)`).Report(`remove a space between // and "nolint" directive`)

	// Using ?s flag to match multiline text.
	m.MatchComment(`(?s)/\*.*nolint.*`).
		Report(`don't put "nolint" inside a multi-line comment`)
}

// Testing a pattern without capture groups.
func nolintGocritic(m dsl.Matcher) {
	m.MatchComment(`//nolint:gocritic`).
		Report(`hey, this is kinda upsetting`)
}

// Testing Suggest.
func nolint2quickfix(m dsl.Matcher) {
	m.MatchComment(`// nolint2(?P<rest>.*)`).Suggest(`//nolint2$rest`)
}

// Testing named groups.
func directiveFormat(m dsl.Matcher) {
	m.MatchComment(`/\*(?P<directive>[\w-]+):`).
		Report(`directive should be written as //$directive`)
}

// Testing Where clause.
func forbiddenGoDirective(m dsl.Matcher) {
	m.MatchComment(`//go:(?P<x>\w+)`).
		Where(m["x"].Text != "generate" && !m["x"].Text.Matches(`embed|noinline`)).
		Report("don't use $x go directive")
}

// Test that file-related things work for MatchComment too.
func fooDirectives(m dsl.Matcher) {
	m.MatchComment(`//go:embed`).
		Where(m.File().Name.Matches(`^.*_foo.go$`)).
		Report("don't use go:embed in _foo files")
}

// Test multi-pattern matching.
// Also test $$ interpolation.
func commentTypo(m dsl.Matcher) {
	m.MatchComment(`begining`, `bizzare`).Report(`"$$" may contain a typo`)
}

// Test $$ with a more complex regexp that captures something.
func commentTypo2(m dsl.Matcher) {
	m.MatchComment(`(?P<word>buisness) advice`).Report(`"$$" may contain a typo`)
}

// Test mixture of the named and unnamed captures.
func commentTypo3(m dsl.Matcher) {
	m.MatchComment(`(?P<first>calender)|(error)`).Report(`first=$first`)
	m.MatchComment(`(error)|(?P<second>cemetary)`).Report(`second=$second`)
}

// Test a case where named group is empty.
func commentTypo4(m dsl.Matcher) {
	m.MatchComment(`(?P<x>collegue)|(commitee)`).Report(`x="$x"`)
}
