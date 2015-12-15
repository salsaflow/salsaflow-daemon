package endpoint

import (
	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/config"

	// Vendor
	"github.com/google/go-github/github"
)

func isStoryIssue(issue *github.Issue, cfg config.Config) bool {
	for _, storyLabel := range cfg.StoryLabels {
		for _, issueLabel := range issue.Labels {
			if *issueLabel.Name == storyLabel {
				return true
			}
		}
	}
	return false
}
