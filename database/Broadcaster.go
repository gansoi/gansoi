package database

type (
	// Broadcaster defines the interface that a type must implement to
	// broadcast changes.
	Broadcaster interface {
		// AddListener adds a new listener for changes.
		RegisterListener(listener Listener)
	}
)
