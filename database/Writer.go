package database

type (
	// Writer defines the interface that database writers types must implement.
	Writer interface {
		// Save will save an object to the database.
		Save(data interface{}) error

		// Delete an object from the database.
		Delete(data interface{}) error
	}
)
