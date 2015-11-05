package tracker

import (
	// Stdlib
	"fmt"
	"net/http"

	// Internal
	"github.com/salsaflow/salsaflow-daemon/internal/modules/issuetracking/pivotaltracker/config"

	// Vendor
	"gopkg.in/salsita/go-pivotaltracker.v1/v5/pivotal"
)

// TODO: This file would need some more love and refactoring,
//       but let's not waste time on this right now.

var _ = Describe("Invoking OnReviewRequestOpened story event handler", func() {

	var (
		stories *testingStoryService
		cfg     *config.Config
		ptStory *pivotal.Story
		story   *commonStory
	)

	BeforeEach(func() {
		stories = &testingStoryService{}
		cfg = &config.Config{}
		ptStory = &pivotal.Story{
			Id:        testingStoryId,
			ProjectId: testingProjectId,
		}
		story = &commonStory{stories, cfg, ptStory}
	})

	It("should result in a comment being added to the relevant story", func() {

		expectedText := fmt.Sprintf("Review request [#%v](%v) opened.",
			testingReviewRequestId, testingReviewRequestURL)

		var addCommentCalled bool

		stories.AddCommentMock = func(
			projectId int,
			storyId int,
			comment *pivotal.Comment,
		) (*pivotal.Comment, *http.Response, error) {

			Expect(projectId).To(Equal(testingProjectId))
			Expect(storyId).To(Equal(testingStoryId))
			Expect(comment).To(Equal(&pivotal.Comment{Text: expectedText}))

			addCommentCalled = true
			return &pivotal.Comment{}, nil, nil
		}

		err := story.OnReviewRequestOpened(testingReviewRequestId, testingReviewRequestURL)

		Expect(err).To(BeNil())
		Expect(addCommentCalled).To(BeTrue())
	})
})

var _ = Describe("Invoking OnReviewRequestClosed story event handler", func() {

	var (
		stories *testingStoryService
		cfg     *config.Config
		ptStory *pivotal.Story
		story   *commonStory
	)

	BeforeEach(func() {
		stories = &testingStoryService{}
		cfg = &config.Config{}
		ptStory = &pivotal.Story{
			Id:        testingStoryId,
			ProjectId: testingProjectId,
		}
		story = &commonStory{stories, cfg, ptStory}
	})

	It("does nothing", func() {

		// We don't set any mock function, which means that calling any service
		// method returns an error, so getting nil error means that no method was called.
		err := story.OnReviewRequestClosed(testingReviewRequestId, testingReviewRequestURL)
		Expect(err).To(BeNil())
	})
})

var _ = Describe("Invoking OnReviewRequestReopened story event handler", func() {

	reviewed := &pivotal.Label{Name: "reviewed"}
	noReview := &pivotal.Label{Name: "no review"}
	qaPlus := &pivotal.Label{Name: "qa+"}
	qaMinus := &pivotal.Label{Name: "qa-"}
	noQA := &pivotal.Label{Name: "no qa"}
	other := &pivotal.Label{Name: "other"}

	var (
		stories *testingStoryService
		cfg     *config.Config
		ptStory *pivotal.Story
		story   *commonStory
	)

	BeforeEach(func() {
		stories = &testingStoryService{}
		cfg = &config.Config{
			ReviewedLabel:       "reviewed",
			ReviewSkippedLabel:  "no review",
			TestingPassedLabel:  "qa+",
			TestingFailedLabel:  "qa-",
			TestingSkippedLabel: "no qa",
		}
		ptStory = &pivotal.Story{
			Id:        testingStoryId,
			ProjectId: testingProjectId,
		}
		story = &commonStory{stories, cfg, ptStory}
	})

	data := []struct {
		state   string
		labels  []*pivotal.Label
		request *pivotal.StoryRequest
	}{
		{
			pivotal.StoryStateUnscheduled,
			nil,
			&pivotal.StoryRequest{State: pivotal.StoryStateFinished},
		},
		{
			pivotal.StoryStatePlanned,
			nil,
			&pivotal.StoryRequest{State: pivotal.StoryStateFinished},
		},
		{
			pivotal.StoryStateUnstarted,
			nil,
			&pivotal.StoryRequest{State: pivotal.StoryStateFinished},
		},
		{
			pivotal.StoryStateStarted,
			nil,
			&pivotal.StoryRequest{State: pivotal.StoryStateFinished},
		},
		{
			pivotal.StoryStateFinished,
			nil,
			nil,
		},
		{
			pivotal.StoryStateDelivered,
			nil,
			nil,
		},
		{
			pivotal.StoryStateAccepted,
			nil,
			nil,
		},
		{
			pivotal.StoryStateRejected,
			nil,
			&pivotal.StoryRequest{State: pivotal.StoryStateFinished},
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{other},
			nil,
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{reviewed, other},
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{other},
			},
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{noReview, other},
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{other},
			},
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{noQA, other},
			nil,
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{qaPlus, other},
			nil,
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{qaMinus, other},
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{other},
			},
		},
		{
			pivotal.StoryStateRejected,
			[]*pivotal.Label{reviewed, other},
			&pivotal.StoryRequest{
				State:  pivotal.StoryStateFinished,
				Labels: &[]*pivotal.Label{other},
			},
		},
	}

	for i := range data {
		func(i int) {
			td := data[i]
			ctx := fmt.Sprintf("state=%v, labels=%v, update=%+v", td.state, td.labels, td.request)

			Context(ctx, func() {

				It("sends out out the expected story update request", func() {

					ptStory.State = td.state
					ptStory.Labels = td.labels

					var updateCalled bool

					stories.UpdateMock = func(
						projectId int,
						storyId int,
						story *pivotal.StoryRequest,
					) (*pivotal.Story, *http.Response, error) {

						Expect(projectId).To(Equal(testingProjectId))
						Expect(storyId).To(Equal(testingStoryId))
						Expect(story).To(Equal(td.request))

						updateCalled = true
						return &pivotal.Story{}, nil, nil
					}

					err := story.OnReviewRequestReopened(
						testingReviewRequestId, testingReviewRequestURL)

					Expect(err).To(BeNil())
					Expect(updateCalled).To(Equal(td.request != nil))
				})
			})
		}(i)
	}
})

