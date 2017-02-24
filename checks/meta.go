package checks

import (
	"math/rand"
	"sync"
	"time"
)

type (
	checkMeta struct {
		running   bool
		runs      int
		NextCheck time.Time
	}
)

var (
	metaLock  sync.RWMutex
	metaStore = make(map[string]*checkMeta)
)

func meta(clock time.Time, check *Check) *checkMeta {
	metaLock.RLock()
	meta, found := metaStore[check.ID]
	metaLock.RUnlock()

	if found {
		return meta
	}

	meta = newMeta(clock, check)

	metaLock.Lock()
	metaStore[check.ID] = meta
	metaLock.Unlock()

	return meta
}

func newMeta(clock time.Time, check *Check) *checkMeta {
	meta := &checkMeta{}

	// Calculate randomized time to execute first check.
	if check.Interval > 0 {
		checkIn := time.Duration(rand.Int63n(int64(check.Interval)))
		meta.NextCheck = clock.Add(checkIn)
	}

	return meta
}
