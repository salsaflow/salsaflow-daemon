package endpoints

import (
	// Stdlib
	"net/http"

	// Internal
	ghReview "github.com/salsaflow/salsaflow-daemon/internal/modules/codereview/github/endpoint"
	ghIssues "github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github/endpoint"
	pt "github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/endpoint"
)

type ModuleEndpoint interface {
	ModuleId() string
	NewHandler() (http.Handler, error)
}

var endpoints = []ModuleEndpoint{
	ghReview.NewEndpoint(),
	ghIssues.NewEndpoint(),
	pt.NewEndpoint(),
}

func Endpoints() []ModuleEndpoint {
	eps := make([]ModuleEndpoint, len(endpoints))
	copy(eps, endpoints)
	return eps
}
