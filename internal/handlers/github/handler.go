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
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	// Internal
	"github.com/tchap/salsaflow-daemon/internal/trackers"

	// Vendor
	"github.com/codegangsta/negroni"
	"github.com/google/go-github/github"
)

const (
	statusUnprocessableEntity     = 422
	statusUnprocessableEntityText = "Unprocessable Entity"
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
	// Parse the payload.
	var event github.IssueActivityEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("WARNING in %v: failed to parse event: %v\n", r.URL.Path, err)
		httpStatus(rw, http.StatusBadRequest)
		return
	}
	issue := event.Issue

	// Make sure this is a review issue event.
	var isReviewIssue bool
	for _, label := range issue.Labels {
		if *label.Name == "review" {
			isReviewIssue = true
			break
		}
	}
	if !isReviewIssue {
		httpStatus(rw, http.StatusAccepted)
		return
	}

	// Do nothing unless this is an opened, closed or reopened event.
	switch *event.Action {
	case "opened":
	case "closed":
	case "reopened":
	default:
		httpStatus(rw, http.StatusAccepted)
		return
	}

	// Parse issue body.
	body, err := parseIssueBody(*issue.Body)
	if err != nil {
		log.Printf("WARNING in %v: failed to parse issue body: %v\n", r.URL.Path, err)
		httpStatus(rw, statusUnprocessableEntity)
		return
	}

	log.Printf("Issue being handled: %#v\n", body)

	// Instantiate the issue tracker.
	tracker, err := trackers.GetIssueTracker(body.TrackerId)
	if err != nil {
		log.Printf("WARNING in %v: %v\n", r.URL.Path, err)
		httpStatus(rw, statusUnprocessableEntity)
		return
	}

	// Find relevant story.
	story, err := tracker.FindStoryById(body.ProjectId, body.StoryId)
	if err != nil {
		log.Printf("WARNING in %v: story not found: %v\n", r.URL.Path, err)
		httpStatus(rw, statusUnprocessableEntity)
		return
	}

	// Invoke relevant event handler.
	var (
		issueNum = strconv.Itoa(*issue.Number)
		issueURL = *issue.HTMLURL
		ex       error
	)
	switch *event.Action {
	case "opened":
		ex = story.OnReviewRequestOpened(issueNum, issueURL)
	case "closed":
		ex = story.OnReviewRequestClosed(issueNum, issueURL)
	case "reopened":
		ex = story.OnReviewRequestReopened(issueNum, issueURL)
	default:
		panic("unreachable code reached")
	}
	if ex != nil {
		log.Printf("WARNING in %v: event handler failed: %v\n", r.URL.Path, ex)
		httpStatus(rw, http.StatusInternalServerError)
		return
	}
	httpStatus(rw, http.StatusAccepted)
}

func newSecretMiddleware(secret string) negroni.HandlerFunc {
	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Read the request body into a buffer.
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("ERROR in %v: %v\n", r.URL.Path, err)
				httpStatus(rw, http.StatusInternalServerError)
				return
			}

			// Fill the request body again so that it is available in the next handler.
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			// Compute the hash.
			mac := hmac.New(sha1.New, []byte(secret))
			if _, err := io.Copy(mac, bytes.NewReader(bodyBytes)); err != nil {
				log.Printf("ERROR in %v: %v\n", r.URL.Path, err)
				httpStatus(rw, http.StatusInternalServerError)
				return
			}

			// Compare with the header provided in the request.
			secretHeader := r.Header.Get("X-Hub-Signature")
			expected := "sha1=" + hex.EncodeToString(mac.Sum(nil))
			if secretHeader != expected {
				log.Printf("WARNING in %v: HMAC mismatch detected: expected='%v', got='%v'\n",
					r.URL.Path, expected, secretHeader)
				httpStatus(rw, http.StatusUnauthorized)
				return
			}

			// Call the next handler.
			next(rw, r)
		})
}

const (
	trackerIdTag = "SF-Issue-Tracker"
	projectIdTag = "SF-Project-Id"
	storyIdTag   = "SF-Story-Id"
)

var (
	trackerIdRegexp = regexp.MustCompile(fmt.Sprintf("^%v: (.+)$", trackerIdTag))
	projectIdRegexp = regexp.MustCompile(fmt.Sprintf("^%v: (.+)$", projectIdTag))
	storyIdRegexp   = regexp.MustCompile(fmt.Sprintf("^%v: (.+)$", storyIdTag))
)

type issueBody struct {
	TrackerId string
	ProjectId string
	StoryId   string
}

func parseIssueBody(content string) (*issueBody, error) {
	var body issueBody
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		match := trackerIdRegexp.FindStringSubmatch(line)
		if len(match) == 2 {
			body.TrackerId = match[1]
			continue
		}

		match = projectIdRegexp.FindStringSubmatch(line)
		if len(match) == 2 {
			body.ProjectId = match[1]
			continue
		}

		match = storyIdRegexp.FindStringSubmatch(line)
		if len(match) == 2 {
			body.StoryId = match[1]
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	switch {
	case body.TrackerId == "":
		return nil, fmt.Errorf("parseIssueBody: %v tag not found", trackerIdTag)
	case body.ProjectId == "":
		return nil, fmt.Errorf("parseIssueBody: %v tag not found", projectIdTag)
	case body.StoryId == "":
		return nil, fmt.Errorf("parseIssueBody: %v tag not found", storyIdTag)
	}

	return &body, nil
}

func httpStatus(rw http.ResponseWriter, status int) {
	switch status {
	case statusUnprocessableEntity:
		http.Error(rw, statusUnprocessableEntityText, statusUnprocessableEntity)
	default:
		http.Error(rw, http.StatusText(status), status)
	}
}
