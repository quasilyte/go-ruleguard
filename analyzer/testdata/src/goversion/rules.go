// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

// This test is executed with go=1.16

func ioutilDeprecated(m dsl.Matcher) {
	m.Match(`ioutil.ReadAll($r)`).
		Where(m.GoVersion().GreaterEqThan("1.16")).
		Suggest(`io.ReadAll($r)`).
		Report(`ioutil.ReadAll is deprecated, use io.ReadAll instead`)
}

func versionConstraints(m dsl.Matcher) {
	m.Match(`test("<2.0 && >1.0")`).Where(m.GoVersion().LessThan("2.0") && m.GoVersion().GreaterThan("1.0")).Report(`true`)
	m.Match(`test("<1.10 && >1.90")`).Where(m.GoVersion().LessThan("2.0") && m.GoVersion().GreaterThan("1.90")).Report(`false`)

	m.Match(`test("<=1.15")`).Where(m.GoVersion().LessEqThan("1.15")).Report(`false`)
	m.Match(`test("<=1.16")`).Where(m.GoVersion().LessEqThan("1.16")).Report(`true`)
	m.Match(`test("<=1.17")`).Where(m.GoVersion().LessEqThan("1.17")).Report(`true`)

	m.Match(`test(">=1.15")`).Where(m.GoVersion().GreaterEqThan("1.15")).Report(`true`)
	m.Match(`test(">=1.16")`).Where(m.GoVersion().GreaterEqThan("1.16")).Report(`true`)
	m.Match(`test(">=1.17")`).Where(m.GoVersion().GreaterEqThan("1.17")).Report(`false`)

	m.Match(`test(">1.15")`).Where(m.GoVersion().GreaterThan("1.15")).Report(`true`)
	m.Match(`test(">1.17")`).Where(m.GoVersion().GreaterThan("1.17")).Report(`false`)

	m.Match(`test("<1.17")`).Where(m.GoVersion().LessThan("1.17")).Report(`true`)
	m.Match(`test("<1.16")`).Where(m.GoVersion().LessThan("1.16")).Report(`false`)

	m.Match(`test("==1.16")`).Where(m.GoVersion().Eq("1.16")).Report(`true`)
	m.Match(`test("==1.17")`).Where(m.GoVersion().Eq("1.17")).Report(`false`)
	m.Match(`test("!=1.16")`).Where(!m.GoVersion().Eq("1.16")).Report(`false`)
}
