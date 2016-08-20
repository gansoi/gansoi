package node

type (
	// PeerStore will store the current list of raft peers.
	PeerStore struct {
		peers []string
		self  string
	}
)

// NewPeerStore will instantiate a new PeerStore.
func NewPeerStore() *PeerStore {
	return &PeerStore{}
}

// Peers implements raft.PeerStore.
func (p *PeerStore) Peers() ([]string, error) {
	return p.peers, nil
}

// SetPeers implements raft.PeerStore.
func (p *PeerStore) SetPeers(peers []string) error {
	p.peers = peers

	return nil
}

// Self will return our own name as set by SetSelf().
func (p *PeerStore) Self() string {
	return p.self
}

// SetSelf will set our own node name.
func (p *PeerStore) SetSelf(self string) {
	p.self = self
}
