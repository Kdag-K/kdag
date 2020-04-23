package node

import (
	"sync"
	"sync/atomic"
)

// State captures the state of a Babble node: Babbling, CatchingUp, Joining,
// Leaving, Suspended, or Shutdown
type State uint32

const (
	//Babbling is the initial state of a Babble node.
	Babbling State = iota
	//CatchingUp implements Fast Sync
	CatchingUp
	//Joining is joining
	Joining
	//Leaving is leaving
	Leaving
	//Shutdown is shutdown
	Shutdown
	//Suspended is initialised, but not gossipping
	Suspended
)
