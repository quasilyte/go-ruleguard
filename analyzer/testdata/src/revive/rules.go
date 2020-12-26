// +build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func testRules(m dsl.Matcher) {
	m.Match(`runtime.GC()`).Report(`explicit call to GC`)

	m.Match(`$x = atomic.AddInt32(&$x, $_)`,
		`$x = atomic.AddInt64(&$x, $_)`,
		`$x = atomic.AddUint32(&$x, $_)`,
		`$x = atomic.AddUint64(&$x, $_)`,
		`*$x = atomic.AddInt32($x, $_)`,
		`*$x = atomic.AddInt64($x, $_)`,
		`*$x = atomic.AddUint32($x, $_)`,
		`*$x = atomic.AddUint64($x, $_)`).
		Report(`direct assignment to atomic value`)

	m.Match(`$x == true`,
		`$x != true`,
		`$x == false`,
		`$x != false`).
		Report(`omit bool literal in expression`)
}
