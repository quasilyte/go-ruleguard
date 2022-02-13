//go:build ignore
// +build ignore

package gorules

import (
	"fmt"
	"strings"

	"github.com/quasilyte/go-ruleguard/dsl"
	"github.com/quasilyte/go-ruleguard/dsl/types"
)

func reportHello(ctx *dsl.DoContext) {
	ctx.SetReport("Hello, World!")
}

func suggestHello(ctx *dsl.DoContext) {
	ctx.SetSuggest("Hello, World!")
}

func reportX(ctx *dsl.DoContext) {
	ctx.SetReport(ctx.Var("x").Text())
}

func unquote(s string) string {
	return s[1 : len(s)-1]
}

func reportTrimPrefix(ctx *dsl.DoContext) {
	s := unquote(ctx.Var("x").Text())
	prefix := unquote(ctx.Var("y").Text())
	ctx.SetReport(strings.TrimPrefix(s, prefix))
}

func reportEmptyString(ctx *dsl.DoContext) {
	x := ctx.Var("x")
	if x.Text() == `""` {
		ctx.SetReport("empty string")
	} else {
		ctx.SetReport("non-empty string")
	}
}

func reportType(ctx *dsl.DoContext) {
	ctx.SetReport(ctx.Var("x").Type().String())
}

func reportTypesIdentical(ctx *dsl.DoContext) {
	xtype := ctx.Var("x").Type()
	ytype := ctx.Var("y").Type()
	ctx.SetReport(fmt.Sprintf("%v", types.Identical(xtype, ytype)))
}

func testRules(m dsl.Matcher) {
	m.Match(`test("custom report")`).
		Do(reportHello)

	m.Match(`test("custom suggest")`).
		Do(suggestHello)

	m.Match(`test("var text", $x)`).
		Do(reportX)

	m.Match(`test("trim prefix", $x, $y)`).
		Do(reportTrimPrefix)

	m.Match(`test("report empty string", $x)`).
		Do(reportEmptyString)

	m.Match(`test("report type", $x)`).
		Do(reportType)

	m.Match(`test("types identical", $x, $y)`).
		Do(reportTypesIdentical)
}
