package node

import (
	"sync"
	"testing"
)

func TestNodeLeaderChange(t *testing.T) {
	var lock sync.Mutex
	leader := false
	n := &Node{}
	ch := n.LeaderCh()

	lock.Lock()
	go func() {
		if leader != <-ch {
			t.Fatalf("Wrong value returned in leader channel")
		}
		lock.Unlock()
	}()

	n.leaderChange(leader)

	// Wait for goroutine listening for changes.
	lock.Lock()
}
