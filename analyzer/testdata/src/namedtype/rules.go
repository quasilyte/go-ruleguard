// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
	m.Import(`namedtype/x/nested`)

	m.Match(`sink = &$t`).
		Where(m["t"].Type.Is(`nested.Element`)).
		Report(`x/nested Element`)
}
