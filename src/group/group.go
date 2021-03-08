package group

import "github.com/Kdag-K/kdag/src/peers"

type Group struct {
	ID           string
	Name         string
	AppID        string
	PubKey       string
	LastUpdated  int64
	Peers        []*peers.Peer
	GenesisPeers []*peers.Peer
}

// NewGroup generates a new Group
func NewGroup(id string, name string, appID string, peers []*peers.Peer) *Group {
	return &Group{
		ID:           id,
		Name:         name,
		AppID:        appID,
		Peers:        peers,
		GenesisPeers: peers,
	}
}
