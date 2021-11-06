// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	bothConst := func(x, y dsl.Var) bool {
		return x.Const && y.Const
	}
	m.Match(`test("both const", $x, $y)`).
		Where(bothConst(m["x"], m["y"])).
		Report(`true`)

	intValue := func(x dsl.Var, val int) bool {
		return x.Value.Int() == val
	}
	m.Match(`test("== 10", $x)`).
		Where(intValue(m["x"], 10)).
		Report(`true`)

	isZero := func(x dsl.Var) bool { return x.Value.Int() == 0 }
	m.Match(`test("== 0", $x)`).
		Where(isZero(m["x"])).
		Report(`true`)

	// Testing closure-captured m variable.
	fmtIsImported := func() bool {
		return m.File().Imports(`fmt`)
	}
	m.Match(`test("fmt is imported")`).
		Where(fmtIsImported()).
		Report(`true`)

	// Testing explicitly passed matcher.
	ioutilIsImported := func(m2 dsl.Matcher) bool {
		return m2.File().Imports(`io/ioutil`)
	}
	m.Match(`test("ioutil is imported")`).
		Where(ioutilIsImported(m)).
		Report(`true`)

	// Test precedence after the macro expansion.
	isSimpleExpr := func(v dsl.Var) bool {
		return v.Const || v.Node.Is(`Ident`)
	}
	m.Match(`test("check precedence", $x, $y)`).
		Where(isSimpleExpr(m["x"]) && m["y"].Text == "err").
		Report(`true`)

	// Test variables referenced through captured m.
	isStringLit := func() bool {
		return m["x"].Node.Is(`BasicLit`) && m["x"].Type.Is(`string`)
	}
	m.Match(`test("is string", $x)`).
		Where(isStringLit()).
		Report(`true`)

	// Test predicate applied to different matcher vars.
	isPureCall := func(v dsl.Var) bool {
		return v.Node.Is(`CallExpr`) && v.Pure
	}
	m.Match(`test("is pure call", $x, $y)`).
		Where(isPureCall(m["x"]) && isPureCall(m["y"])).
		Report(`true`)
}
