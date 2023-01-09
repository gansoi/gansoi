package database

import (
	"errors"
)

type (
	// Reader defines the interface that database reader types must implement.
	Reader interface {
		// One will retrieve one record from the database.
		One(fieldName string, value interface{}, to interface{}) error

		// All lists all kinds of a type.
		All(to interface{}, limit int, skip int, reverse bool) error

		// Find Find returns one or more records by the specified index.
		Find(field string, value interface{}, to interface{}, limit int, skip int, reverse bool) error
	}
)

var (
	// ErrNotFound is returned when the specified record is not saved.
	ErrNotFound = errors.New("not found")
)
