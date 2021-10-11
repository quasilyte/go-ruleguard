package gogrep

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type compileTest struct {
	input  string
	output []string
}

func compileTestsFromMap(m map[string][]string) []compileTest {
	result := make([]compileTest, 0, len(m))
	for input, output := range m {
		result = append(result, compileTest{input: input, output: output})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].input < result[j].input
	})
	return result
}

func TestCompileWildcard(t *testing.T) {
	tests := compileTestsFromMap(map[string][]string{
		`$_`:  {`Node`},
		`$x`:  {`NamedNode x`},
		`$*_`: {`NodeSeq`},
		`$*x`: {`NamedNodeSeq x`},

		`a.$x`: {
			`SelectorExpr`,
			` • NamedNode x`,
			` • Ident a`,
		},
		`$x.b`: {
			`SimpleSelectorExpr b`,
			` • NamedNode x`,
		},

		`if $x != (nil) { $y }`: {
			`IfStmt`,
			` • BinaryExpr !=`,
			` •  • NamedNode x`,
			` •  • ParenExpr`,
			` •  •  • Ident nil`,
			` • BlockStmt`,
			` •  • NamedNode y`,
			` •  • End`,
		},
		`if $*_ {}`: {
			`IfInitStmt`,
			` • OptNode`,
			` • Node`,
			` • BlockStmt`,
			` •  • End`,
		},
		`if $*_ {} else {}`: {
			`IfInitElseStmt`,
			` • OptNode`,
			` • Node`,
			` • BlockStmt`,
			` •  • End`,
			` • BlockStmt`,
			` •  • End`,
		},
		`if $*x {} else {}`: {
			`IfNamedOptElseStmt x`,
			` • BlockStmt`,
			` •  • End`,
			` • BlockStmt`,
			` •  • End`,
		},
		`if $*x {} else if $*x {}`: {
			`IfNamedOptElseStmt x`,
			` • BlockStmt`,
			` •  • End`,
			` • IfNamedOptStmt x`,
			` •  • BlockStmt`,
			` •  •  • End`,
		},
		`if $*x {}`: {
			`IfNamedOptStmt x`,
			` • BlockStmt`,
			` •  • End`,
		},
		`if $_; cond {}`: {
			`IfInitStmt`,
			` • Node`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
		},
		`if $*_; cond {}`: {
			`IfInitStmt`,
			` • OptNode`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
		},
		`if $*x; cond {}`: {
			`IfInitStmt`,
			` • NamedOptNode x`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
		},

		`func ($x typ) {}`: {
			`FuncLit`,
			` • VoidFuncType`,
			` •  • FieldList`,
			` •  •  • Field`,
			` •  •  •  • NamedNode x`,
			` •  •  •  • Ident typ`,
			` •  •  • End`,
			` • BlockStmt`,
			` •  • End`,
		},

		`print($*_, x, $*_)`: {
			`CallExpr`,
			` • Ident print`,
			` • NodeSeq`,
			` • Ident x`,
			` • NodeSeq`,
			` • End`,
		},

		`{ $*_; x; $*_ }`: {
			`BlockStmt`,
			` • NodeSeq`,
			` • ExprStmt`,
			` •  • Ident x`,
			` • NodeSeq`,
			` • End`,
		},
		`{ $*head; x }`: {
			`BlockStmt`,
			` • NamedNodeSeq head`,
			` • ExprStmt`,
			` •  • Ident x`,
			` • End`,
		},

		`$l: if c {}`: {
			`LabeledStmt`,
			` • NamedNode l`,
			` • IfStmt`,
			` •  • Ident c`,
			` •  • BlockStmt`,
			` •  •  • End`,
		},

		`goto $l`: {
			`LabeledBranchStmt goto`,
			` • NamedNode l`,
		},

		`for $*_; $*_; $*_ {}`: {
			`ForInitCondPostStmt`,
			` • OptNode`,
			` • OptNode`,
			` • OptNode`,
			` • BlockStmt`,
			` •  • End`,
		},

		`const $x = $y`: {
			`ConstDecl`,
			` • ValueInitSpec`,
			` •  • NamedNode x`,
			` •  • End`,
			` •  • NamedNode y`,
			` •  • End`,
			` • End`,
		},

		`$_ int`: {
			`TypedValueSpec`,
			` • Node`,
			` • End`,
			` • Ident int`,
		},
		`$_ int = 5`: {
			`TypedValueInitSpec`,
			` • Node`,
			` • End`,
			` • Ident int`,
			` • BasicLit 5`,
			` • End`,
		},

		`switch {$_}`: {
			`SwitchStmt`,
			` • Node`,
			` • End`,
		},

		`switch $*_; x.(type) {}`: {
			`TypeSwitchInitStmt`,
			` • OptNode`,
			` • ExprStmt`,
			` •  • TypeSwitchAssertExpr`,
			` •  •  • Ident x`,
			` • End`,
		},

		`select {$*x}`: {
			`SelectStmt`,
			` • NamedNodeSeq x`,
			` • End`,
		},

		`package $p`: {
			`EmptyPackage`,
			` • NamedNode p`,
		},

		// $*_ in a place of a field list implies a field list of 0 or more fields.
		// It can also match a field list of 1 element and nil.
		`func $_() $*_ { $*_ }`: {
			`FuncDecl`,
			` • Node`,
			` • FuncType`,
			` •  • FieldList`,
			` •  •  • End`,
			` •  • OptNode`,
			` • BlockStmt`,
			` •  • NodeSeq`,
			` •  • End`,
		},

		// $y in a place of a field list implies a field list of exactly 1 field.
		`func $_($x $y) $y { return $x }`: {
			`FuncDecl`,
			` • Node`,
			` • FuncType`,
			` •  • FieldList`,
			` •  •  • Field`,
			` •  •  •  • NamedNode x`,
			` •  •  •  • NamedNode y`,
			` •  •  • End`,
			` •  • FieldList`,
			` •  •  • UnnamedField`,
			` •  •  •  • NamedNode y`,
			` •  •  • End`,
			` • BlockStmt`,
			` •  • ReturnStmt`,
			` •  •  • NamedNode x`,
			` •  •  • End`,
			` •  • End`,
		},

		`func _($*_) {}`: {
			`FuncDecl`,
			` • Ident _`,
			` • VoidFuncType`,
			` •  • OptNode`,
			` • BlockStmt`,
			` •  • End`,
		},

		`f($*_)`: {
			`CallExpr`,
			` • Ident f`,
			` • NodeSeq`,
			` • End`,
		},

		`f(1, $*_)`: {
			`CallExpr`,
			` • Ident f`,
			` • BasicLit 1`,
			` • NodeSeq`,
			` • End`,
		},

		`f($_)`: {
			`NonVariadicCallExpr`,
			` • Ident f`,
			` • Node`,
			` • End`,
		},

		`var x int; if true { f() }`: {
			`MultiStmt`,
			` • DeclStmt`,
			` •  • VarDecl`,
			` •  •  • TypedValueSpec`,
			` •  •  •  • Ident x`,
			` •  •  •  • End`,
			` •  •  •  • Ident int`,
			` •  •  • End`,
			` • IfStmt`,
			` •  • Ident true`,
			` •  • BlockStmt`,
			` •  •  • ExprStmt`,
			` •  •  •  • NonVariadicCallExpr`,
			` •  •  •  •  • Ident f`,
			` •  •  •  •  • End`,
			` •  •  • End`,
			` • End`,
		},
	})

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			input := test.input
			want := test.output
			fset := token.NewFileSet()
			p, err := Compile(fset, input, false)
			if err != nil {
				t.Errorf("compile `%s`: %v", input, err)
				return
			}
			have := formatProgram(p.m.prog)
			if diff := cmp.Diff(have, want); diff != "" {
				t.Errorf("compile `%s` (+want -have):\n%s", input, diff)
				fmt.Printf("Output:\n")
				for _, line := range have {
					fmt.Printf("`%s`,\n", line)
				}
				return
			}
		})
	}
}

