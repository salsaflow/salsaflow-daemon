package tracker

import (
	// Stdlib
	"fmt"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/config"

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
	// Prune workflow labels.
	// Add the reviewed label.
	// Reset to finished.
	state := pivotal.StoryStateFinished
	labels := append(pruneLabels(s.story.Labels), &pivotal.Label{
		Name: config.Get().ReviewedLabel,
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
	cfg := config.Get()
	ls := make([]*pivotal.Label, 0, len(labels))
	for _, label := range labels {
		switch label.Name {
		case cfg.ReviewedLabel:
		case cfg.ReviewSkippedLabel:
		case cfg.TestingPassedLabel:
		case cfg.TestingFailedLabel:
		default:
			ls = append(ls, &pivotal.Label{
				Id: label.Id,
			})
		}
	}
	return ls
}
