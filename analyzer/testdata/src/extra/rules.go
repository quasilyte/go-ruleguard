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

	m.Match(`fmt.Sprint($x)`).
		Where(m["x"].Type.Implements(`fmt.Stringer`)).
		Suggest(`$x.String()`)

	m.Match(`_ = $v`).
		Where(m["v"].Pure).
		Report(`please remove the assignment to _`)

	m.Match(`$err != nil`,
		`$err == nil`).
		Where(!m["err"].Pure && m["err"].Type.Is(`error`)).
		Report(`assign $err to err and then do a nil check`)

	// FIXME: this is not 100% correct.
	// If ($a) contains something that has a higher precedence
	// that ||, the result would not be functionally identical.
	m.Match(`($a) || ($b)`).Report(`rewrite as '$a || $b'`)
	m.Match(`($a) && ($b)`).Report(`rewrite as '$a && $b'`)

	m.Match(`context.TODO()`).Report(`might want to replace context.TODO()`)

	m.Match(`os.Open(path.Join($*_))`,
		`ioutil.ReadFile(path.Join($*_))`,
		`$p := path.Join($*_); $_, $_ := os.Open($p)`,
		`$p := path.Join($*_); $_, $_ := ioutil.ReadFile($p)`).
		Report(`use filepath.Join for file paths`)

	m.Match(`new([$cap]$typ)[:$len]`).
		Report(`rewrite as 'make([]$typ, $len, $cap)'`)

	// Type check of $ch is not strictly needed, since
	// Go would not permit having non-chan type in the select case clause.
	m.Match(`for { select { case $_ := <-$ch: $*_ } }`).
		Report(`can use for range over $ch`)

	m.Match(`time.Duration($x) * time.Second`).
		Where(m["x"].Const).
		Report(`rewrite as '$x * time.Second'`)

	m.Match(`select {case <-$ctx.Done(): return $ctx.Err(); default:}`).
		Where(m["ctx"].Type.Is(`context.Context`)).
		Suggest(`if err := $ctx.Err(); err != nil { return err }`)

	// See https://twitter.com/dvyukov/status/1174698980208513024
	m.Match(`type $x error`).
		Report(`error as underlying type is probably a mistake`).
		Suggest(`type $x struct { error }`)

	m.Match(`var()`).Report(`empty var() block`)
	m.Match(`const()`).Report(`empty const() block`)
	m.Match(`type()`).Report(`empty type() block`)

	m.Match(`int64(time.Since($t) / time.Microsecond)`).
		Suggest(`time.Since($t).Microseconds()`)
	m.Match(`int64(time.Since($t) / time.Millisecond)`).
		Suggest(`time.Since($t).Milliseconds()`)
}
