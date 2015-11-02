package endpoint

import (
	// Stdlib
	"bufio"
	"net/http"
	"strings"

	// Internal
	githubutil "github.com/salsaflow/salsaflow-daemon/internal/github"
	"github.com/salsaflow/salsaflow-daemon/internal/github/events"
	httputil "github.com/salsaflow/salsaflow-daemon/internal/http"
	"github.com/salsaflow/salsaflow-daemon/internal/log"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/config"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/util"

	// Vendor
	"github.com/google/go-github/github"
)

// HandleIssueCommentEvent implements events.IssueCommentEventHandler
// and it is used to handle GitHub issue_comment events.
func (handler *eventHandler) HandleIssueCommentEvent(
	rw http.ResponseWriter,
	r *http.Request,
	event *events.IssueCommentEvent,
) {
	// Make sure this is a story issue event.
	// The label is sometimes missing in the webhook, we need to re-fetch.
	var (
		owner    = *event.Repo.Owner.Login
		repo     = *event.Repo.Name
		issueNum = *event.Issue.Number
	)
	log.Info(r, "Re-fetching issue %v/%v#%v", owner, repo, issueNum)
	issue, _, err := handler.client.Issues.Get(owner, repo, issueNum)
	if err != nil {
		httputil.Error(rw, r, err)
		return
	}

	// Make sure this is a story issue.
	if !githubutil.LabeledWith(issue, config.Get().StoryLabel) {
		log.Info(r, "Issue %v is not a story issue, skipping", *issue.HTMLURL)
		return
	}

	switch *event.Action {
	case "created":
		handler.onIssueCommentCreated(rw, r, event, issue)
	default:
		httputil.Status(rw, http.StatusAccepted)
	}
}

func (handler *eventHandler) onIssueCommentCreated(
	rw http.ResponseWriter,
	r *http.Request,
	event *events.IssueCommentEvent,
	issue *github.Issue,
) {

	scanner := bufio.NewScanner(strings.NewReader(*event.Comment.Body))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		switch word {
		case "!reject":
			if err := handler.rejectIssue(r, event, issue); err != nil {
				httputil.Error(rw, r, err)
				return
			}
		}
	}
	if err := scanner.Err(); err != nil {
		httputil.Error(rw, r, err)
		return
	}

	httputil.Status(rw, http.StatusAccepted)
}

func (handler *eventHandler) rejectIssue(
	r *http.Request,
	event *events.IssueCommentEvent,
	issue *github.Issue,
) error {

	// Mark the issue as rejected.
	var (
		owner  = *event.Repo.Owner.Login
		repo   = *event.Repo.Name
		labels = []string{config.Get().RejectedLabel}
	)
	return util.ReplaceWorkflowLabels(handler.client, owner, repo, issue, labels)
}
