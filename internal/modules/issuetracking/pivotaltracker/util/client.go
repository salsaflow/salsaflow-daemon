package util

import (
	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/errs"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/config"

	// Vendor
	"gopkg.in/salsita/go-pivotaltracker.v1/v5/pivotal"
)

// NewClient returns a new Pivotal Tracker API client
// that uses the access token read from the environment.
//
// An error is returned in case the relevant environment variable is not set.
func NewClient() (*pivotal.Client, error) {
	token := config.Get().Token
	if token == "" {
		return nil, &errs.ErrVarNotSet{"SFD_PIVOTALTRACKER_TOKEN"}
	}

	return pivotal.NewClient(token), nil
}
