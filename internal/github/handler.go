package github

import (
	// Stdlib
	"crypto/hmac"
	"crypto/sha1"
	"io"
	"log"
	"net/http"

	// Vendor
	"github.com/codegangsta/negroni"
	"github.com/google/go-github/github"
)

type Handler struct {
	// Embedded http.Handler
	http.Handler

	// Options
	secret string
}

type OptionFunc func(handler *Handler)

func SetSecret(secret string) OptionFunc {
	return func(handler *Handler) {
		handler.secret = secret
	}
}

func NewHandler(options ...OptionFunc) http.Handler {
	// Create the handler.
	handler := &Handler{}
	for _, opt := range options {
		opt(handler)
	}

	// Set up the middleware chain.
	n := negroni.New()
	if handler.secret != "" {
		n.Use(newSecretMiddleware(handler.secret))
	}
	n.UseHandlerFunc(handler.handleEvent)

	// Set the Negroni instance to be THE handler.
	handler.Handler = n

	// Return the new handler.
	return handler
}

type IssueEvent struct {
	Action string        `json:"action"`
	Issue  *github.Issue `json:"issue"`
	Label  *github.Label `json:"label,omitempty"`
}

func (handler *Handler) handleEvent(rw http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-Github-Event")
	switch eventType {
	case "issues":
		handler.handleIssuesEvent(rw, r)
	default:
		httpStatus(rw, http.StatusAccepted)
	}
}

func (handler *Handler) handleIssuesEvent(rw http.ResponseWriter, r *http.Request) {

}

func newSecretMiddleware(secret string) negroni.HandlerFunc {
	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Compute the hash.
			mac := hmac.New(sha1.New, []byte(secret))
			if _, err := io.Copy(mac, r.Body); err != nil {
				log.Printf("ERROR in %v: %v\n", r.URL.Path, err)
				httpStatus(rw, http.StatusInternalServerError)
				return
			}

			// Compare with the header provided in the request.
			secretHeader := r.Header.Get("X-Hub-Signature")
			expected := mac.Sum(nil)
			if !hmac.Equal(expected, []byte(secretHeader)) {
				log.Printf("WARNING in %v: HMAC mismatch detected: expected=%v, got=%v",
					r.URL.Path, string(expected), secretHeader)
				httpStatus(rw, http.StatusUnauthorized)
				return
			}

			// Call the next handler.
			next(rw, r)
		})
}

func httpStatus(rw http.ResponseWriter, status int) {
	http.Error(rw, http.StatusText(status), status)
}
