package jira

import (
	// Stdlib
	"log"

	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

const (
	iconOpen   = "https://raw.githubusercontent.com/salsaflow/salsaflow-daemon/develop/internal/trackers/jira/img/blacktocat-16.png"
	iconClosed = "https://raw.githubusercontent.com/salsaflow/salsaflow-daemon/develop/internal/trackers/jira/img/greentocat-16.png"
)

type commonStory struct {
	client *jira.Client
	issue  *jira.Issue
}

// common.Story interface implementation ---------------------------------------

func (s *commonStory) OnReviewRequestOpened(rrID, rrURL string) error {
	// Prepare the remote link object.
	var link jira.RemoteIssueLink
	link.GlobalId = rrURL
	link.Object.Title = toTitle(rrID)
	link.Object.URL = rrURL
	link.Object.Icon.URL = iconOpen

	// Create the remote link.
	return s.createRemoteLink(&link)
}

func (s *commonStory) OnReviewRequestClosed(rrID, rrURL string) error {
	// List the remote link associated with this issue.
	links, err := s.listRemoteLinks()
	if err != nil {
		return nil
	}

	var (
		title       = toTitle(rrID)
		linkFound   = false
		allResolved = true
	)
	for _, link := range links {
		if link.Object.Title == title {
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
		return s.markIssueAsReviewed()
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
	// No need to do anything really, everything is in OnReviewRequestClosed.

	// Mark the link as resolved.
	return s.setRemoteLinkResolved(rrID, rrURL, false)
}

func (s *commonStory) MarkAsReviewed() error {
	// No need to do anything really, everything is in OnReviewRequestClosed.
	return nil
}

// Internal methods ------------------------------------------------------------

func (s *commonStory) setRemoteLinkResolved(rrID, rrURL string, resolved bool) error {
	// Prepare the update object.
	var update jira.RemoteIssueLink
	update.GlobalId = rrURL
	update.Object.Title = toTitle(rrID)
	update.Object.URL = rrURL
	update.Object.Status.Resolved = resolved

	var icon string
	if resolved {
		icon = iconClosed
	} else {
		icon = iconOpen
	}
	update.Object.Icon.URL = icon

	// Update the remote link.
	return s.updateRemoteLink(&update)
}

func (s *commonStory) createRemoteLink(link *jira.RemoteIssueLink) error {
	_, err := s.client.RemoteIssueLinks.Create(s.issue.Key, link)
	return err
}

func (s *commonStory) listRemoteLinks() ([]*jira.RemoteIssueLink, error) {
	links, _, err := s.client.RemoteIssueLinks.List(s.issue.Key)
	return links, err
}

func (s *commonStory) updateRemoteLink(link *jira.RemoteIssueLink) error {
	_, err := s.client.RemoteIssueLinks.Update(s.issue.Key, link)
	return err
}

func (s *commonStory) findRemoteLink(rrID string) (*jira.RemoteIssueLink, error) {
	links, _, err := s.client.RemoteIssueLinks.List(s.issue.Key)
	if err != nil {
		return nil, err
	}

	title := toTitle(rrID)
	for _, link := range links {
		if link.Object.Title == title {
			return link, nil
		}
	}
	return nil, nil
}

func (s *commonStory) markIssueAsReviewed() error {
	switch s.issue.Fields.Status.Id {
	// In case the issue is still Being Implemented, we are done.
	case statusIdBeingImplemented:
		return nil

	// In case the issue is Implemented, we proceed with the transition.
	case statusIdImplemented:
		_, err := s.client.Issues.PerformTransition(s.issue.Key, transitionIdMarkAsReviewed)
		return err

	// By default we log a warning and return.
	default:
		log.Printf(
			"JIRA: issue %v: not Implemented nor Being Implemented\n", s.issue.Key)
		return nil
	}
}

func toTitle(rrID string) string {
	return "Review issue #" + rrID
}
