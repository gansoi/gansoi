package notify

import (
	"github.com/abrander/gansoi/logger"
)

type (
	// ContactGroup is a group of contacts.
	ContactGroup struct {
		ID      string   `json:"id"`
		Members []string `json:"members"`
	}
)

// Notify contacts a Contact about a service change.
func (g *ContactGroup) Notify(txt string) error {
	for _, contactID := range g.Members {
		cacheLock.RLock()
		c, found := contactsCache[contactID]
		cacheLock.RUnlock()

		if !found {
			logger.Yellow("notify", "ContactID '%s' not found", contactID)
			continue
		}

		// FIXME: Handle errors somehow.
		go c.Notify(txt)
	}

	return nil
}
