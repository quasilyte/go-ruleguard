//go:build ignore
// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	m.Match(`$x = $x`).Report(`suspicious self-assignment in $$`)

	m.Match(`$tmp := $x; $x = $y; $y = $tmp`).
		Report(`can use parallel assignment like $x,$y=$y,$x`)

	m.Match(`io.Copy($x, $x)`,
		`copy($x, $x)`).
		Report(`suspicious duplicated args in $$`)

	m.Match(`$x && $_ && $x`,
		`$x && $_ && $_ && $x`).
		Report(`suspicious duplicated $x in condition`)

	m.Match(`$x || $x`,
		`$x && $x`,
		`$x | $x`,
		`$x & $x`,
		`$x ^ $x`,
		`$x < $x`,
		`$x > $x`,
		`$x &^ $x`,
		`$x % $s`,
		`$x == $x`,
		`$x != $x`,
		`$x <= $x`,
		`$x >= $x`,
		`$x / $x`,
		`$x - $x`).
		Where(m["x"].Pure).
		Report(`suspicious identical LHS and RHS`)

	m.Match(`strings.Replace($_, $_, $_, -1)`).Report(`use ReplaceAll`)
	m.Match(`strings.SplitN($_, $_, -1)`).Report(`use Split`)

	m.Match(`regexp.Compile($pat)`,
		`regexp.CompilePOSIX($pat)`).
		Where(m["pat"].Const).
		Report(`can use MustCompile for const patterns`)

	m.Match(`map[$_]$_{$*_, $k: $_, $*_, $k: $_, $*_}`).
		Where(m["k"].Pure).
		Report(`suspicious duplicate key $k`).
		At(m["k"])

	m.Match(`$dst = append($x, $a); $dst = append($x, $b)`).
		Report(`$dst=append($x,$a,$b) is faster`)

	m.Match(`strings.Replace($_, $_, $_, 0)`).
		Report(`n=0 argument does nothing, maybe n=-1 is intended?`)

	m.Match(`append($_)`).
		Report(`append called with 1 argument does nothing`)

	m.Match(`copy($b, []byte($s))`).
		Where(m["s"].Type.Is(`string`)).
		Report(`can write copy($b, $s) without type conversion`)

	m.Match(`$x = $x + 1`).Report(`can simplify to $x++`)
	m.Match(`$x = $x - 1`).Report(`can simplify to $x--`)

	m.Match(`$x = $x + $y`).Report(`can simplify to $x+=$y`)
	m.Match(`$x = $x - $y`).Report(`can simplify to $x-=$y`)
	m.Match(`$x = $x * $y`).Report(`can simplify to $x*=$y`)

	m.Match(`!!$x`).Report(`can simplify !!$x to $x`)
	m.Match(`!($x != $y)`).Report(`can simplify !($x!=$y) to $x==$y`)
	m.Match(`!($x == $y)`).Report(`can simplify !($x==$y) to $x!=$y`)

	m.Match(`nil != $_`).Report(`yoda-style expression`)

	m.Match(`(*$arr)[$_]`).
		Where(m["arr"].Type.Is(`*[$_]$_`)).
		Report(`explicit array deref is redundant`)

	m.Match(`$s[:]`).
		Where(m["s"].Type.Is(`string`) || m["s"].Type.Is(`[]$_`)).
		Report(`can simplify $$ to $s`)

	m.Match(`switch $_ {case $_: $*_}`,
		`switch {case $_: $*_}`,
		`switch $_ := $_.(type) {case $_: $*_}`,
		`switch $_.(type) {case $_: $*_}`).
		Report(`should rewrite switch statement to if statement`)

	m.Match(`switch true {$*_}`).Report(`can omit true in switch`)

	m.Match(`len($_) >= 0`).Report(`$$ is always true`)
	m.Match(`len($_) < 0`).Report(`$$ is always false`)
	m.Match(`len($s) <= 0`).Report(`$$ is never negative, can rewrite as len($s)==0`)

	m.Match(`*new(bool)`).Report(`replace $$ with false`)
	m.Match(`*new(string)`).Report(`replace $$ with ""`)
	m.Match(`*new(int)`).Report(`replace $$ with 0`)

	m.Match(`len($s) == 0`).
		Where(m["s"].Type.Is(`string`)).
		Report(`replace $$ with $s == ""`)
	m.Match(`len($s) != 0`).
		Where(m["s"].Type.Is(`string`)).
		Report(`replace $$ with $s != ""`)

	m.Match(`$s[len($s)]`).
		Where(m["s"].Type.Is(`[]$elem`) && m["s"].Pure).
		Report(`index expr always panics; maybe you wanted $s[len($s)-1]?`)

	m.Match(
		`$i := strings.Index($s, $_); $_ := $slicing[$i:]`,
		`$i := strings.Index($s, $_); $_ = $slicing[$i:]`,
		`$i := bytes.Index($s, $_); $_ := $slicing[$i:]`,
		`$i := bytes.Index($s, $_); $_ = $slicing[$i:]`).
		Where(m["s"].Text == m["slicing"].Text).
		Report(`Index() can return -1; maybe you wanted to do $s[$i+1:]`).
		At(m["slicing"])

	m.Match(
		`$s[strings.Index($s, $_):]`,
		`$s[bytes.Index($s, $_):]`).
		Report(`Index() can return -1; maybe you wanted to do Index()+1`)

	m.Match(`*flag.Bool($*_)`,
		`*flag.Float64($*_)`,
		`*flag.Duration($*_)`,
		`*flag.Int($*_)`,
		`*flag.Int64($*_)`,
		`*flag.String($*_)`,
		`*flag.Uint($*_)`,
		`*flag.Uint64($*_)`).
		Report(`immediate deref in $$ is most likely an error`)

	m.Match(`if $*_; $v == nil { return $v }`).
		Report(`returned expr is always nil; replace $v with nil`)

	const badLoop = `for _, $_ := range $x { $*_ }`
	const goodLoop = `for _, $_ = range $x { $*_ }`
	m.Match(badLoop, goodLoop).
		Where(m["x"].Addressable && m["x"].Type.Size >= 512).
		Report(`$x copy can be avoided with &$x`).
		At(m["x"]).
		Suggest(`&$x`)
}

