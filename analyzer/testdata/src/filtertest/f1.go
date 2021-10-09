package filtertest

import (
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"
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

type vector2D struct {
	X, Y float64
}

func (v vector2D) String() string {
	return fmt.Sprintf("{%f, %f}", v.X, v.Y)
}

func _() {
	fileTest("with foo prefix")
	fileTest("f1.go") // want `YES`
}

func detectObject() {
	var vec vector2D

	{
		objectTest(fmt.Println, "object is pkgname") // want `true`

		objectTest(vec.X, "object is pkgname")      // want `false`
		objectTest(vec.String, "object is pkgname") // want `false`
	}

	{
		objectTest(vec, "object is var")         // want `true`
		objectTest(os.Stdout, "object is var")   // want `true`
		objectTest(vec.X, "object is var")       // want `true`
		objectTest((vec), "object is var")       // want `true`
		objectTest((os.Stdout), "object is var") // want `true`
		objectTest((vec.X), "object is var")     // want `true`

		objectTest(fmt.Println, "object is var") // want `false`
		objectTest(vec.X+4, "object is var")     // want `false`
	}

	{
		objectTest("variadic object is var")             // want `true`
		objectTest(vec, "variadic object is var")        // want `true`
		objectTest(vec, vec.Y, "variadic object is var") // want `true`

		objectTest(1, "variadic object is var")      // want `false`
		objectTest(vec, 2, "variadic object is var") // want `false`
		objectTest(1, vec, "variadic object is var") // want `false`
	}
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

func detectValue() {
	valueTest("variadic value 5")           // want `true`
	valueTest(5, "variadic value 5")        // want `true`
	valueTest(2+3, 6-1, "variadic value 5") // want `true`

	valueTest(0, "variadic value 5")    // want `false`
	valueTest("5", "variadic value 5")  // want `false`
	valueTest(5, 0, "variadic value 5") // want `false`
	valueTest(0, 5, "variadic value 5") // want `false`
}

func detectType() {
	{
		var s fmt.Stringer

		typeTest(s, "is interface")
		typeTest(interface{}(nil), "is interface") // want `YES`
		typeTest(implementsAll{}, "is interface")
		typeTest(&implementsAll{}, "is interface")
		typeTest(4, "is interface")
		typeTest("", "is interface")

		typeTest(s, "underlying is interface")                // want `YES`
		typeTest(interface{}(nil), "underlying is interface") // want `YES`
		typeTest(implementsAll{}, "underlying is interface")
		typeTest(&implementsAll{}, "underlying is interface")
		typeTest(4, "underlying is interface")
		typeTest("", "underlying is interface")
	}

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

	typeTest(&implementsAll{}, "variadic implements error")                 // want `true`
	typeTest(&implementsAll{}, error(nil), "variadic implements error")     // want `true`
	typeTest(error(nil), "variadic implements error")                       // want `true`
	typeTest(errors.New("example"), "variadic implements error")            // want `true`
	typeTest(&embedImplementsAll{}, "variadic implements error")            // want `true`
	typeTest(embedImplementsAllPtr{}, "variadic implements error")          // want `true`
	typeTest(&embedImplementsAllPtr{}, "variadic implements error")         // want `true`
	typeTest(implementsAll{}, "variadic implements error")                  // want `false`
	typeTest(i1, "variadic implements error")                               // want `false`
	typeTest(implementsAllNewtype{}, "variadic implements error")           // want `false`
	typeTest(&implementsAllNewtype{}, "variadic implements error")          // want `false`
	typeTest(embedImplementsAll{}, "variadic implements error")             // want `false`
	typeTest(embedImplementsAll{}, 1, "variadic implements error")          // want `false`
	typeTest(embedImplementsAll{}, 1, "variadic implements error")          // want `false`
	typeTest(embedImplementsAll{}, error(nil), "variadic implements error") // want `false`

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

	typeTest("variadic size==4")                                 // want `true`
	typeTest([4]byte{}, "variadic size==4")                      // want `true`
	typeTest(int32(0), rune(0), [2]uint16{}, "variadic size==4") // want `true`

	typeTest([6]byte{}, "variadic size==4")            // want `false`
	typeTest(uint32(0), [6]byte{}, "variadic size==4") // want `false`
	typeTest([6]byte{}, uint32(0), "variadic size==4") // want `false`

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

	{
		type newInt int

		var x int
		y := newInt(5)

		typeTest("variadic int")              // want `true`
		typeTest(1, "variadic int")           // want `true`
		typeTest(x, 2, "variadic int")        // want `true`
		typeTest(-x, x+1, +x, "variadic int") // want `true`

		typeTest("no", "variadic int")       // want `false`
		typeTest(1, "no", "variadic int")    // want `false`
		typeTest(y, "variadic int")          // want `false`
		typeTest("no", 2, "variadic int")    // want `false`
		typeTest(1, "no", 3, "variadic int") // want `false`
		typeTest(1, 2, "no", "variadic int") // want `false`
	}

	{
		type newInt int

		var x int
		y := newInt(5)

		typeTest("variadic underlying int")              // want `true`
		typeTest(1, "variadic underlying int")           // want `true`
		typeTest(y, "variadic underlying int")           // want `true`
		typeTest(x, 2, "variadic underlying int")        // want `true`
		typeTest(-y, y+1, +x, "variadic underlying int") // want `true`

		typeTest("no", "variadic underlying int")       // want `false`
		typeTest(1, "no", "variadic underlying int")    // want `false`
		typeTest("no", 2, "variadic underlying int")    // want `false`
		typeTest(1, "no", 3, "variadic underlying int") // want `false`
		typeTest(1, 2, "no", "variadic underlying int") // want `false`
	}
}

func detectAddressable(x int, xs []int) {
	typeTest("variadic addressable")               // want `true`
	typeTest(x, "variadic addressable")            // want `true`
	typeTest(xs, x, "variadic addressable")        // want `true`
	typeTest(xs, x, xs[0], "variadic addressable") // want `true`

	typeTest(x, "", "variadic addressable")       // want `false`
	typeTest(1, x, xs[0], "variadic addressable") // want `false`
}

func detectConvertibleTo(x int, xs []int) {
	type newString string
	stringSlice := []string{""}

	typeTest("variadic convertible to string")                                                      // want `true`
	typeTest("yes", "variadic convertible to string")                                               // want `true`
	typeTest("yes", newString("yes"), "variadic convertible to string")                             // want `true`
	typeTest([]byte("yes"), newString("yes"), "", stringSlice[0], "variadic convertible to string") // want `true`

	typeTest(xs, "variadic convertible to string")        // want `false`
	typeTest(xs, "", "variadic convertible to string")    // want `false`
	typeTest("", xs[:], "variadic convertible to string") // want `false`
}

func detectAssignableTo(x int, xs []int) {
	type newString string
	type stringAlias = string
	stringSlice := []string{""}

	typeTest("variadic assignable to string")                            // want `true`
	typeTest("yes", "variadic assignable to string")                     // want `true`
	typeTest(stringAlias("yes"), "yes", "variadic assignable to string") // want `true`

	typeTest("yes", newString("yes"), "variadic assignable to string")                             // want `false`
	typeTest([]byte("yes"), newString("yes"), "", stringSlice[0], "variadic assignable to string") // want `false`
	typeTest([]byte("no"), "variadic assignable to string")                                        // want `false`
	typeTest(xs, "variadic assignable to string")                                                  // want `false`
	typeTest(xs, "", "variadic assignable to string")                                              // want `false`
	typeTest("", xs[:], "variadic assignable to string")                                           // want `false`
}

func detectConst(x int, xs []int) {
	const namedIntConst = 130

	constTest("variadic const")                                 // want `true`
	constTest(1, "variadic const")                              // want `true`
	constTest(namedIntConst, 1<<3, "variadic const")            // want `true`
	constTest(namedIntConst, namedIntConst*2, "variadic const") // want `true`
	constTest(1, unsafe.Sizeof(int(0)), 3, "variadic const")    // want `true`

	constTest(x, "variadic const")           // want `false`
	constTest(random(), "variadic const")    // want `false`
	constTest(1, random(), "variadic const") // want `false`
	constTest(random(), 2, "variadic const") // want `false`
}

func detectPure(x int, xs []int) {
	var foo struct {
		a int
	}

	xptr := &x

	pureTest(random())        // want `false`
	pureTest([]int{random()}) // want `false`

	pureTest(*xptr)               // want `true`
	pureTest(int(x))              // want `true`
	pureTest((*int)(&x))          // want `true`
	pureTest((func())(func() {})) // want `true`
	pureTest(foo.a)               // want `true`
	pureTest(x * x)               // want `true`
	pureTest((x * x))             // want `true`
	pureTest(+x)                  // want `true`
	pureTest(xs[0])               // want `true`
	pureTest(xs[x])               // want `true`
	pureTest([]int{0})            // want `true`

	pureTest("variadic pure")                     // want `true`
	pureTest("", "variadic pure")                 // want `true`
	pureTest(1, 2, "variadic pure")               // want `true`
	pureTest(1, xs[0], "variadic pure")           // want `true`
	pureTest(*xptr, xs[0], xptr, "variadic pure") // want `true`

	pureTest(random(), "variadic pure")           // want `false`
	pureTest(1, random(), "variadic pure")        // want `false`
	pureTest(random(), 2, "variadic pure")        // want `false`
	pureTest(1, xs[0], random(), "variadic pure") // want `false`
	pureTest(random(), xs[0], 3, "variadic pure") // want `false`
	pureTest(random(), random(), "variadic pure") // want `false`
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

	textTest("", "root text test") // want `YES`
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

func detectLine() {
	lineTest(1, 2, "same line") // want `YES`
	lineTest(1,
		2, "same line")

	lineTest( // want `YES`
		1,
		2,
		"different line",
	)
	lineTest(1, 2,
		"different line")
	lineTest(1, 2, "different line")
}

func detectNode() {
	var i int
	var s string
	var rows [][]byte

	var f func()

	nodeTest("3 identical expr statements in a row") // want `true`
	f()
	f()
	f()

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
