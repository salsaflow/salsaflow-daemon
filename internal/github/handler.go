package github

import (
	// Stdlib
	"encoding/json"
	stdLog "log"
	"net/http"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/github/events"
	httputil "github.com/salsaflow/salsaflow-daemon/internal/http"

	// Vendor
	"github.com/codegangsta/negroni"
)

// WebhookHandler --------------------------------------------------------------

// WebhookHandler provides a framework for handling GitHub webhooks. It implements
// http.Handler so that it can be used anywhere http.Handler is expected.
//
// Now what is a WebhookHandler good for. All it does is that it takes an event
// handler object and it routes incoming webhooks to the right methods, e.g.
// when an issues_event webhook is received, it passes the request into
// HandleIssuesEvent method. All available event handling methods can be found
// in the events package.
//
// In case the event handler does not implement the method for the event type
// received, WebhookHandler simply returns 201 Accepted and does nothing.
type WebhookHandler struct {
	// Embedded http.Handler
	http.Handler

	// The event handler being used for this webhook handler.
	eventHandler interface{}
}

func NewWebhookHandler(eventHandler interface{}) *WebhookHandler {
	// Create the handler.
	handler := &WebhookHandler{
		eventHandler: eventHandler,
	}

	// Set up the middleware chain.
	n := negroni.New()

	if secret := GetConfig().WebhookSecret; secret != "" {
		n.Use(newSecretMiddleware(secret))
	} else {
		stdLog.Println("WARNING: SFD_GITHUB_WEBHOOK_SECRET is not set")
	}

	n.UseHandlerFunc(handler.handleEvent)

	// Set the Negroni instance to be THE handler.
	handler.Handler = n

	// Return the new handler.
	return handler
}

func (handler *WebhookHandler) handleEvent(rw http.ResponseWriter, r *http.Request) {
	// Get the right event handler and execute it.
	getEventHandler(r.Header.Get("X-GitHub-Event"), handler.eventHandler).ServeHTTP(rw, r)
}

// Event handlers --------------------------------------------------------------

type eventHandlerSpec struct {
	isHandlerCompatible func(eventHandler interface{}) bool
	newHandler          func(eventHandler interface{}) http.Handler
}

var specs = map[string]*eventHandlerSpec{
	"commit_comment": &eventHandlerSpec{
		func(eventHandler interface{}) bool {
			_, ok := eventHandler.(events.CommitCommentEventHandler)
			return ok
		},
		func(eventHandler interface{}) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				var event events.CommitCommentEvent
				if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
					httputil.Error(rw, r, err)
					return
				}

				eventHandler.(events.CommitCommentEventHandler).HandleCommitCommentEvent(rw, r, &event)
			})
		},
	},
	"issue_comment": &eventHandlerSpec{
		func(eventHandler interface{}) bool {
			_, ok := eventHandler.(events.IssueCommentEventHandler)
			return ok
		},
		func(eventHandler interface{}) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				var event events.IssueCommentEvent
				if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
					httputil.Error(rw, r, err)
					return
				}

				eventHandler.(events.IssueCommentEventHandler).HandleIssueCommentEvent(rw, r, &event)
			})
		},
	},
	"issues": &eventHandlerSpec{
		func(eventHandler interface{}) bool {
			_, ok := eventHandler.(events.IssuesEventHandler)
			return ok
		},
		func(eventHandler interface{}) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				var event events.IssuesEvent
				if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
					httputil.Error(rw, r, err)
					return
				}

				eventHandler.(events.IssuesEventHandler).HandleIssuesEvent(rw, r, &event)
			})
		},
	},
}

func getEventHandler(eventType string, eventHandler interface{}) http.Handler {
	// Get the spec for the given event type.
	spec, ok := specs[eventType]
	if !ok {
		return http.HandlerFunc(accepted)
	}

	// Check whether eventHandler implements the right interface.
	// In case this is not the case, we simply return 201 Accepted.
	if !spec.isHandlerCompatible(eventHandler) {
		return http.HandlerFunc(accepted)
	}

	// In case eventHandler implements the right interface,
	// we use eventHandler to handle the request.
	return spec.newHandler(eventHandler)
}

func accepted(rw http.ResponseWriter, r *http.Request) {
	httputil.Status(rw, http.StatusAccepted)
}
