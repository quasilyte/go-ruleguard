package gocritic

import (
	"io"
	"regexp"
	"strings"
)

func selfAssign(x int, ys []string) {
	x = x         // want `warning: suspicious self-assignment in x = x`
	ys[0] = ys[0] // want `warning: suspicious self-assignment in ys\[0\] = ys\[0\]`
}

func valSwap1(x, y int) (int, int) {
	tmp := x // want `hint: can use parallel assignment like x,y=y,x`
	x = y
	y = tmp
	return x, y
}

func valSwap2(xs, ys []int) {
	if len(xs) != 0 && len(ys) != 0 {
		temp := ys[0] // want `hint: can use parallel assignment like ys\[0\],xs\[0\]=xs\[0\],ys\[0\]`
		ys[0] = xs[0]
		xs[0] = temp
	}
}

func dupArgs(xs []int, rw io.ReadWriter) {
	copy(xs, xs)    // want `warning: suspicious duplicated args in copy\(xs, xs\)`
	io.Copy(rw, rw) // want `warning: suspicious duplicated args in io\.Copy\(rw, rw\)`
}

func appendCombine1(xs []int, x, y int) []int {
	xs = append(xs, x) // want `information: xs=append\(xs,x,y\) is faster`
	xs = append(xs, y)
	return xs
}

func appendCombine2(xs []int, aa []int, bb []int) []int {
	// Can't combine here due to the variadic calls.
	xs = append(xs, aa...)
	xs = append(xs, bb...)
	return xs
}

func badCall(s string, xs []int) {
	_ = strings.Replace(s, "a", "b", 0) // want `error: n=0 argument does nothing, maybe n=-1 is indended\?`
	_ = append(xs)                      // want `error: append called with 1 argument does nothing`
}

func stringXbytes(s string, b []byte) {
	copy(b, []byte(s)) // want `hint: can write copy\(b, s\) without type conversion`
}

func assignOp(x, y int) {
	x = x + 3 // want `hint: can simplify to x\+=3`
	x = x - 4 // want `hint: can simplify to x\-=4`
	x = x * 5 // want `hint: can simplify to x\*=5`
	y = y + 1 // want `hint: can simplify to y\+\+`
}

func boolExprSimplify(a, b bool, i1, i2 int) {
	_ = !!a         // want `hint: can simplify !!a to a`
	_ = !(i1 != i2) // want `hint: can simplify !\(i1!=i2\) to i1==i2`
	_ = !(i1 == i2) // want `hint: can simplify !\(i1==i2\) to i1!=i2`
}

func dupSubExprBad(i1, i2 int) {
	_ = i1 != 0 && i1 != 1 && i1 != 0 // want `error: suspicious duplicated i1 != 0 in condition`
	_ = i1 == 0 || i1 == 0            // want `error: suspicious identical LHS and RHS`
	_ = i1 == i1                      // want `error: suspicious identical LHS and RHS`
	_ = i1 != i1                      // want `error: suspicious identical LHS and RHS`
	_ = i1 - i1                       // want `error: suspicious identical LHS and RHS`
}

func mapKey(x, y int) {
	_ = map[int]int{}
	_ = map[int]int{x + 1: 1, x + 2: 2}
	_ = map[int]int{x: 1, x: 2} // want `error: suspicious duplicate key x`
	_ = map[int]int{            // want `error: suspicious duplicate key x`
		10: 1,
		x:  2,
		30: 3,
		x:  4,
		50: 5,
	}
	_ = map[int]int{y: 1, x: 2, y: 3} // want `error: suspicious duplicate key y`
}

func regexpMust(pat string) {
	regexp.Compile(pat)   // OK: dynamic pattern
	regexp.Compile("123") // want `hint: can use MustCompile for const patterns`

	const constPat = `hello`
	regexp.CompilePOSIX(constPat) // want `hint: can use MustCompile for const patterns`
}

func yodaStyleExpr(p *int) {
	_ = nil != p // want `warning: yoda-style expression`
}
