package database

type (
	// IDSetter is the interface required by all database write operations.
	IDSetter interface {
		SetID()
		GetID() string
	}
)
