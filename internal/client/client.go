// Package client provides the HTTP client for NetBird API requests
package client

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

// New creates a new NetBird API client
func New(token, managementURL string) *Client {
	return &Client{
		Token:         token,
		ManagementURL: managementURL,
		HTTPClient:    &http.Client{},
	}
}

// MakeRequest is a helper function to create and send authenticated API requests
func (c *Client) MakeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := c.ManagementURL + endpoint

	// Debug: Log request details
	if c.Debug {
		fmt.Fprintf(os.Stderr, "\n=== DEBUG: HTTP REQUEST ===\n")
		fmt.Fprintf(os.Stderr, "%s %s\n", method, url)
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
		fmt.Fprintf(os.Stderr, "\nHeaders:\n")
		for key, values := range req.Header {
			value := strings.Join(values, ", ")
			if key == "Authorization" {
				// Redact token for security
				value = "Token [REDACTED]"
			}
			fmt.Fprintf(os.Stderr, "  %s: %s\n", key, value)
		}

		// Log request body if present (redact secrets first)
		if len(bodyBytes) > 0 {
			fmt.Fprintf(os.Stderr, "\nRequest Body:\n")
			redacted := redactSensitiveJSON(bodyBytes)
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, redacted, "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", prettyJSON.String())
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", string(redacted))
			}
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if c.Debug {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		}
		return nil, fmt.Errorf("api request failed: %v", err)
	}

	// Debug: Log response details
	if c.Debug {
		fmt.Fprintf(os.Stderr, "\n=== DEBUG: HTTP RESPONSE ===\n")
		fmt.Fprintf(os.Stderr, "Status: %s\n", resp.Status)

		fmt.Fprintf(os.Stderr, "\nHeaders:\n")
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
			fmt.Fprintf(os.Stderr, "\nResponse Body:\n")
			redacted := redactSensitiveJSON(respBody)
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, redacted, "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", prettyJSON.String())
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", string(redacted))
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
			fmt.Fprintf(os.Stderr, "\nResponse Body:\n")
			redacted := redactSensitiveJSON(respBody)
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, redacted, "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", prettyJSON.String())
			} else {
				fmt.Fprintf(os.Stderr, "%s\n", string(redacted))
			}
			// Recreate response body for caller (with the ORIGINAL bytes, not redacted).
			resp.Body = io.NopCloser(bytes.NewReader(respBody))
		}
		fmt.Fprintf(os.Stderr, "===========================\n\n")
	}

	return resp, nil
}

// sensitiveJSONKeys names JSON keys whose string values should be redacted
// whenever they appear in a debug dump. Covers NetBird auth secrets plus
// generic credential fields any future resource might add.
var sensitiveJSONKeys = map[string]bool{
	"password":      true,
	"pin":           true,
	"token":         true,
	"api_key":       true,
	"apikey":        true,
	"secret":        true,
	"access_token":  true,
	"refresh_token": true,
	"client_secret": true,
}

// redactSensitiveJSON walks a JSON document and replaces values of known
// sensitive keys with "[REDACTED]". Also masks the `value` field of
// `header_auths` entries (reverse-proxy header-auth shared secret) since it's
// context-sensitive rather than name-sensitive.
//
// If the input isn't valid JSON, returns the original bytes unchanged —
// the caller will fall through to raw printing.
func redactSensitiveJSON(b []byte) []byte {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return b
	}
	v = redactWalk(v, "")
	out, err := json.Marshal(v)
	if err != nil {
		return b
	}
	return out
}

// redactWalk recursively redacts sensitive values. parentKey is the key
// name of the containing map entry (or "" at the root / inside a slice) so
// context-sensitive rules (e.g. `value` inside `header_auths`) work.
func redactWalk(v interface{}, parentKey string) interface{} {
	switch n := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(n))
		for k, child := range n {
			if sensitiveJSONKeys[k] {
				if _, isString := child.(string); isString {
					out[k] = "[REDACTED]"
					continue
				}
			}
			// Context rule: header_auths[].value is a shared secret.
			if parentKey == "header_auths" && k == "value" {
				if _, isString := child.(string); isString {
					out[k] = "[REDACTED]"
					continue
				}
			}
			out[k] = redactWalk(child, k)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(n))
		for i, child := range n {
			out[i] = redactWalk(child, parentKey)
		}
		return out
	default:
		return n
	}
}
