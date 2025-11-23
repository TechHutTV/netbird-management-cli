// setup_keys.go
package commands

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

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// HandleSetupKeysCommand routes setup-key related commands using the flag package
func (s *Service) HandleSetupKeysCommand(args []string) error {
	// Create a new flag set for the 'setup-key' command
	setupKeyCmd := flag.NewFlagSet("setup-key", flag.ContinueOnError)
	setupKeyCmd.SetOutput(os.Stderr)
	setupKeyCmd.Usage = PrintSetupKeyUsage

	// Query flags
	listFlag := setupKeyCmd.Bool("list", false, "List all setup keys")
	inspectFlag := setupKeyCmd.String("inspect", "", "Inspect a setup key by its ID")
	filterNameFlag := setupKeyCmd.String("filter-name", "", "Filter by name pattern (use with --list)")
	filterTypeFlag := setupKeyCmd.String("filter-type", "", "Filter by type: one-off or reusable (use with --list)")
	validOnlyFlag := setupKeyCmd.Bool("valid-only", false, "Show only valid keys (use with --list)")
	outputFlag := setupKeyCmd.String("output", "table", "Output format: table or json")

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

	// Delete flags
	deleteFlag := setupKeyCmd.String("delete", "", "Delete a setup key by its ID")
	deleteBatchFlag := setupKeyCmd.String("delete-batch", "", "Delete multiple setup keys (comma-separated IDs)")
	deleteAllFlag := setupKeyCmd.Bool("delete-all", false, "Delete all setup keys")

	// If no flags provided, show usage
	if len(args) == 1 {
		PrintSetupKeyUsage()
		return nil
	}

	// Parse the flags
	if err := setupKeyCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags
	if *listFlag {
		return s.listSetupKeys(*filterNameFlag, *filterTypeFlag, *validOnlyFlag, *outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectSetupKey(*inspectFlag, *outputFlag)
	}

	if *createFlag != "" {
		expiresInSec, err := helpers.ParseDuration(*expiresInFlag, helpers.SetupKeyDurationBounds())
		if err != nil {
			return fmt.Errorf("invalid expiration duration: %v", err)
		}
		// Resolve group names/IDs to IDs
		groupIdentifiers := helpers.SplitCommaList(*autoGroupsFlag)
		autoGroupIDs, err := s.resolveMultipleGroupIdentifiers(groupIdentifiers)
		if err != nil {
			return fmt.Errorf("failed to resolve auto-groups: %v", err)
		}
		return s.createSetupKey(*createFlag, *keyTypeFlag, expiresInSec, autoGroupIDs, *usageLimitFlag, *ephemeralFlag, *allowExtraDNSLabelsFlag)
	}

	if *quickFlag != "" {
		// Quick create with sensible defaults
		return s.createSetupKey(*quickFlag, "one-off", 7*24*3600, []string{}, 1, false, false)
	}

	if *revokeFlag != "" {
		return s.updateSetupKeyRevocation(*revokeFlag, true)
	}

	if *enableFlag != "" {
		return s.updateSetupKeyRevocation(*enableFlag, false)
	}

	if *updateGroupsFlag != "" {
		if *groupsFlag == "" {
			return fmt.Errorf("flag --update-groups requires --groups")
		}
		// Resolve group names/IDs to IDs
		groupIdentifiers := helpers.SplitCommaList(*groupsFlag)
		newGroupIDs, err := s.resolveMultipleGroupIdentifiers(groupIdentifiers)
		if err != nil {
			return fmt.Errorf("failed to resolve groups: %v", err)
		}
		return s.updateSetupKeyGroups(*updateGroupsFlag, newGroupIDs)
	}

	if *deleteFlag != "" {
		return s.deleteSetupKey(*deleteFlag)
	}

	if *deleteBatchFlag != "" {
		return s.deleteSetupKeysBatch(*deleteBatchFlag)
	}

	if *deleteAllFlag {
		return s.deleteAllSetupKeys()
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'setup-key' command.")
	PrintSetupKeyUsage()
	return nil
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
func (s *Service) listSetupKeys(filterName, filterType string, validOnly bool, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/setup-keys", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var keys []models.SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// Apply filters
	var filtered []models.SetupKey
	for _, key := range keys {
		// Filter by name
		if filterName != "" && !helpers.MatchesPattern(key.Name, filterName) {
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

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
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
func (s *Service) inspectSetupKey(keyID string, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var key models.SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(key, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
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
func (s *Service) createSetupKey(name, keyType string, expiresIn int, autoGroups []string, usageLimit int, ephemeral, allowExtraDNSLabels bool) error {
	// Validate key type
	if keyType != "one-off" && keyType != "reusable" {
		return fmt.Errorf("invalid key type: %s (must be one-off or reusable)", keyType)
	}

	// Create request
	req := models.SetupKeyCreateRequest{
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

	resp, err := s.Client.MakeRequest("POST", "/setup-keys", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var key models.SetupKey
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
func (s *Service) updateSetupKeyRevocation(keyID string, revoked bool) error {
	// First get the current key to retrieve auto-groups
	resp, err := s.Client.MakeRequest("GET", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentKey models.SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&currentKey); err != nil {
		return fmt.Errorf("failed to decode current key: %v", err)
	}

	// Create update request
	updateReq := models.SetupKeyUpdateRequest{
		Revoked:    revoked,
		AutoGroups: currentKey.AutoGroups,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = s.Client.MakeRequest("PUT", "/setup-keys/"+keyID, bytes.NewReader(bodyBytes))
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
func (s *Service) updateSetupKeyGroups(keyID string, newGroups []string) error {
	// First get the current key to retrieve revoked status
	resp, err := s.Client.MakeRequest("GET", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentKey models.SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&currentKey); err != nil {
		return fmt.Errorf("failed to decode current key: %v", err)
	}

	// Create update request
	updateReq := models.SetupKeyUpdateRequest{
		Revoked:    currentKey.Revoked,
		AutoGroups: newGroups,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = s.Client.MakeRequest("PUT", "/setup-keys/"+keyID, bytes.NewReader(bodyBytes))
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
func (s *Service) deleteSetupKey(keyID string) error {
	// First get the key details to show confirmation info
	resp, err := s.Client.MakeRequest("GET", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var key models.SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return fmt.Errorf("failed to decode setup key: %v", err)
	}

	// Show confirmation prompt with key details
	details := map[string]string{
		"Type":  key.Type,
		"State": formatState(key.State, key.Valid, key.Revoked),
	}

	if !helpers.ConfirmSingleDeletion("setup key", key.Name, key.ID, details) {
		return nil
	}

	// Perform deletion
	resp, err = s.Client.MakeRequest("DELETE", "/setup-keys/"+keyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Setup key %s has been deleted\n", key.Name)
	return nil
}

// deleteSetupKeysBatch deletes multiple setup keys
func (s *Service) deleteSetupKeysBatch(idList string) error {
	keyIDs := helpers.SplitCommaList(idList)
	if len(keyIDs) == 0 {
		return fmt.Errorf("no setup key IDs provided")
	}

	// Fetch key details for confirmation
	keys := make([]models.SetupKey, 0, len(keyIDs))
	itemList := make([]string, 0, len(keyIDs))

	fmt.Println("Fetching setup key details...")
	for _, id := range keyIDs {
		resp, err := s.Client.MakeRequest("GET", "/setup-keys/"+id, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping %s: %v\n", id, err)
			continue
		}

		var key models.SetupKey
		if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
			resp.Body.Close()
			fmt.Fprintf(os.Stderr, "Warning: Skipping %s: failed to decode\n", id)
			continue
		}
		resp.Body.Close()

		keys = append(keys, key)
		state := formatState(key.State, key.Valid, key.Revoked)
		itemList = append(itemList, fmt.Sprintf("%s (ID: %s, Type: %s, State: %s)",
			key.Name, key.ID, key.Type, state))
	}

	if len(keys) == 0 {
		return fmt.Errorf("no valid setup keys found to delete")
	}

	// Confirm bulk deletion
	if !helpers.ConfirmBulkDeletion("setup keys", itemList, len(keys)) {
		return nil
	}

	// Process deletions with progress
	var succeeded, failed int
	for i, key := range keys {
		fmt.Printf("[%d/%d] Deleting setup key '%s'... ", i+1, len(keys), key.Name)

		resp, err := s.Client.MakeRequest("DELETE", "/setup-keys/"+key.ID, nil)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			failed++
			continue
		}
		resp.Body.Close()
		fmt.Println("Done")
		succeeded++
	}

	// Print summary
	fmt.Println()
	if failed > 0 {
		fmt.Fprintf(os.Stderr, "Warning: Completed: %d succeeded, %d failed\n", succeeded, failed)
	} else {
		fmt.Printf("All %d setup keys deleted successfully\n", succeeded)
	}

	return nil
}

// deleteAllSetupKeys deletes all setup keys with confirmation
func (s *Service) deleteAllSetupKeys() error {
	// First, get all setup keys
	resp, err := s.Client.MakeRequest("GET", "/setup-keys", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var keys []models.SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(keys) == 0 {
		fmt.Println("No setup keys found to delete.")
		return nil
	}

	// Build confirmation list
	keyList := make([]string, len(keys))
	for i, key := range keys {
		keyList[i] = fmt.Sprintf("%s (%s, %s)", key.Name, key.Type, formatState(key.State, key.Valid, key.Revoked))
	}

	// Prompt for bulk deletion confirmation
	if !helpers.ConfirmBulkDeletion("setup keys", keyList, len(keys)) {
		return nil
	}

	// Delete all keys
	fmt.Printf("\nDeleting %d setup key(s)...\n", len(keys))
	successCount := 0
	failCount := 0

	for _, key := range keys {
		resp, err := s.Client.MakeRequest("DELETE", "/setup-keys/"+key.ID, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to delete %s (%s): %v\n", key.Name, key.ID, err)
			failCount++
			continue
		}
		resp.Body.Close()

		fmt.Printf("✓ Deleted %s (%s)\n", key.Name, key.ID)
		successCount++
	}

	// Summary
	fmt.Printf("\nDeletion complete: %d successful, %d failed\n", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("failed to delete %d setup key(s)", failCount)
	}

	return nil
}
