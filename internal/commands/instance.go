package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// HandleInstanceCommand handles instance management operations
// Note: These endpoints don't require authentication
func HandleInstanceCommand(args []string, managementURL string) error {
	cmd := flag.NewFlagSet("instance", flag.ContinueOnError)
	cmd.SetOutput(os.Stderr)
	cmd.Usage = PrintInstanceUsage

	statusFlag := cmd.Bool("status", false, "Get instance setup status")
	setupFlag := cmd.Bool("setup", false, "Set up the instance (create initial admin)")

	emailFlag := cmd.String("email", "", "Admin email (required for --setup)")
	passwordFlag := cmd.String("password", "", "Admin password (required for --setup, min 8 chars)")
	nameFlag := cmd.String("name", "", "Admin name (required for --setup)")
	urlFlag := cmd.String("management-url", "", "Management URL (defaults to NetBird cloud)")
	outputFlag := cmd.String("output", "table", "Output format: table or json")

	if len(args) == 1 {
		PrintInstanceUsage()
		return nil
	}

	if err := cmd.Parse(args[1:]); err != nil {
		return nil
	}

	mgmtURL := managementURL
	if *urlFlag != "" {
		mgmtURL = *urlFlag
	}
	if mgmtURL == "" {
		mgmtURL = "https://api.netbird.io/api"
	}

	// Create a client without auth token for instance operations
	c := client.New("", mgmtURL)

	if *statusFlag {
		return getInstanceStatus(c, *outputFlag)
	}

	if *setupFlag {
		if *emailFlag == "" || *passwordFlag == "" || *nameFlag == "" {
			return fmt.Errorf("--email, --password, and --name are required for --setup")
		}
		return setupInstance(c, *emailFlag, *passwordFlag, *nameFlag)
	}

	PrintInstanceUsage()
	return nil
}

// getInstanceStatus retrieves the instance setup status
func getInstanceStatus(c *client.Client, outputFormat string) error {
	resp, err := c.MakeRequest("GET", "/instance", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var status models.InstanceStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if outputFormat == "json" {
		output, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Println("Instance Status:")
	if status.SetupRequired {
		fmt.Println("  Setup Required: Yes")
		fmt.Println("  Run 'netbird-manage instance --setup --email <email> --password <password> --name <name>' to set up")
	} else {
		fmt.Println("  Setup Required: No")
		fmt.Println("  Instance is already configured")
	}
	return nil
}

// setupInstance creates the initial admin user
func setupInstance(c *client.Client, email, password, name string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	req := models.InstanceSetupRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.MakeRequest("POST", "/setup", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result models.InstanceSetupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Instance set up successfully!\n")
	fmt.Printf("  User ID: %s\n", result.UserID)
	fmt.Printf("  Email:   %s\n", result.Email)
	fmt.Println("\nYou can now connect with:")
	fmt.Println("  netbird-manage connect --token <your-token>")
	return nil
}
