// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func testRules(m fluent.Matcher) {
	m.Import(`github.com/quasilyte/go-ruleguard/analyzer/testdata/src/filtertest/foolib`)

	m.Match(`typeTest($x, "contains time.Time")`).
		Where(m["x"].Type.Underlying().Is(`struct{$*_; time.Time; $*_}`)).
		Report(`YES`)

	m.Match(`typeTest($x, "starts with time.Time")`).
		Where(m["x"].Type.Underlying().Is(`struct{time.Time; $*_}`)).
		Report(`YES`)

	m.Match(`typeTest($x, "non-underlying type test; T + T")`).
		Where(m["x"].Type.Is(`struct{$t; $t}`)).
		Report(`YES`)

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
		Where(m["x"].Type.Implements(`io.Reader`)).Report(`YES`)
	m.Match(`typeTest($x, "implements foolib.Stringer")`).
		Where(m["x"].Type.Implements(`foolib.Stringer`)).Report(`YES`)
	m.Match(`typeTest($x, "implements error")`).
		Where(m["x"].Type.Implements(`error`)).Report(`YES`)

	m.Match(`typeTest($x, "size>=100")`).Where(m["x"].Type.Size >= 100).Report(`YES`)
	m.Match(`typeTest($x, "size<=100")`).Where(m["x"].Type.Size <= 100).Report(`YES`)
	m.Match(`typeTest($x, "size>100")`).Where(m["x"].Type.Size > 100).Report(`YES`)
	m.Match(`typeTest($x, "size<100")`).Where(m["x"].Type.Size < 100).Report(`YES`)
	m.Match(`typeTest($x, "size==100")`).Where(m["x"].Type.Size == 100).Report(`YES`)
	m.Match(`typeTest($x, "size!=100")`).Where(m["x"].Type.Size != 100).Report(`YES`)

	m.Match(`typeTest($x(), "func() int")`).
		Where(m["x"].Type.Is("func() int")).
		Report(`YES`)

	m.Match(`typeTest($x($*_), "func(int) int")`).
		Where(m["x"].Type.Is("func(int) int")).
		Report(`YES`)

	m.Match(`typeTest($x(), "func() string")`).
		Where(m["x"].Type.Is("func() string")).
		Report(`YES`)

	m.Match(`typeTest($t0 == $t1, "time==time")`).Where(m["t0"].Type.Is("time.Time")).Report(`YES`)
	m.Match(`typeTest($t0 != $t1, "time!=time")`).Where(m["t1"].Type.Is("time.Time")).Report(`YES`)

	m.Match(`pureTest($x)`).
		Where(m["x"].Pure).
		Report("pure")

	m.Match(`pureTest($x)`).
		Where(!m["x"].Pure).
		Report("!pure")

	m.Match(`textTest($x, "text=foo")`).Where(m["x"].Text == `foo`).Report(`YES`)
	m.Match(`textTest($x, "text='foo'")`).Where(m["x"].Text == `"foo"`).Report(`YES`)
	m.Match(`textTest($x, "text!='foo'")`).Where(m["x"].Text != `"foo"`).Report(`YES`)

	m.Match(`textTest($x, "matches d+")`).Where(m["x"].Text.Matches(`^\d+$`)).Report(`YES`)
	m.Match(`textTest($x, "doesn't match [A-Z]")`).Where(!m["x"].Text.Matches(`[A-Z]`)).Report(`YES`)

	m.Match(`parensFilterTest($x, "type is error")`).Where((m["x"].Type.Is(`error`))).Report(`YES`)

	m.Match(`importsTest(os.PathSeparator, "path/filepath")`).
		Where(m.File().Imports("path/filepath")).
		Report(`YES`)

	m.Match(`importsTest(os.PathListSeparator, "path/filepath")`).
		Where(m.File().Imports("path/filepath")).
		Report(`YES`)

	m.Match(`fileTest("with foo prefix")`).
		Where(m.File().Name.Matches(`^foo_`)).
		Report(`YES`)

	m.Match(`fileTest("f1.go")`).
		Where(m.File().Name.Matches(`^f1.go$`)).
		Report(`YES`)

	m.Match(`(($x))`).Where(!m.File().PkgPath.Matches(`filtertest`)).Report(`suspicious double parens`)
	m.Match(`((($x)))`).Where(m.File().PkgPath.Matches(`filtertest`)).Report(`suspicious tripple parens`)

	m.Match(`nodeTest($x, "Expr")`).Where(m["x"].Node.Is(`Expr`)).Report(`YES`)
	m.Match(`nodeTest($x, "BasicLit")`).Where(m["x"].Node.Is(`BasicLit`)).Report(`YES`)
	m.Match(`nodeTest($x, "Ident")`).Where(m["x"].Node.Is(`Ident`)).Report(`YES`)
	m.Match(`nodeTest($x, "!Ident")`).Where(!m["x"].Node.Is(`Ident`)).Report(`YES`)
	m.Match(`nodeTest($x, "IndexExpr")`).Where(m["x"].Node.Is(`IndexExpr`)).Report(`YES`)
}
