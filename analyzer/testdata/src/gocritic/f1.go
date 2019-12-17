package gocritic

import (
	"io"
	"regexp"
	"strings"
)

func selfAssign(x int, ys []string) {
	x = x         // want `warn: suspicious self-assignment in x = x`
	ys[0] = ys[0] // want `warn: suspicious self-assignment in ys\[0\] = ys\[0\]`
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
	copy(xs, xs)    // want `warn: suspicious duplicated args in copy\(xs, xs\)`
	io.Copy(rw, rw) // want `warn: suspicious duplicated args in io\.Copy\(rw, rw\)`
}

func appendCombine1(xs []int, x, y int) []int {
	xs = append(xs, x) // want `info: xs=append\(xs,x,y\) is faster`
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
	_ = nil != p // want `warn: yoda-style expression`
}

func underef() {
	var k *[5]int
	(*k)[2] = 3 // want `hint: explicit array deref is redundant`

	var k2 **[2]int
	_ = (**k2)[0] // want `hint: explicit array deref is redundant`
	k2ptr := &k2
	_ = (***k2ptr)[1] // want `hint: explicit array deref is redundant`
}

func unslice() {
	var s string
	_ = s[:] // want `hint: can simplify s\[:\] to s`
	_ = s[1:]
	_ = s[:1]
	_ = s

	{
		var xs []byte
		var ys []byte
		copy(
			xs[:], // want `hint: can simplify xs\[:\] to xs`
			ys[:], // want `hint: can simplify ys\[:\] to ys`
		)
	}
	{
		var xs [][]int
		_ = xs[0][:] // want `hint: can simplify xs\[0\]\[:\] to xs\[0\]`
	}

	{
		var xs []byte
		var ys []byte
		copy(xs[1:], ys[:2])
	}
	{
		var xs []int
		_ = xs[:len(xs)-1]
	}
	{
		var xs [][]int
		_ = xs[0][1:]
	}
	{
		var xs []string
		_ = xs[:0]
	}
	{
		var xs []struct{}
		_ = xs[0:]
	}
	{
		var xs map[string][][]int
		_ = xs["0"][0][:10]
	}
	{
		var xs [3]int
		_ = xs[:]
	}
}

func switchWithOneCase1(x int) {
	switch x { // want `warn: should rewrite switch statement to if statement`
	case 1:
		println("ok")
	}
}

func switchWithOneCase2(x int) {
	switch { // want `warn: should rewrite switch statement to if statement`
	case x == 1:
		println("ok")
	}
}

func typeSwitchOneCase1(x interface{}) int {
	switch x := x.(type) { // want `warn: should rewrite switch statement to if statement`
	case int:
		return x
	}
	return 0
}

func typeSwitchOneCase2(x interface{}) int {
	switch x.(type) { // want `warn: should rewrite switch statement to if statement`
	case int:
		return 1
	}
	return 0
}

func switchTrue(b bool) {
	switch true { // want `hint: can omit true in switch`
	case b:
		return
	case !b:
		panic("!b")
	}

	switch {
	}

	switch {
	case true && false:
		println("1")
	case false && true:
		fallthrough
	default:
		println("2")
	}

	switch x := 0; {
	case x < 0:
	case x > 0:
	}
}

func sloppyLen() {
	a := []int{}

	_ = len(a) >= 0 // want `error: len\(a\) >= 0 is always true`
	_ = len(a) < 0  // want `error: len\(a\) < 0 is always false`
	_ = len(a) <= 0 // want `warn: len\(a\) <= 0 is never negative, can rewrite as len\(a\)==0`
}

func newDeref() {
	_ = *new(bool)   // want `hint: replace \*new\(bool\) with false`
	_ = *new(string) // want `hint: replace \*new\(string\) with ""`
	_ = *new(int)    // want `hint: replace \*new\(int\) with 0`
}

func emptyStringTest(s string) {
	sptr := &s

	_ = len(s) == 0 // want `replace len\(s\) == 0 with len\(s\) == ""`
	_ = len(s) != 0 // want `hint: replace len\(s\) != 0 with len\(s\) != ""`

	_ = len(*sptr) == 0 // want `hint: replace len\(\*sptr\) == 0 with len\(\*sptr\) == ""`
	_ = len(*sptr) != 0 // want `hint: replace len\(\*sptr\) != 0 with len\(\*sptr\) != ""`

	_ = s == ""
	_ = s != ""

	_ = *sptr == ""
	_ = *sptr != ""

	var b []byte
	_ = len(b) == 0
	_ = len(b) != 0
}

func offBy1(xs []int, ys []string) {
	_ = xs[len(xs)] // want `error: index expr always panics; maybe you wanted xs\[len\(xs\)-1\]\?`
	_ = ys[len(ys)] // want `error: index expr always panics; maybe you wanted ys\[len\(ys\)-1\]\?`

	_ = xs[len(xs)-1]
	_ = ys[len(ys)-1]

	// Conservative with function call.
	// Might return different lengths for both calls.
	_ = makeSlice()[len(makeSlice())]

	var m map[int]int
	// Not an error. Doesn't panic.
	_ = m[len(m)]
}

func wrapperFunc(s string) {
	_ = strings.SplitN(s, ".", -1)       // want `use Split`
	_ = strings.Replace(s, "a", "b", -1) // want `use Replace`

	_ = strings.Split(s, ".")
	_ = strings.ReplaceAll(s, "a", "b")
}
