package tracker

import (
	// Stdlib
	"fmt"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/config"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/util"

	// Vendor
	"github.com/google/go-github/github"
)

type commonStory struct {
	client *github.Client
	issue  *github.Issue
	owner  string
	repo   string
}

func (s *commonStory) OnReviewRequestOpened(rrID, rrURL string) error {
	return s.addComment(fmt.Sprintf("Review request [#%v](%v) opened.", rrID, rrURL))
}

func (s *commonStory) OnReviewRequestClosed(rrID, rrURL string) error {
	return nil
}

func (s *commonStory) OnReviewRequestReopened(rrID, rrURL string) error {
	return s.setStateLabel(config.Get().BeingImplementedLabel)
}

func (s *commonStory) MarkAsReviewed() error {
	return s.setStateLabel(config.Get().ReviewedLabel)
}

func (s *commonStory) addComment(text string) error {
	var (
		client   = s.client
		owner    = s.owner
		repo     = s.repo
		issueNum = *s.issue.Number
	)
	_, _, err := client.Issues.CreateComment(owner, repo, issueNum, &github.IssueComment{
		Body: &text,
	})
	return err
}

func (s *commonStory) setStateLabel(label string) error {
	return util.ReplaceWorkflowLabels(s.client, s.owner, s.repo, s.issue, []string{label})
}
