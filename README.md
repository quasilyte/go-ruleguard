# go-ruleguard

[![Build Status](https://travis-ci.com/quasilyte/go-ruleguard.svg?branch=master)](https://travis-ci.com/quasilyte/go-ruleguard)
[![GoDoc](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl?status.svg)](https://godoc.org/github.com/quasilyte/go-ruleguard)
[![Go Report Card](https://goreportcard.com/badge/github.com/quasilyte/go-ruleguard)](https://goreportcard.com/report/github.com/quasilyte/go-ruleguard)

![Logo](docs/logo_small.png)

## Overview

[analysis](https://godoc.org/golang.org/x/tools/go/analysis)-based Go linter that runs dynamically loaded rules.

You write the rules, `ruleguard` checks whether they are satisfied.

**Features:**

* No re-compilations is required. It also doesn't use plugins.
* Diagnostics (rules) are written in a declarative way.
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
	Error(`suspicious identical LHS and RHS`,
		Match(
			`$x || $x`,
			`$x && $x`,
		),
		Filter(m["x"].Pure),
	)
	
	Hint(`can simplify !($x==$y) to $x!=$y`, Match(`!($x != $y)`))
	Hint(`can simplify !($x==$y) to $x!=$y`, Match(`!($x == $y)`))
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

A rule is defined by a call to a special function: `Error`, `Warn`, `Info` or `Hint`.  
The only difference is a severity of the report message.

Such function takes a report message template string as well as a list of clauses.

Right now we have these clauses:
1. **match clause** contains a [gogrep](https://github.com/mvdan/gogrep) pattern that is used to match a part of a Go program.
2. **where clause** (optional) applies constraints to a match in order to decide whether its accepted or rejected.

**where** constraints are usually type-based, but can also include properties like "an expression is side-effect free".

To learn more, check out the documentation and/or the source code.

## Documentation

* [Example rules.go file](analyzer/testdata/src/gocritic/gocritic.rules.go)
* [rules.go](docs/gorules.md) documentation
* [dsl package](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl) reference
* [ruleguard package](https://godoc.org/github.com/quasilyte/go-ruleguard/ruleguard) reference

## Extra references

* [gogrep](https://github.com/mvdan/gogrep) - underlying AST matching engine
* [NoVerify: Dynamic Rules for Static Analysis](https://medium.com/@vktech/noverify-dynamic-rules-for-static-analysis-8f42859e9253)
