// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func shiftOverflow(m dsl.Matcher) {
	m.Match(`$x << $n`).
		Where(!m["x"].Const && m["x"].Type.Size == 1 && m["n"].Value.Int() >= 8 && !m.Deadcode()).
		Report(`$x (8 bits) too small for shift of $n`)

	m.Match(`$x << $n`).
		Where(!m["x"].Const && m["x"].Type.Size == 2 && m["n"].Value.Int() >= 16 && !m.Deadcode()).
		Report(`$x (16 bits) too small for shift of $n`)

	m.Match(`$x << $n`).
		Where(!m["x"].Const && m["x"].Type.Size == 4 && m["n"].Value.Int() >= 32 && !m.Deadcode()).
		Report(`$x (32 bits) too small for shift of $n`)
	
	m.Match(`$x << $n`).
		Where(!m["x"].Const && m["x"].Type.Size == 8 && m["n"].Value.Int() >= 64 && !m.Deadcode()).
		Report(`$x (64 bits) too small for shift of $n`)
}
