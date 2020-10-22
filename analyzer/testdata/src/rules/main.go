// nolint
// This package tests rules specified in `rules.go` of ruleguard itself.
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

const () // want `\Qempty const() block`
var ()   // want `\Qempty var() block`
type ()  // want `\Qempty type() block`

func badFmt() {
	fmt.Fprint(os.Stdout, "hello")    // want `\Qsuggestion: fmt.Print("hello")`
	fmt.Fprintln(os.Stdout, "hello")  // want `\Qsuggestion: fmt.Println("hello")`
	fmt.Fprintf(os.Stdout, "%d", 123) // want `\Qsuggestion: fmt.Printf("%d", 123)`
}

func badString() {
	s1 := "oh"
	s2 := "hi"
	s3 := "mark"

	strings.Replace(s1, s2, s3, -4) // want `\Qsuggestion: strings.ReplaceAll(s1, s2, s3)`
	strings.Replace(s1, s2, s3, -1) // want `\Qsuggestion: strings.ReplaceAll(s1, s2, s3)`
	strings.Replace(s1, s2, s3, 0)  // want `\Qsuggestion: strings.ReplaceAll(s1, s2, s3)`
	strings.Replace(s1, s2, s3, 1)
	strings.Replace(s1, s2, s3, 2)

	strings.SplitN(s1, s2, -13) // want `\Qsuggestion: strings.Split(s1, s2)`
	strings.SplitN(s1, s2, -1)  // want `\Qsuggestion: strings.Split(s1, s2)`
	strings.SplitN(s1, s2, 0)   // want `\Qsuggestion: strings.Split(s1, s2)`
	strings.SplitN(s1, s2, 1)   // want `\Qsuggestion: strings.SplitN(s1, s2, 2)`
	strings.SplitN(s1, s2, 2)

	strings.SplitAfterN(s1, s2, -13) // want `\Qsuggestion: strings.SplitAfter(s1, s2)`
	strings.SplitAfterN(s1, s2, -1)  // want `\Qsuggestion: strings.SplitAfter(s1, s2)`
	strings.SplitAfterN(s1, s2, 0)   // want `\Qsuggestion: strings.SplitAfter(s1, s2)`
	strings.SplitAfterN(s1, s2, 1)   // want `\Qsuggestion: strings.SplitAfterN(s1, s2, 2)`
	strings.SplitAfterN(s1, s2, 2)

	p := fmt.Println
	p(strings.Count(s1, s2) == 0) // want `\Qsuggestion: !strings.Contains(s1, s2)`
	p(strings.Count(s1, s2) > 0)  // want `\Qsuggestion: strings.Contains(s1, s2)`
	p(strings.Count(s1, s2) >= 1) // want `\Qsuggestion: strings.Contains(s1, s2)`
	p(strings.Count(s1, s2) > 1)
}

func badBytes() {
	s1 := []byte("oh")
	s2 := []byte("hi")
	s3 := []byte("mark")

	bytes.Replace(s1, s2, s3, -4) // want `\Qsuggestion: bytes.ReplaceAll(s1, s2, s3)`
	bytes.Replace(s1, s2, s3, -1) // want `\Qsuggestion: bytes.ReplaceAll(s1, s2, s3)`
	bytes.Replace(s1, s2, s3, 0)  // want `\Qsuggestion: bytes.ReplaceAll(s1, s2, s3)`
	bytes.Replace(s1, s2, s3, 1)
	bytes.Replace(s1, s2, s3, 2)

	bytes.SplitN(s1, s2, -13) // want `\Qsuggestion: bytes.Split(s1, s2)`
	bytes.SplitN(s1, s2, -1)  // want `\Qsuggestion: bytes.Split(s1, s2)`
	bytes.SplitN(s1, s2, 0)   // want `\Qsuggestion: bytes.Split(s1, s2)`
	bytes.SplitN(s1, s2, 1)   // want `\Qsuggestion: bytes.SplitN(s1, s2, 2)`
	bytes.SplitN(s1, s2, 2)

	bytes.SplitAfterN(s1, s2, -13) // want `\Qsuggestion: bytes.SplitAfter(s1, s2)`
	bytes.SplitAfterN(s1, s2, -1)  // want `\Qsuggestion: bytes.SplitAfter(s1, s2)`
	bytes.SplitAfterN(s1, s2, 0)   // want `\Qsuggestion: bytes.SplitAfter(s1, s2)`
	bytes.SplitAfterN(s1, s2, 1)   // want `\Qsuggestion: bytes.SplitAfterN(s1, s2, 2)`
	bytes.SplitAfterN(s1, s2, 2)
}
