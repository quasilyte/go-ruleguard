//go:build ignore
// +build ignore

package main

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	corerules "github.com/quasilyte/go-ruleguard/rules"
	uber "github.com/quasilyte/uber-rules"
)

func init() {
	dsl.ImportRules("", corerules.Bundle)
	dsl.ImportRules("", uber.Bundle)
}
