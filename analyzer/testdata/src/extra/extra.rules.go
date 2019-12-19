// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
	// We don't want to suggest int64(x) if x is already int64,
	// this is why 2 rules are needed.
	// Maybe there will be a way to group these 2 together in
	// future, but this solution will do for now.
	m.Match(`fmt.Sprintf("%d", $i)`).
		Where(m["i"].Type.Is(`int64`)).
		Report(`use strconv.FormatInt($i, 10)`)
	m.Match(`fmt.Sprintf("%d", $i)`).
		Where(m["i"].Type.ConvertibleTo(`int64`)).
		Report(`use strconv.FormatInt(int64($i), 10)`)

	m.Match(`fmt.Sprintf("%t", $i&1 == 0)`).
		Report(`use strconv.FormatBool($i&1 == 0)`)
}
