package jirautils

import (
	// Stdlib
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"os"

	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

type ErrNotSet struct {
	varName string
}

func (err *ErrNotSet) Error() string {
	return fmt.Sprintf("Environment variable not set: %v", err.varName)
}

func NewClient() (client *jira.Client, err error) {
	mustGetenv := func(varName string) string {
		value := os.Getenv(varName)
		if value == "" {
			panic(&ErrNotSet{varName})
		}
		return value
	}

	defer func() {
		if r := recover(); r != nil {
			if ex, ok := r.(*ErrNotSet); ok {
				err = ex
			} else {
				panic(r)
			}
		}
	}()

	var (
		baseURL          = mustGetenv("JIRA_BASE_URL")
		oauthConsumerKey = mustGetenv("JIRA_OAUTH_CONSUMER_KEY")
		oauthPrivateKey  = mustGetenv("JIRA_OAUTH_PRIVATE_KEY")
		oauthAccessToken = mustGetenv("JIRA_OAUTH_ACCESS_TOKEN")
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
