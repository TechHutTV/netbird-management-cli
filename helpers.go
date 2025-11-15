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
	fmt.Println("  setup-key ...                 Manage device registration keys (run 'netbird-manage setup-key' for options)")
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
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                           List all groups")
	fmt.Println("    --filter-name <pattern>        Filter by name (supports wildcards: prod-*)")
	fmt.Println("  --inspect <group-id>             Inspect a specific group")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --create <group-name>            Create a new group")
	fmt.Println("    --peers <id1,id2,...>          (Optional) Add peers on creation")
	fmt.Println()
	fmt.Println("  --delete <group-id>              Delete a group")
	fmt.Println()
	fmt.Println("  --rename <group-id>              Rename a group")
	fmt.Println("    --new-name <new-name>          New name for the group (required)")
	fmt.Println()
	fmt.Println("  --add-peers <group-id>           Add peers to a group (bulk)")
	fmt.Println("    --peers <id1,id2,...>          Comma-separated peer IDs (required)")
	fmt.Println()
	fmt.Println("  --remove-peers <group-id>        Remove peers from a group (bulk)")
	fmt.Println("    --peers <id1,id2,...>          Comma-separated peer IDs (required)")
}

// printNetworkUsage provides specific help for the 'network' command
func printNetworkUsage() {
	fmt.Println("Usage: netbird-manage network <flag> [arguments]")
	fmt.Println("\nManage networks, resources, and routers.")
	fmt.Println("\n=== Network Operations ===")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                              List all networks")
	fmt.Println("    --filter-name <pattern>           Filter by name (supports wildcards: prod-*)")
	fmt.Println("  --inspect <network-id>              Inspect a specific network")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --create <name>                     Create a new network")
	fmt.Println("    --description <desc>              Network description (optional)")
	fmt.Println()
	fmt.Println("  --delete <network-id>               Delete a network")
	fmt.Println()
	fmt.Println("  --rename <network-id>               Rename a network")
	fmt.Println("    --new-name <name>                 New name (required)")
	fmt.Println()
	fmt.Println("  --update <network-id>               Update network description")
	fmt.Println("    --description <desc>              New description (required)")
	fmt.Println()
	fmt.Println("\n=== Resource Operations ===")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list-resources <network-id>       List all resources in a network")
	fmt.Println("  --inspect-resource                  Inspect a specific resource")
	fmt.Println("    --network-id <id>                 Network ID (required)")
	fmt.Println("    --resource-id <id>                Resource ID (required)")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --add-resource <network-id>         Add a resource to a network")
	fmt.Println("    --name <name>                     Resource name (required)")
	fmt.Println("    --address <address>               IP (1.1.1.1), subnet (192.168.0.0/24), or domain (*.example.com) (required)")
	fmt.Println("    --groups <id1,id2,...>            Comma-separated group IDs (required)")
	fmt.Println("    --description <desc>              Resource description (optional)")
	fmt.Println("    --enabled                         Enable resource (default)")
	fmt.Println("    --disabled                        Disable resource")
	fmt.Println()
	fmt.Println("  --update-resource                   Update a resource")
	fmt.Println("    --network-id <id>                 Network ID (required)")
	fmt.Println("    --resource-id <id>                Resource ID (required)")
	fmt.Println("    --name <name>                     New name (optional)")
	fmt.Println("    --address <address>               New address (optional)")
	fmt.Println("    --groups <id1,id2,...>            New groups (optional)")
	fmt.Println("    --description <desc>              New description (optional)")
	fmt.Println("    --enabled/--disabled              Toggle enabled status")
	fmt.Println()
	fmt.Println("  --remove-resource                   Remove a resource")
	fmt.Println("    --network-id <id>                 Network ID (required)")
	fmt.Println("    --resource-id <id>                Resource ID (required)")
	fmt.Println()
	fmt.Println("\n=== Router Operations ===")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list-routers <network-id>         List all routers in a network")
	fmt.Println("  --list-all-routers                  List all routers across all networks")
	fmt.Println("  --inspect-router                    Inspect a specific router")
	fmt.Println("    --network-id <id>                 Network ID (required)")
	fmt.Println("    --router-id <id>                  Router ID (required)")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --add-router <network-id>           Add a router to a network")
	fmt.Println("    --peer <peer-id>                  Use single peer as router (use this OR --peer-groups)")
	fmt.Println("    --peer-groups <id1,id2,...>       Use peer groups as routers (use this OR --peer)")
	fmt.Println("    --metric <1-9999>                 Route metric, lower = higher priority (default: 100)")
	fmt.Println("    --masquerade                      Enable masquerading (NAT)")
	fmt.Println("    --no-masquerade                   Disable masquerading (default)")
	fmt.Println("    --enabled                         Enable router (default)")
	fmt.Println("    --disabled                        Disable router")
	fmt.Println()
	fmt.Println("  --update-router                     Update a router")
	fmt.Println("    --network-id <id>                 Network ID (required)")
	fmt.Println("    --router-id <id>                  Router ID (required)")
	fmt.Println("    --peer <peer-id>                  Change to single peer (optional)")
	fmt.Println("    --peer-groups <id1,id2,...>       Change to peer groups (optional)")
	fmt.Println("    --metric <1-9999>                 Update metric (optional)")
	fmt.Println("    --masquerade/--no-masquerade      Toggle masquerading")
	fmt.Println("    --enabled/--disabled              Toggle enabled status")
	fmt.Println()
	fmt.Println("  --remove-router                     Remove a router")
	fmt.Println("    --network-id <id>                 Network ID (required)")
	fmt.Println("    --router-id <id>                  Router ID (required)")
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

// validateNetworkAddress validates network resource addresses
// Accepts: IP (1.1.1.1 or 1.1.1.1/32), subnet (192.168.0.0/24), or domain (example.com, *.example.com)
func validateNetworkAddress(address string) error {
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

// splitCommaList splits a comma-separated string into a slice of trimmed strings
func splitCommaList(input string) []string {
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
