package extra

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type canStringer struct{}

func (canStringer) String() string { return "" }

func testRedundantSprint(s canStringer) {
	{
		_ = fmt.Sprint(s) // want `suggestion: s\.String\(\)`
	}
	{
		_ = s.String()
	}
}

func simplifySprintf() {
	var s1 string
	var s2 string
	var err error
	var s fmt.Stringer
	_ = fmt.Sprintf("%s%s", s1, s2) // want `suggestion: s1\+s2`
	_ = fmt.Sprintf("%s%s", s1, err)
	_ = fmt.Sprintf("%s%s", s1, s)
}

func testFormatInt() {
	{
		x16 := int16(342)
		_ = fmt.Sprintf("%d", x16) // want `use strconv.FormatInt\(int64\(x16\), 10\)`
	}
	{
		x64 := int64(32)
		_ = fmt.Sprintf("%d", x64) // want `use strconv.FormatInt\(x64, 10\)`
	}
	{
		// Check that convertibleTo(int64) condition works and rejects this.
		s := struct{}{}
		_ = fmt.Sprintf("%d", s)
	}
}

func testFormatBool() {
	{
		i := int64(4)
		_ = fmt.Sprintf("%t", (i+i)&1 == 0) // want `use strconv.FormatBool\(\(i\+i\)&1 == 0\)`
	}
}

func testBlankAssign() {
	x := foo()
	_ = x // want `please remove the assignment to _`

	// This is OK, could be for side-effects.
	_ = foo()
}

func nilErrCheck() {
	if mightFail() == nil { // want `assign mightFail\(\) to err and then do a nil check`
	}
	if mightFail() != nil { // want `assign mightFail\(\) to err and then do a nil check`
	}

	// Good.
	if err := mightFail(); err != nil {
	}
	err := mightFail()
	if err == nil {
	}

	// Not error-typed LHS.
	if newInt() == nil {
	}
}

func unparen(x, y int) {
	if (x == 0) || (y == 0) { // want `rewrite as 'x == 0 || y == 0'`
	}

	if (x != 5) && (y == 5) { // want `rewrite as 'x != 5 && y == 5'`
	}
}

func contextTodo() {
	_ = context.TODO() // want `might want to replace context.TODO\(\)`
	_ = context.Background()
}

func filtepathJoin(bad, good []bool) []byte {
	if bad[0] {
		data, _ := ioutil.ReadFile(path.Join("a", "b")) // want `use filepath\.Join for file paths`
		return data
	}

	if bad[1] {
		p := path.Join("a", "b") // want `use filepath\.Join for file paths`
		data, _ := ioutil.ReadFile(p)
		return data
	}
	if bad[2] {
		f, _ := os.Open(path.Join("123")) // want `use filepath\.Join for file paths`
		data, _ := ioutil.ReadAll(f)
		return data
	}
	if bad[3] {
		p := path.Join("x") // want `use filepath\.Join for file paths`
		f, _ := os.Open(p)
		data, _ := ioutil.ReadAll(f)
		return data
	}

	if good[0] {
		data, _ := ioutil.ReadFile(filepath.Join("a", "b"))
		return data
	}

	if good[1] {
		p := filepath.Join("a", "b")
		data, _ := ioutil.ReadFile(p)
		return data
	}

	return nil
}

func makeExpr() {
	_ = new([14]int)[:10] // want `rewrite as 'make\(\[\]int, 10, 14\)'`
	_ = make([]int, 10, 14)
}

func chanRange() int {
	ch := make(chan int)
	for { // want `can use for range over ch`
		select {
		case c := <-ch:
			return c
		}
	}
}

func unconvertTime() {
	sink = time.Duration(4) * time.Second // want `rewrite as '4 \* time\.Second'`
	sink = 4 * time.Second
}

func timeCast() {
	var t time.Time
	sink = int64(time.Since(t) / time.Microsecond) // want `suggestion: time\.Since\(t\)\.Microseconds\(\)`
	sink = time.Since(t).Microseconds()

	sink = int64(time.Since(t) / time.Millisecond) // want `suggestion: time\.Since\(t\)\.Milliseconds\(\)`
	sink = time.Since(t).Milliseconds()
}

func argOrder() {
	var s1, s2 string

	_ = strings.HasPrefix("prefix", s2) // want `suggestion: strings\.HasPrefix\(s2, "prefix"\)`
	_ = strings.HasSuffix("suffix", s1) // want `suggestion: strings\.HasPrefix\(s1, "suffix"\)`
	_ = strings.Contains("s", s1)       // want `suggestion: strings.Contains\(s1, "s"\)`

	_ = strings.HasPrefix("prefix", "")
	_ = strings.HasSuffix("suffix", "")
	_ = strings.Contains("", "")
}

