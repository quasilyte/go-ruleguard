package localfunc

import (
	"fmt"
	"io/ioutil"
)

func test(args ...interface{}) {}

func f() interface{} { return nil }

func _() {
	fmt.Println("ok")
	_ = ioutil.Discard

	var i int
	var err error

	test("both const", 1, 2)   // want `true`
	test("both const", 1, 2+2) // want `true`
	test("both const", i, 2)
	test("both const", 1, i)
	test("both const", i, i)

	test("== 10", 10)  // want `true`
	test("== 10", 9+1) // want `true`
	test("== 10", 11)
	test("== 10", i)

	test("== 0", 0)   // want `true`
	test("== 0", 1-1) // want `true`
	test("== 0", 11)
	test("== 0", i)

	test("fmt is imported") // want `true`

	test("ioutil is imported") // want `true`

	test("check precedence", 1, err)   // want `true`
	test("check precedence", 1+2, err) // want `true`
	test("check precedence", i, err)   // want `true`
	test("check precedence", err, err) // want `true`
	test("check precedence", f(), err)
	test("check precedence", 1)
	test("check precedence", 1+2)
	test("check precedence", i)
	test("check precedence", err)
	test("check precedence", f())
	test("check precedence", 1, nil)
	test("check precedence", 1+2, nil)
	test("check precedence", i, nil)
	test("check precedence", err, nil)
	test("check precedence", f(), nil)

	test("is string", "yes") // want `true`
	test("is string", `yes`) // want `true`
	test("is string", 1)
	test("is string", i)

	test("is pure call", int(0), int(1))      // want `true`
	test("is pure call", string("f"), int(1)) // want `true`
	test("is pure call", f(), f())
	test("is pure call", int(0), 1)
	test("is pure call", 0, int(1))
	test("is pure call", f(), int(1))
	test("is pure call", 1, 1)
}
