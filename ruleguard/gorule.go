package ruleguard

import (
	"go/types"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
	"github.com/quasilyte/go-ruleguard/ruleguard/typematch"
)

type scopedGoRuleSet struct {
	uncategorized   []goRule
	categorizedNum  int
	rulesByCategory [nodeCategoriesCount][]goRule
}

type goRule struct {
	severity   string
	pat        *gogrep.Pattern
	msg        string
	location   string
	suggestion string
	filters    map[string]submatchFilter
}

type submatchFilter struct {
	typePred func(typeMatchingContext) bool
	pure     bool3
	constant bool3
}

type typeMatchingContext struct {
	typ types.Type
	env *typematch.Env
}
