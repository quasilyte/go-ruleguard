package gocritic

import (
	"flag"
	"io"
	"regexp"
	"strings"
)

func selfAssign(x int, ys []string) {
	x = x         // want `suspicious self-assignment in x = x`
	ys[0] = ys[0] // want `suspicious self-assignment in ys\[0\] = ys\[0\]`
}

func valSwap1(x, y int) (int, int) {
	tmp := x // want `can use parallel assignment like x,y=y,x`
	x = y
	y = tmp
	return x, y
}

func valSwap2(xs, ys []int) {
	if len(xs) != 0 && len(ys) != 0 {
		temp := ys[0] // want `can use parallel assignment like ys\[0\],xs\[0\]=xs\[0\],ys\[0\]`
		ys[0] = xs[0]
		xs[0] = temp
	}
}

func dupArgs(xs []int, rw io.ReadWriter) {
	copy(xs, xs)    // want `suspicious duplicated args in copy\(xs, xs\)`
	io.Copy(rw, rw) // want `suspicious duplicated args in io\.Copy\(rw, rw\)`
}

func appendCombine1(xs []int, x, y int) []int {
	xs = append(xs, x) // want `xs=append\(xs,x,y\) is faster`
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
	_ = strings.Replace(s, "a", "b", 0) // want `n=0 argument does nothing, maybe n=-1 is indended\?`
	_ = append(xs)                      // want `append called with 1 argument does nothing`
}

func stringXbytes(s string, b []byte) {
	copy(b, []byte(s)) // want `can write copy\(b, s\) without type conversion`
}

func assignOp(x, y int) {
	x = x + 3 // want `can simplify to x\+=3`
	x = x - 4 // want `can simplify to x\-=4`
	x = x * 5 // want `can simplify to x\*=5`
	y = y + 1 // want `can simplify to y\+\+`
}

func boolExprSimplify(a, b bool, i1, i2 int) {
	_ = !!a         // want `can simplify !!a to a`
	_ = !(i1 != i2) // want `can simplify !\(i1!=i2\) to i1==i2`
	_ = !(i1 == i2) // want `can simplify !\(i1==i2\) to i1!=i2`
}

func dupSubExprBad(i1, i2 int) {
	_ = i1 != 0 && i1 != 1 && i1 != 0 // want `suspicious duplicated i1 != 0 in condition`
	_ = i1 == 0 || i1 == 0            // want `suspicious identical LHS and RHS`
	_ = i1 == i1                      // want `suspicious identical LHS and RHS`
	_ = i1 != i1                      // want `suspicious identical LHS and RHS`
	_ = i1 - i1                       // want `suspicious identical LHS and RHS`
}

func mapKey(x, y int) {
	_ = map[int]int{}
	_ = map[int]int{x + 1: 1, x + 2: 2}
	_ = map[int]int{x: 1, x: 2} // want `suspicious duplicate key x`
	_ = map[int]int{
		10: 1,
		x:  2, // want `suspicious duplicate key x`
		30: 3,
		x:  4,
		50: 5,
	}
	_ = map[int]int{y: 1, x: 2, y: 3} // want `suspicious duplicate key y`
}

func regexpMust(pat string) {
	regexp.Compile(pat)   // OK: dynamic pattern
	regexp.Compile("123") // want `can use MustCompile for const patterns`

	const constPat = `hello`
	regexp.CompilePOSIX(constPat) // want `can use MustCompile for const patterns`
}

func yodaStyleExpr(p *int) {
	_ = nil != p // want `yoda-style expression`
}

func underef() {
	var k *[5]int
	(*k)[2] = 3 // want `explicit array deref is redundant`

	var k2 **[2]int
	_ = (**k2)[0] // want `explicit array deref is redundant`
	k2ptr := &k2
	_ = (***k2ptr)[1] // want `explicit array deref is redundant`
}