func TestCompile(t *testing.T) {
	tests := compileTestsFromMap(map[string][]string{
		`package p;`: {
			`EmptyPackage`,
			` • Ident p`,
		},

		`var ()`: {
			`VarDecl`,
			` • End`,
		},
		`type foo = int`: {
			`TypeDecl`,
			` • TypeAliasSpec`,
			` •  • Ident foo`,
			` •  • Ident int`,
			` • End`,
		},
		`type (a int64; b string)`: {
			`TypeDecl`,
			` • TypeSpec`,
			` •  • Ident a`,
			` •  • Ident int64`,
			` • TypeSpec`,
			` •  • Ident b`,
			` •  • Ident string`,
			` • End`,
		},

		`10`:    {`BasicLit 10`},
		`2.4`:   {`BasicLit 2.4`},
		`"foo"`: {`BasicLit "foo"`},
		`'a'`:   {`BasicLit 97`},
		`'\n'`:  {`BasicLit 10`},
		`'✓'`:   {`BasicLit 10003`},

		`*x`: {
			`StarExpr`,
			` • Ident x`,
		},
		`+x`: {
			`UnaryExpr +`,
			` • Ident x`,
		},
		`-x`: {
			`UnaryExpr -`,
			` • Ident x`,
		},
		`((x))`: {
			`ParenExpr`,
			` • ParenExpr`,
			` •  • Ident x`,
		},

		`[]func() int{}`: {
			`TypedCompositeLit`,
			` • SliceType`,
			` •  • FuncType`,
			` •  •  • FieldList`,
			` •  •  •  • End`,
			` •  •  • FieldList`,
			` •  •  •  • UnnamedField`,
			` •  •  •  •  • Ident int`,
			` •  •  •  • End`,
			` • End`,
		},

		`func () {}`: {
			`FuncLit`,
			` • VoidFuncType`,
			` •  • FieldList`,
			` •  •  • End`,
			` • BlockStmt`,
			` •  • End`,
		},
		`func(xs ...int) {}`: {
			`FuncLit`,
			` • VoidFuncType`,
			` •  • FieldList`,
			` •  •  • SimpleField xs`,
			` •  •  •  • TypedEllipsis`,
			` •  •  •  •  • Ident int`,
			` •  •  • End`,
			` • BlockStmt`,
			` •  • End`,
		},
		`func(x, y int, z int) (string, string) {}`: {
			`FuncLit`,
			` • FuncType`,
			` •  • FieldList`,
			` •  •  • MultiField`,
			` •  •  •  • Ident x`,
			` •  •  •  • Ident y`,
			` •  •  •  • End`,
			` •  •  •  • Ident int`,
			` •  •  • SimpleField z`,
			` •  •  •  • Ident int`,
			` •  •  • End`,
			` •  • FieldList`,
			` •  •  • UnnamedField`,
			` •  •  •  • Ident string`,
			` •  •  • UnnamedField`,
			` •  •  •  • Ident string`,
			` •  •  • End`,
			` • BlockStmt`,
			` •  • End`,
		},

		`1 + 2`: {
			`BinaryExpr +`,
			` • BasicLit 1`,
			` • BasicLit 2`,
		},
		`1 - (x)`: {
			`BinaryExpr -`,
			` • BasicLit 1`,
			` • ParenExpr`,
			` •  • Ident x`,
		},

		`f(1, 2)`: {
			`NonVariadicCallExpr`,
			` • Ident f`,
			` • BasicLit 1`,
			` • BasicLit 2`,
			` • End`,
		},

		`f(g(), xs...)`: {
			`VariadicCallExpr`,
			` • Ident f`,
			` • NonVariadicCallExpr`,
			` •  • Ident g`,
			` •  • End`,
			` • Ident xs`,
			` • End`,
		},

		`x[0]`: {
			`IndexExpr`,
			` • Ident x`,
			` • BasicLit 0`,
		},

		`s[:]`: {
			`SliceExpr`,
			` • Ident s`,
		},
		`s[from:]`: {
			`SliceFromExpr`,
			` • Ident s`,
			` • Ident from`,
		},
		`s[:to]`: {
			`SliceToExpr`,
			` • Ident s`,
			` • Ident to`,
		},
		`s[from:to]`: {
			`SliceFromToExpr`,
			` • Ident s`,
			` • Ident from`,
			` • Ident to`,
		},
		`s[:to:max]`: {
			`SliceToCapExpr`,
			` • Ident s`,
			` • Ident to`,
			` • Ident max`,
		},
		`s[from:to:max]`: {
			`SliceFromToCapExpr`,
			` • Ident s`,
			` • Ident from`,
			` • Ident to`,
			` • Ident max`,
		},

		`([2]int)(x)`: {
			`NonVariadicCallExpr`,
			` • ParenExpr`,
			` •  • ArrayType`,
			` •  •  • BasicLit 2`,
			` •  •  • Ident int`,
			` • Ident x`,
			` • End`,
		},
		`([]int)(x)`: {
			`NonVariadicCallExpr`,
			` • ParenExpr`,
			` •  • SliceType`,
			` •  •  • Ident int`,
			` • Ident x`,
			` • End`,
		},

		`[]int{1, 2}`: {
			`TypedCompositeLit`,
			` • SliceType`,
			` •  • Ident int`,
			` • BasicLit 1`,
			` • BasicLit 2`,
			` • End`,
		},
		`[][]int{{1, 2}, {3}}`: {
			`TypedCompositeLit`,
			` • SliceType`,
			` •  • SliceType`,
			` •  •  • Ident int`,
			` • CompositeLit`,
			` •  • BasicLit 1`,
			` •  • BasicLit 2`,
			` •  • End`,
			` • CompositeLit`,
			` •  • BasicLit 3`,
			` •  • End`,
			` • End`,
		},

		`[...]int{5: 1}`: {
			`TypedCompositeLit`,
			` • ArrayType`,
			` •  • Ellipsis`,
			` •  • Ident int`,
			` • KeyValueExpr`,
			` •  • BasicLit 5`,
			` •  • BasicLit 1`,
			` • End`,
		},
		`map[int]string{}`: {
			`TypedCompositeLit`,
			` • MapType`,
			` •  • Ident int`,
			` •  • Ident string`,
			` • End`,
		},

		`go f()`: {
			`GoStmt`,
			` • NonVariadicCallExpr`,
			` •  • Ident f`,
			` •  • End`,
		},

		`defer f()`: {
			`DeferStmt`,
			` • NonVariadicCallExpr`,
			` •  • Ident f`,
			` •  • End`,
		},

		`ch <- 1`: {
			`SendStmt`,
			` • Ident ch`,
			` • BasicLit 1`,
		},

		`x.y.z`: {
			`SimpleSelectorExpr z`,
			` • SimpleSelectorExpr y`,
			` •  • Ident x`,
		},

		`x.(int)`: {
			`TypeAssertExpr`,
			` • Ident x`,
			` • Ident int`,
		},

		`;`: {`EmptyStmt`},

		`x++`: {
			`IncDecStmt ++`,
			` • Ident x`,
		},
		`x--`: {
			`IncDecStmt --`,
			` • Ident x`,
		},

		`{ f(); g(); }`: {
			`BlockStmt`,
			` • ExprStmt`,
			` •  • NonVariadicCallExpr`,
			` •  •  • Ident f`,
			` •  •  • End`,
			` • ExprStmt`,
			` •  • NonVariadicCallExpr`,
			` •  •  • Ident g`,
			` •  •  • End`,
			` • End`,
		},

		`if cond {}`: {
			`IfStmt`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
		},
		`if init; cond {}`: {
			`IfInitStmt`,
			` • ExprStmt`,
			` •  • Ident init`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
		},
		`if cond {} else { f() }`: {
			`IfElseStmt`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
			` • BlockStmt`,
			` •  • ExprStmt`,
			` •  •  • NonVariadicCallExpr`,
			` •  •  •  • Ident f`,
			` •  •  •  • End`,
			` •  • End`,
		},
		`if cond {} else if cond2 { f() } else {}`: {
			`IfElseStmt`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
			` • IfElseStmt`,
			` •  • Ident cond2`,
			` •  • BlockStmt`,
			` •  •  • ExprStmt`,
			` •  •  •  • NonVariadicCallExpr`,
			` •  •  •  •  • Ident f`,
			` •  •  •  •  • End`,
			` •  •  • End`,
			` •  • BlockStmt`,
			` •  •  • End`,
		},
		`if init1; cond {} else if init2; cond2 { f() } else {}`: {
			`IfInitElseStmt`,
			` • ExprStmt`,
			` •  • Ident init1`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
			` • IfInitElseStmt`,
			` •  • ExprStmt`,
			` •  •  • Ident init2`,
			` •  • Ident cond2`,
			` •  • BlockStmt`,
			` •  •  • ExprStmt`,
			` •  •  •  • NonVariadicCallExpr`,
			` •  •  •  •  • Ident f`,
			` •  •  •  •  • End`,
			` •  •  • End`,
			` •  • BlockStmt`,
			` •  •  • End`,
		},

		`return 1, 2`: {
			`ReturnStmt`,
			` • BasicLit 1`,
			` • BasicLit 2`,
			` • End`,
		},

		`break`:       {`BranchStmt break`},
		`continue`:    {`BranchStmt continue`},
		`fallthrough`: {`BranchStmt fallthrough`},
		`break l`:     {`SimpleLabeledBranchStmt break l`},
		`continue l`:  {`SimpleLabeledBranchStmt continue l`},
		`goto l`:      {`SimpleLabeledBranchStmt goto l`},

		`foo: x`: {
			`SimpleLabeledStmt foo`,
			` • ExprStmt`,
			` •  • Ident x`,
		},

		`x = y`: {
			`AssignStmt =`,
			` • Ident x`,
			` • Ident y`,
		},
		`x := y`: {
			`AssignStmt :=`,
			` • Ident x`,
			` • Ident y`,
		},
		`x, y := f()`: {
			`MultiAssignStmt :=`,
			` • Ident x`,
			` • Ident y`,
			` • End`,
			` • NonVariadicCallExpr`,
			` •  • Ident f`,
			` •  • End`,
			` • End`,
		},

		`(chan int)(nil)`: {
			`NonVariadicCallExpr`,
			` • ParenExpr`,
			` •  • ChanType send recv`,
			` •  •  • Ident int`,
			` • Ident nil`,
			` • End`,
		},
		`(chan<- int)(nil)`: {
			`NonVariadicCallExpr`,
			` • ParenExpr`,
			` •  • ChanType send`,
			` •  •  • Ident int`,
			` • Ident nil`,
			` • End`,
		},
		`(<-chan int)(nil)`: {
			`NonVariadicCallExpr`,
			` • ParenExpr`,
			` •  • ChanType recv`,
			` •  •  • Ident int`,
			` • Ident nil`,
			` • End`,
		},

		`for range xs {}`: {
			`RangeStmt`,
			` • Ident xs`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for i := range xs {}`: {
			`RangeKeyStmt :=`,
			` • Ident i`,
			` • Ident xs`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for i = range xs {}`: {
			`RangeKeyStmt =`,
			` • Ident i`,
			` • Ident xs`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for i, x := range xs {}`: {
			`RangeKeyValueStmt :=`,
			` • Ident i`,
			` • Ident x`,
			` • Ident xs`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for i, x = range xs {}`: {
			`RangeKeyValueStmt =`,
			` • Ident i`,
			` • Ident x`,
			` • Ident xs`,
			` • BlockStmt`,
			` •  • End`,
		},

		`for {}`: {
			`ForStmt`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for ;; {}`: {
			`ForStmt`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for ;; post {}`: {
			`ForPostStmt`,
			` • ExprStmt`,
			` •  • Ident post`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for cond {}`: {
			`ForCondStmt`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for ; cond; {}`: {
			`ForCondStmt`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for ; cond; post {}`: {
			`ForCondPostStmt`,
			` • Ident cond`,
			` • ExprStmt`,
			` •  • Ident post`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for init; ; {}`: {
			`ForInitStmt`,
			` • ExprStmt`,
			` •  • Ident init`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for init; ; post {}`: {
			`ForInitPostStmt`,
			` • ExprStmt`,
			` •  • Ident init`,
			` • ExprStmt`,
			` •  • Ident post`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for init; cond; {}`: {
			`ForInitCondStmt`,
			` • ExprStmt`,
			` •  • Ident init`,
			` • Ident cond`,
			` • BlockStmt`,
			` •  • End`,
		},
		`for init; cond; post {}`: {
			`ForInitCondPostStmt`,
			` • ExprStmt`,
			` •  • Ident init`,
			` • Ident cond`,
			` • ExprStmt`,
			` •  • Ident post`,
			` • BlockStmt`,
			` •  • End`,
		},

		`switch x.(type) {}`: {
			`TypeSwitchStmt`,
			` • ExprStmt`,
			` •  • TypeSwitchAssertExpr`,
			` •  •  • Ident x`,
			` • End`,
		},

		`switch x := y.(type) {}`: {
			`TypeSwitchStmt`,
			` • AssignStmt :=`,
			` •  • Ident x`,
			` •  • TypeSwitchAssertExpr`,
			` •  •  • Ident y`,
			` • End`,
		},

		`switch {case 1, 2: f(); default: g() }`: {
			`SwitchStmt`,
			` • CaseClause`,
			` •  • BasicLit 1`,
			` •  • BasicLit 2`,
			` •  • End`,
			` •  • ExprStmt`,
			` •  •  • NonVariadicCallExpr`,
			` •  •  •  • Ident f`,
			` •  •  •  • End`,
			` •  • End`,
			` • DefaultCaseClause`,
			` •  • ExprStmt`,
			` •  •  • NonVariadicCallExpr`,
			` •  •  •  • Ident g`,
			` •  •  •  • End`,
			` •  • End`,
			` • End`,
		},

		`fmt.Println()`: {
			`NonVariadicCallExpr`,
			` • SimpleSelectorExpr Println`,
			` •  • StdlibPkg fmt`,
			` • End`,
		},

		`x = fmt.Sprint(y)`: {
			`AssignStmt =`,
			` • Ident x`,
			` • NonVariadicCallExpr`,
			` •  • SimpleSelectorExpr Sprint`,
			` •  •  • StdlibPkg fmt`,
			` •  • Ident y`,
			` •  • End`,
		},
	})

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			input := test.input
			want := test.output
			fset := token.NewFileSet()
			n := testParseNode(t, fset, input)
			var c compiler
			p, err := c.Compile(fset, n, false)
			if err != nil {
				t.Errorf("compile `%s`: %v", input, err)
				return
			}

			have := formatProgram(p)
			if diff := cmp.Diff(have, want); diff != "" {
				t.Errorf("compile `%s` (+want -have):\n%s", input, diff)
				fmt.Printf("Output:\n")
				for _, line := range have {
					fmt.Printf("`%s`,\n", line)
				}
				return
			}
		})
	}
}

func testParseNode(t testing.TB, fset *token.FileSet, s string) ast.Node {
	if strings.HasPrefix(s, "package ") {
		file, err := parser.ParseFile(fset, "string", s, 0)
		if err != nil {
			t.Fatalf("parse `%s`: %v", s, err)
		}
		return file
	}
	source := `package p; func _() { ` + s + ` }`
	file, err := parser.ParseFile(fset, "string", source, 0)
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
