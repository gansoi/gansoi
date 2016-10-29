package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
	"github.com/hashicorp/raft"
)

type (
	// Database is the lowest level of the gansoi database, it represent the
	// on-disk database. Database implements raft.FSM.
	Database struct {
		dbMutex       sync.RWMutex
		db            *storm.DB
		listenersLock sync.RWMutex
		listeners     []Listener
	}
)

var (
	// ErrNotFound is returned when the specified record is not saved.
	ErrNotFound = storm.ErrNotFound
)

const (
	// stableBucket is the bucket name used for raft stable storage.
	stableBucket = "raft.StableStore"

	// logBucket is used for raft log storage.
	logBucket = "raft.LogStore"
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

// open will open the underlying file storage.
func (d *Database) open(filepath string) error {
	db, err := storm.Open(
		filepath,
		storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}),
		storm.AutoIncrement(),
	)
	if err != nil {
		return err
	}

	d.db = db

	d.Storm().Bolt.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(logBucket))

		return nil
	})

	return nil
}

// Storm will return the underlying Storm database.
func (d *Database) Storm() *storm.DB {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db
}

// ProcessLogEntry will process the log entry and apply whatever needs doing.
func (d *Database) ProcessLogEntry(entry *LogEntry) error {
	var err error

	var v interface{}

	switch entry.Command {
	case CommandSave:
		v, _ = entry.Payload()
		err = d.Save(v)
	case CommandDelete:
		v, _ = entry.Payload()
		err = d.db.DeleteStruct(v)
	default:
		err = fmt.Errorf("not implemented")
	}

	go func(command Command, data interface{}, err error) {
		d.listenersLock.RLock()

		for _, listener := range d.listeners {
			listener.PostLocalApply(command, data, err)
		}

		d.listenersLock.RUnlock()
	}(entry.Command, v, err)

	return err
}

// Apply implements raft.FSM.
func (d *Database) Apply(l *raft.Log) interface{} {
	entry := &LogEntry{}
	err := json.Unmarshal(l.Data, entry)
	if err != nil {
		// This should not happen..?
		fmt.Printf("%s: '%s'\n", err.Error(), string(l.Data))
		return nil
	}

	return d.ProcessLogEntry(entry)
}

// Snapshot implements raft.FSM.
func (d *Database) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot{db: d}, nil
}

// Restore implements raft.FSM.
func (d *Database) Restore(source io.ReadCloser) error {
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
func (d *Database) Save(data interface{}) error {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db.Save(data)
}

// One will retrieve one (or zero) record from the database.
func (d *Database) One(fieldName string, value interface{}, to interface{}) error {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db.One(fieldName, value, to)
}

// All lists all kinds of a type.
func (d *Database) All(to interface{}, limit int, skip int, reverse bool) error {
	d.dbMutex.RLock()
	defer d.dbMutex.RUnlock()

	return d.db.All(to, func(opts *index.Options) {
		opts.Limit = limit
		opts.Skip = skip
		opts.Reverse = reverse
	})
}

// RegisterListener will register a listener for new changes to the database.
func (d *Database) RegisterListener(listener Listener) {
	d.listenersLock.Lock()
	defer d.listenersLock.Unlock()

	d.listeners = append(d.listeners, listener)
}

// Set implements raft.StableStore.
func (d *Database) Set(k, v []byte) error {
	return d.Storm().Set(stableBucket, k, v)
}

// Get implements raft.StableStore.
func (d *Database) Get(k []byte) ([]byte, error) {
	var value []byte
	err := d.Storm().Get(stableBucket, k, &value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// SetUint64 is like Set, but handles uint64 values
func (d *Database) SetUint64(key []byte, val uint64) error {
	return d.Set(key, uint64ToBytes(val))
}

// GetUint64 is like Get, but handles uint64 values
func (d *Database) GetUint64(key []byte) (uint64, error) {
	val, err := d.Get(key)
	if err != nil {
		return 0, err
	}

	return bytesToUint64(val), nil
}

// FirstIndex implements raft.LogStore.
func (d *Database) FirstIndex() (uint64, error) {
	tx, err := d.Storm().Bolt.Begin(false)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	curs := tx.Bucket([]byte(logBucket)).Cursor()

	first, _ := curs.First()
	if first == nil {
		return 0, nil
	}

	return bytesToUint64(first), nil
}

// LastIndex implements raft.LogStore.
func (d *Database) LastIndex() (uint64, error) {
	tx, err := d.Storm().Bolt.Begin(false)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	curs := tx.Bucket([]byte(logBucket)).Cursor()

	last, _ := curs.Last()
	if last == nil {
		return 0, nil
	}

	return bytesToUint64(last), nil
}

// GetLog implements raft.LogStore.
func (d *Database) GetLog(idx uint64, log *raft.Log) error {
	tx, err := d.Storm().Bolt.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	value := tx.Bucket([]byte(logBucket)).Get(uint64ToBytes(idx))

	return json.Unmarshal(value, log)
}

// StoreLog implements raft.LogStore.
func (d *Database) StoreLog(log *raft.Log) error {
	return d.StoreLogs([]*raft.Log{log})
}

// StoreLogs implements raft.LogStore.
func (d *Database) StoreLogs(logs []*raft.Log) error {
	tx, err := d.Storm().Bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, log := range logs {
		key := uint64ToBytes(log.Index)
		value, err := json.Marshal(log)
		if err != nil {
			return err
		}

		bucket := tx.Bucket([]byte(logBucket))
		err = bucket.Put(key, value)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteRange implements raft.LogStore.
func (d *Database) DeleteRange(min, max uint64) error {
	minKey := uint64ToBytes(min)

	// We can safely use standar Bolt functions, we're not interested in the
	// encoded values, we're only using delete, so Storm encoding/decoding
	// doesn't matter.
	tx, err := d.Storm().Bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	curs := tx.Bucket([]byte(logBucket)).Cursor()
	for k, _ := curs.Seek(minKey); k != nil; k, _ = curs.Next() {
		if bytesToUint64(k) > max {
			break
		}

		if err := curs.Delete(); err != nil {
			return err
		}
	}

	return tx.Commit()
}
