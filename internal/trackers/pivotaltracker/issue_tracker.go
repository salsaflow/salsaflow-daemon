package pivotaltracker

import (
	// Stdlib
	"fmt"
	"strconv"
	"strings"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/trackers/common"

	// Vendor
	"github.com/salsita/go-pivotaltracker/v5/pivotal"
)

const Id = "Pivotal Tracker"

func Factory() (common.IssueTracker, error) {
	config := GetConfig()
	return &issueTracker{pivotal.NewClient(config.Token)}, nil
}

type issueTracker struct {
	client *pivotal.Client
}

func (tracker *issueTracker) FindStoryByTag(storyTag string) (common.Story, error) {
	pid, sid, err := parseStoryTag(storyTag)
	if err != nil {
		return nil, err
	}

	story, _, err := tracker.client.Stories.Get(pid, sid)
	if err != nil {
		return nil, err
	}

	return &commonStory{tracker.client, story}, nil
}

func parseStoryTag(storyTag string) (pid, sid int, err error) {
	parts := strings.Split(storyTag, "/")
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("Pivotal Tracker: malformed story tag: %v", storyTag)
	}

	pidString := parts[0]
	pid, err = strconv.Atoi(pidString)
	if err != nil {
		return 0, 0, fmt.Errorf("Pivotal Tracker: malformed project ID: %v", pidString)
	}

	sidString := parts[2]
	sid, err = strconv.Atoi(sidString)
	if err != nil {
		return 0, 0, fmt.Errorf("Pivotal Tracker: malformed story ID: %v", sidString)
	}

	return pid, sid, nil
}
