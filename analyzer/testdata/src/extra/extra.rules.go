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

	m.Match(`_ = $v`).
		Where(m["v"].Pure).
		Report(`please remove the assignment to _`)

	m.Match(`$err != nil`,
		`$err == nil`).
		Where(!m["err"].Pure && m["err"].Type.Is(`error`)).
		Report(`assign $err to err and then do a nil check`)

	m.Match(`($a) || ($b)`).Report(`rewrite as '$a || $b'`)
	m.Match(`($a) && ($b)`).Report(`rewrite as '$a && $b'`)

	m.Match(`context.TODO()`).Report(`might want to replace context.TODO()`)

	m.Match(`$_, $_ := ioutil.ReadFile(path.Join($*_))`,
		`$p := path.Join($*_); $_, $_ := ioutil.ReadFile($p)`).
		Report(`use filepath.Join for file paths`)

	m.Match(`new([$cap]$typ)[:$len]`).
		Report(`rewrite as 'make([]$typ, $len, $cap)'`)

	// Type check of $ch is not strictly needed, since
	// Go would not permit having non-chan type in the select case clause.
	m.Match(`for { select { case $_ := <-$ch: $*_ } }`).
		Report(`can use for range over $ch`)
}
