package github

import (
	// Stdlib
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/log"
	"github.com/salsaflow/salsaflow-daemon/internal/trackers"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/githubutils"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/httputils"

	// Vendor
	"github.com/google/go-github/github"
	"github.com/salsaflow/salsaflow/github/issues"
)

func handleIssuesEvent(rw http.ResponseWriter, r *http.Request) {
	// Parse the payload.
	var event github.IssueActivityEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Warn(r, "failed to parse event: %v", err)
		httputils.Status(rw, http.StatusBadRequest)
		return
	}

	// Make sure this is a review issue event.
	// The label is sometimes missing in the webhook, we need to re-fetch.
	client, err := githubutils.NewClient()
	if err != nil {
		httputils.Error(rw, r, err)
		return
	}
	var (
		owner    = *event.Repo.Owner.Login
		repo     = *event.Repo.Name
		issueNum = *event.Issue.Number
	)
	log.Info(r, "Re-fetching issue %v/%v#%v", owner, repo, issueNum)
	issue, _, err := client.Issues.Get(owner, repo, issueNum)
	if err != nil {
		httputils.Error(rw, r, err)
		return
	}

	// Make sure this is a review issue.
	labeledWith := func(label string) bool {
		for _, labelPtr := range issue.Labels {
			if *labelPtr.Name == label {
				return true
			}
		}
		return false
	}

	if !labeledWith("review") {
		log.Info(r, "Issue %s is not a review issue", *issue.HTMLURL)
		httputils.Status(rw, http.StatusAccepted)
		return
	}

	// Do nothing unless this is an opened, closed or reopened event.
	switch *event.Action {
	case "opened":
	case "closed":
		// Make sure the issue is marked as implemented.
		if !labeledWith("implemented") {
			rejectClose(rw, r, client, &event)
			return
		}

	case "reopened":
	default:
		httputils.Status(rw, http.StatusAccepted)
		return
	}

	// Parse issue body.
	reviewIssue, err := issues.ParseReviewIssue(issue)
	if err != nil {
		log.Error(r, err)
		httputils.Status(rw, httputils.StatusUnprocessableEntity)
		return
	}

	// We are done in case this is a commit review issue.
	storyIssue, ok := reviewIssue.(*issues.StoryReviewIssue)
	if !ok {
		httputils.Status(rw, http.StatusAccepted)
		return
	}

	// Instantiate the issue tracker.
	tracker, err := trackers.GetIssueTracker(storyIssue.TrackerName)
	if err != nil {
		log.Error(r, err)
		httputils.Status(rw, httputils.StatusUnprocessableEntity)
		return
	}

	// Find relevant story.
	story, err := tracker.FindStoryByTag(storyIssue.StoryKey)
	if err != nil {
		log.Error(r, err)
		httputils.Status(rw, httputils.StatusUnprocessableEntity)
		return
	}

	// Invoke relevant event handler.
	var (
		issueNumString = strconv.Itoa(*issue.Number)
		issueURL       = *issue.HTMLURL
		ex             error
	)
	switch *event.Action {
	case "opened":
		ex = story.OnReviewRequestOpened(issueNumString, issueURL)
	case "closed":
		ex = story.OnReviewRequestClosed(issueNumString, issueURL)
	case "reopened":
		ex = story.OnReviewRequestReopened(issueNumString, issueURL)
	default:
		panic("unreachable code reached")
	}
	if ex != nil {
		httputils.Error(rw, r, ex)
		return
	}

	if *event.Action == "closed" {
		if err := story.MarkAsReviewed(); err != nil {
			httputils.Error(rw, r, err)
			return
		}
	}

	httputils.Status(rw, http.StatusAccepted)
}

func rejectClose(
	rw http.ResponseWriter,
	r *http.Request,
	client *github.Client,
	event *github.IssueActivityEvent) {

	var (
		owner    = *event.Repo.Owner.Login
		repo     = *event.Repo.Name
		issueNum = *event.Issue.Number
		sender   = *event.Sender.Login
	)

	// Log stuff.
	log.Info(r, "Reopening review issue %v/%v#%v, not implemented yet", owner, repo, issueNum)

	// Re-open the issue.
	_, _, err := client.Issues.Edit(owner, repo, issueNum, &github.IssueRequest{
		State: github.String("open"),
	})
	if err != nil {
		httputils.Error(rw, r, err)
		return
	}

	// Add a comment to notify the sender.
	var body bytes.Buffer
	fmt.Fprintf(&body,
		"@%v Reopening review issue #%v, the associated story is not implemented yet.\n",
		sender, issueNum)
	fmt.Fprintln(&body,
		"The review issue needs to be labeled with `implemented`, then it can be closed.")

	_, _, err = client.Issues.CreateComment(owner, repo, issueNum, &github.IssueComment{
		Body: github.String(body.String()),
	})
	if err != nil {
		httputils.Error(rw, r, err)
		return
	}

	httputils.Status(rw, http.StatusAccepted)
}
