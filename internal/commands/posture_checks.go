// posture_checks.go
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"text/tabwriter"

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// PostureCheckFilters holds filtering options for listing posture checks
type PostureCheckFilters struct {
	NamePattern string
	CheckType   string
}

// HandlePostureChecksCommand routes posture check-related commands
func (s *Service) HandlePostureChecksCommand(args []string) error {
	postureCmd := flag.NewFlagSet("posture-check", flag.ContinueOnError)
	postureCmd.SetOutput(os.Stderr)
	postureCmd.Usage = PrintPostureCheckUsage

	// Query flags
	listFlag := postureCmd.Bool("list", false, "List all posture checks")
	inspectFlag := postureCmd.String("inspect", "", "Inspect a posture check by ID")
	filterName := postureCmd.String("filter-name", "", "Filter by name pattern")
	filterType := postureCmd.String("filter-type", "", "Filter by check type")
	outputFlag := postureCmd.String("output", "table", "Output format: table or json")

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
		PrintPostureCheckUsage()
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
		return s.createPostureCheck(*createFlag, *descriptionFlag, *checkTypeFlag, postureCmd)
	}

	// Delete posture check
	if *deleteFlag != "" {
		return s.deletePostureCheck(*deleteFlag)
	}

	// Update posture check
	if *updateFlag != "" {
		if *checkTypeFlag == "" {
			return fmt.Errorf("--type is required when updating a posture check")
		}
		return s.updatePostureCheck(*updateFlag, *descriptionFlag, *checkTypeFlag, postureCmd)
	}

	// Inspect posture check
	if *inspectFlag != "" {
		return s.inspectPostureCheck(*inspectFlag, *outputFlag)
	}

	// List posture checks
	if *listFlag {
		filters := &PostureCheckFilters{
			NamePattern: *filterName,
			CheckType:   *filterType,
		}
		return s.listPostureChecks(filters, *outputFlag)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'posture-check' command.")
	PrintPostureCheckUsage()
	return nil
}

