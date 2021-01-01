package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	corerules "github.com/quasilyte/go-ruleguard/rules"
)

func init() {
	dsl.ImportRules("", corerules.Bundle)
}
