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

	// Set up routing.
	mux := http.NewServeMux()
	mux.HandleFunc("/issues", handler.handleIssueEvent)

	// Set up the middleware chain.
	n := negroni.New()
	if handler.secret != "" {
		n.Use(newSecretMiddleware(handler.secret))
	}
	n.UseHandler(mux)

	// Set the Negroni instance to be THE handler.
	handler.Handler = n

	// Return the new handler.
	return handler
}

func (handler *Handler) handleIssueEvent(rw http.ResponseWriter, r *http.Request) {

}

func newSecretMiddleware(secret string) negroni.HandlerFunc {
	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Compute the hash.
			mac := hmac.New(sha1.New, []byte(secret))
			if _, err := io.Copy(mac, r.Body); err != nil {
				log.Printf("ERROR in %v: %v\n", r.URL.Path, err)
				http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Compare with the header provided in the request.
			secretHeader := r.Header.Get("X-Hub-Signature")
			expected := mac.Sum(nil)
			if !hmac.Equal(expected, []byte(secretHeader)) {
				log.Printf("WARNING in %v: HMAC mismatch detected: expected=%v, got=%v",
					r.URL.Path, string(expected), secretHeader)
				http.Error(rw, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Call the next handler.
			next(rw, r)
		})
}
