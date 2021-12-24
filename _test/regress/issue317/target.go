package main

import (
	"fmt"
)

func main() {
	var k interface{}

	msg := k.(string)     // expects warning
	l := fmt.Sprintf(msg) //expects warning

	fmt.Println(l)
}
