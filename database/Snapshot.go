package database

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"
)

type (
	Snapshot struct {
		db *Database
	}
)

// Persist implements raft.FSMSnapshot.
func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	fmt.Printf("Persist()\n")

	err := s.db.Db().View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(sink)

		return err
	})

	if err != nil {
		sink.Cancel()
		return err
	}

	sink.Close()

	return nil
}

// Release implements raft.FSMSnapshot.
func (s *Snapshot) Release() {
}
