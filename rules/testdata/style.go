package target

import (
	"errors"
	"fmt"
)

func exprUnparen() {
	var f func(args ...interface{})

	f((1))      // want `\QexprUnparen: the parentheses around 1 are superfluous`
	f(1, ("x")) // want `\QexprUnparen: the parentheses around "x" are superfluous`
	f(1, ("y")) // want `\QexprUnparen: the parentheses around "y" are superfluous`
}

func emptyDecl() {
	var ()   // want `\QemptyDecl: empty var() block`
	const () // want `\QemptyDecl: empty const() block`
	type ()  // want `\QemptyDecl: empty type() block`
}

func emptyError() {
	_ = fmt.Errorf("") // want `\Qempty errors are hard to debug`
	_ = fmt.Errorf(``) // want `\Qempty errors are hard to debug`
	_ = errors.New("") // want `\Qempty errors are hard to debug`
	_ = errors.New(``) // want `\Qempty errors are hard to debug`
}

func emptySlice() {
	x := []int{}           // want `\QemptySlice: zero-length slice declaring nil slice is better
	a := make([]int, 0, 0) // want `\QemptySlice: zero-length slice declaring nil slice is better
	fmt.Println(x, a)
}
