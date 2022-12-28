package filtertest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	fileTest("f1.go") // want `true`
}

func detectFunc() {
	var fn func()

	{
		typeTest((func(int) bool)(nil), "is predicate func")    // want `true`
		typeTest((func(string) bool)(nil), "is predicate func") // want `true`
		typeTest((func() bool)(nil), "is predicate func")
		typeTest((func(int) string)(nil), "is predicate func")
		typeTest(fn, "is predicate func")
		typeTest(&fn, "is predicate func")
		typeTest(10, "is predicate func")
		typeTest("str", "is predicate func")
	}

	{
		typeTest((func(int) bool)(nil), "is func")    // want `true`
		typeTest((func(string) bool)(nil), "is func") // want `true`
		typeTest((func() bool)(nil), "is func")       // want `true`
		typeTest((func(int) string)(nil), "is func")  // want `true`
		typeTest((func())(nil), "is func")            // want `true`
		typeTest(func() {}, "is func")                // want `true`
		typeTest(fn, "is func")                       // want `true`
		typeTest(&fn, "is func")
		typeTest(53, "is func")
		typeTest([]int{1}, "is func")

	}
}

func detectObject(x int, rest ...interface{}) {
	var vec vector2D

	{
		{
			// Shadowed variadic param.
			rest := []int{1}
			objectTest(rest, "object is variadic param")
		}

		objectTest(rest, "object is variadic param")   // want `true`
		objectTest((rest), "object is variadic param") // want `true`

		objectTest(24, "object is variadic param")
		objectTest(x, "object is variadic param")
		objectTest(rest[:], "object is variadic param")
		objectTest(rest[1:], "object is variadic param")

		reassigned := rest
		objectTest(reassigned, "object is variadic param")
	}

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
	typeTest([2]int{}, "convertible to ([2]int)")      // want `true`
	typeTest(myInt2Array{}, "convertible to ([2]int)") // want `true`
	typeTest([3]int{}, "convertible to ([2]int)")

	type myIntSlice2 [][]int
	typeTest([][]int{{1}}, "convertible to [][]int")     // want `true`
	typeTest(myIntSlice2(nil), "convertible to [][]int") // want `true`
	typeTest([]int{}, "convertible to [][]int")
}

