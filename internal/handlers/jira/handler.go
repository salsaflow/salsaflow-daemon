package jira

import (
	// Stdlib
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	// Vendor
	"github.com/salsita/go-jira/v2/jira"
)

func NewMeHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var (
			baseURL          = os.Getenv("JIRA_BASE_URL")
			oauthConsumerKey = os.Getenv("JIRA_OAUTH_CONSUMER_KEY")
			oauthPrivateKey  = os.Getenv("JIRA_OAUTH_PRIVATE_KEY")
			oauthAccessToken = os.Getenv("JIRA_OAUTH_ACCESS_TOKEN")
		)

		base, err := url.Parse(baseURL)
		if err != nil {
			httpError(rw, err)
			return
		}

		block, _ := pem.Decode([]byte(oauthPrivateKey))
		if block == nil {
			httpError(rw, errors.New("failed to parse OAuth private key"))
			return
		}
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			httpError(rw, err)
			return
		}

		httpClient := jira.NewOAuthClient(base, oauthConsumerKey, privateKey, oauthAccessToken)

		client := jira.NewClient(base, httpClient)

		me, _, err := client.Myself.Get()
		if err != nil {
			httpError(rw, err)
			return
		}

		fmt.Fprintln(rw, me.Name)
	})
}

func httpError(rw http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	http.Error(rw, http.StatusText(code), code)
	log.Println("Error:", err)
}
