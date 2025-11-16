// setup-keys.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// handleSetupKeysCommand routes setup-key related commands using the flag package
func handleSetupKeysCommand(client *Client, args []string) error {
	// Create a new flag set for the 'setup-key' command
	setupKeyCmd := flag.NewFlagSet("setup-key", flag.ContinueOnError)
	setupKeyCmd.SetOutput(os.Stderr)
	setupKeyCmd.Usage = printSetupKeyUsage

	// Query flags
	listFlag := setupKeyCmd.Bool("list", false, "List all setup keys")
	inspectFlag := setupKeyCmd.String("inspect", "", "Inspect a setup key by its ID")
	filterNameFlag := setupKeyCmd.String("filter-name", "", "Filter by name pattern (use with --list)")
	filterTypeFlag := setupKeyCmd.String("filter-type", "", "Filter by type: one-off or reusable (use with --list)")
	validOnlyFlag := setupKeyCmd.Bool("valid-only", false, "Show only valid keys (use with --list)")

	// Create flags
	createFlag := setupKeyCmd.String("create", "", "Create a new setup key with the given name")
	keyTypeFlag := setupKeyCmd.String("type", "one-off", "Key type: one-off or reusable (default: one-off)")
	expiresInFlag := setupKeyCmd.String("expires-in", "7d", "Expiration duration: 1d, 7d, 30d, 90d, 1y (default: 7d)")
	autoGroupsFlag := setupKeyCmd.String("auto-groups", "", "Comma-separated group IDs for auto-assignment")
	usageLimitFlag := setupKeyCmd.Int("usage-limit", 0, "Usage limit (0 = unlimited, default: 0)")
	ephemeralFlag := setupKeyCmd.Bool("ephemeral", false, "Mark peers as ephemeral")
	allowExtraDNSLabelsFlag := setupKeyCmd.Bool("allow-extra-dns-labels", false, "Allow extra DNS labels")

	// Quick create flag
	quickFlag := setupKeyCmd.String("quick", "", "Quick create one-off key with defaults (7d expiration, single use)")

	// Update/revoke flags
	revokeFlag := setupKeyCmd.String("revoke", "", "Revoke a setup key by its ID")
	enableFlag := setupKeyCmd.String("enable", "", "Enable (un-revoke) a setup key by its ID")
	updateGroupsFlag := setupKeyCmd.String("update-groups", "", "Update auto-groups for a setup key by ID")
	groupsFlag := setupKeyCmd.String("groups", "", "New comma-separated group IDs (requires --update-groups)")

	// Delete flag
	deleteFlag := setupKeyCmd.String("delete", "", "Delete a setup key by its ID")

	// If no flags provided, show usage
	if len(args) == 1 {
		printSetupKeyUsage()
		return nil
	}

	// Parse the flags
	if err := setupKeyCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags
	if *listFlag {
		return client.listSetupKeys(*filterNameFlag, *filterTypeFlag, *validOnlyFlag)
	}

	if *inspectFlag != "" {
		return client.inspectSetupKey(*inspectFlag)
	}

	if *createFlag != "" {
		expiresInSec, err := parseDuration(*expiresInFlag)
		if err != nil {
			return fmt.Errorf("invalid expiration duration: %v", err)
		}
		autoGroups := splitCommaList(*autoGroupsFlag)
		return client.createSetupKey(*createFlag, *keyTypeFlag, expiresInSec, autoGroups, *usageLimitFlag, *ephemeralFlag, *allowExtraDNSLabelsFlag)
	}

	if *quickFlag != "" {
		// Quick create with sensible defaults
		return client.createSetupKey(*quickFlag, "one-off", 7*24*3600, []string{}, 1, false, false)
	}

	if *revokeFlag != "" {
		return client.updateSetupKeyRevocation(*revokeFlag, true)
	}

	if *enableFlag != "" {
		return client.updateSetupKeyRevocation(*enableFlag, false)
	}

	if *updateGroupsFlag != "" {
		if *groupsFlag == "" {
			return fmt.Errorf("flag --update-groups requires --groups")
		}
		newGroups := splitCommaList(*groupsFlag)
		return client.updateSetupKeyGroups(*updateGroupsFlag, newGroups)
	}

	if *deleteFlag != "" {
		return client.deleteSetupKey(*deleteFlag)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'setup-key' command.")
	printSetupKeyUsage()
	return nil
}

// printSetupKeyUsage provides specific help for the 'setup-key' command
func printSetupKeyUsage() {
	fmt.Println("Usage: netbird-manage setup-key <flag> [arguments]")
	fmt.Println("\nManage device registration and onboarding keys.")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                              List all setup keys")
	fmt.Println("    --filter-name <pattern>           Filter by name (supports wildcards: office-*)")
	fmt.Println("    --filter-type <one-off|reusable>  Filter by key type")
	fmt.Println("    --valid-only                      Show only valid (non-revoked, non-expired) keys")
	fmt.Println("  --inspect <key-id>                  Inspect a specific setup key")
	fmt.Println()
	fmt.Println("Create Flags:")
	fmt.Println("  --create <name>                     Create a new setup key")
	fmt.Println("    --type <one-off|reusable>         Key type (default: one-off)")
	fmt.Println("    --expires-in <duration>           Expiration: 1d, 7d, 30d, 90d, 1y (default: 7d)")
	fmt.Println("    --auto-groups <id1,id2,...>       Comma-separated group IDs for auto-assignment")
	fmt.Println("    --usage-limit <number>            Max uses, 0 = unlimited (default: 0)")
	fmt.Println("    --ephemeral                       Mark peers as ephemeral")
	fmt.Println("    --allow-extra-dns-labels          Allow additional DNS labels")
	fmt.Println()
	fmt.Println("  --quick <name>                      Quick create one-off key (7d expiration, single use)")
	fmt.Println()
	fmt.Println("Update Flags:")
	fmt.Println("  --revoke <key-id>                   Revoke a setup key")
	fmt.Println("  --enable <key-id>                   Enable (un-revoke) a setup key")
	fmt.Println("  --update-groups <key-id>            Update auto-groups for a key")
	fmt.Println("    --groups <id1,id2,...>            New group IDs (required)")
	fmt.Println()
	fmt.Println("Delete Flags:")
	fmt.Println("  --delete <key-id>                   Delete a setup key")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  netbird-manage setup-key --list")
	fmt.Println("  netbird-manage setup-key --quick office-laptop")
	fmt.Println("  netbird-manage setup-key --create team-key --type reusable --expires-in 30d")
	fmt.Println("  netbird-manage setup-key --inspect 12345")
	fmt.Println("  netbird-manage setup-key --revoke 12345")
}

// parseDuration converts human-readable duration to seconds
func parseDuration(duration string) (int, error) {
	duration = strings.TrimSpace(strings.ToLower(duration))

	// Extract number and unit
	var num string
	var unit string

	for i, char := range duration {
		if char >= '0' && char <= '9' {
			num += string(char)
		} else {
			unit = duration[i:]
			break
		}
	}

	if num == "" {
		return 0, fmt.Errorf("no numeric value found in duration: %s", duration)
	}

	value, err := strconv.Atoi(num)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", num)
	}

	// Convert to seconds based on unit
	var seconds int
	switch unit {
	case "d", "day", "days":
		seconds = value * 24 * 3600
	case "w", "week", "weeks":
		seconds = value * 7 * 24 * 3600
	case "m", "month", "months":
		seconds = value * 30 * 24 * 3600
	case "y", "year", "years":
		seconds = value * 365 * 24 * 3600
	case "h", "hour", "hours":
		seconds = value * 3600
	default:
		return 0, fmt.Errorf("unknown duration unit: %s (use d, w, m, y, or h)", unit)
	}

	// Validate API constraints (86400-31536000 seconds = 1 day to 1 year)
	if seconds < 86400 {
		return 0, fmt.Errorf("expiration must be at least 1 day (got %d seconds)", seconds)
	}
	if seconds > 31536000 {
		return 0, fmt.Errorf("expiration cannot exceed 1 year (got %d seconds)", seconds)
	}

	return seconds, nil
}

