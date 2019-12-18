# go-ruleguard

[![Build Status](https://travis-ci.com/quasilyte/go-ruleguard.svg?branch=master)](https://travis-ci.com/quasilyte/go-ruleguard)
[![GoDoc](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl?status.svg)](https://godoc.org/github.com/quasilyte/go-ruleguard)
[![Go Report Card](https://goreportcard.com/badge/github.com/quasilyte/go-ruleguard)](https://goreportcard.com/report/github.com/quasilyte/go-ruleguard)

![Logo](docs/logo_small.png)

## Overview

[analysis](https://godoc.org/golang.org/x/tools/go/analysis)-based Go linter that runs dynamically loaded rules.

You write the rules, `ruleguard` checks whether they are satisfied.

**Features:**

* Custom linting rules without re-compilation and Go plugins.
* Diagnostics are written in a declarative way.
* Powerful match filtering features, like expression type pattern matching.

## Quick start

To install `ruleguard` binary under your `$(go env GOPATH)/bin`:

```bash
$ go get -v github.com/quasilyte/go-ruleguard/cmd/ruleguard
```

If `$GOPATH/bin` is under your system `$PATH`, `ruleguard` command should be available after that.<br>

```bash
$ ruleguard -help
ruleguard: execute dynamic gogrep-based rules

Usage: ruleguard [-flag] [package]

Flags:
  -rules string
    	path to a rules.go file
  -c int
    	display offending line with this many lines of context (default -1)
  -json
    	emit JSON output
```

Create a test `example.rules.go` file:

```go
// +build ignore

package gorules

import . "github.com/quasilyte/go-ruleguard/dsl"

func _(m MatchResult) {
	Report(`suspicious identical LHS and RHS`).
		Matches(
			`$x || $x`,
			`$x && $x`,
		).
		Where(m["x"].Pure)

	Report(`can simplify !($x==$y) to $x!=$y`).Matches(`!($x != $y)`)
	Report(`can simplify !($x==$y) to $x!=$y`).Matches(`!($x == $y)`)
}
```

Create a test `example.go` target file:

```go
package main

func main() {
	var v1, v2 int
	println(!(v1 != v2))
	println(!(v1 == v2))
	if v1 == 0 && v1 == 0 {
		println("hello, world!")
	}
}
```

Run `ruleguard` on that target file:

```
$ ruleguard -rules example.rules.go example.go
example.go:5:10: hint: can simplify !(v1!=v2) to v1==v2
example.go:6:10: hint: can simplify !(v1==v2) to v1!=v2
example.go:7:5: error: suspicious identical LHS and RHS
```

## How does it work?

`ruleguard` parses [gorules](docs/gorules.md) during the start to load the rule set.  
Instantiated rules are then used to check the specified targets (Go files, packages).

A rule definition starts from a call to a special function `Report` that takes a single string argument - a template for a report message on a successful match.

Second mandatory part is a `Matches` method call which contains a list of [gogrep](https://github.com/mvdan/gogrep) patterns that are used to match a relevant part of a Go program. If there is a match and all filter conditions are satisfied, a rule report is produced.

Above mentioned filter conditions can be added by a call to a `Where` method, which applies constraints to a match to decide whether its accepted or rejected.

To learn more, check out the documentation and/or the source code.

## Documentation

* [Example rules.go file](analyzer/testdata/src/gocritic/gocritic.rules.go)
* [rules.go](docs/gorules.md) documentation
* [dsl package](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl) reference
* [ruleguard package](https://godoc.org/github.com/quasilyte/go-ruleguard/ruleguard) reference

## Extra references

* [gogrep](https://github.com/mvdan/gogrep) - underlying AST matching engine
* [NoVerify: Dynamic Rules for Static Analysis](https://medium.com/@vktech/noverify-dynamic-rules-for-static-analysis-8f42859e9253)