func unslice() {
	var s string
	_ = s[:] // want `can simplify s\[:\] to s`
	_ = s[1:]
	_ = s[:1]
	_ = s

	{
		var xs []byte
		var ys []byte
		copy(
			xs[:], // want `can simplify xs\[:\] to xs`
			ys[:], // want `can simplify ys\[:\] to ys`
		)
	}
	{
		var xs [][]int
		_ = xs[0][:] // want `can simplify xs\[0\]\[:\] to xs\[0\]`
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
	switch x { // want `should rewrite switch statement to if statement`
	case 1:
		println("ok")
	}
}

func switchWithOneCase2(x int) {
	switch { // want `should rewrite switch statement to if statement`
	case x == 1:
		println("ok")
	}
}

func typeSwitchOneCase1(x interface{}) int {
	switch x := x.(type) { // want `should rewrite switch statement to if statement`
	case int:
		return x
	}
	return 0
}

func typeSwitchOneCase2(x interface{}) int {
	switch x.(type) { // want `should rewrite switch statement to if statement`
	case int:
		return 1
	}
	return 0
}

func switchTrue(b bool) {
	switch true { // want `can omit true in switch`
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

	_ = len(a) >= 0 // want `len\(a\) >= 0 is always true`
	_ = len(a) < 0  // want `len\(a\) < 0 is always false`
	_ = len(a) <= 0 // want `len\(a\) <= 0 is never negative, can rewrite as len\(a\)==0`
}

func newDeref() {
	_ = *new(bool)   // want `replace \*new\(bool\) with false`
	_ = *new(string) // want `replace \*new\(string\) with ""`
	_ = *new(int)    // want `replace \*new\(int\) with 0`
}

func emptyStringTest(s string) {
	sptr := &s

	_ = len(s) == 0 // want `replace len\(s\) == 0 with len\(s\) == ""`
	_ = len(s) != 0 // want `replace len\(s\) != 0 with len\(s\) != ""`

	_ = len(*sptr) == 0 // want `replace len\(\*sptr\) == 0 with len\(\*sptr\) == ""`
	_ = len(*sptr) != 0 // want `replace len\(\*sptr\) != 0 with len\(\*sptr\) != ""`

	_ = s == ""
	_ = s != ""

	_ = *sptr == ""
	_ = *sptr != ""

	var b []byte
	_ = len(b) == 0
	_ = len(b) != 0
}

func offBy1(xs []int, ys []string) {
	_ = xs[len(xs)] // want `index expr always panics; maybe you wanted xs\[len\(xs\)-1\]\?`
	_ = ys[len(ys)] // want `index expr always panics; maybe you wanted ys\[len\(ys\)-1\]\?`

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

func flagDeref() {
	_ = *flag.Bool("b", false, "")  // want `immediate deref in \*flag\.Bool\("b", false, ""\) is most likely an error`
	_ = *flag.Duration("d", 0, "")  // want `immediate deref in \*flag\.Duration\("d", 0, ""\) is most likely an error`
	_ = *flag.Float64("f64", 0, "") // want `immediate deref in \*flag\.Float64\("f64", 0, ""\) is most likely an error`
	_ = *flag.Int("i", 0, "")       // want `immediate deref in \*flag\.Int\("i", 0, ""\) is most likely an error`
	_ = *flag.Int64("i64", 0, "")   // want `immediate deref in \*flag\.Int64\("i64", 0, ""\) is most likely an error`
	_ = *flag.String("s", "", "")   // want `immediate deref in \*flag\.String\("s", "", ""\) is most likely an error`
	_ = *flag.Uint("u", 0, "")      // want `immediate deref in \*flag\.Uint\("u", 0, ""\) is most likely an error`
	_ = *flag.Uint64("u64", 0, "")  // want `immediate deref in \*flag\.Uint64\("u64", 0, ""\) is most likely an error`
}

type object struct {
	data *byte
}

func suspiciousReturns() {
	_ = func(err error) error {
		if err == nil { // want `returned expr is always nil; replace err with nil`
			return err
		}
		return nil
	}

	_ = func(o *object) *object {
		if o == nil { // want `returned expr is always nil; replace o with nil`
			return o
		}
		return &object{}
	}

	_ = func(o *object) *byte {
		if o.data == nil { // want `returned expr is always nil; replace o.data with nil`
			return o.data
		}
		return nil
	}

	_ = func(pointers [][][]map[string]*int) *int {
		if pointers[0][1][2]["ptr"] == nil { // want `returned expr is always nil; replace pointers\[0\]\[1\]\[2\]\["ptr"\] with nil`
			return pointers[0][1][2]["ptr"]
		}
		if ptr := pointers[0][1][2]["ptr"]; ptr == nil { // want `returned expr is always nil; replace ptr with nil`
			return ptr
		}
		return nil
	}
}

func explicitNil() {
	_ = func(err error) error {
		if err == nil {
			return nil
		}
		return nil
	}

	_ = func(o *object) *object {
		if o == nil {
			return nil
		}
		return &object{}
	}

	_ = func(o *object) *byte {
		if o.data == nil {
			return nil
		}
		return nil
	}

	_ = func(pointers [][][]map[string]*int) *int {
		if pointers[0][1][2]["ptr"] == nil {
			return nil
		}
		if ptr := pointers[0][1][2]["ptr"]; ptr == nil {
			return nil
		}
		return nil
	}
}

func explicitNotEqual() {
	_ = func(err error) error {
		if err != nil {
			return err
		}
		return nil
	}

	_ = func(o *object) *object {
		if o != nil {
			return o
		}
		return &object{}
	}

	_ = func(o *object) *byte {
		if o.data != nil {
			return o.data
		}
		return nil
	}

	_ = func(pointers [][][]map[string]*int) *int {
		if pointers[0][1][2]["ptr"] != nil {
			return pointers[0][1][2]["ptr"]
		}
		if ptr := pointers[0][1][2]["ptr"]; ptr != nil {
			return ptr
		}
		return nil
	}
}

func rangeExprCopy() {
	// OK: returned valus is not addressible, can't take address.
	for _, x := range makeArray() {
		_ = x
	}

	{
		var xs [200]byte
		// OK: already iterating over a pointer.
		for _, x := range &xs {
			_ = x
		}
		// OK: only index is used. No copy is generated.
		for i := range xs {
			_ = xs[i]
		}
		// OK: like in case above, no copy, so it's OK.
		for range xs {
		}
	}

	{
		var xs [10]byte
		// OK: xs is a very small array that can be trivially copied.
		for _, x := range xs {
			_ = x
		}
	}

	{
		var xs [777]byte
		for _, x := range xs { // want `xs copy can be avoided with &xs`
			_ = x
		}
	}

	{
		var foo struct {
			arr [768]byte
		}
		for _, x := range foo.arr { // want `foo\.arr copy can be avoided with &foo\.arr`
			_ = x
		}
	}

	{
		xsList := make([][512]byte, 1)
		for _, x := range xsList[0] { // want `xsList\[0\] copy can be avoided with &xsList\[0\]`
			_ = x
		}
	}
}
