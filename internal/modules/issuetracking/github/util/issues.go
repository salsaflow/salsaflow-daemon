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
	add []string,
	keep []string,
) error {
	// Get the list of labels to be used.
	shouldKeep := func(label string) bool {
		for _, keepName := range keep {
			if keepName == label {
				return true
			}
		}
		return false
	}

	c := config.Get()
	labelNames := make([]string, 0, len(issue.Labels)+len(add))
	labelNames = append(labelNames, add...)
	for _, label := range issue.Labels {
		name := *label.Name

		if shouldKeep(name) {
			labelNames = append(labelNames, name)
			continue
		}

		switch name {
		case c.ApprovedLabel:
		case c.BeingImplementedLabel:
		case c.ImplementedLabel:
		case c.ReviewedLabel:
		case c.SkipReviewLabel:
		case c.PassedTestingLabel:
		case c.FailedTestingLabel:
		case c.SkipTestingLabel:
		case c.StagedLabel:
		case c.RejectedLabel:
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
