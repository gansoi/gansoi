package database

import (
	"errors"
)

type (
	// Database defines the interface that database types must implement.
	Database interface {
		// Save will save an object to the database.
		Save(data interface{}) error

		// One will retrieve one record from the database.
		One(fieldName string, value interface{}, to interface{}) error

		// All lists all kinds of a type.
		All(to interface{}, limit int, skip int, reverse bool) error

		// Delete an object from the database.
		Delete(data interface{}) error

		// RegisterListener adds a new listener for changes.
		RegisterListener(listener Listener)
	}
)

var (
	// ErrNotFound is returned when the specified record is not saved.
	ErrNotFound = errors.New("not found")
)
