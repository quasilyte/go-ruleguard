//go:build ignore
// +build ignore

package main

import (
	rules "github.com/delivery-club/delivery-club-rules"
	"github.com/quasilyte/go-ruleguard/dsl"
	uber "github.com/quasilyte/uber-rules"
)

func init() {
	dsl.ImportRules("", rules.Bundle)
	dsl.ImportRules("", uber.Bundle)
}
