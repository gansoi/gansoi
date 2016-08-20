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
		raft   *raft.Raft
		stream *HTTPStream
	}
)

// NewNode will initialize a new node.
func NewNode(db *database.Database, peerStore *PeerStore) (*Node, error) {
	var err error
	n := &Node{}

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

	return n, nil
}

// ServeHTTP implements the http.Handler interface.
func (n *Node) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.stream.ServeHTTP(w, r)
}
