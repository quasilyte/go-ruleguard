package filtertest

import (
	"errors"
	"os"
	"time"
)

type implementsAll struct{}

func (implementsAll) Read([]byte) (int, error) { return 0, nil }
func (implementsAll) String() string           { return "" }
func (*implementsAll) Error() string           { return "" }

type implementsAllNewtype implementsAll

type embedImplementsAll struct {
	implementsAll
}

type embedImplementsAllPtr struct {
	*implementsAll
}

func _() {
	fileTest("with foo prefix")
	fileTest("f1.go") // want `YES`
}

func convertibleTo() {
	type myInt2Array [2]int
	typeTest([2]int{}, "convertible to ([2]int)")      // want `YES`
	typeTest(myInt2Array{}, "convertible to ([2]int)") // want `YES`
	typeTest([3]int{}, "convertible to ([2]int)")

	type myIntSlice2 [][]int
	typeTest([][]int{{1}}, "convertible to [][]int")     // want `YES`
	typeTest(myIntSlice2(nil), "convertible to [][]int") // want `YES`
	typeTest([]int{}, "convertible to [][]int")
}

func assignableTo() {
	typeTest(map[*string]error{}, "assignable to map[*string]error") // want `YES`
	typeTest(map[*string]int{}, "assignable to map[*string]error")

	typeTest(0, "assignable to interface{}")   // want `YES`
	typeTest(5.6, "assignable to interface{}") // want `YES`
	typeTest("", "assignable to interface{}")  // want `YES`
}

func detectType() {
	{
		type withNamedTime struct {
			x int
			y time.Time
		}
		var foo struct {
			x time.Time
		}
		var bar withNamedTime
		type indirectFoo1 withNamedTime
		type indirectFoo2 indirectFoo1
		typeTest(withNamedTime{}, "contains time.Time") // want `YES`
		typeTest(foo, "contains time.Time")             // want `YES`
		typeTest(bar, "contains time.Time")             // want `YES`
		typeTest(indirectFoo2{}, "contains time.Time")  // want `YES`
	}

	{
		type timeFirst struct {
			y time.Time
			x int
		}
		var foo struct {
			x time.Time
		}
		var bar timeFirst
		typeTest(timeFirst{}, "starts with time.Time") // want `YES`
		typeTest(foo, "starts with time.Time")         // want `YES`
		typeTest(bar, "starts with time.Time")         // want `YES`
	}

	{
		type intPair struct {
			y int
			x int
		}
		var foo struct {
			x float32
			y float32
		}
		var bar intPair
		typeTest(struct { // want `YES`
			_ string
			_ string
		}{}, "non-underlying type test; T + T")
		typeTest(intPair{}, "non-underlying type test; T + T") // type is Named, not struct
		typeTest(foo, "non-underlying type test; T + T")       // want `YES`
		typeTest(bar, "non-underlying type test; T + T")       // type is Named, not struct
	}

	var i1, i2 int
	var ii []int
	var s1, s2 string
	var ss []string
	typeTest(s1 + s2) // want `concat`
	typeTest(i1 + i2) // want `addition`
	typeTest(s1 > s2) // want `\Qs1 !is(int)`
	typeTest(i1 > i2) // want `\Qi1 !is(string) && pure`
	typeTest(random() > i2)
	typeTest(ss, ss) // want `\Qss is([]string)`
	typeTest(ii, ii)
	typeTest("2 type filters", i1)
	typeTest("2 type filters", s1)
	typeTest("2 type filters", ii) // want `\Qii !is(string) && !is(int)`

	typeTest(implementsAll{}, "implements io.Reader") // want `YES`
	typeTest(i1, "implements io.Reader")
	typeTest(ss, "implements io.Reader")
	typeTest(implementsAll{}, "implements foolib.Stringer") // want `YES`
	typeTest(i1, "implements foolib.Stringer")
	typeTest(ss, "implements foolib.Stringer")
	typeTest(implementsAll{}, "implements error")
	typeTest(&implementsAll{}, "implements error") // want `YES`
	typeTest(i1, "implements error")
	typeTest(error(nil), "implements error")            // want `YES`
	typeTest(errors.New("example"), "implements error") // want `YES`
	typeTest(implementsAllNewtype{}, "implements error")
	typeTest(&implementsAllNewtype{}, "implements error")
	typeTest(embedImplementsAll{}, "implements error")
	typeTest(&embedImplementsAll{}, "implements error")    // want `YES`
	typeTest(embedImplementsAllPtr{}, "implements error")  // want `YES`
	typeTest(&embedImplementsAllPtr{}, "implements error") // want `YES`

	typeTest([100]byte{}, "size>=100") // want `YES`
	typeTest([105]byte{}, "size>=100") // want `YES`
	typeTest([10]byte{}, "size>=100")
	typeTest([100]byte{}, "size<=100") // want `YES`
	typeTest([105]byte{}, "size<=100")
	typeTest([10]byte{}, "size<=100") // want `YES`
	typeTest([100]byte{}, "size>100")
	typeTest([105]byte{}, "size>100") // want `YES`
	typeTest([10]byte{}, "size>100")
	typeTest([100]byte{}, "size<100")
	typeTest([105]byte{}, "size<100")
	typeTest([10]byte{}, "size<100")   // want `YES`
	typeTest([100]byte{}, "size==100") // want `YES`
	typeTest([105]byte{}, "size==100")
	typeTest([10]byte{}, "size==100")
	typeTest([100]byte{}, "size!=100")
	typeTest([105]byte{}, "size!=100") // want `YES`
	typeTest([10]byte{}, "size!=100")  // want `YES`

	var time1, time2 time.Time
	var err error
	typeTest(time1 == time2, "time==time") // want `YES`
	typeTest(err == nil, "time==time")
	typeTest(nil == err, "time==time")
	typeTest(time1 != time2, "time!=time") // want `YES`
	typeTest(err != nil, "time!=time")
	typeTest(nil != err, "time!=time")

	intFunc := func() int { return 10 }
	intToIntFunc := func(x int) int { return x }
	typeTest(intFunc(), "func() int")                 // want `YES`
	typeTest(func() int { return 0 }(), "func() int") // want `YES`
	typeTest(func() string { return "" }(), "func() int")
	typeTest(intToIntFunc(1), "func() int")

	typeTest(intToIntFunc(2), "func(int) int") // want `YES`
	typeTest(intToIntFunc, "func(int) int")
	typeTest(intFunc, "func(int) int")

	var v implementsAll
	typeTest(v.String(), "func() string") // want `YES`
	typeTest(implementsAll.String(v), "func() string")
	typeTest(implementsAll.String, "func() string")

}

