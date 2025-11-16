// client.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client holds the API token and HTTP client
type Client struct {
	Token         string
	ManagementURL string // URL to the NetBird Management API
	HTTPClient    *http.Client
}

// NewClient creates a new NetBird API client
func NewClient(token, managementURL string) *Client {
	return &Client{
		Token:         token,
		ManagementURL: managementURL,
		HTTPClient:    &http.Client{},
	}
}

// makeRequest is a helper function to create and send authenticated API requests
func (c *Client) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := c.ManagementURL + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set authentication and content type headers
	req.Header.Set("Authorization", "Token "+c.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request failed: %v", err)
	}

	// Check for non-success status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var apiError struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}
		// Try to decode the error response from NetBird
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err == nil {
			return nil, fmt.Errorf("api request failed: %d %s (status code: %d) %s", apiError.Code, apiError.Message, resp.StatusCode, resp.Status)
		}
		// Fallback for non-JSON errors
		return nil, fmt.Errorf("api request failed: %s", resp.Status)
	}

	return resp, nil
}
