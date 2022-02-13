package quasigo_test

import (
	"runtime"
	"testing"

	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo"
	"github.com/quasilyte/go-ruleguard/ruleguard/quasigo/stdlib/qfmt"
)

type benchTestCase struct {
	name   string
	src    string
	params string
	args   []interface{}
}

var benchmarksNoAlloc = []*benchTestCase{
	{
		name: `ReturnFalse`,
		src:  `return false`,
	},

	{
		name: `ReturnInt`,
		src:  `return 384723`,
	},

	{
		name:   `ParamInt`,
		src:    `return x + y + z`,
		params: `x, y, z int`,
		args:   []interface{}{10, 20, 30},
	},

	{
		name:   `ParamString`,
		src:    `return len(s)`,
		params: `s string`,
		args:   []interface{}{"hello, world"},
	},

	{
		name: `LocalVars`,
		src:  `x := 1; y := x; return y`,
	},

	{
		name: `IfStmt`,
		src:  `x := 100; if x == 1 { x = 10 } else if x == 2 { x = 20 } else { x = 30 }; return x`,
	},

	{
		name: `CallNative`,
		src:  `return imul(1, 5) + imul(2, 2)`,
	},

	{
		name: `CounterLoop`,
		src:  `j := 0; for j < 10000 { j++ }; return j`,
	},

	{
		name: `CounterLoopNested`,
		src:  `j := 0; for j < 10000 { k := 0; for k < 10 { k++; j++; } }; return j`,
	},
}

func TestNoAllocs(t *testing.T) {
	for _, test := range benchmarksNoAlloc {
		env, compiled := compileBenchFunc(t, test.params, test.src)
		evalEnv := env.GetEvalEnv()
		pushArgs(evalEnv, test.args...)

		const numTests = 5
		failures := 0
		allocated := uint64(0)
		for i := 0; i < numTests; i++ {
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			quasigo.Call(evalEnv, compiled)
			runtime.ReadMemStats(&after)
			allocated = after.Alloc - before.Alloc
			if allocated != 0 {
				failures++
			}
		}
		if failures == numTests {
			t.Errorf("%s does allocate (%d bytes)", test.name, allocated)
		}
	}
}

func BenchmarkEval(b *testing.B) {
	var tests = []*benchTestCase{
		{
			name: `CallNativeVariadic0`,
			src:  `return fmt.Sprintf("no formatting")`,
		},
		{
			name: `CallNativeVariadic1`,
			src:  `return fmt.Sprintf("Hello, %s!", "world")`,
		},

		{
			name: `CallNativeVariadic2`,
			src:  `return fmt.Sprintf("%s:%d", "foo.go", 105)`,
		},
	}

	tests = append(tests, benchmarksNoAlloc...)

	runBench := func(b *testing.B, env *quasigo.EvalEnv, fn *quasigo.Func) {
		for i := 0; i < b.N; i++ {
			_ = quasigo.Call(env, fn)
		}
	}

	for _, test := range tests {
		test := test
		b.Run(test.name, func(b *testing.B) {
			env, compiled := compileBenchFunc(b, test.params, test.src)
			evalEnv := env.GetEvalEnv()
			pushArgs(evalEnv, test.args...)
			b.ResetTimer()
			runBench(b, evalEnv, compiled)
		})
	}
}

func pushArgs(env *quasigo.EvalEnv, args ...interface{}) {
	for _, arg := range args {
		switch arg := arg.(type) {
		case int:
			env.Stack.PushInt(arg)
		default:
			env.Stack.Push(arg)
		}
	}
}

func compileBenchFunc(t testing.TB, paramsSig, bodySrc string) (*quasigo.Env, *quasigo.Func) {
	makePackageSource := func(body string) string {
		return `
		  package ` + testPackage + `
		  import "fmt"
		  var _ = fmt.Sprintf
		  func f(` + paramsSig + `) interface{} {
			  ` + body + `
		  }
		  func imul(x, y int) int
		  `
	}

	env := quasigo.NewEnv()
	env.AddNativeFunc(testPackage, "imul", func(stack *quasigo.ValueStack) {
		y := stack.PopInt()
		x := stack.PopInt()
		stack.PushInt(x * y)
	})
	qfmt.ImportAll(env)
	src := makePackageSource(bodySrc)
	parsed, err := parseGoFile(testPackage, src)
	if err != nil {
		t.Fatalf("parse %s: %v", bodySrc, err)
	}
	compiled, err := compileTestFunc(env, "f", parsed)
	if err != nil {
		t.Fatalf("compile %s: %v", bodySrc, err)
	}
	return env, compiled
}
