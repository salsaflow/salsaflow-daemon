package jira

import (
	// Stdlib
	"fmt"
	"net/http"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/utils/httputils"
	"github.com/salsaflow/salsaflow-daemon/internal/utils/jirautils"
)

func NewMeHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		client, err := jirautils.NewClient()
		if err != nil {
			httputils.Error(rw, err)
			return
		}

		me, _, err := client.Myself.Get()
		if err != nil {
			httputils.Error(rw, err)
			return
		}

		fmt.Fprintln(rw, me.Name)
	})
}
