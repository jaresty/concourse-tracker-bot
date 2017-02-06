package concourse_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/zankich/concourse-tracker-bot/concourse"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	pipelines = `[
  {
    "name": "p1",
    "url": "/teams/main/pipelines/p1",
    "paused": false,
    "public": true,
    "groups": [
      {
        "name": "g1",
        "jobs": [
			"g1j1",
			"g1j2"
        ]
      },
      {
        "name": "g2",
        "jobs": [
			"g2j1",
			"g2j2"
        ]
      }
    ],
    "team_name": "main"
  },
  {
    "name": "p2",
    "url": "/teams/main/pipelines/p2",
    "paused": false,
    "public": true,
    "groups": [
      {
        "name": "g1",
        "jobs": [
			"g1j1",
			"g1j2"
        ]
      }
    ],
    "team_name": "main"
  },
  {
    "name": "p3",
    "url": "/teams/main/pipelines/p3",
    "paused": true,
    "public": true,
    "groups": [
      {
        "name": "g1",
        "jobs": [
			"g1j1",
			"g1j2"
        ]
      }
    ],
    "team_name": "main"
  }
]`
)

var _ = Describe("GetJobURLs", func() {
	It("returns a list of public jobs for a given team", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/api/v1/teams/main/pipelines" {
				w.Write([]byte(pipelines))
				return
			}

			w.WriteHeader(http.StatusTeapot)
		}))
		defer ts.Close()

		urls, err := concourse.GetJobURLs(ts.URL, "main")
		Expect(err).NotTo(HaveOccurred())

		Expect(urls).To(Equal([]string{
			ts.URL + "/api/v1/teams/main/pipelines/p1/jobs/g1j1",
			ts.URL + "/api/v1/teams/main/pipelines/p1/jobs/g1j2",
			ts.URL + "/api/v1/teams/main/pipelines/p1/jobs/g2j1",
			ts.URL + "/api/v1/teams/main/pipelines/p1/jobs/g2j2",
			ts.URL + "/api/v1/teams/main/pipelines/p2/jobs/g1j1",
			ts.URL + "/api/v1/teams/main/pipelines/p2/jobs/g1j2",
		}))
	})

	Context("failure cases", func() {
		It("returns an error when the host is bad", func() {
			_, err := concourse.GetJobURLs("%%%%%", "main")

			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("returns an error when the json is bad", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && r.URL.Path == "/api/v1/teams/main/pipelines" {
					w.Write([]byte("%%%%%%%%"))
					return
				}

				w.WriteHeader(http.StatusTeapot)
			}))
			defer ts.Close()

			_, err := concourse.GetJobURLs(ts.URL, "main")
			Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
		})
	})
})
