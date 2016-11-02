package stats

import (
	"sync/atomic"
)

// GetAll will return all stats.
func GetAll() map[string]interface{} {
	ret := make(map[string]interface{})

	for key, counter := range counters {
		val := atomic.LoadInt64(counter)
		ret[key] = val
	}

	return ret
}
