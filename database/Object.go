package database

import (
	"github.com/gansoi/gansoi/ca"
)

type (
	// Object can be embbed in other types for saving to the database and
	// thereby easily implementing IDSetter.
	Object struct {
		ID string `json:"id"`
	}
)

// SetID will assign a random ID to o if none is set.
func (o *Object) SetID() {
	if o.ID == "" {
		o.ID = ca.RandomString(24)
	}
}

// GetID will return the ID of o or an empty string if none is set.
func (o *Object) GetID() string {
	return o.ID
}
