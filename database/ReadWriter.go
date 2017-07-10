package database

type (
	// ReadWriter must be implemented by types able to both read and write.
	ReadWriter interface {
		Reader
		Writer
	}
)
