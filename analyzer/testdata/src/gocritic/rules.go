// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl/fluent"

func _(m fluent.Matcher) {
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
		Report(`n=0 argument does nothing, maybe n=-1 is indended?`)

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

	// Can factor into a single rule when || operator
	// is supported in filters.
	m.Match(`$s[:]`).
		Where(m["s"].Type.Is(`string`)).
		Report(`can simplify $$ to $s`)
	m.Match(`$s[:]`).
		Where(m["s"].Type.Is(`[]$_`)).
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
		Report(`replace $$ with len($s) == ""`)
	m.Match(`len($s) != 0`).
		Where(m["s"].Type.Is(`string`)).
		Report(`replace $$ with len($s) != ""`)

	m.Match(`$s[len($s)]`).
		Where(m["s"].Type.Is(`[]$elem`) && m["s"].Pure).
		Report(`index expr always panics; maybe you wanted $s[len($s)-1]?`)

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
