package group

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
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
	DelGroup(groupID string) error
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
		println(" gid : " + gid)
		println(res[gid])
	}

	return res, nil
}
func (igr *InmemGroupRepo) SetGroup(group *Group) (string, error) {
	if group.AppID == "" {
		return "", fmt.Errorf("Group AppID not specified")
	}

	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	group.LastUpdated = time.Now().Unix()

	igr.Lock()
	defer igr.Unlock()
	// If the group does not exist, add it to the AppID index
	if _, gok := igr.groupsByID[group.ID]; !gok {
		appGroups, aok := igr.groupsByAppID[group.AppID]
		if !aok {
			appGroups = []string{}
		}
		appGroups = append(appGroups, group.ID)
		igr.groupsByAppID[group.AppID] = appGroups
	}
	// Set group in main index
	igr.groupsByID[group.ID] = group
	return group.ID, nil
}

// DelGroup implements the GroupRepo interface and removes a group from
// the map
func (igr *InmemGroupRepo) DelGroup(id string) error {
	igr.Lock()
	defer igr.Unlock()
	// If the group exists, remove it from the AppID index
	if g, gok := igr.groupsByID[id]; gok {
		appGroups, aok := igr.groupsByAppID[g.AppID]
		if aok {
			igr.groupsByAppID[g.AppID] = appGroups
		}
	}
	delete(igr.groupsByID, id)
	return nil
}

// GetAllGroupsByAppID implements the GroupRepo interface and returns all
// the groups associated with an AppID
func (igr *InmemGroupRepo) GetAllGroupsByAppID(appID string) (map[string]*Group, error) {
	igr.Lock()
	defer igr.Unlock()

	res := make(map[string]*Group)

	return res, nil
}
