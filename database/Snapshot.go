package database

import (
	"fmt"

	"github.com/hashicorp/raft"
)

type (
	Snapshot struct {
	}
)

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	fmt.Printf("Persist()\n")
	return nil
}

func (s *Snapshot) Release() {
	fmt.Printf("Release()\n")
}
