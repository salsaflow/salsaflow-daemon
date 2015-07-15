package jirautils

import (
	// Stdlib
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/url"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/env"

	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

func NewClient() (client *jira.Client, err error) {
	defer env.Recover(&err)

	var (
		baseURL          = env.MustGetenv("JIRA_BASE_URL")
		oauthConsumerKey = env.MustGetenv("JIRA_OAUTH_CONSUMER_KEY")
		oauthPrivateKey  = env.MustGetenv("JIRA_OAUTH_PRIVATE_KEY")
		oauthAccessToken = env.MustGetenv("JIRA_OAUTH_ACCESS_TOKEN")
	)

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(oauthPrivateKey))
	if block == nil {
		return nil, errors.New("failed to parse OAuth private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	httpClient := jira.NewOAuthClient(base, oauthConsumerKey, privateKey, oauthAccessToken)
	return jira.NewClient(base, httpClient), nil
}
