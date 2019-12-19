package ruleguard

import (
	"go/types"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
)

type scopedGoRuleSet struct {
	uncategorized   []goRule
	categorizedNum  int
	rulesByCategory [nodeCategoriesCount][]goRule
}

type goRule struct {
	severity string
	pat      *gogrep.Pattern
	msg      string
	location string
	filters  map[string]submatchFilter
}

type submatchFilter struct {
	typePred func(types.Type) bool
	pure     bool3
	constant bool3
}
