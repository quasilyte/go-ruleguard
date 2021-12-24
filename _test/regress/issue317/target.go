package main

import (
	"fmt"
	"strings"
)

func main() {
	var k interface{}

	msg := k.(string) // expects warning

	fmt.Println(msg, strings.Count("foo", "bar") >= 0)
}
