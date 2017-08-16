package notify

import (
	"encoding/json"
	"fmt"

	"gopkg.in/go-playground/validator.v9"

	"github.com/gansoi/gansoi/database"
	"github.com/gansoi/gansoi/plugins"
)

type (
	// Contact is a person or service capable of receiving notifications.
	Contact struct {
		database.Object `storm:"inline"`
		Name            string          `json:"name" validate:"required"`
		Notifier        string          `json:"notifier" validate:"required"`
		Arguments       json.RawMessage `json:"arguments"`
	}
)

// LoadContact will read a contact from db.
func LoadContact(db database.Reader, ID string) (*Contact, error) {
	var contact Contact

	err := db.One("ID", ID, &contact)
	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// Notify contacts a Contact about a service change.
func (c *Contact) Notify(text string) error {
	notifier := plugins.GetNotifier(c.Notifier)
	if notifier == nil {
		return fmt.Errorf("Unknown notifier: %s", c.Notifier)
	}

	err := json.Unmarshal(c.Arguments, &notifier)
	if err != nil {
		return err
	}

	return notifier.Notify(text)
}

// Validate implements database.Validator.
func (c *Contact) Validate(db database.Reader) error {
	v := validator.New()
	return v.Struct(c)
}
