package tracker_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/zankich/concourse-tracker-bot/tracker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	filteredStories = `[
	{
		"created_at":      "2017-01-23T06:02:36Z",
		"updated_at":      "2017-01-23T06:02:36Z",
		"current_state":   "started",
		"description":     "story 2 description",
		"id":              556,
		"kind":            "story",
		"labels":          [],
		"name":            "story 2",
		"owner_ids":       [10199],
		"project_id":      99,
		"requested_by_id": 101,
		"story_type":      "chore",
		"url":             "http://localhost/story/show/556"
	}
]`
	stories = `[
	{
		"created_at":      "2017-01-23T06:02:36Z",
		"updated_at":      "2017-01-23T06:02:36Z",
		"current_state":   "unstarted",
		"description":     "story 1 description",
		"estimate":        2,
		"id":              555,
		"kind":            "story",
		"labels":          [],
		"name":            "story 1 name",
		"owner_ids":       [],
		"project_id":      99,
		"requested_by_id": 101,
		"story_type":      "feature",
		"url":             "http://localhost/story/show/555"
	},
	{
		"created_at":      "2017-01-23T06:02:36Z",
		"updated_at":      "2017-01-23T06:02:36Z",
		"current_state":   "started",
		"description":     "story 2 description",
		"id":              556,
		"kind":            "story",
		"labels":          [],
		"name":            "story 2",
		"owner_ids":       [10199],
		"project_id":      99,
		"requested_by_id": 101,
		"story_type":      "chore",
		"url":             "http://localhost/story/show/556"
	},
	{
		"created_at":    "2017-01-23T06:02:36Z",
		"updated_at":    "2017-01-23T06:02:36Z",
		"current_state": "unstarted",
		"description":   "story 3 description",
		"estimate":      2,
		"id":            557,
		"kind":          "story",
		"labels": [
			{
				"id":          17434625,
				"project_id":  99,
				"kind":        "label",
				"name":        "story 3 label",
				"created_at":  "2017-01-19T05:15:37Z",
				"updated_at":  "2017-01-19T05:15:37Z"
			}
		],
		"name":            "story 3 name",
		"owner_ids":       [],
		"project_id":      99,
		"requested_by_id": 101,
		"story_type":      "feature",
		"url":             "http://localhost/story/show/557"
	}
]`
)

var _ = Describe("Client", func() {
	Describe("Stories", func() {
		var (
			ts     *httptest.Server
			client tracker.Client
		)

		BeforeEach(func() {
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-TrackerToken") != "my-tracker-token" {
					w.WriteHeader(http.StatusUnauthorized)
				}

				if r.Method == "GET" && r.URL.Path == "/projects/99/stories" {
					switch r.URL.Query().Get("filter") {
					case "":
						w.Write([]byte(stories))
						return
					case "state:started type:chore":
						w.Write([]byte(filteredStories))
						return
					default:
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				}

				w.WriteHeader(http.StatusTeapot)
			}))

			client = tracker.Client{
				APIToken:   "my-tracker-token",
				TrackerAPI: ts.URL,
			}
		})

		It("returns the list of stories for the given tracker", func() {
			stories, err := client.Stories(99, "")
			Expect(err).NotTo(HaveOccurred())

			Expect(stories).To(Equal([]tracker.Story{
				{
					CreatedAt:     "2017-01-23T06:02:36Z",
					UpdatedAt:     "2017-01-23T06:02:36Z",
					CurrentState:  "unstarted",
					Description:   "story 1 description",
					Estimate:      2,
					ID:            555,
					Kind:          "story",
					Labels:        []tracker.Label{},
					Name:          "story 1 name",
					OwnerIDs:      []int{},
					ProjectID:     99,
					RequestedByID: 101,
					StoryType:     "feature",
					URL:           "http://localhost/story/show/555",
				},
				{
					CreatedAt:     "2017-01-23T06:02:36Z",
					UpdatedAt:     "2017-01-23T06:02:36Z",
					CurrentState:  "started",
					Description:   "story 2 description",
					ID:            556,
					Kind:          "story",
					Labels:        []tracker.Label{},
					Name:          "story 2",
					OwnerIDs:      []int{10199},
					ProjectID:     99,
					RequestedByID: 101,
					StoryType:     "chore",
					URL:           "http://localhost/story/show/556",
				},
				{
					CreatedAt:    "2017-01-23T06:02:36Z",
					UpdatedAt:    "2017-01-23T06:02:36Z",
					CurrentState: "unstarted",
					Description:  "story 3 description",
					Estimate:     2,
					ID:           557,
					Kind:         "story",
					Labels: []tracker.Label{
						{
							ID:        17434625,
							ProjectID: 99,
							Kind:      "label",
							Name:      "story 3 label",
							CreatedAt: "2017-01-19T05:15:37Z",
							UpdatedAt: "2017-01-19T05:15:37Z",
						},
					},
					Name:          "story 3 name",
					OwnerIDs:      []int{},
					ProjectID:     99,
					RequestedByID: 101,
					StoryType:     "feature",
					URL:           "http://localhost/story/show/557",
				},
			}))
		})

		It("returns the list of stories with the provided filter", func() {
			stories, err := client.Stories(99, "state:started type:chore")
			Expect(err).NotTo(HaveOccurred())

			Expect(stories).To(Equal([]tracker.Story{
				{
					CreatedAt:     "2017-01-23T06:02:36Z",
					UpdatedAt:     "2017-01-23T06:02:36Z",
					CurrentState:  "started",
					Description:   "story 2 description",
					ID:            556,
					Kind:          "story",
					Labels:        []tracker.Label{},
					Name:          "story 2",
					OwnerIDs:      []int{10199},
					ProjectID:     99,
					RequestedByID: 101,
					StoryType:     "chore",
					URL:           "http://localhost/story/show/556",
				},
			}))
		})

		Context("failure cases", func() {
			It("returns an error on a non 200 status code", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
					w.Write([]byte("something bad happened"))
				}))

				client := tracker.Client{
					TrackerAPI: ts.URL,
				}

				_, err := client.Stories(99, "")
				Expect(err).To(MatchError("418 I'm a teapot - something bad happened"))
			})

			It("returns an error when url is malformed", func() {
				client := tracker.Client{
					TrackerAPI: "%%",
				}

				_, err := client.Stories(99, "")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("returns an error when request fails", func() {
				client := tracker.Client{
					TrackerAPI: "http://fakehost.fake",
				}

				_, err := client.Stories(99, "")
				Expect(err).To(MatchError(ContainSubstring("dial tcp: lookup fakehost.fake: no such host")))
			})

			It("returns an error when the json is malformed", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("%%"))
				}))

				client := tracker.Client{
					TrackerAPI: ts.URL,
				}

				_, err := client.Stories(99, "")
				Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
			})
		})

	})
})
