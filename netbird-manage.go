// netbird-manage.go
package main

import (
	"fmt"
	"os"
)

func main() {
	// Parse command-line arguments
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// The 'connect' command is special: it creates the config, so it doesn't need to load one.
	if command == "connect" {
		if len(args) != 3 || args[1] != "--token" {
			fmt.Fprintln(os.Stderr, "Usage: netbird-manage connect --token <api_token>")
			os.Exit(1)
		}
		token := args[2]
		client := NewClient(token) // Use the provided token
		if err := testAndSaveToken(client, token); err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// For all other commands, load the token first
	apiToken, err := loadToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: API token not found.")
		fmt.Fprintln(os.Stderr, "Please run 'netbird-manage connect --token <your_token>'")
		fmt.Fprintln(os.Stderr, "or set the NETBIRD_API_TOKEN environment variable.")
		os.Exit(1)
	}

	client := NewClient(apiToken)

	// Route the command to the correct handler
	switch command {
	case "peer":
		if err := handlePeersCommand(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "networks":
		if err := handleNetworksCommand(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "policy":
		if err := handlePoliciesCommand(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "group", "groups":
		if err := handleGroupsCommand(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}
