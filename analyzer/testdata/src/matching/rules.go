//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	m.Match(`$x, $x`).Report(`repeated expression in list`)

	m.Match(`range $x[:]`).
		Where(m["x"].Type.Is(`[]$_`)).
		Report(`redundant slicing of a range expression`)
}