// listPostureChecks implements the "posture-check --list" command
func (s *Service) listPostureChecks(filters *PostureCheckFilters, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/posture-checks", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var checks []models.PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
		return fmt.Errorf("failed to decode posture checks response: %v", err)
	}

	// Apply filters
	var filtered []models.PostureCheck
	for _, check := range checks {
		// Filter by name pattern
		if filters.NamePattern != "" && !helpers.MatchesPattern(check.Name, filters.NamePattern) {
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

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
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
func (s *Service) inspectPostureCheck(checkID string, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/posture-checks/"+checkID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var check models.PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&check); err != nil {
		return fmt.Errorf("failed to decode posture check response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(check, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
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
func (s *Service) createPostureCheck(name, description, checkType string, flags *flag.FlagSet) error {
	// Build check definition based on type
	checks, err := buildCheckDefinition(checkType, flags)
	if err != nil {
		return err
	}

	reqBody := models.PostureCheckRequest{
		Name:        name,
		Description: description,
		Checks:      checks,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/posture-checks", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var createdCheck models.PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&createdCheck); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Posture check created successfully!\n")
	fmt.Printf("  ID:   %s\n", createdCheck.ID)
	fmt.Printf("  Name: %s\n", createdCheck.Name)
	fmt.Printf("  Type: %s\n", checkType)
	return nil
}

// updatePostureCheck implements the "posture-check --update" command
func (s *Service) updatePostureCheck(checkID, description, checkType string, flags *flag.FlagSet) error {
	// First, get the current check
	resp, err := s.Client.MakeRequest("GET", "/posture-checks/"+checkID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentCheck models.PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&currentCheck); err != nil {
		return fmt.Errorf("failed to decode current posture check: %v", err)
	}

	// Build check definition based on type
	checks, err := buildCheckDefinition(checkType, flags)
	if err != nil {
		return err
	}

	// Build update request
	updateReq := models.PostureCheckRequest{
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

	resp, err = s.Client.MakeRequest("PUT", "/posture-checks/"+checkID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Posture check %s updated successfully\n", checkID)
	return nil
}

// deletePostureCheck implements the "posture-check --delete" command
func (s *Service) deletePostureCheck(checkID string) error {
	// Fetch posture check details first
	resp, err := s.Client.MakeRequest("GET", "/posture-checks/"+checkID, nil)
	if err != nil {
		return err
	}
	var check models.PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&check); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode posture check: %v", err)
	}
	resp.Body.Close()

	// Build details map
	checkType := getCheckType(check.Checks)
	details := map[string]string{
		"Type": checkType,
	}
	if check.Description != "" {
		details["Description"] = check.Description
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("posture check", check.Name, checkID, details) {
		return nil // User cancelled
	}

	resp, err = s.Client.MakeRequest("DELETE", "/posture-checks/"+checkID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Posture check %s deleted successfully\n", checkID)
	return nil
}

// buildCheckDefinition builds a PostureCheckDefinition from command flags
func buildCheckDefinition(checkType string, flags *flag.FlagSet) (models.PostureCheckDefinition, error) {
	var checks models.PostureCheckDefinition

	switch checkType {
	case "nb-version":
		minVersion := flags.Lookup("min-version").Value.String()
		if minVersion == "" {
			return checks, fmt.Errorf("--min-version is required for nb-version check")
		}
		checks.NBVersionCheck = &models.NBVersionCheck{
			MinVersion: minVersion,
		}

	case "os-version":
		osType := flags.Lookup("os").Value.String()
		if osType == "" {
			return checks, fmt.Errorf("--os is required for os-version check")
		}

		osCheck := &models.OSVersionCheck{}

		switch osType {
		case "android":
			minVersion := flags.Lookup("min-os-version").Value.String()
			if minVersion == "" {
				return checks, fmt.Errorf("--min-os-version is required for Android")
			}
			osCheck.Android = &models.MinVersionConfig{MinVersion: minVersion}

		case "darwin":
			minVersion := flags.Lookup("min-os-version").Value.String()
			if minVersion == "" {
				return checks, fmt.Errorf("--min-os-version is required for macOS")
			}
			osCheck.Darwin = &models.MinVersionConfig{MinVersion: minVersion}

		case "ios":
			minVersion := flags.Lookup("min-os-version").Value.String()
			if minVersion == "" {
				return checks, fmt.Errorf("--min-os-version is required for iOS")
			}
			osCheck.IOS = &models.MinVersionConfig{MinVersion: minVersion}

		case "linux":
			minKernel := flags.Lookup("min-kernel").Value.String()
			if minKernel == "" {
				return checks, fmt.Errorf("--min-kernel is required for Linux")
			}
			osCheck.Linux = &models.MinKernelVersionConfig{MinKernelVersion: minKernel}

		case "windows":
			minKernel := flags.Lookup("min-kernel").Value.String()
			if minKernel == "" {
				return checks, fmt.Errorf("--min-kernel is required for Windows")
			}
			osCheck.Windows = &models.MinKernelVersionConfig{MinKernelVersion: minKernel}

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

		checks.GeoLocationCheck = &models.GeoLocationCheck{
			Locations: locations,
			Action:    action,
		}

	case "network-range":
		rangesStr := flags.Lookup("ranges").Value.String()
		if rangesStr == "" {
			return checks, fmt.Errorf("--ranges is required for network-range check")
		}

		ranges := helpers.SplitCommaList(rangesStr)
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

		checks.PeerNetworkRangeCheck = &models.PeerNetworkRangeCheck{
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

		process := models.Process{
			LinuxPath:   linuxPath,
			MacPath:     macPath,
			WindowsPath: windowsPath,
		}

		checks.ProcessCheck = &models.ProcessCheck{
			Processes: []models.Process{process},
		}

	default:
		return checks, fmt.Errorf("invalid check type: %s (must be nb-version, os-version, geo-location, network-range, or process)", checkType)
	}

	return checks, nil
}

// parseLocations parses location strings
// Format: "US:NewYork,GB:London" or "US,GB" (country only)
func parseLocations(locationsStr string) ([]models.Location, error) {
	parts := strings.Split(locationsStr, ",")
	var locations []models.Location

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var loc models.Location
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
func getCheckType(checks models.PostureCheckDefinition) string {
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

// validateCIDR validates a CIDR notation string
func validateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR notation '%s': %v", cidr, err)
	}
	return nil
}
