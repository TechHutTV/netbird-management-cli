// config.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Environment variable that holds the API token
const apiTokenEnvVar = "NETBIRD_API_TOKEN"

// configFileName is the name of the config file stored in the user's home dir
const configFileName = ".netbird-manage.conf"

// Config struct for saving the token
type Config struct {
	APIToken string `json:"api_token"`
}

// testAndSaveToken tests a token by fetching the peer list and saves it to the config
func testAndSaveToken(client *Client, token string) error {
	// Use "GET /api/peers" as a simple check, as "GET /api/users/current" can be problematic for service users.
	// This endpoint is used in the quickstart guide.
	resp, err := client.makeRequest("GET", "/peers", nil)
	if err != nil {
		return fmt.Errorf("invalid token: %v", err)
	}
	defer resp.Body.Close()

	// We just need to know the request was successful (200 OK)
	// We can try to decode to give the user some feedback
	var peers []Peer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		// io.EOF is fine if the list is empty, but any other decode error is bad
		if err != io.EOF && err != nil {
			return fmt.Errorf("failed to decode peers response during test: %v", err)
		}
	}

	fmt.Printf("Successfully authenticated token! Found %d peers in your network.\n", len(peers))

	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find home directory: %v", err)
	}
	configPath := filepath.Join(home, configFileName)

	// Create config struct and marshal to JSON
	config := Config{APIToken: token}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write to config file with secure permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to save token to %s: %v", configPath, err)
	}

	fmt.Printf("Token saved to %s\n", configPath)
	return nil
}

// loadToken loads the API token from the config file or environment variable
func loadToken() (string, error) {
	// 1. Try loading from config file
	home, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(home, configFileName)
		data, err := os.ReadFile(configPath)
		if err == nil {
			var config Config
			if json.Unmarshal(data, &config) == nil && config.APIToken != "" {
				return config.APIToken, nil
			}
		}
	}

	// 2. Fallback to environment variable
	token := os.Getenv(apiTokenEnvVar)
	if token != "" {
		return token, nil
	}

	// 3. If no token is found, return an error
	return "", fmt.Errorf("api token not found")
}
