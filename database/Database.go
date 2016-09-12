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
	// Database is the lowest level of the gansoi database.
	Database struct {
		dbMutex       sync.RWMutex
		db            *storm.DB
		listenersLock sync.RWMutex
		listeners     []Listener
	}
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

	if err != nil {
		panic(err.Error())
	}
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

// One will retrieve one record from the database.
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
