package golint

import (
	"errors"
	"fmt"
	"testing"
)

func f(x int) error {
	if x > 10 {
		return errors.New(fmt.Sprintf("something %d", x)) // want `\Qshould replace error.New(fmt.Sprintf(...)) with fmt.Errorf(...)`
	}
	if x > 5 {
		return errors.New(g("blah")) // ok
	}
	if x > 4 {
		return errors.New("something else") // ok
	}
	return nil
}

func TestF(t *testing.T) error {
	x := 1
	if x > 10 {
		t.Error(fmt.Sprintf("something %d", x)) // want `\Qshould replace t.Error(fmt.Sprintf(...)) with t.Errorf(...)`
	}
	if x > 5 {
		t.Error(g("blah")) // ok
	}
	if x > 4 {
		t.Error("something else") // ok
	}
	return nil
}

func g(s string) string { return "prefix: " + s }

// func golintRange() {
// 	var m map[string]int
// 	for x, _ := range m { // `should omit 2nd value from range; this loop is equivalent to 'for x := range \.\.\.'`
// 		_ = x
// 	}
// 	var y string
// 	_ = y
// 	for y, _ = range m { // `should omit 2nd value from range; this loop is equivalent to 'for y = range \.\.\.'`
// 	}
// }

func golintIfreturn() {
	var conds []bool

	_ = func() {
		if conds[0] {
			return
		}
		println("good")
	}

	_ = func() {
		if conds[0] {
			println("ok")
		} else if conds[1] { // want `\Qif block ends with a return statement, so drop this else and outdent its block`
			return
		} else {
			println("bad")
		}
	}

	_ = func() {
		if conds[0] { // want `\Qif block ends with a return statement, so drop this else and outdent its block`
			return
		} else {
			println("bad")
		}
	}

	_ = func(cond bool) int {
		if cond { // want `\Qif block ends with a return statement, so drop this else and outdent its block`
			return 10
		} else {
			return 20
		}
	}
}
