package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	API_VERSION_KEY = "Notion-Version"
	API_VERSION_VAL = "2021-05-13"
)
var saveResp = ""

type Response struct {
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
	Object     string `json:"object"`
	// Results    interface{} `json:"results"`
}

func (c *Client) doRequest(method, path string, body interface{}, response interface{}) error {
	req, err := c.makeRequest(method, path, body)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return handleResponse(resp, response)
}

func (c *Client) makeRequest(method, path string, body interface{}) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to json marshal body: %w", err)
	}
	buf := bytes.NewBuffer(b)

	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", c.baseURL, path), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add(API_VERSION_KEY, API_VERSION_VAL)
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func handleResponse(resp *http.Response, response interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	if saveResp != "" {
		ioutil.WriteFile(saveResp, b, 0644)
	}

	if err := json.Unmarshal(b, &response); err != nil {
		return fmt.Errorf("failed to json unmarshal the response: %w", err)
	}

	return nil
}
