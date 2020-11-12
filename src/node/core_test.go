package node

import (
	"crypto/ecdsa"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/Kdag-K/kdag/src/common"
	"github.com/Kdag-K/kdag/src/crypto/keys"
	hg "github.com/Kdag-K/kdag/src/hashgraph"
	"github.com/Kdag-K/kdag/src/peers"
	"github.com/Kdag-K/kdag/src/proxy"
)

func initCores(n int, t *testing.T) ([]*core, map[uint32]*ecdsa.PrivateKey, map[string]string) {
	cacheSize := 1000

	cores := []*core{}
	index := make(map[string]string)
	participantKeys := map[uint32]*ecdsa.PrivateKey{}
	pirs := []*peers.Peer{}

	for i := 0; i < n; i++ {
		key, _ := keys.GenerateECDSAKey()
		peer := peers.NewPeer(keys.PublicKeyHex(&key.PublicKey), "", "")
		pirs = append(pirs, peer)
		participantKeys[peer.ID()] = key
	}

	peerSet := peers.NewPeerSet(pirs)

	genesisPeerSet := clonePeerSet(t, peerSet.Peers)

	for i, peer := range peerSet.Peers {
		key, _ := participantKeys[peer.ID()]

		core := newCore(
			NewValidator(key, peer.Moniker),
			peerSet,
			genesisPeerSet,
			hg.NewInmemStore(cacheSize),
			proxy.DummyCommitCallback,
			false,
			common.NewTestEntry(t, common.TestLogLevel))

		//Create and save the first Event
		initialEvent := hg.NewEvent([][]byte(nil),
			[]hg.InternalTransaction{},
			nil,
			[]string{"", ""},
			core.validator.PublicKeyBytes(),
			0)

		err := core.signAndInsertSelfEvent(initialEvent)
		if err != nil {
			t.Fatal(err)
		}

		cores = append(cores, core)
		index[fmt.Sprintf("e%d", i)] = core.head
	}

	return cores, participantKeys, index
}

/*
|  e12  |
|   | \ |
|   |   e20
|   | / |
|   /   |
| / |   |
e01 |   |
| \ |   |
e0  e1  e2
0   1   2
*/
func initHashgraph(cores []*core, keys map[uint32]*ecdsa.PrivateKey, index map[string]string, participant uint32) {
	for i := 0; i < len(cores); i++ {
		if uint32(i) != participant {
			event, _ := cores[i].getEvent(index[fmt.Sprintf("e%d", i)])
			if err := cores[participant].insertEventAndRunConsensus(event, true); err != nil {
				fmt.Printf("error inserting %s: %s\n", getName(index, event.Hex()), err)
			}
		}
	}

	event01 := hg.NewEvent([][]byte{},
		[]hg.InternalTransaction{},
		nil,
		[]string{index["e0"], index["e1"]}, //e0 and e1
		cores[0].validator.PublicKeyBytes(), 1)
	if err := insertEvent(cores, keys, index, event01, "e01", participant, cores[0].validator.ID()); err != nil {
		fmt.Printf("error inserting e01: %s\n", err)
	}

	event20 := hg.NewEvent([][]byte{},
		[]hg.InternalTransaction{},
		nil,
		[]string{index["e2"], index["e01"]}, //e2 and e01
		cores[2].validator.PublicKeyBytes(), 1)
	if err := insertEvent(cores, keys, index, event20, "e20", participant, cores[2].validator.ID()); err != nil {
		fmt.Printf("error inserting e20: %s\n", err)
	}

	event12 := hg.NewEvent([][]byte{},
		[]hg.InternalTransaction{},
		nil,
		[]string{index["e1"], index["e20"]}, //e1 and e20
		cores[1].validator.PublicKeyBytes(), 1)
	if err := insertEvent(cores, keys, index, event12, "e12", participant, cores[1].validator.ID()); err != nil {
		fmt.Printf("error inserting e12: %s\n", err)
	}
}

