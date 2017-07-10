package database

type (
	// ReadWriterBroadcaster must be implemented by types able to both read,
	// write and broadcast.
	ReadWriteBroadcaster interface {
		Reader
		Writer
		Broadcaster
	}
)
