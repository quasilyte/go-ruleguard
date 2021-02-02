// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	m.Match(`$x, $x`).Report(`repeated expression in list`)
}
