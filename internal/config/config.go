// Package config handles configuration file management for the CLI
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// configFileName is the name of the config file in the user's home directory
const configFileName = ".netbird-manage.json"

// DefaultCloudURL is the default NetBird cloud API URL
const DefaultCloudURL = "https://api.netbird.io/api"

// GetConfigPath returns the full path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find user home directory: %v", err)
	}
	return filepath.Join(homeDir, configFileName), nil
}

// TestAndSave validates a token by making an API call and saves it if successful
func TestAndSave(token, managementURL string) error {
	fmt.Println("Testing connection to NetBird API at", managementURL)

	// Create a temporary client to test the new credentials
	testClient := client.New(token, managementURL)

	// Use "GET /api/peers" as the test endpoint
	resp, err := testClient.MakeRequest("GET", "/peers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Connection successful. Saving configuration...")
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create the config struct
	cfg := models.Config{
		Token:         token,
		ManagementURL: managementURL,
	}

	// Marshal to JSON
	configData, err := json.MarshalIndent(cfg, "", "  ")
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

// Load loads the API token and URL from the config file or environment variable
func Load() (*models.Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Try loading from config file first
	configData, err := os.ReadFile(configPath)
	if err == nil {
		var cfg models.Config
		if err := json.Unmarshal(configData, &cfg); err == nil {
			// If URL is somehow empty in file, set default
			if cfg.ManagementURL == "" {
				cfg.ManagementURL = DefaultCloudURL
			}
			if cfg.Token != "" {
				return &cfg, nil
			}
		}
	}

	// If file doesn't exist or is empty, try environment variable
	token := os.Getenv("NETBIRD_API_TOKEN")
	if token != "" {
		// If using env var, assume default cloud URL
		return &models.Config{
			Token:         token,
			ManagementURL: DefaultCloudURL,
		}, nil
	}

	return nil, fmt.Errorf("no token found")
}
