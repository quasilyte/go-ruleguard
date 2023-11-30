package gocritic

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func selfAssign(x int, ys []string) {
	x = x         // want `\Qsuspicious self-assignment in x = x`
	ys[0] = ys[0] // want `\Qsuspicious self-assignment in ys[0] = ys[0]`
}

func valSwap1(x, y int) (int, int) {
	tmp := x // want `\Qcan use parallel assignment like x,y=y,x`
	x = y
	y = tmp
	return x, y
}

func valSwap2(xs, ys []int) {
	if len(xs) != 0 && len(ys) != 0 {
		temp := ys[0] // want `\Qcan use parallel assignment like ys[0],xs[0]=xs[0],ys[0]`
		ys[0] = xs[0]
		xs[0] = temp
	}

	{
		temp := ys[0] // want `\Qcan use parallel assignment like ys[0],xs[0]=xs[0],ys[0]`
		ys[0] = xs[0]
		xs[0] = temp

		temp2 := ys[0] // want `\Qcan use parallel assignment like ys[0],xs[0]=xs[0],ys[0]`
		ys[0] = xs[0]
		xs[0] = temp2
	}
}

func dupArgs(xs []int, rw io.ReadWriter) {
	copy(xs, xs)    // want `\Qsuspicious duplicated args in copy(xs, xs)`
	io.Copy(rw, rw) // want `\Qsuspicious duplicated args in io.Copy(rw, rw)`
}

func appendCombine1(xs []int, x, y int) []int {
	xs = append(xs, x) // want `\Qxs=append(xs,x,y) is faster`
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
	_ = strings.Replace(s, "a", "b", 0) // want `\Qn=0 argument does nothing, maybe n=-1 is intended?`
	_ = append(xs)                      // want `\Qappend called with 1 argument does nothing`
}

func stringXbytes(s string, b []byte) {
	copy(b, []byte(s)) // want `\Qcan write copy(b, s) without type conversion`
}

func assignOp(x, y int) {
	x = x + 3 // want `\Qcan simplify to x+=3`
	x = x - 4 // want `\Qcan simplify to x-=4`
	x = x * 5 // want `\Qcan simplify to x*=5`
	y = y + 1 // want `\Qcan simplify to y++`
}

func boolExprSimplify(a, b bool, i1, i2 int) {
	_ = !!a         // want `\Qcan simplify !!a to a`
	_ = !(i1 != i2) // want `\Qcan simplify !(i1!=i2) to i1==i2`
	_ = !(i1 == i2) // want `\Qcan simplify !(i1==i2) to i1!=i2`
}

func dupSubExprBad(i1, i2 int) {
	_ = i1 != 0 && i1 != 1 && i1 != 0 // want `\Qsuspicious duplicated i1 != 0 in condition`
	_ = i1 == 0 || i1 == 0            // want `\Qsuspicious identical LHS and RHS`
	_ = i1 == i1                      // want `\Qsuspicious identical LHS and RHS`
	_ = i1 != i1                      // want `\Qsuspicious identical LHS and RHS`
	_ = i1 - i1                       // want `\Qsuspicious identical LHS and RHS`
}

func mapKey(x, y int) {
	_ = map[int]int{}
	_ = map[int]int{x + 1: 1, x + 2: 2}
	_ = map[int]int{x: 1, x: 2} // want `\Qsuspicious duplicate key x`
	_ = map[int]int{
		10: 1,
		x:  2, // want `\Qsuspicious duplicate key x`
		30: 3,
		x:  4,
		50: 5,
	}
	_ = map[int]int{y: 1, x: 2, y: 3} // want `\Qsuspicious duplicate key y`
}

func regexpMust(pat string) {
	regexp.Compile(pat)   // OK: dynamic pattern
	regexp.Compile("123") // want `\Qcan use MustCompile for const patterns`

	const constPat = `hello`
	regexp.CompilePOSIX(constPat) // want `\Qcan use MustCompile for const patterns`
}

func yodaStyleExpr(p *int) {
	_ = nil != p // want `\Qyoda-style expression`
}

func underef() {
	var k *[5]int
	(*k)[2] = 3 // want `\Qexplicit array deref is redundant`

	var k2 **[2]int
	_ = (**k2)[0] // want `\Qexplicit array deref is redundant`
	k2ptr := &k2
	_ = (***k2ptr)[1] // want `\Qexplicit array deref is redundant`
}

