package node

import (
	"github.com/abrander/gansoi/database"
)

type (
	// Listener is an interface for type capable of listening to changes
	// in the cluster database.
	Listener interface {
		// PostClusterApply will be called in its own goroutine when the node
		// detects a change in the cluster database. leader will be true if the
		// current node is leader.
		PostClusterApply(leader bool, command database.Command, data interface{}, err error)
	}
)