func insertEvent(cores []*core, keys map[uint32]*ecdsa.PrivateKey, index map[string]string,
	event *hg.Event, name string, particant uint32, creator uint32) error {

	if particant == creator {
		if err := cores[particant].signAndInsertSelfEvent(event); err != nil {
			return err
		}
		//event is not signed because passed by value
		index[name] = cores[particant].head
	} else {
		event.Sign(keys[creator])
		if err := cores[particant].insertEventAndRunConsensus(event, true); err != nil {
			return err
		}
		index[name] = event.Hex()
	}
	return nil
}

func TestEventDiff(t *testing.T) {
	cores, keys, index := initCores(3, t)

	initHashgraph(cores, keys, index, 0)

	/*
	   P0 knows

	   |  e12  |
	   |   | \ |
	   |   |   e20
	   |   | / |
	   |   /   |
	   | / |   |
	   e01 |   |        P1 knows
	   | \ |   |
	   e0  e1  e2       |   e1  |
	   0   1   2        0   1   2
	*/

	knownBy1 := cores[1].knownEvents()
	unknownBy1, err := cores[0].eventDiff(knownBy1)
	if err != nil {
		t.Fatal(err)
	}

	if l := len(unknownBy1); l != 5 {
		t.Fatalf("length of unknown should be 5, not %d", l)
	}

	expectedOrder := []string{"e0", "e2", "e01", "e20", "e12"}
	for i, e := range unknownBy1 {
		if name := getName(index, e.Hex()); name != expectedOrder[i] {
			t.Fatalf("element %d should be %s, not %s", i, expectedOrder[i], name)
		}
	}
}

func TestSync(t *testing.T) {
	cores, _, index := initCores(3, t)

	/*
	   core 0           core 1          core 2

	   e0  |   |        |   e1  |       |   |   e2
	   0   1   2        0   1   2       0   1   2
	*/

	//core 1 is going to tell core 0 everything it knows
	if err := synchronizeCores(cores, 1, 0, [][]byte{}, []hg.InternalTransaction{}); err != nil {
		t.Fatal(err)
	}

	/*
	   core 0           core 1          core 2

	   e01 |   |
	   | \ |   |
	   e0  e1  |        |   e1  |       |   |   e2
	   0   1   2        0   1   2       0   1   2
	*/

	knownBy0 := cores[0].knownEvents()
	if k := knownBy0[cores[0].validator.ID()]; k != 1 {
		t.Fatalf("core 0 should have last-index 1 for core 0, not %d", k)
	}
	if k := knownBy0[cores[1].validator.ID()]; k != 0 {
		t.Fatalf("core 0 should have last-index 0 for core 1, not %d", k)
	}
	if k := knownBy0[cores[2].validator.ID()]; k != -1 {
		t.Fatalf("core 0 should have last-index -1 for core 2, not %d", k)
	}
	core0Head, _ := cores[0].getHead()
	if core0Head.SelfParent() != index["e0"] {
		t.Fatalf("core 0 head self-parent should be e0")
	}
	if core0Head.OtherParent() != index["e1"] {
		t.Fatalf("core 0 head other-parent should be e1")
	}
	index["e01"] = core0Head.Hex()

	//core 0 is going to tell core 2 everything it knows
	if err := synchronizeCores(cores, 0, 2, [][]byte{}, []hg.InternalTransaction{}); err != nil {
		t.Fatal(err)
	}

	/*

	   core 0           core 1          core 2

	                                    |   |  e20
	                                    |   | / |
	                                    |   /   |
	                                    | / |   |
	   e01 |   |                        e01 |   |
	   | \ |   |                        | \ |   |
	   e0  e1  |        |   e1  |       e0  e1  e2
	   0   1   2        0   1   2       0   1   2
	*/

	knownBy2 := cores[2].knownEvents()
	if k := knownBy2[cores[0].validator.ID()]; k != 1 {
		t.Fatalf("core 2 should have last-index 1 for core 0, not %d", k)
	}
	if k := knownBy2[cores[1].validator.ID()]; k != 0 {
		t.Fatalf("core 2 should have last-index 0 core 1, not %d", k)
	}
	if k := knownBy2[cores[2].validator.ID()]; k != 1 {
		t.Fatalf("core 2 should have last-index 1 for core 2, not %d", k)
	}
	core2Head, _ := cores[2].getHead()
	if core2Head.SelfParent() != index["e2"] {
		t.Fatalf("core 2 head self-parent should be e2")
	}
	if core2Head.OtherParent() != index["e01"] {
		t.Fatalf("core 2 head other-parent should be e01")
	}
	index["e20"] = core2Head.Hex()

	//core 2 is going to tell core 1 everything it knows
	if err := synchronizeCores(cores, 2, 1, [][]byte{}, []hg.InternalTransaction{}); err != nil {
		t.Fatal(err)
	}

	/*

	   core 0           core 1          core 2

	                    |  e12  |
	                    |   | \ |
	                    |   |  e20      |   |  e20
	                    |   | / |       |   | / |
	                    |   /   |       |   /   |
	                    | / |   |       | / |   |
	   e01 |   |        e01 |   |       e01 |   |
	   | \ |   |        | \ |   |       | \ |   |
	   e0  e1  |        e0  e1  e2      e0  e1  e2
	   0   1   2        0   1   2       0   1   2
	*/

	knownBy1 := cores[1].knownEvents()
	if k := knownBy1[cores[0].validator.ID()]; k != 1 {
		t.Fatalf("core 1 should have last-index 1 for core 0, not %d", k)
	}
	if k := knownBy1[cores[1].validator.ID()]; k != 1 {
		t.Fatalf("core 1 should have last-index 1 for core 1, not %d", k)
	}
	if k := knownBy1[cores[2].validator.ID()]; k != 1 {
		t.Fatalf("core 1 should have last-index 1 for core 2, not %d", k)
	}
	core1Head, _ := cores[1].getHead()
	if core1Head.SelfParent() != index["e1"] {
		t.Fatalf("core 1 head self-parent should be e1")
	}
	if core1Head.OtherParent() != index["e20"] {
		t.Fatalf("core 1 head other-parent should be e20")
	}
	index["e12"] = core1Head.Hex()
}

