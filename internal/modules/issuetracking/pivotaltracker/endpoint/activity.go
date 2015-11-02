package endpoint

import (
	// Stdlib
	"encoding/json"
	"net/http"

	// Internal
	httputil "github.com/salsaflow/salsaflow-daemon/internal/http"
	"github.com/salsaflow/salsaflow-daemon/internal/log"
)

type Activity struct {
	Changes []*Change `json:"changes"`
	Project struct {
		Id int `json:"id"`
	} `json:"project"`
}

type Change struct {
	ResourceKind string `json:"kind"`
	ResourceID   int    `json:"id"`
	NewValues    Values `json:"new_values"`
}

type Values struct {
	State string `json:"current_state"`
}

type activityHandlerFunc func(r *http.Request, projectId int, change *Change) error

var activityHandlers = []activityHandlerFunc{
	handleRejectedStories,
}

func handleActivity(rw http.ResponseWriter, r *http.Request) {
	// Decode the activity object.
	var activity Activity
	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
		httputil.Error(rw, r, err)
		return
	}

	// Process the changes.
	var errorOccured bool

	pid := activity.Project.Id
	for _, change := range activity.Changes {
		for _, handler := range activityHandlers {
			if err := handler(r, pid, change); err != nil {
				log.Error(r, err)
				errorOccured = true
			}
		}
	}

	if errorOccured {
		httputil.Status(rw, http.StatusInternalServerError)
	} else {
		httputil.Status(rw, http.StatusAccepted)
	}
}
