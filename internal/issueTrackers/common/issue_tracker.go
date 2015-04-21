package common

// IssueTracker is a common interface that must be implemented by
// all modules representing an issue tracker.
type IssueTracker interface {

	// FindStory can be used to find a story by its ID.
	FindStory(projectId, storyId string) (Story, error)
}

// Story represents a common interface for issue tracker stories.
// This is where the event handling occurs.
type Story interface {

	// OnReviewRequestOpened is called to handle the RR opened event.
	OnReviewRequestOpened(rrID, rrURL string) error

	// OnReviewRequestClosed is called to handle the RR closed event.
	OnReviewRequestClosed(rrID, rrURL string) error

	// OnReviewRequestReopened is called to handle the RR reopened event.
	OnReviewRequestReopened(rrID, rrURL string) error
}
