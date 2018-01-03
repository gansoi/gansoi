package boltdb

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"sync"

	"github.com/gansoi/gansoi/database"
)

type (
	// TestStore can be used when testing database functions.
	TestStore struct {
		BoltStore
		FailSave bool
	}
)

// NewTestStore returns a store suited for testing.
func NewTestStore() *TestStore {
	p := path.Join(os.TempDir(), fmt.Sprintf(".gansoi-test-%d.db", rand.Int63()))
	d := BoltStore{
		dbMutex:       new(sync.RWMutex),
		broadcastFrom: math.MaxUint64,
		listenersLock: new(sync.RWMutex),
	}

	d.open(p)
	os.Remove(p)

	return &TestStore{
		BoltStore: d,
	}
}

// Save will save an object to the database.
func (t *TestStore) Save(data interface{}) error {
	if t.FailSave {
		return errors.New("failed")
	}

	return t.save(data)
}

// Delete an object from the database.
func (t *TestStore) Delete(data interface{}) error {
	return t.delete(data)
}

// RegisterListener implements database.Broadcaster - but does nothing.
func (t *TestStore) RegisterListener(listener database.Listener) {
}
