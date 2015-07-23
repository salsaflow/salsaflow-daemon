package github

import (
	// Stdlib
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/log"
	"github.com/salsaflow/salsaflow-daemon/internal/trackers"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/githubutils"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/httputils"

	// Vendor
	"github.com/codegangsta/negroni"
	"github.com/google/go-github/github"
	ghissues "github.com/salsaflow/salsaflow/github/issues"
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
	case "commit_comment":
		handleCommitCommentEvent(rw, r)
	case "issues":
		handleIssuesEvent(rw, r)
	case "push":
		handlePushEvent(rw, r)
	default:
		httpStatus(rw, http.StatusAccepted)
	}
}
