package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

func exprUnparen(m dsl.Matcher) {
	m.Match(`$f($*_, ($x), $*_)`).
		Report(`the parentheses around $x are superfluous`).
		Suggest(`$f($x)`)
}

func emptyDecl(m dsl.Matcher) {
	m.Match(`var()`).Report(`empty var() block`)
	m.Match(`const()`).Report(`empty const() block`)
	m.Match(`type()`).Report(`empty type() block`)
}

func emptyError(m dsl.Matcher) {
	m.Match(`fmt.Errorf("")`, `errors.New("")`).
		Report(`empty errors are hard to debug`)
}
