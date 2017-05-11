package status_groomer_test

import (
	"fmt"
	"net/http"

	. "github.com/jaresty/concourse-tracker-bot/status_groomer"
	"github.com/jaresty/concourse-tracker-bot/status_groomer/fakes"
	"github.com/jaresty/concourse-tracker-bot/tracker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("StatusGroomer", func() {
	var (
		failedJob           Job
		successfulJob       Job
		mockServerUrl       string
		mockServer          *ghttp.Server
		mockTrackerClient   *fakes.FakeTrackerClient
		mockConcourseClient *fakes.FakeConcourseClient
		mockLog             *fakes.FakeLogger
	)

	Context("when there are failed builds", func() {
		BeforeEach(func() {
			mockServer = ghttp.NewServer()
			mockTrackerClient = new(fakes.FakeTrackerClient)
			mockConcourseClient = new(fakes.FakeConcourseClient)
			mockLog = new(fakes.FakeLogger)

			mockServerUrl = mockServer.URL()
			mockConcourseClient.GetJobURLsReturns([]string{mockServerUrl + "/pipeline/build1", mockServerUrl + "/pipeline/build2"}, nil)
			failedJob = Job{
				FinishedBuild: Build{
					Status:       "failed",
					PipelineName: "fooPipeline",
					URL:          "/pipeline/build1",
				}}
			successfulJob = Job{
				FinishedBuild: Build{
					Status:       "success",
					PipelineName: "fooPipeline",
					URL:          "/pipeline/build2",
				}}
			mockServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/pipeline/build1"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, failedJob),
				))
			mockServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/pipeline/build2"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, successfulJob),
				))
		})

		Context("when a pre-existing story does not exist", func() {
			BeforeEach(func() {
				mockTrackerClient.StoriesReturns([]tracker.Story{tracker.Story{ID: 2}}, nil)
			})
			It("creates a new story and adds the failed build as a comment", func() {
				Groom(mockServerUrl, "husbandandwife", 12345, mockTrackerClient, mockConcourseClient, mockLog, 0)

				Expect(mockTrackerClient.CreateStoryCallCount()).To(Equal(1))
				_, createdStory := mockTrackerClient.CreateStoryArgsForCall(0)
				Expect(createdStory.StoryType).To(Equal("chore"))
				Expect(createdStory.CurrentState).To(Equal("unstarted"))
				Expect(createdStory.Labels[0].Name).To(Equal("broken build"))
				Expect(len(createdStory.Labels)).To(Equal(1))
				Expect(createdStory.Comments[0].Text).To(Equal(fmt.Sprintf("%s/%s", mockServerUrl, failedJob.FinishedBuild.URL)))
				Expect(len(createdStory.Comments)).To(Equal(1))
				Expect(createdStory.BeforeID).To(Equal(2))
			})
		})

		Context("when a pre-existing story does exist", func() {
			var existingStory tracker.Story
			BeforeEach(func() {
				existingStory = tracker.Story{
					Name: fmt.Sprintf("%s/%s has %s", failedJob.FinishedBuild.PipelineName, failedJob.FinishedBuild.JobName, "failed"),
					ID:   2,
				}
				mockTrackerClient.StoriesReturns([]tracker.Story{existingStory}, nil)
			})
			It("adds the failed build as a new comment", func() {
				Groom(mockServerUrl, "husbandandwife", 12345, mockTrackerClient, mockConcourseClient, mockLog, 0)
				Expect(mockTrackerClient.ListCommentsCallCount()).To(Equal(1))
				trackerProjectID, storyID := mockTrackerClient.ListCommentsArgsForCall(0)
				Expect(trackerProjectID).To(Equal(12345))
				Expect(storyID).To(Equal(existingStory.ID))
				Expect(mockTrackerClient.AddCommentCallCount()).To(Equal(1))
				_, _, commentText := mockTrackerClient.AddCommentArgsForCall(0)
				Expect(commentText).To(Equal(fmt.Sprintf("%s/%s", mockServerUrl, failedJob.FinishedBuild.URL)))

			})
		})
	})

})
