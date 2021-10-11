// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func ifacePtr(m dsl.Matcher) {
	m.Match(`*$x`).
		Where(m["x"].Type.Underlying().Is(`interface{ $*_ }`)).
		Report(`don't use pointers to an interface`)
}

func newMutex(m dsl.Matcher) {
	m.Match(`$mu := new(sync.Mutex); $mu.Lock()`).
		Report(`use zero mutex value instead, 'var $mu sync.Mutex'`).
		At(m["mu"])
}

func deferCleanup(m dsl.Matcher) {
	m.Match(`$mu.Lock(); $next`).
		Where(!m["next"].Text.Matches(`defer .*\.Unlock\(\)`)).
		Report(`$mu.Lock() should be followed by a deferred Unlock`)

	m.Match(`$res, $err := $open($*_); if $*_ { $*_ }; $next`).
		Where(m["res"].Type.Implements(`io.Closer`) &&
			m["err"].Type.Implements(`error`) &&
			!m["next"].Text.Matches(`defer .*[cC]lose`)).
		Report(`$res.Close() should be deferred right after the $open error check`)
}

func channelSize(m dsl.Matcher) {
	m.Match(`make(chan $_, $size)`).
		Where(m["size"].Value.Int() != 0 && m["size"].Value.Int() != 1).
		Report(`channels should have a size of one or be unbuffered`)
}

func enumStartsAtOne(m dsl.Matcher) {
	m.Match(`const ($_ $_ = iota; $*_)`).
		Report(`enums should start from 1, not 0; use iota+1`)
}

func uncheckedTypeAssert(m dsl.Matcher) {
	m.Match(
		`$_ := $_.($_)`,
		`$_ = $_.($_)`,
		`$_($*_, $_.($_), $*_)`,
		`$_{$*_, $_.($_), $*_}`,
		`$_{$*_, $_: $_.($_), $*_}`).
		Report(`avoid unchecked type assertions as they can panic`)
}

func unnecessaryElse(m dsl.Matcher) {
	m.Match(`var $v $_; if $cond { $v = $x } else { $v = $y }`).
		Where(m["y"].Pure).
		Report(`rewrite as '$v := $y; if $cond { $v = $x }'`)
}

func localVarDecl(m dsl.Matcher) {
	m.Match(`var $x = $y`).
		Where(!m["$$"].Node.Parent().Is(`File`)).
		Suggest(`$x := $y`).
		Report(`use := for local variables declaration`)

	m.Match(`var $x $_ = $y`).
		Where(!m["$$"].Node.Parent().Is(`File`)).
		Report(`use := for local variables declaration`)
}
