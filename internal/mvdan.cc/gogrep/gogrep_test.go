package gogrep

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWildNameEncDec(t *testing.T) {
	tests := []string{
		"_",
		"x",
		"foo",
		"Foo_Bar",
	}

	for _, any := range []bool{true, false} {
		for _, name := range tests {
			want := varInfo{Name: name, Any: any}
			enc := encodeWildName(name, any)
			dec := decodeWildName(enc)
			if diff := cmp.Diff(dec, want); diff != "" {
				t.Errorf("decode diff (+want -have):\n%s", diff)
				continue
			}
		}
	}
}

func TestCapture(t *testing.T) {
	type vars = map[string]string

	tests := []struct {
		pat     string
		input   string
		capture map[string]string
	}{
		{
			`$x + $y`,
			`1 + 2`,
			vars{"x": "1", "y": "2"},
		},

		{
			`f($*args)`,
			`f(1, 2, 3)`,
			vars{"args": "1, 2, 3"},
		},

		{
			`f($x, $*tail)`,
			`f(1)`,
			vars{"tail": "", "x": "1"},
		},
		{
			`f($x, $*tail)`,
			`f(1, 2)`,
			vars{"tail": "2", "x": "1"},
		},
		{
			`f($x, $*tail)`,
			`f(1, 2, 3)`,
			vars{"tail": "2, 3", "x": "1"},
		},

		{
			`f($left, $*mid, $right)`,
			`f(1, 2)`,
			vars{"left": "1", "mid": "", "right": "2"},
		},
		// TODO: #192
		// {
		// 	`f($left, $*mid, $right)`,
		// 	`f(1, 2, 3)`,
		// 	vars{"left": "1", "mid": "2", "right": "3"},
		// },
		// {
		// 	`f($left, $*mid, $right)`,
		// 	`f(1, 2, 3, 4)`,
		// 	vars{"left": "1", "mid": "2, 3", "right": "4"},
		// },

		{
			`f($*butlast, "last")`,
			`f("last")`,
			vars{"butlast": ""},
		},
		{
			`f($*butlast, "last")`,
			`f(1, "last")`,
			vars{"butlast": "1"},
		},
		{
			`f($*butlast, "last")`,
			`f(1, 2, "last")`,
			vars{"butlast": "1, 2"},
		},
		{
			`f($*v, "x", "y")`,
			`f(1, 2, "x", "y")`,
			vars{"v": "1, 2"},
		},

		{
			`f($*butlast, $x)`,
			`f(1)`,
			vars{"butlast": "", "x": "1"},
		},
		// TODO: #192
		// {
		// 	`f($*butlast, $x)`,
		// 	`f(1, 2, 3)`,
		// 	vars{"butlast": "1, 2", "x": "3"},
		// },
		// {
		// 	`f($first, $*butlast, $x)`,
		// 	`f(1, 2, 3)`,
		// 	vars{"first": "1", "butlast": "2", "x": "2"},
		// },
	}

	emptyFset := token.NewFileSet()
	sprintNode := func(n ast.Node) string {
		var buf strings.Builder
		testPrintNode(&buf, emptyFset, n)
		return buf.String()
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			fset := token.NewFileSet()
			pat, err := Parse(fset, test.pat)
			if err != nil {
				t.Errorf("parse `%s`: %v", test.pat, err)
				return
			}
			target := testParseNode(t, test.input)
			if err != nil {
				t.Errorf("parse target `%s`: %v", test.input, err)
				return
			}
			capture := vars{}
			pat.MatchNode(target, func(m MatchData) {
				for k, n := range m.Values {
					capture[k] = sprintNode(n)
				}
			})
			if diff := cmp.Diff(capture, test.capture); diff != "" {
				t.Errorf("test `%s`:\ntarget: `%s`\ndiff (+want -have)\n%s", test.pat, test.input, diff)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		pat        string
		numMatches int
		input      string
	}{
		{`123`, 1, `123`},
		{`false`, 0, `true`},

		{`$x`, 1, `rune`},
		{`foo($x, $x)`, 0, `foo(1, 2)`},
		{`foo($_, $_)`, 1, `foo(1, 2)`},
		{`foo($x, $y, $y)`, 1, `foo(1, 2, 2)`},
		{`$x`, 1, `"foo"`},

		{`$x`, 3, `a + b`},
		{`$x + $x`, 0, `foo(a) + foo(b)`},
		{`$x + $x`, 1, `foo(a) + foo(a)`},
		{`$x`, 5, `var a int`},
		{`go foo()`, 1, `go foo()`},

		{`$x; $y`, 1, `{1; 2}`},
		{`go foo()`, 1, `for { a(); go foo(); a() }`},

		// many value expressions
		{`$x, $y`, 1, `foo(1, 2)`},
		{`2, 3`, 1, `foo(1, 2, 3)`},
		{`$x, $x, $x`, 1, `foo(1, 1, 1)`},
		{`$x, $x, $x`, 1, `foo(2, 1, 1, 1)`},
		{`$x, $x, $x`, 1, `foo(1, 1, 1, 2)`},
		{`$x, $x, $x`, 1, `foo(2, 1, 1, 1, 2)`},
		{`$x, $x, $x`, 1, `[]int{1, 1, 1, 2}`},
		{`$x, $x, $x`, 1, `[]int{2, 1, 1, 1}`},
		{`$x, $x, $x`, 1, `[]int{2, 1, 1, 1, 2}`},
		{`$x, $y`, 0, `1`},
		{`$x`, 5, `[]string{a, b}`},
		{`$x, $x`, 1, `return 1, 1`},
		{`$x, $x`, 1, `return 0, 1, 0, 1, 1`},
		{`$x, $x`, 1, `return 0, 1, 1, 0`},

		{`b, c`, 1, `[]int{a, b, c, d}`},
		{`b, c`, 1, `foo(a, b, c, d)`},
		{`print($*_, $x)`, 1, `print(a, b, c)`},

		// any number of expressions
		// {`$*x`, 1, `f(a, b)`},
		{`print($*x)`, 1, `print()`},
		{`print($*x)`, 1, `print(a, b)`},
		{`print($*x, $y, $*z)`, 0, `print()`},
		{`print($*x, $y, $*z)`, 1, `print(a)`},
		{`print($*x, $y, $*z)`, 1, `print(a, b, c)`},
		{`{ $*_; return nil }`, 1, `{ return nil }`},
		{`{ $*_; return nil }`, 1, `{ a(); b(); return nil }`},
		{`c($*x); c($*x)`, 1, `{ c(); c() }`},
		{`c($*x); c()`, 1, `{ c(); c() }`},
		{`c($*x); c($*x)`, 0, `if cond { c(x); c(y) }`},
		{`c($*x); c($*x)`, 0, `if cond { c(x, y); c(z) }`},
		{`c($*x); c($*x)`, 1, `if cond { c(x, y); c(x, y) }`},
		{`c($*x, y); c($*x, y)`, 1, `if cond { c(x, y); c(x, y) }`},
		{`c($*x, $*y); c($*x, $*y)`, 1, `{ c(x, y); c(x, y) }`},

		// composite lits
		{`[]float64{$x}`, 1, `[]float64{3}`},
		{`[2]bool{$x, 0}`, 0, `[2]bool{3, 1}`},
		{`someStruct{fld: $x}`, 0, `someStruct{fld: a, fld2: b}`},
		{`map[int]int{1: $x}`, 1, `map[int]int{1: a}`},

		// func lits
		{`func($s string) { print($s) }`, 1, `func(a string) { print(a) }`},
		{`func($x ...$t) {}`, 1, `func(a ...int) {}`},

		// type exprs
		{`[8]$x`, 1, `[8]int{4: 1}`},
		{`struct{field $t}`, 1, `struct{field int}{}`},
		{`struct{field $t}`, 1, `struct{field int}{}`},
		{`struct{field $t}`, 0, `struct{other int}{}`},
		{`struct{field $t}`, 0, `(struct{f1, f2 int}{})`},
		{`interface{$x() int}`, 1, `(interface{i() int})(nil)`},
		{`chan $x`, 1, `new(chan bool)`},
		{`<-chan $x`, 0, `make(chan bool)`},
		{`chan $x`, 0, `(chan<- bool)(nil)`},

		// parens
		{`($x)`, 1, `(a + b)`},
		{`($x)`, 0, `a + b`},

		// unary ops
		{`-someConst`, 1, `- someConst`},
		{`*someVar`, 1, `* someVar`},
		{`-someConst`, 0, `someConst`},
		{`*someVar`, 0, `someVar`},

		// binary ops
		{`$x == $y`, 1, `a == b`},
		{`$x == $y`, 0, `123`},
		{`$x == $y`, 0, `a != b`},
		{`$x - $x`, 0, `a - b`},

		// calls
		{`someFunc($x)`, 1, `someFunc(a > b)`},

		// selector
		{`$x.Field`, 1, `a.Field`},
		{`$x.Field`, 0, `a.field`},
		{`$x.Method()`, 1, `a.Method()`},
		{`a.b`, 1, `a.b.c`},
		{`b.c`, 0, `a.b.c`},
		{`$x.c`, 1, `a.b.c`},
		{`a.$x`, 1, `a.b.c`},

		// indexes
		{`$x[len($x)-1]`, 1, `a[len(a)-1]`},
		{`$x[len($x)-1]`, 0, `a[len(b)-1]`},

		// slicing
		{`$x[:$y]`, 1, `a[:1]`},
		{`$x[3:]`, 0, `a[3:5:5]`},

		// type asserts
		{`$x.(string)`, 1, `a.(string)`},

		// key-value expression
		{`"a": 1`, 1, `map[string]int{"a": 1}`},

		// elipsis
		{`append($x, $y...)`, 1, `append(a, bs...)`},
		{`foo($x...)`, 0, `foo(a)`},
		{`foo($x...)`, 0, `foo(a, b)`},

		// forcing node to be a statement
		{`append($*_);`, 0, `{ f(); x = append(x, a) }`},
		{`append($*_);`, 1, `{ f(); append(x, a) }`},

		// many statements
		{`$x(); $y()`, 1, `{ a(); b() }`},
		{`$x(); $y()`, 0, `{ a() }`},
		{`$x`, 5, `{a; b}`},
		{`b; c`, 0, `{b}`},
		{`b; c`, 1, `{b; c}`},
		{`b; c`, 0, `{b; x; c}`},
		{`b; c`, 1, `{ a; b; c; d }`},
		{`b; c`, 1, `{b; c; d}`},
		{`b; c`, 1, `{a; b; c}`},
		{`b; c`, 1, `{b; b; c; c}`},
		{`$x++; $x--`, 1, `{ n; a++; b++; b-- }`},
		{`$*_; b; $*_`, 1, `{a; b; c; d}`},
		{`{$*_; $x}`, 1, `{a; b; c}`},
		{`{b; c}`, 0, `{a; b; c}`},
		{`$x := $_; $x = $_`, 1, `{ a := n; b := n; b = m }`},
		{`$x := $_; $*_; $x = $_`, 1, `{ a := n; b := n; b = m }`},

		// mixing lists
		{`$x, $y`, 0, `{ 1; 2 }`},
		{`$x; $y`, 0, `f(1, 2)`},

		// any number of statements
		// {`$*x`, 1, `{ a; b }`},
		{`$*x; b; $*y`, 1, `{ a; b; c }`},
		{`$*x; b; $*x`, 0, `{ a; b; c }`},

		// const/var declarations
		{`const $x = $y`, 1, `const a = b`},
		{`const $x = $y`, 1, `const (a = b)`},
		{`const $x = $y`, 0, "const (a = b\nc = d)"},
		{`var $x int`, 1, `var a int`},
		{`var $x int`, 0, `var a int = 3`},
		{`var ()`, 1, `var()`},
		{`var ()`, 0, `var(x int)`},

		// func declarations
		{
			`func $_($x $y) $y { return $x }`,
			1,
			`package p; func a(i int) int { return i }`,
		},
		{`func $x(i int)`, 1, `package p; func a(i int)`},
		{`func $x(i int) {}`, 0, `package p; func a(i int)`},
		{
			`func $_() $*_ { $*_ }`,
			1,
			`package p; func f() {}`,
		},
		{
			`func $_() $*_ { $*_ }`,
			1,
			`package p; func f() (int, error) { return 3, nil }`,
		},

		// type declarations
		{`struct{}`, 1, `type T struct{}`},
		{`type $x struct{}`, 1, `type T struct{}`},
		{`struct{$_ int}`, 1, `type T struct{n int}`},
		{`struct{$_ int}`, 1, `var V struct{n int}`},
		{`struct{$_}`, 1, `type T struct{n int}`},
		{`struct{$*_}`, 1, `type T struct{n int}`},
		{
			`struct{$*_; Foo $t; $*_}`,
			1,
			`type T struct{Foo string; a int; B}`,
		},
		// structure literal
		{`struct{a int}{a: $_}`, 1, `struct{a int}{a: 1}`},
		{`struct{a int}{a: $*_}`, 1, `struct{a int}{a: 1}`},

		// value specs
		{`$_ int`, 1, `var a int`},
		{`$_ int`, 0, `var a bool`},
		// TODO: consider these
		{`$_ int`, 0, `var a int = 3`},
		{`$_ int`, 0, `var a, b int`},
		{`$_ int`, 0, `func(i int) { println(i) }`},

		// entire files
		{`package $_`, 0, `package p; var a = 1`},
		{`package $_; func Foo() { $*_ }`, 1, `package p; func Foo() {}`},

		// blocks
		{`{ $x }`, 1, `{ a() }`},
		{`{ $x }`, 0, `{ a(); b() }`},
		{`{}`, 1, `package p; func f() {}`},

		// assigns
		{`$x = $y`, 1, `a = b`},
		{`$x := $y`, 0, `a, b := c()`},

		// if stmts
		{`if $x != nil { $y }`, 1, `if p != nil { p.foo() }`},
		{`if $x { $y }`, 0, `if a { b() } else { c() }`},
		{`if $x != nil { $y }`, 1, `if a != nil { return a }`},

		// for and range stmts
		{`for $x { $y }`, 1, `for b { c() }`},
		{`for $x := range $y { $z }`, 1, `for i := range l { c() }`},
		{`for $x := range $y { $z }`, 0, `for i = range l { c() }`},
		{`for $x = range $y { $z }`, 0, `for i := range l { c() }`},
		{`for range $y { $z }`, 0, `for _, e := range l { e() }`},

		// $*_ matching stmt+expr combos (ifs)
		{`if $*x {}`, 1, `if a {}`},
		{`if $*x {}`, 1, `for { if a(); b {} }`},
		{`if $*x {}; if $*x {}`, 1, `for cond() { if a(); b {}; if a(); b {} }`},
		{`if $*x {}; if $*x {}`, 0, `for cond() { if a(); b {}; if b {} }`},
		{`if $*_ {} else {}`, 0, `if a(); b {}`},
		{`if $*_ {} else {}`, 1, `if a(); b {} else {}`},
		{`if a(); $*_ {}`, 0, `if b {}`},

		// $*_ matching stmt+expr combos (fors)
		{`for $*x {}`, 1, `for {}`},
		{`for $*x {}`, 1, `for a {}`},
		{`for $*x {}`, 1, `for i(); a; p() {}`},
		{`for $*x {}; for $*x {}`, 1, `if ok { for i(); a; p() {}; for i(); a; p() {} }`},
		{`for $*x {}; for $*x {}`, 0, `if ok { for i(); a; p() {}; for i(); b; p() {} }`},
		{`for a(); $*_; {}`, 0, `for b {}`},
		{`for ; $*_; c() {}`, 0, `for b {}`},

		// $*_ matching stmt+expr combos (switches)
		{`switch $*x {}`, 1, `switch a {}`},
		{`switch $*x {}`, 1, `switch a(); b {}`},
		{`switch $*x {}; switch $*x {}`, 1, `{ switch a(); b {}; switch a(); b {} }`},
		{`switch $*x {}; switch $*x {}`, 0, `{ switch a(); b {}; switch b {} }`},
		{`switch a(); $*_ {}`, 0, `for b {}`},

		// $*_ matching stmt+expr combos (node type mixing)
		{`if $*x {}; for $*x {}`, 1, `{ if a(); b {}; for a(); b; {} }`},
		{`if $*x {}; for $*x {}`, 0, `{ if a(); b {}; for a(); b; c() {} }`},

		// for $*_ {} matching a range for
		{`for $_ {}`, 0, `for range x {}`},
		{`for $*_ {}`, 1, `for range x {}`},
		{`for $*_ {}`, 1, `for _, v := range x {}`},

		// $*_ matching optional statements (ifs)
		{`if $*_; b {}`, 1, `if b {}`},
		{`if $*_; b {}`, 1, `if a := f(); b {}`},
		// TODO: should these match?
		//{`if a(); $*x { f($*x) }`, `if a(); b { f(b) }`, 1},
		//{`if a(); $*x { f($*x) }`, `if a(); b { f(b, c) }`, 0},
		//{`if $*_; $*_ {}`, `if a(); b {}`, 1},

		// $*_ matching optional statements (fors)
		{`for $*x; b; $*x {}`, 1, `for b {}`},
		{`for $*x; b; $*x {}`, 1, `for a(); b; a() {}`},
		{`for $*x; b; $*x {}`, 0, `for a(); b; c() {}`},

		// $*_ matching optional statements (switches)
		{`switch $*_; b {}`, 1, `switch b := f(); b {}`},
		{`switch $*_; b {}`, 0, `switch b := f(); c {}`},

		// inc/dec stmts
		{`$x++`, 1, `a[b]++`},
		{`$x--`, 0, `a++`},

		// returns
		{`return nil, $x`, 1, `{ return nil, err }`},
		{`return nil, $x`, 0, `{ return nil, 0, err }`},

		// go stmts
		{`go $x()`, 1, `go func() { a() }()`},
		{`go func() { $x }()`, 1, `go func() { a() }()`},
		{`go func() { $x }()`, 0, `go a()`},

		// defer stmts
		{`defer $x()`, 1, `defer func() { a() }()`},
		{`defer func() { $x }()`, 1, `defer func() { a() }()`},
		{`defer func() { $x }()`, 0, `defer a()`},

		// empty statement
		{`;`, 1, `;`},

		// labeled statement
		{`foo: if x {}`, 1, `foo: if x {}`},
		{`foo: if x {}`, 0, `foo: if y {}`},

		// send statement
		{`x <- 1`, 1, `x <- 1`},
		{`x <- 1`, 0, `y <- 1`},
		{`x <- 1`, 0, `x <- 2`},

		// branch statement
		{`break foo`, 1, `break foo`},
		{`break foo`, 0, `break bar`},
		{`break foo`, 0, `continue foo`},
		{`break foo`, 0, `break`},
		{`break`, 1, `break`},

		// case clause
		{`switch x {case 4: x}`, 1, `switch x {case 4: x}`},
		{`switch x {case 4: x}`, 0, `switch y {case 4: x}`},
		{`switch x {case 4: x}`, 0, `switch x {case 5: x}`},
		{`switch {$_}`, 1, `switch {case 5: x}`},
		{`switch x {$_}`, 1, `switch x {case 5: x}`},
		{`switch x {$*_}`, 1, `switch x {case 5: x}`},
		{`switch x {$*_}`, 1, `switch x {}`},
		{`switch x {$*_}`, 1, `switch x {case 1: a; case 2: b}`},
		{`switch {$a; $a}`, 1, `switch {case true: a; case true: a}`},
		{`switch {$a; $a}`, 0, `switch {case true: a; case true: b}`},

		// switch statement
		{`switch x; y {}`, 1, `switch x; y {}`},
		{`switch x {}`, 0, `switch x; y {}`},
		{`switch {}`, 1, `switch {}`},
		{`switch {}`, 0, `switch x {}`},
		{`switch {}`, 0, `switch {case y:}`},
		{`switch $_ {}`, 1, `switch x {}`},
		{`switch $_ {}`, 0, `switch x; y {}`},
		{`switch $_; $_ {}`, 0, `switch x {}`},
		{`switch $_; $_ {}`, 1, `switch x; y {}`},
		{`switch { $*_; case $*_: $*a }`, 0, `switch { case x: y() }`},

		// type switch statement
		{`switch x := y.(z); x {}`, 1, `switch x := y.(z); x {}`},
		{`switch x := y.(z); x {}`, 0, `switch y := y.(z); x {}`},
		{`switch x := y.(z); x {}`, 0, `switch y := y.(z); x {}`},
		{`switch x := $x.(type) {}`, 1, `switch x := y.(type) {}`},
		{`switch x := $x.(type) {}`, 1, `switch x := xs[0].(type) {}`},
		{`switch x := $x.(type) {}`, 0, `{}`},
		{`switch $x.(type) {}`, 1, `switch v.(type) {}`},
		// TODO more switch variations.

		// TODO select statement
		// TODO communication clause
		{`select {$*_}`, 1, `select {case <-x: a}`},
		{`select {$*_}`, 1, `select {}`},
		{`select {$a; $a}`, 1, `select {case <-x: a; case <-x: a}`},
		{`select {$a; $a}`, 0, `select {case <-x: a; case <-x: b}`},
		{`select {case x := <-y: f(x)}`, 1, `select {case x := <-y: f(x)}`},

		{
			`if len($xs) != 0 { for _, $x = range $xs { $*_ } }`,
			1,
			`if len(xs) != 0 { for _, v = range xs { println(v) } }`,
		},
		{
			`if len($xs) != 0 { for _, $x := range $xs { $*_ } }`,
			0,
			`if len(xs) != 0 { for _, v = range xs { println(v) } }`,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			fset := token.NewFileSet()
			pat, err := Parse(fset, test.pat)
			if err != nil {
				t.Errorf("parse `%s`: %v", test.pat, err)
				return
			}
			target := testParseNode(t, test.input)
			if err != nil {
				t.Errorf("parse target `%s`: %v", test.input, err)
				return
			}
			matches := 0
			testAllMatches(pat, target, func(m MatchData) {
				matches++
			})
			if matches != test.numMatches {
				t.Errorf("test `%s`:\ntarget: `%s`\nhave: %v\nwant: %v",
					test.pat, test.input, matches, test.numMatches)
			}
		})
	}
}

