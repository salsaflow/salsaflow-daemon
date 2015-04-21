package pivotaltracker

import (
	"github.com/tchap/salsaflow-daemon/internal/trackers/common"
)

const Id = "Pivotal Tracker"

func Factory() common.IssueTracker {
	return &issueTracker{}
}

type issueTracker struct{}

func (tracker *issueTracker) FindStoryById(projectId, storyId string) (common.Story, error) {
	panic("Not implemented")
}
