//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func externalErrorReassign(m dsl.Matcher) {
	m.Match(`$pkg.$err = $x`).
		Where(m["err"].Type.Is(`error`) && m["pkg"].Object.Is(`PkgName`)).
		Report(`suspicious reassigment of error from another package`)
}

func largeLoopCopy(m dsl.Matcher) {
	m.Match(
		`for $_, $v := range $_ { $*_ }`,
	).
		Where(m["v"].Type.Size > 512).
		Report(`loop copies large value each iteration`)
}