func detectPure(x int, xs []int) {
	var foo struct {
		a int
	}

	xptr := &x

	pureTest(random())        // want `!pure`
	pureTest([]int{random()}) // want `!pure`

	pureTest(*xptr)               // want `pure`
	pureTest(int(x))              // want `pure`
	pureTest((*int)(&x))          // want `pure`
	pureTest((func())(func() {})) // want `pure`
	pureTest(foo.a)               // want `pure`
	pureTest(x * x)               // want `pure`
	pureTest((x * x))             // want `pure`
	pureTest(+x)                  // want `pure`
	pureTest(xs[0])               // want `pure`
	pureTest(xs[x])               // want `pure`
	pureTest([]int{0})            // want `pure`
}

func detectText(foo, bar int) {
	textTest(foo, "text=foo") // want `YES`
	textTest(bar, "text=foo")

	textTest("foo", "text='foo'") // want `YES`
	textTest("bar", "text='foo'")

	textTest("bar", "text!='foo'") // want `YES`
	textTest("foo", "text!='foo'")

	textTest(32, "matches d+") // want `YES`
	textTest(0x32, "matches d+")
	textTest("foo", "matches d+")

	textTest(1, "doesn't match [A-Z]") // want `YES`
	textTest("ABC", "doesn't match [A-Z]")
}

func detectParensFilter() {
	var err error
	parensFilterTest(err, "type is error") // want `YES`
}

func fileFilters1() {
	// No matches as this file doesn't import "path/filepath".
	importsTest(os.PathSeparator, "path/filepath")
	importsTest(os.PathListSeparator, "path/filepath")
}

func detectNode() {
	var i int
	var s string
	var rows [][]byte

	nodeTest("123", "Expr") // want `YES`
	nodeTest(`123`, "Expr") // want `YES`
	nodeTest(12, "Expr")    // want `YES`
	nodeTest(1.56, "Expr")  // want `YES`
	nodeTest(1+2, "Expr")   // want `YES`
	nodeTest(i, "Expr")     // want `YES`
	nodeTest(s, "Expr")     // want `YES`

	nodeTest("123", "BasicLit") // want `YES`
	nodeTest(`123`, "BasicLit") // want `YES`
	nodeTest(12, "BasicLit")    // want `YES`
	nodeTest(1.56, "BasicLit")  // want `YES`
	nodeTest(1+2, "BasicLit")
	nodeTest(i, "BasicLit")
	nodeTest(s, "BasicLit")

	nodeTest("123", "Ident")
	nodeTest(12, "Ident")
	nodeTest(i, "Ident") // want `YES`
	nodeTest(s, "Ident") // want `YES`

	nodeTest("42", "!Ident") // want `YES`
	nodeTest(12, "!Ident")   // want `YES`
	nodeTest(s[0], "!Ident") // want `YES`
	nodeTest(i, "!Ident")
	nodeTest(s, "!Ident")

	nodeTest(s[0], "IndexExpr")       // want `YES`
	nodeTest(rows[0][5], "IndexExpr") // want `YES`
	nodeTest("42", "IndexExpr")
}
