package trackers

import (
	// Stdlib
	"fmt"

	// Internal
	"github.com/tchap/salsaflow-daemon/internal/trackers/common"
	"github.com/tchap/salsaflow-daemon/internal/trackers/pivotaltracker"
)

// Errors ----------------------------------------------------------------------

type ErrUnknownTrackerId struct {
	id string
}

func (err *ErrUnknownTrackerId) Error() string {
	return fmt.Sprintf("unknown issue tracker id: %v", err.id)
}

// Factory ---------------------------------------------------------------------

type factoryFunc func() common.IssueTracker

var factories = map[string]factoryFunc{
	pivotaltracker.Id: pivotaltracker.Factory,
}

// GetIssueTracker can be used to get a common.IssueTracker for the given ID.
// In case there is no factory registered for the given ID, *ErrUnknownTrackerId is returned.
func GetIssueTracker(id string) (common.IssueTracker, error) {
	factory, ok := factories[id]
	if !ok {
		return nil, &ErrUnknownTrackerId{id}
	}
	return factory(), nil
}