// formatDuration converts seconds to human-readable duration
func formatDuration(seconds int) string {
	if seconds == 0 {
		return "Never"
	}

	days := seconds / (24 * 3600)
	if days >= 365 {
		years := days / 365
		if years == 1 {
			return "1 year"
		}
		return fmt.Sprintf("%d years", years)
	}
	if days >= 30 {
		months := days / 30
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	}
	if days >= 7 {
		weeks := days / 7
		if weeks == 1 {
			return "1 week"
		}
		return fmt.Sprintf("%d weeks", weeks)
	}
	if days > 0 {
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}

	hours := seconds / 3600
	if hours > 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}

	return fmt.Sprintf("%d seconds", seconds)
}

// formatExpiration formats expiration timestamp with time remaining
func formatExpiration(expiresStr string) string {
	if expiresStr == "" {
		return "Never"
	}

	expires, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return expiresStr
	}

	now := time.Now()
	if expires.Before(now) {
		return fmt.Sprintf("Expired (%s)", expires.Format("2006-01-02"))
	}

	remaining := expires.Sub(now)
	days := int(remaining.Hours() / 24)

	if days > 0 {
		return fmt.Sprintf("%s (in %d days)", expires.Format("2006-01-02"), days)
	}

	hours := int(remaining.Hours())
	if hours > 0 {
		return fmt.Sprintf("%s (in %d hours)", expires.Format("2006-01-02 15:04"), hours)
	}

	return fmt.Sprintf("%s (soon)", expires.Format("2006-01-02 15:04"))
}

