package extra

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

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
		_ = fmt.Sprintf("%t", (i+i)&1 == 0) // want `use strconv.FormatBool\(\(i \+ i\)&1 == 0\)`
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
