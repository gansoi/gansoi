package notify

import (
	"encoding/json"

	"github.com/abrander/gansoi/plugins"
)

type (
	// Contact is a person or service capable of receiving notifications.
	Contact struct {
		ID        string          `json:"id"`
		Notifier  string          `json:"notifier"`
		Arguments json.RawMessage `json:"arguments"`
	}
)

// Notify contacts a Contact about a service change.
func (c *Contact) Notify(text string) error {
	notifier := plugins.GetNotifier(c.Notifier)
	err := json.Unmarshal(c.Arguments, &notifier)
	if err != nil {
		return err
	}

	return notifier.Notify(text)
}
