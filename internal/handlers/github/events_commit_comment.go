package github

import (
	// Stdlib
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/log"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/githubutils"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/httputils"

	// Vendor
	"github.com/google/go-github/github"
	"github.com/salsaflow/salsaflow/github/issues"
)

type commitCommentEvent struct {
	Comment    *github.RepositoryComment `json:"comment"`
	Repository *github.Repository        `json:"repository"`
}

func handleCommitComment(rw http.ResponseWriter, r *http.Request) {
	// Parse the payload.
	var event commitCommentEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Warn(r, "failed to parse event: %v", err)
		httputils.Status(rw, http.StatusBadRequest)
		return
	}

	// A command is always placed at the beginning of the line
	// and it is prefixed with '!'.
	cmdRegexp := regexp.MustCompile("^[!]([a-zA-Z]+) (.*)$")

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

	httputils.Status(rw, http.StatusAccepted)
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
	var (
		commitSHA     = *comment.CommitID
		commentURL    = *comment.HTMLURL
		commentAuthor = *comment.User.Login
	)

	issue, err := issues.FindReviewIssueByCommitItem(client, owner, repo, commitSHA)
	if err != nil {
		return err
	}

	// Parse issue body.
	reviewIssue, err := issues.ParseReviewIssue(issue)
	if err != nil {
		return err
	}

	// Add the new review issue record.
	reviewIssue.AddReviewBlocker(false, commentURL, commitSHA, blockerSummary)

	// Update the review issue.
	issueNum := *issue.Number
	_, _, err = client.Issues.Edit(owner, repo, issueNum, &github.IssueRequest{
		Body:  github.String(reviewIssue.FormatBody()),
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