var _ = Describe("Calling commonStory.MarkAsReviewed", func() {

	reviewed := &pivotal.Label{Name: "reviewed"}

	noReview := &pivotal.Label{Name: "no review"}
	qaPlus := &pivotal.Label{Name: "qa+"}
	qaMinus := &pivotal.Label{Name: "qa-"}
	noQA := &pivotal.Label{Name: "no qa"}
	other := &pivotal.Label{Name: "other"}

	var (
		stories *testingStoryService
		cfg     *config.Config
		ptStory *pivotal.Story
		story   *commonStory
	)

	BeforeEach(func() {
		stories = &testingStoryService{}
		cfg = &config.Config{
			ReviewedLabel:       "reviewed",
			ReviewSkippedLabel:  "no review",
			TestingPassedLabel:  "qa+",
			TestingFailedLabel:  "qa-",
			TestingSkippedLabel: "no qa",
		}
		ptStory = &pivotal.Story{
			Id:        testingStoryId,
			ProjectId: testingProjectId,
		}
		story = &commonStory{stories, cfg, ptStory}
	})

	data := []struct {
		state   string
		labels  []*pivotal.Label
		request *pivotal.StoryRequest
	}{
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{reviewed},
			nil,
		},
		{
			pivotal.StoryStateUnscheduled,
			nil,
			&pivotal.StoryRequest{
				State:  pivotal.StoryStateFinished,
				Labels: &[]*pivotal.Label{reviewed},
			},
		},
		{
			pivotal.StoryStatePlanned,
			nil,
			&pivotal.StoryRequest{
				State:  pivotal.StoryStateFinished,
				Labels: &[]*pivotal.Label{reviewed},
			},
		},
		{
			pivotal.StoryStateUnstarted,
			nil,
			&pivotal.StoryRequest{
				State:  pivotal.StoryStateFinished,
				Labels: &[]*pivotal.Label{reviewed},
			},
		},
		{
			pivotal.StoryStateStarted,
			nil,
			&pivotal.StoryRequest{
				State:  pivotal.StoryStateFinished,
				Labels: &[]*pivotal.Label{reviewed},
			},
		},
		{
			pivotal.StoryStateFinished,
			nil,
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{reviewed},
			},
		},
		{
			pivotal.StoryStateDelivered,
			nil,
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{reviewed},
			},
		},
		{
			pivotal.StoryStateAccepted,
			nil,
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{reviewed},
			},
		},
		{
			pivotal.StoryStateRejected,
			nil,
			&pivotal.StoryRequest{
				State:  pivotal.StoryStateFinished,
				Labels: &[]*pivotal.Label{reviewed},
			},
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{other},
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{other, reviewed},
			},
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{noReview, other},
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{other, reviewed},
			},
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{noQA, other},
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{noQA, other, reviewed},
			},
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{qaPlus, other},
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{qaPlus, other, reviewed},
			},
		},
		{
			pivotal.StoryStateFinished,
			[]*pivotal.Label{qaMinus, other},
			&pivotal.StoryRequest{
				Labels: &[]*pivotal.Label{other, reviewed},
			},
		},
	}

	for i := range data {
		func(i int) {
			td := data[i]
			ctx := fmt.Sprintf("state=%v, labels=%v, update=%+v", td.state, td.labels, td.request)

			Context(ctx, func() {

				It("sends out out the expected story update request", func() {

					ptStory.State = td.state
					ptStory.Labels = td.labels

					var updateCalled bool

					stories.UpdateMock = func(
						projectId int,
						storyId int,
						story *pivotal.StoryRequest,
					) (*pivotal.Story, *http.Response, error) {

						Expect(projectId).To(Equal(testingProjectId))
						Expect(storyId).To(Equal(testingStoryId))
						Expect(story).To(Equal(td.request))

						updateCalled = true
						return &pivotal.Story{}, nil, nil
					}

					err := story.MarkAsReviewed()

					Expect(err).To(BeNil())
					Expect(updateCalled).To(Equal(td.request != nil))
				})
			})
		}(i)
	}
})
