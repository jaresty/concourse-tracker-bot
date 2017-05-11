package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/jaresty/concourse-tracker-bot/concourse"
	"github.com/jaresty/concourse-tracker-bot/parser"
	"github.com/jaresty/concourse-tracker-bot/status_groomer"
	"github.com/jaresty/concourse-tracker-bot/tracker"
	"gopkg.in/yaml.v2"
)

func parse(groupConfigFile string) map[string]string {
	data, err := ioutil.ReadFile(groupConfigFile)
	if err != nil {
		panic(err)
	}
	c := make(map[string][]string)
	err = yaml.Unmarshal(data, c)
	if err != nil {
		panic(err)
	}
	return parser.Parse(c)
}

func main() {
	var groupConfigFile string
	flag.StringVar(&groupConfigFile, "group-config-file", "", "path to the group config file")
	flag.Parse()

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
	status_groomer.Groom(parse(groupConfigFile), host, team, trackerProjectID, client, concourse.ConcourseClient{}, log, -1)
}
