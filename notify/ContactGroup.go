package notify

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/gansoi/gansoi/database"
)

type (
	// ContactGroup is a group of contacts.
	ContactGroup struct {
		database.Object `storm:"inline"`
		Name            string   `json:"name" validate:"required"`
		Members         []string `json:"members"`
	}
)

// LoadContactGroup will read a contact from db.
func LoadContactGroup(db database.Reader, id string) (*ContactGroup, error) {
	var group ContactGroup

	err := db.One("ID", id, &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

// GetContacts returns the list of contacts in g.
func (g *ContactGroup) GetContacts(db database.Reader) ([]*Contact, error) {
	contacts := make([]*Contact, 0, len(g.Members))

	for _, memberID := range g.Members {
		contact, err := LoadContact(db, memberID)
		if err != nil {
			return nil, err
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// Validate implements database.Validator.
func (g *ContactGroup) Validate(db database.Reader) error {
	v := validator.New()
	err := v.Struct(g)
	if err != nil {
		return err
	}

	for _, m := range g.Members {
		_, err = LoadContact(db, m)
		if err != nil {
			return fmt.Errorf("member %s not found", m)
		}
	}

	return nil
}
