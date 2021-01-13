package target

import "fmt"

func add(x, y int) int {
	return x + y
}

func test(b bool) {
	println(add((1), 2))
	println(add(1, (2)))

	println(b == true)
	println(!!b)

	var eface interface{}
	println(&eface)

	fooPtr := &Foo{}
	foo := Foo{}
	println(fmt.Sprint(foo))
	println(fmt.Sprint(fooPtr))
	println(fmt.Sprint(0))    // Not fmt.Stringer
	println(fmt.Sprint(&foo)) // Not addressable
}

type Foo struct{}

func (*Foo) String() string { return "Foo" }
