// migrate.go - Peer migration between NetBird accounts
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"netbird-manage/internal/client"
	"netbird-manage/internal/config"
	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// MigrateOptions holds the configuration for a migration operation
type MigrateOptions struct {
	SourceToken string
	SourceURL   string
	DestToken   string
	DestURL     string
	PeerID      string
	GroupName   string
	CreateGroups bool
	KeyExpiry   string
	Cleanup     bool
}

// HandleMigrateCommand handles the migrate command for peer migration between accounts
func HandleMigrateCommand(args []string, debug bool) error {
	migrateCmd := flag.NewFlagSet("migrate", flag.ContinueOnError)
	migrateCmd.SetOutput(os.Stderr)
	migrateCmd.Usage = PrintMigrateUsage

	// Source account flags
	sourceToken := migrateCmd.String("source-token", "", "API token for the source account")
	sourceURL := migrateCmd.String("source-url", config.DefaultCloudURL, "Management URL for the source account")

	// Destination account flags
	destToken := migrateCmd.String("dest-token", "", "API token for the destination account")
	destURL := migrateCmd.String("dest-url", config.DefaultCloudURL, "Management URL for the destination account")

	// Migration target flags
	peerID := migrateCmd.String("peer", "", "Peer ID to migrate")
	groupName := migrateCmd.String("group", "", "Migrate all peers in this group")

	// Options
	createGroups := migrateCmd.Bool("create-groups", true, "Create missing groups in destination")
	keyExpiry := migrateCmd.String("key-expiry", "24h", "Setup key expiration duration (e.g., 1h, 24h, 7d)")
	cleanup := migrateCmd.Bool("cleanup", false, "Remove peer from source after generating migration command")

	if len(args) == 1 {
		PrintMigrateUsage()
		return nil
	}

	if err := migrateCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Validate required flags
	if *sourceToken == "" {
		return fmt.Errorf("--source-token is required")
	}
	if *destToken == "" {
		return fmt.Errorf("--dest-token is required")
	}
	if *peerID == "" && *groupName == "" {
		return fmt.Errorf("either --peer or --group is required")
	}

	opts := MigrateOptions{
		SourceToken:  *sourceToken,
		SourceURL:    *sourceURL,
		DestToken:    *destToken,
		DestURL:      *destURL,
		PeerID:       *peerID,
		GroupName:    *groupName,
		CreateGroups: *createGroups,
		KeyExpiry:    *keyExpiry,
		Cleanup:      *cleanup,
	}

	// Create clients for both accounts
	sourceClient := client.New(opts.SourceToken, opts.SourceURL)
	sourceClient.Debug = debug
	destClient := client.New(opts.DestToken, opts.DestURL)
	destClient.Debug = debug

	if *peerID != "" {
		return migrateSinglePeer(sourceClient, destClient, opts)
	}

	return migrateGroupPeers(sourceClient, destClient, opts)
}

// migrateSinglePeer handles migration of a single peer
func migrateSinglePeer(sourceClient, destClient *client.Client, opts MigrateOptions) error {
	fmt.Println("Fetching peer from source account...")
	fmt.Printf("  Source: %s\n\n", opts.SourceURL)

	// Fetch peer from source
	peer, err := getPeerByID(sourceClient, opts.PeerID)
	if err != nil {
		return fmt.Errorf("failed to fetch peer from source: %v", err)
	}

	// Display peer details
	displaySourcePeer(peer)

	fmt.Println("Connecting to destination account...")
	fmt.Printf("  Destination: %s\n\n", opts.DestURL)

	// Validate destination connection
	if err := validateConnection(destClient); err != nil {
		return fmt.Errorf("failed to connect to destination: %v", err)
	}

	// Get group names from peer
	groupNames := make([]string, len(peer.Groups))
	for i, g := range peer.Groups {
		groupNames[i] = g.Name
	}

	// Resolve/create groups in destination
	var autoGroupIDs []string
	var createdGroups []string
	if len(groupNames) > 0 && opts.CreateGroups {
		autoGroupIDs, createdGroups, err = resolveOrCreateGroups(destClient, groupNames)
		if err != nil {
			return fmt.Errorf("failed to resolve groups: %v", err)
		}
	}

	// Create setup key in destination
	fmt.Println("Creating setup key in destination...")
	keyName := fmt.Sprintf("migrate-%s-%s", peer.Name, time.Now().Format("20060102"))

	expiresIn, err := helpers.ParseDuration(opts.KeyExpiry, helpers.MigrationKeyDurationBounds())
	if err != nil {
		return fmt.Errorf("invalid key expiry: %v", err)
	}

	setupKey, err := createMigrationSetupKey(destClient, keyName, autoGroupIDs, expiresIn)
	if err != nil {
		return fmt.Errorf("failed to create setup key: %v", err)
	}

	// Display setup key info
	fmt.Printf("  Key Name:   %s\n", keyName)
	fmt.Printf("  Type:       one-off\n")
	if len(autoGroupIDs) > 0 {
		fmt.Printf("  Auto-Groups: %s\n", strings.Join(groupNames, ", "))
	}
	if len(createdGroups) > 0 {
		fmt.Printf("  Groups created in destination: %s\n", strings.Join(createdGroups, ", "))
	}
	fmt.Println("\nSetup key created successfully.")

	// Output the migration command
	outputMigrationCommand(peer, setupKey.Key, opts.DestURL)

	// Output cleanup command
	outputCleanupNote(peer, opts)

	return nil
}

