// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules2(m dsl.Matcher) {
	// Override the default text/template to html/template.
	m.Import(`html/template`)

	m.Match(`sink = &$t`).
		Where(m["t"].Type.Is(`template.Template`)).
		Report(`html Template`)
}

func testRules(m dsl.Matcher) {
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
}