func badCond(m dsl.Matcher) {
	m.Match(`$x < $a && $x > $b`).
		Where(m["a"].Value.Int() <= m["b"].Value.Int()).
		Report("the condition is always false because $a <= $b")

	m.Match(`$x > $a && $x < $b`).
		Where(m["a"].Value.Int() >= m["b"].Value.Int()).
		Report("the condition is always false because $a >= $b")
}

func appendAssign(m dsl.Matcher) {
	m.Match(`$x = append($y, $_)`).
		Where(m["x"].Text != m["y"].Text &&
			m["x"].Text != "_" &&
			m["x"].Node.Is(`Ident`) &&
			m["y"].Node.Is(`Ident`)).
		Report("append result not assigned to the same slice")
}

//doc:summary Detects fmt.Sprint(f|ln) calls which can be replaced with fmt.Fprint(f|ln)
//doc:tags    performance experimental
//doc:before  w.Write([]byte(fmt.Sprintf("%x", 10)))
//doc:after   fmt.Fprintf(w, "%x", 10)
func preferFprint(m dsl.Matcher) {
	isFmtPackage := func(v dsl.Var) bool {
		return v.Text == "fmt" && v.Object.Is(`PkgName`)
	}

	m.Match(`$w.Write([]byte($fmt.Sprint($*args)))`).
		Where(m["w"].Type.Implements("io.Writer") && isFmtPackage(m["fmt"])).
		Suggest("fmt.Fprint($w, $args)").
		Report(`fmt.Fprint($w, $args) should be preferred to the $$`)

	m.Match(`$w.Write([]byte($fmt.Sprintf($*args)))`).
		Where(m["w"].Type.Implements("io.Writer") && isFmtPackage(m["fmt"])).
		Suggest("fmt.Fprintf($w, $args)").
		Report(`fmt.Fprintf($w, $args) should be preferred to the $$`)

	m.Match(`$w.Write([]byte($fmt.Sprintln($*args)))`).
		Where(m["w"].Type.Implements("io.Writer") && isFmtPackage(m["fmt"])).
		Suggest("fmt.Fprintln($w, $args)").
		Report(`fmt.Fprintln($w, $args) should be preferred to the $$`)
}

