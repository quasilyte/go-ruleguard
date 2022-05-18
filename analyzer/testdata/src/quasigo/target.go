package quasigo

import (
	"fmt"
	"sync"
)

func f() {
	var i int
	var stringer fmt.Stringer
	var err error

	test([3]int{}, "is [3]int") // want `true`
	test([2]int{}, "is [3]int")
	test(0, "is [3]int")

	test([1]int{}, "is int array") // want `true`
	test([3]int{}, "is int array") // want `true`
	test([3]string{}, "is int array")
	test([]int{}, "is int array")
	test(1, "is int array")

	test([]int{}, "is int slice") // want `true`
	test([2]int{}, "is int slice")
	test([]string{}, "is int slice")

	test("foo", "underlying type is string")           // want `true`
	test(myString("123"), "underlying type is string") // want `true`
	test(0, "underlying type is string")
	test(myEmptyStruct{}, "underlying type is string")

	test(myEmptyStruct{}, "zero sized") // want `true`
	test(struct{}{}, "zero sized")      // want `true`
	test([0]func(){}, "zero sized")     // want `true`
	test("", "zero sized")
	test(10, "zero sized")
	test(true, "zero sized")

	test(new(bool), "type is pointer")        // want `true`
	test((*int)(nil), "type is pointer")      // want `true`
	test(&myEmptyStruct{}, "type is pointer") // want `true`
	test(&i, "type is pointer")               // want `true`
	test([]int(nil), "type is pointer")
	test(interface{}(nil), "type is pointer")
	test(10, "type is pointer")

	test(10, "type is not interface")        // want `true`
	test(&i, "type is not interface")        // want `true`
	test(true, "type is not interface")      // want `true`
	test(&stringer, "type is not interface") // want `true`
	test(stringer, "type is not interface")
	test(interface{}(nil), "type is not interface")

	test(MyError(""), "type name has Error suffix")   // want `true`
	test(new(MyError), "type name has Error suffix")  // want `true`
	test(parseError{}, "type name has Error suffix")  // want `true`
	test(&parseError{}, "type name has Error suffix") // want `true`
	test(0, "type name has Error suffix")
	test((error)(nil), "type name has Error suffix")

	test((error)(nil), "type is error") // want `true`
	test(err, "type is error")          // want `true`
	test(0, "type is error")
	test("", "type is error")

	test(&err, "pointer to interface")              // want `true`
	test((*error)(nil), "pointer to interface")     // want `true`
	test(&stringer, "pointer to interface")         // want `true`
	test(new(fmt.Stringer), "pointer to interface") // want `true`
	test(0, "pointer to interface")
	test("", "pointer to interface")
	test(err, "pointer to interface")
	test(i, "pointer to interface")
	test(parseError{}, "pointer to interface")
	test(&parseError{}, "pointer to interface")

	test(stringer, "implements fmt.Stringer")           // want `true`
	test(&stringerByValue{}, "implements fmt.Stringer") // want `true`
	test(stringerByValue{}, "implements fmt.Stringer")  // want `true`
	test(&stringerByPtr{}, "implements fmt.Stringer")   // want `true`
	test(stringerByPtr{}, "implements fmt.Stringer")    // want `true`
	test(nil, "implements fmt.Stringer")
	test("", "implements fmt.Stringer")

	test(new(byte), "pointer elem value size is smaller than uintptr")        // want `true`
	test(new(int16), "pointer elem value size is smaller than uintptr")       // want `true`
	test(&stringerByPtr{}, "pointer elem value size is smaller than uintptr") // want `true`
	test(new(uintptr), "pointer elem value size is smaller than uintptr")
	test(true, "pointer elem value size is smaller than uintptr")

	// Note that new(*int) returns **int.
	test(new(***int), "indirection of 3 or more pointers") // want `true`
	test(new(**int), "indirection of 3 or more pointers")  // want `true`
	test(new(*int), "indirection of 3 or more pointers")
	test(new(int), "indirection of 3 or more pointers")
	test(true, "indirection of 3 or more pointers")

	test(withEmbeddedMutex1{}, "embeds a mutex")  // want `true`
	test(withEmbeddedMutex2{}, "embeds a mutex")  // want `true`
	test(&withEmbeddedMutex1{}, "embeds a mutex") // want `true`
	test(&withEmbeddedMutex2{}, "embeds a mutex") // want `true`
	{
		var m withEmbeddedMutex2
		test(m, "embeds a mutex")  // want `true`
		test(&m, "embeds a mutex") // want `true`
		m2 := &withEmbeddedMutex1{}
		test(m2, "embeds a mutex")  // want `true`
		test(&m2, "embeds a mutex") // OK: double pointer
	}
	test(withMutex{}, "embeds a mutex")       // OK: not embedded
	test(withNestedMutex{}, "embeds a mutex") // OK: not embedded
	test(withoutMutex{}, "embeds a mutex")    // OK: no mutex at all
	test(1, "embeds a mutex")                 // OK: not a struct
}

type myString string

type myEmptyStruct struct{}

type parseError struct{}

type MyError myString

type stringerByValue struct{}
type stringerByPtr struct{}

func (*stringerByPtr) String() string  { return "" }
func (stringerByValue) String() string { return "" }

func test(args ...interface{}) {}

type withEmbeddedMutex1 struct {
	sync.Mutex
}

type withEmbeddedMutex2 struct {
	x int
	sync.Mutex
	y int
}

type withMutex struct {
	mu sync.Mutex
	x  int
}

type withNestedMutex struct {
	nested withEmbeddedMutex1
}

type withoutMutex struct {
	x int
}
