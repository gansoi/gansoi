package database

type (
	// LocalListener is an interface for type capable of listening to changes
	// in the local database.
	LocalListener interface {
		// PostLocalApply will be called in its own goroutine when a change have
		// been applied to the local database.
		PostLocalApply(command Command, data interface{}, err error)
	}
)
