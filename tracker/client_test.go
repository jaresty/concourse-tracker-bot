package tracker_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"

	"github.com/zankich/concourse-tracker-bot/tracker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	createStoryRequest = `{
	"name": "my story",
	"story_type": "chore",
	"current_state": "unstarted",
	"labels": [{
		"name": "my label"
	}],
	"before_id": 100,
	"comments": [{
		"text": "my comment"
	}]
}`

	createStoryResponse = `{
	"current_state": "unstarted",
	"id": 1098,
	"labels": [{
		"name": "my label"
	}],
	"name": "my story",
	"story_type": "chore"
}`

	filteredStories = `[{
	"current_state": "started",
	"id": 556,
	"labels": [],
	"name": "story 2",
	"story_type": "chore"
}]`
	stories = `[
  {
  	"current_state": "unstarted",
  	"id": 555,
  	"labels": [],
  	"name": "story 1 name",
  	"story_type": "feature"
  },
  {
  	"current_state": "started",
  	"id": 556,
  	"labels": [],
  	"name": "story 2",
  	"story_type": "chore"
  },
  {
  	"current_state": "unstarted",
  	"id": 557,
  	"labels": [{
  		"name": "story 3 label"
  	}],
  	"name": "story 3 name",
  	"story_type": "feature"
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
					CurrentState: "unstarted",
					ID:           555,
					Labels:       []tracker.Label{},
					Name:         "story 1 name",
					StoryType:    "feature",
				},
				{
					CurrentState: "started",
					ID:           556,
					Labels:       []tracker.Label{},
					Name:         "story 2",
					StoryType:    "chore",
				},
				{
					CurrentState: "unstarted",
					ID:           557,
					Labels: []tracker.Label{
						{
							Name: "story 3 label",
						},
					},
					Name:      "story 3 name",
					StoryType: "feature",
				},
			}))
		})

		It("returns the list of stories with the provided filter", func() {
			stories, err := client.Stories(99, "state:started type:chore")
			Expect(err).NotTo(HaveOccurred())

			Expect(stories).To(Equal([]tracker.Story{
				{
					CurrentState: "started",
					ID:           556,
					Labels:       []tracker.Label{},
					Name:         "story 2",
					StoryType:    "chore",
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

	Describe("CreateStory", func() {
		var (
			ts     *httptest.Server
			client tracker.Client
		)

		BeforeEach(func() {
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-TrackerToken") != "my-tracker-token" {
					w.WriteHeader(http.StatusUnauthorized)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					w.WriteHeader(http.StatusNoContent)
				}

				if r.Method == "POST" && r.URL.Path == "/projects/99/stories" {
					body, err := ioutil.ReadAll(r.Body)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
					}

					var got interface{}
					if err := json.Unmarshal(body, &got); err != nil {
						panic(err)
					}

					var want interface{}
					if err := json.Unmarshal([]byte(createStoryRequest), &want); err != nil {
						panic(err)
					}

					if reflect.DeepEqual(got, want) {
						w.Write([]byte(createStoryResponse))
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("got %s - want %s", string(body), createStoryRequest)))
					return
				}

				w.WriteHeader(http.StatusTeapot)
			}))

			client = tracker.Client{
				APIToken:   "my-tracker-token",
				TrackerAPI: ts.URL,
			}
		})

		It("creates a new story in the specified tracker backlog", func() {
			input := tracker.Story{
				Name:         "my story",
				StoryType:    "chore",
				CurrentState: "unstarted",
				Labels:       []tracker.Label{{Name: "my label"}},
				BeforeID:     100,
				Comments:     []tracker.Comment{{Text: "my comment"}},
			}

			story, err := client.CreateStory(99, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(story).To(Equal(tracker.Story{
				CurrentState: "unstarted",
				ID:           1098,
				Labels:       []tracker.Label{{Name: "my label"}},
				Name:         "my story",
				StoryType:    "chore",
			}))
		})

		Context("failure cases", func() {
			It("returns an error when the return code is not a 200", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
					w.Write([]byte("something bad happened"))
				}))

				client := tracker.Client{
					TrackerAPI: ts.URL,
				}

				_, err := client.CreateStory(99, tracker.Story{})
				Expect(err).To(MatchError("418 I'm a teapot - something bad happened"))
			})

			It("returns an error when the request fails", func() {
				client := tracker.Client{
					TrackerAPI: "http://fakehost.fake",
				}

				_, err := client.CreateStory(99, tracker.Story{})
				Expect(err).To(MatchError(ContainSubstring("dial tcp: lookup fakehost.fake: no such host")))
			})

			It("returns an error when the json is malformed", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("%%"))
				}))

				client := tracker.Client{
					TrackerAPI: ts.URL,
				}

				_, err := client.CreateStory(99, tracker.Story{})
				Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
			})

			It("returns an error when url is malformed", func() {
				client := tracker.Client{
					TrackerAPI: "%%",
				}

				_, err := client.CreateStory(99, tracker.Story{})
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})
	})

	Describe("AddComment", func() {
		var (
			ts     *httptest.Server
			client tracker.Client
		)

		BeforeEach(func() {
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-TrackerToken") != "my-tracker-token" {
					w.WriteHeader(http.StatusUnauthorized)
				}

				if r.Header.Get("Content-Type") != "application/json" {
					w.WriteHeader(http.StatusNoContent)
				}

				if r.Method == "POST" && r.URL.Path == "/projects/99/stories/101/comments" {
					body, err := ioutil.ReadAll(r.Body)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
					}

					commentJSON := []byte(`{"text":"my comment"}`)

					var got interface{}
					if err := json.Unmarshal(body, &got); err != nil {
						panic(err)
					}

					var want interface{}
					if err := json.Unmarshal(commentJSON, &want); err != nil {
						panic(err)
					}

					if reflect.DeepEqual(got, want) {
						w.Write(commentJSON)
						return
					}
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("got %s - want %s", string(body), string(commentJSON))))
					return
				}

				w.WriteHeader(http.StatusTeapot)
			}))

			client = tracker.Client{
				APIToken:   "my-tracker-token",
				TrackerAPI: ts.URL,
			}
		})

		It("adds a comment to an existing story", func() {
			err := client.AddComment(99, 101, "my comment")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("failure cases", func() {
			It("returns an error when the return code is not a 200", func() {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
					w.Write([]byte("something bad happened"))
				}))

				client := tracker.Client{
					TrackerAPI: ts.URL,
				}

				err := client.AddComment(99, 101, "")
				Expect(err).To(MatchError("418 I'm a teapot - something bad happened"))
			})

			It("returns an error when the request fails", func() {
				client := tracker.Client{
					TrackerAPI: "http://fakehost.fake",
				}

				err := client.AddComment(99, 101, "")
				Expect(err).To(MatchError(ContainSubstring("dial tcp: lookup fakehost.fake: no such host")))
			})

			It("returns an error when url is malformed", func() {
				client := tracker.Client{
					TrackerAPI: "%%",
				}

				err := client.AddComment(99, 101, "")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})
	})
})
