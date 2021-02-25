package group
import (
	"github.com/Kdag-K/kdag/src/peers"
	"sync"
)
// InmemGroupRepo implements the GroupRepo interface with an inmem
// map of groups. It is thread safe.
type InmemGroupRepo struct {
	sync.Mutex
	groupsByID    map[string]*Group   // [group ID] => Group
	groupsByAppID map[string][]string // [app ID] => [GroupID,...]
}

// GroupRepo defines an interface for a repository where groups can be
// queried, added, and manipulated. It should be thread safe.
type GroupRepo interface {
	GetAllGroups() (map[string]*Group, error)
	GetAllGroupsByAppID(appID string) (map[string]*Group, error)
	GetGroup(groupID string) (*Group, error)
	SetGroup(group *Group) (string, error)
	DeleteGroup(groupID string) error
}
// NewInmemGroupRepo instantiates a new InmemGroupRepo
func NewInmemGroupRepo() *InmemGroupRepo {
	return &InmemGroupRepo{
		groupsByID:    make(map[string]*Group),
		groupsByAppID: make(map[string][]string),
	}
}

func TestGetGroups() {
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

	allGroups,_  := repo.GetAllGroups()

}

// GetAllGroups implements the GroupRepo interface and returns all the
// groups
func (igr *InmemGroupRepo) GetAllGroups() (map[string]*Group, error) {
	igr.Lock()
	defer igr.Unlock()

	return igr.groupsByID, nil
}