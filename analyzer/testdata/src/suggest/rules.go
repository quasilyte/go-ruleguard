// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func testRules(m fluent.Matcher) {
	m.Match(`($a) || ($b)`).Suggest(`$a || $b`)
}
