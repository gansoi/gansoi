package cluster

type (
	// PeerAdder must be implemented by types that can add peers.
	PeerAdder interface {
		AddPeer(name string) error
	}
)
