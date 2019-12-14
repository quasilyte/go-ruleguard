// +build ignore

package gorules

import . "github.com/quasilyte/go-ruleguard/dsl"

func _(m MatchResult) {
	Match(`typeTest($x + $y)`)
	Filter(m["x"].Type.Is(`string`) && m["y"].Type.Is("string"))
	Info(`concat`)

	Match(`typeTest($x + $y)`)
	Filter(m["x"].Type.Is(`int`) && m["y"].Type.Is("int"))
	Info(`addition`)

	Match(`typeTest($x > $y)`)
	Filter(!m["x"].Type.Is(`int`))
	Info(`$x !is(int)`)

	Match(`typeTest($x > $y)`)
	Filter(!m["x"].Type.Is(`string`) && m["x"].Pure)
	Info(`$x !is(string) && pure`)

	Match(`typeTest($s, $s)`)
	Filter(m["s"].Type.Is(`[]string`))
	Info(`$s is([]string)`)

	Match(`pureTest($x)`)
	Filter(m["x"].Pure)
	Info("pure")

	Match(`pureTest($x)`)
	Filter(!m["x"].Pure)
	Info("!pure")
}
