package modules

import (
	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/common"
	gh "github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github"
	ghTracker "github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/tracker"
	pt "github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker"
	ptTracker "github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/tracker"
)

type factoryFunc func() (common.IssueTracker, error)

var factories = map[string]factoryFunc{
	gh.ModuleId: ghTracker.Factory,
	pt.ModuleId: ptTracker.Factory,
}

// GetIssueTracker can be used to get a common.IssueTracker for the given module ID.
// In case there is no factory registered for the given ID, *ErrUnknownTrackerId is returned.
func GetIssueTracker(moduleId string) (common.IssueTracker, error) {
	// Rewrite deprecated values.
	switch moduleId {
	case "GitHub Issues":
		moduleId = gh.ModuleId
	case "Pivotal Tracker":
		moduleId = pt.ModuleId
	}

	// Get factory associated with the given module ID.
	factory, ok := factories[moduleId]
	if !ok {
		return nil, &ErrUnknownModuleId{moduleId}
	}

	// Return a new IssueTracker instance.
	return factory()
}
