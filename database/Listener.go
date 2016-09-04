package database

type (
	// Listener is an interface for type capable of listening to changes
	// in the cluster database.
	Listener interface {
		// PostLocalApply will be called in its own goroutine when a change have
		// been applied to the local database node.
		PostLocalApply(command Command, data interface{}, err error)
	}
)
