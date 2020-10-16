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
