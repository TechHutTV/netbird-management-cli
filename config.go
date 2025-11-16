// config.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// configFileName is the name of the config file in the user's home directory
const configFileName = ".netbird-manage.json"
const defaultCloudURL = "https://api.netbird.io/api"

// getConfigPath returns the full path to the configuration file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find user home directory: %v", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

// testAndSaveConfig validates a token by making an API call and saves it if successful
func testAndSaveConfig(token, managementURL string) error {
	fmt.Println("Testing connection to NetBird API at", managementURL)

	// Create a temporary client to test the new credentials
	testClient := NewClient(token, managementURL)

	// Use "GET /api/peers" as the test endpoint
	resp, err := testClient.makeRequest("GET", "/peers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Connection successful. Saving configuration...")
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create the config struct
	config := Config{
		Token:         token,
		ManagementURL: managementURL,
	}

	// Marshal to JSON
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %v", err)
	}

	// Write the token to the config file
	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Printf("Configuration saved successfully to %s\n", configPath)
	return nil
}

// loadConfig loads the API token and URL from the config file or environment variable
func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Try loading from config file first
	configData, err := os.ReadFile(configPath)
	if err == nil {
		var config Config
		if err := json.Unmarshal(configData, &config); err == nil {
			// If URL is somehow empty in file, set default
			if config.ManagementURL == "" {
				config.ManagementURL = defaultCloudURL
			}
			if config.Token != "" {
				return &config, nil
			}
		}
	}

	// If file doesn't exist or is empty, try environment variable
	token := os.Getenv("NETBIRD_API_TOKEN")
	if token != "" {
		// If using env var, assume default cloud URL
		return &Config{
			Token:         token,
			ManagementURL: defaultCloudURL,
		}, nil
	}

	return nil, fmt.Errorf("no token found")
}
