package quasigo

import (
	"runtime"
	"testing"
)

type benchTestCase struct {
	name string
	src  string
}

var benchmarksNoAlloc = []*benchTestCase{
	{
		`ReturnFalse`,
		`return false`,
	},

	{
		`ReturnInt`,
		`return 384723`,
	},

	{
		`LocalVars`,
		`x := 1; y := x; return y`,
	},

	{
		`IfStmt`,
		`x := 100; if x == 1 { x = 10 } else if x == 2 { x = 20 } else { x = 30 }; return x`,
	},

	{
		`CallNative`,
		`return imul(1, 5) + imul(2, 2)`,
	},

	{
		`CounterLoop`,
		`j := 0; for j < 10000 { j++ }; return j`,
	},

	{
		`CounterLoopNested`,
		`j := 0; for j < 10000 { k := 0; for k < 10 { k++; j++; } }; return j`,
	},
}

func TestNoAllocs(t *testing.T) {
	for _, test := range benchmarksNoAlloc {
		env, compiled := compileBenchFunc(t, test.src)
		evalEnv := env.GetEvalEnv()

		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		Call(evalEnv, compiled)
		runtime.ReadMemStats(&after)

		if allocated := after.Alloc - before.Alloc; allocated != 0 {
			t.Errorf("%s does allocate (%d bytes)", test.name, allocated)
		}
	}
}

func BenchmarkEval(b *testing.B) {
	var tests []*benchTestCase
	tests = append(tests, benchmarksNoAlloc...)

	runBench := func(b *testing.B, env *EvalEnv, fn *Func) {
		for i := 0; i < b.N; i++ {
			_ = Call(env, fn)
		}
	}

	for _, test := range tests {
		test := test
		b.Run(test.name, func(b *testing.B) {
			env, compiled := compileBenchFunc(b, test.src)
			b.ResetTimer()
			runBench(b, env.GetEvalEnv(), compiled)
		})
	}
}

func compileBenchFunc(t testing.TB, bodySrc string) (*Env, *Func) {
	makePackageSource := func(body string) string {
		return `
		  package test
		  func f() interface{} {
			  ` + body + `
		  }
		  func imul(x, y int) int
		  `
	}

	env := NewEnv()
	env.AddNativeFunc(testPackage, "imul", func(stack *ValueStack) {
		x, y := stack.popInt2()
		stack.PushInt(x * y)
	})
	src := makePackageSource(bodySrc)
	parsed, err := parseGoFile(src)
	if err != nil {
		t.Fatalf("parse %s: %v", bodySrc, err)
	}
	compiled, err := compileTestFunc(env, "f", parsed)
	if err != nil {
		t.Fatalf("compile %s: %v", bodySrc, err)
	}
	return env, compiled
}
