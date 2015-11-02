package endpoint

import (
	// Stdlib
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/github/events"
	httputil "github.com/salsaflow/salsaflow-daemon/internal/http"
	"github.com/salsaflow/salsaflow-daemon/internal/log"

	// Vendor
	"github.com/google/go-github/github"
	"github.com/salsaflow/salsaflow/github/issues"
)

// HandleCommitCommentEvent implements events.CommitCommentEventHandler
// and it is used to handle GitHub commit_comment events.
func (handler *eventHandler) HandleCommitCommentEvent(
	rw http.ResponseWriter,
	r *http.Request,
	event *events.CommitCommentEvent,
) {
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
		case "mustfix":
			err = handler.createReviewBlockerFromCommitComment(
				r,
				*event.Repo.Owner.Login,
				*event.Repo.Name,
				event.Comment,
				arg)
		}
		if err != nil {
			httputil.Error(rw, r, err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		httputil.Error(rw, r, err)
		return
	}

	httputil.Status(rw, http.StatusAccepted)
}

func (handler *eventHandler) createReviewBlockerFromCommitComment(
	r *http.Request,
	owner string,
	repo string,
	comment *github.RepositoryComment,
	blockerSummary string,
) error {

	// Find the right review issue.
	//
	// We search the content of all review issues for the right commit hash.
	var (
		client        = handler.client
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
	var bodyBuffer bytes.Buffer
	fmt.Fprintf(&bodyBuffer,
		"A new [review blocker](%v) was opened by @%v for review issue #%v. The summary follows:\n",
		commentURL, commentAuthor, issueNum)
	fmt.Fprintf(&bodyBuffer, "> %v\n", blockerSummary)

	_, _, err = client.Issues.CreateComment(owner, repo, issueNum, &github.IssueComment{
		Body: github.String(bodyBuffer.String()),
	})
	return err
}
