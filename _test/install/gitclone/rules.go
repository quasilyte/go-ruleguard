// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	"github.com/quasilyte/go-ruleguard/dsl/types"
)

func implementsStringer(ctx *dsl.VarFilterContext) bool {
	stringer := ctx.GetInterface(`fmt.Stringer`)
	return types.Implements(ctx.Type, stringer) ||
		types.Implements(types.NewPointer(ctx.Type), stringer)
}

func exprUnparen(m dsl.Matcher) {
	m.Match(`$f($*_, ($x), $*_)`).
		Report(`the parentheses around $x are superfluous`).
		Suggest(`$f($x)`)
}

func sprintStringer(m dsl.Matcher) {
	m.Match(`fmt.Sprint($x)`).
		Where(m["x"].Filter(implementsStringer) && m["x"].Addressable).
		Report(`can use $x.String() directly`)
}