func assignableTo() {
	typeTest(map[*string]error{}, "assignable to map[*string]error") // want `true`
	typeTest(map[*string]int{}, "assignable to map[*string]error")

	typeTest(0, "assignable to interface{}")   // want `true`
	typeTest(5.6, "assignable to interface{}") // want `true`
	typeTest("", "assignable to interface{}")  // want `true`
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

func detectComparable() {
	typeTest("", "comparable") // want `true`
	typeTest(0, "comparable")  // want `true`

	type good1 struct {
		x, y int
	}
	type good2 struct {
		nested good1
		x      [2]byte
		s      string
	}
	type good3 struct {
		x *int
	}
	type good4 struct {
		*good3
		good2
	}

	typeTest(good1{}, "comparable")  // want `true`
	typeTest(good2{}, "comparable")  // want `true`
	typeTest(good3{}, "comparable")  // want `true`
	typeTest(good4{}, "comparable")  // want `true`
	typeTest(&good1{}, "comparable") // want `true`
	typeTest(&good2{}, "comparable") // want `true`
	typeTest(&good3{}, "comparable") // want `true`
	typeTest(&good4{}, "comparable") // want `true`

	var (
		g1 good1
		g2 good2
		g3 good3
		g4 good4
	)
	_ = g1 == good1{}
	_ = g2 == good2{}
	_ = g3 == good3{}
	_ = g4 == good4{}
	_ = g1 != good1{}
	_ = g2 != good2{}
	_ = g3 != good3{}
	_ = g4 != good4{}

	type bad1 struct {
		_ [1]func()
	}
	type bad2 struct {
		slice []int
	}
	type bad3 struct {
		bad2
	}

	typeTest(bad1{}, "comparable")
	typeTest(bad2{}, "comparable")
	typeTest(bad3{}, "comparable")

	typeTest(&bad1{}, "comparable") // want `true`
	typeTest(&bad2{}, "comparable") // want `true`
	typeTest(&bad3{}, "comparable") // want `true`
}

func detectType() {
	{
		var s fmt.Stringer

		typeTest(s, "is interface")
		typeTest(interface{}(nil), "is interface") // want `true`
		typeTest(implementsAll{}, "is interface")
		typeTest(&implementsAll{}, "is interface")
		typeTest(4, "is interface")
		typeTest("", "is interface")

		typeTest(s, "underlying is interface")                // want `true`
		typeTest(interface{}(nil), "underlying is interface") // want `true`
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
		typeTest(withNamedTime{}, "contains time.Time") // want `true`
		typeTest(foo, "contains time.Time")             // want `true`
		typeTest(bar, "contains time.Time")             // want `true`
		typeTest(indirectFoo2{}, "contains time.Time")  // want `true`
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
		typeTest(timeFirst{}, "starts with time.Time") // want `true`
		typeTest(foo, "starts with time.Time")         // want `true`
		typeTest(bar, "starts with time.Time")         // want `true`
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
		typeTest(struct { // want `true`
			_ string
			_ string
		}{}, "non-underlying type test; T + T")
		typeTest(intPair{}, "non-underlying type test; T + T") // type is Named, not struct
		typeTest(foo, "non-underlying type test; T + T")       // want `true`
		typeTest(bar, "non-underlying type test; T + T")       // type is Named, not struct
	}

	var i1, i2 int
	var ii []int
	var s1, s2 string
	var ss []string
	typeTest(s1 + s2) // want `\Qconcat`
	typeTest(i1 + i2) // want `\Qaddition`
	typeTest(s1 > s2) // want `\Qs1 !is(int)`
	typeTest(i1 > i2) // want `\Qi1 !is(string) && pure`
	typeTest(random() > i2)
	typeTest(ss, ss) // want `\Qss is([]string)`
	typeTest(ii, ii)
	typeTest("2 type filters", i1)
	typeTest("2 type filters", s1)
	typeTest("2 type filters", ii) // want `\Qii !is(string) && !is(int)`

	typeTest(implementsAll{}, "implements io.Reader") // want `true`
	typeTest(i1, "implements io.Reader")
	typeTest(ss, "implements io.Reader")
	typeTest(implementsAll{}, "implements foolib.Stringer") // want `true`
	typeTest(i1, "implements foolib.Stringer")
	typeTest(ss, "implements foolib.Stringer")
	typeTest(implementsAll{}, "implements error")
	typeTest(&implementsAll{}, "implements error") // want `true`
	typeTest(i1, "implements error")
	typeTest(error(nil), "implements error")            // want `true`
	typeTest(errors.New("example"), "implements error") // want `true`
	typeTest(implementsAllNewtype{}, "implements error")
	typeTest(&implementsAllNewtype{}, "implements error")
	typeTest(embedImplementsAll{}, "implements error")
	typeTest(&embedImplementsAll{}, "implements error")    // want `true`
	typeTest(embedImplementsAllPtr{}, "implements error")  // want `true`
	typeTest(&embedImplementsAllPtr{}, "implements error") // want `true`

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

	typeTest([100]byte{}, "size>=100") // want `true`
	typeTest([105]byte{}, "size>=100") // want `true`
	typeTest([10]byte{}, "size>=100")
	typeTest([100]byte{}, "size<=100") // want `true`
	typeTest([105]byte{}, "size<=100")
	typeTest([10]byte{}, "size<=100") // want `true`
	typeTest([100]byte{}, "size>100")
	typeTest([105]byte{}, "size>100") // want `true`
	typeTest([10]byte{}, "size>100")
	typeTest([100]byte{}, "size<100")
	typeTest([105]byte{}, "size<100")
	typeTest([10]byte{}, "size<100")   // want `true`
	typeTest([100]byte{}, "size==100") // want `true`
	typeTest([105]byte{}, "size==100")
	typeTest([10]byte{}, "size==100")
	typeTest([100]byte{}, "size!=100")
	typeTest([105]byte{}, "size!=100") // want `true`
	typeTest([10]byte{}, "size!=100")  // want `true`

	typeTest("variadic size==4")                                 // want `true`
	typeTest([4]byte{}, "variadic size==4")                      // want `true`
	typeTest(int32(0), rune(0), [2]uint16{}, "variadic size==4") // want `true`

	typeTest([6]byte{}, "variadic size==4")            // want `false`
	typeTest(uint32(0), [6]byte{}, "variadic size==4") // want `false`
	typeTest([6]byte{}, uint32(0), "variadic size==4") // want `false`

	var time1, time2 time.Time
	var err error
	typeTest(time1 == time2, "time==time") // want `true`
	typeTest(err == nil, "time==time")
	typeTest(nil == err, "time==time")
	typeTest(time1 != time2, "time!=time") // want `true`
	typeTest(err != nil, "time!=time")
	typeTest(nil != err, "time!=time")

	intFunc := func() int { return 10 }
	intToIntFunc := func(x int) int { return x }
	typeTest(intFunc(), "func() int")                 // want `true`
	typeTest(func() int { return 0 }(), "func() int") // want `true`
	typeTest(func() string { return "" }(), "func() int")
	typeTest(intToIntFunc(1), "func() int")

	typeTest(intToIntFunc(2), "func(int) int") // want `true`
	typeTest(intToIntFunc, "func(int) int")
	typeTest(intFunc, "func(int) int")

	var v implementsAll
	typeTest(v.String(), "func() string") // want `true`
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

	{
		type myInt int
		typeTest(1, "is numeric") // want `true`
		typeTest(myInt(1), "is numeric")
		typeTest("not numeric", "is numeric")

		typeTest(1, "underlying is numeric")        // want `true`
		typeTest(1.63, "underlying is numeric")     // want `true`
		typeTest(myInt(1), "underlying is numeric") // want `true`
		typeTest("not", "underlying is numeric")
		typeTest([]int{1}, "underlying is numeric")

		typeTest(uintptr(5), "is unsigned") // want `true`
		typeTest(uint(5), "is unsigned")    // want `true`
		typeTest(uint8(5), "is unsigned")   // want `true`
		typeTest(uint16(5), "is unsigned")  // want `true`
		typeTest(uint32(5), "is unsigned")  // want `true`
		typeTest(uint64(5), "is unsigned")  // want `true`
		typeTest(5, "is unsigned")
		typeTest(int32(5), "is unsigned")
		typeTest(int(5), "is unsigned")
		typeTest(rune(5), "is unsigned")
		typeTest("934", "is unsigned")
		typeTest(4.6, "is unsigned")

		typeTest(uint(5), "is signed")
		typeTest(uint8(5), "is signed")
		typeTest(uint16(5), "is signed")
		typeTest(uint32(5), "is signed")
		typeTest(uint64(5), "is signed")
		typeTest(uintptr(5), "is signed")
		typeTest(5, "is signed")        // want `true`
		typeTest(int8(5), "is signed")  // want `true`
		typeTest(int16(5), "is signed") // want `true`
		typeTest(int32(5), "is signed") // want `true`
		typeTest(int64(5), "is signed") // want `true`
		typeTest(int(5), "is signed")   // want `true`
		typeTest(rune(5), "is signed")  // want `true`
		typeTest("934", "is signed")
		typeTest(4.6, "is signed")
		typeTest([]int{225}, "is signed")

		typeTest(uint(5), "is float")
		typeTest(uint8(5), "is float")
		typeTest(uint16(5), "is float")
		typeTest(uint32(5), "is float")
		typeTest(uint64(5), "is float")
		typeTest(uintptr(5), "is float")
		typeTest(5, "is float")
		typeTest(int8(5), "is float")
		typeTest(int16(5), "is float")
		typeTest(int32(5), "is float")
		typeTest(int64(5), "is float")
		typeTest(int(5), "is float")
		typeTest("934", "is float")
		typeTest(4.6, "is float")          // want `true`
		typeTest(float32(4.6), "is float") // want `true`
		typeTest(float64(4.6), "is float") // want `true`
		typeTest([]int{225}, "is float")

		typeTest(5, "is int")        // want `true`
		typeTest(int8(1), "is int")  // want `true`
		typeTest(int16(1), "is int") // want `true`
		typeTest(int32(1), "is int") // want `true`
		typeTest(int64(1), "is int") // want `true`
		typeTest(rune(1), "is int")  // want `true`
		typeTest(byte(1), "is int")
		typeTest(uint(1), "is int")
		typeTest(uint8(1), "is int")
		typeTest(uint16(1), "is int")
		typeTest(uint32(1), "is int")
		typeTest(uint64(1), "is int")
		typeTest([]int{3}, "is int")
		typeTest("ds", "is int")
		typeTest(54.2, "is int")
		typeTest(float64(5.3), "is int")
		typeTest(float32(5.3), "is int")

		typeTest(5, "is uint")
		typeTest(int8(1), "is uint")
		typeTest(int16(1), "is uint")
		typeTest(int32(1), "is uint")
		typeTest(int64(1), "is uint")
		typeTest(rune(1), "is uint")
		typeTest(byte(1), "is uint")   // want `true`
		typeTest(uint(1), "is uint")   // want `true`
		typeTest(uint8(1), "is uint")  // want `true`
		typeTest(uint16(1), "is uint") // want `true`
		typeTest(uint32(1), "is uint") // want `true`
		typeTest(uint64(1), "is uint") // want `true`
		typeTest([]int{3}, "is uint")
		typeTest("ds", "is uint")
		typeTest(54.2, "is uint")
		typeTest(float64(5.3), "is uint")
		typeTest(float32(5.3), "is uint")
	}

	{
		const untypedStr = "123"
		type myString string

		type scalarObject struct {
			x int
			y int
		}

		type withPointers struct {
			x int
			y *int
		}

		var intarr [4]int
		var intptrarr [4]*int
		var r io.Reader
		var ch chan int

		typeTest(1, "pointer-free")                  // want `true`
		typeTest(1.6, "pointer-free")                // want `true`
		typeTest(true, "pointer-free")               // want `true`
		typeTest(scalarObject{1, 2}, "pointer-free") // want `true`
		typeTest(intarr, "pointer-free")             // want `true`
		typeTest([2]scalarObject{}, "pointer-free")  // want `true`

		typeTest(withPointers{}, "pointer-free")
		typeTest(&withPointers{}, "pointer-free")
		typeTest(ch, "pointer-free")
		typeTest(r, "pointer-free")
		typeTest(&r, "pointer-free")
		typeTest(&intarr, "pointer-free")
		typeTest(intptrarr, "pointer-free")
		typeTest(&intptrarr, "pointer-free")
		typeTest(&scalarObject{1, 2}, "pointer-free")
		typeTest("str", "pointer-free")
		typeTest(untypedStr, "pointer-free")
		typeTest(myString("123"), "pointer-free")
		typeTest(unsafe.Pointer(nil), "pointer-free")
		typeTest(nil, "pointer-free")
		typeTest([]int{1}, "pointer-free")
		typeTest([]string{""}, "pointer-free")
		typeTest(map[string]string{}, "pointer-free")
		typeTest(new(int), "pointer-free")
		typeTest(new(string), "pointer-free")

		typeTest(1, "has pointers")
		typeTest(1.6, "has pointers")
		typeTest(true, "has pointers")
		typeTest(scalarObject{1, 2}, "has pointers")
		typeTest(intarr, "has pointers")
		typeTest([2]scalarObject{}, "has pointers")

		typeTest(withPointers{}, "has pointers")      // want `true`
		typeTest(&withPointers{}, "has pointers")     // want `true`
		typeTest(ch, "has pointers")                  // want `true`
		typeTest(r, "has pointers")                   // want `true`
		typeTest(&r, "has pointers")                  // want `true`
		typeTest(&intarr, "has pointers")             // want `true`
		typeTest(intptrarr, "has pointers")           // want `true`
		typeTest(&intptrarr, "has pointers")          // want `true`
		typeTest(&scalarObject{1, 2}, "has pointers") // want `true`
		typeTest("str", "has pointers")               // want `true`
		typeTest(untypedStr, "has pointers")          // want `true`
		typeTest(myString("123"), "has pointers")     // want `true`
		typeTest(unsafe.Pointer(nil), "has pointers") // want `true`
		typeTest(nil, "has pointers")                 // want `true`
		typeTest([]int{1}, "has pointers")            // want `true`
		typeTest([]string{""}, "has pointers")        // want `true`
		typeTest(map[string]string{}, "has pointers") // want `true`
		typeTest(new(int), "has pointers")            // want `true`
		typeTest(new(string), "has pointers")         // want `true`
	}
}

func detectHasMethod() {
	type embedsStringWriter struct {
		io.StringWriter
	}

	type embedsBuffer struct {
		bytes.Buffer
	}

	type embedsBufferPtr struct {
		*bytes.Buffer
	}

	{
		var buf bytes.Buffer
		bufPtr := &buf
		typeTest(buf, "has WriteString method")    // want `true`
		typeTest(bufPtr, "has WriteString method") // want `true`
		typeTest(buf, "has String method")         // want `true`
		typeTest(bufPtr, "has String method")      // want `true`
		buf.WriteString("")
		bufPtr.WriteString("")
	}
	{
		var w io.StringWriter
		wPtr := &w
		typeTest(w, "has WriteString method") // want `true`
		typeTest(wPtr, "has WriteString method")
		typeTest(w, "has String method")
		typeTest(wPtr, "has String method")
		w.WriteString("")
	}
	{
		var e embedsStringWriter
		ePtr := &e
		typeTest(e, "has WriteString method")    // want `true`
		typeTest(ePtr, "has WriteString method") // want `true`
		typeTest(e, "has String method")
		typeTest(ePtr, "has String method")
		e.WriteString("")
		ePtr.WriteString("")
	}
	{
		var e embedsBuffer
		ePtr := &e
		typeTest(e, "has WriteString method")    // want `true`
		typeTest(ePtr, "has WriteString method") // want `true`
		typeTest(e, "has String method")         // want `true`
		typeTest(ePtr, "has String method")      // want `true`
		e.WriteString("")
		ePtr.WriteString("")
	}
	{
		var e embedsBufferPtr
		ePtr := &e
		typeTest(e, "has WriteString method")    // want `true`
		typeTest(ePtr, "has WriteString method") // want `true`
		typeTest(e, "has String method")         // want `true`
		typeTest(ePtr, "has String method")      // want `true`
		e.WriteString("")
		ePtr.WriteString("")
	}

	{
		typeTest(1, "has WriteString method")
		typeTest("", "has WriteString method")
		typeTest(1, "has String method")
		typeTest("", "has String method")
	}
	{
		var w io.Writer
		typeTest(w, "has WriteString method")
	}
	{
		type withBufferField struct {
			buf *bytes.Buffer
		}
		var x withBufferField
		xPtr := &x
		typeTest(x, "has WriteString method")
		typeTest(xPtr, "has WriteString method")
		typeTest(x, "has String method")
		typeTest(xPtr, "has String method")
	}
	{
		var w io.StringWriter
		var eface interface{} = w
		typeTest(eface, "has WriteString method")
		typeTest(eface, "has String method")
	}
}

func detectSameTypeSizes() {
	{
		typeTest(int8(1), int8(2), "same type sizes")         // want `true`
		typeTest(float64(1), float64(2.5), "same type sizes") // want `true`
		typeTest(float32(1), float64(2), "same type sizes")
	}

	{
		var s string
		var b []byte
		typeTest(s, s, "same type sizes") // want `true`
		typeTest(b, b, "same type sizes") // want `true`
		typeTest(s, b, "same type sizes")
	}

	{
		var a10 [10]byte
		var a15 [15]byte
		typeTest(a10, a10, "same type sizes")       // want `true`
		typeTest(a15, a15, "same type sizes")       // want `true`
		typeTest(a15[:], a15[:], "same type sizes") // want `true`
		typeTest(a10[:], a15[:], "same type sizes") // want `true`
		typeTest(a15[:], a10[:], "same type sizes") // want `true`
		typeTest(a15, a10, "same type sizes")
		typeTest(a10, a15, "same type sizes")
	}

	{
		type vector2 struct {
			x, y float64
		}
		type vector3 struct {
			x, y, z float64
		}
		var a, b vector2
		typeTest(a, b, "same type sizes")                 // want `true`
		typeTest(vector2{}, vector2{}, "same type sizes") // want `true`
		typeTest(vector3{}, vector3{}, "same type sizes") // want `true`
		typeTest(vector2{}, vector3{}, "same type sizes")
		typeTest(vector3{}, vector2{}, "same type sizes")
		typeTest(vector2{}, 14, "same type sizes")
		typeTest(14, vector2{}, "same type sizes")
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

func detectIdenticalTypes() {
	{
		typeTest(1, 1, "identical types")                  // want `true`
		typeTest("a", "b", "identical types")              // want `true`
		typeTest([]int{}, []int{1}, "identical types")     // want `true`
		typeTest([]int32{}, []int32{1}, "identical types") // want `true`
		typeTest([]int32{}, []rune{}, "identical types")   // want `true`
		typeTest([4]int{}, [4]int{}, "identical types")    // want `true`

		typeTest(1, 1.5, "identical types")
		typeTest("ok", 1.5, "identical types")
		typeTest([]int{}, []int32{1}, "identical types")
		typeTest([4]int{}, [3]int{}, "identical types")

		type point struct {
			x, y float64
		}
		type myString string
		var s string
		var myStr myString
		var pt point
		ppt := &point{}
		var eface interface{}
		var i int
		typeTest([]point{}, []point{}, "identical types") // want `true`
		typeTest(s, s, "identical types")                 // want `true`
		typeTest(myStr, myStr, "identical types")         // want `true`
		typeTest(pt, pt, "identical types")               // want `true`
		typeTest(&pt, ppt, "identical types")             // want `true`
		typeTest(eface, eface, "identical types")         // want `true`
		typeTest([]point{}, [1]point{}, "identical types")
		typeTest(myStr, s, "identical types")
		typeTest(s, myStr, "identical types")
		typeTest(ppt, pt, "identical types")
		typeTest(eface, pt, "identical types")
		typeTest(eface, ppt, "identical types")
		typeTest(i, eface, "identical types")
		typeTest(eface, i, "identical types")
	}
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
	textTest(foo, "text=foo") // want `true`
	textTest(bar, "text=foo")

	textTest("foo", "text='foo'") // want `true`
	textTest("bar", "text='foo'")

	textTest("bar", "text!='foo'") // want `true`
	textTest("foo", "text!='foo'")

	textTest(32, "matches d+") // want `true`
	textTest(0x32, "matches d+")
	textTest("foo", "matches d+")

	textTest(1, "doesn't match [A-Z]") // want `true`
	textTest("ABC", "doesn't match [A-Z]")

	textTest("", "root text test") // want `true`
}

func detectParensFilter() {
	var err error
	parensFilterTest(err, "type is error") // want `true`
}

func fileFilters1() {
	// No matches as this file doesn't import "path/filepath".
	importsTest(os.PathSeparator, "path/filepath")
	importsTest(os.PathListSeparator, "path/filepath")
}

func detectLine() {
	lineTest(1, 2, "same line") // want `true`
	lineTest(1,
		2, "same line")

	lineTest( // want `true`
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

	nodeTest("123", "Expr") // want `true`
	nodeTest(`123`, "Expr") // want `true`
	nodeTest(12, "Expr")    // want `true`
	nodeTest(1.56, "Expr")  // want `true`
	nodeTest(1+2, "Expr")   // want `true`
	nodeTest(i, "Expr")     // want `true`
	nodeTest(s, "Expr")     // want `true`

	nodeTest("123", "BasicLit") // want `true`
	nodeTest(`123`, "BasicLit") // want `true`
	nodeTest(12, "BasicLit")    // want `true`
	nodeTest(1.56, "BasicLit")  // want `true`
	nodeTest(1+2, "BasicLit")
	nodeTest(i, "BasicLit")
	nodeTest(s, "BasicLit")

	nodeTest("123", "Ident")
	nodeTest(12, "Ident")
	nodeTest(i, "Ident") // want `true`
	nodeTest(s, "Ident") // want `true`

	nodeTest("42", "!Ident") // want `true`
	nodeTest(12, "!Ident")   // want `true`
	nodeTest(s[0], "!Ident") // want `true`
	nodeTest(i, "!Ident")
	nodeTest(s, "!Ident")

	nodeTest(s[0], "IndexExpr")       // want `true`
	nodeTest(rows[0][5], "IndexExpr") // want `true`
	nodeTest("42", "IndexExpr")
}

var globalVar string
var globalVar2 string = time.Now().String() // want `\Qglobal var`
var globalVar3 = time.Now().String()        // want `\Qglobal var`
var (
	globalVar4 string
)

func detectGlobal() {
	globalVar = time.Now().String()  // want `\Qglobal var`
	globalVar4 = time.Now().String() // want `\Qglobal var`
	{
		globalVar := time.Now().String() // shadowed global var
		print(globalVar)
	}
	{
		var globalVar = time.Now().String() // shadowed global var
		print(globalVar)
	}
}

func detectSinkType() {
	// Call argument context.
	_ = acceptReader(newIface("sink is io.Reader").(*bytes.Buffer)) // want `true`
	_ = acceptReader(newIface("sink is io.Reader").(io.Reader))     // want `true`
	_ = acceptReader((newIface("sink is io.Reader").(io.Reader)))   // want `true`
	_ = acceptBuffer(newIface("sink is io.Reader").(*bytes.Buffer))
	_ = acceptReaderVariadic(10, newIface("sink is io.Reader").(*bytes.Buffer))      // want `true`
	_ = acceptReaderVariadic(10, newIface("sink is io.Reader").(*bytes.Buffer))      // want `true`
	_ = acceptReaderVariadic(10, nil, newIface("sink is io.Reader").(*bytes.Buffer)) // want `true`
	_ = acceptReaderVariadic(10, newIface("sink is io.Reader").([]io.Reader)...)
	_ = acceptWriterVariadic(10, newIface("sink is io.Reader").(*bytes.Buffer))
	_ = acceptWriterVariadic(10, nil, newIface("sink is io.Reader").(*bytes.Buffer))
	_ = acceptWriterVariadic(10, nil, nil, newIface("sink is io.Reader").(*bytes.Buffer))
	_ = acceptVariadic(10, newIface("sink is io.Reader").(*bytes.Buffer))
	_ = acceptVariadic(10, nil, newIface("sink is io.Reader").(*bytes.Buffer))
	_ = acceptVariadic(10, nil, nil, newIface("sink is io.Reader").(*bytes.Buffer))
	fmt.Println(newIface("sink is interface{}").(int))          // want `true`
	fmt.Println(1, newIface("sink is interface{}").(io.Reader)) // want `true`

	// Type conversion context.
	_ = io.Reader(newIface("sink is io.Reader").(*bytes.Buffer)) // want `true`
	_ = io.Writer(newIface("sink is io.Reader").(*bytes.Buffer))

	// Return stmt context.
	{
		_ = func() (io.Reader, io.Writer) {
			return newIface("sink is io.Reader").(*bytes.Buffer), nil // want `true`
		}
		_ = func() (io.Reader, io.Writer) {
			return nil, newIface("sink is io.Reader").(*bytes.Buffer)
		}
		_ = func() (io.Writer, io.Reader) {
			return nil, newIface("sink is io.Reader").(*bytes.Buffer) // want `true`
		}
	}

	// Assignment context.
	{
		var r io.Reader = (newIface("sink is io.Reader").(*bytes.Buffer)) // want `true`
		var _ io.Reader = newIface("sink is io.Reader").(*bytes.Buffer)   // want `true`
		var w io.Writer = newIface("sink is io.Reader").(*bytes.Buffer)
		x := newIface("sink is io.Reader").(*bytes.Buffer)
		_ = r
		_ = w
		_ = x
		var readers map[string]io.Reader
		readers["foo"] = newIface("sink is io.Reader").(*bytes.Buffer) // want `true`
		var writers map[string]io.Writer
		writers["foo"] = newIface("sink is io.Reader").(*bytes.Buffer)
		var foo exampleStruct
		foo.r = newIface("sink is io.Reader").(*bytes.Buffer) // want `true`
		foo.buf = newIface("sink is io.Reader").(*bytes.Buffer)
		foo.w = newIface("sink is io.Reader").(*bytes.Buffer)
	}

	// Index expr context
	{
		var readerKeys map[io.Reader]string
		readerKeys[newIface("sink is io.Reader").(*bytes.Buffer)] = "ok"   // want `true`
		readerKeys[(newIface("sink is io.Reader").(*bytes.Buffer))] = "ok" // want `true`
		var writerKeys map[io.Writer]string
		writerKeys[newIface("sink is io.Reader").(*bytes.Buffer)] = "ok"
		writerKeys[(newIface("sink is io.Reader").(*bytes.Buffer))] = "ok"
	}

	// Composite lit element context.
	_ = []io.Reader{
		newIface("sink is io.Reader").(*bytes.Buffer), // want `true`
	}
	_ = []io.Reader{
		10: newIface("sink is io.Reader").(*bytes.Buffer), // want `true`
	}
	_ = [10]io.Reader{
		4: newIface("sink is io.Reader").(*bytes.Buffer), // want `true`
	}
	_ = map[string]io.Reader{
		"foo": newIface("sink is io.Reader").(*bytes.Buffer), // want `true`
	}
	_ = map[io.Reader]string{
		newIface("sink is io.Reader").(*bytes.Buffer): "foo", // want `true`
	}
	_ = map[io.Reader]string{
		(newIface("sink is io.Reader").(*bytes.Buffer)): "foo", // want `true`
	}
	_ = []io.Writer{
		(newIface("sink is io.Reader").(*bytes.Buffer)),
	}
	_ = exampleStruct{
		w: newIface("sink is io.Reader").(*bytes.Buffer),
		r: newIface("sink is io.Reader").(*bytes.Buffer), // want `true`
	}
	_ = []interface{}{
		newIface("sink is interface{}").(*bytes.Buffer), // want `true`
		newIface("sink is interface{}").(int),           // want `true`
	}
}

func detectSinkType2() io.Reader {
	return newIface("sink is io.Reader").(*bytes.Buffer) // want `true`
}

func detectSinkType3() io.Writer {
	return newIface("sink is io.Reader").(*bytes.Buffer)
}

func newIface(key string) interface{} { return nil }

func acceptReaderVariadic(a int, r ...io.Reader) int { return 0 }
func acceptWriterVariadic(a int, r ...io.Writer) int { return 0 }
func acceptVariadic(a int, r ...interface{}) int     { return 0 }
func acceptReader(r io.Reader) int                   { return 0 }
func acceptWriter(r io.Writer) int                   { return 0 }
func acceptBuffer(b *bytes.Buffer) int               { return 0 }

type exampleStruct struct {
	r   io.Reader
	w   io.Writer
	buf *bytes.Buffer
}
