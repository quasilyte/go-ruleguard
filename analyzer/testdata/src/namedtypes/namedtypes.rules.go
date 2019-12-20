// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
	m.Match(`sink = &$t`).
		Where(m["t"].Type.Is(`list.Element`)).
		Report(`list Element`)

	m.Match(`sink = &$t`).
		Where(m["t"].Type.Is(`Element`)).
		Report(`Element`)
}
