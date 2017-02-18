package boltdb

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"

	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/stats"
)

type (
	// BoltStore is the lowest level of the gansoi database, it represent the
	// on-disk database. BoltStore implements raft.FSM and database.Database.
	BoltStore struct {
		dbMutex       sync.RWMutex
		db            *storm.DB
		listenersLock sync.RWMutex
		listeners     []database.Listener
	}
)

func init() {
	stats.CounterInit("database_saves")
	stats.CounterInit("database_deletes")
	stats.CounterInit("database_applied")
	stats.CounterInit("database_snapshot")
}

// NewBoltStore will instantiate a new BoltStore. path will be created if it
// doesn't exist.
func NewBoltStore(path string) (*BoltStore, error) {
	d := &BoltStore{}

	err := d.open(path)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// Close will close the database. Accessing the database after this will
// result in a deadlock.
func (d *BoltStore) Close() error {
	d.dbMutex.RLock()
	return d.db.Close()
}

// open will open the underlying file storage.
func (d *BoltStore) open(filepath string) error {
	db, err := storm.Open(
		filepath,
		storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}),
		storm.AutoIncrement(),
	)
	if err != nil {
		return err
	}

	d.db = db

	return nil
}

// Storm will return the underlying Storm database.
func (d *BoltStore) Storm() *storm.DB {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db
}

// ProcessLogEntry will process the log entry and apply whatever needs doing.
func (d *BoltStore) ProcessLogEntry(entry *database.LogEntry) error {
	var err error

	var v interface{}

	switch entry.Command {
	case database.CommandSave:
		v, _ = entry.Payload()
		stats.CounterInc("database_saves", 1)
		err = d.Save(v)
	case database.CommandDelete:
		stats.CounterInc("database_deletes", 1)
		v, _ = entry.Payload()
		err = d.Delete(v)
	default:
		err = fmt.Errorf("not implemented")
	}

	go func(command database.Command, data interface{}, err error) {
		d.listenersLock.RLock()

		for _, listener := range d.listeners {
			go listener.PostApply(false, command, data, err)
		}

		d.listenersLock.RUnlock()
	}(entry.Command, v, err)

	return err
}

// Apply implements raft.FSM.
func (d *BoltStore) Apply(l *raft.Log) interface{} {
	stats.CounterInc("database_applied", 1)
	entry := &database.LogEntry{}
	err := json.Unmarshal(l.Data, entry)
	if err != nil {
		// This should not happen..?
		fmt.Printf("%s: '%s'\n", err.Error(), string(l.Data))
		return nil
	}

	return d.ProcessLogEntry(entry)
}

// Snapshot implements raft.FSM.
func (d *BoltStore) Snapshot() (raft.FSMSnapshot, error) {
	stats.CounterInc("database_snapshot", 1)
	return &Snapshot{db: d}, nil
}

// Restore implements raft.FSM.
func (d *BoltStore) Restore(source io.ReadCloser) error {
	db := d.Storm().Bolt
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

// Save will save an object to the database.
func (d *BoltStore) Save(data interface{}) error {
	idsetter, ok := data.(database.IDSetter)
	if ok {
		idsetter.SetID()
	}

	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db.Save(data)
}

// One will retrieve one (or zero) record from the database.
func (d *BoltStore) One(fieldName string, value interface{}, to interface{}) error {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	err := d.db.One(fieldName, value, to)
	if err == storm.ErrNotFound {
		return database.ErrNotFound
	}

	return err
}

// All lists all kinds of a type.
func (d *BoltStore) All(to interface{}, limit int, skip int, reverse bool) error {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	err := d.db.All(to, func(opts *index.Options) {
		opts.Limit = limit
		opts.Skip = skip
		opts.Reverse = reverse
	})

	if err == storm.ErrNotFound {
		return database.ErrNotFound
	}

	return err
}

// Delete deletes a record from the store.
func (d *BoltStore) Delete(data interface{}) error {
	return d.db.DeleteStruct(data)
}

// RegisterListener implements database.Database.
func (d *BoltStore) RegisterListener(listener database.Listener) {
	d.listenersLock.Lock()
	defer d.listenersLock.Unlock()

	d.listeners = append(d.listeners, listener)
}

// WriteTo implements io.WriterTo. WriteTo will write a consitent snapshot of
// the database to w.
func (d *BoltStore) WriteTo(w io.Writer) (int64, error) {
	var len int64
	err := d.Storm().Bolt.View(func(tx *bolt.Tx) error {
		var err error

		len, err = tx.WriteTo(w)

		return err
	})

	return len, err
}
