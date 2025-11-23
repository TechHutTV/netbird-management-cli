// Package commands provides CLI command handlers for the NetBird Management CLI
package commands

import (
	"netbird-manage/internal/client"
)

// Service wraps the API client and provides high-level API operations
type Service struct {
	Client *client.Client
}

// NewService creates a new Service with the given client
func NewService(c *client.Client) *Service {
	return &Service{Client: c}
}