func syncMapLoadAndDelete(m dsl.Matcher) {
	m.Match(`$_, $ok := $m.Load($k); if $ok { $m.Delete($k); $*_ }`).
		Where(m.GoVersion().GreaterEqThan("1.15") &&
			m["m"].Type.Is(`*sync.Map`)).
		Report(`use $m.LoadAndDelete to perform load+delete operations atomically`)
}

func argOrder(m dsl.Matcher) {
	m.Match(
		`strings.HasPrefix($lit, $s)`,
		`bytes.HasPrefix($lit, $s)`,
		`strings.HasSuffix($lit, $s)`,
		`bytes.HasSuffix($lit, $s)`,
		`strings.Contains($lit, $s)`,
		`bytes.Contains($lit, $s)`,
		`strings.TrimPrefix($lit, $s)`,
		`bytes.TrimPrefix($lit, $s)`,
		`strings.TrimSuffix($lit, $s)`,
		`bytes.TrimSuffix($lit, $s)`,
		`strings.Split($lit, $s)`,
		`bytes.Split($lit, $s)`).
		Where((m["lit"].Const || m["lit"].ConstSlice) &&
			!(m["s"].Const || m["s"].ConstSlice) &&
			!m["lit"].Node.Is(`Ident`)).
		Report(`$lit and $s arguments order looks reversed`)
}

func equalFold(m dsl.Matcher) {
	// We specify so many patterns to avoid too generic
	// patterns that would match things like
	// `strings.ToLower(x) == strings.ToUpper(y)`
	// While it could be an EqualFold candidate,
	// it just looks wrong and should probably be
	// marked by some other checker.

	// string== patterns
	m.Match(
		`strings.ToLower($x) == $y`,
		`strings.ToLower($x) == strings.ToLower($y)`,
		`$x == strings.ToLower($y)`,
		`strings.ToUpper($x) == $y`,
		`strings.ToUpper($x) == strings.ToUpper($y)`,
		`$x == strings.ToUpper($y)`,
	).
		Where(m["x"].Pure && m["y"].Pure && m["x"].Text != m["y"].Text).
		Suggest(`strings.EqualFold($x, $y)]`).
		Report(`consider replacing with strings.EqualFold($x, $y)`)

	// string!= patterns
	m.Match(
		`strings.ToLower($x) != $y`,
		`strings.ToLower($x) != strings.ToLower($y)`,
		`$x != strings.ToLower($y)`,
		`strings.ToUpper($x) != $y`,
		`strings.ToUpper($x) != strings.ToUpper($y)`,
		`$x != strings.ToUpper($y)`,
	).
		Where(m["x"].Pure && m["y"].Pure && m["x"].Text != m["y"].Text).
		Suggest(`!strings.EqualFold($x, $y)]`).
		Report(`consider replacing with !strings.EqualFold($x, $y)`)

	// bytes.Equal patterns
	m.Match(
		`bytes.Equal(bytes.ToLower($x), $y)`,
		`bytes.Equal(bytes.ToLower($x), bytes.ToLower($y))`,
		`bytes.Equal($x, bytes.ToLower($y))`,
		`bytes.Equal(bytes.ToUpper($x), $y)`,
		`bytes.Equal(bytes.ToUpper($x), bytes.ToUpper($y))`,
		`bytes.Equal($x, bytes.ToUpper($y))`,
	).
		Where(m["x"].Pure && m["y"].Pure && m["x"].Text != m["y"].Text).
		Suggest(`bytes.EqualFold($x, $y)]`).
		Report(`consider replacing with bytes.EqualFold($x, $y)`)
}

func deferUnlambda(m dsl.Matcher) {
	m.Match(`defer func() { $f($*args) }()`).
		Where(m["f"].Node.Is(`Ident`) && m["f"].Text != "panic" && m["f"].Text != "recover" && m["args"].Const).
		Report("can rewrite as `defer $f($args)`")

	m.Match(`defer func() { $pkg.$f($*args) }()`).
		Where(m["f"].Node.Is(`Ident`) && m["args"].Const && m["pkg"].Object.Is(`PkgName`)).
		Report("can rewrite as `defer $pkg.$f($args)`")
}
