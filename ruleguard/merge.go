package ruleguard

import (
	"fmt"
	"go/token"
)

func mergeRuleSets(toMerge []*GoRuleSet) (*GoRuleSet, error) {
	out := &GoRuleSet{
		local:     &scopedGoRuleSet{},
		universal: &scopedGoRuleSet{},
		groups:    make(map[string]token.Position),
		Imports:   make(map[string]struct{}),
	}

	for _, x := range toMerge {
		out.local = appendScopedRuleSet(out.local, x.local)
		out.universal = appendScopedRuleSet(out.universal, x.universal)
		for pkgPath := range x.Imports {
			out.Imports[pkgPath] = struct{}{}
		}
		for group, pos := range x.groups {
			if prevPos, ok := out.groups[group]; ok {
				newRef := fmt.Sprintf("%s:%d", pos.Filename, pos.Line)
				oldRef := fmt.Sprintf("%s:%d", prevPos.Filename, prevPos.Line)
				return nil, fmt.Errorf("%s: redefenition of %s(), previously defined at %s", newRef, group, oldRef)
			}
			out.groups[group] = pos
		}
	}

	return out, nil
}

func appendScopedRuleSet(dst, src *scopedGoRuleSet) *scopedGoRuleSet {
	dst.uncategorized = append(dst.uncategorized, src.uncategorized...)
	for cat, rules := range src.rulesByCategory {
		dst.rulesByCategory[cat] = append(dst.rulesByCategory[cat], rules...)
		dst.categorizedNum += len(rules)
	}
	return dst
}
