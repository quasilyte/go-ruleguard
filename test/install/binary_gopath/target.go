package target

func add(x, y int) int {
	return x + y
}

func test() {
	println(add((1), 2))
	println(add(1, (2)))
}
