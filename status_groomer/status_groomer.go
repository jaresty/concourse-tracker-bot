package status_groomer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
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

func updateStory(client TrackerClient, host string, buildURL string, trackerProjectID, storyID int, log Logger) error {
	commentText := fmt.Sprintf("%s/%s", host, buildURL)

	comments, err := client.ListComments(trackerProjectID, storyID)
	for _, c := range comments {
		if c.Text == commentText {
			return nil
		}
	}

	log.Println("commenting on previously created story...")
	err = client.AddComment(trackerProjectID, storyID, commentText)
	if err != nil {
		return err
	}
	return nil
}

func createStory(storyName string, log Logger, client TrackerClient, trackerProjectID int, host string, job Job) error {
	log.Println("creating a new story...")

	log.Println("retrieving top of backlog story id...")
	tobStory, err := client.Stories(trackerProjectID, `-type:release state:unstarted`)
	if err != nil {
		return err
	}
	log.Printf("found story %v\n", tobStory[0].ID)

	story, err := client.CreateStory(trackerProjectID, tracker.Story{
		Name:         storyName,
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
		return err
	}

	log.Printf("new story created %v\n", story.ID)
	return nil
}

func findExistingStory(storyName string, stories []tracker.Story) *tracker.Story {
	for _, story := range stories {
		if story.Name == storyName {
			return &story
		}
	}
	return nil
}

func getStoryName(job Job, groupingStrategy map[string]string) string {
	for regex, groupName := range groupingStrategy {
		matched, err := regexp.MatchString(regex, fmt.Sprintf("%s-%s", job.FinishedBuild.PipelineName, job.FinishedBuild.JobName))
		if err == nil && matched {
			return fmt.Sprintf("%s has failed", groupName)
		}
	}
	return fmt.Sprintf("%s/%s has %s", job.FinishedBuild.PipelineName, job.FinishedBuild.JobName, job.FinishedBuild.Status)
}

func handleFailedBuild(groupingStrategy map[string]string, job Job, client TrackerClient, host string, trackerProjectID int, log Logger) error {
	log.Println("build status failed")
	stories, err := client.Stories(trackerProjectID, `-state:accepted label:"broken build"`)
	if err != nil {
		return err
	}

	log.Println("checking for a previously created story...")
	storyName := getStoryName(job, groupingStrategy)

	existingStory := findExistingStory(storyName, stories)
	if existingStory != nil {
		log.Printf("found story %v\n", existingStory.ID)
		err = updateStory(client, host, job.FinishedBuild.URL, trackerProjectID, existingStory.ID, log)
		if err != nil {
			return err
		}
		return nil
	}

	err = createStory(storyName, log, client, trackerProjectID, host, job)
	if err != nil {
		return err
	}
	return nil
}

func processURLs(groupingStrategy map[string]string, urls []string, client TrackerClient, host string, trackerProjectID int, log Logger) error {
	for _, url := range urls {
		log.Printf("checking %s...\n", url)

		res, err := http.Get(url)
		if err != nil {
			return err
		}

		job := Job{}
		if err := json.NewDecoder(res.Body).Decode(&job); err != nil {
			return err
		}

		if job.FinishedBuild.Status == "failed" {
			err := handleFailedBuild(groupingStrategy, job, client, host, trackerProjectID, log)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func Groom(groupingStrategy map[string]string, host, team string, trackerProjectID int, client TrackerClient, concourse ConcourseClient, log Logger, maxIterations int) {
	var currentIteration int
	for {
		log.Println("retrieving jobs...")
		urls, err := concourse.GetJobURLs(host, team)
		if err != nil {
			log.Println(err)
		}

		log.Println("checking for build errors...")
		err = processURLs(groupingStrategy, urls, client, host, trackerProjectID, log)
		if err != nil {
			log.Println(err)
		}

		currentIteration = currentIteration + 1
		if currentIteration > maxIterations && maxIterations >= 0 {
			break
		}
		log.Println("sleeping...")
		time.Sleep(300 * time.Second)
	}
}
