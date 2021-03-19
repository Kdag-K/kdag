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
func TestDelGroup(t *testing.T) {
	repo := NewInmemGroupRepo()

	group := NewGroup(
		"",
		"TestGroup",
		"TestApp",
		[]*peers.Peer{
			peers.NewPeer("pub1", "net1", "peer1"),
		},
	)

	groupID, err := repo.SetGroup(group)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DelGroup(groupID)
	if err != nil {
		t.Fatal(err)
	}

	retrievedGroup, err := repo.GetGroup(groupID)
	if retrievedGroup != nil || err == nil {
		t.Fatalf("Retrieving deleted group should be return nil and error")
	}
}

// Test inserting groups with different AppIDs and fetching all groups or groups
// by AppID.
func TestGetGroups(t *testing.T) {
	repo := NewInmemGroupRepo()

	group1 := NewGroup(
		"",
		"TestGroup1",
		"TestApp1",
		[]*peers.Peer{
			peers.NewPeer("pub1", "net1", "peer1"),
		},
	)

	group2 := NewGroup(
		"",
		"TestGroup2",
		"TestApp2",
		[]*peers.Peer{
			peers.NewPeer("pub1", "net1", "peer1"),
		},
	)

	_, err := repo.SetGroup(group1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.SetGroup((group2))
	if err != nil {
		t.Fatal(err)
	}

	allGroups, err := repo.GetAllGroups()
	if err != nil {
		t.Fatal(err)
	}

	if len(allGroups) != 2 {
		t.Fatalf("Repo should contain 2 groups, not %d", len(allGroups))
	}

	app1Groups, err := repo.GetAllGroupsByAppID("TestApp1")
	if err != nil {
		t.Fatal(err)
	}

	if len(app1Groups) != 1 {
		t.Fatalf("App1 should contain 1 group, not %d", len(app1Groups))
	}

	app2Groups, err := repo.GetAllGroupsByAppID("TestApp2")
	if err != nil {
		t.Fatal(err)
	}

	if len(app2Groups) != 1 {
		t.Fatalf("App2 should contain 1 group, not %d", len(app2Groups))
	}
}
