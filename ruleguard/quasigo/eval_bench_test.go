package quasigo

import (
	"testing"
)

func BenchmarkEval(b *testing.B) {
	type testCase struct {
		name string
		src  string
	}
	tests := []*testCase{
		{
			`ReturnFalse`,
			`return false`,
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
	}

	runBench := func(b *testing.B, env *EvalEnv, fn *Func) {
		for i := 0; i < b.N; i++ {
			_ = Call(env, fn)
		}
	}

	makePackageSource := func(body string) string {
		return `
		  package test
		  func f() interface{} {
			  ` + body + `
		  }
		  func imul(x, y int) int
		  `
	}

	for _, test := range tests {
		test := test
		b.Run(test.name, func(b *testing.B) {
			env := NewEnv()
			env.AddNativeFunc(testPackage, "imul", func(stack *ValueStack) {
				x, y := stack.Pop2()
				stack.Push(x.(int) * y.(int))
			})
			src := makePackageSource(test.src)
			parsed, err := parseGoFile(src)
			if err != nil {
				b.Fatalf("parse %s: %v", test.src, err)
			}
			compiled, err := compileTestFunc(env, "f", parsed)
			if err != nil {
				b.Fatalf("compile %s: %v", test.src, err)
			}

			b.ResetTimer()
			runBench(b, env.GetEvalEnv(), compiled)
		})
	}

}
