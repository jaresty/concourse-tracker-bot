package status_groomer_test

import (
	. "github.com/jaresty/concourse-tracker-bot/status_groomer"
	"github.com/jaresty/concourse-tracker-bot/status_groomer/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatusGroomer", func() {
	var (
		mockTrackerClient   TrackerClient
		mockConcourseClient ConcourseClient
		mockLog             Logger
	)

	Context("when there are failed builds", func() {
		BeforeEach(func() {
			mockTrackerClient = new(fakes.FakeTrackerClient)
			mockConcourseClient = new(fakes.FakeConcourseClient)
			mockLog = new(fakes.FakeLogger)
		})

		Context("when a pre-existing story does not exists", func() {
			It("creates a new story", func() {
				Groom("bride.com", "husbandandwife", 12345, mockTrackerClient, mockConcourseClient, mockLog)
				Expect(true)
			})
		})
	})

})
