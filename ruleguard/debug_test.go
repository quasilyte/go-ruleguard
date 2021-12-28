package ruleguard

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDebug(t *testing.T) {
	allTests := map[string]map[string][]string{
		`m.MatchComment("// (?P<x>\\w+)").Where(m["x"].Text.Matches("^Test"))`: {
			`// TestFoo`: nil,

			`// Foo`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Text.Matches("^Test")`,
				`  $x: Foo`,
			},
		},

		`m.Match("f()").Where(!m.Deadcode())`: {
			`f()`:             nil,
			`if true { f() }`: nil,

			`if false { f() }`: {
				`input.go:4: [rules.go:5] rejected by !m.Deadcode()`,
			},
		},

		`m.Match("f($x)").Where(m["x"].Type.Is("string"))`: {
			`f("abc")`: nil,

			`f(10)`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Type.Is("string")`,
				`  $x int: 10`,
			},
		},

		`m.Match("$x + $y").Where(m["x"].Const && m["y"].Const)`: {
			`sink = 1 + 2`: nil,

			`sink = f().(int) + 2`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Const`,
				`  $x int: f().(int)`,
				`  $y int: 2`,
			},

			`sink = 1 + f().(int)`: {
				`input.go:4: [rules.go:5] rejected by m["y"].Const`,
				`  $x int: 1`,
				`  $y int: f().(int)`,
			},
		},

		`m.Match("$x + $_").Where(!m["x"].Type.Is("int"))`: {
			`sink = "a" + "b"`: nil,

			`sink = int(10) + 20`: {
				`input.go:4: [rules.go:5] rejected by !m["x"].Type.Is("int")`,
				`  $x int: int(10)`,
			},
		},

		`m.Match("$x + $_").Where(m["x"].Value.Int() >= 10)`: {
			`sink = 20 + 1`: nil,

			// OK: $x is const-folded.
			`sink = (2 << 3) + 1`: nil,

			// Not an int.
			`sink = "20" + "x"`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Value.Int() >= 10`,
				`  $x untyped string: "20"`,
			},

			// Not a const value.
			`sink = f().(int) + 0`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Value.Int() >= 10`,
				`  $x int: f().(int)`,
			},

			// Less than 10.
			`sink = 4 + 1`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Value.Int() >= 10`,
				`  $x untyped int: 4`,
			},
		},

		`m.Match("_ = $x").Where(m["x"].Node.Is("ParenExpr"))`: {
			`_ = (1)`: nil,

			`_ = 10`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Node.Is("ParenExpr")`,
				`  $x int: 10`,
			},
			`_ = "hello"`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Node.Is("ParenExpr")`,
				`  $x string: "hello"`,
			},
			`_ = f((10))`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Node.Is("ParenExpr")`,
				`  $x interface{}: f((10))`,
			},
		},

		// When debugging OR, the last alternative will be reported as the failure reason,
		// although it should be obvious that all operands are falsy.
		// We don't return the entire OR expression as a reason to avoid the output cluttering.
		`m.Match("_ = $x").Where(m["x"].Type.Is("int") || m["x"].Type.Is("string"))`: {
			`_ = ""`: nil,
			`_ = 10`: nil,

			`_ = []int{}`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Type.Is("string")`,
				`  $x []int: []int{}`,
			},

			`_ = int32(0)`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Type.Is("string")`,
				`  $x int32: int32(0)`,
			},
		},

		// Using 3 operands for || and different ()-groupings.
		`m.Match("_ = $x").Where(m["x"].Type.Is("int") || m["x"].Type.Is("string") || m["x"].Text == "f()")`: {
			`_ = ""`:  nil,
			`_ = 10`:  nil,
			`_ = f()`: nil,

			`_ = []string{"x"}`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Text == "f()"`,
				`  $x []string: []string{"x"}`,
			},
		},
		`m.Match("_ = $x").Where(m["x"].Type.Is("int") || (m["x"].Type.Is("string") || m["x"].Text == "f()"))`: {
			`_ = ""`:  nil,
			`_ = 10`:  nil,
			`_ = f()`: nil,

			`_ = []string{"x"}`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Text == "f()"`,
				`  $x []string: []string{"x"}`,
			},
		},
		`m.Match("_ = $x").Where((m["x"].Type.Is("int") || m["x"].Type.Is("string")) || m["x"].Text == "f()")`: {
			`_ = ""`:  nil,
			`_ = 10`:  nil,
			`_ = f()`: nil,

			`_ = []string{"x"}`: {
				`input.go:4: [rules.go:5] rejected by m["x"].Text == "f()"`,
				`  $x []string: []string{"x"}`,
			},
		},

		`isConst := func(v dsl.Var) bool { return v.Const }; m.Match("_ = $x").Where(isConst(m["x"]) && !m["x"].Type.Is("string"))`: {
			`_ = 10`: nil,

			`_ = "str"`: {
				`input.go:4: [rules.go:5] rejected by !m["x"].Type.Is("string")`,
				`  $x string: "str"`,
			},

			`_ = f()`: {
				`input.go:4: [rules.go:5] rejected by isConst(m["x"])`,
				`  $x interface{}: f()`,
			},
		},
	}

	loadRulesFromExpr := func(e *Engine, s string) {
		file := fmt.Sprintf(`
			package gorules
			import "github.com/quasilyte/go-ruleguard/dsl"
			func testrule(m dsl.Matcher) {
				%s.Report("$$")
			}`,
			s)
		ctx := &LoadContext{
			Fset: token.NewFileSet(),
		}
		err := e.Load(ctx, "rules.go", strings.NewReader(file))
		if err != nil {
			t.Fatalf("parse %s: %v", s, err)
		}
	}

	for expr, testCases := range allTests {
		e := NewEngine()
		loadRulesFromExpr(e, expr)
		for input, lines := range testCases {
			runner, err := newDebugTestRunner(input)
			if err != nil {
				t.Fatalf("init %s: %s: %v", expr, input, err)
			}
			if err := runner.Run(t, e); err != nil {
				t.Fatalf("run %s: %s: %v", expr, input, err)
			}
			if diff := cmp.Diff(runner.out, lines); diff != "" {
				t.Errorf("check %s: %s:\n(+want -have)\n%s", expr, input, diff)
			}
		}
	}
}

type debugTestRunner struct {
	ctx *RunContext
	f   *ast.File
	out []string
}

func (r debugTestRunner) Run(t *testing.T, e *Engine) error {
	if err := e.Run(r.ctx, r.f); err != nil {
		return err
	}
	return nil
}

func newDebugTestRunner(input string) (*debugTestRunner, error) {
	fullInput := fmt.Sprintf(`
		package testrule
		func testfunc() {
		  %s
		}
		func f(...interface{}) interface{} { return 10 }
		var sink interface{}`, input)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "input.go", []byte(fullInput), parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var typecheker types.Config
	info := types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
	}
	pkg, err := typecheker.Check("testrule", fset, []*ast.File{f}, &info)
	if err != nil {
		return nil, err
	}
	runner := &debugTestRunner{f: f}
	ctx := &RunContext{
		Debug: "testrule",
		DebugPrint: func(s string) {
			runner.out = append(runner.out, s)
		},
		Pkg:   pkg,
		Types: &info,
		Sizes: types.SizesFor("gc", runtime.GOARCH),
		Fset:  fset,
		Report: func(data *ReportData) {
			// Do nothing.
		},
	}
	runner.ctx = ctx
	return runner, nil
}
