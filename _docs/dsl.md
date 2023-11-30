# Ruleguard DSL documentation

## Overview

Ruleguard takes special Go files as its configuration. These files define custom rules and have strict structure that is described in this document.

The advantage of a Go-compatible syntax is having convenient tooling working for ruleguard files.

The Go code from these files is never compiled by `go build` and/or executed by `go run`. Ruleguard parses these files and creates a special internal representation that can be used to efficiently execute them on the fly.

You write the rules, ruleguard tries to execute them precisely and efficiently.

## Ruleguard file structure

We can describe a file structure like this:

1. It has a package clause (package name should be `gorules`).
2. An import clause (at the bare minimum, you'll need [`dsl`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl) package).
3. Function declarations.

There are 3 kinds of functions you can declare:

1. Matcher functions that define a **rule group**
2. Custom filter functions
3. `init()` function

### Matcher functions

Every **matcher function** accepts exactly 1 argument, a [`dsl.Matcher`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher), and defines some **rules**.

Every **rule** definition starts with a [`Match()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.Match) or [`MatchComment()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.MatchComment) method call.

* For `Match()`, you specify one or more [AST patterns](https://github.com/quasilyte/gogrep) that should represent what kind of Go code a rule is supposed to match.
* For `MatchComment()`, you provide one or more regular expressions that should match a comment of interest.

Another mandatory part is [`Report()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.Report) or [`Suggest()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.Suggest) that describe a rule match action. `Report()` will print a warning message while `Suggest()` can be used to provide a quickfix action (a syntax rewrite pattern).

Here is a small yet useful, example of ruleguard file:

```go
package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func regexpMust(m dsl.Matcher) {                                  //        - regexpMust matcher func
	m.Match(`regexp.Compile($pat)`,                           // - rule  | (or "regexpMust rules group")
		`regexp.CompilePOSIX($pat)`).                     //  |      |
		Where(m["pat"].Const).                            //  |      |
		Report(`can use MustCompile for const patterns`). //  |      |
		Suggest(`regexp.MustCompile($pat)`)               // -       |
}                                                                 //        -
```

A `Report()` argument string can use `$<varname>` notation to interpolate the named pattern submatches into the report message.

There is a special variable `$$` which can be used to inject the entire pattern match into the message (like `$0` in regular expressions).

### Documenting your rules

It's a good practice to add structured documentation for your rule groups.

To add such documentation, use special pragmas when commenting a matcher function.

```go
//doc:summary reports always false/true conditions
//doc:before  strings.Count(s, "/") >= 0
//doc:after   strings.Count(s, "/") > 0
//doc:tags    diagnostic experimental
func badCond(m dsl.Matcher) {
	m.Match(`strings.Count($_, $_) >= 0`).Report(`statement always true`)
	m.Match(`bytes.Count($_, $_) >= 0`).Report(`statement always true`)
}
```

* `summary` - short one sentence description
* `before` - code snippet of code that will violate rule
* `after` - code after a fix (one that complies to the rule)
* `tags` - space separated list of custom tags
* `note` - extra information, like issue links

### Filters

The rule is matched if:

1. At least 1 AST pattern from `Match()` is matched
2. Filters from `Where()` accept the given match

There are 2 types of filters that can be used in [`Where()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.Where) call:

1. Submatch (named variable-based) filters
2. Context filters (current file, etc)

A match variable describes a named submatch of a pattern.

Here are some examples of supported filters:

* Submatch expression type is identical to `T`
* Submatch expression type is assignable to `T`
* Submatch expression type implements interface `I`
* Submatch expression is side-effect free
* Submatch expression is a const expression
* Submatch expression const value check
* Submatch text matches provided regexp
* Current files imports package `P`

A match variable can be accessed with `dsl.Matcher` function argument indexing:

```go
// m["a"] -- $a
// m["b"] -- $b
Where(m["a"].Type.Is(`int`) && !m["b"].Type.AssignableTo(`[]string`))
```

If we had a pattern with `$a` and `$b` match variables, a filter above would only accept it
if `$a` expression had a type of `int` while `$b` is anything that is **not** assignable to `[]string`.

Context-related filters can be applied through `m` members:
```go
// Using m.File() to apply a file-related filter.
Where(m.File().Imports("io/ioutil"))
```

When using `MatchComment`, submatches will have a type of `*ast.Comment`. Text-related filters can be used as usual.

The filter concept is crucial to avoid false-positives in rules.

Please refer to the godoc page of a [`dsl`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl) package to get an up-to-date list of supported filters.

## Custom filters

When none of the DSL filters seem to do what you want, you can write a custom filter function.

```go
package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	"github.com/quasilyte/go-ruleguard/dsl/types"
)

func implementsStringer(ctx *dsl.VarFilterContext) bool {
	stringer := ctx.GetInterface(`fmt.Stringer`)
	return types.Implements(ctx.Type, stringer) ||
		types.Implements(types.NewPointer(ctx.Type), stringer)
}

func stringerLiteral(m dsl.Matcher) {
	m.Match(`$x{$*_}`).
		Where(m["x"].Filter(implementsStringer)).
		Report("$x implements stringer")
}
```

Suppose that we have this Go file we want to check:

```go
package target

type byValue struct{}
type byPtr struct{}

func (*byPtr) String() string  { return "" }
func (byValue) String() string { return "" }

func f() {
	_ = &byValue{}
	_ = byValue{}
	_ = &byPtr{}
	_ = byPtr{} // does not implement fooer, but we still want it
}
```

We'll get this output if we apply a `stringerLiteral` rule:

```
example.go:10:7: stringerLiterals: byValue implements stringer
example.go:11:6: stringerLiterals: byValue implements stringer
example.go:12:7: stringerLiterals: byPtr implements stringer
example.go:13:6: stringerLiterals: byPtr implements stringer
```

Custom filter functions are byte-compiled and interpreted like a scripting language. There are some limitations in the implementations; if you would like to see some feature to be implemented, please [tell about it](https://github.com/quasilyte/go-ruleguard/issues/new).

## Named types and import tables

When you use a type filter, the parser must know how to match a given type against the type that is going to be found during execution.

In normal Go programs, the unqualified type name like `Foo` makes sense, it refers to a current package symbol table.

In ruleguard files, you either have to [`Import()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Matcher.Import) the package and use the qualified name or use a fully-qualified name without import.

Here are two ways to use a `Bar` type from the `foo/bar` package in filters:

```go
func qualifiedName(m dsl.Matcher) {
	m.Import(`foo/bar`)
	m.Match(`f($x)`).Where(m["x"].Type.Is(`bar.Baz`)).Report("$x is bar.Baz")
}

func fullyQualifiedName(m dsl.Matcher) {
	m.Match(`f($x)`).
		Where(m["x"].Type.Is(`foo/bar.Baz`)).
		Report("$x is bar.Baz")
}
```

For convenience, [stdlib packages](https://gist.github.com/quasilyte/2bbe64a0ec92c217d8e5f534d9781fcf) are pre-loaded into the imports table.

There are some collisions in the standard library, like `text/template` and `html/template`. By default, `template` is imported as `text/template`, but if you want `template.Template` to refer to the HTML package template, you can override the default imports by doing an explicit import:

```go
m.Import(`html/template`)
// Now template.Template refers to a type from html/template package.
```

### Type pattern matching

Methods like [`ExprType.Is()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#ExprType.Is) accept a string argument that describes a Go type. It can be as simple as `"[]string"` that matches only a string slice, but it can also include a pattern-like variables:

* `[]$T` matches any slice.
* `[$len]$T` matches any array.
* `map[$K]$V` matches any map.
* `map[$T]$T` matches a map where a key and value types are the same.
* `struct{$*_}` any struct type.
* `struct{$x; $*_}` struct that has $x-typed first field.
* `struct{$*_; $x; $*_}` struct that contains $x-typed field.
* `struct{$*_; $x}` struct that has $x-typed last field.

Note: when matching types, make sure to think whether you need to match a type or the **underlying type**.
To match the underlying type, use [`ExprType.Underlying()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#ExprType.Underlying) method.

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

Be careful when using `Suggest()` with `MatchComment()`. As regexp may match a subset of the comment, you'll replace that exact comment portion with `Suggest()` pattern. If you want to replace an entire comment, be sure that your pattern contains `^` and `$` anchors.

## Ruleguard bundles

If you want to use a ruleguard file that is written by someone else, you have 2 main options:

1. Copy the file and keep it somewhere inside your repository
2. Import that ruleguard file as a bundle

The latter option gives you extra benefits:

* Versioning is easier. You can pin a bundle version in your `go.mod` file
* You can add your own rules without a risk of running into collisions

### Installing bundles

Bundles are installed with `go get`.

Here is how we can install the [github.com/quasilyte/ruleguard-rules-test](https://github.com/quasilyte/ruleguard-rules-test) bundle:

```bash
# Make sure that Go modules are turned on.
export GO111MODULE=on

go get -v -u github.com/quasilyte/ruleguard-rules-test@master
```

If your ruleguard file has a `// +build ignore` build tag, `go get` would install a bundle as **indirect** dependency. You'll probably want to use a [tools.go](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module) idiom to make it a direct dependency.

```go
// +build tools

package tools

import (
	_ "github.com/quasilyte/go-ruleguard/dsl"
)
```

If `rules.go` file has no ignore tag, `go get` should work properly without extra efforts.

Ruleguard bundle becomes your project explicit dependency that you can use in your ruleguard files. Note that it does not make your project bloated: bundle packages are never used in your builds.

In case if you don't want to have a direct bundle dependency, run a `go get` before running a ruleguard and then remove the installed package (`go mod tidy` will be enough if bundle package is an indirect dependency).

### Importing bundle rules: init() function

Installed bundle packages can be imported as normal Go packages.

Importing the bundle package is not enough to add its rules, you need to call [`ImportRules()`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#ImportRules) function for that.

```go
package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
	quasilyterules "github.com/quasilyte/ruleguard-rules-test"
)

func init() {
	// Imported rules will have a "qrules" prefix.
	dsl.ImportRules("qrules", quasilyterules.Bundle)
}

// Then you can define your own rules.

func emptyStringTest(m dsl.Matcher) {
	m.Match(`len($s) == 0`).
		Where(m["s"].Type.Is("string")).
		Report(`maybe use $s == "" instead?`)

	m.Match(`len($s) != 0`).
		Where(m["s"].Type.Is("string")).
		Report(`maybe use $s != "" instead?`)
}
```

Let's try running that file:

```bash
$ ruleguard -rules rules.go test.go 
test.go:4:6: emptyStringTest: maybe use s == "" instead? (rules.go:13)
test.go:5:6: qrules/boolComparison: omit bool literal in expression (rules1.go:8)
```

It’s possible to use an empty (`""`) prefix, but you’ll risk getting a name collision. If you don’t define your own rules, then it’s perfectly fine to use an empty prefix.

### Creating a ruleguard bundle

A package that exports rules must define a [`Bundle`](https://pkg.go.dev/github.com/quasilyte/go-ruleguard/dsl#Bundle) object:

```go
// Bundle holds the rules package metadata.
//
// In order to be importable from other gorules package,
// a package must define a Bundle variable.
var Bundle = dsl.Bundle{}
```

That package should be a separate [Go module](https://github.com/golang/go/wiki/Modules). A rules bundle is versioned by its Go module.

It's possible to have several ruleguard files inside one Go module. Only one file should define a Bundle object. During a bundle import, all files will be exported.