/*
h0  |   h2
| \ | / |
|   h1  |
|  /|   |--------------------
g02 |   | R2
| \ |   |
|   \   |
|   | \ |
|   |  g21
|   | / |
|  g10  |
| / |   |
g0  |   g2
| \ | / |
|   g1  |
|  /|   |--------------------
f02 |   | R1
| \ |   |
|   \   |
|   | \ |
|   |  f21
|   | / |
|  f10  |
| / |   |
f0  |   f2
| \ | / |
|   f1  |
|  /|   |--------------------
e02 |   | R0 Consensus
| \ |   |
|   \   |
|   | \ |
|   |  e21
|   | / |
|  e10  |
| / |   |
e0  e1  e2
0   1    2
*/
type play struct {
	from        int
	to          int
	payload     [][]byte
	internalTxs []hg.InternalTransaction
}

func initConsensusHashgraph(t *testing.T) []*core {
	cores, _, _ := initCores(3, t)
	playbook := []play{
		{from: 0, to: 1, payload: [][]byte{[]byte("e10")}},
		{from: 1, to: 2, payload: [][]byte{[]byte("e21")}},
		{from: 2, to: 0, payload: [][]byte{[]byte("e02")}},
		{from: 0, to: 1, payload: [][]byte{[]byte("f1")}},
		{from: 1, to: 0, payload: [][]byte{[]byte("f0")}},
		{from: 1, to: 2, payload: [][]byte{[]byte("f2")}},

		{from: 0, to: 1, payload: [][]byte{[]byte("f10")}},
		{from: 1, to: 2, payload: [][]byte{[]byte("f21")}},
		{from: 2, to: 0, payload: [][]byte{[]byte("f02")}},
		{from: 0, to: 1, payload: [][]byte{[]byte("g1")}},
		{from: 1, to: 0, payload: [][]byte{[]byte("g0")}},
		{from: 1, to: 2, payload: [][]byte{[]byte("g2")}},

		{from: 0, to: 1, payload: [][]byte{[]byte("g10")}},
		{from: 1, to: 2, payload: [][]byte{[]byte("g21")}},
		{from: 2, to: 0, payload: [][]byte{[]byte("g02")}},
		{from: 0, to: 1, payload: [][]byte{[]byte("h1")}},
		{from: 1, to: 0, payload: [][]byte{[]byte("h0")}},
		{from: 1, to: 2, payload: [][]byte{[]byte("h2")}},
	}

	for _, play := range playbook {
		if err := syncAndRunConsensus(cores, play.from, play.to, play.payload, play.internalTxs); err != nil {
			t.Fatal(err)
		}
	}
	return cores
}

