//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	m.Import(`github.com/quasilyte/go-ruleguard/analyzer/testdata/src/filtertest/foolib`)

	m.Match(`objectTest($pkg.$_, "object is pkgname")`).Where(m["pkg"].Object.Is(`PkgName`)).Report(`true`)
	m.Match(`objectTest($pkg.$_, "object is pkgname")`).Where(!m["pkg"].Object.Is(`PkgName`)).Report(`false`)

	m.Match(`objectTest($v, "object is var")`).Where(!m["v"].Object.Is(`Var`)).Report(`false`)
	m.Match(`objectTest($v, "object is var")`).Where(m["v"].Object.Is(`Var`)).Report(`true`)

	m.Match(`objectTest($*xs, "variadic object is var")`).Where(!m["xs"].Object.Is(`Var`)).Report(`false`)
	m.Match(`objectTest($*xs, "variadic object is var")`).Where(m["xs"].Object.Is(`Var`)).Report(`true`)

	m.Match(`typeTest($x, "contains time.Time")`).
		Where(m["x"].Type.Underlying().Is(`struct{$*_; time.Time; $*_}`)).
		Report(`true`)

	m.Match(`typeTest($x, "starts with time.Time")`).
		Where(m["x"].Type.Underlying().Is(`struct{time.Time; $*_}`)).
		Report(`true`)

	m.Match(`typeTest($x, "non-underlying type test; T + T")`).
		Where(m["x"].Type.Is(`struct{$t; $t}`)).
		Report(`true`)

	m.Match(`typeTest($x + $y)`).
		Where(m["x"].Type.Is(`string`) && m["y"].Type.Is("string")).
		Report(`concat`)

	m.Match(`typeTest($x + $y)`).
		Where(m["x"].Type.Is(`string`) && m["y"].Type.Is("string")).
		Report(`concat`)

	m.Match(`typeTest($x + $y)`).
		Where(m["x"].Type.Is(`int`) && m["y"].Type.Is("int")).
		Report(`addition`)

	m.Match(`typeTest($x > $y)`).
		Where(!m["x"].Type.Is(`int`)).
		Report(`$x !is(int)`)

	m.Match(`typeTest($*xs, "variadic int")`).Where(m["xs"].Type.Is(`int`)).Report(`true`)
	m.Match(`typeTest($*xs, "variadic int")`).Where(!m["xs"].Type.Is(`int`)).Report(`false`)

	m.Match(`typeTest($*xs, "variadic underlying int")`).Where(m["xs"].Type.Underlying().Is(`int`)).Report(`true`)
	m.Match(`typeTest($*xs, "variadic underlying int")`).Where(!m["xs"].Type.Underlying().Is(`int`)).Report(`false`)

	m.Match(`typeTest($x > $y)`).
		Where(!m["x"].Type.Is(`string`) && m["x"].Pure).
		Report(`$x !is(string) && pure`)

	m.Match(`typeTest($s, $s)`).
		Where(m["s"].Type.Is(`[]string`)).
		Report(`$s is([]string)`)

	m.Match(`typeTest("2 type filters", $x)`).
		Where(!m["x"].Type.Is(`string`) && !m["x"].Type.Is(`int`)).
		Report(`$x !is(string) && !is(int)`)

	m.Match(`typeTest($x, "implements io.Reader")`).
		Where(m["x"].Type.Implements(`io.Reader`)).Report(`true`)
	m.Match(`typeTest($x, "implements foolib.Stringer")`).
		Where(m["x"].Type.Implements(`foolib.Stringer`)).Report(`true`)
	m.Match(`typeTest($x, "implements error")`).
		Where(m["x"].Type.Implements(`error`)).Report(`true`)

	m.Match(`typeTest($*xs, "variadic implements error")`).
		Where(m["xs"].Type.Implements(`error`)).Report(`true`)
	m.Match(`typeTest($*xs, "variadic implements error")`).
		Where(!m["xs"].Type.Implements(`error`)).Report(`false`)

	m.Match(`typeTest($*xs, "variadic size==4")`).Where(m["xs"].Type.Size == 4).Report(`true`)
	m.Match(`typeTest($*xs, "variadic size==4")`).Where(!(m["xs"].Type.Size == 4)).Report(`false`)

	m.Match(`typeTest($x, "size>=100")`).Where(m["x"].Type.Size >= 100).Report(`true`)
	m.Match(`typeTest($x, "size<=100")`).Where(m["x"].Type.Size <= 100).Report(`true`)
	m.Match(`typeTest($x, "size>100")`).Where(m["x"].Type.Size > 100).Report(`true`)
	m.Match(`typeTest($x, "size<100")`).Where(m["x"].Type.Size < 100).Report(`true`)
	m.Match(`typeTest($x, "size==100")`).Where(m["x"].Type.Size == 100).Report(`true`)
	m.Match(`typeTest($x, "size!=100")`).Where(m["x"].Type.Size != 100).Report(`true`)

	m.Match(`typeTest($x(), "func() int")`).
		Where(m["x"].Type.Is("func() int")).
		Report(`true`)

	m.Match(`typeTest($x($*_), "func(int) int")`).
		Where(m["x"].Type.Is("func(int) int")).
		Report(`true`)

	m.Match(`typeTest($x(), "func() string")`).
		Where(m["x"].Type.Is("func() string")).
		Report(`true`)

	m.Match(`typeTest($t0 == $t1, "time==time")`).Where(m["t0"].Type.Is("time.Time")).Report(`true`)
	m.Match(`typeTest($t0 != $t1, "time!=time")`).Where(m["t1"].Type.Is("time.Time")).Report(`true`)

	m.Match(`pureTest($x)`).
		Where(m["x"].Pure).
		Report("true")

	m.Match(`pureTest($x)`).
		Where(!m["x"].Pure).
		Report("false")

	m.Match(`pureTest($*xs, "variadic pure")`).
		Where(m["xs"].Pure).
		Report("true")

	m.Match(`pureTest($*xs, "variadic pure")`).
		Where(!m["xs"].Pure).
		Report("false")

	m.Match(`constTest($*xs, "variadic const")`).
		Where(m["xs"].Const).
		Report("true")

	m.Match(`constTest($*xs, "variadic const")`).
		Where(!m["xs"].Const).
		Report("false")

	m.Match(`typeTest($*xs, "variadic addressable")`).Where(m["xs"].Addressable).Report("true")
	m.Match(`typeTest($*xs, "variadic addressable")`).Where(!m["xs"].Addressable).Report("false")

	m.Match(`typeTest($*xs, "variadic convertible to string")`).Where(m["xs"].Type.ConvertibleTo(`string`)).Report("true")
	m.Match(`typeTest($*xs, "variadic convertible to string")`).Where(!m["xs"].Type.ConvertibleTo(`string`)).Report("false")

	m.Match(`typeTest($*xs, "variadic assignable to string")`).Where(m["xs"].Type.AssignableTo(`string`)).Report("true")
	m.Match(`typeTest($*xs, "variadic assignable to string")`).Where(!m["xs"].Type.AssignableTo(`string`)).Report("false")

	m.Match(`valueTest($*xs, "variadic value 5")`).Where(m["xs"].Value.Int() == 5).Report(`true`)
	m.Match(`valueTest($*xs, "variadic value 5")`).Where(!(m["xs"].Value.Int() == 5)).Report(`false`)

	m.Match(`lineTest($x, "line 4")`).Where(m["x"].Line == 4).Report(`true`)
	m.Match(`lineTest($x, $y, "same line")`).Where(m["x"].Line == m["y"].Line).Report(`true`)
	m.Match(`lineTest($x, $y, "different line")`).Where(m["x"].Line != m["y"].Line).Report(`true`)

	m.Match(`textTest($x, "text=foo")`).Where(m["x"].Text == `foo`).Report(`true`)
	m.Match(`textTest($x, "text='foo'")`).Where(m["x"].Text == `"foo"`).Report(`true`)
	m.Match(`textTest($x, "text!='foo'")`).Where(m["x"].Text != `"foo"`).Report(`true`)

	m.Match(`textTest($x, "matches d+")`).Where(m["x"].Text.Matches(`^\d+$`)).Report(`true`)
	m.Match(`textTest($x, "doesn't match [A-Z]")`).Where(!m["x"].Text.Matches(`[A-Z]`)).Report(`true`)

	m.Match(`parensFilterTest($x, "type is error")`).Where((m["x"].Type.Is(`error`))).Report(`true`)

	m.Match(`importsTest(os.PathSeparator, "path/filepath")`).
		Where(m.File().Imports("path/filepath")).
		Report(`true`)

	m.Match(`importsTest(os.PathListSeparator, "path/filepath")`).
		Where(m.File().Imports("path/filepath")).
		Report(`true`)

	m.Match(`fileTest("with foo prefix")`).
		Where(m.File().Name.Matches(`^foo_`)).
		Report(`true`)

	m.Match(`fileTest("f1.go")`).
		Where(m.File().Name.Matches(`^f1.go$`)).
		Report(`true`)

	m.Match(`(($x))`).Where(!m.File().PkgPath.Matches(`filtertest`)).Report(`suspicious double parens`)
	m.Match(`((($x)))`).Where(m.File().PkgPath.Matches(`filtertest`)).Report(`suspicious tripple parens`)

	m.Match(`nodeTest("3 identical expr statements in a row"); $x; $x; $x`).Where(m["x"].Node.Is(`ExprStmt`)).Report(`true`)

	m.Match(`nodeTest($x, "Expr")`).Where(m["x"].Node.Is(`Expr`)).Report(`true`)
	m.Match(`nodeTest($x, "BasicLit")`).Where(m["x"].Node.Is(`BasicLit`)).Report(`true`)
	m.Match(`nodeTest($x, "Ident")`).Where(m["x"].Node.Is(`Ident`)).Report(`true`)
	m.Match(`nodeTest($x, "!Ident")`).Where(!m["x"].Node.Is(`Ident`)).Report(`true`)
	m.Match(`nodeTest($x, "IndexExpr")`).Where(m["x"].Node.Is(`IndexExpr`)).Report(`true`)

	m.Match(`typeTest($x, "convertible to ([2]int)")`).
		Where(m["x"].Type.ConvertibleTo(`([2]int)`)).
		Report(`true`)

	m.Match(`typeTest($x, "convertible to [][]int")`).
		Where(m["x"].Type.ConvertibleTo(`[][]int`)).
		Report(`true`)

	m.Match(`typeTest($x, "assignable to map[*string]error")`).
		Where(m["x"].Type.AssignableTo(`map[*string]error`)).
		Report(`true`)

	m.Match(`typeTest($x, "assignable to interface{}")`).
		Where(m["x"].Type.AssignableTo(`interface{}`)).
		Report(`true`)

	m.Match(`typeTest($x, "is interface")`).
		Where(m["x"].Type.Is(`interface{ $*_ }`)).
		Report(`true`)

	m.Match(`typeTest($x, "underlying is interface")`).
		Where(m["x"].Type.Underlying().Is(`interface{ $*_ }`)).
		Report(`true`)

	m.Match(`textTest("", "root text test")`).
		Where(m["$$"].Text == `textTest("", "root text test")`).
		Report(`true`)

	m.Match(`typeTest($x, "is numeric")`).
		Where(m["x"].Type.OfKind("numeric")).
		Report(`true`)

	m.Match(`typeTest($x, "underlying is numeric")`).
		Where(m["x"].Type.Underlying().OfKind("numeric")).
		Report(`true`)

	m.Match(`typeTest($x, "is unsigned")`).
		Where(m["x"].Type.Underlying().OfKind("unsigned")).
		Report(`true`)

	m.Match(`typeTest($x, "is signed")`).
		Where(m["x"].Type.Underlying().OfKind("signed")).
		Report(`true`)

	m.Match(`typeTest($x, "is float")`).
		Where(m["x"].Type.Underlying().OfKind("float")).
		Report(`true`)

	m.Match(`typeTest($x, "is int")`).
		Where(m["x"].Type.Underlying().OfKind("int")).
		Report(`true`)

	m.Match(`typeTest($x, "is uint")`).
		Where(m["x"].Type.Underlying().OfKind("uint")).
		Report(`true`)

	m.Match(`typeTest($x, "pointer-free")`).
		Where(!m["x"].Type.HasPointers()).
		Report(`true`)

	m.Match(`typeTest($x, "has pointers")`).
		Where(m["x"].Type.HasPointers()).
		Report(`true`)

	m.Match(`typeTest($x, "has WriteString method")`).
		Where(m["x"].Type.HasMethod(`io.StringWriter.WriteString`)).
		Report(`true`)

	m.Match(`typeTest($x, "has String method")`).
		Where(m["x"].Type.HasMethod(`fmt.Stringer.String`)).
		Report(`true`)

	m.Match(`typeTest($x, $y, "same type sizes")`).
		Where(m["x"].Type.Size == m["y"].Type.Size).
		Report(`true`)

	m.Match(`typeTest($x, "is predicate func")`).
		Where(m["x"].Type.Is(`func ($_) bool`)).
		Report(`true`)

	m.Match(`typeTest($x, "is func")`).
		Where(m["x"].Type.Is(`func ($*_) $*_`)).
		Report(`true`)

	m.Match(`typeTest($x, $y, "identical types")`).
		Where(m["x"].Type.IdenticalTo(m["y"])).
		Report(`true`)

	m.Match(`typeTest($x, "comparable")`).
		Where(m["x"].Comparable).
		Report(`true`)

	m.Match(`$x = time.Now().String()`,
		`var $x = time.Now().String()`,
		`var $x $_ = time.Now().String()`,
		`$x := time.Now().String()`).
		Where(m["x"].Object.IsGlobal()).
		Report(`global var`)

	m.Match(`newIface("sink is io.Reader").($_)`).
		Where(m["$$"].SinkType.Is(`io.Reader`)).
		Report(`true`)

	m.Match(`newIface("sink is interface{}").($_)`).
		Where(m["$$"].SinkType.Is(`interface{}`)).
		Report(`true`)

	m.Match(`objectTest($x, "object is variadic param")`).
		Where(m["x"].Object.IsVariadicParam()).
		Report(`true`)
}
