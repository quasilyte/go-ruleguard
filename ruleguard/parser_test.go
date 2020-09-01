package ruleguard

import (
	"bytes"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchParenthesizedExpr(t *testing.T) {
	p := newRulesParser()
	fset := token.NewFileSet()
	rs, err := p.ParseFile("", fset, bytes.NewBufferString(testRules))
	require.NoError(t, err)
	rl := rs.universal.rulesByCategory[nodeCallExpr]
	require.Equal(t, 1, len(rl))
	_, ok := rl[0].filters["a"]
	require.True(t, ok)
}

const testRules = `
// +build ignore

package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl/fluent"
)

func _(m fluent.Matcher) {
	
	// rule for matching github.com/Masterminds/squirrel.SelectBuilder.Where misuses
	m.Match("$_.Where($a,$_)").Where(
		!(m["a"].Type.Is("string") && m["a"].Text.Matches("\\?")) &&
			!m["a"].Type.Is("map[string]interface{}"),
	).
		Report("squirrel.SelectBuilder.Where second parameter is ignored when first cannot " +
			"be transformed to a string containing ? placeholder. " +
			"See https://pkg.go.dev/github.com/Masterminds/squirrel?tab=doc#SelectBuilder.Where")
}
`
