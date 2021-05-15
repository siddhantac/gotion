package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var saveResp = ""

func makeRequest(method, path string, body io.Reader, response interface{}) error {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", baseURL, path), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+API_KEY)
	req.Header.Add(API_VERSION_KEY, API_VERSION_VAL)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status: %s", resp.Status)
	}

	if saveResp != "" {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}

		ioutil.WriteFile(saveResp, b, 0644)
		return fmt.Errorf("file written")
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to json unmarshal the response: %w", err)
	}

	return nil
}

type Response struct {
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
	Object     string `json:"object"`
	// Results    interface{} `json:"results"`
}
