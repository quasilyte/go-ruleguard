// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	m.Match(`errors.New(fmt.Sprintf($*_))`).
		Report(`should replace error.New(fmt.Sprintf(...)) with fmt.Errorf(...)`)
	m.Match(`t.Error(fmt.Sprintf($*_))`).
		Report(`should replace t.Error(fmt.Sprintf(...)) with t.Errorf(...)`)

	// m.Match(`for $x, _ := range $_ { $*_ }`).Report(`should omit 2nd value from range; this loop is equivalent to 'for $x := range ...'`)
	// m.Match(`for $x, _ = range $_ { $*_ }`).Report(`should omit 2nd value from range; this loop is equivalent to 'for $x = range ...'`)

	m.Match(`if $_ { $*_; return $*_ } else { $*_ }`).
		Report(`if block ends with a return statement, so drop this else and outdent its block`)
}