// migrateGroupPeers handles migration of all peers in a group
func migrateGroupPeers(sourceClient, destClient *client.Client, opts MigrateOptions) error {
	fmt.Printf("Fetching peers in group '%s' from source...\n", opts.GroupName)
	fmt.Printf("  Source: %s\n\n", opts.SourceURL)

	// Find group and get its peers
	group, err := getGroupByName(sourceClient, opts.GroupName)
	if err != nil {
		return fmt.Errorf("failed to find group: %v", err)
	}

	if len(group.Peers) == 0 {
		fmt.Printf("No peers found in group '%s'.\n", opts.GroupName)
		return nil
	}

	fmt.Printf("Found %d peers to migrate.\n\n", len(group.Peers))

	fmt.Println("Connecting to destination account...")
	fmt.Printf("  Destination: %s\n\n", opts.DestURL)

	// Validate destination connection
	if err := validateConnection(destClient); err != nil {
		return fmt.Errorf("failed to connect to destination: %v", err)
	}

	// Collect all unique group names from all peers
	allGroupNames := make(map[string]bool)
	for _, peer := range group.Peers {
		for _, g := range peer.Groups {
			allGroupNames[g.Name] = true
		}
	}

	// Resolve/create all groups in destination first
	groupNameList := make([]string, 0, len(allGroupNames))
	for name := range allGroupNames {
		groupNameList = append(groupNameList, name)
	}

	var groupIDMap map[string]string
	var createdGroups []string
	if opts.CreateGroups && len(groupNameList) > 0 {
		groupIDMap, createdGroups, err = resolveOrCreateGroupsMap(destClient, groupNameList)
		if err != nil {
			return fmt.Errorf("failed to resolve groups: %v", err)
		}
		if len(createdGroups) > 0 {
			fmt.Printf("Groups created in destination: %s\n\n", strings.Join(createdGroups, ", "))
		}
	}

	// Parse key expiry once
	expiresIn, err := helpers.ParseDuration(opts.KeyExpiry, helpers.MigrationKeyDurationBounds())
	if err != nil {
		return fmt.Errorf("invalid key expiry: %v", err)
	}

	// Create setup keys for each peer
	type migrationInfo struct {
		Peer     models.Peer
		SetupKey string
	}
	var migrations []migrationInfo

	for i, peer := range group.Peers {
		fmt.Printf("Peer %d/%d: %s\n", i+1, len(group.Peers), peer.Name)

		// Get auto-groups for this peer
		var autoGroupIDs []string
		if groupIDMap != nil {
			for _, g := range peer.Groups {
				if id, ok := groupIDMap[g.Name]; ok {
					autoGroupIDs = append(autoGroupIDs, id)
				}
			}
		}

		keyName := fmt.Sprintf("migrate-%s-%s", peer.Name, time.Now().Format("20060102"))
		setupKey, err := createMigrationSetupKey(destClient, keyName, autoGroupIDs, expiresIn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Failed to create setup key: %v\n", err)
			continue
		}
		fmt.Printf("  Creating setup key... Done\n")

		migrations = append(migrations, migrationInfo{
			Peer:     peer,
			SetupKey: setupKey.Key,
		})
	}

	// Output all migration commands
	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("MIGRATION COMMANDS - Run on each peer:")
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println()

	for _, m := range migrations {
		fmt.Printf("# %s (%s)\n", m.Peer.Name, m.Peer.IP)
		outputMigrationCommandInline(m.Peer, m.SetupKey, opts.DestURL)
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 72))

	return nil
}

// getPeerByID fetches a peer by ID from the given client
func getPeerByID(c *client.Client, peerID string) (*models.Peer, error) {
	resp, err := c.MakeRequest("GET", "/peers/"+peerID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var peer models.Peer
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		return nil, fmt.Errorf("failed to decode peer: %v", err)
	}
	return &peer, nil
}

// getGroupByName finds a group by name from the source account
func getGroupByName(c *client.Client, name string) (*models.GroupDetail, error) {
	resp, err := c.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("failed to decode groups: %v", err)
	}

	for _, g := range groups {
		if strings.EqualFold(g.Name, name) {
			// Need to fetch full details to get peers
			return getGroupByID(c, g.ID)
		}
	}

	return nil, fmt.Errorf("group '%s' not found", name)
}

