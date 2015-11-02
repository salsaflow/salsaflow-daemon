package github

import (
	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/errs"

	// Vendor
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func NewClient() (*github.Client, error) {
	token := GetConfig().Token
	if token == "" {
		return nil, &errs.ErrVarNotSet{"SFD_GITHUB_TOKEN"}
	}

	httpClient := oauth2.NewClient(oauth2.NoContext, &tokenSource{token})
	return github.NewClient(httpClient), nil
}

type tokenSource struct {
	token string
}

// Token implements oauth2.TokenSource interface.
func (ts *tokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: ts.token}, nil
}
