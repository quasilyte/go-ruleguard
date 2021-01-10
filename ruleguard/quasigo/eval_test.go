package quasigo

import (
	"fmt"
	"testing"

	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo/internal/evaltest"
)

func TestEval(t *testing.T) {
	type testCase struct {
		src    string
		result interface{}
	}

	exprTests := []testCase{
		// Const literals.
		{`1`, 1},
		{`"foo"`, "foo"},
		{`true`, true},
		{`false`, false},

		// Function args.
		{`b`, true},
		{`i`, 10},

		// Arith operators.
		{`5 + 5`, 10},
		{`i + i`, 20},
		{`i - 5`, 5},
		{`5 - i`, -5},

		// String operators.
		{`s + s`, "foofoo"},

		// Bool operators.
		{`!b`, false},
		{`!!b`, true},
		{`i == 2`, false},
		{`i == 10`, true},
		{`i >= 10`, true},
		{`i >= 9`, true},
		{`i >= 11`, false},
		{`i > 10`, false},
		{`i > 9`, true},
		{`i > -1`, true},
		{`i < 10`, false},
		{`i < 11`, true},
		{`i <= 10`, true},
		{`i <= 11`, true},
		{`i != 2`, true},
		{`i != 10`, false},
		{`s != "foo"`, false},
		{`s != "bar"`, true},

		// || operator.
		{`i == 2 || i == 10`, true},
		{`i == 10 || i == 2`, true},
		{`i == 2 || i == 3 || i == 10`, true},
		{`i == 2 || i == 10 || i == 3`, true},
		{`i == 10 || i == 2 || i == 3`, true},
		{`!(i == 10 || i == 2 || i == 3)`, false},

		// && operator.
		{`i == 10 && s == "foo"`, true},
		{`i == 10 && s == "foo" && true`, true},
		{`i == 20 && s == "foo"`, false},
		{`i == 10 && s == "bar"`, false},
		{`i == 10 && s == "foo" && false`, false},

		// Builtin func call.
		{`imul(2, 3)`, 6},
		{`idiv(9, 3)`, 3},
		{`idiv(imul(2, 3), 1 + 1)`, 3},

		// Method call.
		{`foo.Method1(40)`, "Hello40"},
		{`newFoo("x").Method1(11)`, "x11"},

		// Accesing the fields.
		{`foo.Prefix`, "Hello"},

		// Nil checks.
		{`nilfoo == nil`, true},
		{`nileface == nil`, true},
		{`nil == nilfoo`, true},
		{`nil == nileface`, true},
		{`nilfoo != nil`, false},
		{`nileface != nil`, false},
		{`nil != nilfoo`, false},
		{`nil != nileface`, false},
		{`foo == nil`, false},
		{`foo != nil`, true},

		// String slicing.
		{`s[:]`, "foo"},
		{`s[0:]`, "foo"},
		{`s[1:]`, "oo"},
		{`s[:1]`, "f"},
		{`s[:0]`, ""},
		{`s[1:2]`, "o"},
		{`s[1:3]`, "oo"},

		// Builtin len().
		{`len(s)`, 3},
		{`len(s) == 3`, true},
		{`len(s[1:])`, 2},
	}

	tests := []testCase{
		{`if b { return 1 }; return 0`, 1},
		{`if !b { return 1 }; return 0`, 0},
		{`if b { return 1 } else { return 0 }`, 1},
		{`if !b { return 1 } else { return 0 }`, 0},

		{`x := 2; if x == 2 { return "a" } else if x == 0 { return "b" }; return "c"`, "a"},
		{`x := 2; if x == 0 { return "a" } else if x == 2 { return "b" }; return "c"`, "b"},
		{`x := 2; if x == 0 { return "a" } else if x == 1 { return "b" }; return "c"`, "c"},
		{`x := 2; if x == 2 { return "a" } else if x == 0 { return "b" } else { return "c" }`, "a"},
		{`x := 2; if x == 0 { return "a" } else if x == 2 { return "b" } else { return "c" }`, "b"},
		{`x := 2; if x == 0 { return "a" } else if x == 1 { return "b" } else { return "c" }`, "c"},
		{`x := 0; if b { x = 5 } else { x = 50 }; return x`, 5},
		{`x := 0; if !b { x = 5 } else { x = 50 }; return x`, 50},
		{`x := 0; if b { x = 1 } else if x == 0 { x = 2 } else { x = 3 }; return x`, 1},
		{`x := 0; if !b { x = 1 } else if x == 0 { x = 2 } else { x = 3 }; return x`, 2},
		{`x := 0; if !b { x = 1 } else if x == 1 { x = 2 } else { x = 3 }; return x`, 3},

		{`x := 0; x++; return x`, 1},
		{`x := i; x++; return x`, 11},
		{`x := 0; x--; return x`, -1},
		{`x := i; x--; return x`, 9},

		{`j := 0; for { j = j + 1; break; }; return j`, 1},
		{`j := -5; for { if j > 0 { break }; j++; }; return j`, 1},
		{`j := -5; for { if j >= 0 { break }; j++; }; return j`, 0},
		{`j := 0; for j < 0 { j++; break; }; return j`, 0},
		{`j := -5; for j < 0 { j++ }; return j`, 0},
		{`j := -5; for j <= 0 { j++; }; return j`, 1},
		{`j := 0; for j < 100 { k := 0; for { if k > 40 { break }; k++; j++; } }; return j`, 123},
		{`j := 0; for j < 10000 { k := 0; for k < 10 { k++; j++; } }; return j`, 10000},
	}

	for _, test := range exprTests {
		test.src = `return ` + test.src
		tests = append(tests, test)
	}

	makePackageSource := func(body string, result interface{}) string {
		var returnType string
		switch result.(type) {
		case int:
			returnType = "int"
		case string:
			returnType = "string"
		case bool:
			returnType = "bool"
		default:
			t.Fatalf("unexpected result type: %T", result)
		}
		return `
		  package test
		  import "github.com/quasilyte/go-ruleguard/ruleguard/quasigo/internal/evaltest"
		  func target(i int, s string, b bool, foo, nilfoo *evaltest.Foo, nileface interface{}) ` + returnType + ` {
		    ` + body + `
		  }
		  func imul(x, y int) int
		  func idiv(x, y int) int
		  func newFoo(prefix string) * evaltest.Foo
		  `
	}

	env := NewEnv()
	env.AddNativeFunc(testPackage, "imul", func(stack *ValueStack) {
		x, y := stack.popInt2()
		stack.PushInt(x * y)
	})
	env.AddNativeFunc(testPackage, "idiv", func(stack *ValueStack) {
		x, y := stack.popInt2()
		stack.PushInt(x / y)
	})
	env.AddNativeFunc(testPackage, "newFoo", func(stack *ValueStack) {
		prefix := stack.Pop().(string)
		stack.Push(&evaltest.Foo{Prefix: prefix})
	})

	const evaltestPkgPath = `github.com/quasilyte/go-ruleguard/ruleguard/quasigo/internal/evaltest`
	const evaltestFoo = `*` + evaltestPkgPath + `.Foo`
	env.AddNativeMethod(evaltestFoo, "Method1", func(stack *ValueStack) {
		x := stack.PopInt()
		obj := stack.Pop()
		foo := obj.(*evaltest.Foo)
		stack.Push(foo.Prefix + fmt.Sprint(x))
	})
	env.AddNativeMethod(evaltestFoo, "Prefix", func(stack *ValueStack) {
		foo := stack.Pop().(*evaltest.Foo)
		stack.Push(foo.Prefix)
	})

	for _, test := range tests {
		src := makePackageSource(test.src, test.result)
		parsed, err := parseGoFile(src)
		if err != nil {
			t.Errorf("parse %s: %v", test.src, err)
			continue
		}
		compiled, err := compileTestFunc(env, "target", parsed)
		if err != nil {
			t.Errorf("compile %s: %v", test.src, err)
			continue
		}
		result := Call(env.GetEvalEnv(), compiled,
			10, "foo", true, &evaltest.Foo{Prefix: "Hello"}, (*evaltest.Foo)(nil), nil)
		var unboxedResult interface{}
		if _, ok := test.result.(int); ok {
			unboxedResult = result.IntValue()
		} else {
			unboxedResult = result.Value()
		}
		if unboxedResult != test.result {
			t.Errorf("eval %s:\nhave: %#v\nwant: %#v", test.src, unboxedResult, test.result)
		}
	}
}
