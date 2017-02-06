package tracker

import (
	"bytes"
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
	Name         string    `json:"name"`
	ID           int       `json:"id,omitempty"`
	CurrentState string    `json:"current_state,omitempty"`
	Labels       []Label   `json:"labels,omitempty"`
	StoryType    string    `json:"story_type,omitempty"`
	BeforeID     int       `json:"before_id,omitempty"`
	Comments     []Comment `json:"comments,omitempty"`
}

type Comment struct {
	Text string `json:"text"`
}

type Label struct {
	Name string `json:"name"`
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

func (c Client) CreateStory(projectID int, input Story) (Story, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(input); err != nil {
		return Story{}, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects/%d/stories", c.TrackerAPI, projectID), body)
	if err != nil {
		return Story{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return Story{}, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return Story{}, fmt.Errorf("%s - %s", resp.Status, string(body))
	}

	var story Story
	if err := json.NewDecoder(resp.Body).Decode(&story); err != nil {
		return Story{}, err
	}

	return story, nil
}

func (c Client) AddComment(projectID int, storyID int, comment string) error {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(Comment{Text: comment}); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects/%d/stories/%d/comments", c.TrackerAPI, projectID, storyID), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s - %s", resp.Status, string(body))
	}

	return nil
}
