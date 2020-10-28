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
