package tracker

import (
	// Stdlib
	"fmt"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/config"

	// Vendor
	"gopkg.in/salsita/go-pivotaltracker.v1/v5/pivotal"
)

type commonStory struct {
	stories storyService
	config  *config.Config
	story   *pivotal.Story
}

func (s *commonStory) OnReviewRequestOpened(rrID, rrURL string) error {
	return s.addComment(fmt.Sprintf("Review request [#%v](%v) opened.", rrID, rrURL))
}

func (s *commonStory) OnReviewRequestClosed(rrID, rrURL string) error {
	return nil
}

func (s *commonStory) OnReviewRequestReopened(rrID, rrURL string) error {
	// Check whether the state needs updating.
	state, changed := shouldUpdateState(s.story.State)

	// Get the new list of labels.
	// We drop 'reviewed', 'no review' and 'qa-'.
	labels := filterLabels(s.story.Labels, func(label *pivotal.Label) bool {
		switch label.Name {
		case s.config.ReviewedLabel:
		case s.config.ReviewSkippedLabel:
		case s.config.TestingFailedLabel:
		default:
			return true
		}
		changed = true
		return false
	})

	// Return in case we don't need to update the story.
	if !changed {
		return nil
	}

	return s.update(state, labels)
}

func (s *commonStory) MarkAsReviewed() error {
	// Check whether the state needs updating.
	state, changed := shouldUpdateState(s.story.State)

	// Get the new list of labels.
	// We drop 'no review' and 'qa-' and append 'reviewed'.
	// We actually remove 'reviewed' and we add it again,
	// but in that case we do not mark the story as changed.
	var skipAppend bool
	labels := filterLabels(s.story.Labels, func(label *pivotal.Label) bool {
		switch label.Name {
		case s.config.ReviewedLabel:
			skipAppend = true
			return true
		case s.config.ReviewSkippedLabel:
			fallthrough
		case s.config.TestingFailedLabel:
			changed = true
			return false
		default:
			return true
		}
	})
	if !skipAppend {
		// We append 'reviewed'.
		labels = append(labels, &pivotal.Label{Name: s.config.ReviewedLabel})
		changed = true
	}

	// Return in case we don't need to update the story.
	if !changed {
		return nil
	}

	// Update the story.
	return s.update(state, labels)
}

func (s *commonStory) addComment(text string) error {
	var (
		pid = s.story.ProjectId
		sid = s.story.Id
	)
	comment, _, err := s.stories.AddComment(pid, sid, &pivotal.Comment{
		Text: text,
	})
	if err != nil {
		return err
	}

	s.story.CommentIds = append(s.story.CommentIds, comment.Id)
	return nil
}

func (s *commonStory) update(state string, labels []*pivotal.Label) error {
	// Set the state field.
	req := &pivotal.StoryRequest{State: state}

	// Set the labels field.
	if len(labels) != 0 {
		ls := mapLabels(labels, func(label *pivotal.Label) *pivotal.Label {
			return &pivotal.Label{Name: label.Name}
		})
		req.Labels = &ls
	}

	// Update.
	story, _, err := s.stories.Update(s.story.ProjectId, s.story.Id, req)
	if err != nil {
		return err
	}
	s.story = story
	return nil
}

func shouldUpdateState(state string) (string, bool) {
	switch state {
	case pivotal.StoryStateFinished:
	case pivotal.StoryStateDelivered:
	case pivotal.StoryStateAccepted:
	default:
		return pivotal.StoryStateFinished, true
	}
	return "", false
}

func filterLabels(
	labels []*pivotal.Label,
	filterFunc func(*pivotal.Label) bool,
) []*pivotal.Label {

	ls := make([]*pivotal.Label, 0, len(labels))
	for _, label := range labels {
		if filterFunc(label) {
			ls = append(ls, label)
		}
	}
	return ls
}

func mapLabels(
	labels []*pivotal.Label,
	mapFunc func(*pivotal.Label) *pivotal.Label,
) []*pivotal.Label {

	ls := make([]*pivotal.Label, len(labels))
	for i := range labels {
		ls[i] = mapFunc(labels[i])
	}
	return ls
}
