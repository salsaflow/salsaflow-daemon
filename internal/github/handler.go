package github

import (
	// Stdlib
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	var event github.IssueActivityEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		httpStatus(rw, http.StatusBadRequest)
	}

	fmt.Println("Accepted issues event for", *event.Issue.HTMLURL)
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
			expected := "sha1=" + hex.EncodeToString(mac.Sum(nil))
			if secretHeader != expected {
				log.Printf("WARNING in %v: HMAC mismatch detected: expected='%v', got='%v'",
					r.URL.Path, expected, secretHeader)
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
