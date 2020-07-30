package hashgraph

import "github.com/Kdag-K/kdag/src/peers"

// Store ...
type Store interface {
	CacheSize() int
	GetPeerSet(round int) (*peers.PeerSet, error)
	// SetPeerSet sets the peer-set effective at a given round.
	SetPeerSet(round int, peers *peers.PeerSet) error
	GetAllPeerSets() (map[int][]*peers.Peer, error)
	FirstRound(participantID uint32) (int, bool)
	RepertoireByPubKey() map[string]*peers.Peer
	RepertoireByID() map[uint32]*peers.Peer
	GetEvent(hash string) (*Event, error)
	// SetEvent inserts an envent in the store.
	SetEvent(event *Event) error
	// ParticipantEvents returns all the sorted event hashes of a participant
	// starting at index skip+1.
	ParticipantEvents(participant string, skip int) ([]string, error)
	// ParticipantEvent returns a participant's event with a given index.
	ParticipantEvent(participant string, index int) (string, error)
	// LastEventFrom returns the last event of a participant.
	LastEventFrom(participant string) (string, error)
	LastConsensusEventFrom(string) (string, error)
	KnownEvents() map[uint32]int
	ConsensusEvents() []string
	ConsensusEventsCount() int
	AddConsensusEvent(*Event) error
	GetRound(roundIndex int) (*RoundInfo, error)
	// SetRound stores a round.
	SetRound(roundIndex int, roundInfo *RoundInfo) error
	LastRound() int
	RoundWitnesses(roundIndex int) []string
	// RoundEvents returns the number of events in a round.
	RoundEvents(roundIndex int) int
	// GetRoot returns a participant's root.
	GetRoot(participant string) (*Root, error)
	GetBlock(int) (*Block, error)
	SetBlock(*Block) error
	LastBlockIndex() int
	GetFrame(roundReceived int) (*Frame, error)
	SetFrame(*Frame) error
	Reset(*Frame) error
	Close() error
	StorePath() string
}
