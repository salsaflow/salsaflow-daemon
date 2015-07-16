package issues

import (
	// Stdlib
	"fmt"
	"strings"

	// Vendor
	"github.com/google/go-github/github"
)

// ReviewIssue represents the common interface for all review issue types.
type ReviewIssue interface {

	// AddCommit adds the commit to the commit checklist.
	AddCommit(commitSHA, commitTitle string, done bool) (added bool)

	// CommitItems returns the list of commits contained in the commit checklist.
	CommitItems() []*CommitItem

	// AddReviewBlocker adds the review blocker to the blocker checkbox.
	AddReviewBlocker(commitSHA, commentURL, blockerSummary string, fixed bool) (added bool)

	// ReviewBlockerItems returns the list of blockers contained in the blocker checklist.
	ReviewBlockerItems() []*ReviewBlockerItem

	// FormatTitle returns the review issue title for the given issue type.
	FormatTitle() string

	// FormatBody returns the review issue body for the given issue type.
	FormatBody() string
}

// ParseReviewIssue parses the given GitHub review issue and returns
// a *StoryReviewIssue or *CommitReviewIssue based on the issue type,
// which both implement ReviewIssue interface.
func ParseReviewIssue(issue *github.Issue) (ReviewIssue, error) {
	// Use the title prefix to decide the review issue type.
	switch {
	case strings.HasPrefix(*issue.Title, "Review story"):
		return parseStoryReviewIssue(issue)
	case strings.HasPrefix(*issue.Title, "Review commit"):
		return parseCommitReviewIssue(issue)
	default:
		return nil, &ErrUnknownReviewIssueType{issue}
	}
}

// FindIssueForCommitItem returns the GitHub issue that contains
// the given commit in its commit checklist.
func FindIssueForCommitItem(
	client *github.Client,
	owner string,
	repo string,
	commitSHA string,
	commitTitle string,
) (*github.Issue, error) {

	// Find the relevant review issue.
	pattern := fmt.Sprintf("] %v: %v", commitSHA[:7], commitTitle)

	query := fmt.Sprintf(
		`"%v" repo:"%v/%v" label:review type:issue in:body`, pattern, owner, repo)

	result, _, err := client.Search.Issues(query, &github.SearchOptions{})
	if err != nil {
		return nil, err
	}
	if total := *result.Total; total != 1 {
		return nil, &ErrReviewIssueNotFound{commitSHA, result}
	}
	issue := result.Issues[0]
	return &issue, nil
}
