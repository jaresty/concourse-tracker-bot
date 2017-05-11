package status_groomer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jaresty/concourse-tracker-bot/tracker"
)

type Build struct {
	Status       string `json:"status"`
	JobName      string `json:"job_name"`
	URL          string `json:"url"`
	PipelineName string `json:"pipeline_name"`
}

type Job struct {
	FinishedBuild Build `json:"finished_build"`
}

type TrackerClient interface {
	Stories(int, string) ([]tracker.Story, error)
	CreateStory(int, tracker.Story) (tracker.Story, error)
	ListComments(int, int) ([]tracker.Comment, error)
	AddComment(int, int, string) error
}

type ConcourseClient interface {
	GetJobURLs(string, string) ([]string, error)
}

type Logger interface {
	Println(...interface{})
	Printf(string, ...interface{})
}

func Groom(host, team string, trackerProjectID int, client TrackerClient, concourse ConcourseClient, log Logger) {
	for {
		log.Println("retrieving jobs...")
		urls, err := concourse.GetJobURLs(host, team)
		if err != nil {
			log.Println(err)
		}

		log.Println("checking for build errors...")
	Checks:
		for _, url := range urls {
			log.Printf("checking %s...\n", url)

			res, err := http.Get(url)
			if err != nil {
				log.Println(err)
				continue
			}

			job := Job{}
			if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
				log.Println(err)
				continue
			}

			if job.FinishedBuild.Status == "failed" {
				log.Println("build status failed")
				stories, err := client.Stories(trackerProjectID, `-state:accepted label:"broken build"`)
				if err != nil {
					log.Println(err)
					continue
				}

				log.Println("checking for a previously created story...")
				for _, story := range stories {
					if story.Name == fmt.Sprintf("%s/%s has %s", job.FinishedBuild.PipelineName, job.FinishedBuild.JobName, job.FinishedBuild.Status) {
						commentText := fmt.Sprintf("%s/%s", host, job.FinishedBuild.URL)

						log.Printf("found story %v\n", story.ID)

						comments, err := client.ListComments(trackerProjectID, story.ID)
						for _, c := range comments {
							if c.Text == commentText {
								continue Checks
							}
						}

						log.Println("commenting on previously created story...")
						err = client.AddComment(trackerProjectID, story.ID, commentText)
						if err != nil {
							log.Println(err)
							continue
						}
						continue Checks
					}
				}

				log.Println("creating a new story...")

				log.Println("retrieving top of backlog story id...")
				tobStory, err := client.Stories(trackerProjectID, `-type:release state:unstarted`)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("found story %v\n", tobStory[0].ID)

				story, err := client.CreateStory(trackerProjectID, tracker.Story{
					Name:         fmt.Sprintf("%s/%s has %s", job.FinishedBuild.PipelineName, job.FinishedBuild.JobName, job.FinishedBuild.Status),
					StoryType:    "chore",
					CurrentState: "unstarted",
					Labels: []tracker.Label{
						{Name: "broken build"},
					},
					Comments: []tracker.Comment{
						{Text: fmt.Sprintf("%s/%s", host, job.FinishedBuild.URL)},
					},
					BeforeID: tobStory[0].ID,
				})
				if err != nil {
					log.Println(err)
					continue
				}

				log.Printf("new story created %v\n", story.ID)
			}
		}

		log.Println("sleeping...")
		time.Sleep(300 * time.Second)
	}
}
