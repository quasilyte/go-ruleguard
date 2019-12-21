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

A rule group that has `_` function name is called **anonymous**. You can have as much anonymous groups as you like.

A `Report` argument string can use `$<varname>` notation to interpolate the named pattern submatches into the report message.
There is a special case of `$$` which can be used to inject the entire pattern match into the message.

## Rule group statements

Apart from the rules, function can contain group statements.

As everything else, statements are `Matcher` methods. [`Import()`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Import) is one of these special methods.

Rule group statements only affect the current rule group and last from the line they were defined until the end of a function block.

```go
func _(m fluent.Matcher) {
	// <- Empty imports table.

	m.Import(`github.com/some/pkg`)
	// <- "github.com/some/pkg" is loaded into the imports table.

} // <- End of the group statements effect
```

`Import()` is explained further in this document.

> There were plans to permit methods like `Import()` inside rule defining chains, making their effect rule-local,
> but since there is no request for such feature yet, we stick to the group-local approach for now.

## Filters

Right now there are only match variable-based filters that can be added with a [`Where`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Where) call.

A match variable describes a named submatch of a pattern.

Here are some examples of supported filters:
* Submatch expression type is identical to `T`
* Submatch expression type is assignable to `T`
* Submatch expression type implements interface `I`
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

## Named types and import tables

When you use a type filter, the rule parser must know how to match a given type against the type that is going to be found during execution.

In normal Go programs, the unqualified type name like `Foo` makes sense, it refers to a current package symbol table. In the simplest case `Foo` is a type defined inside that package. It can also be coming from another package if dot import is used.

In `gorules`, unqualified type name is hard to interpret right. We could use the same logic as in normal Go programs, but why would you need to define a type filter that matches a local package type? It could be different for every package being analyzed. You usually want to use a specific type constraints that do not depend on the current package.

Our resolution is to reject all the unqualified names. If you want a `Foo` type from `a/b/c` package, you need to:

1. Do an [`Import("a/b/c")`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#Matcher.Import) call, so the package is loaded into the current imports table.
2. Use `c.Foo` type name.

We need the `Import()` step to match `c` package name with its path, `a/b/c`.

For convenience, [stdlib packages](https://gist.github.com/quasilyte/2bbe64a0ec92c217d8e5f534d9781fcf) are pre-imported into the table. There are some collisions in the standard library, like `text/template` and `html/template`. By default, `template` is imported as `text/template`, but if you want `template.Template` to refer to the HTML package template, you can override the default imports by doing an explicit import:

```go
m.Import(`html/template`)
// Now template.Template refers to a type from html/template package.
```

## Type pattern matching

Methods like [`ExprType.Is()`](https://godoc.org/github.com/quasilyte/go-ruleguard/dsl/fluent#ExprType.Is) accept a string argument that describes a Go type. It can be as simple as `"[]string"` that matches only a string slice, but it can also include a pattern-like variables:

* `[]$T` matches any slice.
* `[$len]$T` matches any array.
* `map[$K]$V` matches any map.
* `map[$T]$T` matches a map where a key and value types are the same.

You may recognize that it's the same pattern behavior as in AST patterns.

## Suggestions (quickfix support)

Some rules have a clear suggestion that can be used as a direct replacement of a matched code.

Consider this rule:

```go
m.Match(`!!$x`).Report(`can simplify !!$x to $x`)
```

It contains a fix inside its message. The user will have to fix it manually, by copy/pasting the
suggestion from the report message. We can do better.

```go
m.Match(`!!$x`).Suggest(`$x`)
```

Now we're enabled the `-fix` support for that rule. If `ruleguard` is invoked with that argument,
code from the `Suggest` argument will replace the matched code chunk.

You can have both `Report()` and `Suggest()`, so the user can also have a more detailed
warning message when not using `-fix`.

When you use `Suggest()` and omit `Report()`, suggested string is used as a foundation of a report message.  
The following 2 lines are identical:

```go
m.Match(`!!$x`).Suggest(`$x`)
m.Match(`!!$x`).Suggest(`$x`).Report(`suggested: $x`)
```
