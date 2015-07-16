package github

import (
	// Stdlib
	"bufio"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/log"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/githubutils"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/httputils"

	// Vendor
	"github.com/google/go-github/github"
	"github.com/salsaflow/salsaflow/github/issues"
)

func handlePushEvent(rw http.ResponseWriter, r *http.Request) {
	// Parse the payload.
	var event github.PushEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Warn(r, "failed to parse event: %v", err)
		httpStatus(rw, http.StatusBadRequest)
		return
	}

	// Go through the commits and search for SalsaFlow commands.
	for _, commit := range event.Commits {
		scanner := bufio.NewScanner(strings.NewReader(*commit.Message))
		scanner.Split(bufio.ScanWords)
	Scanning:
		for scanner.Scan() {
			// word is never and empty string.
			word := scanner.Text()
			// Every command starts with '!'.
			if word[0] != '!' {
				continue
			}
			cmd := word[1:]
			switch cmd {
			case "unblock":
				// Get the argument.
				if !scanner.Scan() {
					log.Warn(r, "EOF encountered while getting !unblock argument")
					break Scanning
				}
				blockerNum, err := strconv.Atoi(scanner.Text())
				if err != nil {
					log.Warn(r, "!unblock: argument is not a number: %v", err)
					continue Scanning
				}

				// Mark the relevant review blocker as unblocked.
				owner, repo := *event.Repo.Owner.Login, *event.Repo.Name
				updated, err := unblockReviewBlocker(owner, repo, &commit, blockerNum)
				if err != nil {
					log.Error(r, err)
					continue
				}
				if !updated {
					log.Warn(r, "!unblock: unknown blocker number [commit=%v, blocker=%v]",
						*commit.SHA, blockerNum)
					continue
				}

				// Add a comment to the review issue.
				err = addUnblockComment(owner, repo, &commit, blockerNum)
				if err != nil {
					log.Error(r, err)
					continue
				}
			}
		}
		if err := scanner.Err(); err != nil {
			httputils.Error(rw, r, err)
			return
		}
	}

	httpStatus(rw, http.StatusAccepted)
}

// unblockReviewBlocker returns
//
//   true,  nil - success
//   false, nil - unknown review blocker number
//   false, err - something exploded
func unblockReviewBlocker(
	owner string,
	repo string,
	commit *github.PushEventCommit,
	blockerNum int,
) (updated bool, err error) {

	// Get GitHub API client.
	client, err := githubutils.NewClient()
	if err != nil {
		return false, err
	}

	// Find the review assue where the commit is registered.
	issue, err := issues.FindReviewIssueForCommit(client, owner, repo, *commit.SHA)
	if err != nil {
		return false, err
	}

	// Find the relevant blocker item and mark it as unblocked.
	blockers := issue.ReviewBlockers()
	if len(blockers) < blockerNum {
		// Invalid blocker number.
		return false, nil
	}
	blocker = blockers[blockerNum-1]
	blocker.Fixed = true

	_, _, err = client.Issues.Edit(owner, repo, *issue.Number, &github.IssueRequest{
		Body: github.String(issue.FormatBody()),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func addUnblockComment(
	owner string,
	repo string,
	commit *github.PushEventCommit,
	blockerNum int,
) error {

	// Add a comment to the review issue.
	commentBody := fmt.Sprintf(
		"Review blocker [[%v]](%v) was unblocked by commit %v (authored by %v).",
		blockerNum, blocker.CommentURL, *commit.SHA, *commit.Author)

	_, _, err := client.Issues.CreateComment(owner, repo, *issue.Number, &github.IssueComment{
		Body: github.Strng(commentBody),
	})
	return err
}
