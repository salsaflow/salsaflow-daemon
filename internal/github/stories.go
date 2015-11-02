package github

import "github.com/google/go-github/github"

func LabeledWith(issue *github.Issue, labelName string) bool {
	for _, label := range issue.Labels {
		if *label.Name == labelName {
			return true
		}
	}
	return false
}
