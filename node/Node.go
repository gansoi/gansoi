package node

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/raft"

	"github.com/abrander/gansoi/database"
)

type (
	// Node represents a single gansoi node.
	Node struct {
		db     *database.Database
		raft   *raft.Raft
		leader bool
		stream *HTTPStream
	}
)

// NewNode will initialize a new node.
func NewNode(db *database.Database, peerStore *PeerStore) (*Node, error) {
	var err error
	n := &Node{
		db: db,
	}

	// Raft config.
	conf := raft.DefaultConfig()
	conf.HeartbeatTimeout = 1000 * time.Millisecond
	conf.ElectionTimeout = 1000 * time.Millisecond
	conf.LeaderLeaseTimeout = 500 * time.Millisecond
	conf.CommitTimeout = 200 * time.Millisecond
	conf.Logger = log.New(os.Stdout, "      RAFT ", log.Lmicroseconds|log.Lshortfile)

	// Set up nice HTTP based transport.
	n.stream, err = NewHTTPStream(peerStore.Self())
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, " TRANSPORT ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	transport := raft.NewNetworkTransportWithLogger(n.stream, 1, 0, logger)

	n.raft, err = raft.NewRaft(
		conf,
		db,
		raft.NewInmemStore(),
		raft.NewInmemStore(),
		raft.NewDiscardSnapshotStore(),
		peerStore,
		transport)
	if err != nil {
		return nil, err
	}

	lch := n.raft.LeaderCh()
	go func() {
		select {
		case leader := <-lch:
			n.leader = leader
		}
	}()

	return n, nil
}

// ServeHTTP implements the http.Handler interface.
func (n *Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.stream.ServeHTTP(w, r)
}

// Set will set a key in the generic Raft-backed key/value store.
func (n *Node) Set(key string, value []byte) error {
	if !n.leader {
		/* FIXME: Relay to leader somehow */
		return raft.ErrNotLeader
	}

	entry := database.NewLogEntry(database.CommandSet, key, value)
	n.raft.Apply(entry.Byte(), time.Minute)

	return nil
}

// Get will retrieve a key from the generic Raft-backed K/V store.
func (n *Node) Get(key string) ([]byte, error) {
	return n.db.Get(key)
}
