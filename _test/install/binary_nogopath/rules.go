// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	"github.com/quasilyte/go-ruleguard/dsl/types"
	testrules "github.com/quasilyte/ruleguard-rules-test"
	subtestrules "github.com/quasilyte/ruleguard-rules-test/sub2"
)

func init() {
	dsl.ImportRules("", testrules.Bundle)
	dsl.ImportRules("", subtestrules.Bundle)
}

func isInterface(ctx *dsl.VarFilterContext) bool {
	// Could be written as m["x"].Type.Underlying().Is(`interface{$*_}`) in DSL.
	return types.AsInterface(ctx.Type.Underlying()) != nil
}

func exprUnparen(m dsl.Matcher) {
	m.Match(`$f($*_, ($x), $*_)`).
		Report(`the parentheses around $x are superfluous`).
		Suggest(`$f($x)`)
}

func interfaceAddr(m dsl.Matcher) {
	m.Match(`&$x`).
		Where(m["x"].Filter(isInterface)).
		Report(`taking address of interface-typed value`)
}
