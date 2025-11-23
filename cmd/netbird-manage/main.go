// Package main provides the entry point for the NetBird Management CLI
package main

import (
	"flag"
	"fmt"
	"os"

	"netbird-manage/internal/client"
	"netbird-manage/internal/commands"
	"netbird-manage/internal/config"
	"netbird-manage/internal/helpers"
)

var (
	// debugMode is set to true when --debug flag is provided
	debugMode = false
)

func main() {
	// Parse command-line arguments
	args := os.Args[1:]
	if len(args) == 0 {
		commands.PrintUsage()
		os.Exit(1)
	}

	// Check for global flags (--yes, --debug)
	filteredArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--yes" || arg == "-y" {
			helpers.SkipConfirmation = true
		} else if arg == "--debug" || arg == "-d" {
			debugMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	// Re-check after filtering
	if len(args) == 0 {
		commands.PrintUsage()
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
			commands.PrintPeerUsage()
			os.Exit(0)
		case "group", "groups":
			commands.PrintGroupUsage()
			os.Exit(0)
		case "network":
			commands.PrintNetworkUsage()
			os.Exit(0)
		case "policy":
			commands.PrintPolicyUsage()
			os.Exit(0)
		case "setup-key":
			commands.PrintSetupKeyUsage()
			os.Exit(0)
		case "user":
			commands.PrintUserUsage()
			os.Exit(0)
		case "token":
			commands.PrintTokenUsage()
			os.Exit(0)
		case "route":
			commands.PrintRouteUsage()
			os.Exit(0)
		case "dns":
			commands.PrintDNSUsage()
			os.Exit(0)
		case "posture-check", "posture":
			commands.PrintPostureCheckUsage()
			os.Exit(0)
		case "event", "events":
			commands.PrintEventUsage()
			os.Exit(0)
		case "geo", "geo-location", "location":
			commands.PrintGeoLocationUsage()
			os.Exit(0)
		case "account", "accounts":
			commands.PrintAccountUsage()
			os.Exit(0)
		case "ingress-port", "ingress":
			commands.PrintIngressPortUsage()
			os.Exit(0)
		case "ingress-peer":
			commands.PrintIngressPeerUsage()
			os.Exit(0)
		case "export":
			commands.PrintExportUsage()
			os.Exit(0)
		case "import":
			commands.PrintImportUsage()
			os.Exit(0)
		case "help", "--help":
			commands.PrintUsage()
			os.Exit(0)
		}
	}

	// For all other commands, load the config first
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Not connected.")
		fmt.Fprintln(os.Stderr, "Please run 'netbird-manage connect --token <your_token>'")
		fmt.Fprintln(os.Stderr, "or set the NETBIRD_API_TOKEN environment variable.")
		os.Exit(1)
	}

	c := client.New(cfg.Token, cfg.ManagementURL)
	c.Debug = debugMode

	svc := commands.NewService(c)

	// Route the command to the correct handler
	switch command {
	case "peer":
		if err := svc.HandlePeersCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "network":
		if err := svc.HandleNetworkCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "policy":
		if err := svc.HandlePoliciesCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "group", "groups":
		if err := svc.HandleGroupsCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "setup-key":
		if err := svc.HandleSetupKeysCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "user":
		if err := svc.HandleUsersCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "token":
		if err := svc.HandleTokensCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "route":
		if err := svc.HandleRoutesCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "dns":
		if err := svc.HandleDNSCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "posture-check", "posture":
		if err := svc.HandlePostureChecksCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "event", "events":
		if err := svc.HandleEventsCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "geo", "geo-location", "location":
		if err := svc.HandleGeoLocationsCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "account", "accounts":
		if err := svc.HandleAccountsCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "ingress-port", "ingress":
		if err := svc.HandleIngressPortsCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "ingress-peer":
		if err := svc.HandleIngressPeersCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "export":
		if err := svc.HandleExportCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "import":
		if err := svc.HandleImportCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help":
		commands.PrintUsage()

	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n", command)
		commands.PrintUsage()
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
		mgmtURL = config.DefaultCloudURL
	}

	// Test and save the new configuration
	return config.TestAndSave(*tokenFlag, mgmtURL)
}

// handleConnectStatus shows the current connection status
func handleConnectStatus() error {
	fmt.Println("Checking connection status...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Status: Not connected.")
		fmt.Println("Run 'netbird-manage connect --token <token>' to connect.")
		return nil
	}

	fmt.Printf("Status:         Connected\n")
	fmt.Printf("Management URL: %s\n", cfg.ManagementURL)

	// Try to validate the token
	c := client.New(cfg.Token, cfg.ManagementURL)
	resp, err := c.MakeRequest("GET", "/peers", nil)
	if err != nil {
		fmt.Printf("Token Status:   Validation Failed (%v)\n", err)
		return nil
	}
	defer resp.Body.Close()
	fmt.Printf("Token Status:   Valid\n")
	return nil
}
