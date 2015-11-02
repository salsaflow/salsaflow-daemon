package endpoint

import (
	// Stdlib
	"net/http"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/github"
	module "github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/github"
)

type Endpoint struct{}

func NewEndpoint() *Endpoint {
	return &Endpoint{}
}

func (ep *Endpoint) ModuleId() string {
	return module.ModuleId
}

func (ep *Endpoint) NewHandler() (http.Handler, error) {
	client, err := github.NewClient()
	if err != nil {
		return nil, err
	}

	handler := github.NewWebhookHandler(&eventHandler{client})

	mux := http.NewServeMux()
	mux.Handle("/events", handler)
	return mux, nil
}
