package github

import (
	// Stdlib
	"fmt"

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
	config := GetConfig()
	return s.setStateLabel(config.BeingImplementedLabel)
}

func (s *commonStory) MarkAsReviewed() error {
	config := GetConfig()
	return s.setStateLabel(config.ReviewedLabel)
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
	// Get the list of labels to be used.
	labelNames := setStateLabel(s.issue.Labels, label)

	// Replace the labels.
	var (
		client   = s.client
		owner    = s.owner
		repo     = s.repo
		issueNum = *s.issue.Number
	)
	labels, _, err := client.Issues.ReplaceLabelsForIssue(owner, repo, issueNum, labelNames)
	if err != nil {
		return err
	}

	s.issue.Labels = labels
	return nil
}

func setStateLabel(currentLabels []github.Label, stateLabel string) (labels []string) {
	config := GetConfig()
	labels = make([]string, 0, len(currentLabels)+1)
	for _, label := range currentLabels {
		name := *label.Name
		switch name {
		case config.ApprovedLabel:
		case config.BeingImplementedLabel:
		case config.ImplementedLabel:
		case config.ReviewedLabel:
		case config.SkipReviewLabel:
		case config.PassedTestingLabel:
		case config.FailedTestingLabel:
		case config.SkipTestingLabel:
		case config.StagedLabel:
		case config.RejectedLabel:
		default:
			labels = append(labels, name)
		}
	}
	labels = append(labels, stateLabel)
	return labels
}
