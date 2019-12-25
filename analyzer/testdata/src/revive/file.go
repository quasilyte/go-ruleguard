package revive

import (
	"runtime"
	"sync/atomic"
)

func callToGC() {
	runtime.GC() // want `explicit call to GC`
}

func atomicAssign() {
	var i64 int64

	i64 = atomic.AddInt64(&i64, 10) // want `direct assignment to atomic value`
	i64p := &i64
	*i64p = atomic.AddInt64(i64p, 10) // want `direct assignment to atomic value`
}

func boolLiteralInExpr(a, b, c, d int) bool {
	var bar, yes bool

	if bar == true { // want `omit bool literal in expression`
	}

	for getBool() || yes != false { // want `omit bool literal in expression`
	}

	return b > c == false // want `omit bool literal in expression`
}
