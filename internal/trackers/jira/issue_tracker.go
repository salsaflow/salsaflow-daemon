package pivotaltracker

import (
	// Stdlib
	"fmt"
	"os"
	"strconv"
	"strings"

	// Internal
	"github.com/tchap/salsaflow-daemon/internal/trackers/common"

	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

const Id = "JIRA"

func Factory() (common.IssueTracker, error) {
	panic("Not implemented")
}

type issueTracker struct {
	client *jira.Client
}

func (tracker *issueTracker) FindStoryByTag(storyTag string) (common.Story, error) {
	story, _, err := tracker.client.Issues.Get(storyTag)
	if err != nil {
		return nil, err
	}
	return &commonStory{tracker.client, story}, nil
}
