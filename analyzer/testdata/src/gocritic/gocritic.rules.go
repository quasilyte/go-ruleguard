// +build ignore

package gorules

import . "github.com/quasilyte/go-ruleguard/dsl"

func _(m MatchResult) {
	Match(`$x = $x`)
	Warn(`suspicious self-assignment in $$`)

	Match(`$tmp := $x; $x = $y; $y = $tmp`)
	Hint(`can use parallel assignment like $x,$y=$y,$x`)

	Match(
		`io.Copy($x, $x)`,
		`copy($x, $x)`,
	)
	Warn(`suspicious duplicated args in $$`)

	Match(
		`$x && $_ && $x`,
		`$x && $_ && $_ && $x`,
	)
	Error(`suspicious duplicated $x in condition`)

	Match(
		`$x || $x`,
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
		`$x - $x`,
	)
	Filter(m["x"].Pure)
	Error(`suspicious identical LHS and RHS`)

	Match(
		`regexp.Compile($pat)`,
		`regexp.CompilePOSIX($pat)`,
	)
	Filter(m["pat"].Const)
	Hint(`can use MustCompile for const patterns`)

	Match(`map[$_]$_{$*_, $k: $_, $*_, $k: $_, $*_}`)
	Filter(m["k"].Pure)
	Error(`suspicious duplicate key $k`)

	Match(`$dst = append($x, $a); $dst = append($x, $b)`)
	Info(`$dst=append($x,$a,$b) is faster`)

	Match(`strings.Replace($_, $_, $_, 0)`)
	Error(`n=0 argument does nothing, maybe n=-1 is indended?`)

	Match(`append($_)`)
	Error(`append called with 1 argument does nothing`)

	Match(`copy($b, []byte($s))`)
	Filter(m["s"].Type.Is(`string`))
	Hint(`can write copy($b, $s) without type conversion`)

	Match(`$x = $x + 1`)
	Hint(`can simplify to $x++`)
	Match(`$x = $x - 1`)
	Hint(`can simplify to $x--`)

	Match(`$x = $x + $y`)
	Hint(`can simplify to $x+=$y`)
	Match(`$x = $x - $y`)
	Hint(`can simplify to $x-=$y`)
	Match(`$x = $x * $y`)
	Hint(`can simplify to $x*=$y`)

	Match(`!!$x`)
	Hint(`can simplify !!$x to $x`)
	Match(`!($x != $y)`)
	Hint(`can simplify !($x!=$y) to $x==$y`)
	Match(`!($x == $y)`)
	Hint(`can simplify !($x==$y) to $x!=$y`)

	Match(`nil != $_`)
	Warn(`yoda-style expression`)
}
