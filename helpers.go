// helpers.go
package main

import (
	"fmt"
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
	fmt.Println("  networks                      List all networks")
	fmt.Println()
	fmt.Println("  policy                        List all access control policies")
	fmt.Println()
	fmt.Println("  group                         List all groups")
	fmt.Println()
	fmt.Println("  help                          Show this help message")
}

// printPeerUsage provides specific help for the 'peer' command
func printPeerUsage() {
	fmt.Println("Usage: netbird-manage peer <flag> [arguments]")
	fmt.Println("\nManage network peers.")
	fmt.Println("\nFlags:")
	fmt.Println("  --list                     List all peers")
	fmt.Println("  --inspect <peer-id>          Inspect a single peer")
	fmt.Println("  --remove <peer-id>           Remove a peer by its ID")
	fmt.Println("  --edit <peer-id>             Specify a peer to edit (used with group flags)")
	fmt.Println("  --add-group <group-name>     Add peer to a group (requires --edit)")
	fmt.Println("  --remove-group <group-name>  Remove peer from a group (requires --edit)")
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
