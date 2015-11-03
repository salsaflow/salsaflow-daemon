package tracker

import (
	// Stdlib
	"errors"
	"net/http"
	"testing"

	// Vendor
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"gopkg.in/salsita/go-pivotaltracker.v1/v5/pivotal"
)

// Set up Ginkgo and Gomega ----------------------------------------------------

func TestIssueTracker(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Tracker Suite")
}

// Testing imports -------------------------------------------------------------

var (
	BeforeEach = ginkgo.BeforeEach
	Context    = ginkgo.Context
	Describe   = ginkgo.Describe
	It         = ginkgo.It

	BeTrue = gomega.BeTrue
	BeNil  = gomega.BeNil
	Equal  = gomega.Equal
	Expect = gomega.Expect
)

// Shared testing infrastructure -----------------------------------------------

const (
	testingProjectId        = 102030
	testingStoryId          = 302010
	testingReviewRequestId  = "10"
	testingReviewRequestURL = "https://some-review-request-url"
)

type (
	StoryGetFunc        func(int, int) (*pivotal.Story, *http.Response, error)
	StoryUpdateFunc     func(int, int, *pivotal.StoryRequest) (*pivotal.Story, *http.Response, error)
	StoryAddCommentFunc func(int, int, *pivotal.Comment) (*pivotal.Comment, *http.Response, error)
)

type testingStoryService struct {
	GetMock        StoryGetFunc
	UpdateMock     StoryUpdateFunc
	AddCommentMock StoryAddCommentFunc
}

func (srv *testingStoryService) Get(
	projectId int,
	storyId int,
) (*pivotal.Story, *http.Response, error) {

	if srv.GetMock == nil {
		return nil, nil, errors.New("GetMock function is not set")
	}
	return srv.GetMock(projectId, storyId)
}

func (srv *testingStoryService) Update(
	projectId int,
	storyId int,
	story *pivotal.StoryRequest,
) (*pivotal.Story, *http.Response, error) {

	if srv.UpdateMock == nil {
		return nil, nil, errors.New("UpdateMock function is not set")
	}
	return srv.UpdateMock(projectId, storyId, story)
}

func (srv *testingStoryService) AddComment(
	projectId int,
	storyId int,
	comment *pivotal.Comment,
) (*pivotal.Comment, *http.Response, error) {

	if srv.AddCommentMock == nil {
		return nil, nil, errors.New("AddCommentMock function is not set")
	}
	return srv.AddCommentMock(projectId, storyId, comment)
}
