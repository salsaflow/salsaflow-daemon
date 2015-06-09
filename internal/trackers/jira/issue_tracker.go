package jira

import (
	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/trackers/common"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/jirautils"

	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

const Id = "JIRA"

func Factory() (common.IssueTracker, error) {
	client, err := jirautils.NewClient()
	if err != nil {
		return nil, err
	}
	return &issueTracker{client}, nil
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
