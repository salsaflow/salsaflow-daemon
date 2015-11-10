package tracker

import (
	// Stdlib
	"fmt"

	// Internal
	githubutil "github.com/salsaflow/salsaflow-daemon/internal/github"
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
	// Add 'implemented' unless 'qa+' or 'no qa' is there already.
	var (
		c    = config.Get()
		add  []string
		keep []string
	)
	switch {
	case githubutil.LabeledWith(s.issue, c.PassedTestingLabel):
		keep = []string{c.PassedTestingLabel}
	case githubutil.LabeledWith(s.issue, c.SkipTestingLabel):
		keep = []string{c.SkipTestingLabel}
	default:
		add = []string{c.ImplementedLabel}
	}
	return s.updateStateLabels(add, keep)
}

func (s *commonStory) MarkAsReviewed() error {
	// Add 'reviewed', but also keep 'qa+' and 'no qa'.
	var (
		c    = config.Get()
		add  = []string{c.ReviewedLabel}
		keep []string
	)
	switch {
	case githubutil.LabeledWith(s.issue, c.PassedTestingLabel):
		keep = []string{c.PassedTestingLabel}
	case githubutil.LabeledWith(s.issue, c.SkipTestingLabel):
		keep = []string{c.SkipTestingLabel}
	}
	return s.updateStateLabels(add, keep)
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

func (s *commonStory) updateStateLabels(add, keep []string) error {
	return util.ReplaceWorkflowLabels(s.client, s.owner, s.repo, s.issue, add, keep)
}
