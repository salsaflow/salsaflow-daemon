package jira

import (
	// Stdlib
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
			baseURL      = os.Getenv("JIRA_BASE_URL")
			clientKey    = os.Getenv("JIRA_OAUTH_CONSUMER_KEY")
			clientSecret = os.Getenv("JIRA_OAUTH_CONSUMER_SECRET")
			accessToken  = os.Getenv("JIRA_OAUTH_ACCESS_TOKEN")
		)

		base, err := url.Parse(baseURL)
		if err != nil {
			httpError(rw, err)
			return
		}

		httpClient := jira.NewOAuthClient(clientKey, clientSecret, accessToken)

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
