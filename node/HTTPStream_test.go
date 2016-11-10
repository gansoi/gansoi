package node

import (
	"github.com/hashicorp/raft"
)

var _ raft.StreamLayer = (*HTTPStream)(nil)
