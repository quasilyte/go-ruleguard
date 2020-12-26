// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	testrules "github.com/quasilyte/ruleguard-rules-test"
)

func init() {
	dsl.ImportRules("testrules", testrules.Bundle)
}

func exprUnparen(m dsl.Matcher) {
	m.Match(`$f($*_, ($x), $*_)`).
		Report(`the parentheses around $x are superfluous`).
		Suggest(`$f($x)`)
}
