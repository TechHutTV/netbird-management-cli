// helpers.go
package main

import (
	"fmt"
	"net"
	"strings"
)

func printUsage() {
	fmt.Println("NetBird Management CLI")
	fmt.Println("----------------------")
	fmt.Println("A simple tool to manage your NetBird network via the API.")
	fmt.Println("\nUsage:")
	fmt.Println("  netbird-manage <command> [arguments]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  connect                       Check current connection status")
	fmt.Println("  connect [flags]               Connect and save your API token")
	fmt.Println("    --token <key>               (Required) Your NetBird API token")
	fmt.Println("    --management-url <url>      (Optional) Your self-hosted management URL")
	fmt.Println()
	fmt.Println("  peer ...                      Manage peers (run 'netbird-manage peer' for options)")
	fmt.Println()
	fmt.Println("  group ...                     Manage groups (run 'netbird-manage group' for options)")
	fmt.Println()
	fmt.Println("  network ...                   Manage networks (run 'netbird-manage network' for options)")
	fmt.Println()
	fmt.Println("  policy ...                    Manage access control policies (run 'netbird-manage policy' for options)")
	fmt.Println()
	fmt.Println("  help                          Show this help message")
}

// printPeerUsage provides specific help for the 'peer' command
func printPeerUsage() {
	fmt.Println("Usage: netbird-manage peer <flag> [arguments]")
	fmt.Println("\nManage network peers.")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                            List all peers")
	fmt.Println("    --filter-name <pattern>         Filter by name (supports wildcards: ubuntu*)")
	fmt.Println("    --filter-ip <pattern>           Filter by IP address pattern")
	fmt.Println("  --inspect <peer-id>               Inspect a single peer")
	fmt.Println("  --accessible-peers <peer-id>      List peers accessible from the specified peer")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --remove <peer-id>                Remove a peer from your network")
	fmt.Println()
	fmt.Println("  --edit <peer-id>                  Edit peer group membership")
	fmt.Println("    --add-group <group-id>          Add peer to a group (requires --edit)")
	fmt.Println("    --remove-group <group-id>       Remove peer from a group (requires --edit)")
	fmt.Println()
	fmt.Println("  --update <peer-id>                Update peer settings")
	fmt.Println("    --rename <new-name>             Change peer name")
	fmt.Println("    --ssh-enabled <true|false>      Enable/disable SSH access")
	fmt.Println("    --login-expiration <true|false> Enable/disable login expiration")
	fmt.Println("    --inactivity-expiration <true|false> Enable/disable inactivity expiration")
	fmt.Println("    --approval-required <true|false> Require approval (cloud-only)")
	fmt.Println("    --ip <ip-address>               Set IP (must be in 100.64.0.0/10 range)")
}

// printGroupUsage provides specific help for the 'group' command
func printGroupUsage() {
	fmt.Println("Usage: netbird-manage group <flag> [arguments]")
	fmt.Println("\nManage network groups.")
	fmt.Println("\nFlags:")
	fmt.Println("  --list                       List all groups")
}

// printNetworkUsage provides specific help for the 'network' command
func printNetworkUsage() {
	fmt.Println("Usage: netbird-manage network <flag> [arguments]")
	fmt.Println("\nManage network networks.")
	fmt.Println("\nFlags:")
	fmt.Println("  --list                       List all networks")
}

// printPolicyUsage provides specific help for the 'policy' command
func printPolicyUsage() {
	fmt.Println("Usage: netbird-manage policy <flag> [arguments]")
	fmt.Println("\nManage access control policies.")
	fmt.Println("\nFlags:")
	fmt.Println("  --list                       List all policies")
}

func formatOS(osStr string) string {
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

// validateNetBirdIP validates that an IP address is within the NetBird CGNAT range
// NetBird uses 100.64.0.0/10 (100.64.0.0 to 100.127.255.255)
func validateNetBirdIP(ipStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// NetBird CGNAT range: 100.64.0.0/10
	_, cgnatRange, _ := net.ParseCIDR("100.64.0.0/10")
	if !cgnatRange.Contains(ip) {
		return fmt.Errorf("IP address %s is outside NetBird's allowed range (100.64.0.0/10)", ipStr)
	}

	return nil
}

// matchesPattern checks if a string matches a glob-style pattern (* wildcard)
func matchesPattern(str, pattern string) bool {
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
