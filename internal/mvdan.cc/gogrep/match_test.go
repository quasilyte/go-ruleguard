// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package gogrep

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"testing"
)

type wantErr string

func tokErr(msg string) wantErr   { return wantErr("cannot tokenize expr: " + msg) }
func modErr(msg string) wantErr   { return wantErr("cannot parse mods: " + msg) }
func parseErr(msg string) wantErr { return wantErr("cannot parse expr: " + msg) }

func TestErrors(t *testing.T) {
	tests := []struct {
		args []string
		want interface{}
	}{

		// expr tokenize errors
		{[]string{"-x", "$"}, tokErr(`1:2: $ must be followed by ident, got EOF`)},
		{[]string{"-x", `"`}, tokErr(`1:1: string literal not terminated`)},
		{[]string{"-x", ""}, parseErr(`empty source code`)},
		{[]string{"-x", "\t"}, parseErr(`empty source code`)},
		{
			[]string{"-x", "$x", "-a", "a"},
			modErr(`1:2: wanted (`),
		},
		{
			[]string{"-x", "$x", "-a", "a("},
			modErr(`1:1: unknown op "a"`),
		},
		{
			[]string{"-x", "$x", "-a", "is(foo)"},
			modErr(`1:4: unknown type: "foo"`),
		},
		{
			[]string{"-x", "$x", "-a", "type("},
			modErr(`1:5: expected ) to close (`),
		},
		{
			[]string{"-x", "$x", "-a", "type({)"},
			modErr(`1:1: expected ';', found '{'`),
		},
		{
			[]string{"-x", "$x", "-a", "type(notType + expr)"},
			modErr(`1:9: expected ';', found '+'`),
		},
		{
			[]string{"-x", "$x", "-a", "comp etc"},
			modErr(`1:6: wanted EOF, got IDENT`),
		},
		{
			[]string{"-x", "$x", "-a", "is(slice) etc"},
			modErr(`1:11: wanted EOF, got IDENT`),
		},

		// expr parse errors
		{[]string{"-x", "foo)"}, parseErr(`1:4: expected statement, found ')'`)},
		{[]string{"-x", "{"}, parseErr(`1:4: expected '}', found 'EOF'`)},
		{[]string{"-x", "$x)"}, parseErr(`1:3: expected statement, found ')'`)},
		{[]string{"-x", "$x("}, parseErr(`1:5: expected operand, found '}'`)},
		{[]string{"-x", "$*x)"}, parseErr(`1:4: expected statement, found ')'`)},
		{[]string{"-x", "a\n$x)"}, parseErr(`2:3: expected statement, found ')'`)},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			grepTest(t, tc.args, "nosrc", tc.want)
		})
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		args []string
		src  string
		want interface{}
	}{
		// basic lits
		{[]string{"-x", "123"}, "123", 1},
		{[]string{"-x", "false"}, "true", 0},

		// wildcards
		{[]string{"-x", "$x"}, "rune", 1},
		{[]string{"-x", "foo($x, $x)"}, "foo(1, 2)", 0},
		{[]string{"-x", "foo($_, $_)"}, "foo(1, 2)", 1},
		{[]string{"-x", "foo($x, $y, $y)"}, "foo(1, 2, 2)", 1},
		{[]string{"-x", "$x"}, `"foo"`, 1},

		// recursion
		{[]string{"-x", "$x"}, "a + b", 3},
		{[]string{"-x", "$x + $x"}, "foo(a + a, b + b)", 2},
		{[]string{"-x", "$x"}, "var a int", 4},
		{[]string{"-x", "go foo()"}, "a(); go foo(); a()", 1},

		// ident regex matches
		{
			[]string{"-x", "$x", "-a", "rx(`foo`)"},
			"bar", 0,
		},
		{
			[]string{"-x", "$x", "-a", "rx(`foo`)"},
			"foo", 1,
		},
		{
			[]string{"-x", "$x", "-a", "rx(`foo`)"},
			"_foo", 0,
		},
		{
			[]string{"-x", "$x", "-a", "rx(`foo`)"},
			"foo_", 0,
		},
		{
			[]string{"-x", "$x", "-a", "rx(`.*foo.*`)"},
			"_foo_", 1,
		},
		{
			[]string{"-x", "$x = $_", "-x", "$x", "-a", "rx(`.*`)"},
			"a = b", 1,
		},
		{
			[]string{"-x", "$x = $_", "-x", "$x", "-a", "rx(`.*`)"},
			"a.field = b", 0,
		},
		{
			[]string{"-x", "$x", "-a", "rx(`.*foo.*`)", "-a", "rx(`.*bar.*`)"},
			"foobar; barfoo; foo; barbar", 2,
		},

		// type equality
		{
			[]string{"-x", "$x", "-a", "type(int)"},
			"var i int", 2, // includes "int" the type
		},
		{
			[]string{"-x", "append($x)", "-x", "$x", "-a", "type([]int)"},
			"var _ = append([]int32{3})", 0,
		},
		{
			[]string{"-x", "append($x)", "-x", "$x", "-a", "type([]int)"},
			"var _ = append([]int{3})", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type([2]int)"},
			"var _ = [...]int{1}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type([2]int)"},
			"var _ = [...]int{1, 2}", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type([2]int)"},
			"var _ = []int{1, 2}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(*int)"},
			"var _ = int(3)", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(*int)"},
			"var _ = new(int)", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(io.Reader)"},
			`import "io"; var _ = io.Writer(nil)`, 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "type(io.Reader)"},
			`import "io"; var _ = io.Reader(nil)`, 1,
		},
		{
			[]string{"-x", "$x", "-a", "type(int)"},
			`type I int; func (i I) p() { print(i) }`, 1,
		},
		{
			[]string{"-x", "$x", "-a", "type(*I)"},
			`type I int; var i *I`, 2,
		},
		// TODO
		// {
		// 	[]string{"-x", "$x", "-a", "type(chan int)"},
		// 	`ch := make(chan int)`, 2,
		// },

		// type assignability
		{
			[]string{"-x", "const _ = $x", "-x", "$x", "-a", "type(int)"},
			"const _ = 3", 0,
		},
		{
			[]string{"-x", "var $x $_", "-x", "$x", "-a", "type(io.Reader)"},
			`import "os"; var f *os.File`, 0,
		},
		{
			[]string{"-x", "var $x $_", "-x", "$x", "-a", "asgn(io.Reader)"},
			`import "os"; var f *os.File`, 1,
		},
		{
			[]string{"-x", "var $x $_", "-x", "$x", "-a", "asgn(io.Writer)"},
			`import "io"; var r io.Reader`, 0,
		},
		{
			[]string{"-x", "var $_ $_ = $x", "-x", "$x", "-a", "asgn(*url.URL)"},
			`var _ interface{} = 0`, 0,
		},
		{
			[]string{"-x", "var $_ $_ = $x", "-x", "$x", "-a", "asgn(*url.URL)"},
			`var _ interface{} = nil`, 1,
		},

		// type conversions
		{
			[]string{"-x", "const _ = $x", "-x", "$x", "-a", "type(int)"},
			"const _ = 3", 0,
		},
		{
			[]string{"-x", "const _ = $x", "-x", "$x", "-a", "conv(int)"},
			"const _ = 3", 1,
		},
		{
			[]string{"-x", "const _ = $x", "-x", "$x", "-a", "conv(int32)"},
			"const _ = 3", 1,
		},
		{
			[]string{"-x", "const _ = $x", "-x", "$x", "-a", "conv([]byte)"},
			"const _ = 3", 0,
		},
		{
			[]string{"-x", "var $x $_", "-x", "$x", "-a", "type(int)"},
			"type I int; var i I", 0,
		},
		{
			[]string{"-x", "var $x $_", "-x", "$x", "-a", "conv(int)"},
			"type I int; var i I", 1,
		},

		// comparable types
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "comp"},
			"var _ = []byte{0}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "comp"},
			"var _ = [...]byte{0}", 1,
		},

		// addressable expressions
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "addr"},
			"var _ = []byte{0}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "addr"},
			"var s struct { i int }; var _ = s.i", 1,
		},

		// underlying types
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(basic)"},
			"var _ = []byte{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(basic)"},
			"var _ = 3", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(basic)"},
			`import "io"; var _ = io.SeekEnd`, 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(array)"},
			"var _ = []byte{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(array)"},
			"var _ = [...]byte{}", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(slice)"},
			"var _ = []byte{}", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(slice)"},
			"var _ = [...]byte{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(struct)"},
			"var _ = []byte{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(struct)"},
			"var _ = struct{}{}", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(interface)"},
			"var _ = struct{}{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(interface)"},
			"var _ = interface{}(nil)", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(pointer)"},
			"var _ = []byte{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(pointer)"},
			"var _ = new(byte)", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(func)"},
			"var _ = []byte{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(func)"},
			"var _ = func() {}", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(map)"},
			"var _ = []byte{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(map)"},
			"var _ = map[int]int{}", 1,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(chan)"},
			"var _ = []byte{}", 0,
		},
		{
			[]string{"-x", "var _ = $x", "-x", "$x", "-a", "is(chan)"},
			"var _ = make(chan int)", 1,
		},

		// many value expressions
		{[]string{"-x", "$x, $y"}, "foo(1, 2)", 1},
		{[]string{"-x", "$x, $y"}, "1", 0},
		{[]string{"-x", "$x"}, "a, b", 3},
		// unlike statements, expressions don't automatically
		// imply partial matches
		{[]string{"-x", "b, c"}, "a, b, c, d", 0},
		{[]string{"-x", "b, c"}, "foo(a, b, c, d)", 0},
		{[]string{"-x", "print($*_, $x)"}, "print(a, b, c)", 1},

		// any number of expressions
		{[]string{"-x", "$*x"}, "a, b", "a, b"},
		{[]string{"-x", "print($*x)"}, "print()", 1},
		{[]string{"-x", "print($*x)"}, "print(a, b)", 1},
		{[]string{"-x", "print($*x, $y, $*z)"}, "print()", 0},
		{[]string{"-x", "print($*x, $y, $*z)"}, "print(a)", 1},
		{[]string{"-x", "print($*x, $y, $*z)"}, "print(a, b, c)", 1},
		{[]string{"-x", "{ $*_; return nil }"}, "{ return nil }", 1},
		{[]string{"-x", "{ $*_; return nil }"}, "{ a(); b(); return nil }", 1},
		{[]string{"-x", "c($*x); c($*x)"}, "c(); c()", 1},
		{[]string{"-x", "c($*x); c()"}, "c(); c()", 1},
		{[]string{"-x", "c($*x); c($*x)"}, "c(x); c(y)", 0},
		{[]string{"-x", "c($*x); c($*x)"}, "c(x, y); c(z)", 0},
		{[]string{"-x", "c($*x); c($*x)"}, "c(x, y); c(x, y)", 1},
		{[]string{"-x", "c($*x, y); c($*x, y)"}, "c(x, y); c(x, y)", 1},
		{[]string{"-x", "c($*x, $*y); c($*x, $*y)"}, "c(x, y); c(x, y)", 1},

		// composite lits
		{[]string{"-x", "[]float64{$x}"}, "[]float64{3}", 1},
		{[]string{"-x", "[2]bool{$x, 0}"}, "[2]bool{3, 1}", 0},
		{[]string{"-x", "someStruct{fld: $x}"}, "someStruct{fld: a, fld2: b}", 0},
		{[]string{"-x", "map[int]int{1: $x}"}, "map[int]int{1: a}", 1},

		// func lits
		{[]string{"-x", "func($s string) { print($s) }"}, "func(a string) { print(a) }", 1},
		{[]string{"-x", "func($x ...$t) {}"}, "func(a ...int) {}", 1},

		// type exprs
		{[]string{"-x", "[8]$x"}, "[8]int", 1},
		{[]string{"-x", "struct{field $t}"}, "struct{field int}", 1},
		{[]string{"-x", "struct{field $t}"}, "struct{field int}", 1},
		{[]string{"-x", "struct{field $t}"}, "struct{other int}", 0},
		{[]string{"-x", "struct{field $t}"}, "struct{f1, f2 int}", 0},
		{[]string{"-x", "interface{$x() int}"}, "interface{i() int}", 1},
		{[]string{"-x", "chan $x"}, "chan bool", 1},
		{[]string{"-x", "<-chan $x"}, "chan bool", 0},
		{[]string{"-x", "chan $x"}, "chan<- bool", 0},

		// many types (TODO; revisit)
		// {[]string{"-x", "chan $x, interface{}"}, "chan int, interface{}", 1},
		// {[]string{"-x", "chan $x, interface{}"}, "chan int", 0},
		// {[]string{"-x", "$x string, $y int"}, "func(s string, i int) {}", 1},

		// parens
		{[]string{"-x", "($x)"}, "(a + b)", 1},
		{[]string{"-x", "($x)"}, "a + b", 0},

		// unary ops
		{[]string{"-x", "-someConst"}, "- someConst", 1},
		{[]string{"-x", "*someVar"}, "* someVar", 1},

		// binary ops
		{[]string{"-x", "$x == $y"}, "a == b", 1},
		{[]string{"-x", "$x == $y"}, "123", 0},
		{[]string{"-x", "$x == $y"}, "a != b", 0},
		{[]string{"-x", "$x - $x"}, "a - b", 0},

		// calls
		{[]string{"-x", "someFunc($x)"}, "someFunc(a > b)", 1},

		// selector
		{[]string{"-x", "$x.Field"}, "a.Field", 1},
		{[]string{"-x", "$x.Field"}, "a.field", 0},
		{[]string{"-x", "$x.Method()"}, "a.Method()", 1},
		{[]string{"-x", "a.b"}, "a.b.c", 1},
		{[]string{"-x", "b.c"}, "a.b.c", 0},
		{[]string{"-x", "$x.c"}, "a.b.c", 1},
		{[]string{"-x", "a.$x"}, "a.b.c", 1},

		// indexes
		{[]string{"-x", "$x[len($x)-1]"}, "a[len(a)-1]", 1},
		{[]string{"-x", "$x[len($x)-1]"}, "a[len(b)-1]", 0},

		// slicing
		{[]string{"-x", "$x[:$y]"}, "a[:1]", 1},
		{[]string{"-x", "$x[3:]"}, "a[3:5:5]", 0},

		// type asserts
		{[]string{"-x", "$x.(string)"}, "a.(string)", 1},

		// elipsis
		{[]string{"-x", "append($x, $y...)"}, "append(a, bs...)", 1},
		{[]string{"-x", "foo($x...)"}, "foo(a)", 0},
		{[]string{"-x", "foo($x...)"}, "foo(a, b)", 0},

		// forcing node to be a statement
		{[]string{"-x", "append($*_);"}, "f(); x = append(x, a)", 0},
		{[]string{"-x", "append($*_);"}, "f(); append(x, a)", 1},

		// many statements
		{[]string{"-x", "$x(); $y()"}, "a(); b()", 1},
		{[]string{"-x", "$x(); $y()"}, "a()", 0},
		{[]string{"-x", "$x"}, "a; b", 3},
		{[]string{"-x", "b; c"}, "b", 0},
		{[]string{"-x", "b; c"}, "b; c", 1},
		{[]string{"-x", "b; c"}, "b; x; c", 0},
		{[]string{"-x", "b; c"}, "a; b; c; d", "b; c"},
		{[]string{"-x", "b; c"}, "{b; c; d}", 1},
		{[]string{"-x", "b; c"}, "{a; b; c}", 1},
		{[]string{"-x", "b; c"}, "{b; b; c; c}", "b; c"},
		{[]string{"-x", "$x++; $x--"}, "n; a++; b++; b--", "b++; b--"},
		{[]string{"-x", "$*_; b; $*_"}, "{a; b; c; d}", "a; b; c; d"},
		{[]string{"-x", "{$*_; $x}"}, "{a; b; c}", 1},
		{[]string{"-x", "{b; c}"}, "{a; b; c}", 0},
		{[]string{"-x", "$x := $_; $x = $_"}, "a := n; b := n; b = m", "b := n; b = m"},
		{[]string{"-x", "$x := $_; $*_; $x = $_"}, "a := n; b := n; b = m", "b := n; b = m"},

		// mixing lists
		{[]string{"-x", "$x, $y"}, "1; 2", 0},
		{[]string{"-x", "$x; $y"}, "1, 2", 0},

		// any number of statements
		{[]string{"-x", "$*x"}, "a; b", "a; b"},
		{[]string{"-x", "$*x; b; $*y"}, "a; b; c", 1},
		{[]string{"-x", "$*x; b; $*x"}, "a; b; c", 0},

		// const/var declarations
		{[]string{"-x", "const $x = $y"}, "const a = b", 1},
		{[]string{"-x", "const $x = $y"}, "const (a = b)", 1},
		{[]string{"-x", "const $x = $y"}, "const (a = b\nc = d)", 0},
		{[]string{"-x", "var $x int"}, "var a int", 1},
		{[]string{"-x", "var $x int"}, "var a int = 3", 0},

		// func declarations
		{
			[]string{"-x", "func $_($x $y) $y { return $x }"},
			"func a(i int) int { return i }", 1,
		},
		{[]string{"-x", "func $x(i int)"}, "func a(i int)", 1},
		{[]string{"-x", "func $x(i int) {}"}, "func a(i int)", 0},

		// type declarations
		{[]string{"-x", "struct{}"}, "type T struct{}", 1},
		{[]string{"-x", "type $x struct{}"}, "type T struct{}", 1},
		{[]string{"-x", "struct{$_ int}"}, "type T struct{n int}", 1},
		{[]string{"-x", "struct{$_ int}"}, "var V struct{n int}", 1},
		{[]string{"-x", "struct{$_}"}, "type T struct{n int}", 1},
		{[]string{"-x", "struct{$*_}"}, "type T struct{n int}", 1},
		{
			[]string{"-x", "struct{$*_; Foo $t; $*_}"},
			"type T struct{Foo string; a int; B}", 1,
		},

		// value specs
		{[]string{"-x", "$_ int"}, "var a int", 1},
		{[]string{"-x", "$_ int"}, "var a bool", 0},
		// TODO: consider these
		{[]string{"-x", "$_ int"}, "var a int = 3", 0},
		{[]string{"-x", "$_ int"}, "var a, b int", 0},
		{[]string{"-x", "$_ int"}, "func(i int) { println(i) }", 0},

		// entire files
		{[]string{"-x", "package $_"}, "package p; var a = 1", 0},
		{[]string{"-x", "package $_; func Foo() { $*_ }"}, "package p; func Foo() {}", 1},

		// blocks
		{[]string{"-x", "{ $x }"}, "{ a() }", 1},
		{[]string{"-x", "{ $x }"}, "{ a(); b() }", 0},

		// assigns
		{[]string{"-x", "$x = $y"}, "a = b", 1},
		{[]string{"-x", "$x := $y"}, "a, b := c()", 0},

		// if stmts
		{[]string{"-x", "if $x != nil { $y }"}, "if p != nil { p.foo() }", 1},
		{[]string{"-x", "if $x { $y }"}, "if a { b() } else { c() }", 0},
		{[]string{"-x", "if $x != nil { $y }"}, "if a != nil { return a }", 1},

		// for and range stmts
		{[]string{"-x", "for $x { $y }"}, "for b { c() }", 1},
		{[]string{"-x", "for $x := range $y { $z }"}, "for i := range l { c() }", 1},
		{[]string{"-x", "for range $y { $z }"}, "for _, e := range l { e() }", 0},

		// $*_ matching stmt+expr combos (ifs)
		{[]string{"-x", "if $*x {}"}, "if a {}", 1},
		{[]string{"-x", "if $*x {}"}, "if a(); b {}", 1},
		{[]string{"-x", "if $*x {}; if $*x {}"}, "if a(); b {}; if a(); b {}", 1},
		{[]string{"-x", "if $*x {}; if $*x {}"}, "if a(); b {}; if b {}", 0},
		{[]string{"-x", "if $*_ {} else {}"}, "if a(); b {}", 0},
		{[]string{"-x", "if $*_ {} else {}"}, "if a(); b {} else {}", 1},
		{[]string{"-x", "if a(); $*_ {}"}, "if b {}", 0},

		// $*_ matching stmt+expr combos (fors)
		{[]string{"-x", "for $*x {}"}, "for {}", 1},
		{[]string{"-x", "for $*x {}"}, "for a {}", 1},
		{[]string{"-x", "for $*x {}"}, "for i(); a; p() {}", 1},
		{[]string{"-x", "for $*x {}; for $*x {}"}, "for i(); a; p() {}; for i(); a; p() {}", 1},
		{[]string{"-x", "for $*x {}; for $*x {}"}, "for i(); a; p() {}; for i(); b; p() {}", 0},
		{[]string{"-x", "for a(); $*_; {}"}, "for b {}", 0},
		{[]string{"-x", "for ; $*_; c() {}"}, "for b {}", 0},

		// $*_ matching stmt+expr combos (switches)
		{[]string{"-x", "switch $*x {}"}, "switch a {}", 1},
		{[]string{"-x", "switch $*x {}"}, "switch a(); b {}", 1},
		{[]string{"-x", "switch $*x {}; switch $*x {}"}, "switch a(); b {}; switch a(); b {}", 1},
		{[]string{"-x", "switch $*x {}; switch $*x {}"}, "switch a(); b {}; switch b {}", 0},
		{[]string{"-x", "switch a(); $*_ {}"}, "for b {}", 0},

		// $*_ matching stmt+expr combos (node type mixing)
		{[]string{"-x", "if $*x {}; for $*x {}"}, "if a(); b {}; for a(); b; {}", 1},
		{[]string{"-x", "if $*x {}; for $*x {}"}, "if a(); b {}; for a(); b; c() {}", 0},

		// for $*_ {} matching a range for
		{[]string{"-x", "for $_ {}"}, "for range x {}", 0},
		{[]string{"-x", "for $*_ {}"}, "for range x {}", 1},
		{[]string{"-x", "for $*_ {}"}, "for _, v := range x {}", 1},

		// $*_ matching optional statements (ifs)
		{[]string{"-x", "if $*_; b {}"}, "if b {}", 1},
		{[]string{"-x", "if $*_; b {}"}, "if a := f(); b {}", 1},
		// TODO: should these match?
		//{[]string{"-x", "if a(); $*x { f($*x) }"}, "if a(); b { f(b) }", 1},
		//{[]string{"-x", "if a(); $*x { f($*x) }"}, "if a(); b { f(b, c) }", 0},
		//{[]string{"-x", "if $*_; $*_ {}"}, "if a(); b {}", 1},

		// $*_ matching optional statements (fors)
		{[]string{"-x", "for $*x; b; $*x {}"}, "for b {}", 1},
		{[]string{"-x", "for $*x; b; $*x {}"}, "for a(); b; a() {}", 1},
		{[]string{"-x", "for $*x; b; $*x {}"}, "for a(); b; c() {}", 0},

		// $*_ matching optional statements (switches)
		{[]string{"-x", "switch $*_; b {}"}, "switch b := f(); b {}", 1},
		{[]string{"-x", "switch $*_; b {}"}, "switch b := f(); c {}", 0},

		// inc/dec stmts
		{[]string{"-x", "$x++"}, "a[b]++", 1},
		{[]string{"-x", "$x--"}, "a++", 0},

		// returns
		{[]string{"-x", "return nil, $x"}, "{ return nil, err }", 1},
		{[]string{"-x", "return nil, $x"}, "{ return nil, 0, err }", 0},

		// go stmts
		{[]string{"-x", "go $x()"}, "go func() { a() }()", 1},
		{[]string{"-x", "go func() { $x }()"}, "go func() { a() }()", 1},
		{[]string{"-x", "go func() { $x }()"}, "go a()", 0},

		// defer stmts
		{[]string{"-x", "defer $x()"}, "defer func() { a() }()", 1},
		{[]string{"-x", "defer func() { $x }()"}, "defer func() { a() }()", 1},
		{[]string{"-x", "defer func() { $x }()"}, "defer a()", 0},

		// empty statement
		{[]string{"-x", ";"}, ";", 1},

		// labeled statement
		{[]string{"-x", "foo: a"}, "foo: a", 1},
		{[]string{"-x", "foo: a"}, "foo: b", 0},

		// send statement
		{[]string{"-x", "x <- 1"}, "x <- 1", 1},
		{[]string{"-x", "x <- 1"}, "y <- 1", 0},
		{[]string{"-x", "x <- 1"}, "x <- 2", 0},

		// branch statement
		{[]string{"-x", "break foo"}, "break foo", 1},
		{[]string{"-x", "break foo"}, "break bar", 0},
		{[]string{"-x", "break foo"}, "continue foo", 0},
		{[]string{"-x", "break"}, "break", 1},
		{[]string{"-x", "break foo"}, "break", 0},

		// case clause
		{[]string{"-x", "switch x {case 4: x}"}, "switch x {case 4: x}", 1},
		{[]string{"-x", "switch x {case 4: x}"}, "switch y {case 4: x}", 0},
		{[]string{"-x", "switch x {case 4: x}"}, "switch x {case 5: x}", 0},
		{[]string{"-x", "switch {$_}"}, "switch {case 5: x}", 1},
		{[]string{"-x", "switch x {$_}"}, "switch x {case 5: x}", 1},
		{[]string{"-x", "switch x {$*_}"}, "switch x {case 5: x}", 1},
		{[]string{"-x", "switch x {$*_}"}, "switch x {}", 1},
		{[]string{"-x", "switch x {$*_}"}, "switch x {case 1: a; case 2: b}", 1},
		{[]string{"-x", "switch {$a; $a}"}, "switch {case true: a; case true: a}", 1},
		{[]string{"-x", "switch {$a; $a}"}, "switch {case true: a; case true: b}", 0},

		// switch statement
		{[]string{"-x", "switch x; y {}"}, "switch x; y {}", 1},
		{[]string{"-x", "switch x {}"}, "switch x; y {}", 0},
		{[]string{"-x", "switch {}"}, "switch {}", 1},
		{[]string{"-x", "switch {}"}, "switch x {}", 0},
		{[]string{"-x", "switch {}"}, "switch {case y:}", 0},
		{[]string{"-x", "switch $_ {}"}, "switch x {}", 1},
		{[]string{"-x", "switch $_ {}"}, "switch x; y {}", 0},
		{[]string{"-x", "switch $_; $_ {}"}, "switch x {}", 0},
		{[]string{"-x", "switch $_; $_ {}"}, "switch x; y {}", 1},
		{[]string{"-x", "switch { $*_; case $*_: $*a }"}, "switch { case x: y() }", 0},

		// type switch statement
		{[]string{"-x", "switch x := y.(z); x {}"}, "switch x := y.(z); x {}", 1},
		{[]string{"-x", "switch x := y.(z); x {}"}, "switch y := y.(z); x {}", 0},
		{[]string{"-x", "switch x := y.(z); x {}"}, "switch y := y.(z); x {}", 0},
		// TODO more switch variations.

		// TODO select statement
		// TODO communication clause
		{[]string{"-x", "select {$*_}"}, "select {case <-x: a}", 1},
		{[]string{"-x", "select {$*_}"}, "select {}", 1},
		{[]string{"-x", "select {$a; $a}"}, "select {case <-x: a; case <-x: a}", 1},
		{[]string{"-x", "select {$a; $a}"}, "select {case <-x: a; case <-x: b}", 0},
		{[]string{"-x", "select {case x := <-y: f(x)}"}, "select {case x := <-y: f(x)}", 1},

		// aggressive mode
		{[]string{"-x", "for range $x {}"}, "for _ = range a {}", 0},
		{[]string{"-x", "~ for range $x {}"}, "for _ = range a {}", 1},
		{[]string{"-x", "~ for _ = range $x {}"}, "for range a {}", 1},
		{[]string{"-x", "a int"}, "var (a, b int; c bool)", 0},
		{[]string{"-x", "~ a int"}, "var (a, b uint; c bool)", 0},
		{[]string{"-x", "~ a int"}, "var (a, b int; c bool)", 1},
		{[]string{"-x", "~ a int"}, "var (a, b int; c bool)", 1},
		{[]string{"-x", "{ x; }"}, "switch { case true: x; }", 0},
		{[]string{"-x", "~ { x; }"}, "switch { case true: x; }", 1},
		{[]string{"-x", "a = b"}, "a = b; a := b", 1},
		{[]string{"-x", "a := b"}, "a = b; a := b", 1},
		{[]string{"-x", "~ a = b"}, "a = b; a := b; var a = b", 3},
		{[]string{"-x", "~ a := b"}, "a = b; a := b; var a = b", 3},

		// many cmds
		{
			[]string{"-x", "break"},
			"switch { case x: break }; for { y(); break; break }",
			3,
		},
		{
			[]string{"-x", "for { $*_ }", "-x", "break"},
			"switch { case x: break }; for { y(); break; break }",
			2,
		},
		{
			[]string{"-x", "for { $*_ }", "-g", "break"},
			"break; for {}; for { if x { break } else { break } }",
			1,
		},
		{
			[]string{"-x", "for { $*_ }", "-v", "break"},
			"break; for {}; for { x() }; for { break }",
			2,
		},
		{
			[]string{"-x", "for { $*sts }", "-x", "$*sts"},
			"for { a(); b() }",
			"a(); b()",
		},
		{
			[]string{"-x", "for { $*sts }", "-x", "$*sts"},
			"for { if x { a(); b() } }",
			"if x { a(); b(); }",
		},
		{
			[]string{"-x", "foo", "-s", "bar", "-w"},
			`foo(); println("foo"); println(foo, foobar)`,
			`bar(); println("foo"); println(bar, foobar)`,
		},
		{
			[]string{"-x", "$f()", "-s", "$f(nil)", "-w"},
			`foo(); bar(); baz(x)`,
			`foo(nil); bar(nil); baz(x)`,
		},
		{
			[]string{"-x", "foo($*_)", "-s", "foo()", "-w"},
			`foo(); foo(a, b); bar(x)`,
			`foo(); foo(); bar(x)`,
		},
		{
			[]string{"-x", "a, b", "-s", "c, d", "-w"},
			`foo(); foo(a, b); bar(a, b)`,
			`foo(); foo(c, d); bar(c, d)`,
		},
		{
			[]string{"-x", "a(); b()", "-s", "c(); d()", "-w"},
			`{ a(); b(); c(); }; { a(); a(); b(); }`,
			`{ c(); d(); c(); }; { a(); c(); d(); }`,
		},
		{
			[]string{"-x", "a()", "-s", "c()", "-w"},
			`{ a(); b(); a(); }`,
			`{ c(); b(); c(); }`,
		},
		{
			[]string{"-x", "go func() { $f() }()", "-s", "go $f()", "-w"},
			`{ go func() { f.Close() }(); }`,
			`{ go f.Close(); }`,
		},
		{
			[]string{"-x", "foo", "-s", "bar", "-w"},
			`package p; var foo int`,
			`package p; var bar int`,
		},
		{
			[]string{"-x", "foo($*a)", "-s", "bar($*a)", "-w"},
			`{ foo(); }`,
			`{ bar(); }`,
		},
		{
			[]string{"-x", "foo($*a)", "-s", "bar($*a)", "-w"},
			`{ foo(0); }`,
			`{ bar(0); }`,
		},
		{
			[]string{"-x", "a(); b()", "-s", "x = a()", "-w"},
			`{ a(); b(); }`,
			`{ x = a(); }`,
		},
		{
			[]string{"-x", "a(); b()", "-s", "a()", "-w"},
			`{ a(); b(); }`,
			`{ a(); }`,
		},
		{
			[]string{"-x", "a, b", "-s", "c", "-w"},
			`foo(a, b)`,
			`foo(c)`,
		},
		{
			[]string{"-x", "b = a()", "-s", "c()", "-w"},
			`if b = a(); b { }`,
			`if c(); b { }`,
		},
		{
			[]string{"-x", "foo()", "-p", "1"},
			`{ if foo() { bar(); }; etc(); }`,
			`if foo() { bar(); }`,
		},
		{
			[]string{"-x", "f($*a)", "-s", "f2(x, $a)", "-w"},
			`f(c, d)`,
			`f2(x, c, d)`,
		},
		{
			[]string{"-x", "err = f(); if err != nil { $*then }", "-s", "if err := f(); err != nil { $then }", "-w"},
			`{ err = f(); if err != nil { handle(err); }; }`,
			`{ if err := f(); err != nil { handle(err); }; }`,
		},
		{
			[]string{"-x", "List{$e}", "-s", "$e", "-w"},
			`List{foo()}`,
			`foo()`,
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			grepTest(t, tc.args, tc.src, tc.want)
		})
	}
}

