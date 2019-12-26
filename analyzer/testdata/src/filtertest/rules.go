// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
	m.Import(`github.com/quasilyte/go-ruleguard/analyzer/testdata/src/filtertest/foolib`)

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

	m.Match(`typeTest("2 type filters", $x)`).
		Where(!m["x"].Type.Is(`string`) && !m["x"].Type.Is(`int`)).
		Report(`$x !is(string) && !is(int)`)

	m.Match(`typeTest($x, "implements io.Reader")`).
		Where(m["x"].Type.Implements(`io.Reader`)).Report(`YES`)
	m.Match(`typeTest($x, "implements foolib.Stringer")`).
		Where(m["x"].Type.Implements(`foolib.Stringer`)).Report(`YES`)

	m.Match(`typeTest($x, "size>=100")`).Where(m["x"].Type.Size >= 100).Report(`YES`)
	m.Match(`typeTest($x, "size<=100")`).Where(m["x"].Type.Size <= 100).Report(`YES`)
	m.Match(`typeTest($x, "size>100")`).Where(m["x"].Type.Size > 100).Report(`YES`)
	m.Match(`typeTest($x, "size<100")`).Where(m["x"].Type.Size < 100).Report(`YES`)
	m.Match(`typeTest($x, "size==100")`).Where(m["x"].Type.Size == 100).Report(`YES`)
	m.Match(`typeTest($x, "size!=100")`).Where(m["x"].Type.Size != 100).Report(`YES`)

	m.Match(`pureTest($x)`).
		Where(m["x"].Pure).
		Report("pure")

	m.Match(`pureTest($x)`).
		Where(!m["x"].Pure).
		Report("!pure")
}
