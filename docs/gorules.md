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

Every `gorules` file is a valid Go file.

We can describe a file structure like this:

1. It has a package clause (package name should be `gorules`).
2. An import clause (you need at least `github.com/quasilyte/go-ruleguard/dsl/fluent`).
3. Function declarations.

Functions play a special role: they serve as a **rule groups**.

Every function accepts exactly 1 arguement, a [`fluent.Matcher`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher), and defines some **rules**.

Every **rule** definition starts with a [`Match`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Match) method call that specifies one or more [AST patterns](https://github.com/mvdan/gogrep) that should represent what kind of Go code rule supposed to match. Another mandatory method is [`Report`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Report) that describes a message template that is going to be printed when the rule match is accepted.

Here is a small yet useful, example of `gorules` file:

```go
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
	m.Match(`regexp.Compile($pat)`,
		`regexp.CompilePOSIX($pat)`).
		Where(m["pat"].Const).
		Report(`can use MustCompile for const patterns`)
}
```

A rule group that has `_` function name is called anonymous. You can have as much anonymous groups as you like.

A `Report` argument string can use `$<varname>` notation to interpolate the named pattern submatches into the report message.
There is a special case of `$$` which can be used to inject the entire pattern match into the message.

## Filters

Right now there are only match variable-based filters that can be added with a [`Where`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Where) call.

A match variable describes a named submatch of a pattern.

Here are some examples of supported filters:
* Submatch expression type is identical to `T`
* Submatch expression type is assignable to `T`
* Submatch expression is side-effect free
* Submatch expression is a const expression

A match variable can be accessed with `fluent.Matcher` function argument indexing:

```go
Filter(m["a"].Type.Is(`int`) && !m["b"].Type.AssignableTo(`[]string`))
```

If we had a pattern with `$a` and `$b` match variables, a filter above would only accept it
if `$a` expression had a type of `int` while `$b` is anything that is **not** assignable to `[]string`.

The filter concept is crucial to avoid false-positives in rules.

Please refer to the godoc page of a [`dsl/fluent`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent) package to get an up-to-date list of supported filters.

## Type pattern matching

Methods like [`ExprType.Is()`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#ExprType.Is) accept a string argument that describes a Go type. It can be as simple as `"[]string"` that matches only a string slice, but it can also include a pattern-like variables:

* `[]$T` matches any slice.
* `[$len]$T` matches any array.
* `map[$K]$V` matches any map.
* `map[$T]$T` matches a map where a key and value types are the same.

You may recognize that it's the same pattern behavior as in AST patterns.
