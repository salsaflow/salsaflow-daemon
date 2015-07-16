package github

import (
	// Stdlib
	"bufio"
	"encoding/json"
	"fmt"
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
				owner, repo := *event.Repo.Owner.Name, *event.Repo.Name
				issue, blocker, err := unblockReviewBlocker(owner, repo, &commit, blockerNum)
				if err != nil {
					log.Error(r, err)
					continue
				}
				if issue == nil {
					log.Warn(r, "!unblock: unknown blocker number [commit=%v, blocker=%v]",
						*commit.SHA, blockerNum)
					continue
				}

				// Add a comment to the review issue.
				err = addUnblockComment(owner, repo, *issue.Number, blocker, &commit)
				if err != nil {
					log.Error(r, err)
					continue
				}

				log.Info(r, "Review blocker %v for issue %v marked as unblocked",
					blocker.BlockerNumber, *issue.HTMLURL)
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
//   issue, blocker, nil - success
//   nil,   nil,     nil - unknown review blocker number
//   nil,   nil,     err - something exploded
func unblockReviewBlocker(
	owner string,
	repo string,
	commit *github.PushEventCommit,
	blockerNum int,
) (*github.Issue, *issues.ReviewBlockerItem, error) {

	// Get GitHub API client.
	client, err := githubutils.NewClient()
	if err != nil {
		return nil, nil, err
	}

	// Find the review assue where the commit is registered.
	commitTitle, err := bufio.NewReader(strings.NewReader(*commit.Message)).ReadString('\n')
	if err != nil {
		return nil, nil, err
	}

	fmt.Printlf("COMMIT %+v\n", commit)

	issue, err := issues.FindIssueForCommitItem(client, owner, repo, *commit.SHA, commitTitle)
	if err != nil {
		return nil, nil, err
	}

	reviewIssue, err := issues.ParseReviewIssue(issue)
	if err != nil {
		return nil, nil, err
	}

	// Find the relevant blocker item and mark it as unblocked.
	blockers := reviewIssue.ReviewBlockerItems()
	if len(blockers) < blockerNum {
		// Unknown blocker number.
		return nil, nil, nil
	}
	blocker := blockers[blockerNum-1]
	blocker.Fixed = true

	newIssue, _, err := client.Issues.Edit(owner, repo, *issue.Number, &github.IssueRequest{
		Body: github.String(reviewIssue.FormatBody()),
	})
	if err != nil {
		return nil, nil, err
	}
	return newIssue, blocker, nil
}

func addUnblockComment(
	owner string,
	repo string,
	issueNum int,
	blocker *issues.ReviewBlockerItem,
	commit *github.PushEventCommit,
) error {

	// Get GitHub API client.
	client, err := githubutils.NewClient()
	if err != nil {
		return err
	}

	// Add a comment to the review issue.
	commentBody := fmt.Sprintf(
		"Review blocker [[%v]](%v) was unblocked by commit %v (authored by %v).",
		blocker.BlockerNumber, blocker.CommentURL, *commit.SHA, *commit.Author)

	_, _, err = client.Issues.CreateComment(owner, repo, issueNum, &github.IssueComment{
		Body: github.String(commentBody),
	})
	return err
}
