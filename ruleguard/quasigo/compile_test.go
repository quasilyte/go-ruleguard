package quasigo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompile(t *testing.T) {
	tests := map[string][]string{
		`return 1`: {
			`  PushConst 0 # value=1`,
			`  ReturnTop`,
		},

		`return false`: {
			`  ReturnFalse`,
		},

		`return true`: {
			`  ReturnTrue`,
		},

		`return b`: {
			`  PushParam 2 # b`,
			`  ReturnTop`,
		},

		`return i == 2`: {
			`  PushParam 0 # i`,
			`  PushConst 0 # value=2`,
			`  EqInt`,
			`  ReturnTop`,
		},

		`return i == 10 || i == 2`: {
			`  PushParam 0 # i`,
			`  PushConst 0 # value=10`,
			`  EqInt`,
			`  JumpTrue 9 # L0`,
			`  Pop`,
			`  PushParam 0 # i`,
			`  PushConst 1 # value=2`,
			`  EqInt`,
			`L0:`,
			`  ReturnTop`,
		},

		`return i == 10 && s == "foo"`: {
			`  PushParam 0 # i`,
			`  PushConst 0 # value=10`,
			`  EqInt`,
			`  JumpFalse 9 # L0`,
			`  Pop`,
			`  PushParam 1 # s`,
			`  PushConst 1 # value="foo"`,
			`  EqString`,
			`L0:`,
			`  ReturnTop`,
		},

		`return imul(i, 5) == 10`: {
			`  PushParam 0 # i`,
			`  PushConst 0 # value=5`,
			`  CallBuiltin 0 # testpkg.imul`,
			`  PushConst 1 # value=10`,
			`  EqInt`,
			`  ReturnTop`,
		},

		`x := 10; y := x; return y`: {
			`  PushConst 0 # value=10`,
			`  SetLocal 0 # x`,
			`  PushLocal 0 # x`,
			`  SetLocal 1 # y`,
			`  PushLocal 1 # y`,
			`  ReturnTop`,
		},

		`if b { return 1 }; return 0`: {
			`  PushParam 2 # b`,
			`  JumpFalse 7 # L0`,
			`  Pop`,
			`  PushConst 0 # value=1`,
			`  ReturnTop`,
			`L0:`,
			`  PushConst 1 # value=0`,
			`  ReturnTop`,
		},

		`if b { return 1 } else { return 0 }`: {
			`  PushParam 2 # b`,
			`  JumpFalse 7 # L0`,
			`  Pop`,
			`  PushConst 0 # value=1`,
			`  ReturnTop`,
			`L0:`,
			`  PushConst 1 # value=0`,
			`  ReturnTop`,
		},

		`x := 0; if b { x = 5 } else { x = 50 }; return x`: {
			`  PushConst 0 # value=0`,
			`  SetLocal 0 # x`,
			`  PushParam 2 # b`,
			`  JumpFalse 11 # L0`,
			`  Pop`,
			`  PushConst 1 # value=5`,
			`  SetLocal 0 # x`,
			`  Jump 7 # L1`,
			`L0:`,
			`  PushConst 2 # value=50`,
			`  SetLocal 0 # x`,
			`L1:`,
			`  PushLocal 0 # x`,
			`  ReturnTop`,
		},

		`if i != 2 { return "a" } else if b { return "b" }; return "c"`: {
			`  PushParam 0 # i`,
			`  PushConst 0 # value=2`,
			`  NotEqInt`,
			`  JumpFalse 7 # L0`,
			`  Pop`,
			`  PushConst 1 # value="a"`,
			`  ReturnTop`,
			`L0:`,
			`  PushParam 2 # b`,
			`  JumpFalse 7 # L1`,
			`  Pop`,
			`  PushConst 2 # value="b"`,
			`  ReturnTop`,
			`L1:`,
			`  PushConst 3 # value="c"`,
			`  ReturnTop`,
		},

		`return eface == nil`: {
			`  PushParam 3 # eface`,
			`  IsNil`,
			`  ReturnTop`,
		},

		`return nil == eface`: {
			`  PushParam 3 # eface`,
			`  IsNil`,
			`  ReturnTop`,
		},

		`return eface != nil`: {
			`  PushParam 3 # eface`,
			`  IsNotNil`,
			`  ReturnTop`,
		},

		`return nil != eface`: {
			`  PushParam 3 # eface`,
			`  IsNotNil`,
			`  ReturnTop`,
		},

		`return s[:]`: {
			`  PushParam 1 # s`,
			`  ReturnTop`,
		},

		`return s[1:]`: {
			`  PushParam 1 # s`,
			`  PushConst 0 # value=1`,
			`  StringSliceFrom`,
			`  ReturnTop`,
		},

		`return s[:1]`: {
			`  PushParam 1 # s`,
			`  PushConst 0 # value=1`,
			`  StringSliceTo`,
			`  ReturnTop`,
		},

		`return s[1:2]`: {
			`  PushParam 1 # s`,
			`  PushConst 0 # value=1`,
			`  PushConst 1 # value=2`,
			`  StringSlice`,
			`  ReturnTop`,
		},

		`return len(s) >= 0`: {
			`  PushParam 1 # s`,
			`  StringLen`,
			`  PushConst 0 # value=0`,
			`  GtEqInt`,
			`  ReturnTop`,
		},

		`return i > 0`: {
			`  PushParam 0 # i`,
			`  PushConst 0 # value=0`,
			`  GtInt`,
			`  ReturnTop`,
		},

		`return i < 0`: {
			`  PushParam 0 # i`,
			`  PushConst 0 # value=0`,
			`  LtInt`,
			`  ReturnTop`,
		},

		`return i <= 0`: {
			`  PushParam 0 # i`,
			`  PushConst 0 # value=0`,
			`  LtEqInt`,
			`  ReturnTop`,
		},
	}

	makePackageSource := func(body string) string {
		return `
		  package test
		  func f(i int, s string, b bool, eface interface{}) interface{} {
			` + body + `
		  }
		  func imul(x, y int) int
		  func idiv(x, y int) int
		  `
	}

	env := NewEnv()
	env.AddNativeFunc(testPackage, "imul", func(stack *ValueStack) {
		x, y := stack.Pop2()
		stack.Push(x.(int) * y.(int))
	})
	env.AddNativeFunc(testPackage, "idiv", func(stack *ValueStack) {
		x, y := stack.Pop2()
		stack.Push(x.(int) / y.(int))
	})

	for testSrc, disasmLines := range tests {
		src := makePackageSource(testSrc)
		parsed, err := parseGoFile(src)
		if err != nil {
			t.Errorf("parse %s: %v", testSrc, err)
			continue
		}
		compiled, err := compileTestFunc(env, "f", parsed)
		if err != nil {
			t.Errorf("compile %s: %v", testSrc, err)
			continue
		}
		want := disasmLines
		have := strings.Split(Disasm(env, compiled), "\n")
		have = have[:len(have)-1] // Drop an empty line
		if diff := cmp.Diff(have, want); diff != "" {
			t.Errorf("compile %s (-have +want):\n%s", testSrc, diff)
			fmt.Println("For copy/paste:")
			for _, l := range have {
				fmt.Printf("  `%s`,\n", l)
			}
			continue
		}
	}
}
