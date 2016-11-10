package database

type (
	// ClusterListener is an interface for type capable of listening to changes
	// in the cluster database.
	ClusterListener interface {
		// PostClusterApply will be called in its own goroutine when the node
		// detects a change in the cluster database. leader will be true if the
		// current node is leader.
		PostClusterApply(leader bool, command Command, data interface{}, err error)
	}
)
