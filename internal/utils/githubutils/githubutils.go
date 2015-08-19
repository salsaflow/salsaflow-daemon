package githubutils

import (
	// Stdlib
	"os"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/env"

	// Vendor
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const EnvAccessToken = "SFD_GH_TOKEN"

func NewClient() (client *github.Client, err error) {
	token := os.Getenv(EnvAccessToken)
	if token == "" {
		return nil, &env.ErrNotSet{EnvAccessToken}
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
