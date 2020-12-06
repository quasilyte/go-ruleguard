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

func usefulLenCheck(xs, ys []int, op func()) {
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

	var v int
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
