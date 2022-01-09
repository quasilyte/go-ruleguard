//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func rangeRuneSlice(m dsl.Matcher) {
	m.Match(`range []rune($s)`).
		Where(m["s"].Type.Is(`string`)).
		Suggest(`range $s`)
}

func writeString(m dsl.Matcher) {
	m.Match(`io.WriteString($w, $s)`).
		Where(m["w"].Type.HasMethod(`io.StringWriter.WriteString`)).
		Suggest(`$w.WriteString($s)`)
}