// formatState formats the key state with visual indicators
func formatState(state string, valid, revoked bool) string {
	if revoked {
		return "✗ Revoked"
	}
	if !valid {
		return "✗ Expired"
	}
	if state == "valid" {
		return "✓ Valid"
	}
	return state
}

// listSetupKeys lists all setup keys with optional filters
func (c *Client) listSetupKeys(filterName, filterType string, validOnly bool) error {
	resp, err := c.makeRequest("GET", "/setup-keys", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var keys []SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Apply filters
	var filtered []SetupKey
	for _, key := range keys {
		// Filter by name
		if filterName != "" && !matchesPattern(key.Name, filterName) {
			continue
		}

		// Filter by type
		if filterType != "" && !strings.EqualFold(key.Type, filterType) {
			continue
		}

		// Filter by validity
		if validOnly && (!key.Valid || key.Revoked) {
			continue
		}

		filtered = append(filtered, key)
	}

	if len(filtered) == 0 {
		fmt.Println("No setup keys found.")
		return nil
	}

	// Display in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTYPE\tSTATE\tUSED/LIMIT\tEXPIRES\tGROUPS")
	fmt.Fprintln(w, "--\t----\t----\t-----\t----------\t-------\t------")

	for _, key := range filtered {
		usageLimit := "∞"
		if key.UsageLimit > 0 {
			usageLimit = strconv.Itoa(key.UsageLimit)
		}

		groupCount := len(key.AutoGroups)
		groupsStr := "-"
		if groupCount > 0 {
			groupsStr = fmt.Sprintf("%d groups", groupCount)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d/%s\t%s\t%s\n",
			key.ID,
			key.Name,
			key.Type,
			formatState(key.State, key.Valid, key.Revoked),
			key.UsedTimes,
			usageLimit,
			formatExpiration(key.Expires),
			groupsStr,
		)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d setup keys\n", len(filtered))
	return nil
}

// inspectSetupKey shows detailed information about a setup key
func (c *Client) inspectSetupKey(keyID string) error {
	resp, err := c.makeRequest("GET", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var key SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Display key details
	fmt.Printf("Setup Key Details\n")
	fmt.Printf("=================\n\n")
	fmt.Printf("ID:                    %s\n", key.ID)
	fmt.Printf("Name:                  %s\n", key.Name)
	fmt.Printf("Type:                  %s\n", key.Type)
	fmt.Printf("State:                 %s\n", formatState(key.State, key.Valid, key.Revoked))
	fmt.Printf("Valid:                 %v\n", key.Valid)
	fmt.Printf("Revoked:               %v\n", key.Revoked)
	fmt.Printf("\n")

	fmt.Printf("Usage Statistics\n")
	fmt.Printf("----------------\n")
	fmt.Printf("Used Times:            %d", key.UsedTimes)
	if key.UsageLimit > 0 {
		fmt.Printf(" / %d", key.UsageLimit)
		remaining := key.UsageLimit - key.UsedTimes
		if remaining > 0 {
			fmt.Printf(" (%d remaining)", remaining)
		} else {
			fmt.Printf(" (exhausted)")
		}
	} else {
		fmt.Printf(" (unlimited)")
	}
	fmt.Println()

	if key.LastUsed != "" && key.LastUsed != "0001-01-01T00:00:00Z" {
		lastUsed, err := time.Parse(time.RFC3339, key.LastUsed)
		if err == nil {
			fmt.Printf("Last Used:             %s\n", lastUsed.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Printf("Last Used:             Never\n")
	}
	fmt.Printf("\n")

	fmt.Printf("Expiration\n")
	fmt.Printf("----------\n")
	fmt.Printf("Expires:               %s\n", formatExpiration(key.Expires))
	fmt.Printf("\n")

	fmt.Printf("Configuration\n")
	fmt.Printf("-------------\n")
	fmt.Printf("Ephemeral Peers:       %v\n", key.Ephemeral)
	fmt.Printf("Extra DNS Labels:      %v\n", key.AllowExtraDNSLabels)
	fmt.Printf("\n")

	fmt.Printf("Auto-Groups\n")
	fmt.Printf("-----------\n")
	if len(key.AutoGroups) == 0 {
		fmt.Printf("None\n")
	} else {
		for _, groupID := range key.AutoGroups {
			fmt.Printf("  - %s\n", groupID)
		}
	}
	fmt.Printf("\n")

	if key.UpdatedAt != "" {
		updated, err := time.Parse(time.RFC3339, key.UpdatedAt)
		if err == nil {
			fmt.Printf("Last Updated:          %s\n", updated.Format("2006-01-02 15:04:05"))
		}
	}

	// Security note: Key value is masked after creation
	if key.Key != "" {
		fmt.Printf("\nKey Value:             %s\n", key.Key)
	} else {
		fmt.Printf("\nNote: Key value is only displayed once during creation.\n")
	}

	return nil
}

// createSetupKey creates a new setup key
func (c *Client) createSetupKey(name, keyType string, expiresIn int, autoGroups []string, usageLimit int, ephemeral, allowExtraDNSLabels bool) error {
	// Validate key type
	if keyType != "one-off" && keyType != "reusable" {
		return fmt.Errorf("invalid key type: %s (must be one-off or reusable)", keyType)
	}

	// Create request
	req := SetupKeyCreateRequest{
		Name:                name,
		Type:                keyType,
		ExpiresIn:           expiresIn,
		AutoGroups:          autoGroups,
		UsageLimit:          usageLimit,
		Ephemeral:           ephemeral,
		AllowExtraDNSLabels: allowExtraDNSLabels,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/setup-keys", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var key SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Display success message with key details
	fmt.Printf("✓ Setup key created successfully!\n\n")
	fmt.Printf("Key ID:       %s\n", key.ID)
	fmt.Printf("Name:         %s\n", key.Name)
	fmt.Printf("Type:         %s\n", key.Type)
	fmt.Printf("Expires:      %s (%s)\n", formatExpiration(key.Expires), formatDuration(expiresIn))

	if usageLimit > 0 {
		fmt.Printf("Usage Limit:  %d\n", usageLimit)
	} else {
		fmt.Printf("Usage Limit:  Unlimited\n")
	}

	if len(autoGroups) > 0 {
		fmt.Printf("Auto-Groups:  %s\n", strings.Join(autoGroups, ", "))
	}

	// Display the key value - CRITICAL: Only shown once!
	if key.Key != "" {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("SETUP KEY (save this now - won't be shown again!):\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("%s\n", key.Key)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		fmt.Printf("Use this key to register new peers:\n")
		fmt.Printf("  netbird up --setup-key %s\n\n", key.Key)
	}

	return nil
}

// updateSetupKeyRevocation updates the revocation status of a setup key
func (c *Client) updateSetupKeyRevocation(keyID string, revoked bool) error {
	// First get the current key to retrieve auto-groups
	resp, err := c.makeRequest("GET", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentKey SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&currentKey); err != nil {
		return fmt.Errorf("failed to decode current key: %v", err)
	}

	// Create update request
	updateReq := SetupKeyUpdateRequest{
		Revoked:    revoked,
		AutoGroups: currentKey.AutoGroups,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/setup-keys/"+keyID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if revoked {
		fmt.Printf("✓ Setup key %s has been revoked.\n", keyID)
	} else {
		fmt.Printf("✓ Setup key %s has been enabled (un-revoked).\n", keyID)
	}

	return nil
}

// updateSetupKeyGroups updates the auto-groups for a setup key
func (c *Client) updateSetupKeyGroups(keyID string, newGroups []string) error {
	// First get the current key to retrieve revoked status
	resp, err := c.makeRequest("GET", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentKey SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&currentKey); err != nil {
		return fmt.Errorf("failed to decode current key: %v", err)
	}

	// Create update request
	updateReq := SetupKeyUpdateRequest{
		Revoked:    currentKey.Revoked,
		AutoGroups: newGroups,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/setup-keys/"+keyID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Auto-groups updated for setup key %s.\n", keyID)
	if len(newGroups) == 0 {
		fmt.Printf("  No auto-groups assigned.\n")
	} else {
		fmt.Printf("  New groups: %s\n", strings.Join(newGroups, ", "))
	}

	return nil
}

// deleteSetupKey deletes a setup key
func (c *Client) deleteSetupKey(keyID string) error {
	// First get the key details to show confirmation info
	resp, err := c.makeRequest("GET", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Perform deletion
	resp, err = c.makeRequest("DELETE", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Setup key %s has been deleted.\n", keyID)
	return nil
}
