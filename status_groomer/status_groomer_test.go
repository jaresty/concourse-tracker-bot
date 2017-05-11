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
		failedJob2          Job
		failedJob3          Job
		successfulJob       Job
		mockServerUrl       string
		mockServer          *ghttp.Server
		mockTrackerClient   *fakes.FakeTrackerClient
		mockConcourseClient *fakes.FakeConcourseClient
		mockLog             *fakes.FakeLogger
		groupingStrategy    map[string]string
	)

	Context("when there are failed builds", func() {
		BeforeEach(func() {
			mockServer = ghttp.NewServer()
			mockTrackerClient = new(fakes.FakeTrackerClient)
			mockConcourseClient = new(fakes.FakeConcourseClient)
			mockLog = new(fakes.FakeLogger)
			groupingStrategy = make(map[string]string)
			groupingStrategy["(fooPipeline-.*-groupa)"] = "groupa"

			mockServerUrl = mockServer.URL()
			// two failures to group
			failedJob = Job{
				FinishedBuild: Build{
					JobName:      "job-groupa",
					Status:       "failed",
					PipelineName: "fooPipeline",
					URL:          "/failed/group/1",
				}}
			failedJob2 = Job{
				FinishedBuild: Build{
					JobName:      "job2-groupa",
					Status:       "failed",
					PipelineName: "fooPipeline",
					URL:          "/failed/group/2",
				}}
			successfulJob = Job{
				FinishedBuild: Build{
					JobName:      "job-groupb",
					Status:       "success",
					PipelineName: "fooPipeline",
					URL:          "/success/nogroup/1",
				}}
			// one failure to not group
			failedJob3 = Job{
				FinishedBuild: Build{
					JobName:      "job3-groupc",
					Status:       "failed",
					PipelineName: "fooPipeline",
					URL:          "/failed/nogroup/1",
				}}
		})

		Context("with a matching group", func() {
			Context("groups by pipeline and job name", func() {
				Context("when a story does not exist", func() {
					BeforeEach(func() {
						mockConcourseClient.GetJobURLsReturns([]string{
							mockServerUrl + "/failed/group/1",
						}, nil)
						mockServer.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/failed/group/1"),
								ghttp.RespondWithJSONEncoded(http.StatusOK, failedJob),
							))
						mockTrackerClient.StoriesReturns([]tracker.Story{tracker.Story{ID: 2}}, nil)
					})
					It("creates a new story and adds the failed build as a comment", func() {
						Groom(groupingStrategy, mockServerUrl, "husbandandwife", 12345, mockTrackerClient, mockConcourseClient, mockLog, 0)

						Expect(mockTrackerClient.CreateStoryCallCount()).To(Equal(1))
						_, createdStory := mockTrackerClient.CreateStoryArgsForCall(0)
						Expect(createdStory.Name).To(Equal("groupa has failed"))
						Expect(createdStory.StoryType).To(Equal("chore"))
						Expect(createdStory.CurrentState).To(Equal("unstarted"))
						Expect(createdStory.Labels[0].Name).To(Equal("broken build"))
						Expect(len(createdStory.Labels)).To(Equal(1))
						Expect(len(createdStory.Comments)).To(Equal(1))
						Expect(createdStory.Comments[0].Text).To(Equal(fmt.Sprintf("%s/%s", mockServerUrl, failedJob.FinishedBuild.URL)))
						Expect(createdStory.BeforeID).To(Equal(2))
					})
				})

				Context("when a story already exists", func() {
					var existingStory tracker.Story
					BeforeEach(func() {
						mockConcourseClient.GetJobURLsReturns([]string{
							mockServerUrl + "/failed/group/1",
							mockServerUrl + "/failed/group/2",
						}, nil)
						mockServer.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/failed/group/1"),
								ghttp.RespondWithJSONEncoded(http.StatusOK, failedJob),
							))
						mockServer.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/failed/group/2"),
								ghttp.RespondWithJSONEncoded(http.StatusOK, failedJob2),
							))
						existingStory = tracker.Story{
							Name: "groupa has failed",
							ID:   2,
						}
						mockTrackerClient.StoriesReturns([]tracker.Story{existingStory}, nil)
					})
					It("adds the failed build as a new comment", func() {
						Groom(groupingStrategy, mockServerUrl, "husbandandwife", 12345, mockTrackerClient, mockConcourseClient, mockLog, 0)
						Expect(mockTrackerClient.ListCommentsCallCount()).To(Equal(2))
						trackerProjectID, storyID := mockTrackerClient.ListCommentsArgsForCall(0)
						Expect(trackerProjectID).To(Equal(12345))
						Expect(storyID).To(Equal(existingStory.ID))
						Expect(mockTrackerClient.AddCommentCallCount()).To(Equal(2))
						_, _, commentText := mockTrackerClient.AddCommentArgsForCall(0)
						Expect(commentText).To(Equal(fmt.Sprintf("%s/%s", mockServerUrl, failedJob.FinishedBuild.URL)))
						_, _, commentText = mockTrackerClient.AddCommentArgsForCall(1)
						Expect(commentText).To(Equal(fmt.Sprintf("%s/%s", mockServerUrl, failedJob2.FinishedBuild.URL)))
					})
				})
			})
		})

		Context("without a matching group", func() {
			BeforeEach(func() {
				mockConcourseClient.GetJobURLsReturns([]string{
					mockServerUrl + "/failed/nogroup/1",
					mockServerUrl + "/success/nogroup/1",
				}, nil)
				mockServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/failed/nogroup/1"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, failedJob3),
					))
				mockServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/success/nogroup/1"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, successfulJob),
					))
			})
			Context("when a pre-existing story does not exist", func() {
				BeforeEach(func() {
					mockTrackerClient.StoriesReturns([]tracker.Story{tracker.Story{ID: 2}}, nil)
				})
				It("creates a new story and adds the failed build as a comment", func() {
					Groom(groupingStrategy, mockServerUrl, "husbandandwife", 12345, mockTrackerClient, mockConcourseClient, mockLog, 0)

					Expect(mockTrackerClient.CreateStoryCallCount()).To(Equal(1))
					_, createdStory := mockTrackerClient.CreateStoryArgsForCall(0)
					Expect(createdStory.Name).To(Equal(fmt.Sprintf("%s/%s has failed", failedJob3.FinishedBuild.PipelineName, failedJob3.FinishedBuild.JobName)))
					Expect(createdStory.StoryType).To(Equal("chore"))
					Expect(createdStory.CurrentState).To(Equal("unstarted"))
					Expect(createdStory.Labels[0].Name).To(Equal("broken build"))
					Expect(len(createdStory.Labels)).To(Equal(1))
					Expect(createdStory.Comments[0].Text).To(Equal(fmt.Sprintf("%s/%s", mockServerUrl, failedJob3.FinishedBuild.URL)))
					Expect(len(createdStory.Comments)).To(Equal(1))
					Expect(createdStory.BeforeID).To(Equal(2))
				})
			})

			Context("when a pre-existing story does exist", func() {
				var existingStory tracker.Story
				BeforeEach(func() {
					existingStory = tracker.Story{
						Name: fmt.Sprintf("%s/%s has %s", failedJob3.FinishedBuild.PipelineName, failedJob3.FinishedBuild.JobName, "failed"),
						ID:   2,
					}
					mockTrackerClient.StoriesReturns([]tracker.Story{existingStory}, nil)
				})
				It("adds the failed build as a new comment", func() {
					Groom(groupingStrategy, mockServerUrl, "husbandandwife", 12345, mockTrackerClient, mockConcourseClient, mockLog, 0)
					Expect(mockTrackerClient.ListCommentsCallCount()).To(Equal(1))
					trackerProjectID, storyID := mockTrackerClient.ListCommentsArgsForCall(0)
					Expect(trackerProjectID).To(Equal(12345))
					Expect(storyID).To(Equal(existingStory.ID))
					Expect(mockTrackerClient.AddCommentCallCount()).To(Equal(1))
					_, _, commentText := mockTrackerClient.AddCommentArgsForCall(0)
					Expect(commentText).To(Equal(fmt.Sprintf("%s/%s", mockServerUrl, failedJob3.FinishedBuild.URL)))

				})
			})
		})
	})

})