// getGroupByID fetches full group details
func getGroupByID(c *client.Client, groupID string) (*models.GroupDetail, error) {
	resp, err := c.MakeRequest("GET", "/groups/"+groupID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var group models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode group: %v", err)
	}
	return &group, nil
}

// validateConnection tests if the client can connect to the API
func validateConnection(c *client.Client) error {
	resp, err := c.MakeRequest("GET", "/peers", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// resolveOrCreateGroups resolves group names to IDs, creating missing groups
func resolveOrCreateGroups(c *client.Client, groupNames []string) ([]string, []string, error) {
	// Get existing groups from destination
	resp, err := c.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var existingGroups []models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&existingGroups); err != nil {
		return nil, nil, fmt.Errorf("failed to decode groups: %v", err)
	}

	// Build map of existing groups
	existingMap := make(map[string]string) // name -> id
	for _, g := range existingGroups {
		existingMap[strings.ToLower(g.Name)] = g.ID
	}

	var groupIDs []string
	var created []string

	for _, name := range groupNames {
		if id, exists := existingMap[strings.ToLower(name)]; exists {
			groupIDs = append(groupIDs, id)
		} else {
			// Create the group
			newGroup, err := createGroup(c, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to create group '%s': %v\n", name, err)
				continue
			}
			groupIDs = append(groupIDs, newGroup.ID)
			created = append(created, name)
		}
	}

	return groupIDs, created, nil
}

// resolveOrCreateGroupsMap returns a map of name -> ID for all groups
func resolveOrCreateGroupsMap(c *client.Client, groupNames []string) (map[string]string, []string, error) {
	// Get existing groups from destination
	resp, err := c.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var existingGroups []models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&existingGroups); err != nil {
		return nil, nil, fmt.Errorf("failed to decode groups: %v", err)
	}

	// Build map of existing groups
	existingMap := make(map[string]string) // lowercase name -> id
	nameMap := make(map[string]string)     // lowercase name -> actual name
	for _, g := range existingGroups {
		lower := strings.ToLower(g.Name)
		existingMap[lower] = g.ID
		nameMap[lower] = g.Name
	}

	result := make(map[string]string)
	var created []string

	for _, name := range groupNames {
		lower := strings.ToLower(name)
		if id, exists := existingMap[lower]; exists {
			result[name] = id
		} else {
			// Create the group
			newGroup, err := createGroup(c, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to create group '%s': %v\n", name, err)
				continue
			}
			result[name] = newGroup.ID
			created = append(created, name)
		}
	}

	return result, created, nil
}

