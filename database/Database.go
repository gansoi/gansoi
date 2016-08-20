package database

import (
	"encoding/json"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"
)

type (
	// Database is the lowest level of the gansoi database.
	Database struct {
		db *bolt.DB
	}

	// Command is used to denote which operation should be carried out as a
	// result of a Raft commit.
	Command int

	// LogEntry is an entry in the Raft log (?).
	LogEntry struct {
		Command Command
		Key     string
		Value   []byte
	}
)

const (
	// CommandSet will set a new value in the database.
	CommandSet = iota

	// CommandDelete will delete a key in the database.
	CommandDelete
)

// NewDatabase will instantiate a new database placed in filepath.
func NewDatabase(filepath string) (*Database, error) {
	db, err := bolt.Open(
		path.Join(filepath, "gansoi.db"),
		0600,
		&bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}

// NewLogEntry will return a new LogEntry ready for committing to the Raft log.
func NewLogEntry(cmd Command, key string, value []byte) *LogEntry {
	return &LogEntry{
		Command: cmd,
		Key:     key,
		Value:   value,
	}
}

// Byte is a simple helper, that will marshal the entry to a byte slice.
func (e *LogEntry) Byte() []byte {
	b, _ := json.Marshal(e)

	return b
}

// ProcessLogEntry will process the log entry and apply whatever needs doing.
func (d *Database) ProcessLogEntry(entry *LogEntry) error {
	switch entry.Command {
	case CommandSet:
		tx, err := d.db.Begin(true)
		if err != nil {
			return err
		}

		bucket, err := tx.CreateBucketIfNotExists([]byte("keyvalue"))
		if err != nil {
			tx.Rollback()
			return err
		}

		err = bucket.Put([]byte(entry.Key), entry.Value)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

	case CommandDelete:
		tx, err := d.db.Begin(true)
		if err != nil {
			return err
		}

		bucket, err := tx.CreateBucketIfNotExists([]byte("keyvalue"))
		if err != nil {
			tx.Rollback()
			return err
		}

		err = bucket.Delete([]byte(entry.Key))
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// Get will return a value from the generic key/value store.
func (d *Database) Get(key string) ([]byte, error) {
	var value []byte

	err := d.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("keyvalue"))
		value = bucket.Get([]byte(key))
		return nil
	})

	return value, err
}

// Apply implements raft.FSM.
func (d *Database) Apply(l *raft.Log) interface{} {
	entry := &LogEntry{}
	err := json.Unmarshal(l.Data, entry)
	if err != nil {
		// This should not happen..?
		panic(err.Error())
	}

	return d.ProcessLogEntry(entry)
}

// Snapshot implements raft.FSM.
func (d *Database) Snapshot() (raft.FSMSnapshot, error) {
	fmt.Printf("Snapshot()\n")

	return &Snapshot{}, nil
}

// Restore implements raft.FSM.
func (d *Database) Restore(io.ReadCloser) error {
	fmt.Printf("Restore()\n")
	return nil
}
