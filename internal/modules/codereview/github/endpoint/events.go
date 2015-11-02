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
	// Panic in case eventHandler is not implement the right interfaces.
	// Optimally the compiler would check this, but that is not possible here,
	// so we at least panic at program startup when this is not correct.
	if err := ensureInterfaces(); err != nil {
		panic(err)
	}
}

func ensureInterfaces() error {
	var handler interface{} = &eventHandler{}

	if _, ok := handler.(events.CommitCommentEventHandler); !ok {
		return errors.New("eventHandler does not implement events.CommitCommentEventHandler")
	}

	if _, ok := handler.(events.IssuesEventHandler); !ok {
		return errors.New("eventHandler does not implement events.IssuesEventHandler")
	}

	return nil
}
