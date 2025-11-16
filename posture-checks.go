// posture-checks.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// handlePostureChecksCommand routes posture check-related commands
func handlePostureChecksCommand(client *Client, args []string) error {
	postureCmd := flag.NewFlagSet("posture-check", flag.ContinueOnError)
	postureCmd.SetOutput(os.Stderr)
	postureCmd.Usage = printPostureCheckUsage

	// Query flags
	listFlag := postureCmd.Bool("list", false, "List all posture checks")
	inspectFlag := postureCmd.String("inspect", "", "Inspect a posture check by ID")
	filterName := postureCmd.String("filter-name", "", "Filter by name pattern")
	filterType := postureCmd.String("filter-type", "", "Filter by check type")

	// Create flags
	createFlag := postureCmd.String("create", "", "Create a new posture check with the given name")
	descriptionFlag := postureCmd.String("description", "", "Posture check description")
	checkTypeFlag := postureCmd.String("type", "", "Check type: nb-version, os-version, geo-location, network-range, process")

	// Check-specific flags (defined but accessed via flag lookups in buildCheckDefinition)
	postureCmd.String("min-version", "", "Minimum NetBird version (for nb-version)")
	postureCmd.String("os", "", "OS type: android, darwin, ios, linux, windows (for os-version)")
	postureCmd.String("min-os-version", "", "Minimum OS version (for os-version)")
	postureCmd.String("min-kernel", "", "Minimum kernel version (for os-version on linux/windows)")
	postureCmd.String("locations", "", "Locations (e.g., US:NewYork,GB:London, or just US,GB)")
	postureCmd.String("action", "allow", "Action: allow or deny (for geo-location and network-range)")
	postureCmd.String("ranges", "", "CIDR ranges (comma-separated, for network-range)")
	postureCmd.String("linux-path", "", "Linux process path (for process)")
	postureCmd.String("mac-path", "", "macOS process path (for process)")
	postureCmd.String("windows-path", "", "Windows process path (for process)")

	// Update flags
	updateFlag := postureCmd.String("update", "", "Update a posture check by ID")

	// Delete flags
	deleteFlag := postureCmd.String("delete", "", "Delete a posture check by ID")

	// If no flags provided, show usage
	if len(args) == 1 {
		printPostureCheckUsage()
		return nil
	}

	// Parse the flags
	if err := postureCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags in priority order

	// Create posture check
	if *createFlag != "" {
		if *checkTypeFlag == "" {
			return fmt.Errorf("--type is required when creating a posture check")
		}
		return client.createPostureCheck(*createFlag, *descriptionFlag, *checkTypeFlag, postureCmd)
	}

	// Delete posture check
	if *deleteFlag != "" {
		return client.deletePostureCheck(*deleteFlag)
	}

	// Update posture check
	if *updateFlag != "" {
		if *checkTypeFlag == "" {
			return fmt.Errorf("--type is required when updating a posture check")
		}
		return client.updatePostureCheck(*updateFlag, *descriptionFlag, *checkTypeFlag, postureCmd)
	}

	// Inspect posture check
	if *inspectFlag != "" {
		return client.inspectPostureCheck(*inspectFlag)
	}

	// List posture checks
	if *listFlag {
		filters := &PostureCheckFilters{
			NamePattern: *filterName,
			CheckType:   *filterType,
		}
		return client.listPostureChecks(filters)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'posture-check' command.")
	printPostureCheckUsage()
	return nil
}

// PostureCheckFilters holds filtering options for listing posture checks
type PostureCheckFilters struct {
	NamePattern string
	CheckType   string
}

// listPostureChecks implements the "posture-check --list" command
func (c *Client) listPostureChecks(filters *PostureCheckFilters) error {
	resp, err := c.makeRequest("GET", "/posture-checks", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var checks []PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
		return fmt.Errorf("failed to decode posture checks response: %v", err)
	}

	// Apply filters
	var filtered []PostureCheck
	for _, check := range checks {
		// Filter by name pattern
		if filters.NamePattern != "" && !matchesPattern(check.Name, filters.NamePattern) {
			continue
		}

		// Filter by check type
		if filters.CheckType != "" {
			checkType := getCheckType(check.Checks)
			if !strings.EqualFold(checkType, filters.CheckType) {
				continue
			}
		}

		filtered = append(filtered, check)
	}

	if len(filtered) == 0 {
		fmt.Println("No posture checks found.")
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tDESCRIPTION")
	fmt.Fprintln(w, "--\t----\t----\t-----------")

	for _, check := range filtered {
		checkType := getCheckType(check.Checks)
		desc := check.Description
		if desc == "" {
			desc = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			check.ID,
			check.Name,
			checkType,
			desc,
		)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d posture checks\n", len(filtered))
	return nil
}

// inspectPostureCheck implements the "posture-check --inspect" command
func (c *Client) inspectPostureCheck(checkID string) error {
	resp, err := c.makeRequest("GET", "/posture-checks/"+checkID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var check PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&check); err != nil {
		return fmt.Errorf("failed to decode posture check response: %v", err)
	}

	// Print detailed posture check information
	fmt.Println("Posture Check Details:")
	fmt.Println("======================")
	fmt.Printf("ID:          %s\n", check.ID)
	fmt.Printf("Name:        %s\n", check.Name)
	fmt.Printf("Type:        %s\n", getCheckType(check.Checks))

	if check.Description != "" {
		fmt.Printf("Description: %s\n", check.Description)
	}

	fmt.Println()
	fmt.Println("Check Configuration:")
	fmt.Println("--------------------")

	// Display check-specific details
	if check.Checks.NBVersionCheck != nil {
		fmt.Printf("NetBird Version Check:\n")
		fmt.Printf("  Minimum Version: %s\n", check.Checks.NBVersionCheck.MinVersion)
	}

	if check.Checks.OSVersionCheck != nil {
		fmt.Printf("OS Version Check:\n")
		if check.Checks.OSVersionCheck.Android != nil {
			fmt.Printf("  Android:  min version %s\n", check.Checks.OSVersionCheck.Android.MinVersion)
		}
		if check.Checks.OSVersionCheck.Darwin != nil {
			fmt.Printf("  macOS:    min version %s\n", check.Checks.OSVersionCheck.Darwin.MinVersion)
		}
		if check.Checks.OSVersionCheck.IOS != nil {
			fmt.Printf("  iOS:      min version %s\n", check.Checks.OSVersionCheck.IOS.MinVersion)
		}
		if check.Checks.OSVersionCheck.Linux != nil {
			fmt.Printf("  Linux:    min kernel %s\n", check.Checks.OSVersionCheck.Linux.MinKernelVersion)
		}
		if check.Checks.OSVersionCheck.Windows != nil {
			fmt.Printf("  Windows:  min kernel %s\n", check.Checks.OSVersionCheck.Windows.MinKernelVersion)
		}
	}

	if check.Checks.GeoLocationCheck != nil {
		fmt.Printf("Geo-Location Check:\n")
		fmt.Printf("  Action: %s\n", check.Checks.GeoLocationCheck.Action)
		fmt.Printf("  Locations:\n")
		for _, loc := range check.Checks.GeoLocationCheck.Locations {
			if loc.CityName != "" {
				fmt.Printf("    - %s:%s\n", loc.CountryCode, loc.CityName)
			} else {
				fmt.Printf("    - %s (entire country)\n", loc.CountryCode)
			}
		}
	}

	if check.Checks.PeerNetworkRangeCheck != nil {
		fmt.Printf("Peer Network Range Check:\n")
		fmt.Printf("  Action: %s\n", check.Checks.PeerNetworkRangeCheck.Action)
		fmt.Printf("  Ranges:\n")
		for _, cidr := range check.Checks.PeerNetworkRangeCheck.Ranges {
			fmt.Printf("    - %s\n", cidr)
		}
	}

	if check.Checks.ProcessCheck != nil {
		fmt.Printf("Process Check:\n")
		for _, proc := range check.Checks.ProcessCheck.Processes {
			if proc.LinuxPath != "" {
				fmt.Printf("  Linux:   %s\n", proc.LinuxPath)
			}
			if proc.MacPath != "" {
				fmt.Printf("  macOS:   %s\n", proc.MacPath)
			}
			if proc.WindowsPath != "" {
				fmt.Printf("  Windows: %s\n", proc.WindowsPath)
			}
		}
	}

	return nil
}

// createPostureCheck implements the "posture-check --create" command
func (c *Client) createPostureCheck(name, description, checkType string, flags *flag.FlagSet) error {
	// Build check definition based on type
	checks, err := buildCheckDefinition(checkType, flags)
	if err != nil {
		return err
	}

	reqBody := PostureCheckRequest{
		Name:        name,
		Description: description,
		Checks:      checks,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/posture-checks", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var createdCheck PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&createdCheck); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Posture check created successfully!\n")
	fmt.Printf("  ID:   %s\n", createdCheck.ID)
	fmt.Printf("  Name: %s\n", createdCheck.Name)
	fmt.Printf("  Type: %s\n", checkType)
	return nil
}

// updatePostureCheck implements the "posture-check --update" command
func (c *Client) updatePostureCheck(checkID, description, checkType string, flags *flag.FlagSet) error {
	// First, get the current check
	resp, err := c.makeRequest("GET", "/posture-checks/"+checkID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentCheck PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&currentCheck); err != nil {
		return fmt.Errorf("failed to decode current posture check: %v", err)
	}

	// Build check definition based on type
	checks, err := buildCheckDefinition(checkType, flags)
	if err != nil {
		return err
	}

	// Build update request
	updateReq := PostureCheckRequest{
		Name:        currentCheck.Name,
		Description: description,
		Checks:      checks,
	}

	if description == "" {
		updateReq.Description = currentCheck.Description
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/posture-checks/"+checkID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Posture check %s updated successfully\n", checkID)
	return nil
}

// deletePostureCheck implements the "posture-check --delete" command
func (c *Client) deletePostureCheck(checkID string) error {
	resp, err := c.makeRequest("DELETE", "/posture-checks/"+checkID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Posture check %s deleted successfully\n", checkID)
	return nil
}

// buildCheckDefinition builds a PostureCheckDefinition from command flags
func buildCheckDefinition(checkType string, flags *flag.FlagSet) (PostureCheckDefinition, error) {
	var checks PostureCheckDefinition

	switch checkType {
	case "nb-version":
		minVersion := flags.Lookup("min-version").Value.String()
		if minVersion == "" {
			return checks, fmt.Errorf("--min-version is required for nb-version check")
		}
		checks.NBVersionCheck = &NBVersionCheck{
			MinVersion: minVersion,
		}

	case "os-version":
		osType := flags.Lookup("os").Value.String()
		if osType == "" {
			return checks, fmt.Errorf("--os is required for os-version check")
		}

		osCheck := &OSVersionCheck{}

		switch osType {
		case "android":
			minVersion := flags.Lookup("min-os-version").Value.String()
			if minVersion == "" {
				return checks, fmt.Errorf("--min-os-version is required for Android")
			}
			osCheck.Android = &MinVersionConfig{MinVersion: minVersion}

		case "darwin":
			minVersion := flags.Lookup("min-os-version").Value.String()
			if minVersion == "" {
				return checks, fmt.Errorf("--min-os-version is required for macOS")
			}
			osCheck.Darwin = &MinVersionConfig{MinVersion: minVersion}

		case "ios":
			minVersion := flags.Lookup("min-os-version").Value.String()
			if minVersion == "" {
				return checks, fmt.Errorf("--min-os-version is required for iOS")
			}
			osCheck.IOS = &MinVersionConfig{MinVersion: minVersion}

		case "linux":
			minKernel := flags.Lookup("min-kernel").Value.String()
			if minKernel == "" {
				return checks, fmt.Errorf("--min-kernel is required for Linux")
			}
			osCheck.Linux = &MinKernelVersionConfig{MinKernelVersion: minKernel}

		case "windows":
			minKernel := flags.Lookup("min-kernel").Value.String()
			if minKernel == "" {
				return checks, fmt.Errorf("--min-kernel is required for Windows")
			}
			osCheck.Windows = &MinKernelVersionConfig{MinKernelVersion: minKernel}

		default:
			return checks, fmt.Errorf("invalid OS type: %s (must be android, darwin, ios, linux, or windows)", osType)
		}

		checks.OSVersionCheck = osCheck

	case "geo-location":
		locationsStr := flags.Lookup("locations").Value.String()
		if locationsStr == "" {
			return checks, fmt.Errorf("--locations is required for geo-location check")
		}

		locations, err := parseLocations(locationsStr)
		if err != nil {
			return checks, err
		}

		action := flags.Lookup("action").Value.String()
		if action != "allow" && action != "deny" {
			return checks, fmt.Errorf("action must be 'allow' or 'deny' (got '%s')", action)
		}

		checks.GeoLocationCheck = &GeoLocationCheck{
			Locations: locations,
			Action:    action,
		}

	case "network-range":
		rangesStr := flags.Lookup("ranges").Value.String()
		if rangesStr == "" {
			return checks, fmt.Errorf("--ranges is required for network-range check")
		}

		ranges := splitCommaList(rangesStr)
		// Validate each CIDR
		for _, cidr := range ranges {
			if err := validateCIDR(cidr); err != nil {
				return checks, fmt.Errorf("invalid CIDR '%s': %v", cidr, err)
			}
		}

		action := flags.Lookup("action").Value.String()
		if action != "allow" && action != "deny" {
			return checks, fmt.Errorf("action must be 'allow' or 'deny' (got '%s')", action)
		}

		checks.PeerNetworkRangeCheck = &PeerNetworkRangeCheck{
			Ranges: ranges,
			Action: action,
		}

	case "process":
		linuxPath := flags.Lookup("linux-path").Value.String()
		macPath := flags.Lookup("mac-path").Value.String()
		windowsPath := flags.Lookup("windows-path").Value.String()

		if linuxPath == "" && macPath == "" && windowsPath == "" {
			return checks, fmt.Errorf("at least one process path is required (--linux-path, --mac-path, or --windows-path)")
		}

		process := Process{
			LinuxPath:   linuxPath,
			MacPath:     macPath,
			WindowsPath: windowsPath,
		}

		checks.ProcessCheck = &ProcessCheck{
			Processes: []Process{process},
		}

	default:
		return checks, fmt.Errorf("invalid check type: %s (must be nb-version, os-version, geo-location, network-range, or process)", checkType)
	}

	return checks, nil
}

// parseLocations parses location strings
// Format: "US:NewYork,GB:London" or "US,GB" (country only)
func parseLocations(locationsStr string) ([]Location, error) {
	parts := strings.Split(locationsStr, ",")
	var locations []Location

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var loc Location
		if strings.Contains(part, ":") {
			// Country:City format
			locParts := strings.Split(part, ":")
			if len(locParts) != 2 {
				return nil, fmt.Errorf("invalid location format '%s': expected CountryCode:CityName or CountryCode", part)
			}
			loc.CountryCode = strings.TrimSpace(locParts[0])
			loc.CityName = strings.TrimSpace(locParts[1])
		} else {
			// Country only
			loc.CountryCode = part
		}

		// Validate country code (should be 2 letters)
		if len(loc.CountryCode) != 2 {
			return nil, fmt.Errorf("invalid country code '%s': must be 2-letter ISO 3166-1 alpha-2 code", loc.CountryCode)
		}

		locations = append(locations, loc)
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("at least one location is required")
	}

	return locations, nil
}

// getCheckType returns a human-readable check type
func getCheckType(checks PostureCheckDefinition) string {
	if checks.NBVersionCheck != nil {
		return "nb-version"
	}
	if checks.OSVersionCheck != nil {
		return "os-version"
	}
	if checks.GeoLocationCheck != nil {
		return "geo-location"
	}
	if checks.PeerNetworkRangeCheck != nil {
		return "network-range"
	}
	if checks.ProcessCheck != nil {
		return "process"
	}
	return "unknown"
}

// printPostureCheckUsage prints usage information for the posture-check command
func printPostureCheckUsage() {
	fmt.Println("Usage: netbird-manage posture-check [options]")
	fmt.Println("\nManage device posture checks for zero-trust security.")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                              List all posture checks")
	fmt.Println("    --filter-name <pattern>           Filter by name pattern")
	fmt.Println("    --filter-type <type>              Filter by check type")
	fmt.Println("  --inspect <check-id>                Show detailed posture check information")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --create <name>                     Create a posture check")
	fmt.Println("    --description <desc>              Description (optional)")
	fmt.Println("    --type <type>                     Check type (required, see below)")
	fmt.Println()
	fmt.Println("  --update <check-id>                 Update a posture check")
	fmt.Println("    --description <desc>              New description")
	fmt.Println("    --type <type>                     Check type (required)")
	fmt.Println()
	fmt.Println("  --delete <check-id>                 Delete a posture check")
	fmt.Println()
	fmt.Println("Check Types and Their Flags:")
	fmt.Println("----------------------------")
	fmt.Println()
	fmt.Println("  nb-version                          NetBird version check")
	fmt.Println("    --min-version <version>           Minimum version (e.g., 0.28.0)")
	fmt.Println()
	fmt.Println("  os-version                          OS version check")
	fmt.Println("    --os <os-type>                    OS: android, darwin, ios, linux, windows")
	fmt.Println("    --min-os-version <version>        Min version (for android, darwin, ios)")
	fmt.Println("    --min-kernel <version>            Min kernel (for linux, windows)")
	fmt.Println()
	fmt.Println("  geo-location                        Geographic location check")
	fmt.Println("    --locations <CC:City,...>         Locations (e.g., US:NewYork,GB or US,GB)")
	fmt.Println("    --action <allow|deny>             Action to take (default: allow)")
	fmt.Println()
	fmt.Println("  network-range                       Peer network range check")
	fmt.Println("    --ranges <cidr1,cidr2,...>        CIDR ranges (e.g., 192.168.0.0/16)")
	fmt.Println("    --action <allow|deny>             Action to take (default: allow)")
	fmt.Println()
	fmt.Println("  process                             Process running check")
	fmt.Println("    --linux-path <path>               Linux process path")
	fmt.Println("    --mac-path <path>                 macOS process path")
	fmt.Println("    --windows-path <path>             Windows process path")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List all posture checks")
	fmt.Println("  netbird-manage posture-check --list")
	fmt.Println()
	fmt.Println("  # Create NetBird version check")
	fmt.Println("  netbird-manage posture-check --create \"min-nb-version\" \\")
	fmt.Println("    --type nb-version \\")
	fmt.Println("    --min-version \"0.28.0\"")
	fmt.Println()
	fmt.Println("  # Create geo-location check (US only)")
	fmt.Println("  netbird-manage posture-check --create \"us-only\" \\")
	fmt.Println("    --type geo-location \\")
	fmt.Println("    --locations \"US\" \\")
	fmt.Println("    --action allow")
	fmt.Println()
	fmt.Println("  # Create macOS version check")
	fmt.Println("  netbird-manage posture-check --create \"macos-13+\" \\")
	fmt.Println("    --type os-version \\")
	fmt.Println("    --os darwin \\")
	fmt.Println("    --min-os-version \"13.0\"")
	fmt.Println()
	fmt.Println("  # Create corporate network check")
	fmt.Println("  netbird-manage posture-check --create \"corporate-net\" \\")
	fmt.Println("    --type network-range \\")
	fmt.Println("    --ranges \"192.168.0.0/16,10.0.0.0/8\" \\")
	fmt.Println("    --action allow")
	fmt.Println()
	fmt.Println("  # Inspect a posture check")
	fmt.Println("  netbird-manage posture-check --inspect <check-id>")
}
