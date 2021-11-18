package main

import "strings"

func main() {
	s := "Hello, World!"

	println(strings.HasPrefix("foo", "foo"))
	println(strings.HasPrefix(s, "Hello"))
	println(strings.HasPrefix(s, "$"))

	println(strings.HasSuffix("foo", "foo"))
	println(strings.HasSuffix(s, "World!"))
	println(strings.HasSuffix(s, "Hello"))

	println(strings.Contains("foo", "foo"))
	println(strings.Contains(s, ","))
	println(strings.Contains(s, "$"))

	println(strings.Replace("foo", "f", "b", 1))
	println(strings.Replace("foo", "f", "b", 0))
	println(strings.Replace("foo", "f", "b", -1))
	println(strings.Replace("foo", "o", "??", 1))
	println(strings.Replace("foo", "o", "??", 0))
	println(strings.Replace("foo", "o", "??", -1))
	println(strings.ReplaceAll("foo", "o", "f"))
	println(strings.ReplaceAll(s, "l", "12"))
	println(strings.ReplaceAll(s, "ll", ""))
}
