package tracker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	TrackerAPI string
	APIToken   string
}
type Story struct {
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	CurrentState  string  `json:"current_state"`
	Description   string  `json:"description"`
	Estimate      int     `json:"estimate"`
	ID            int     `json:"id"`
	Kind          string  `json:"kind"`
	Labels        []Label `json:"labels"`
	Name          string  `json:"name"`
	OwnerIDs      []int   `json:"owner_ids"`
	ProjectID     int     `json:"project_id"`
	RequestedByID int     `json:"requested_by_id"`
	StoryType     string  `json:"story_type"`
	URL           string  `json:"url"`
}
type Label struct {
	ID        int    `json:"id"`
	ProjectID int    `json:"project_id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (c Client) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-TrackerToken", c.APIToken)
	return http.DefaultClient.Do(req)
}

func (c Client) Stories(projectID int, filter string) ([]Story, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%d/stories", c.TrackerAPI, projectID), nil)
	if err != nil {
		return []Story{}, err
	}

	if filter != "" {
		q := req.URL.Query()
		q.Set("filter", filter)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return []Story{}, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return []Story{}, fmt.Errorf("%s - %s", resp.Status, string(body))
	}

	var stories []Story
	if err := json.NewDecoder(resp.Body).Decode(&stories); err != nil {
		return []Story{}, err
	}

	return stories, nil
}
