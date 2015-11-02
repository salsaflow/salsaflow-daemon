package tracker

import (
	// Stdlib
	"fmt"
	"regexp"
	"strconv"

	// Internal
	githubutil "github.com/salsaflow/salsaflow-daemon/internal/github"
	"github.com/salsaflow/salsaflow-daemon/internal/modules/common"

	// Vendor
	"github.com/google/go-github/github"
)

const Id = "GitHub Issues"

func Factory() (common.IssueTracker, error) {
	client, err := githubutil.NewClient()
	if err != nil {
		return nil, err
	}

	return &issueTracker{client}, nil
}

type issueTracker struct {
	client *github.Client
}

func (tracker *issueTracker) FindStoryByTag(storyTag string) (common.Story, error) {
	owner, repo, issueNum, err := parseStoryTag(storyTag)
	if err != nil {
		return nil, err
	}

	issue, _, err := tracker.client.Issues.Get(owner, repo, issueNum)
	if err != nil {
		return nil, err
	}

	return &commonStory{tracker.client, issue, owner, repo}, nil
}

func parseStoryTag(storyTag string) (owner, repo string, issueNum int, err error) {
	// The format is owner/repo#issueNum
	re := regexp.MustCompile("^([^/]+)/([^#]+)#([0-9]+)$")
	match := re.FindStringSubmatch(storyTag)
	if len(match) != 4 {
		return "", "", 0, fmt.Errorf("GitHub Issues: malformed story tag: %v", storyTag)
	}

	issueNum, _ = strconv.Atoi(match[3])
	return match[1], match[2], issueNum, nil
}
