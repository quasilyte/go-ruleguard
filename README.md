# go-ruleguard

[analysis](https://godoc.org/golang.org/x/tools/go/analysis)-based Go linter that runs dynamically loaded rules.

No compilation or plugins are needed.

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
    	path to a gorules file
  -c int
    	display offending line with this many lines of context (default -1)
  -json
    	emit JSON output
```

Create a test `example.gorules` file:

```js
// Find suspicious expressions that have duplicated side-effect free LHS and RHS.
//
//error: suspicious identical LHS and RHS
//$x: pure
$x || $x
$x && $x

//hint: can simplify !($x!=$y) to $x==$y
!($x != $y)
//hint: can simplify !($x==$y) to $x!=$y
!($x == $y)
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
$ ruleguard -rules example.gorules example.go
example.go:5:10: hint: can simplify !(v1!=v2) to v1==v2
example.go:6:10: hint: can simplify !(v1==v2) to v1!=v2
example.go:7:5: error: suspicious identical LHS and RHS
```

## References

* [gogrep](https://github.com/mvdan/gogrep)
* [Example rule file](analyzer/testdata/go-critic/go-critic.gorules)
* [NoVerify: Dynamic Rules for Static Analysis](https://medium.com/@vktech/noverify-dynamic-rules-for-static-analysis-8f42859e9253).
