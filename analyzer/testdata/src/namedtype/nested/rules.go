// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
	m.Import(`namedtype/x/nested`)
	m.Import(`extra`)

	m.Match(`sink = $t`).
		Where(m["t"].Type.Is(`extra.Value`)).
		Report(`extra Value`)

	m.Match(`sink = &$t`).
		Where(m["t"].Type.Is(`nested.Element`)).
		Report(`x/nested Element`)

	m.Match(`sink = &$t`).
		Where(m["t"].Type.Is(`list.Element`)).
		Report(`list Element`)

	m.Match(`sink = &$t`).
		Where(m["t"].Type.Is(`template.Template`)).
		Report(`text Template`)

	// Now override the default text/template to html/template.
	m.Import(`html/template`)

	m.Match(`sink = &$t`).
		Where(m["t"].Type.Is(`template.Template`)).
		Report(`html Template`)

}
