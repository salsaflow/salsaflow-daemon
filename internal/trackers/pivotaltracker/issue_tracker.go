package pivotaltracker

import (
	// Internal
	"github.com/tchap/salsaflow-daemon/internal/trackers/common"

	// Vendor
	"github.com/salsita/go-pivotaltracker/v5/pivotal"
)

const Id = "Pivotal Tracker"

const EnvToken = "PIVOTALTRACKER_TOKEN"

func Factory() common.IssueTracker {
	return &issueTracker{pivotal.NewClient(os.Getenv(EnvToken))}
}

type issueTracker struct {
	client *pivotal.Client
}

func (tracker *issueTracker) FindStoryById(projectId, storyId string) (common.Story, error) {
	pid, err := strconv.Atoi(projectId)
	if err != nil {
		return nil, fmt.Errorf("not a valid Pivotal Tracker project ID: %v", projectId)
	}

	sid, err := strconv.Atoi(storyId)
	if err != nil {
		return nil, fmt.Errorf("not a valid Pivotal Tracker story ID: %v", storyId)
	}

	story, _, err := client.Stories.Get(pid, sid)
	if err != nil {
		return nil, err
	}

	return &commonStory{client, story}, nil
}
