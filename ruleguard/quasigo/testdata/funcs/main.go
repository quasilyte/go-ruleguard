package main

func ten() int { return 10 }

func helloWorld() {
	println("Hello, world!")
}

func add1(x int) int {
	return x + 1
}

func strlen(s string) int {
	return len(s)
}

func concat(s1, s2 string) string {
	return s1 + s2
}

func concat3(s1, s2, s3 string) string {
	return concat(concat(s1, s2), s3)
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func main() {
	helloWorld()
	println(ten())
	println(add1(ten()))
	println(strlen("hello"))
	println(concat("foo", "bar"))
	println(concat3("", "", ""))
	println(concat3("x", "", ""))
	println(concat3("", "x", ""))
	println(concat3("", "", "x"))
	println(concat3("a", "b", "c"))
	println(concat3("hello", "world", ""))
	println(hasPrefix("", ""))
	println(hasPrefix("", "hello"))
	println(hasPrefix("hello", ""))
	println(hasPrefix("hello", "hello"))
	println(hasPrefix("hello, world", "hello"))
}
