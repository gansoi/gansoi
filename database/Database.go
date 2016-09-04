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
		dbMutex sync.RWMutex
		db      *storm.DB
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

	switch entry.Command {
	case CommandSave:
		v := GetType(entry.Type)
		json.Unmarshal(entry.Value, v)
		err = d.Save(v)
	case CommandDelete:
		v := GetType(entry.Type)
		json.Unmarshal(entry.Value, v)
		err = d.db.DeleteStruct(v)
	default:
		err = fmt.Errorf("not implemented")
	}

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
