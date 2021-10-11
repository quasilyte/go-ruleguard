package target

import (
	"fmt"
	"os"
	"sync"
)

func sink(args ...interface{}) {}

func ifacePtr() {
	type structType struct {
		_ *fmt.Stringer // want `\Qdon't use pointers to an interface`
	}

	type ifacePtrAlias = *fmt.Stringer // want `\Qdon't use pointers to an interface`

	{
		var x *interface{} // want `\Qdon't use pointers to an interface`
		_ = x
		_ = *x
	}

	{
		var x **interface{} // want `\Qdon't use pointers to an interface`
		_ = x
		_ = *x
	}
}

func newMutex() {
	mu := new(sync.Mutex)
	_ = mu

	mu2 := new(sync.Mutex) // want `\Quse zero mutex value instead, 'var mu2 sync.Mutex'`
	mu2.Lock()
}

func deferCleanup() {
	var mu sync.Mutex

	mu.Lock() // want `\Qmu.Lock() should be followed by a deferred Unlock`
	println("ok")
	mu.Unlock()

	// Still bad: unlock is not deferred.
	mu.Lock() // want `\Qmu.Lock() should be followed by a deferred Unlock`
	mu.Unlock()

	// OK: deferred unlock.
	mu.Lock()
	defer mu.Unlock()

	{
		f, err := os.Open("foo") // want `\Qf.Close() should be deferred right after the os.Open error check`
		if err != nil {
			panic(err)
		}
		sink(f)
	}

	{
		// Still bad: close is not deferred.
		f, err := os.Open("foo") // want `\Qf.Close() should be deferred right after the os.Open error check`
		if err != nil {
			panic(err)
		}
		f.Close()
	}

	{
		// OK: deferred close
		f, err := os.Open("foo")
		if err != nil {
			panic(err)
		}
		defer f.Close()
	}

	{
		// OK: close wrapper is used
		closeFile := func(f *os.File) {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}
		f, err := os.Open("foo")
		if err != nil {
			panic(err)
		}
		defer closeFile(f)
	}
}

func channelSize() {
	_ = make(chan int, 1)    // OK: size of 1
	_ = make(chan string, 0) // OK: explicit size of 0
	_ = make(chan float32)   // OK: unbuffered, implicit size of 0

	size := 1
	_ = make(chan int, size) // OK: can't analyze

	_ = make(chan int, 2)     // want `\Qchannels should have a size of one or be unbuffered`
	_ = make(chan []int, 128) // want `\Qchannels should have a size of one or be unbuffered`
}

func enumStartsAtOne() {
	const ( // want `\Qenums should start from 1, not 0; use iota+1`
		a int = iota
		b
		c
	)

	// OK: untyped, may not be enum
	const (
		ok1 = iota
		ok2
		ok3
	)

	const (
		good int = iota + 1
		good2
		good3
	)

	const (
		alsoGood uint8 = 1 << iota
		alsoGood2
	)
}

func uncheckedTypeAssert() {
	var v interface{}

	_ = v.(int) // want `\Qavoid unchecked type assertions as they can panic`
	{
		x := v.(int) // want `\Qavoid unchecked type assertions as they can panic`
		_ = x
	}

	sink(v.(int))          // want `\Qavoid unchecked type assertions as they can panic`
	sink(0, v.(int))       // want `\Qavoid unchecked type assertions as they can panic`
	sink(v.(int), 0)       // want `\Qavoid unchecked type assertions as they can panic`
	sink(1, 2, v.(int), 3) // want `\Qavoid unchecked type assertions as they can panic`

	{
		type structSink struct {
			f0 interface{}
			f1 interface{}
			f2 interface{}
		}
		_ = structSink{v.(int), 0, 0}    // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{0, v.(string), 0} // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{0, 0, v.([]int)}  // want `\Qavoid unchecked type assertions as they can panic`

		_ = structSink{f0: v.(int)}                  // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{f0: 0, f1: v.(int)}           // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{f0: 0, f1: 0, f2: v.(int)}    // want `\Qavoid unchecked type assertions as they can panic`
		_ = structSink{f0: v.(string), f1: 0, f2: 0} // want `\Qavoid unchecked type assertions as they can panic`
	}

	{
		_ = []interface{}{v.(int)}       // want `\Qavoid unchecked type assertions as they can panic`
		_ = []interface{}{0, v.(int)}    // want `\Qavoid unchecked type assertions as they can panic`
		_ = []interface{}{v.(int), 0}    // want `\Qavoid unchecked type assertions as they can panic`
		_ = []interface{}{0, v.(int), 0} // want `\Qavoid unchecked type assertions as they can panic`

		_ = [...]interface{}{10: v.(int)}               // want `\Qavoid unchecked type assertions as they can panic`
		_ = [...]interface{}{10: 0, 20: v.(int)}        // want `\Qavoid unchecked type assertions as they can panic`
		_ = [...]interface{}{10: v.(int), 20: 0}        // want `\Qavoid unchecked type assertions as they can panic`
		_ = [...]interface{}{10: 0, 20: v.(int), 30: 0} // want `\Qavoid unchecked type assertions as they can panic`
	}
}

func unnecessaryElse() {
	var cond bool

	{
		var x int // want `\Qrewrite as 'x := 5; if cond { x = 10 }'`
		if cond {
			x = 10
		} else {
			x = 5
		}
		_ = x
	}
}

var globalVar = 10
var globalVar2 uint8 = 10

func localVarDecl() {
	var i = 10        // want `\Quse := for local variables declaration`
	var i2 uint8 = 10 // want `\Quse := for local variables declaration`
	sink(i)
	sink(i2)
}
