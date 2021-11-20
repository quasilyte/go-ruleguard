package main

import "fmt"

func main() {
	println(fmt.Sprintf("%s:%d", "hello", 10))

	formatString := "hello, %s!"
	println(fmt.Sprintf(formatString, "world"))

	println(fmt.Sprintf("no formatting args"))
}
