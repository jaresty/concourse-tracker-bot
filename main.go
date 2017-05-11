package main

import (
	"log"
	"os"
	"strconv"

	"github.com/jaresty/concourse-tracker-bot/concourse"
	"github.com/jaresty/concourse-tracker-bot/status_groomer"
	"github.com/jaresty/concourse-tracker-bot/tracker"
)

func main() {
	host := os.Getenv("CONCOURSE_HOST")
	team := os.Getenv("CONCOURSE_TEAM")
	trackerToken := os.Getenv("TRACKER_API_TOKEN")
	trackerProjectID, err := strconv.Atoi(os.Getenv("TRACKER_PROJECT_ID"))
	if err != nil {
		panic(err)
	}
	client := tracker.Client{
		APIToken:   trackerToken,
		TrackerAPI: "https://www.pivotaltracker.com/services/v5",
	}

	log := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	status_groomer.Groom(host, team, trackerProjectID, client, concourse.ConcourseClient{}, log, -1)
}
