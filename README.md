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
* [Quickfix](docs/gorules.md#suggestions-quickfix-support) actions support.
* Powerful match filtering features, like expression [type pattern matching](docs/gorules.md#type-pattern-matching).

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
  -e string
    	execute a single rule from a given string
  -fix
    	apply all suggested fixes
  -c int
    	display offending line with this many lines of context (default -1)
  -json
    	emit JSON output
```

Create a test `example.rules.go` file:

```go
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
	m.Match(`$x || $x`,
		`$x && $x`).
		Where(m["x"].Pure).
		Report(`suspicious identical LHS and RHS`)

	m.Match(`!($x != $y)`).Suggest(`$x == $y`)
	m.Match(`!($x == $y)`).Suggest(`$x != $y`)
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

```bash
$ ruleguard -rules example.rules.go -fix example.go
example.go:5:10: hint: suggested: v1 == v2
example.go:6:10: hint: suggested: v1 != v2
example.go:7:5: error: suspicious identical LHS and RHS
```

Since we ran `ruleguard` with `-fix` argument, both **suggested** changes are applied to `example.go`.

There is also a `-e` mode that is useful during pattern debugging:

```bash
$ ruleguard -e 'm.Match(`!($x != $y)`)' example.go
example.go:5:10: !(v1 != v2)
```

It automatically inserts `Report("$$")` into the specified pattern.

## How does it work?

`ruleguard` parses [gorules](docs/gorules.md) (e.g. `rules.go`) during the start to load the rule set.  
Loaded rules are then used to check the specified targets (Go files, packages).  
The `rules.go` file itself is never compiled, nor executed.

A `rules.go` file, as interpreted by a [`dsl/fluent`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent) API, is a set of functions that serve as a rule groups. Every function accepts a single [`fluent.Matcher`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher) argument that is then used to define and configure rules inside the group.

A rule definition always starts from a [`Match(patterns...)`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Match) method call and ends with a [`Report(message)`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Report) method call.

There can be additional calls in between these two. For example, a [`Where(cond)`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Where) call applies constraints to a match to decide whether its accepted or rejected. So even if there is a match for a pattern, it won't produce a report message unless it satisfies a `Where()` condition.

To learn more, check out the documentation and/or the source code.

## Documentation

* Example rule files: [example1](analyzer/testdata/src/gocritic/rules.go), [example2](https://github.com/quasilyte/go-ruleguard/blob/master/analyzer/testdata/src/extra/rules.go)
* [gorules](docs/gorules.md) format documentation
* [dsl/fluent package](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent) reference
* [ruleguard package](https://godoc.org/github.com/quasilyte/go-ruleguard/ruleguard) reference

## Extra references

* [gogrep](https://github.com/mvdan/gogrep) - underlying AST matching engine
* [NoVerify: Dynamic Rules for Static Analysis](https://medium.com/@vktech/noverify-dynamic-rules-for-static-analysis-8f42859e9253)
