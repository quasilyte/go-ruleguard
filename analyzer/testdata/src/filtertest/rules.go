// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
	m.Match(`typeTest($x + $y)`).
		Where(m["x"].Type.Is(`string`) && m["y"].Type.Is("string")).
		Report(`concat`)

	m.Match(`typeTest($x + $y)`).
		Where(m["x"].Type.Is(`string`) && m["y"].Type.Is("string")).
		Report(`concat`)

	m.Match(`typeTest($x + $y)`).
		Where(m["x"].Type.Is(`int`) && m["y"].Type.Is("int")).
		Report(`addition`)

	m.Match(`typeTest($x > $y)`).
		Where(!m["x"].Type.Is(`int`)).
		Report(`$x !is(int)`)

	m.Match(`typeTest($x > $y)`).
		Where(!m["x"].Type.Is(`string`) && m["x"].Pure).
		Report(`$x !is(string) && pure`)

	m.Match(`typeTest($s, $s)`).
		Where(m["s"].Type.Is(`[]string`)).
		Report(`$s is([]string)`)

	m.Match(`pureTest($x)`).
		Where(m["x"].Pure).
		Report("pure")

	m.Match(`pureTest($x)`).
		Where(!m["x"].Pure).
		Report("!pure")
}
