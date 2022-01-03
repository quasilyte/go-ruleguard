//go:build ignore
// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

func testMathRand(m dsl.Matcher) {
	m.Match(`rand.Read($*_)`).Report(`math/rand`)
}

func testCryptoRand(m dsl.Matcher) {
	m.Import(`crypto/rand`)
	m.Match(`rand.Read($*_)`).Report(`crypto/rand`)
}
