package database

type (
	// ClusterDatabase must be implemented by types representing a cluster-wide
	// database.
	ClusterDatabase interface {
		Database

		RegisterClusterListener(listener ClusterListener)
	}
)
