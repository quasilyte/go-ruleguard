//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func rangeRuneSlice(m dsl.Matcher) {
	m.Match(`range []rune($s)`).
		Where(m["s"].Type.Is(`string`)).
		Suggest(`range $s`)
}
