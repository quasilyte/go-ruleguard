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
			`  PushIntConst 0 # value=1`,
			`  ReturnIntTop`,
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
			`  PushIntParam 0 # i`,
			`  PushIntConst 0 # value=2`,
			`  EqInt`,
			`  ReturnTop`,
		},

		`return i == 10 || i == 2`: {
			`  PushIntParam 0 # i`,
			`  PushIntConst 0 # value=10`,
			`  EqInt`,
			`  Dup`,
			`  JumpTrue 8 # L0`,
			`  PushIntParam 0 # i`,
			`  PushIntConst 1 # value=2`,
			`  EqInt`,
			`L0:`,
			`  ReturnTop`,
		},

		`return i == 10 && s == "foo"`: {
			`  PushIntParam 0 # i`,
			`  PushIntConst 0 # value=10`,
			`  EqInt`,
			`  Dup`,
			`  JumpFalse 8 # L0`,
			`  PushParam 1 # s`,
			`  PushConst 0 # value="foo"`,
			`  EqString`,
			`L0:`,
			`  ReturnTop`,
		},

		`return imul(i, 5) == 10`: {
			`  PushIntParam 0 # i`,
			`  PushIntConst 0 # value=5`,
			`  CallBuiltin 0 # testpkg.imul`,
			`  PushIntConst 1 # value=10`,
			`  EqInt`,
			`  ReturnTop`,
		},

		`x := 10; y := x; return y`: {
			`  PushIntConst 0 # value=10`,
			`  SetIntLocal 0 # x`,
			`  PushIntLocal 0 # x`,
			`  SetIntLocal 1 # y`,
			`  PushIntLocal 1 # y`,
			`  ReturnIntTop`,
		},

		`if b { return 1 }; return 0`: {
			`  PushParam 2 # b`,
			`  JumpFalse 6 # L0`,
			`  PushIntConst 0 # value=1`,
			`  ReturnIntTop`,
			`L0:`,
			`  PushIntConst 1 # value=0`,
			`  ReturnIntTop`,
		},

		`if b { return 1 } else { return 0 }`: {
			`  PushParam 2 # b`,
			`  JumpFalse 6 # L0`,
			`  PushIntConst 0 # value=1`,
			`  ReturnIntTop`,
			`L0:`,
			`  PushIntConst 1 # value=0`,
			`  ReturnIntTop`,
		},

		`x := 0; if b { x = 5 } else { x = 50 }; return x`: {
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 0 # x`,
			`  PushParam 2 # b`,
			`  JumpFalse 10 # L0`,
			`  PushIntConst 1 # value=5`,
			`  SetIntLocal 0 # x`,
			`  Jump 7 # L1`,
			`L0:`,
			`  PushIntConst 2 # value=50`,
			`  SetIntLocal 0 # x`,
			`L1:`,
			`  PushIntLocal 0 # x`,
			`  ReturnIntTop`,
		},

		`if i != 2 { return "a" } else if b { return "b" }; return "c"`: {
			`  PushIntParam 0 # i`,
			`  PushIntConst 0 # value=2`,
			`  NotEqInt`,
			`  JumpFalse 6 # L0`,
			`  PushConst 0 # value="a"`,
			`  ReturnTop`,
			`L0:`,
			`  PushParam 2 # b`,
			`  JumpFalse 6 # L1`,
			`  PushConst 1 # value="b"`,
			`  ReturnTop`,
			`L1:`,
			`  PushConst 2 # value="c"`,
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
			`  PushIntConst 0 # value=1`,
			`  StringSliceFrom`,
			`  ReturnTop`,
		},

		`return s[:1]`: {
			`  PushParam 1 # s`,
			`  PushIntConst 0 # value=1`,
			`  StringSliceTo`,
			`  ReturnTop`,
		},

		`return s[1:2]`: {
			`  PushParam 1 # s`,
			`  PushIntConst 0 # value=1`,
			`  PushIntConst 1 # value=2`,
			`  StringSlice`,
			`  ReturnTop`,
		},

		`return len(s) >= 0`: {
			`  PushParam 1 # s`,
			`  StringLen`,
			`  PushIntConst 0 # value=0`,
			`  GtEqInt`,
			`  ReturnTop`,
		},

		`return i > 0`: {
			`  PushIntParam 0 # i`,
			`  PushIntConst 0 # value=0`,
			`  GtInt`,
			`  ReturnTop`,
		},

		`return i < 0`: {
			`  PushIntParam 0 # i`,
			`  PushIntConst 0 # value=0`,
			`  LtInt`,
			`  ReturnTop`,
		},

		`return i <= 0`: {
			`  PushIntParam 0 # i`,
			`  PushIntConst 0 # value=0`,
			`  LtEqInt`,
			`  ReturnTop`,
		},

		`x := 0; x++; return x`: {
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 0 # x`,
			`  IncLocal 0 # x`,
			`  PushIntLocal 0 # x`,
			`  ReturnIntTop`,
		},

		`x := 0; x--; return x`: {
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 0 # x`,
			`  DecLocal 0 # x`,
			`  PushIntLocal 0 # x`,
			`  ReturnIntTop`,
		},

		`j := 0; for { j++; break; }; return j`: {
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 0 # j`,
			`L1:`,
			`  IncLocal 0 # j`,
			`  Jump 6 # L0`,
			`  Jump -5 # L1`,
			`L0:`,
			`  PushIntLocal 0 # j`,
			`  ReturnIntTop`,
		},

		`j := -5; for { if j > 0 { break }; j++; }; return j`: {
			`  PushIntConst 0 # value=-5`,
			`  SetIntLocal 0 # j`,
			`L2:`,
			`  PushIntLocal 0 # j`,
			`  PushIntConst 1 # value=0`,
			`  GtInt`,
			`  JumpFalse 6 # L0`,
			`  Jump 8 # L1`,
			`L0:`,
			`  IncLocal 0 # j`,
			`  Jump -13 # L2`,
			`L1:`,
			`  PushIntLocal 0 # j`,
			`  ReturnIntTop`,
		},

		`j := 0; for j < 1000 { j++ }; return j`: {
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 0 # j`,
			`  Jump 5 # L0`,
			`L1:`,
			`  IncLocal 0 # j`,
			`L0:`,
			`  PushIntLocal 0 # j`,
			`  PushIntConst 1 # value=1000`,
			`  LtInt`,
			`  JumpTrue -7 # L1`,
			`  PushIntLocal 0 # j`,
			`  ReturnIntTop`,
		},

		`j := 0; for j < 100 { k := 0; for { if k > 40 { break }; k++; j++; } }; return j`: {
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 0 # j`,
			`  Jump 25 # L0`,
			`L3:`,
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 1 # k`,
			`L2:`,
			`  PushIntLocal 1 # k`,
			`  PushIntConst 1 # value=40`,
			`  GtInt`,
			`  JumpFalse 6 # L1`,
			`  Jump 10 # L0`,
			`L1:`,
			`  IncLocal 1 # k`,
			`  IncLocal 0 # j`,
			`  Jump -15 # L2`,
			`L0:`,
			`  PushIntLocal 0 # j`,
			`  PushIntConst 2 # value=100`,
			`  LtInt`,
			`  JumpTrue -27 # L3`,
			`  PushIntLocal 0 # j`,
			`  ReturnIntTop`,
		},

		`j := 0; for j < 10000 { k := 0; for k < 10 { k++; j++; } }; return j`: {
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 0 # j`,
			`  Jump 22 # L0`,
			`L3:`,
			`  PushIntConst 0 # value=0`,
			`  SetIntLocal 1 # k`,
			`  Jump 7 # L1`,
			`L2:`,
			`  IncLocal 1 # k`,
			`  IncLocal 0 # j`,
			`L1:`,
			`  PushIntLocal 1 # k`,
			`  PushIntConst 1 # value=10`,
			`  LtInt`,
			`  JumpTrue -9 # L2`,
			`L0:`,
			`  PushIntLocal 0 # j`,
			`  PushIntConst 2 # value=10000`,
			`  LtInt`,
			`  JumpTrue -24 # L3`,
			`  PushIntLocal 0 # j`,
			`  ReturnIntTop`,
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
		x, y := stack.pop2()
		stack.Push(x.(int) * y.(int))
	})
	env.AddNativeFunc(testPackage, "idiv", func(stack *ValueStack) {
		x, y := stack.pop2()
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
