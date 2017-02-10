package notify

import (
	"github.com/gansoi/gansoi/database"
)

type (
	// ContactGroup is a group of contacts.
	ContactGroup struct {
		database.Object `storm:"inline"`
		Name            string   `json:"name"`
		Members         []string `json:"members"`
	}
)

// LoadContactGroup will read a contact from db.
func LoadContactGroup(db database.Database, ID string) (*ContactGroup, error) {
	var group ContactGroup

	err := db.One("ID", ID, &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

// GetContacts returns the list of contacts in g.
func (g *ContactGroup) GetContacts(db database.Database) ([]*Contact, error) {
	var contacts []*Contact
	for _, memberID := range g.Members {
		contact, err := LoadContact(db, memberID)
		if err != nil {
			return nil, err
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}
