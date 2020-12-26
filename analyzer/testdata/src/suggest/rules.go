// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	m.Match(`($a) || ($b)`).Suggest(`$a || $b`)
}
