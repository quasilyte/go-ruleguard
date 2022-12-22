# go-ruleguard

![Build Status](https://github.com/quasilyte/go-ruleguard/workflows/Go/badge.svg)
![Build Status](https://github.com/quasilyte/go-ruleguard/workflows/Merge/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/quasilyte/go-ruleguard)](https://pkg.go.dev/mod/github.com/quasilyte/go-ruleguard)

![Logo](_docs/logo2.png)

## Overview

[analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)-based Go linter that runs dynamically loaded rules.

You write the rules, `ruleguard` checks whether they are satisfied.

`ruleguard` has some similarities with [GitHub CodeQL](https://securitylab.github.com/tools/codeql), but it's dedicated to Go only.

**Features:**

* Custom linting rules without re-compilation and Go plugins
* Diagnostics are written in a declarative way
* [Quickfix](_docs/dsl.md#suggestions-quickfix-support) actions support
* Powerful match filtering features, like expression [type pattern matching](_docs/dsl.md#type-pattern-matching)
* Not restricted to AST rules; it's possible to write a comment-related rule, for example
* Rules can be installed and distributed as [Go modules](https://quasilyte.dev/blog/post/ruleguard-modules/)
* Rules can be developed, debugged and profiled using the conventional Go tooling
* Integrated into [golangci-lint](https://github.com/golangci/golangci-lint)

It can also be easily embedded into other static analyzers. [go-critic](https://github.com/go-critic/go-critic) can be used as an example.

## Quick start

It's advised that you get a binary from the [latest release](https://github.com/quasilyte/go-ruleguard/releases/tag/v0.3.18) {[linux/amd64](https://github.com/quasilyte/go-ruleguard/releases/download/v0.3.18/ruleguard-linux-amd64.zip), [linux/arm64](https://github.com/quasilyte/go-ruleguard/releases/download/v0.3.18/ruleguard-linux-arm64.zip), [darwin/amd64](https://github.com/quasilyte/go-ruleguard/releases/download/v0.3.18/ruleguard-darwin-amd64.zip), [darwin/arm64](https://github.com/quasilyte/go-ruleguard/releases/download/v0.3.18/ruleguard-darwin-arm64.zip), [windows/amd64](https://github.com/quasilyte/go-ruleguard/releases/download/v0.3.18/ruleguard-windows-amd64.zip), [windows/arm64](https://github.com/quasilyte/go-ruleguard/releases/download/v0.3.18/ruleguard-windows-arm64.zip)}.

If you want to install the ruleguard from source, it's as simple as:

```bash
# Installs a `ruleguard` binary under your `$(go env GOPATH)/bin`
$ go install -v github.com/quasilyte/go-ruleguard/cmd/ruleguard@latest

# Get the DSL package (needed to execute the ruleguard files)
$ go get -v -u github.com/quasilyte/go-ruleguard/dsl@latest
```

> If inside a Go module, the `dsl` package will be installed for the current module,
> otherwise it installs the package into the $GOPATH and it will be globally available.

If `$GOPATH/bin` is under your system `$PATH`, `ruleguard` command should be available after that:

```bash
$ ruleguard -help
ruleguard: execute dynamic gogrep-based rules

Usage: ruleguard [-flag] [package]

Flags:
  -rules string
    	comma-separated list of ruleguard file paths
  -e string
    	execute a single rule from a given string
  -fix
    	apply all suggested fixes
  -c int
    	display offending line with this many lines of context (default -1)
  -json
    	emit JSON output
```

Create a test `rules.go` file:

```go
//go:build ruleguard
// +build ruleguard

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func dupSubExpr(m dsl.Matcher) {
	m.Match(`$x || $x`,
		`$x && $x`,
		`$x | $x`,
		`$x & $x`).
		Where(m["x"].Pure).
		Report(`suspicious identical LHS and RHS`)
}

func boolExprSimplify(m dsl.Matcher) {
	m.Match(`!($x != $y)`).Suggest(`$x == $y`)
	m.Match(`!($x == $y)`).Suggest(`$x != $y`)
}

func exposedMutex(m dsl.Matcher) {
	isExported := func(v dsl.Var) bool {
		return v.Text.Matches(`^\p{Lu}`)
	}

	m.Match(`type $name struct { $*_; sync.Mutex; $*_ }`).
		Where(isExported(m["name"])).
		Report("do not embed sync.Mutex")

	m.Match(`type $name struct { $*_; sync.RWMutex; $*_ }`).
		Where(isExported(m["name"])).
		Report("do not embed sync.RWMutex")
}
```

Create a test `example.go` target file:

```go
package main

import "sync"

type EmbedsMutex struct {
	key int
	sync.Mutex
}

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
$ ruleguard -rules rules.go -fix example.go
example.go:5:1: exposedMutex: do not embed sync.Mutex (rules.go:24)
example.go:12:10: boolExprSimplify: suggestion: v1 == v2 (rules.go:15)
example.go:13:10: boolExprSimplify: suggestion: v1 != v2 (rules.go:16)
example.go:14:5: dupSubExpr: suspicious identical LHS and RHS (rules.go:7)
```

Since we ran `ruleguard` with `-fix` argument, both **suggested** changes are applied to `example.go`.

There is also a `-e` mode that is useful during the pattern debugging:

```bash
$ ruleguard -e 'm.Match(`!($x != $y)`)' example.go
example.go:12:10: !(v1 != v2)
```

It automatically inserts `Report("$$")` into the specified pattern.

You can use `-debug-group <name>` flag to see explanations
on why some rules rejected the match (e.g. which `Where()` condition failed).

The `-e` generated rule will have `e` name, so it can be debugged as well.

## How does it work?

First, it parses [ruleguard](_docs/dsl.md) files (e.g. `rules.go`) during the start to load the rule set.  

Loaded rules are then used to check the specified targets (Go files, packages).

The `rules.go` file is written in terms of [`dsl`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl) API. Ruleguard files contain a set of functions that serve as a rule groups. Every such function accepts a single [`dsl.Matcher`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher) argument that is then used to define and configure rules inside the group.

A rule definition always starts with [`Match(patterns...)`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.Match) method call and ends with [`Report(message)`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.Report) method call.

There can be additional calls in between these two. For example, a [`Where(cond)`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.Where) call applies constraints to a match to decide whether it's accepted or rejected. So even if there is a match for a pattern, it won't produce a report message unless it satisfies a `Where()` condition.

## Troubleshooting

For ruleguard to work the `dsl` package must be available at _runtime_. If it's not, you are going to see an error like:

```
$ ruleguard -rules rules.go .
ruleguard: load rules: parse rules file: typechecker error: rules.go:6:8: could not import github.com/quasilyte/go-ruleguard/dsl (can't find import: "github.com/quasilyte/go-ruleguard/dsl")
```

This is fixed by adding the dsl package to the module:

```
$ ruleguard-test go get github.com/quasilyte/go-ruleguard/dsl
go: downloading github.com/quasilyte/go-ruleguard v0.3.18
go: downloading github.com/quasilyte/go-ruleguard/dsl v0.3.21
go: added github.com/quasilyte/go-ruleguard/dsl v0.3.21
$ ruleguard-test ruleguard -rules rules.go .
.../test.go:6:5: boolExprSimplify: suggestion: 1 == 0 (rules.go:9)
```

If you have followed past advise of using a build constraint in the rules.go file, like this:

```
$ ruleguard-test head -4 rules.go
//go:build ignore
// +build ignore

package gorules
```

you are going to notice that `go.mod` file lists the dsl as an indirect dependency:

```
$ grep dsl go.mod
require github.com/quasilyte/go-ruleguard/dsl v0.3.21 // indirect
```

If you run `go mod tidy` now, you are going to notice the dsl package disappears from the `go.mod` file:

```
$ go mod tidy
$ grep dsl go.mod
$ 
```

This is because `go mod tidy` behaves as if all the build constraints are in effect, _with the exception of `ignore`_. [This is documented in the go website](https://go.dev/ref/mod#go-mod-tidy).

This is fixed by using a different build constraint, like `ruleguard` or `rules`.

## Documentation

* [Ruleguard by example](https://go-ruleguard.github.io/by-example/) tour
* [Ruleguard files](_docs/dsl.md) format documentation
* [dsl package](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl) reference
* [ruleguard package](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/ruleguard) reference
* Introduction article: [EN](https://quasilyte.dev/blog/post/ruleguard/), [RU](https://habr.com/ru/post/481696/)
* [Using ruleguard from the golangci-lint](https://quasilyte.dev/blog/post/ruleguard/#using-from-the-golangci-lint)

## Rule set examples

* Basic rule set from the ruleguard: [go-ruleguard/rules](rules)
* [Damian Gryski](github.com/dgryski/) rule set: [github.com/dgryski/semgrep-go/ruleguard.rules.go](https://github.com/dgryski/semgrep-go)
* [go-critic](https://github.com/go-critic/go-critic) rule set: [github.com/go-critic/go-critic/checkers/rules/rules.go](https://github.com/go-critic/go-critic/blob/master/checkers/rules/rules.go)
* Partial [Uber-Go](https://github.com/uber-go/guide) style rule set: [github.com/quasilyte/uber-rules](https://github.com/quasilyte/uber-rules)
* [go-perfguard](https://github.com/quasilyte/go-perfguard) rule set: [github.com/quasilyte/go-perfguard/perfguard/_rules/universal_rules.go](https://github.com/quasilyte/go-perfguard/blob/master/perfguard/_rules/universal_rules.go)

Note: `go-critic` and `go-perfguard` embed the rules using the IR precompilation feature.

## Extra references

* Online ruleguard playground: [go-ruleguard.github.io/play](https://go-ruleguard.github.io/play)
* [gogrep](https://github.com/quasilyte/gogrep) - underlying AST matching engine
* [NoVerify: Dynamic Rules for Static Analysis](https://medium.com/@vktech/noverify-dynamic-rules-for-static-analysis-8f42859e9253)
* [Ruleguard comparison with Semgrep and CodeQL](https://speakerdeck.com/quasilyte/ruleguard-vs-semgrep-vs-codeql)
