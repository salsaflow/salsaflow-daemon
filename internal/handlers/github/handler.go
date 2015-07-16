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
	case "commit_comment":
		handler.handleCommitComment(rw, r)
	case "issues":
		handler.handleIssuesEvent(rw, r)
	case "push":
		handlePushEvent(rw, r)
	default:
		httpStatus(rw, http.StatusAccepted)
	}
}

type commitCommentEvent struct {
	Comment    *github.RepositoryComment `json:"comment"`
	Repository *github.Repository        `json:"repository"`
}

func (handler *Handler) handleCommitComment(rw http.ResponseWriter, r *http.Request) {
	// Parse the payload.
	var event commitCommentEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Warn(r, "failed to parse event: %v", err)
		httpStatus(rw, http.StatusBadRequest)
		return
	}

	// A command is always placed at the beginning of the line
	// and it is prefixed with '!'.
	cmdRegexp := regexp.MustCompile("^[!]([a-zA-Z]+)(.*)$")

	// Process the comment body.
	scanner := bufio.NewScanner(strings.NewReader(*event.Comment.Body))
	for scanner.Scan() {
		// Check whether this is a command and continue if not.
		match := cmdRegexp.FindStringSubmatch(scanner.Text())
		if len(match) == 0 {
			continue
		}
		cmd, arg := match[1], strings.TrimSpace(match[2])

		var err error
		switch cmd {
		case "blocker":
			err = createReviewBlockerFromCommitComment(
				r,
				*event.Repository.Owner.Login,
				*event.Repository.Name,
				event.Comment,
				arg)
		}
		if err != nil {
			httputils.Error(rw, r, err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		httputils.Error(rw, r, err)
		return
	}

	httpStatus(rw, http.StatusAccepted)
}

func createReviewBlockerFromCommitComment(
	r *http.Request,
	owner string,
	repo string,
	comment *github.RepositoryComment,
	blockerSummary string,
) error {

	// Get GitHub API client.
	client, err := githubutils.NewClient()
	if err != nil {
		return err
	}

	// Find the right review issue.
	//
	// We search the content of all review issues for the right commit hash.
	// This is not terribly robust but that is all we can do right now.
	//
	// GitHub shortens commit hashes to 7 leading characters, hence [:7].
	var (
		commitSHA     = *comment.CommitID
		commentURL    = *comment.HTMLURL
		commentAuthor = *comment.User.Login
		pattern       = fmt.Sprintf("] %v:", commitSHA[:7])
	)

	query := fmt.Sprintf(
		`"%v" repo:"%v/%v" type:issue state:open state:closed label:review in:body`,
		pattern, owner, repo)

	res, _, err := client.Search.Issues(query, &github.SearchOptions{})
	if err != nil {
		return err
	}
	if num := *res.Total; num != 1 {
		log.Warn(r, "failed to find the review issue for commit %v (%v issues found)",
			commitSHA, num)
		return nil
	}
	issue := res.Issues[0]

	// Parse issue body.
	issueCtx, err := ghissues.ParseReviewIssue(&issue)
	if err != nil {
		return err
	}

	// Add the new review issue record.
	issueCtx.AddReviewBlocker(commitSHA, commentURL, blockerSummary, false)

	// Update the review issue.
	issueNum := *issue.Number
	_, _, err = client.Issues.Edit(owner, repo, issueNum, &github.IssueRequest{
		Body:  github.String(issueCtx.FormatBody()),
		State: github.String("open"),
	})
	if err != nil {
		return err
	}

	log.Info(r, "Linked a new review comment to review issue %v/%v#%v", owner, repo, issueNum)

	// Add the blocker comment.
	body := fmt.Sprintf("A new [review blocker](%v) was opened by @%v for review issue #%v.",
		commentURL, commentAuthor, issueNum)

	_, _, err = client.Issues.CreateComment(owner, repo, issueNum, &github.IssueComment{
		Body: github.String(body),
	})
	return err
}

func (handler *Handler) handleIssuesEvent(rw http.ResponseWriter, r *http.Request) {
	// Parse the payload.
	var event github.IssueActivityEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Warn(r, "failed to parse event: %v", err)
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
	issueCtx, err := ghissues.ParseReviewIssue(issue)
	if err != nil {
		log.Error(r, err)
		httpStatus(rw, statusUnprocessableEntity)
		return
	}

	// We are done in case this is a commit review issue.
	ctx, ok := issueCtx.(*ghissues.StoryReviewIssue)
	if !ok {
		httpStatus(rw, http.StatusAccepted)
		return
	}

	// Instantiate the issue tracker.
	tracker, err := trackers.GetIssueTracker(ctx.TrackerName)
	if err != nil {
		log.Error(r, err)
		httpStatus(rw, statusUnprocessableEntity)
		return
	}

	// Find relevant story.
	story, err := tracker.FindStoryByTag(ctx.StoryKey)
	if err != nil {
		log.Error(r, err)
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
		httputils.Error(rw, r, err)
		return
	}

	if *event.Action == "closed" {
		if err := story.MarkAsReviewed(); err != nil {
			httputils.Error(rw, r, err)
			return
		}
	}

	httpStatus(rw, http.StatusAccepted)
}

func newSecretMiddleware(secret string) negroni.HandlerFunc {
	return negroni.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			// Read the request body into a buffer.
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				httputils.Error(rw, r, err)
				return
			}

			// Fill the request body again so that it is available in the next handler.
			r.Body.Close()
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			// Compute the hash.
			mac := hmac.New(sha1.New, []byte(secret))
			if _, err := io.Copy(mac, bytes.NewReader(bodyBytes)); err != nil {
				httputils.Error(rw, r, err)
				return
			}

			// Compare with the header provided in the request.
			secretHeader := r.Header.Get("X-Hub-Signature")
			expected := "sha1=" + hex.EncodeToString(mac.Sum(nil))
			if secretHeader != expected {
				log.Warn(r, "HMAC mismatch detected: expected='%v', got='%v'\n",
					expected, secretHeader)
				httpStatus(rw, http.StatusUnauthorized)
				return
			}

			// Call the next handler.
			next(rw, r)
		})
}

func httpStatus(rw http.ResponseWriter, status int) {
	switch status {
	case statusUnprocessableEntity:
		http.Error(rw, statusUnprocessableEntityText, statusUnprocessableEntity)
	default:
		http.Error(rw, http.StatusText(status), status)
	}
}
