package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/salsita/go-pivotaltracker/v5/pivotal"
	"github.com/zankich/concourse-tracker-bot/concourse"
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

func main() {
	host := os.Getenv("CONCOURSE_HOST")
	team := os.Getenv("CONCOURSE_TEAM")
	trackerToken := os.Getenv("TRACKER_API_TOKEN")
	trackerProjectID, err := strconv.Atoi(os.Getenv("TRACKER_PROJECT_ID"))
	if err != nil {
		panic(err)
	}
	trackerClient := pivotal.NewClient(trackerToken)

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
				stories, err := trackerClient.Stories.List(trackerProjectID, `-state:accepted label:"broken build"`)
				if err != nil {
					log.Println(err)
					continue
				}

				log.Println("checking for a previously created story...")
				for _, story := range stories {
					if story.Name == fmt.Sprintf("%s/%s has %s", job.FinishedBuild.PipelineName, job.FinishedBuild.JobName, job.FinishedBuild.Status) {
						commentText := fmt.Sprintf("%s/%s", host, job.FinishedBuild.URL)

						log.Printf("found story %v\n", story.Id)

						comments, _, err := trackerClient.Stories.ListComments(trackerProjectID, story.Id)
						for _, c := range comments {
							if c.Text == commentText {
								continue Checks
							}
						}

						log.Println("commenting on previously created story...")
						comment := &pivotal.Comment{Text: commentText}
						_, _, err = trackerClient.Stories.AddComment(trackerProjectID, story.Id, comment)
						if err != nil {
							log.Println(err)
							continue
						}
						continue Checks
					}
				}

				log.Println("creating a new story...")

				log.Println("retrieveing top of backlog story id...")
				tobStory, err := trackerClient.Stories.List(trackerProjectID, `-type:release state:unstarted`)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("found story %v\n", tobStory[0].Id)

				s := &pivotal.StoryRequest{
					Name:  fmt.Sprintf("%s/%s has %s", job.FinishedBuild.PipelineName, job.FinishedBuild.JobName, job.FinishedBuild.Status),
					Type:  pivotal.StoryTypeChore,
					State: pivotal.StoryStateUnstarted,
					Labels: &[]*pivotal.Label{
						&pivotal.Label{Name: "broken build"},
					},
					BeforeId: &tobStory[0].Id,
				}

				story, _, err := trackerClient.Stories.Create(trackerProjectID, s)
				if err != nil {
					log.Println(err)
					continue
				}

				log.Printf("new story created %v\n", story.Id)

				log.Println("commenting on new story...")
				comment := &pivotal.Comment{Text: fmt.Sprintf("%s/%s", host, job.FinishedBuild.URL)}
				_, _, err = trackerClient.Stories.AddComment(trackerProjectID, story.Id, comment)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}

		log.Println("sleeping...")
		time.Sleep(300 * time.Second)
	}
}
