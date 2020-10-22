// nolint
// This package tests rules specified in `rules.go` of ruleguard itself.
package main

import (
	"fmt"
	"os"
)

const () // want `\Qempty const() block`
var ()   // want `\Qempty var() block`
type ()  // want `\Qempty type() block`

func main() {
	// test `fmt` related fixers
	fmt.Fprint(os.Stdout, "hello")    // want `\Qsuggestion: fmt.Print("hello")`
	fmt.Fprintln(os.Stdout, "hello")  // want `\Qsuggestion: fmt.Println("hello")`
	fmt.Fprintf(os.Stdout, "%d", 123) // want `\Qsuggestion: fmt.Printf("%d", 123)`
}
