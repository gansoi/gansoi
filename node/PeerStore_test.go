package node

import (
	"github.com/hashicorp/raft"
)

var _ raft.PeerStore = (*PeerStore)(nil)