type wantMultiline string

func TestMatchMultiline(t *testing.T) {
	tests := []struct {
		args []string
		src  string
		want string
	}{
		{
			[]string{"-x", "List{$e}", "-s", "$e", "-w"},
			"return List{\n\tfoo(),\n}",
			"return foo()",
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			grepTest(t, tc.args, tc.src, wantMultiline(tc.want))
		})
	}
}

func grepTest(t *testing.T, args []string, src string, want interface{}) {
	tfatalf := func(format string, a ...interface{}) {
		t.Fatalf("%v | %q: %s", args, src, fmt.Sprintf(format, a...))
	}
	m := matcher{fset: token.NewFileSet()}
	cmds, paths, err := m.parseCmds(args)
	if len(paths) > 0 {
		t.Fatalf("non-zero paths: %v", paths)
	}
	srcNode, file, srcErr := parseDetectingNode(m.fset, src)
	if srcErr != nil {
		t.Fatal(srcErr)
	}

	// Type-checking is attempted on a best-effort basis.
	m.Info = &types.Info{
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Scopes: make(map[ast.Node]*types.Scope),
	}
	pkg := types.NewPackage("", "")
	config := &types.Config{
		Importer: importer.Default(),
		Error:    func(error) {}, // don't stop at the first error
	}
	check := types.NewChecker(config, m.fset, pkg, m.Info)
	_ = check.Files([]*ast.File{file})
	m.scope = pkg.Scope()

	matches := m.matches(cmds, []ast.Node{srcNode})
	if want, ok := want.(wantErr); ok {
		if err == nil {
			tfatalf("wanted error %q, got none", want)
		} else if got := err.Error(); got != string(want) {
			tfatalf("wanted error %q, got %q", want, got)
		}
		return
	}
	if err != nil {
		tfatalf("unexpected error: %v", err)
	}
	if want, ok := want.(int); ok {
		if len(matches) != want {
			tfatalf("wanted %d matches, got %d", want, len(matches))
		}
		return
	}
	if len(matches) != 1 {
		tfatalf("wanted 1 match, got %d", len(matches))
	}
	var got, wantStr string
	switch want := want.(type) {
	case string:
		wantStr = want
		got = singleLinePrint(matches[0])
	case wantMultiline:
		wantStr = string(want)
		var buf bytes.Buffer
		printNode(&buf, m.fset, matches[0])
		got = buf.String()
	default:
		panic(fmt.Sprintf("unexpected want type: %T", want))
	}
	if got != wantStr {
		tfatalf("wanted %q, got %q", wantStr, got)
	}
}