func stringsReplace() {
	var s string
	_ = strings.Replace(s, " ", " ", -1) // want `replace 'old' and 'new' parameters are identical`
}

func stringsRepeat() {
	var l int
	var part string
	{
		s := make([]string, l) // want `suggestion: strings\.Repeat\("foo", i\)`
		for i := range s {
			s[i] = "foo"
		}
		println(s)
	}
	{
		s := make([]string, 10) // want `suggestion: strings\.Repeat\(part, i\)`
		for i := 0; i < len(s); i++ {
			s[i] = part
		}
		println(s)
	}
}

func stringsCompare() {
	var s1, s2 string

	_ = strings.Compare(s1, s2) == 0  // want `suggestion: s1 == s2`
	_ = strings.Compare(s1, s2) < 0   // want `suggestion: s1 < s2`
	_ = strings.Compare(s1, s2) == -1 // want `suggestion: s1 < s2`
	_ = strings.Compare(s1, s2) > 0   // want `suggestion: s1 > s2`
	_ = strings.Compare(s1, s2) == 1  // want `suggestion: s1 > s2`

	if s1 == s2 {
	}
	if s1 < s2 {
	}
	if s1 > s2 {
	}
}

func hasPrefixSuffix() {
	var s1, s2 string
	if len(s1) >= len(s2) && s1[:len(s2)] == s2 { // want `strings\.HasPrefix\(s1, s2\)`
	}
	if len(s1) >= len(s2) && s1[len(s1)-len(s2):] == s2 { // want `strings\.HasSuffix\(s1, s2\)`
	}
}

func stringsContains() {
	var s1, s2 string

	_ = strings.Count(s1, s2) > 0  // want `suggestion: strings\.Contains\(s1, s2\)`
	_ = strings.Count(s1, s2) >= 1 // want `suggestion: strings\.Contains\(s1, s2\)`
	_ = strings.Count(s1, s2) == 0 // want `suggestion: !strings\.Contains\(s1, s2\)`
}

func fmtFprintf(x int) {
	os.Stderr.WriteString(fmt.Sprintf("foo: %d", x))  // want `suggestion: fmt\.Fprintf\(os\.Stderr, "foo: %d", x\)`
	os.Stderr.WriteString(fmt.Sprintf("message"))     // want `suggestion: fmt\.Fprintf\(os\.Stderr, "message"\)`
	os.Stderr.WriteString(fmt.Sprintf("%d%d", x, 10)) // want `suggestion: fmt\.Fprintf\(os\.Stderr, "%d%d", x, 10\)`
	fmt.Fprintf(os.Stderr, "foo: %d", x)
	fmt.Fprintf(os.Stderr, "message")
	fmt.Fprintf(os.Stderr, "%d%d", x, 10)

	fmt.Fprintf(os.Stdout, "foo: %d", x)  // want `suggestion: fmt\.Printf\("foo: %d", x\)`
	fmt.Fprintf(os.Stdout, "message")     // want `suggestion: fmt\.Printf\("message"\)`
	fmt.Fprintf(os.Stdout, "%d%d", x, 10) // want `suggestion: fmt\.Printf\("%d%d", x, 10\)`
	fmt.Printf("foo: %d", x)
	fmt.Printf("message")
	fmt.Printf("%d%d", x, 10)
}

func sortSlice() {
	var s1, s2 []string
	var ints []int

	sort.Slice(s1, func(i, j int) bool { return s1[i] < s1[j] })       // want `suggestion: sort\.Strings\(s1\)`
	sort.Slice(ints, func(a, b int) bool { return ints[a] < ints[b] }) // want `suggestion: sort\.Ints\(ints\)`

	// No warning: invalid index order.
	sort.Slice(s2, func(a, b int) bool { return s2[b] < s2[a] })

	// No warning: operator differs.
	sort.Slice(s2, func(a, b int) bool { return s2[b] > s2[a] })
	sort.Slice(s2, func(a, b int) bool { return s2[b] >= s2[a] })

	// No warning: not a proper slice type.
	var i32s []int32
	sort.Slice(i32s, func(i, j int) bool { return i32s[i] < i32s[j] })
}

func testCtx(ctx context.Context) error {
	var withCtx struct {
		theContext context.Context
	}

	select { // want `suggestion: if err := withCtx.theContext.Err\(\); err != nil { return err }`
	case <-withCtx.theContext.Done():
		return withCtx.theContext.Err()
	default:
	}

	select { // want `suggestion: if err := ctx.Err\(\); err != nil { return err }`
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

type errDontLog error // want `error as underlying type is probably a mistake`

var ( // want `empty var\(\) block`
// Empty decl...
)

type () // want `empty type\(\) block`

const () // want `empty const\(\) block`

func testEmptyVarBlock() {
	var ()   // want `empty var\(\) block`
	type ()  // want `empty type\(\) block`
	const () // want `empty const\(\) block`
}
