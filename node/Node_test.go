package node

import (
	"sync"
	"testing"

	"github.com/gansoi/gansoi/database"
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

// Make sure we implement the needed interfaces.
var _ database.ReadWriter = (*Node)(nil)
var _ database.Broadcaster = (*Node)(nil)
