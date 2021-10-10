// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

func testRules(m dsl.Matcher) {
	m.Match(`io.WriteString($*_)`).Report(`WriteString from stdlib`)

	m.Match(`sink(fmt.Sprint($_), fmt.Sprint($_))`).Report(`sink with two Sprint from stdlib`)
}
