package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

var Bundle = dsl.Bundle{}

//doc:summary reports always false/true conditions
//doc:before  strings.Count(s, "/") >= 0
//doc:after   strings.Count(s, "/") > 0
//doc:tags    diagnostic
func badCond(m dsl.Matcher) {
	m.Match(`strings.Count($_, $_) >= 0`).Report(`statement always true`)
	m.Match(`bytes.Count($_, $_) >= 0`).Report(`statement always true`)
}
