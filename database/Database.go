package database

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"
)

type (
	// Database is the lowest level of the gansoi database.
	Database struct {
		dbMutex sync.RWMutex
		db      *bolt.DB
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
	d := &Database{}

	err := d.open(path.Join(filepath, "gansoi.db"))
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) open(filepath string) error {
	db, err := bolt.Open(
		filepath,
		0600,
		&bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("keyvalue"))
		return nil
	})
	if err != nil {
		return err
	}

	d.db = db

	return nil
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

// Db will return the underlying Bolt database.
func (d *Database) Db() *bolt.DB {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db
}

// ProcessLogEntry will process the log entry and apply whatever needs doing.
func (d *Database) ProcessLogEntry(entry *LogEntry) error {
	db := d.Db()
	switch entry.Command {
	case CommandSet:
		tx, err := db.Begin(true)
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
		tx, err := db.Begin(true)
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

	err := d.Db().View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("keyvalue"))
		if bucket == nil {
			panic("bucket == nil")
		}

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
	return &Snapshot{db: d}, nil
}

// Restore implements raft.FSM.
func (d *Database) Restore(source io.ReadCloser) error {
	db := d.Db()
	d.dbMutex.Lock()
	defer d.dbMutex.Unlock()
	defer source.Close()

	path := db.Path()
	restorePath := path + ".restoretmp"

	file, err := os.Create(restorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, source)
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil {
		return err
	}

	err = os.Rename(restorePath, path)
	if err != nil {
		return err
	}

	err = d.open(path)
	if err != nil {
		return err
	}

	return nil
}
