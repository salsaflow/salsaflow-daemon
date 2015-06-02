package pivotaltracker

import (
	// Stdlib
	"log"

	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

type commonStory struct {
	client *jira.Client
	issue  *jira.Issue
}

// common.Story interface implementation ---------------------------------------

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
	// List the remote link associated with this issue.
	links, err := s.listRemoteLinks()
	if err != nil {
		return nil, err
	}

	var (
		linkFound   = false
		allResolved = true
	)
	for _, link := range links {
		if link.Object.Title == rrID {
			linkFound = true
			continue
		}
		if !link.Object.Status.Resolved {
			allResolved = false
		}
	}

	// In case the link is there, and it should be, mark it as resolved.
	if linkFound {
		if err := s.setRemoteLinkResolved(rrID, rrURL, true); err != nil {
			return err
		}
	} else {
		log.Printf("JIRA: remote link not found: issue %v, review issue %v\n", s.issue.Key, rrURL)
	}

	// In case all links are resolved, mark the issue as reviewed.
	if allResolved {
		return s.MarkAsReviewed()
	}

	return nil
}

func (s *commonStory) OnReviewRequestReopened(rrID, rrURL string) error {
	// Make sure the remote link exists.
	link, err := s.findRemoteLink(rrID)
	if err != nil {
		return err
	}
	if link == nil {
		log.Printf("JIRA: remote link not found: issue %v, review issue %v\n", s.issue.Key, rrURL)
		return nil
	}

	// Mark the link as resolved.
	return s.setRemoteLinkResolved(rrID, rrURL, false)
}

func (s *commonStory) MarkAsReviewed() error {
	panic("Not implemented")
}

// Internal methods ------------------------------------------------------------

func (s *commonStory) setRemoteLinkResolved(rrID, rrURL string, resolved bool) error {
	// Prepare the update object.
	var update jira.IssueRemoteLink
	update.GlobalId = rrURL
	update.Object.Status.Resolved = resolved

	// Update the remote link.
	return s.updateRemoteLink(&update)
}

func (s *commonStory) createRemoteLink(link *jira.IssueRemoteLink) error {
	_, _, err := s.client.IssueRemoteLinks.Create(s.issue.Key, link)
	return err
}

func (s *commonStory) listRemoteLinks() ([]*jira.IssueRemoteLink, error) {
	links, _, err := s.client.IssueRemoteLinks.List(s.issue.Key)
	return links, err
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
