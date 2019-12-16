# gorules documentation

## Overview

`gorules` is a configuration file format for a `ruleguard` program.

The proposed filename for these files is `<name>.rules.go`, where `<name>` is up to you.

## Syntax

Go syntax is used, although `gorules` files are never executed or involved into any kind of `go build`.
It's therefore recommended to add a `// +build ignore` comment to avoid any issues during the build.

The advantage of a Go-compatible syntax is having convenient tooling working for rule files.
The downside is that it makes rule files slightly more verbose.

## Structure

Every `gorules` file is a valid Go file:

* It has a package clause. By convenience, we use `gorules` package name, but
  it doesn't have any effect right now.
* An import clause (you need at least `github.com/quasilyte/go-ruleguard/dsl`).
* Every **rule** lives inside a **rule group**. A rule group is a function that is defined
  on a top level.
* Every **rule** has mandatory **match** and **yield** clauses. There can also
  be an optional **filter** clause.
* Every clause if a special function call.

Here is a small yet useful, example of `gorules` file:

```go
// +build ignore

package gorules

import . "github.com/quasilyte/go-ruleguard/dsl"

func _(m MatchResult) {
	Match(
		`regexp.Compile($pat)`,
		`regexp.CompilePOSIX($pat)`,
	)
	Filter(m["pat"].Const)
	Hint(`can use MustCompile for const patterns`)
}
```

> Note: right now it's impossible not to use dot-import for a `dsl`, but it can be fixed in future.

A rule group that has `_` function name is called anonymous. You can have as much anonymous groups as you like.

The `MatchResult`, `Match`, `Filter` and `Warn` symbols are defined in the [dsl](https://github.com/quasilyte/go-ruleguard/blob/master/dsl/dsl.go) package.

* `Match()` is for **match clause**,
* `Filter()` is for **filter clause** and
* `Hint()` is for **yield clause**.

You can also use `Error`, `Warn` and `Info` functions in **yield clause**. They control the severity level of a produced report.

A warning message can use `$<varname>` to interpolate the named pattern submatches into the report message.
There is a special case of `$$` which can be used to inject the entire pattern match into the message.

## Filters

Right now there are only variable-based filters.

A variable describes a named submatch of a pattern.

Here are some examples of supported filters:
* Submatch expression type is identical to `T`
* Submatch expression type is assignable to `T`
* Submatch expression is side-effect free
* Submatch expression is a const expression

Please refer to the godoc page of a `dsl` package to get an up-to-date documentation on what filters are supported.
