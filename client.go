// client.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Client holds the API token and HTTP client
type Client struct {
	Token         string
	ManagementURL string // URL to the NetBird Management API
	HTTPClient    *http.Client
	Debug         bool // Enable verbose debug output
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

	// Debug: Log request details
	if c.Debug {
		fmt.Fprintf(os.Stderr, "\n"+cyan("═══ DEBUG: HTTP REQUEST ═══")+"\n")
		fmt.Fprintf(os.Stderr, "%s %s\n", bold(method), url)
	}

	// Read body for debug logging (need to recreate reader after)
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

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

	// Debug: Log request headers (redact token)
	if c.Debug {
		fmt.Fprintf(os.Stderr, "\n"+dim("Headers:")+"\n")
		for key, values := range req.Header {
			value := strings.Join(values, ", ")
			if key == "Authorization" {
				// Redact token for security
				value = "Token " + dim("[REDACTED]")
			}
			fmt.Fprintf(os.Stderr, "  %s: %s\n", key, value)
		}

		// Log request body if present
		if len(bodyBytes) > 0 {
			fmt.Fprintf(os.Stderr, "\n"+dim("Request Body:")+"\n")
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", prettyJSON.String())
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", string(bodyBytes))
			}
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if c.Debug {
			fmt.Fprintf(os.Stderr, "\n"+red("Error: %v")+"\n", err)
		}
		return nil, fmt.Errorf("api request failed: %v", err)
	}

	// Debug: Log response details
	if c.Debug {
		fmt.Fprintf(os.Stderr, "\n"+cyan("═══ DEBUG: HTTP RESPONSE ═══")+"\n")
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Fprintf(os.Stderr, "Status: %s\n", green(resp.Status))
		} else {
			fmt.Fprintf(os.Stderr, "Status: %s\n", red(resp.Status))
		}

		fmt.Fprintf(os.Stderr, "\n"+dim("Headers:")+"\n")
		for key, values := range resp.Header {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", key, strings.Join(values, ", "))
		}
	}

	// Check for non-success status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()

		// Read response body for error and debug logging
		respBody, _ := io.ReadAll(resp.Body)

		if c.Debug && len(respBody) > 0 {
			fmt.Fprintf(os.Stderr, "\n"+dim("Response Body:")+"\n")
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", prettyJSON.String())
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", string(respBody))
			}
		}

		var apiError struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}
		// Try to decode the error response from NetBird
		if err := json.Unmarshal(respBody, &apiError); err == nil {
			return resp, fmt.Errorf("api request failed: %d %s (status code: %d) %s", apiError.Code, apiError.Message, resp.StatusCode, resp.Status)
		}
		// Fallback for non-JSON errors
		return resp, fmt.Errorf("api request failed: %s", resp.Status)
	}

	// Debug: Log successful response body
	if c.Debug {
		respBody, err := io.ReadAll(resp.Body)
		if err == nil && len(respBody) > 0 {
			fmt.Fprintf(os.Stderr, "\n"+dim("Response Body:")+"\n")
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", prettyJSON.String())
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", string(respBody))
			}
			// Recreate response body for caller
			resp.Body = io.NopCloser(bytes.NewReader(respBody))
		}
		fmt.Fprintf(os.Stderr, cyan("═══════════════════════════")+"\n\n")
	}

	return resp, nil
}