// createGroup creates a new group in the destination account
func createGroup(c *client.Client, name string) (*models.GroupDetail, error) {
	reqBody := map[string]interface{}{
		"name":  name,
		"peers": []string{},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.MakeRequest("POST", "/groups", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var group models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

// createMigrationSetupKey creates a one-off setup key for migration
func createMigrationSetupKey(c *client.Client, name string, autoGroups []string, expiresIn int) (*models.SetupKey, error) {
	if autoGroups == nil {
		autoGroups = []string{}
	}

	req := models.SetupKeyCreateRequest{
		Name:       name,
		Type:       "one-off",
		ExpiresIn:  expiresIn,
		AutoGroups: autoGroups,
		UsageLimit: 1,
		Ephemeral:  false,
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.MakeRequest("POST", "/setup-keys", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var key models.SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&key); err != nil {
		return nil, err
	}

	return &key, nil
}

// displaySourcePeer displays peer details from source
func displaySourcePeer(peer *models.Peer) {
	fmt.Println("Source Peer Details:")
	fmt.Printf("  Name:       %s\n", peer.Name)
	fmt.Printf("  Hostname:   %s\n", peer.Hostname)
	fmt.Printf("  ID:         %s\n", peer.ID)
	fmt.Printf("  IP:         %s\n", peer.IP)
	fmt.Printf("  OS:         %s\n", helpers.FormatOS(peer.OS))
	fmt.Printf("  Version:    %s\n", peer.Version)

	if len(peer.Groups) > 0 {
		groupNames := make([]string, len(peer.Groups))
		for i, g := range peer.Groups {
			groupNames[i] = g.Name
		}
		fmt.Printf("  Groups:     %s\n", strings.Join(groupNames, ", "))
	} else {
		fmt.Printf("  Groups:     (none)\n")
	}
	fmt.Println()
}

// outputMigrationCommand outputs the migration command for a peer
func outputMigrationCommand(peer *models.Peer, setupKey, destURL string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println("MIGRATION COMMAND - Run this on the peer device:")
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println()

	// Build the command
	cmd := fmt.Sprintf("  sudo netbird down && sudo netbird up --setup-key %s --hostname %s",
		setupKey, peer.Hostname)

	// Add management URL if not the default cloud
	if destURL != config.DefaultCloudURL {
		// Extract management URL (remove /api suffix for netbird client)
		mgmtURL := strings.TrimSuffix(destURL, "/api")
		cmd += fmt.Sprintf(" \\\n    --management-url %s", mgmtURL)
	}

	fmt.Println(cmd)
	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
}

// outputMigrationCommandInline outputs the command on one/two lines for batch output
func outputMigrationCommandInline(peer models.Peer, setupKey, destURL string) {
	cmd := fmt.Sprintf("sudo netbird down && sudo netbird up --setup-key %s --hostname %s",
		setupKey, peer.Hostname)

	if destURL != config.DefaultCloudURL {
		mgmtURL := strings.TrimSuffix(destURL, "/api")
		cmd += fmt.Sprintf(" --management-url %s", mgmtURL)
	}

	fmt.Println(cmd)
}

// outputCleanupNote outputs notes about post-migration cleanup
func outputCleanupNote(peer *models.Peer, opts MigrateOptions) {
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - The peer will disconnect from the source network")
	fmt.Println("  - A new peer ID and IP will be assigned in the destination")

	// Calculate expiry
	expiresIn, _ := helpers.ParseDuration(opts.KeyExpiry, helpers.MigrationKeyDurationBounds())
	hours := expiresIn / 3600
	if hours >= 24 {
		days := hours / 24
		fmt.Printf("  - The setup key expires in %d day(s) and is single-use\n", days)
	} else {
		fmt.Printf("  - The setup key expires in %d hour(s) and is single-use\n", hours)
	}
	fmt.Println("  - Old peer entry in source account must be manually removed")
	fmt.Println()

	// Show cleanup command
	fmt.Println("To remove the old peer from source after migration:")

	cleanupCmd := fmt.Sprintf("  netbird-manage peer --remove %s", peer.ID)
	if opts.SourceURL != config.DefaultCloudURL {
		// Need to use the source token for cleanup
		fmt.Println("  # First, connect to source account:")
		fmt.Printf("  netbird-manage connect --token \"<source-token>\" --management-url \"%s\"\n", opts.SourceURL)
		fmt.Println("  # Then remove the peer:")
	}
	fmt.Println(cleanupCmd)
}

// PrintMigrateUsage displays help for the migrate command
func PrintMigrateUsage() {
	fmt.Println("Usage: netbird-manage migrate [options]")
	fmt.Println("\nMigrate peers between NetBird accounts.")
	fmt.Println("\nThis command generates migration commands that can be run on peer devices")
	fmt.Println("to move them from one NetBird account to another.")
	fmt.Println()
	fmt.Println("Required Flags:")
	fmt.Println("  --source-token <token>       API token for the source (exporting) account")
	fmt.Println("  --dest-token <token>         API token for the destination (importing) account")
	fmt.Println()
	fmt.Println("  --peer <peer-id>             Migrate a single peer by ID")
	fmt.Println("    OR")
	fmt.Println("  --group <group-name>         Migrate all peers in a group")
	fmt.Println()
	fmt.Println("Optional Flags:")
	fmt.Println("  --source-url <url>           Source management URL (default: NetBird Cloud)")
	fmt.Println("  --dest-url <url>             Destination management URL (default: NetBird Cloud)")
	fmt.Println("  --create-groups              Create missing groups in destination (default: true)")
	fmt.Println("  --key-expiry <duration>      Setup key expiration: 1h, 24h, 7d (default: 24h)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println()
	fmt.Println("  # Migrate a single peer between cloud accounts:")
	fmt.Println("  netbird-manage migrate \\")
	fmt.Println("    --source-token \"nbp_source...\" \\")
	fmt.Println("    --dest-token \"nbp_dest...\" \\")
	fmt.Println("    --peer \"abc123def\"")
	fmt.Println()
	fmt.Println("  # Migrate from cloud to self-hosted:")
	fmt.Println("  netbird-manage migrate \\")
	fmt.Println("    --source-token \"nbp_cloud...\" \\")
	fmt.Println("    --dest-token \"nbp_selfhost...\" \\")
	fmt.Println("    --dest-url \"https://netbird.mycompany.com/api\" \\")
	fmt.Println("    --peer \"abc123def\"")
	fmt.Println()
	fmt.Println("  # Migrate all peers in a group:")
	fmt.Println("  netbird-manage migrate \\")
	fmt.Println("    --source-token \"nbp_source...\" \\")
	fmt.Println("    --dest-token \"nbp_dest...\" \\")
	fmt.Println("    --group \"production-servers\"")
}
