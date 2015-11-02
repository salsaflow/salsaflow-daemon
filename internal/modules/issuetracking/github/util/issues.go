package util

import (
	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/config"

	// Vendor
	"github.com/google/go-github/github"
)

func ReplaceWorkflowLabels(
	client *github.Client,
	owner string,
	repo string,
	issue *github.Issue,
	labels []string,
) error {
	// Get the list of labels to be used.
	cfg := config.Get()
	labelNames := make([]string, 0, len(issue.Labels)+len(labels))
	labelNames = append(labelNames, labels...)
	for _, label := range issue.Labels {
		name := *label.Name
		switch name {
		case cfg.ApprovedLabel:
		case cfg.BeingImplementedLabel:
		case cfg.ImplementedLabel:
		case cfg.ReviewedLabel:
		case cfg.SkipReviewLabel:
		case cfg.PassedTestingLabel:
		case cfg.FailedTestingLabel:
		case cfg.SkipTestingLabel:
		case cfg.StagedLabel:
		case cfg.RejectedLabel:
		default:
			labelNames = append(labelNames, name)
		}
	}

	// Replace the labels.
	ls, _, err := client.Issues.ReplaceLabelsForIssue(owner, repo, *issue.Number, labelNames)
	if err != nil {
		return err
	}

	issue.Labels = ls
	return nil
}
