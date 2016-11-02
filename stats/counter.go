package stats

import (
	"sync"
	"sync/atomic"
)

var (
	counterMutex sync.RWMutex

	counters = make(map[string]*int64)
)

func counter(key string) *int64 {
	counterMutex.RLock()
	counter, found := counters[key]
	counterMutex.RUnlock()

	if !found {
		panic(key + " not found")
	}

	return counter
}

// CounterInit will add a new counter.
func CounterInit(key string) {
	counterMutex.RLock()
	counter, found := counters[key]
	counterMutex.RUnlock()

	if found {
		panic(key + " already exists")
	}

	var i int64
	counter = &i

	counterMutex.Lock()
	counters[key] = counter
	counterMutex.Unlock()
}

// CounterInc increments a counter by val.
func CounterInc(key string, val int64) {
	atomic.AddInt64(counter(key), val)
}

// CounterDec decrements a counter by val.
func CounterDec(key string, val int64) {
	atomic.AddInt64(counter(key), -val)
}

// CounterSet will set a counter to val.
func CounterSet(key string, val int64) {
	atomic.StoreInt64(counter(key), val)
}

// CounterGet will return the current value of counter.
func CounterGet(key string) int64 {
	return atomic.LoadInt64(counter(key))
}
