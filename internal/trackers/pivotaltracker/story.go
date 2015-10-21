package pivotaltracker

import (
	// Stdlib
	"fmt"

	// Vendor
	"github.com/salsita/go-pivotaltracker/v5/pivotal"
)

type commonStory struct {
	client *pivotal.Client
	story  *pivotal.Story
}

func (s *commonStory) OnReviewRequestOpened(rrID, rrURL string) error {
	return s.addComment(fmt.Sprintf("Review request [#%v](%v) opened.", rrID, rrURL))
}

func (s *commonStory) OnReviewRequestClosed(rrID, rrURL string) error {
	return nil
}

func (s *commonStory) OnReviewRequestReopened(rrID, rrURL string) error {
	// Prune workflow labels.
	// Reset to finished.
	state := pivotal.StoryStateFinished
	labels := pruneLabels(s.story.Labels)

	// Update the story.
	req := &pivotal.StoryRequest{
		State:  state,
		Labels: &labels,
	}
	story, _, err := s.client.Stories.Update(s.story.ProjectId, s.story.Id, req)
	if err != nil {
		return err
	}

	s.story = story
	return nil
}

func (s *commonStory) MarkAsReviewed() error {
	// Get PT config.
	config := GetConfig()
	qaLabel := testingLabel(s.story)

	var (
		newState  string
		newLabels []*pivotal.Label
	)

	switch s.story.State {
	case pivotal.StoryStateUnscheduled:
		fallthrough
	case pivotal.StoryStatePlanned:
		fallthrough
	case pivotal.StoryStateUnstarted:
		fallthrough
	case pivotal.StoryStateStarted:
		fallthrough
	case pivotal.StoryStateFinished:
		fallthrough
	case pivotal.StoryStateRejected:
		// Set story state to finished.
		newState = pivotal.StoryStateFinished

		// Prune all workflow labels and add 'reviewed'.
		newLabels = append(pruneLabels(s.story.Labels), &pivotal.Label{
			Name: config.ReviewedLabel,
		})

	case pivotal.StoryStateDelivered:
		// In case the story is delivered, we leave it that way.
		newState = pivotal.StoryStateDelivered

		// We make sure 'reviewed' label is there, but we also
		// keep the testing label there unless it is 'qa-', in which case we drop it.
		newLabels = append(pruneLabels(s.story.Labels), &pivotal.Label{
			Name: config.ReviewedLabel,
		})
		switch qaLabel {
		case config.TestingFailedLabel:
			// Drop 'qa-' in case it is there.
		case config.TestingPassedLabel:
			// Keep 'qa+' in case it is there.
			fallthrough
		case config.TestingSkippedLabel:
			// Keep 'no qa' in case it is there.
			newLabels = append(newLabels, &pivotal.Label{
				Name: qaLabel,
			})
		}

	case pivotal.StoryStateAccepted:
		newState = pivotal.StoryStateAccepted
	}

	// Prune workflow labels.
	// Add the reviewed label.
	// Reset to finished.
	state := pivotal.StoryStateFinished
	labels := append(pruneLabels(s.story.Labels), &pivotal.Label{
		Name: config.ReviewedLabel,
	})

	// Update the story.
	req := &pivotal.StoryRequest{
		State:  state,
		Labels: &labels,
	}
	story, _, err := s.client.Stories.Update(s.story.ProjectId, s.story.Id, req)
	if err != nil {
		return err
	}

	s.story = story
	return nil
}

func (s *commonStory) addComment(text string) error {
	var (
		pid = s.story.ProjectId
		sid = s.story.Id
	)
	comment, _, err := s.client.Stories.AddComment(pid, sid, &pivotal.Comment{
		Text: text,
	})
	if err != nil {
		return err
	}

	s.story.CommentIds = append(s.story.CommentIds, comment.Id)
	return nil
}

func pruneLabels(labels []*pivotal.Label) []*pivotal.Label {
	ls := make([]*pivotal.Label, 0, len(labels))
	for _, label := range labels {
		switch label.Name {
		case config.ReviewedLabel:
		case config.ReviewSkippedLabel:
		case config.TestingPassedLabel:
		case config.TestingFailedLabel:
		default:
			ls = append(ls, &pivotal.Label{
				Id: label.Id,
			})
		}
	}
	return ls
}

func isLabeled(story *pivotal.Story, labelName string) bool {
	for _, label := range story.Labels {
		if label.Name == labelName {
			return true
		}
	}
	return false
}

func testingLabel(story *pivotal.Story) string {
	config := GetConfig()
	qaLabels := [...]string{
		config.TestingPassedLabel,
		config.TestingFailedLabel,
		config.TestingSkippedlabel,
	}

	for _, label := range labels {
		if isLabeled(story, label) {
			return label
		}
	}
	return ""
}
