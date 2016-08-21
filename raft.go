package main

import (
	"net/http"
	"os"
	"time"

	"github.com/abrander/gansoi/database"
	"github.com/abrander/gansoi/node"
)

func main() {
	self := os.Args[1] + ":9118"
	var peers = []string{"127.0.0.1:9118", "127.0.0.2:9118", "127.0.0.3:9118"}

	peerstore := node.NewPeerStore()
	peerstore.SetSelf(self)
	peerstore.SetPeers(peers)

	db, err := database.NewDatabase("/tmp/" + self)
	if err != nil {
		panic(err.Error())
	}

	node, err := node.NewNode(db, peerstore)
	if err != nil {
		panic(err.Error())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/raft", node.ServeHTTP)
	mux.HandleFunc("/raft/stats", node.Stats)
	mux.HandleFunc("/raft/apply", node.Apply)
	mux.HandleFunc("/raft/nodes", node.Nodes)

	s := &http.Server{
		Addr:           self,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err = s.ListenAndServe()
	if err != nil {
		panic(err.Error())
	}
}
