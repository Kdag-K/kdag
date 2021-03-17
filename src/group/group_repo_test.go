package group

import (
	"reflect"
	"testing"

	"github.com/Kdag-K/kdag/src/peers"
)

// Test inserting and updating a single group.
func TestSetGroup(t *testing.T) {
	repo := NewInmemGroupRepo()

	group := NewGroup(
		"",
		"TestGroup",
		"TestApp",
		[]*peers.Peer{
			peers.NewPeer("pub1", "net1", "peer1"),
		},
	)

	var time int64 = 0

	// Insert new group, verify that it can be retrieved and that the LastUpdate
	// time was updated

	groupID, err := repo.SetGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	retrievedGroup, err := repo.GetGroup(groupID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(group, retrievedGroup) {
		t.Fatalf("Retrieved group should be %#v, not %#v", group, retrievedGroup)
	}

	if retrievedGroup.LastUpdated < time {
		t.Fatalf("group LastUpdated should have increased")
	}

	// Update the group and verify that its LastUpdate time has been updated

	time = retrievedGroup.LastUpdated

	// Add a peer
	group.Peers = append(group.Peers, peers.NewPeer("pub2", "net2", "peer2"))

	_, err = repo.SetGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	retrievedGroup2, err := repo.GetGroup(groupID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(group, retrievedGroup2) {
		t.Fatalf("Retrieved group should be %#v, not %#v", group, retrievedGroup2)
	}

	if retrievedGroup2.LastUpdated < time {
		t.Fatalf("group LastUpdated should have increased")
	}
}
