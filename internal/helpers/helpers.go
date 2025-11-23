// Package helpers provides utility functions for the CLI
package helpers

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	// netbirdCGNATRange is the NetBird CGNAT range (100.64.0.0/10)
	// Parsed once at initialization for performance
	netbirdCGNATRange *net.IPNet

	// SkipConfirmation is set to true when --yes flag is provided
	SkipConfirmation = false
)

func init() {
	_, netbirdCGNATRange, _ = net.ParseCIDR("100.64.0.0/10")
}

// FormatOS formats OS string for display
func FormatOS(osStr string) string {
	if strings.Contains(osStr, "Darwin") {
		return "macOS"
	}
	if strings.Contains(osStr, "Linux") {
		return "Linux"
	}
	if strings.Contains(osStr, "Windows") {
		return "Windows"
	}
	return osStr
}

// ValidateNetBirdIP validates that an IP address is within the NetBird CGNAT range
// NetBird uses 100.64.0.0/10 (100.64.0.0 to 100.127.255.255)
func ValidateNetBirdIP(ipStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipStr)
	}

	if !netbirdCGNATRange.Contains(ip) {
		return fmt.Errorf("IP address %s is outside NetBird's allowed range (100.64.0.0/10)", ipStr)
	}

	return nil
}

// ValidateNetworkAddress validates network resource addresses
// Accepts: IP (1.1.1.1 or 1.1.1.1/32), subnet (192.168.0.0/24), or domain (example.com, *.example.com)
func ValidateNetworkAddress(address string) error {
	// Check if it's a CIDR notation (IP with /prefix)
	if strings.Contains(address, "/") {
		_, _, err := net.ParseCIDR(address)
		if err != nil {
			return fmt.Errorf("invalid CIDR notation: %s", address)
		}
		return nil
	}

	// Check if it's a plain IP address
	if ip := net.ParseIP(address); ip != nil {
		return nil
	}

	// Must be a domain name (supports wildcards like *.example.com)
	// Simple validation: check for valid domain characters
	if len(address) == 0 {
		return fmt.Errorf("address cannot be empty")
	}

	// Domain can contain: letters, numbers, hyphens, dots, and wildcards (*)
	// Basic validation - more permissive to allow wildcard domains
	for _, char := range address {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '-' || char == '*') {
			return fmt.Errorf("invalid domain name: %s (contains invalid character: %c)", address, char)
		}
	}

	return nil
}

// MatchesPattern checks if a string matches a glob-style pattern (* wildcard)
func MatchesPattern(str, pattern string) bool {
	// If no wildcard, do exact match
	if !strings.Contains(pattern, "*") {
		return strings.Contains(strings.ToLower(str), strings.ToLower(pattern))
	}

	// Simple glob matching - convert * to regex-like matching
	pattern = strings.ToLower(pattern)
	str = strings.ToLower(str)

	// Split on * to get parts that must be present
	parts := strings.Split(pattern, "*")

	// Check if string contains all parts in order
	currentPos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}

		idx := strings.Index(str[currentPos:], part)
		if idx == -1 {
			return false
		}

		// For first part, must be at beginning if pattern doesn't start with *
		if i == 0 && pattern[0] != '*' && idx != 0 {
			return false
		}

		currentPos += idx + len(part)
	}

	// If pattern doesn't end with *, ensure we matched to the end
	if !strings.HasSuffix(pattern, "*") {
		return currentPos == len(str)
	}

	return true
}

// SplitCommaList splits a comma-separated string into a slice of trimmed strings
func SplitCommaList(input string) []string {
	if input == "" {
		return []string{}
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// ReorderArgsForFlags reorders command arguments to put flags before positional arguments.
// This allows users to write: command file.yml --flag
// instead of requiring: command --flag file.yml
// Go's flag package requires flags before positional arguments.
func ReorderArgsForFlags(args []string) []string {
	if len(args) == 0 {
		return args
	}

	var flags []string
	var positional []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
		} else {
			positional = append(positional, arg)
		}
	}

	// Return flags first, then positional arguments
	return append(flags, positional...)
}

// ConfirmSingleDeletion shows resource details and asks for Y/N confirmation
// Returns true if user confirms, false otherwise
func ConfirmSingleDeletion(resourceType, resourceName, resourceID string, details map[string]string) bool {
	// Skip confirmation if --yes flag was provided
	if SkipConfirmation {
		return true
	}

	fmt.Fprintf(os.Stderr, "\nAbout to remove %s:\n", resourceType)

	// Always show name and ID first if available
	if resourceName != "" {
		fmt.Fprintf(os.Stderr, "  Name:      %s\n", resourceName)
	}
	if resourceID != "" {
		fmt.Fprintf(os.Stderr, "  ID:        %s\n", resourceID)
	}

	// Show additional details in consistent order
	for key, value := range details {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", key+":", value)
	}

	fmt.Fprintf(os.Stderr, "\nThis action cannot be undone. Continue? [y/N]: ")

	return ReadYesNo()
}

// ConfirmBulkDeletion shows a summary list and requires typing to confirm
// Returns true if user types the correct confirmation text
func ConfirmBulkDeletion(resourceType string, items []string, count int) bool {
	// Skip confirmation if --yes flag was provided
	if SkipConfirmation {
		return true
	}

	fmt.Fprintf(os.Stderr, "\nThis will delete %d %s:\n", count, resourceType)

	// Show up to 10 items in the list
	maxShow := 10
	for i, item := range items {
		if i >= maxShow {
			fmt.Fprintf(os.Stderr, "  ... and %d more\n", count-maxShow)
			break
		}
		fmt.Fprintf(os.Stderr, "  - %s\n", item)
	}

	// Generate confirmation text
	confirmText := fmt.Sprintf("delete %d %s", count, resourceType)

	fmt.Fprintf(os.Stderr, "\nType '%s' to confirm:\n> ", confirmText)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(input)

	if input == confirmText {
		return true
	}

	fmt.Fprintln(os.Stderr, "Operation cancelled")
	return false
}

// ReadYesNo reads a y/N response from the user
// Returns true if user types 'y' or 'yes' (case insensitive)
func ReadYesNo() bool {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		return true
	}

	fmt.Fprintln(os.Stderr, "Operation cancelled")
	return false
}
