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

	var foo Foo
	fooptr := &Foo{}

	println(fmt.Sprint(0))
	println(fmt.Sprint(foo))
	println(fmt.Sprint(fooptr))
	println(fmt.Sprint(&foo))
}

type Foo struct{}

func (Foo) String() string { return "Foo" }
