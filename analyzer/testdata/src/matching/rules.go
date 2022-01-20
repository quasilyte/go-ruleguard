//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	m.Match(`$x, $x`).Report(`repeated expression in list`)

	m.Match(`range $x[:]`).
		Where(m["x"].Type.Is(`[]$_`)).
		Report(`redundant slicing of a range expression`)

	m.Match(`$lhs = append($lhs, $x)`).
		Where(m["lhs"].Contains(`$x`)).
		Report(`$lhs contains $x`)

	m.Match(`copy($*args)`).
		Where(m["args"].Contains(`$_[:]`)).
		Report(`copy() contains a slicing operation`)

	m.Match(`sink($err); $x`).
		Where(!m["x"].Contains(`return $err`) && m["err"].Type.Is(`error`)).
		Report(`sink($err) call not followed by return $err`)

	m.Match(`for { $*body }`).
		Where(m["body"].Contains(`expensive()`)).
		Report(`expensive call inside a loop`)

	m.Match(`$x != nil && $y`).
		Where(m["y"].Contains(`$x[0]`)).
		Report(`nil check may not be enough to access $x[0], check for len`)
}
