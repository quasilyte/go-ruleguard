// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl/fluent"
)

func issue68(m fluent.Matcher) {
	m.Match(`func $_($_ *testing.T) { $p; $*_ }`).Where(m["p"].Text != "t.Parallel()").Report(`Not a parallel test`)
	m.Match(`func $_($_ *testing.T) { $p; $*_ }`).Where(m["p"].Text == "t.Parallel()").Report(`Parallel test`)
}
