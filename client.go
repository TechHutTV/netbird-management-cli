// client.go
package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// BaseURL for the NetBird API
const apiBaseURL = "https://api.netbird.io/api"

// Client holds the HTTP client and API token
type Client struct {
	httpClient *http.Client
	apiToken   string
}

// NewClient creates a new API client
func NewClient(apiToken string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   apiToken,
	}
}

// makeRequest is a helper function to create and send API requests
func (c *Client) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := apiBaseURL + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set required headers
	req.Header.Set("Authorization", "Token "+c.apiToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api request failed: %s (status code: %d) %s", resp.Status, resp.StatusCode, string(respBody))
	}

	return resp, nil
}
