package target

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
}
