package events

import (
	// Stdlib
	"net/http"
)

type CommitCommentEventHandler interface {
	HandleCommitCommentEvent(rw http.ResponseWriter, r *http.Request, e *CommitCommentEvent)
}

type IssueCommentEventHandler interface {
	HandleIssueCommentEvent(rw http.ResponseWriter, r *http.Request, e *IssueCommentEvent)
}

type IssuesEventHandler interface {
	HandleIssuesEvent(rw http.ResponseWriter, r *http.Request, e *IssuesEvent)
}
