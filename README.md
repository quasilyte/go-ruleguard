# go-ruleguard

[analysis](https://godoc.org/golang.org/x/tools/go/analysis)-based Go linter that runs dynamically loaded rules.

You write the rules, `ruleguard` checks whether they are satisfied.

* No re-compilations is required. It also doesn't use plugins.
* Diagnostics (rules) are written in a declarative way.

## Overview

TODO.

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
package gorules

import . "github.com/quasilyte/go-ruleguard/dsl"

func _(x Var) {
	Match(
		`$x || $x`,
		`$x && $x`,
	)
	Filter(x.Pure)
	Error(`suspicious identical LHS and RHS`)
}

func _() {
	// It's possible to write several match-filter-yield sequences
	// inside one rule function.

	Match(`!($x != $y)`)
	Hint(`can simplify !($x==$y) to $x!=$y`)

	Match(`!($x == $y)`)
	Hint(`can simplify !($x==$y) to $x!=$y`)
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

## Documentation

* [Example rules.go file](analyzer/testdata/src/gocritic/gocritic.rules.go)

## Extra references

* [gogrep](https://github.com/mvdan/gogrep) - underlying AST matching engine
* [NoVerify: Dynamic Rules for Static Analysis](https://medium.com/@vktech/noverify-dynamic-rules-for-static-analysis-8f42859e9253)