func TestConsensus(t *testing.T) {
	cores := initConsensusHashgraph(t)

	if l := len(cores[0].getConsensusEvents()); l != 6 {
		t.Fatalf("length of consensus should be 6 not %d", l)
	}

	core0Consensus := cores[0].getConsensusEvents()
	core1Consensus := cores[1].getConsensusEvents()
	core2Consensus := cores[2].getConsensusEvents()

	for i, e := range core0Consensus {
		if core1Consensus[i] != e {
			t.Fatalf("core 1 consensus[%d] does not match core 0's", i)
		}
		if core2Consensus[i] != e {
			t.Fatalf("core 2 consensus[%d] does not match core 0's", i)
		}
	}
}
func TestCoreFastForwardAfterJoin(t *testing.T) {
	cores, bobPeer, bobKey := initR2DynHashgraph(t)

	initPeerSet, err := cores[0].hg.Store.GetPeerSet(0)
	if err != nil {
		t.Fatal(err)
	}

	genesisPeerSet := clonePeerSet(t, initPeerSet.Peers)

	bobCore := newCore(
		NewValidator(bobKey, bobPeer.Moniker),
		initPeerSet,
		genesisPeerSet,
		hg.NewInmemStore(1000),
		proxy.DummyCommitCallback,
		false,
		common.NewTestEntry(t, common.TestLogLevel))

	bobCore.setHeadAndSeq()

	cores = append(cores, bobCore)

	/***************************************************************************
		Manually FastForward Bob from cores[2]

		Testing 2 scenarios:

			- AnchorBlock: check that the AnchorBlock (Block 5) is selected
						   correctly and that FuturePeerSets works.

			- Block 0: check that FastForwarding from a Round below the PeerSet
			           change works.
	***************************************************************************/

	type play struct {
		block           *hg.Block
		frame           *hg.Frame
		roundLowerBound int
	}

	plays := []play{}

	//Prepare Block 0 scenario
	block0, err := cores[2].hg.Store.GetBlock(0)
	if err != nil {
		t.Fatal(err)
	}

	frame0, err := cores[2].hg.Store.GetFrame(block0.RoundReceived())
	if err != nil {
		t.Fatal(err)
	}

	plays = append(plays, play{block0, frame0, 0})

	//Prepare AnchorBlock scenario
	anchorBlock, anchorFrame, err := cores[2].hg.GetAnchorBlockWithFrame()
	if err != nil {
		t.Fatal(err)
	}

	plays = append(plays, play{anchorBlock, anchorFrame, 6})

	/***************************************************************************
		Run the same test for both scenarios
	***************************************************************************/

	for _, p := range plays {

		/***********************************************************************
			FastForward, Sync, and Run Consensus
		***********************************************************************/

		marshalledBlock, err := p.block.Marshal()
		if err != nil {
			t.Fatal(err)
		}

		var unmarshalledBlock hg.Block
		err = unmarshalledBlock.Unmarshal(marshalledBlock)
		if err != nil {
			t.Fatal(err)
		}

		marshalledFrame, err := p.frame.Marshal()
		if err != nil {
			t.Fatal(err)
		}

		var unmarshalledFrame hg.Frame
		err = unmarshalledFrame.Unmarshal(marshalledFrame)
		if err != nil {
			t.Fatal(err)
		}

		err = cores[3].fastForward(&unmarshalledBlock, &unmarshalledFrame)
		if err != nil {
			t.Fatal(err)
		}

		//continue after FastForward
		err = syncAndRunConsensus(cores, 2, 3, [][]byte{}, []hg.InternalTransaction{})
		if err != nil {
			t.Fatal(err)
		}

		/***********************************************************************
			Check Known
		***********************************************************************/

		knownBy3 := cores[3].knownEvents()
		if err != nil {
			t.Fatal(err)
		}

		expectedKnown := map[uint32]int{
			cores[0].validator.ID(): 9,
			cores[1].validator.ID(): 15,
			cores[2].validator.ID(): 10,
			cores[3].validator.ID(): 0,
		}

		if !reflect.DeepEqual(knownBy3, expectedKnown) {
			t.Fatalf("Cores[3].Known should be %v, not %v", expectedKnown, knownBy3)
		}

		/***********************************************************************
			Check Rounds
		***********************************************************************/

		//The fame of witnesses of the FastForward's Block RoundReceived and
		//below are not reprocessed after Reset. No need to test those rounds.
		for i := p.roundLowerBound; i <= 8; i++ {
			c3RI, err := cores[3].hg.Store.GetRound(i)
			if err != nil {
				t.Fatal(err)
			}

			c2RI, err := cores[2].hg.Store.GetRound(i)
			if err != nil {
				t.Fatal(err)
			}

			c3RIw := c3RI.Witnesses()
			c2RIw := c2RI.Witnesses()
			sort.Strings(c3RIw)
			sort.Strings(c2RIw)

			if !reflect.DeepEqual(c3RIw, c2RIw) {
				t.Logf("Round(%d).Witnesses do not match", i)
			}

			if !reflect.DeepEqual(c3RI.CreatedEvents, c3RI.CreatedEvents) {
				t.Logf("Round(%d).CreatedEvents do not match", i)
			}

			c3RIr := c3RI.ReceivedEvents
			c2RIr := c2RI.ReceivedEvents
			sort.Strings(c3RIr)
			sort.Strings(c2RIr)

			if !reflect.DeepEqual(c3RIw, c2RIw) {
				t.Logf("Round(%d).ReceivedEvents do not match", i)
			}
		}

		/***********************************************************************
			Check PeerSets
		***********************************************************************/

		for i := p.roundLowerBound; i <= 8; i++ {
			c3PS, err := cores[3].hg.Store.GetPeerSet(i)
			if err != nil {
				t.Fatal(err)
			}

			c2PS, err := cores[2].hg.Store.GetPeerSet(i)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(c3PS.Hex(), c2PS.Hex()) {
				t.Fatalf("PeerSet(%d) does not match", i)
			}
		}

		/***********************************************************************
			Check Consensus Rounds and Blocks
		***********************************************************************/

		if r := cores[3].getLastConsensusRoundIndex(); r == nil || *r != 6 {
			t.Fatalf("Cores[3] last consensus Round should be 4, not %v", *r)
		}

		if lbi := cores[3].hg.Store.LastBlockIndex(); lbi != 5 {
			t.Fatalf("Cores[3].hg.LastBlockIndex should be 5, not %d", lbi)
		}
	}

}

/******************************************************************************/

func synchronizeCores(cores []*core, from int, to int, payload [][]byte, internalTxs []hg.InternalTransaction) error {
	knownByTo := cores[to].knownEvents()
	unknownByTo, err := cores[from].eventDiff(knownByTo)
	if err != nil {
		return err
	}

	unknownWire, err := cores[from].toWire(unknownByTo)
	if err != nil {
		return err
	}

	cores[to].addTransactions(payload)

	for _, it := range internalTxs {
		cores[to].addInternalTransaction(it)
	}

	return cores[to].sync(cores[from].validator.ID(), unknownWire)
}

func syncAndRunConsensus(cores []*core, from int, to int, payload [][]byte, internalTxs []hg.InternalTransaction) error {
	if err := synchronizeCores(cores, from, to, payload, internalTxs); err != nil {
		return err
	}
	cores[to].processSigPool()
	return nil
}
func getName(index map[string]string, hash string) string {
	for name, h := range index {
		if h == hash {
			return name
		}
	}
	return fmt.Sprintf("%s not found", hash)
}
