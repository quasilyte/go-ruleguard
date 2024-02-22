package quasigo_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo"
	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo/internal/evaltest"
	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo/stdlib/qfmt"
	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo/stdlib/qstrconv"
	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo/stdlib/qstrings"
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

		// Accessing the fields.
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
		  package ` + testPackage + `
		  import "github.com/quasilyte/go-ruleguard/ruleguard/quasigo/internal/evaltest"
		  func target(i int, s string, b bool, foo, nilfoo *evaltest.Foo, nileface interface{}) ` + returnType + ` {
		    ` + body + `
		  }
		  func imul(x, y int) int
		  func idiv(x, y int) int
		  func newFoo(prefix string) * evaltest.Foo
		  `
	}

	env := quasigo.NewEnv()
	env.AddNativeFunc(testPackage, "imul", func(stack *quasigo.ValueStack) {
		y := stack.PopInt()
		x := stack.PopInt()
		stack.PushInt(x * y)
	})
	env.AddNativeFunc(testPackage, "idiv", func(stack *quasigo.ValueStack) {
		y := stack.PopInt()
		x := stack.PopInt()
		stack.PushInt(x / y)
	})
	env.AddNativeFunc(testPackage, "newFoo", func(stack *quasigo.ValueStack) {
		prefix := stack.Pop().(string)
		stack.Push(&evaltest.Foo{Prefix: prefix})
	})

	const evaltestPkgPath = `github.com/quasilyte/go-ruleguard/ruleguard/quasigo/internal/evaltest`
	const evaltestFoo = `*` + evaltestPkgPath + `.Foo`
	env.AddNativeMethod(evaltestFoo, "Method1", func(stack *quasigo.ValueStack) {
		x := stack.PopInt()
		obj := stack.Pop()
		foo := obj.(*evaltest.Foo)
		stack.Push(foo.Prefix + fmt.Sprint(x))
	})
	env.AddNativeMethod(evaltestFoo, "Prefix", func(stack *quasigo.ValueStack) {
		foo := stack.Pop().(*evaltest.Foo)
		stack.Push(foo.Prefix)
	})

	for i := range tests {
		test := tests[i]
		src := makePackageSource(test.src, test.result)
		parsed, err := parseGoFile(testPackage, src)
		if err != nil {
			t.Fatalf("parse %s: %v", test.src, err)
		}
		compiled, err := compileTestFunc(env, "target", parsed)
		if err != nil {
			t.Fatalf("compile %s: %v", test.src, err)
		}
		evalEnv := env.GetEvalEnv()
		evalEnv.Stack.PushInt(10)
		evalEnv.Stack.Push("foo")
		evalEnv.Stack.Push(true)
		evalEnv.Stack.Push(&evaltest.Foo{Prefix: "Hello"})
		evalEnv.Stack.Push((*evaltest.Foo)(nil))
		evalEnv.Stack.Push(nil)
		result := quasigo.Call(evalEnv, compiled)
		var unboxedResult interface{}
		if _, ok := test.result.(int); ok {
			unboxedResult = result.IntValue()
		} else {
			unboxedResult = result.Value()
		}
		if unboxedResult != test.result {
			t.Fatalf("eval %s:\nhave: %#v\nwant: %#v", test.src, unboxedResult, test.result)
		}
	}
}

func TestEvalFile(t *testing.T) {
	files, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	runGo := func(main string) (string, error) {
		out, err := exec.Command("go", "run", main).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%v: %s", err, out)
		}
		return string(out), nil
	}

	runQuasigo := func(main string) (string, error) {
		src, err := os.ReadFile(main)
		if err != nil {
			return "", err
		}
		env := quasigo.NewEnv()
		parsed, err := parseGoFile("main", string(src))
		if err != nil {
			return "", fmt.Errorf("parse: %v", err)
		}

		var stdout bytes.Buffer
		env.AddNativeFunc(`builtin`, `Print`, func(stack *quasigo.ValueStack) {
			arg := stack.Pop()
			fmt.Fprintln(&stdout, arg)
		})
		env.AddNativeFunc(`builtin`, `PrintInt`, func(stack *quasigo.ValueStack) {
			fmt.Fprintln(&stdout, stack.PopInt())
		})

		env.AddNativeMethod(`error`, `Error`, func(stack *quasigo.ValueStack) {
			err := stack.Pop().(error)
			stack.Push(err.Error())
		})

		qstrings.ImportAll(env)
		qstrconv.ImportAll(env)
		qfmt.ImportAll(env)

		mainFunc, err := compileTestFile(env, "main", "main", parsed)
		if err != nil {
			return "", err
		}
		if mainFunc == nil {
			return "", errors.New("can't find main() function")
		}

		quasigo.Call(env.GetEvalEnv(), mainFunc)
		return stdout.String(), nil
	}

	runTest := func(t *testing.T, mainFile string) {
		goResult, err := runGo(mainFile)
		if err != nil {
			t.Fatalf("run go: %v", err)
		}
		quasigoResult, err := runQuasigo(mainFile)
		if err != nil {
			t.Fatalf("run quasigo: %v", err)
		}
		if diff := cmp.Diff(quasigoResult, goResult); diff != "" {
			t.Errorf("output mismatch:\nhave (+): `%s`\nwant (-): `%s`\ndiff: %s", quasigoResult, goResult, diff)
		}
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		mainFile := filepath.Join("testdata", f.Name(), "main.go")
		t.Run(f.Name(), func(t *testing.T) {
			runTest(t, mainFile)
		})
	}
}
