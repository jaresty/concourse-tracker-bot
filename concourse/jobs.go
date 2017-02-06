package concourse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Pipeline struct {
	Name   string  `json:"name"`
	URL    string  `json:"url"`
	Paused bool    `json:"paused"`
	Groups []Group `json:"groups"`
}

type Group struct {
	Name string   `json:"name"`
	Jobs []string `json:"jobs"`
}

func GetJobURLs(host string, team string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/teams/%s/pipelines", host, team))
	if err != nil {
		return []string{}, err
	}

	var pipelines []Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipelines); err != nil {
		return []string{}, err
	}

	urls := []string{}
	for _, pipeline := range pipelines {
		if pipeline.Paused {
			continue
		}

		for _, group := range pipeline.Groups {
			for _, job := range group.Jobs {
				urls = append(urls,
					fmt.Sprintf("%s/api/v1/teams/%s/pipelines/%s/jobs/%s", host, team, pipeline.Name, job),
				)
			}
		}
	}

	return urls, nil
}