func testAllMatches(p *Pattern, target ast.Node, cb func(MatchData)) {
	ast.Inspect(target, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if _, ok := p.Expr.(stmtList); ok {
			switch n := n.(type) {
			case *ast.BlockStmt:
				p.MatchStmtList(n.List, cb)
			case *ast.CaseClause:
				p.MatchStmtList(n.Body, cb)
			case *ast.CommClause:
				p.MatchStmtList(n.Body, cb)
			}
		}
		if _, ok := p.Expr.(exprList); ok {
			switch n := n.(type) {
			case *ast.CallExpr:
				p.MatchExprList(n.Args, cb)
			case *ast.CompositeLit:
				p.MatchExprList(n.Elts, cb)
			case *ast.ReturnStmt:
				p.MatchExprList(n.Results, cb)
			}
		}
		p.MatchNode(n, cb)
		return true
	})
}

func testParseNode(t *testing.T, s string) ast.Node {
	if strings.HasPrefix(s, "package ") {
		file, err := parser.ParseFile(token.NewFileSet(), "string", s, 0)
		if err != nil {
			t.Fatalf("parse `%s`: %v", s, err)
		}
		return file
	}
	source := `package p; func _() { ` + s + ` }`
	file, err := parser.ParseFile(token.NewFileSet(), "string", source, 0)
	if err != nil {
		t.Fatalf("parse `%s`: %v", s, err)
	}
	fn := file.Decls[0].(*ast.FuncDecl)
	n := fn.Body.List[0]
	if e, ok := n.(*ast.ExprStmt); ok {
		return e.X
	}
	return n
}

func testPrintNode(w io.Writer, fset *token.FileSet, node ast.Node) {
	switch x := node.(type) {
	case exprList:
		if len(x) == 0 {
			return
		}
		testPrintNode(w, fset, x[0])
		for _, n := range x[1:] {
			fmt.Fprintf(w, ", ")
			testPrintNode(w, fset, n)
		}
	case stmtList:
		if len(x) == 0 {
			return
		}
		testPrintNode(w, fset, x[0])
		for _, n := range x[1:] {
			fmt.Fprintf(w, "; ")
			testPrintNode(w, fset, n)
		}
	default:
		err := printer.Fprint(w, fset, node)
		if err != nil && strings.Contains(err.Error(), "go/printer: unsupported node type") {
			// Should never happen, but make it obvious when it does.
			panic(fmt.Errorf("cannot print node %T: %v", node, err))
		}
	}
}
