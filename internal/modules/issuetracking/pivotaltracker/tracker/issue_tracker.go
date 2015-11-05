package tracker

import (
	// Stdlib
	"fmt"
	"net/http"
	"strconv"
	"strings"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/common"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/config"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/util"

	// Vendor
	"gopkg.in/salsita/go-pivotaltracker.v1/v5/pivotal"
)

const Id = "Pivotal Tracker"

type storyService interface {
	Get(projectId, storyId int) (*pivotal.Story, *http.Response, error)
	Update(projectId, storyId int, story *pivotal.StoryRequest) (*pivotal.Story, *http.Response, error)
	AddComment(projectId, storyId int, comment *pivotal.Comment) (*pivotal.Comment, *http.Response, error)
}

type issueTracker struct {
	stories storyService
	config  config.Config
}

func Factory() (common.IssueTracker, error) {
	client, err := util.NewClient()
	if err != nil {
		return nil, err
	}

	return &issueTracker{
		stories: client.Stories,
		config:  config.Get(),
	}, nil
}

func (tracker *issueTracker) FindStoryByTag(storyTag string) (common.Story, error) {
	pid, sid, err := parseStoryTag(storyTag)
	if err != nil {
		return nil, err
	}

	story, _, err := tracker.stories.Get(pid, sid)
	if err != nil {
		return nil, err
	}

	return &commonStory{tracker.stories, &tracker.config, story}, nil
}

func parseStoryTag(storyTag string) (pid, sid int, err error) {
	parts := strings.Split(storyTag, "/")
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("Pivotal Tracker: malformed story tag: %v", storyTag)
	}

	pidString := parts[0]
	pid, err = strconv.Atoi(pidString)
	if err != nil {
		return 0, 0, fmt.Errorf("Pivotal Tracker: malformed project ID: %v", pidString)
	}

	sidString := parts[2]
	sid, err = strconv.Atoi(sidString)
	if err != nil {
		return 0, 0, fmt.Errorf("Pivotal Tracker: malformed story ID: %v", sidString)
	}

	return pid, sid, nil
}