func unslice() {
	var s string
	_ = s[:] // want `\Qcan simplify s[:] to s`
	_ = s[1:]
	_ = s[:1]
	_ = s

	{
		var xs []byte
		var ys []byte
		copy(
			xs[:], // want `\Qcan simplify xs[:] to xs`
			ys[:], // want `\Qcan simplify ys[:] to ys`
		)
	}
	{
		var xs [][]int
		_ = xs[0][:] // want `\Qcan simplify xs[0][:] to xs[0]`
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
	switch x { // want `\Qshould rewrite switch statement to if statement`
	case 1:
		println("ok")
	}
}

func switchWithOneCase2(x int) {
	switch { // want `\Qshould rewrite switch statement to if statement`
	case x == 1:
		println("ok")
	}
}

func typeSwitchOneCase1(x interface{}) int {
	switch x := x.(type) { // want `\Qshould rewrite switch statement to if statement`
	case int:
		return x
	}
	return 0
}

func typeSwitchOneCase2(x interface{}) int {
	switch x.(type) { // want `\Qshould rewrite switch statement to if statement`
	case int:
		return 1
	}
	return 0
}

func switchTrue(b bool) {
	switch true { // want `\Qcan omit true in switch`
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

	_ = len(a) >= 0 // want `\Qlen(a) >= 0 is always true`
	_ = len(a) < 0  // want `\Qlen(a) < 0 is always false`
	_ = len(a) <= 0 // want `\Qlen(a) <= 0 is never negative, can rewrite as len(a)==0`
}

func newDeref() {
	_ = *new(bool)   // want `\Qreplace *new(bool) with false`
	_ = *new(string) // want `\Qreplace *new(string) with ""`
	_ = *new(int)    // want `\Qreplace *new(int) with 0`
}

func emptyStringTest(s string) {
	sptr := &s

	_ = len(s) == 0 // want `\Qreplace len(s) == 0 with s == ""`
	_ = len(s) != 0 // want `\Qreplace len(s) != 0 with s != ""`

	_ = len(*sptr) == 0 // want `\Qreplace len(*sptr) == 0 with *sptr == ""`
	_ = len(*sptr) != 0 // want `\Qreplace len(*sptr) != 0 with *sptr != ""`

	_ = s == ""
	_ = s != ""

	_ = *sptr == ""
	_ = *sptr != ""

	var b []byte
	_ = len(b) == 0
	_ = len(b) != 0
}

func offBy1(xs []int, ys []string) {
	_ = xs[len(xs)] // want `\Qindex expr always panics; maybe you wanted xs[len(xs)-1]?`
	_ = ys[len(ys)] // want `\Qindex expr always panics; maybe you wanted ys[len(ys)-1]?`

	_ = xs[len(xs)-1]
	_ = ys[len(ys)-1]

	// Conservative with function call.
	// Might return different lengths for both calls.
	_ = makeSlice()[len(makeSlice())]

	var m map[int]int
	// Not an error. Doesn't panic.
	_ = m[len(m)]

	var s string
	var b []byte
	{
		start := strings.Index(s, "/")
		_ = s[start:] // want `\QIndex() can return -1; maybe you wanted to do s[start+1:]`
	}
	{
		start := bytes.Index(b, []byte("/"))
		_ = b[start:] // want `\QIndex() can return -1; maybe you wanted to do b[start+1:]`
	}
	{
		_ = s[strings.Index(s, "/"):] // want `\QIndex() can return -1; maybe you wanted to do Index()+1`
	}
	{
		_ = b[bytes.Index(b, []byte("/")):] // want `\QIndex() can return -1; maybe you wanted to do Index()+1`
	}
	{
		start := strings.Index(s, "/")
		res := s[start:] // want `\QIndex() can return -1; maybe you wanted to do s[start+1:]`
		sink(res)
	}
	{
		start := bytes.Index(b, []byte("/"))
		res := b[start:] // want `\QIndex() can return -1; maybe you wanted to do b[start+1:]`
		sink(res)
	}

	{
		start := strings.Index(s, "/") + 1
		_ = s[start:]
	}
	{
		start := bytes.Index(b, []byte("/")) + 1
		_ = b[start:]
	}
	{
		start := strings.Index(s, "/")
		_ = s[start+1:]
	}
	{
		start := bytes.Index(b, []byte("/"))
		_ = b[start+1:]
	}
	{
		_ = s[strings.Index(s, "/")+1:]
	}
	{
		_ = b[bytes.Index(b, []byte("/"))+1:]
	}
	{
		start := strings.Index(s, "/") + 1
		res := s[start:]
		sink(res)
	}
	{
		start := bytes.Index(b, []byte("/")) + 1
		res := b[start:]
		sink(res)
	}
	{
		start := strings.Index(s, "/")
		res := s[start+1:]
		sink(res)
	}
	{
		start := bytes.Index(b, []byte("/"))
		res := b[start+1:]
		sink(res)
	}
}

func wrapperFunc(s string) {
	_ = strings.SplitN(s, ".", -1)       // want `\Quse Split`
	_ = strings.Replace(s, "a", "b", -1) // want `\Quse Replace`

	_ = strings.Split(s, ".")
	_ = strings.ReplaceAll(s, "a", "b")
}

func flagDeref() {
	_ = *flag.Bool("b", false, "")  // want `\Qimmediate deref in *flag.Bool("b", false, "") is most likely an error`
	_ = *flag.Duration("d", 0, "")  // want `\Qimmediate deref in *flag.Duration("d", 0, "") is most likely an error`
	_ = *flag.Float64("f64", 0, "") // want `\Qimmediate deref in *flag.Float64("f64", 0, "") is most likely an error`
	_ = *flag.Int("i", 0, "")       // want `\Qimmediate deref in *flag.Int("i", 0, "") is most likely an error`
	_ = *flag.Int64("i64", 0, "")   // want `\Qimmediate deref in *flag.Int64("i64", 0, "") is most likely an error`
	_ = *flag.String("s", "", "")   // want `\Qimmediate deref in *flag.String("s", "", "") is most likely an error`
	_ = *flag.Uint("u", 0, "")      // want `\Qimmediate deref in *flag.Uint("u", 0, "") is most likely an error`
	_ = *flag.Uint64("u64", 0, "")  // want `\Qimmediate deref in *flag.Uint64("u64", 0, "") is most likely an error`
}

type object struct {
	data *byte
}

func suspiciousReturns() {
	_ = func(err error) error {
		if err == nil { // want `\Qreturned expr is always nil; replace err with nil`
			return err
		}
		return nil
	}

	_ = func(o *object) *object {
		if o == nil { // want `\Qreturned expr is always nil; replace o with nil`
			return o
		}
		return &object{}
	}

	_ = func(o *object) *byte {
		if o.data == nil { // want `\Qreturned expr is always nil; replace o.data with nil`
			return o.data
		}
		return nil
	}

	_ = func(pointers [][][]map[string]*int) *int {
		if pointers[0][1][2]["ptr"] == nil { // want `\Qreturned expr is always nil; replace pointers[0][1][2]["ptr"] with nil`
			return pointers[0][1][2]["ptr"]
		}
		if ptr := pointers[0][1][2]["ptr"]; ptr == nil { // want `\Qreturned expr is always nil; replace ptr with nil`
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
		for _, x := range xs { // want `\Qxs copy can be avoided with &xs`
			_ = x
		}
	}

	{
		var foo struct {
			arr [768]byte
		}
		for _, x := range foo.arr { // want `\Qfoo.arr copy can be avoided with &foo.arr`
			_ = x
		}
	}

	{
		xsList := make([][512]byte, 1)
		for _, x := range xsList[0] { // want `\QxsList[0] copy can be avoided with &xsList[0]`
			_ = x
		}
	}
}

func badCond(x, y int) {
	if x < -10 && x > 10 { // want `\Qthe condition is always false because -10 <= 10`
	}

	if x > 10 && x < -10 { // want `\Qthe condition is always false because 10 >= -10`
	}

	const ten = 10
	if x > ten+1 && x < -10 { // want `\Qthe condition is always false because ten+1 >= -10`
	}

	// Don't know what value `y` have.
	if x < y && x > 10 {
	}

	if x < -10 && y > 10 {
	}
}

func _(w io.Writer) {
	w.Write([]byte(fmt.Sprintf("%x", 10)))    // want `\Qfmt.Fprintf(w, "%x", 10) should be preferred`
	w.Write([]byte(fmt.Sprint(1, 2, 3, 4)))   // want `\Qfmt.Fprint(w, 1, 2, 3, 4) should be preferred`
	w.Write([]byte(fmt.Sprintln(1, 2, 3, 4))) // want `\Qfmt.Fprintln(w, 1, 2, 3, 4) should be preferred`
}

type exampleStruct struct{}

func (exampleStruct) Sprintf() string {
	return "abc"
}

func _(w io.Writer) {
	var fmt exampleStruct
	w.Write([]byte(fmt.Sprintf()))
}

func sink(args ...interface{}) {}

func _(cond bool, m, m2 *sync.Map) {
	{
		actual, ok := m.Load("key") // want `\Quse m.LoadAndDelete to perform load+delete operations atomically`
		if ok {
			m.Delete("key")
			sink(actual)
		}
	}

	{
		// Condition mismatched.
		v, ok := m.Load("key")
		if ok && cond {
			m.Delete("key")
			sink(v)
		}
	}

	{
		// Maps mismatched.
		actual, ok := m.Load("key")
		if ok {
			m2.Delete("key")
			sink(actual)
		}
	}

	{
		// Keys mismatched.
		actual, ok := m.Load("key")
		if ok {
			m.Delete("key2")
			sink(actual)
		}
	}

	{
		// Return values are ignored.
		m.Load("key")
		if cond {
			m.Delete("key2")
		}
	}

	{
		v, deleted := m.LoadAndDelete("key")
		if deleted {
			sink(v)
		}
	}
}

func argOrder() {
	var s string
	var b []byte

	_ = strings.HasPrefix("http://", s) // want `\Q"http://" and s arguments order looks reversed`

	_ = bytes.HasPrefix([]byte("http://"), b)                         // want `\Q[]byte("http://") and b arguments order looks reversed`
	_ = bytes.HasPrefix([]byte{'h', 't', 't', 'p', ':', '/', '/'}, b) // want `\Q[]byte{'h', 't', 't', 'p', ':', '/', '/'} and b arguments order looks reversed`

	_ = strings.Contains(":", s)       // want `\Q":" and s arguments order looks reversed`
	_ = bytes.Contains([]byte(":"), b) // want `\Q[]byte(":") and b arguments order looks reversed`

	_ = strings.TrimPrefix(":", s)       // want `\Q":" and s arguments order looks reversed`
	_ = bytes.TrimPrefix([]byte(":"), b) // want `\Q[]byte(":") and b arguments order looks reversed`

	_ = strings.TrimSuffix(":", s)       // want `\Q":" and s arguments order looks reversed`
	_ = bytes.TrimSuffix([]byte(":"), b) // want `\Q[]byte(":") and b arguments order looks reversed`

	_ = strings.Split("/", s)       // want `\Q"/" and s arguments order looks reversed`
	_ = bytes.Split([]byte("/"), b) // want `\Q[]byte("/") and b arguments order looks reversed`

	_ = strings.Contains("uint uint8 uint16 uint32", s) // want `\Q"uint uint8 uint16 uint32" and s arguments order looks reversed`
	_ = strings.TrimPrefix("optional foo bar", s)       // want `\Q"optional foo bar" and s arguments order looks reversed`
}

func argOrderNonConstArgs(s1, s2 string, b1, b2 []byte) {
	_ = strings.HasPrefix(s1, s2)
	_ = bytes.HasPrefix(b1, b2)

	x := byte('x')
	_ = bytes.HasPrefix([]byte{x}, b1)
	_ = bytes.HasPrefix([]byte(s1), b1)
}

func argOrderConstOnlyArgs() {
	_ = strings.HasPrefix("", "http://")
	_ = bytes.HasPrefix([]byte{}, []byte("http://"))
	_ = bytes.HasPrefix([]byte{}, []byte{'h', 't', 't', 'p', ':', '/', '/'})
	_ = strings.Contains("", ":")
	_ = bytes.Contains([]byte{}, []byte(":"))
	_ = strings.TrimPrefix("", ":")
	_ = bytes.TrimPrefix([]byte{}, []byte(":"))
	_ = strings.TrimSuffix("", ":")
	_ = bytes.TrimSuffix([]byte{}, []byte(":"))
	_ = strings.Split("", "/")
	_ = bytes.Split([]byte{}, []byte("/"))
}

func argOderProperArgsOrder(s string, b []byte) {
	_ = strings.HasPrefix(s, "http://")
	_ = bytes.HasPrefix(b, []byte("http://"))
	_ = bytes.HasPrefix(b, []byte{'h', 't', 't', 'p', ':', '/', '/'})
	_ = strings.Contains(s, ":")
	_ = bytes.Contains(b, []byte(":"))
	_ = strings.TrimPrefix(s, ":")
	_ = bytes.TrimPrefix(b, []byte(":"))
	_ = strings.TrimSuffix(s, ":")
	_ = bytes.TrimSuffix(b, []byte(":"))
	_ = strings.Split(s, "/")
	_ = bytes.Split(b, []byte("/"))

	const configFileName = "foo.json"
	_ = strings.TrimSuffix(configFileName, filepath.Ext(configFileName))
}

func equalFold(s1, s2 string, b1, b2 []byte) {
	_ = strings.ToLower(s1) == s2                  // want `\Qconsider replacing with strings.EqualFold(s1, s2)`
	_ = s1 == strings.ToLower(s2)                  // want `\Qconsider replacing with strings.EqualFold(s1, s2)`
	_ = strings.ToLower(s1) == strings.ToLower(s2) // want `\Qconsider replacing with strings.EqualFold(s1, s2)`
	_ = strings.ToLower(s1) == "select"            // want `\Qconsider replacing with strings.EqualFold(s1, "select")`
	_ = "select" == strings.ToLower(s1)            // want `\Qconsider replacing with strings.EqualFold("select", s1)`
	_ = strings.ToUpper(s1) == s2                  // want `\Qconsider replacing with strings.EqualFold(s1, s2)`
	_ = s1 == strings.ToUpper(s2)                  // want `\Qconsider replacing with strings.EqualFold(s1, s2)`
	_ = strings.ToUpper(s1) == strings.ToUpper(s2) // want `\Qconsider replacing with strings.EqualFold(s1, s2)`

	_ = strings.ToLower(s1) != s2                  // want `\Qconsider replacing with !strings.EqualFold(s1, s2)`
	_ = s1 != strings.ToLower(s2)                  // want `\Qconsider replacing with !strings.EqualFold(s1, s2)`
	_ = strings.ToLower(s1) != strings.ToLower(s2) // want `\Qconsider replacing with !strings.EqualFold(s1, s2)`
	_ = strings.ToLower(s1) != "select"            // want `\Qconsider replacing with !strings.EqualFold(s1, "select")`
	_ = "select" != strings.ToLower(s1)            // want `\Qconsider replacing with !strings.EqualFold("select", s1)`
	_ = strings.ToUpper(s1) != s2                  // want `\Qconsider replacing with !strings.EqualFold(s1, s2)`
	_ = s1 != strings.ToUpper(s2)                  // want `\Qconsider replacing with !strings.EqualFold(s1, s2)`
	_ = strings.ToUpper(s1) != strings.ToUpper(s2) // want `\Qconsider replacing with !strings.EqualFold(s1, s2)`

	_ = bytes.Equal(bytes.ToLower(b1), b2)                // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = bytes.Equal(b1, bytes.ToLower(b2))                // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = bytes.Equal(bytes.ToLower(b1), bytes.ToLower(b2)) // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = bytes.Equal(bytes.ToLower(b1), []byte("select"))  // want `\Qconsider replacing with bytes.EqualFold(b1, []byte("select"))`
	_ = bytes.Equal([]byte("select"), bytes.ToLower(b1))  // want `\Qconsider replacing with bytes.EqualFold([]byte("select"), b1)`
	_ = bytes.Equal(bytes.ToUpper(b1), b2)                // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = bytes.Equal(b1, bytes.ToUpper(b2))                // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = bytes.Equal(bytes.ToUpper(b1), bytes.ToUpper(b2)) // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`

	_ = !bytes.Equal(bytes.ToLower(b1), b2)                // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = !bytes.Equal(b1, bytes.ToLower(b2))                // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = !bytes.Equal(bytes.ToLower(b1), bytes.ToLower(b2)) // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = !bytes.Equal(bytes.ToLower(b1), []byte("select"))  // want `\Qconsider replacing with bytes.EqualFold(b1, []byte("select"))`
	_ = !bytes.Equal([]byte("select"), bytes.ToLower(b1))  // want `\Qconsider replacing with bytes.EqualFold([]byte("select"), b1)`
	_ = !bytes.Equal(bytes.ToUpper(b1), b2)                // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = !bytes.Equal(b1, bytes.ToUpper(b2))                // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`
	_ = !bytes.Equal(bytes.ToUpper(b1), bytes.ToUpper(b2)) // want `\Qconsider replacing with bytes.EqualFold(b1, b2)`

	_ = strings.ToLower(s1) == strings.ToLower(strings.ToUpper(s2))
	_ = strings.ToUpper(s1) == strings.ToLower(s2)
	_ = strings.ToUpper(s1) == strings.ToUpper(s1)
	_ = strings.ToLower(s1) != strings.ToLower(strings.ToUpper(s2))
	_ = strings.ToUpper(s1) != strings.ToLower(s2)
	_ = strings.ToUpper(s1) != strings.ToUpper(s1)

	_ = bytes.Equal(bytes.ToLower(b1), bytes.ToLower(bytes.ToUpper(b2)))
	_ = bytes.Equal(bytes.ToUpper(b1), bytes.ToLower(b2))
	_ = bytes.Equal(bytes.ToUpper(b1), bytes.ToUpper(b1))
	_ = !bytes.Equal(bytes.ToLower(b1), bytes.ToLower(bytes.ToUpper(b2)))
	_ = !bytes.Equal(bytes.ToUpper(b1), bytes.ToLower(b2))
	_ = !bytes.Equal(bytes.ToUpper(b1), bytes.ToUpper(b1))
}
