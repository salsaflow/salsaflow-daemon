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
	"github.com/salsaflow/salsaflow-daemon/internal/utils/httputils"

	// Vendor
	"github.com/google/go-github/github"
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
					log.Warn(r, "!unblock argument is not a number: %v", err)
					continue Scanning
				}

				// Mark the relevant review blocker as unblocked.
				owner, repo := *event.Repo.Owner.Login, *event.Repo.Name
				if err := unblockReviewIssue(owner, repo, &commit, blockerNum); err != nil {
					log.Error(r, err)
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

func unblockReviewIssue(owner, repo string, commit *github.PushEventCommit, blockerNum int) error {
	panic("Not implemented")
}
