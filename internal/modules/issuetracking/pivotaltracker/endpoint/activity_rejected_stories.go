package endpoint

import (
	// Stdlib
	"net/http"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/log"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/config"

	// Vendor
	"github.com/salsita/go-pivotaltracker/v5/pivotal"
)

func handleRejectedStories(r *http.Request, projectId int, change *Change) error {
	// Check whether we want to process this change or not.
	switch {
	case change.ResourceKind != "story":
		fallthrough
	case change.NewValues.State != pivotal.StoryStateRejected:
		return nil
	}

	// Fetch the story resource.
	var (
		cfg = config.Get()
		pid = projectId
		sid = change.ResourceID
	)
	client := pivotal.NewClient(cfg.Token)
	story, _, err := client.Stories.Get(pid, sid)
	if err != nil {
		return err
	}

	// Drop relevant labels.
	var newLabels []*pivotal.Label
	for _, label := range story.Labels {
		switch label.Name {
		case cfg.ReviewedLabel:
		case cfg.ReviewSkippedLabel:
		case cfg.TestingPassedLabel:
		case cfg.TestingFailedLabel:
		case cfg.TestingSkippedLabel:
		default:
			newLabels = append(newLabels, &pivotal.Label{Name: label.Name})
		}
	}

	// No change, we are done.
	if len(newLabels) == len(story.Labels) {
		return nil
	}

	// Update the story.
	_, _, err = client.Stories.Update(pid, sid, &pivotal.StoryRequest{
		Labels: &newLabels,
	})
	if err != nil {
		return err
	}

	log.Info(r, "Pivotal Tracker: story %v rejected, pruned the workflow labels", sid)
	return nil
}
