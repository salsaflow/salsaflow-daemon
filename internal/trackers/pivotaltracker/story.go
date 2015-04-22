package pivotaltracker

import (
	// Vendor
	"github.com/salsita/go-pivotaltracker/v5/pivotal"
)

type commonStory struct {
	client *pivotal.Client
	story  *pivotal.Story
}

func (s *commonStory) OnReviewRequestOpened(rrID, rrURL string) error {
	return s.addComment(fmt.Sprintf("GitHub review issue [#%v](%v) opened.", rrID, rrURL))
}

func (s *commonStory) OnReviewRequestClosed(rrID, rrURL string) error {
	return s.addComment(fmt.Sprintf("GitHub review issue [#%v](%v) closed.", rrID, rrURL))
}

func (s *commonStory) OnReviewRequestReopened(rrID, rrURL string) error {
	return s.addComment(fmt.Sprintf("GitHub review issue [#%v](%v) reopened.", rrID, rrURL))
}

func (s *commonStory) MarkAsReviewed() error {
	// Add the 'reviewed' label.
	req := &pivotal.Story{
		Labels: append(s.story.Labels, &pivotal.Label{Name: "reviewed"}),
	}

	_, _, err := s.client.Stories.Update(s.story.ProjectId, s.story.StoryId, req)
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
	comment, _, err := client.Stories.AddComment(pid, sid, &client.StoryComment{
		Text: text,
	})
	if err != nil {
		return err
	}

	s.story.CommentIds = append(s.story.CommentIds, comment.Id)
	return
}
