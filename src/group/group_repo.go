package group

import (
	"fmt"
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

// GetAllGroups implements the GroupRepo interface and returns all the
// groups
func (igr *InmemGroupRepo) GetAllGroups() (map[string]*Group, error) {
	igr.Lock()
	defer igr.Unlock()

	return igr.groupsByID, nil
}

// GetGroup implements the GroupRepo interface and returns a group by ID
func (igr *InmemGroupRepo) GetGroup(id string) (*Group, error) {
	igr.Lock()
	defer igr.Unlock()

	g, ok := igr.groupsByID[id]
	if !ok {
		return nil, fmt.Errorf("Group %s not found", id)
	}
	return g, nil
}
// GetAllGroupsByAppID implements the GroupRepo interface and returns all
// the groups associated with an AppID
func (igr *InmemGroupRepo) GetAllGroupsByAppID(appID string) (map[string]*Group, error) {
	igr.Lock()
	defer igr.Unlock()

	res := make(map[string]*Group)

	appGroups, ok := igr.groupsByAppID[appID]
	if !ok {
		return res, nil
	}

	for _, gid := range appGroups {
		res[gid] = igr.groupsByID[gid]
		println ( " gid : " + gid)
		println(res[gid])
	}

	return res, nil
}