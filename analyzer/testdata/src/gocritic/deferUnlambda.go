package gocritic

import "fmt"

func f(...interface{}) int { return 1 }

const ten = 10

func positiveTests() {
	defer func() { f() }() // want `\Qdefer f()`

	defer func() { f(1) }() // want `\Qdefer f(1)`

	defer func() { f(ten, ten+1) }() // want `\Qdefer f(ten, ten+1)`

	defer func() { fmt.Println("hello") }() // want `\Qdefer fmt.Println("hello")`
}
