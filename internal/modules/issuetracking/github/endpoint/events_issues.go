package endpoint

import (
	// Stdlib
	"net/http"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/github/events"
	httputil "github.com/salsaflow/salsaflow-daemon/internal/http"
	"github.com/salsaflow/salsaflow-daemon/internal/log"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/config"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/util"

	// Vendor
	"github.com/google/go-github/github"
)

// HandleIssuesEvent implements events.IssuesEventHandler
// and it is used to handle GitHub issues events.
func (handler *eventHandler) HandleIssuesEvent(
	rw http.ResponseWriter,
	r *http.Request,
	event *events.IssuesEvent,
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
	if !isStoryIssue(issue, config.Get()) {
		log.Info(r, "Issue %v is not a story issue, skipping", *issue.HTMLURL)
		return
	}

	switch *event.Action {
	case "closed":
		handler.onIssueClosed(rw, r, event, issue)
	case "reopened":
		handler.onIssueReopened(rw, r, event, issue)
	default:
		httputil.Status(rw, http.StatusAccepted)
	}
}

func (handler *eventHandler) onIssueClosed(
	rw http.ResponseWriter,
	r *http.Request,
	event *events.IssuesEvent,
	issue *github.Issue,
) {

	// When an issue is closed, we want to prune all SalsaFlow labels.
	var (
		owner = *event.Repo.Owner.Login
		repo  = *event.Repo.Name
	)
	err := util.ReplaceWorkflowLabels(handler.client, owner, repo, issue, nil, nil)
	if err != nil {
		httputil.Error(rw, r, err)
	} else {
		httputil.Status(rw, http.StatusAccepted)
	}
}

func (handler *eventHandler) onIssueReopened(
	rw http.ResponseWriter,
	r *http.Request,
	event *events.IssuesEvent,
	issue *github.Issue,
) {

	// When an issue is reopened, we want to move it into Being Implemented.
	var (
		owner  = *event.Repo.Owner.Login
		repo   = *event.Repo.Name
		labels = []string{config.Get().BeingImplementedLabel}
	)
	err := util.ReplaceWorkflowLabels(handler.client, owner, repo, issue, labels, nil)
	if err != nil {
		httputil.Error(rw, r, err)
	} else {
		httputil.Status(rw, http.StatusAccepted)
	}
}
