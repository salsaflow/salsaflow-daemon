package events

import (
	// Vendor
	"github.com/google/go-github/github"
)

type CommitCommentEvent struct {
	Action  *string                   `json:"action,omitempty"`
	Comment *github.RepositoryComment `json:"comment,omitempty"`
	Repo    *github.Repository        `json:"repository,omitempty"`
	Sender  *github.User              `json:"sender,omitempty"`
}

type IssueCommentEvent struct {
	Action  *string              `json:"action,omitempty"`
	Issue   *github.Issue        `json:"issue,omitempty"`
	Comment *github.IssueComment `json:"comment,omitempty"`
	Repo    *github.Repository   `json:"repository,omitempty"`
	Sender  *github.User         `json:"sender,omitempty"`
}

type IssuesEvent struct {
	Action *string            `json:"action,omitempty"`
	Issue  *github.Issue      `json:"issue,omitempty"`
	Label  *github.Label      `json:"label,omitempty"`
	Repo   *github.Repository `json:"repository,omitempty"`
	Sender *github.User       `json:"sender,omitempty"`
}
