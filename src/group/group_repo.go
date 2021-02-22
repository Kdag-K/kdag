package group
import (
)

// GroupRepo defines an interface for a repository where groups can be
// queried, added, and manipulated. It should be thread safe.
type GroupRepo interface {
	GetAllGroups() (map[string]*Group, error)
	GetAllGroupsByAppID(appID string) (map[string]*Group, error)
	GetGroup(groupID string) (*Group, error)
	SetGroup(group *Group) (string, error)
	DeleteGroup(groupID string) error
}