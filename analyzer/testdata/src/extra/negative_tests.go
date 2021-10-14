package extra

import "sync"

func deferredUnlock(mu *sync.Mutex, op func()) {
	mu.Lock()
	defer mu.Unlock()
	op()
}

func goodUnlock(mu *sync.RWMutex, op func()) {
	mu.Lock()
	defer mu.Unlock()
	op()
}

func goodRUnlock(mu *sync.RWMutex, op func()) {
	mu.RLock()
	defer mu.RUnlock()
	op()
}

func goodStrangeLocks(x1, x2 *sync.RWMutex, op func()) {
	x1.Lock()
	defer x2.Lock()
	op()
}

func goodStrangeRLocks(x1, x2 *sync.RWMutex, op func()) {
	x1.RLock()
	defer x2.RLock()
	op()
}

func differentMutexes(mu1, mu2 *sync.RWMutex, op func()) {
	{
		mu2.RLock()
		mu1.Lock()
		mu2.RUnlock()
		mu1.Unlock()
	}

	{
		mu1.Lock()
		mu2.Lock()
		mu1.Unlock()
		mu2.Unlock()
	}
}

func usefulLenCheck(xs, ys []int, v int, op func()) {
	if len(xs) != 0 {
		for range xs {
			// nothing to do
		}
		op()
	}

	if len(xs) != 0 {
		op()
		for i := range xs {
			println(i)
		}
	}

	if len(xs) == 0 {
		for _, v := range xs {
			println(v)
		}
		op()
	}

	if len(xs) == 0 {
		op()
		for _, v = range xs {
			println(v)
		}
	}

	if len(xs) != 0 {
		for range ys {
			// nothing to do
		}
	}
}

type noMutexEmbed1 struct {
	mu sync.Mutex
}

type noMutexEmbed2 struct {
	mu *sync.Mutex
}

type noRWMutexEmbed1 struct {
	mu sync.RWMutex
}

type noRWMutexEmbed2 struct {
	mu *sync.RWMutex
}

type noMutexEmbed3 struct {
	x  int
	mu sync.Mutex
}

type noMutexEmbed4 struct {
	mu *sync.Mutex
	x  int
	y  int
}

type noRWMutexEmbed3 struct {
	x  int
	y  int
	mu sync.RWMutex
}

type noRWMutexEmbed4 struct {
	mu *sync.RWMutex
	x  int
}
