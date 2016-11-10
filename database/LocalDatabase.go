package database

type (
	// LocalDatabase must be implemented by types implementing a *local*
	// database.
	LocalDatabase interface {
		Database

		RegisterLocalListener(listener LocalListener)
	}
)
