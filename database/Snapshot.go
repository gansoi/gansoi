package database

import (
	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"
)

type (
	// Snapshot allows raft to retrieve a complete snapshot of the database.
	Snapshot struct {
		db *Database
	}
)

// Persist implements raft.FSMSnapshot.
func (s *Snapshot) Persist(sink raft.SnapshotSink) error {

	err := s.db.Storm().Bolt.View(func(tx *bolt.Tx) error {
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
