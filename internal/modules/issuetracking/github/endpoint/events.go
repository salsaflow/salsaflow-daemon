package endpoint

import (
	// Stdlib
	"errors"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/github/events"

	// Vendor
	"github.com/google/go-github/github"
)

type eventHandler struct {
	client *github.Client
}

func init() {
	if err := ensureInterfaces(); err != nil {
		panic(err)
	}
}

func ensureInterfaces() error {
	var handler interface{} = &eventHandler{}

	if _, ok := handler.(events.IssueCommentEventHandler); !ok {
		return errors.New("eventHandler does not implement events.IssueCommentEventHandler")
	}

	if _, ok := handler.(events.IssuesEventHandler); !ok {
		return errors.New("eventHandler does not implement events.IssuesEventHandler")
	}

	return nil
}
