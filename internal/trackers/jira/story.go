package pivotaltracker

import (
	// Stdlib
	"fmt"

	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

type commonStory struct {
	client *jira.Client
	issue  *jira.Issue
}

func (s *commonStory) OnReviewRequestOpened(rrID, rrURL string) error {
	var link jira.IssueRemoteLink
	link.Object.Title = rrID
	link.Object.URL = rrURL

	// TODO: We need the issue key here.

	_, err := s.client.IssueRemoteLinks.Create(&link)
	return err
}

func (s *commonStory) OnReviewRequestClosed(rrID, rrURL string) error {
	panic("Not implemented")
}

func (s *commonStory) OnReviewRequestReopened(rrID, rrURL string) error {
	panic("Not implemented")
}

func (s *commonStory) MarkAsReviewed() error {
	panic("Not implemented")
}
