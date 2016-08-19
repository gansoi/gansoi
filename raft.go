package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

type (
	// PeerStore will store the current list of raft peers.
	PeerStore struct {
		peers []string
	}

	// Fsm FIXME.
	Fsm struct{}

	// FsmSnapshot FIXME.
	FsmSnapshot struct{}
)

func (p *PeerStore) Peers() ([]string, error) {
	return p.peers, nil
}

func (p *PeerStore) SetPeers(peers []string) error {
	p.peers = peers

	return nil
}

func (f Fsm) Apply(l *raft.Log) interface{} {
	return nil
}

func (f Fsm) Snapshot() (raft.FSMSnapshot, error) {
	fmt.Printf("Snapshot()\n")
	return FsmSnapshot{}, nil
}

func (f Fsm) Restore(io.ReadCloser) error {
	fmt.Printf("Restore()\n")
	return nil
}

func (s FsmSnapshot) Persist(sink raft.SnapshotSink) error {
	fmt.Printf("Persist()\n")
	return nil
}

func (s FsmSnapshot) Release() {
	fmt.Printf("Release()\n")
}

func main() {
	self := os.Args[1] + ":9118"

	h, err := NewHTTPStream(self)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/raft", h)

	s := &http.Server{
		Addr:           self,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go s.ListenAndServe()

	logger := log.New(os.Stdout, " TRANSPORT ", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	transport := raft.NewNetworkTransportWithLogger(h, 1, 0, logger)

	conf := raft.DefaultConfig()
	conf.HeartbeatTimeout = 1000 * time.Millisecond
	conf.ElectionTimeout = 1000 * time.Millisecond
	conf.LeaderLeaseTimeout = 500 * time.Millisecond
	conf.CommitTimeout = 200 * time.Millisecond
	conf.Logger = log.New(os.Stdout, "      RAFT ", log.Lmicroseconds|log.Lshortfile)

	logs := raft.NewInmemStore()
	stable := raft.NewInmemStore()
	snaps := raft.NewDiscardSnapshotStore()

	var peers = []string{"127.0.0.1:9118", "127.0.0.2:9118", "127.0.0.3:9118"}
	peerStore := PeerStore{}
	peerStore.SetPeers(peers)

	fsm := Fsm{}

	r, _ := raft.NewRaft(
		conf,
		fsm,
		logs,
		stable,
		snaps,
		&peerStore,
		transport)
	lch := r.LeaderCh()

	mux.HandleFunc("/raft/stats", func(w http.ResponseWriter, req *http.Request) {
		stats := r.Stats()
		e := json.NewEncoder(w)
		e.Encode(stats)
		req.Body.Close()
	})

	c := time.NewTicker(time.Second * 10).C
	for {
		select {
		case leader := <-lch:
			if leader {
				fmt.Printf("I WAS ELECTED LEADER\n")
			}
		case <-c:
			r.Apply([]byte("buh"), time.Second)
		}
	}
}
