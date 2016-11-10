package boltdb

import (
	"github.com/hashicorp/raft"
)

var _ raft.FSMSnapshot = (*Snapshot)(nil)
