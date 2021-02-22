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