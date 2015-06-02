package pivotaltracker

import (
	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

type commonStory struct {
	client *jira.Client
	issue  *jira.Issue
}

func (s *commonStory) OnReviewRequestOpened(rrID, rrURL string) error {
	// Prepare the remote link object.
	var link jira.IssueRemoteLink
	link.GlobalId = rrURL
	link.Object.Title = rrID
	link.Object.URL = rrURL

	// Create the remote link.
	return s.createRemoteLink(&link)
}

func (s *commonStory) OnReviewRequestClosed(rrID, rrURL string) error {
	return s.setRemoteLinkResolved(rrID, rrURL, true)
}

func (s *commonStory) OnReviewRequestReopened(rrID, rrURL string) error {
	return s.setRemoteLinkResolved(rrID, rrURL, false)
}

func (s *commonStory) setRemoteLinkResolved(rrID, rrURL string, resolved bool) error {
	// Find the relevant remote link.
	link, err := s.findRemoteLink(rrID)
	if err != nil {
		return err
	}
	if link == nil {
		return nil
	}

	// Prepare the update object.
	var update jira.IssueRemoteLink
	update.GlobalId = rrURL
	update.Object.Status.Resolved = resolved

	// Update the remote link.
	return s.updateRemoteLink(&update)
}

func (s *commonStory) MarkAsReviewed() error {
	panic("Not implemented")
}

func (s *commonStory) createRemoteLink(link *jira.IssueRemoteLink) error {
	_, _, err := s.client.IssueRemoteLinks.Create(s.issue.Key, link)
	return err
}

func (s *commonStory) updateRemoteLink(link *jira.IssueRemoteLink) error {
	_, _, err := s.client.IssueRemoteLinks.Update(s.issue.Key, link)
	return err
}

func (s *commonStory) findRemoteLink(rrID string) (*jira.IssueRemoteLink, error) {
	links, _, err := s.client.IssueRemoteLinks.List(s.issue.Key)
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if link.Object.Title == rrID {
			return &link, nil
		}
	}
	return nil, nil
}
