// netbird-manage.go
package main

import (
	"flag"
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

	// The 'connect' command is special: it can create or show the config.
	if command == "connect" {
		if err := handleConnectCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Show help without requiring connection if just the command name is provided
	if len(args) == 1 {
		switch command {
		case "peer":
			printPeerUsage()
			os.Exit(0)
		case "group", "groups":
			printGroupUsage()
			os.Exit(0)
		case "network":
			printNetworkUsage()
			os.Exit(0)
		case "policy":
			printPolicyUsage()
			os.Exit(0)
		case "setup-key":
			printSetupKeyUsage()
			os.Exit(0)
		case "user":
			printUserUsage()
			os.Exit(0)
		case "token":
			printTokenUsage()
			os.Exit(0)
		case "help", "--help":
			printUsage()
			os.Exit(0)
		}
	}

	// For all other commands, load the config first
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Not connected.")
		fmt.Fprintln(os.Stderr, "Please run 'netbird-manage connect --token <your_token>'")
		fmt.Fprintln(os.Stderr, "or set the NETBIRD_API_TOKEN environment variable.")
		os.Exit(1)
	}

	client := NewClient(config.Token, config.ManagementURL)

	// Route the command to the correct handler
	switch command {
	case "peer":
		if err := handlePeersCommand(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "network":
		if err := handleNetworkCommand(client, args); err != nil {
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
	case "setup-key":
		if err := handleSetupKeysCommand(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "user":
		if err := handleUsersCommand(client, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "token":
		if err := handleTokensCommand(client, args); err != nil {
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

// handleConnectCommand parses flags for the connect command
func handleConnectCommand(args []string) error {
	connectCmd := flag.NewFlagSet("connect", flag.ContinueOnError)
	tokenFlag := connectCmd.String("token", "", "Your NetBird API token (Personal Access Token or Service User token)")
	urlFlag := connectCmd.String("management-url", "", "Your self-hosted management URL (optional, defaults to NetBird cloud)")

	if err := connectCmd.Parse(args[1:]); err != nil {
		return nil // flag package will print error
	}

	// If no flags are provided, show status
	if *tokenFlag == "" && *urlFlag == "" {
		return handleConnectStatus()
	}

	// If token is missing
	if *tokenFlag == "" {
		return fmt.Errorf("missing required flag: --token")
	}

	// If URL is missing, use default
	mgmtURL := *urlFlag
	if mgmtURL == "" {
		mgmtURL = defaultCloudURL
	}

	// Test and save the new configuration
	return testAndSaveConfig(*tokenFlag, mgmtURL)
}

// handleConnectStatus shows the current connection status
func handleConnectStatus() error {
	fmt.Println("Checking connection status...")
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Status: Not connected.")
		fmt.Println("Run 'netbird-manage connect --token <token>' to connect.")
		return nil
	}

	fmt.Printf("Status:         Connected\n")
	fmt.Printf("Management URL: %s\n", config.ManagementURL)

	// Try to validate the token
	client := NewClient(config.Token, config.ManagementURL)
	resp, err := client.makeRequest("GET", "/peers", nil)
	if err != nil {
		fmt.Printf("Token Status:   Validation Failed (%v)\n", err)
		return nil
	}
	defer resp.Body.Close()
	fmt.Printf("Token Status:   Valid\n")
	return nil
}
