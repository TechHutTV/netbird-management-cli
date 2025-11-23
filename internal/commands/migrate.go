// migrate.go - Full migration between NetBird accounts
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
	SourceToken  string
	SourceURL    string
	DestToken    string
	DestURL      string
	PeerID       string
	GroupName    string
	CreateGroups bool
	KeyExpiry    string
	Cleanup      bool
	// Full configuration migration options
	MigrateConfig   bool
	MigrateGroups   bool
	MigratePolicies bool
	MigrateNetworks bool
	MigrateRoutes   bool
	MigrateDNS      bool
	MigratePosture  bool
	MigrateSetupKeys bool
	SkipExisting    bool
	Update          bool
	DryRun          bool
	Verbose         bool
}

// HandleMigrateCommand handles the migrate command for peer and configuration migration between accounts
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

	// Peer migration target flags
	peerID := migrateCmd.String("peer", "", "Peer ID to migrate")
	groupName := migrateCmd.String("group", "", "Migrate all peers in this group")

	// Peer migration options
	createGroups := migrateCmd.Bool("create-groups", true, "Create missing groups in destination")
	keyExpiry := migrateCmd.String("key-expiry", "24h", "Setup key expiration duration (e.g., 1h, 24h, 7d)")
	cleanup := migrateCmd.Bool("cleanup", false, "Remove peer from source after generating migration command")

	// Full configuration migration flag
	migrateConfig := migrateCmd.Bool("config", false, "Migrate configuration (groups, policies, networks, routes, DNS, posture checks)")
	migrateAll := migrateCmd.Bool("all", false, "Migrate everything (configuration + generate peer migration commands)")

	// Selective configuration migration flags
	migrateGroupsOnly := migrateCmd.Bool("groups", false, "Migrate only groups")
	migratePoliciesOnly := migrateCmd.Bool("policies", false, "Migrate only policies")
	migrateNetworksOnly := migrateCmd.Bool("networks", false, "Migrate only networks")
	migrateRoutesOnly := migrateCmd.Bool("routes", false, "Migrate only routes")
	migrateDNSOnly := migrateCmd.Bool("dns", false, "Migrate only DNS nameserver groups")
	migratePostureOnly := migrateCmd.Bool("posture-checks", false, "Migrate only posture checks")
	migrateSetupKeysOnly := migrateCmd.Bool("setup-keys", false, "Migrate only setup keys")

	// Configuration migration options
	skipExisting := migrateCmd.Bool("skip-existing", false, "Skip resources that already exist in destination")
	update := migrateCmd.Bool("update", false, "Update existing resources in destination")
	dryRun := migrateCmd.Bool("dry-run", false, "Preview changes without applying them")
	verbose := migrateCmd.Bool("verbose", false, "Show detailed output")

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

	// Determine migration type
	isConfigMigration := *migrateConfig || *migrateAll ||
		*migrateGroupsOnly || *migratePoliciesOnly || *migrateNetworksOnly ||
		*migrateRoutesOnly || *migrateDNSOnly || *migratePostureOnly || *migrateSetupKeysOnly

	isPeerMigration := *peerID != "" || *groupName != ""

	// If neither config nor peer migration specified, require one
	if !isConfigMigration && !isPeerMigration {
		return fmt.Errorf("specify migration type: --config, --all, --peer, --group, or specific resource flags (--groups, --policies, etc.)")
	}

	// Determine which resources to migrate for config migration
	migrateGroups := *migrateConfig || *migrateAll || *migrateGroupsOnly
	migratePolicies := *migrateConfig || *migrateAll || *migratePoliciesOnly
	migrateNetworks := *migrateConfig || *migrateAll || *migrateNetworksOnly
	migrateRoutes := *migrateConfig || *migrateAll || *migrateRoutesOnly
	migrateDNS := *migrateConfig || *migrateAll || *migrateDNSOnly
	migratePosture := *migrateConfig || *migrateAll || *migratePostureOnly
	migrateSetupKeys := *migrateConfig || *migrateAll || *migrateSetupKeysOnly

	opts := MigrateOptions{
		SourceToken:      *sourceToken,
		SourceURL:        *sourceURL,
		DestToken:        *destToken,
		DestURL:          *destURL,
		PeerID:           *peerID,
		GroupName:        *groupName,
		CreateGroups:     *createGroups,
		KeyExpiry:        *keyExpiry,
		Cleanup:          *cleanup,
		MigrateConfig:    isConfigMigration,
		MigrateGroups:    migrateGroups,
		MigratePolicies:  migratePolicies,
		MigrateNetworks:  migrateNetworks,
		MigrateRoutes:    migrateRoutes,
		MigrateDNS:       migrateDNS,
		MigratePosture:   migratePosture,
		MigrateSetupKeys: migrateSetupKeys,
		SkipExisting:     *skipExisting,
		Update:           *update,
		DryRun:           *dryRun,
		Verbose:          *verbose,
	}

	// Create clients for both accounts
	sourceClient := client.New(opts.SourceToken, opts.SourceURL)
	sourceClient.Debug = debug
	destClient := client.New(opts.DestToken, opts.DestURL)
	destClient.Debug = debug

	// Handle configuration migration first if requested
	if isConfigMigration {
		if err := migrateConfiguration(sourceClient, destClient, opts); err != nil {
			return err
		}
	}

	// Handle peer migration if requested (either standalone or with --all)
	if isPeerMigration || *migrateAll {
		// For --all without explicit peer/group, migrate all peers
		if *migrateAll && !isPeerMigration {
			return migrateAllPeers(sourceClient, destClient, opts)
		}

		if *peerID != "" {
			return migrateSinglePeer(sourceClient, destClient, opts)
		}

		if *groupName != "" {
			return migrateGroupPeers(sourceClient, destClient, opts)
		}
	}

	return nil
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

// MigrateContext holds the state for a configuration migration
type MigrateContext struct {
	SourceClient *client.Client
	DestClient   *client.Client
	Opts         MigrateOptions

	// Source state
	SourceGroups        []models.GroupDetail
	SourcePolicies      []models.Policy
	SourceNetworks      []models.Network
	SourceRoutes        []models.Route
	SourceDNS           []models.DNSNameserverGroup
	SourcePostureChecks []models.PostureCheck
	SourceSetupKeys     []models.SetupKey
	SourcePeers         []models.Peer

	// Destination state
	DestGroups        map[string]*models.GroupDetail
	DestPolicies      map[string]*models.Policy
	DestNetworks      map[string]*models.Network
	DestDNS           map[string]*models.DNSNameserverGroup
	DestPostureChecks map[string]*models.PostureCheck
	DestSetupKeys     map[string]*models.SetupKey
	DestPeers         map[string]*models.Peer

	// Mappings (source name -> destination ID)
	GroupNameToDestID   map[string]string
	PostureNameToDestID map[string]string

	// Results
	Created []string
	Updated []string
	Skipped []string
	Failed  []string
}

// migrateConfiguration handles full configuration migration between accounts
func migrateConfiguration(sourceClient, destClient *client.Client, opts MigrateOptions) error {
	ctx := &MigrateContext{
		SourceClient:        sourceClient,
		DestClient:          destClient,
		Opts:                opts,
		DestGroups:          make(map[string]*models.GroupDetail),
		DestPolicies:        make(map[string]*models.Policy),
		DestNetworks:        make(map[string]*models.Network),
		DestDNS:             make(map[string]*models.DNSNameserverGroup),
		DestPostureChecks:   make(map[string]*models.PostureCheck),
		DestSetupKeys:       make(map[string]*models.SetupKey),
		DestPeers:           make(map[string]*models.Peer),
		GroupNameToDestID:   make(map[string]string),
		PostureNameToDestID: make(map[string]string),
	}

	if opts.DryRun {
		fmt.Println("Configuration Migration Preview (Dry Run)")
		fmt.Println("==========================================")
	} else {
		fmt.Println("Migrating Configuration...")
		fmt.Println("==========================")
	}

	fmt.Printf("  Source: %s\n", opts.SourceURL)
	fmt.Printf("  Destination: %s\n\n", opts.DestURL)

	// Fetch source and destination state
	fmt.Println("Fetching current state...")
	if err := ctx.fetchSourceState(); err != nil {
		return fmt.Errorf("failed to fetch source state: %v", err)
	}
	if err := ctx.fetchDestState(); err != nil {
		return fmt.Errorf("failed to fetch destination state: %v", err)
	}

	// Check for peer dependencies and warn if needed
	ctx.checkPeerDependencies()

	// Migrate resources in dependency order
	if opts.MigrateGroups {
		if err := ctx.migrateGroups(); err != nil {
			return err
		}
	}

	if opts.MigratePosture {
		if err := ctx.migratePostureChecks(); err != nil {
			return err
		}
	}

	if opts.MigratePolicies {
		if err := ctx.migratePolicies(); err != nil {
			return err
		}
	}

	if opts.MigrateRoutes {
		if err := ctx.migrateRoutes(); err != nil {
			return err
		}
	}

	if opts.MigrateDNS {
		if err := ctx.migrateDNS(); err != nil {
			return err
		}
	}

	if opts.MigrateNetworks {
		if err := ctx.migrateNetworks(); err != nil {
			return err
		}
	}

	if opts.MigrateSetupKeys {
		if err := ctx.migrateSetupKeys(); err != nil {
			return err
		}
	}

	// Print summary
	ctx.printMigrationSummary()

	return nil
}

// fetchSourceState fetches all resources from the source account
func (ctx *MigrateContext) fetchSourceState() error {
	var err error

	// Fetch peers first (needed for dependency checks)
	resp, err := ctx.SourceClient.MakeRequest("GET", "/peers", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch peers: %v", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&ctx.SourcePeers); err != nil {
		return fmt.Errorf("failed to decode peers: %v", err)
	}

	if ctx.Opts.MigrateGroups || ctx.Opts.MigratePolicies || ctx.Opts.MigrateNetworks || ctx.Opts.MigrateRoutes || ctx.Opts.MigrateDNS {
		resp, err := ctx.SourceClient.MakeRequest("GET", "/groups", nil)
		if err != nil {
			return fmt.Errorf("failed to fetch groups: %v", err)
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&ctx.SourceGroups); err != nil {
			return fmt.Errorf("failed to decode groups: %v", err)
		}
	}

	if ctx.Opts.MigratePolicies {
		resp, err := ctx.SourceClient.MakeRequest("GET", "/policies", nil)
		if err != nil {
			return fmt.Errorf("failed to fetch policies: %v", err)
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&ctx.SourcePolicies); err != nil {
			return fmt.Errorf("failed to decode policies: %v", err)
		}
	}

	if ctx.Opts.MigrateNetworks {
		resp, err := ctx.SourceClient.MakeRequest("GET", "/networks", nil)
		if err != nil {
			return fmt.Errorf("failed to fetch networks: %v", err)
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&ctx.SourceNetworks); err != nil {
			return fmt.Errorf("failed to decode networks: %v", err)
		}
	}

	if ctx.Opts.MigrateRoutes {
		resp, err := ctx.SourceClient.MakeRequest("GET", "/routes", nil)
		if err != nil {
			return fmt.Errorf("failed to fetch routes: %v", err)
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&ctx.SourceRoutes); err != nil {
			return fmt.Errorf("failed to decode routes: %v", err)
		}
	}

	if ctx.Opts.MigrateDNS {
		resp, err := ctx.SourceClient.MakeRequest("GET", "/dns/nameservers", nil)
		if err != nil {
			return fmt.Errorf("failed to fetch DNS: %v", err)
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&ctx.SourceDNS); err != nil {
			return fmt.Errorf("failed to decode DNS: %v", err)
		}
	}

	if ctx.Opts.MigratePosture || ctx.Opts.MigratePolicies {
		resp, err := ctx.SourceClient.MakeRequest("GET", "/posture-checks", nil)
		if err != nil {
			return fmt.Errorf("failed to fetch posture checks: %v", err)
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&ctx.SourcePostureChecks); err != nil {
			return fmt.Errorf("failed to decode posture checks: %v", err)
		}
	}

	if ctx.Opts.MigrateSetupKeys {
		resp, err := ctx.SourceClient.MakeRequest("GET", "/setup-keys", nil)
		if err != nil {
			return fmt.Errorf("failed to fetch setup keys: %v", err)
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&ctx.SourceSetupKeys); err != nil {
			return fmt.Errorf("failed to decode setup keys: %v", err)
		}
	}

	if ctx.Opts.Verbose {
		fmt.Printf("  Source: %d groups, %d policies, %d networks, %d routes, %d DNS, %d posture checks, %d setup keys, %d peers\n",
			len(ctx.SourceGroups), len(ctx.SourcePolicies), len(ctx.SourceNetworks),
			len(ctx.SourceRoutes), len(ctx.SourceDNS), len(ctx.SourcePostureChecks),
			len(ctx.SourceSetupKeys), len(ctx.SourcePeers))
	}

	return nil
}

// fetchDestState fetches all resources from the destination account
func (ctx *MigrateContext) fetchDestState() error {
	// Fetch destination peers
	resp, err := ctx.DestClient.MakeRequest("GET", "/peers", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch destination peers: %v", err)
	}
	defer resp.Body.Close()
	var destPeers []models.Peer
	if err := json.NewDecoder(resp.Body).Decode(&destPeers); err != nil {
		return fmt.Errorf("failed to decode destination peers: %v", err)
	}
	for _, peer := range destPeers {
		peerCopy := peer
		ctx.DestPeers[peer.Name] = &peerCopy
	}

	// Fetch destination groups
	resp, err = ctx.DestClient.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch destination groups: %v", err)
	}
	defer resp.Body.Close()
	var destGroups []models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&destGroups); err != nil {
		return fmt.Errorf("failed to decode destination groups: %v", err)
	}
	for _, group := range destGroups {
		groupCopy := group
		ctx.DestGroups[group.Name] = &groupCopy
		ctx.GroupNameToDestID[group.Name] = group.ID
	}

	// Fetch destination policies
	resp, err = ctx.DestClient.MakeRequest("GET", "/policies", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch destination policies: %v", err)
	}
	defer resp.Body.Close()
	var destPolicies []models.Policy
	if err := json.NewDecoder(resp.Body).Decode(&destPolicies); err != nil {
		return fmt.Errorf("failed to decode destination policies: %v", err)
	}
	for _, policy := range destPolicies {
		policyCopy := policy
		ctx.DestPolicies[policy.Name] = &policyCopy
	}

	// Fetch destination networks
	resp, err = ctx.DestClient.MakeRequest("GET", "/networks", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch destination networks: %v", err)
	}
	defer resp.Body.Close()
	var destNetworks []models.Network
	if err := json.NewDecoder(resp.Body).Decode(&destNetworks); err != nil {
		return fmt.Errorf("failed to decode destination networks: %v", err)
	}
	for _, network := range destNetworks {
		networkCopy := network
		ctx.DestNetworks[network.Name] = &networkCopy
	}

	// Fetch destination DNS
	resp, err = ctx.DestClient.MakeRequest("GET", "/dns/nameservers", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch destination DNS: %v", err)
	}
	defer resp.Body.Close()
	var destDNS []models.DNSNameserverGroup
	if err := json.NewDecoder(resp.Body).Decode(&destDNS); err != nil {
		return fmt.Errorf("failed to decode destination DNS: %v", err)
	}
	for _, dns := range destDNS {
		dnsCopy := dns
		ctx.DestDNS[dns.Name] = &dnsCopy
	}

	// Fetch destination posture checks
	resp, err = ctx.DestClient.MakeRequest("GET", "/posture-checks", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch destination posture checks: %v", err)
	}
	defer resp.Body.Close()
	var destPosture []models.PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&destPosture); err != nil {
		return fmt.Errorf("failed to decode destination posture checks: %v", err)
	}
	for _, check := range destPosture {
		checkCopy := check
		ctx.DestPostureChecks[check.Name] = &checkCopy
		ctx.PostureNameToDestID[check.Name] = check.ID
	}

	// Fetch destination setup keys
	resp, err = ctx.DestClient.MakeRequest("GET", "/setup-keys", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch destination setup keys: %v", err)
	}
	defer resp.Body.Close()
	var destSetupKeys []models.SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&destSetupKeys); err != nil {
		return fmt.Errorf("failed to decode destination setup keys: %v", err)
	}
	for _, key := range destSetupKeys {
		keyCopy := key
		ctx.DestSetupKeys[key.Name] = &keyCopy
	}

	if ctx.Opts.Verbose {
		fmt.Printf("  Destination: %d groups, %d policies, %d networks, %d DNS, %d posture checks, %d setup keys, %d peers\n",
			len(ctx.DestGroups), len(ctx.DestPolicies), len(ctx.DestNetworks),
			len(ctx.DestDNS), len(ctx.DestPostureChecks), len(ctx.DestSetupKeys), len(ctx.DestPeers))
	}
	fmt.Println()

	return nil
}

// checkPeerDependencies warns about resources that reference peers not yet migrated
func (ctx *MigrateContext) checkPeerDependencies() {
	// Build set of source peer IDs
	sourcePeerIDs := make(map[string]string) // ID -> name
	for _, peer := range ctx.SourcePeers {
		sourcePeerIDs[peer.ID] = peer.Name
	}

	// Build set of destination peer names
	destPeerNames := make(map[string]bool)
	for name := range ctx.DestPeers {
		destPeerNames[name] = true
	}

	// Check routes for peer references
	var missingPeers []string
	for _, route := range ctx.SourceRoutes {
		if route.Peer != "" {
			if name, ok := sourcePeerIDs[route.Peer]; ok {
				if !destPeerNames[name] {
					missingPeers = append(missingPeers, fmt.Sprintf("Route '%s' references peer '%s'", route.Description, name))
				}
			}
		}
	}

	// Check network routers for peer references
	for _, network := range ctx.SourceNetworks {
		// Would need to fetch routers to check peer references
		// For now, just warn about networks that have routers
		if len(network.Routers) > 0 {
			missingPeers = append(missingPeers, fmt.Sprintf("Network '%s' has %d routers that may reference peers", network.Name, len(network.Routers)))
		}
	}

	if len(missingPeers) > 0 {
		fmt.Println("⚠️  WARNING: Some resources reference peers")
		fmt.Println("================================================")
		fmt.Println("The following resources reference peers that may not exist in the destination:")
		for _, msg := range missingPeers {
			fmt.Printf("  - %s\n", msg)
		}
		fmt.Println()
		fmt.Println("Recommendation: Migrate peers first using:")
		fmt.Println("  netbird-manage migrate --source-token <token> --dest-token <token> --all")
		fmt.Println()
		fmt.Println("Or migrate specific peers/groups before configuration.")
		fmt.Println("================================================")
		fmt.Println()
	}
}

// migrateGroups migrates groups from source to destination
func (ctx *MigrateContext) migrateGroups() error {
	if len(ctx.SourceGroups) == 0 {
		return nil
	}

	fmt.Println("Groups:")

	for _, group := range ctx.SourceGroups {
		// Check if group exists in destination
		if existing, exists := ctx.DestGroups[group.Name]; exists {
			if ctx.Opts.SkipExisting {
				fmt.Printf("  SKIP     %s (already exists)\n", group.Name)
				ctx.Skipped = append(ctx.Skipped, "Group "+group.Name)
				continue
			}
			if !ctx.Opts.Update {
				fmt.Printf("  CONFLICT %s (already exists, use --update or --skip-existing)\n", group.Name)
				ctx.Failed = append(ctx.Failed, "Group "+group.Name+": already exists")
				continue
			}

			// Update existing group
			if ctx.Opts.DryRun {
				fmt.Printf("  UPDATE   %s (would update)\n", group.Name)
			} else {
				if err := ctx.updateGroup(group, existing.ID); err != nil {
					fmt.Printf("  FAILED   %s (%v)\n", group.Name, err)
					ctx.Failed = append(ctx.Failed, "Group "+group.Name+": "+err.Error())
					continue
				}
				fmt.Printf("  UPDATED  %s\n", group.Name)
				ctx.Updated = append(ctx.Updated, "Group "+group.Name)
			}
			continue
		}

		// Create new group
		if ctx.Opts.DryRun {
			fmt.Printf("  CREATE   %s (would create)\n", group.Name)
		} else {
			newID, err := ctx.createGroup(group)
			if err != nil {
				fmt.Printf("  FAILED   %s (%v)\n", group.Name, err)
				ctx.Failed = append(ctx.Failed, "Group "+group.Name+": "+err.Error())
				continue
			}
			fmt.Printf("  CREATED  %s\n", group.Name)
			ctx.Created = append(ctx.Created, "Group "+group.Name)
			ctx.GroupNameToDestID[group.Name] = newID
		}
	}

	fmt.Println()
	return nil
}

// createGroup creates a group in the destination
func (ctx *MigrateContext) createGroup(group models.GroupDetail) (string, error) {
	// Create group without peers (peers must be migrated separately)
	reqBody := map[string]interface{}{
		"name":  group.Name,
		"peers": []string{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("POST", "/groups", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API error: %s", resp.Status)
	}

	var created models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", err
	}

	return created.ID, nil
}

// updateGroup updates a group in the destination
func (ctx *MigrateContext) updateGroup(group models.GroupDetail, destID string) error {
	// Get existing group to preserve peers
	existing := ctx.DestGroups[group.Name]
	existingPeerIDs := []string{}
	for _, peer := range existing.Peers {
		existingPeerIDs = append(existingPeerIDs, peer.ID)
	}

	reqBody := models.GroupPutRequest{
		Name:      group.Name,
		Peers:     existingPeerIDs, // Preserve existing peers
		Resources: []models.GroupResourcePutRequest{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("PUT", "/groups/"+destID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// migratePostureChecks migrates posture checks from source to destination
func (ctx *MigrateContext) migratePostureChecks() error {
	if len(ctx.SourcePostureChecks) == 0 {
		return nil
	}

	fmt.Println("Posture Checks:")

	for _, check := range ctx.SourcePostureChecks {
		if existing, exists := ctx.DestPostureChecks[check.Name]; exists {
			if ctx.Opts.SkipExisting {
				fmt.Printf("  SKIP     %s (already exists)\n", check.Name)
				ctx.Skipped = append(ctx.Skipped, "Posture Check "+check.Name)
				ctx.PostureNameToDestID[check.Name] = existing.ID
				continue
			}
			if !ctx.Opts.Update {
				fmt.Printf("  CONFLICT %s (already exists)\n", check.Name)
				ctx.Failed = append(ctx.Failed, "Posture Check "+check.Name+": already exists")
				ctx.PostureNameToDestID[check.Name] = existing.ID
				continue
			}

			if ctx.Opts.DryRun {
				fmt.Printf("  UPDATE   %s (would update)\n", check.Name)
			} else {
				if err := ctx.updatePostureCheck(check, existing.ID); err != nil {
					fmt.Printf("  FAILED   %s (%v)\n", check.Name, err)
					ctx.Failed = append(ctx.Failed, "Posture Check "+check.Name+": "+err.Error())
					continue
				}
				fmt.Printf("  UPDATED  %s\n", check.Name)
				ctx.Updated = append(ctx.Updated, "Posture Check "+check.Name)
			}
			continue
		}

		if ctx.Opts.DryRun {
			fmt.Printf("  CREATE   %s (would create)\n", check.Name)
		} else {
			newID, err := ctx.createPostureCheck(check)
			if err != nil {
				fmt.Printf("  FAILED   %s (%v)\n", check.Name, err)
				ctx.Failed = append(ctx.Failed, "Posture Check "+check.Name+": "+err.Error())
				continue
			}
			fmt.Printf("  CREATED  %s\n", check.Name)
			ctx.Created = append(ctx.Created, "Posture Check "+check.Name)
			ctx.PostureNameToDestID[check.Name] = newID
		}
	}

	fmt.Println()
	return nil
}

// createPostureCheck creates a posture check in the destination
func (ctx *MigrateContext) createPostureCheck(check models.PostureCheck) (string, error) {
	reqBody := models.PostureCheckRequest{
		Name:        check.Name,
		Description: check.Description,
		Checks:      check.Checks,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("POST", "/posture-checks", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API error: %s", resp.Status)
	}

	var created models.PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", err
	}

	return created.ID, nil
}

// updatePostureCheck updates a posture check in the destination
func (ctx *MigrateContext) updatePostureCheck(check models.PostureCheck, destID string) error {
	reqBody := models.PostureCheckRequest{
		Name:        check.Name,
		Description: check.Description,
		Checks:      check.Checks,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("PUT", "/posture-checks/"+destID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// migratePolicies migrates policies from source to destination
func (ctx *MigrateContext) migratePolicies() error {
	if len(ctx.SourcePolicies) == 0 {
		return nil
	}

	fmt.Println("Policies:")

	for _, policy := range ctx.SourcePolicies {
		if _, exists := ctx.DestPolicies[policy.Name]; exists {
			if ctx.Opts.SkipExisting {
				fmt.Printf("  SKIP     %s (already exists)\n", policy.Name)
				ctx.Skipped = append(ctx.Skipped, "Policy "+policy.Name)
				continue
			}
			if !ctx.Opts.Update {
				fmt.Printf("  CONFLICT %s (already exists)\n", policy.Name)
				ctx.Failed = append(ctx.Failed, "Policy "+policy.Name+": already exists")
				continue
			}

			if ctx.Opts.DryRun {
				fmt.Printf("  UPDATE   %s (would update)\n", policy.Name)
			} else {
				if err := ctx.updatePolicy(policy); err != nil {
					fmt.Printf("  FAILED   %s (%v)\n", policy.Name, err)
					ctx.Failed = append(ctx.Failed, "Policy "+policy.Name+": "+err.Error())
					continue
				}
				fmt.Printf("  UPDATED  %s\n", policy.Name)
				ctx.Updated = append(ctx.Updated, "Policy "+policy.Name)
			}
			continue
		}

		if ctx.Opts.DryRun {
			fmt.Printf("  CREATE   %s (would create)\n", policy.Name)
		} else {
			if err := ctx.createPolicy(policy); err != nil {
				fmt.Printf("  FAILED   %s (%v)\n", policy.Name, err)
				ctx.Failed = append(ctx.Failed, "Policy "+policy.Name+": "+err.Error())
				continue
			}
			fmt.Printf("  CREATED  %s\n", policy.Name)
			ctx.Created = append(ctx.Created, "Policy "+policy.Name)
		}
	}

	fmt.Println()
	return nil
}

// createPolicy creates a policy in the destination
func (ctx *MigrateContext) createPolicy(policy models.Policy) error {
	// Convert rules with resolved group IDs
	rules := []models.PolicyRuleForWrite{}
	for _, rule := range policy.Rules {
		newRule := models.PolicyRuleForWrite{
			Name:          rule.Name,
			Description:   rule.Description,
			Enabled:       rule.Enabled,
			Action:        rule.Action,
			Bidirectional: rule.Bidirectional,
			Protocol:      rule.Protocol,
			Ports:         rule.Ports,
			PortRanges:    rule.PortRanges,
		}

		// Resolve source groups
		for _, src := range rule.Sources {
			if destID, ok := ctx.GroupNameToDestID[src.Name]; ok {
				newRule.Sources = append(newRule.Sources, destID)
			} else {
				return fmt.Errorf("source group '%s' not found in destination", src.Name)
			}
		}

		// Resolve destination groups
		for _, dest := range rule.Destinations {
			if destID, ok := ctx.GroupNameToDestID[dest.Name]; ok {
				newRule.Destinations = append(newRule.Destinations, destID)
			} else {
				return fmt.Errorf("destination group '%s' not found in destination", dest.Name)
			}
		}

		rules = append(rules, newRule)
	}

	// Resolve posture check IDs
	var postureCheckIDs []string
	for _, pcID := range policy.SourcePostureChecks {
		// Find posture check by ID in source, get name, then get dest ID
		for _, pc := range ctx.SourcePostureChecks {
			if pc.ID == pcID {
				if destID, ok := ctx.PostureNameToDestID[pc.Name]; ok {
					postureCheckIDs = append(postureCheckIDs, destID)
				}
				break
			}
		}
	}

	reqBody := models.PolicyCreateRequest{
		Name:                policy.Name,
		Description:         policy.Description,
		Enabled:             policy.Enabled,
		Rules:               rules,
		SourcePostureChecks: postureCheckIDs,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("POST", "/policies", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// updatePolicy updates a policy in the destination
func (ctx *MigrateContext) updatePolicy(policy models.Policy) error {
	destPolicy := ctx.DestPolicies[policy.Name]

	// Convert rules with resolved group IDs
	rules := []models.PolicyRuleForWrite{}
	for _, rule := range policy.Rules {
		newRule := models.PolicyRuleForWrite{
			Name:          rule.Name,
			Description:   rule.Description,
			Enabled:       rule.Enabled,
			Action:        rule.Action,
			Bidirectional: rule.Bidirectional,
			Protocol:      rule.Protocol,
			Ports:         rule.Ports,
			PortRanges:    rule.PortRanges,
		}

		for _, src := range rule.Sources {
			if destID, ok := ctx.GroupNameToDestID[src.Name]; ok {
				newRule.Sources = append(newRule.Sources, destID)
			}
		}

		for _, dest := range rule.Destinations {
			if destID, ok := ctx.GroupNameToDestID[dest.Name]; ok {
				newRule.Destinations = append(newRule.Destinations, destID)
			}
		}

		rules = append(rules, newRule)
	}

	reqBody := models.PolicyUpdateRequest{
		Name:        policy.Name,
		Description: policy.Description,
		Enabled:     policy.Enabled,
		Rules:       rules,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("PUT", "/policies/"+destPolicy.ID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// migrateRoutes migrates routes from source to destination
func (ctx *MigrateContext) migrateRoutes() error {
	if len(ctx.SourceRoutes) == 0 {
		return nil
	}

	fmt.Println("Routes:")

	for _, route := range ctx.SourceRoutes {
		routeName := route.Description
		if routeName == "" {
			routeName = route.Network
		}

		// Skip routes that reference specific peers (can't migrate peer references)
		if route.Peer != "" {
			fmt.Printf("  SKIP     %s (references peer, migrate peer first)\n", routeName)
			ctx.Skipped = append(ctx.Skipped, "Route "+routeName+": references peer")
			continue
		}

		if ctx.Opts.DryRun {
			fmt.Printf("  CREATE   %s (would create)\n", routeName)
		} else {
			if err := ctx.createRoute(route); err != nil {
				fmt.Printf("  FAILED   %s (%v)\n", routeName, err)
				ctx.Failed = append(ctx.Failed, "Route "+routeName+": "+err.Error())
				continue
			}
			fmt.Printf("  CREATED  %s\n", routeName)
			ctx.Created = append(ctx.Created, "Route "+routeName)
		}
	}

	fmt.Println()
	return nil
}

// createRoute creates a route in the destination
func (ctx *MigrateContext) createRoute(route models.Route) error {
	// Resolve group IDs
	var groupIDs []string
	for _, groupID := range route.Groups {
		// Find group by ID in source, get name, then get dest ID
		for _, g := range ctx.SourceGroups {
			if g.ID == groupID {
				if destID, ok := ctx.GroupNameToDestID[g.Name]; ok {
					groupIDs = append(groupIDs, destID)
				}
				break
			}
		}
	}

	// Resolve peer group IDs
	var peerGroupIDs []string
	for _, pgID := range route.PeerGroups {
		for _, g := range ctx.SourceGroups {
			if g.ID == pgID {
				if destID, ok := ctx.GroupNameToDestID[g.Name]; ok {
					peerGroupIDs = append(peerGroupIDs, destID)
				}
				break
			}
		}
	}

	reqBody := models.RouteRequest{
		Description: route.Description,
		NetworkID:   route.NetworkID,
		Network:     route.Network,
		Domains:     route.Domains,
		PeerGroups:  peerGroupIDs,
		Metric:      route.Metric,
		Masquerade:  route.Masquerade,
		Enabled:     route.Enabled,
		Groups:      groupIDs,
		KeepRoute:   route.KeepRoute,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("POST", "/routes", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// migrateDNS migrates DNS nameserver groups from source to destination
func (ctx *MigrateContext) migrateDNS() error {
	if len(ctx.SourceDNS) == 0 {
		return nil
	}

	fmt.Println("DNS Nameserver Groups:")

	for _, dns := range ctx.SourceDNS {
		if _, exists := ctx.DestDNS[dns.Name]; exists {
			if ctx.Opts.SkipExisting {
				fmt.Printf("  SKIP     %s (already exists)\n", dns.Name)
				ctx.Skipped = append(ctx.Skipped, "DNS "+dns.Name)
				continue
			}
			if !ctx.Opts.Update {
				fmt.Printf("  CONFLICT %s (already exists)\n", dns.Name)
				ctx.Failed = append(ctx.Failed, "DNS "+dns.Name+": already exists")
				continue
			}

			if ctx.Opts.DryRun {
				fmt.Printf("  UPDATE   %s (would update)\n", dns.Name)
			} else {
				if err := ctx.updateDNS(dns); err != nil {
					fmt.Printf("  FAILED   %s (%v)\n", dns.Name, err)
					ctx.Failed = append(ctx.Failed, "DNS "+dns.Name+": "+err.Error())
					continue
				}
				fmt.Printf("  UPDATED  %s\n", dns.Name)
				ctx.Updated = append(ctx.Updated, "DNS "+dns.Name)
			}
			continue
		}

		if ctx.Opts.DryRun {
			fmt.Printf("  CREATE   %s (would create)\n", dns.Name)
		} else {
			if err := ctx.createDNS(dns); err != nil {
				fmt.Printf("  FAILED   %s (%v)\n", dns.Name, err)
				ctx.Failed = append(ctx.Failed, "DNS "+dns.Name+": "+err.Error())
				continue
			}
			fmt.Printf("  CREATED  %s\n", dns.Name)
			ctx.Created = append(ctx.Created, "DNS "+dns.Name)
		}
	}

	fmt.Println()
	return nil
}

// createDNS creates a DNS nameserver group in the destination
func (ctx *MigrateContext) createDNS(dns models.DNSNameserverGroup) error {
	// Resolve group IDs
	var groupIDs []string
	for _, groupID := range dns.Groups {
		for _, g := range ctx.SourceGroups {
			if g.ID == groupID {
				if destID, ok := ctx.GroupNameToDestID[g.Name]; ok {
					groupIDs = append(groupIDs, destID)
				}
				break
			}
		}
	}

	reqBody := models.DNSNameserverGroupRequest{
		Name:                 dns.Name,
		Description:          dns.Description,
		Nameservers:          dns.Nameservers,
		Groups:               groupIDs,
		Domains:              dns.Domains,
		SearchDomainsEnabled: dns.SearchDomainsEnabled,
		Primary:              dns.Primary,
		Enabled:              dns.Enabled,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("POST", "/dns/nameservers", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// updateDNS updates a DNS nameserver group in the destination
func (ctx *MigrateContext) updateDNS(dns models.DNSNameserverGroup) error {
	destDNS := ctx.DestDNS[dns.Name]

	var groupIDs []string
	for _, groupID := range dns.Groups {
		for _, g := range ctx.SourceGroups {
			if g.ID == groupID {
				if destID, ok := ctx.GroupNameToDestID[g.Name]; ok {
					groupIDs = append(groupIDs, destID)
				}
				break
			}
		}
	}

	reqBody := models.DNSNameserverGroupRequest{
		Name:                 dns.Name,
		Description:          dns.Description,
		Nameservers:          dns.Nameservers,
		Groups:               groupIDs,
		Domains:              dns.Domains,
		SearchDomainsEnabled: dns.SearchDomainsEnabled,
		Primary:              dns.Primary,
		Enabled:              dns.Enabled,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("PUT", "/dns/nameservers/"+destDNS.ID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// migrateNetworks migrates networks from source to destination
func (ctx *MigrateContext) migrateNetworks() error {
	if len(ctx.SourceNetworks) == 0 {
		return nil
	}

	fmt.Println("Networks:")

	for _, network := range ctx.SourceNetworks {
		if _, exists := ctx.DestNetworks[network.Name]; exists {
			if ctx.Opts.SkipExisting {
				fmt.Printf("  SKIP     %s (already exists)\n", network.Name)
				ctx.Skipped = append(ctx.Skipped, "Network "+network.Name)
				continue
			}
			if !ctx.Opts.Update {
				fmt.Printf("  CONFLICT %s (already exists)\n", network.Name)
				ctx.Failed = append(ctx.Failed, "Network "+network.Name+": already exists")
				continue
			}

			if ctx.Opts.DryRun {
				fmt.Printf("  UPDATE   %s (would update)\n", network.Name)
			} else {
				if err := ctx.updateNetwork(network); err != nil {
					fmt.Printf("  FAILED   %s (%v)\n", network.Name, err)
					ctx.Failed = append(ctx.Failed, "Network "+network.Name+": "+err.Error())
					continue
				}
				fmt.Printf("  UPDATED  %s\n", network.Name)
				ctx.Updated = append(ctx.Updated, "Network "+network.Name)
			}
			continue
		}

		if ctx.Opts.DryRun {
			fmt.Printf("  CREATE   %s (would create)\n", network.Name)
		} else {
			if err := ctx.createNetwork(network); err != nil {
				fmt.Printf("  FAILED   %s (%v)\n", network.Name, err)
				ctx.Failed = append(ctx.Failed, "Network "+network.Name+": "+err.Error())
				continue
			}
			fmt.Printf("  CREATED  %s\n", network.Name)
			ctx.Created = append(ctx.Created, "Network "+network.Name)
		}
	}

	fmt.Println()
	return nil
}

// createNetwork creates a network in the destination
func (ctx *MigrateContext) createNetwork(network models.Network) error {
	reqBody := models.NetworkCreateRequest{
		Name:        network.Name,
		Description: network.Description,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("POST", "/networks", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// updateNetwork updates a network in the destination
func (ctx *MigrateContext) updateNetwork(network models.Network) error {
	destNetwork := ctx.DestNetworks[network.Name]

	reqBody := models.NetworkUpdateRequest{
		Name:        network.Name,
		Description: network.Description,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("PUT", "/networks/"+destNetwork.ID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// migrateSetupKeys migrates setup keys from source to destination
func (ctx *MigrateContext) migrateSetupKeys() error {
	if len(ctx.SourceSetupKeys) == 0 {
		return nil
	}

	fmt.Println("Setup Keys:")

	for _, key := range ctx.SourceSetupKeys {
		if _, exists := ctx.DestSetupKeys[key.Name]; exists {
			if ctx.Opts.SkipExisting {
				fmt.Printf("  SKIP     %s (already exists)\n", key.Name)
				ctx.Skipped = append(ctx.Skipped, "Setup Key "+key.Name)
				continue
			}
			fmt.Printf("  SKIP     %s (setup keys cannot be updated)\n", key.Name)
			ctx.Skipped = append(ctx.Skipped, "Setup Key "+key.Name)
			continue
		}

		if ctx.Opts.DryRun {
			fmt.Printf("  CREATE   %s (would create)\n", key.Name)
		} else {
			if err := ctx.createSetupKey(key); err != nil {
				fmt.Printf("  FAILED   %s (%v)\n", key.Name, err)
				ctx.Failed = append(ctx.Failed, "Setup Key "+key.Name+": "+err.Error())
				continue
			}
			fmt.Printf("  CREATED  %s\n", key.Name)
			ctx.Created = append(ctx.Created, "Setup Key "+key.Name)
		}
	}

	fmt.Println()
	return nil
}

// createSetupKey creates a setup key in the destination
func (ctx *MigrateContext) createSetupKey(key models.SetupKey) error {
	// Resolve auto-group IDs
	var autoGroupIDs []string
	for _, groupID := range key.AutoGroups {
		for _, g := range ctx.SourceGroups {
			if g.ID == groupID {
				if destID, ok := ctx.GroupNameToDestID[g.Name]; ok {
					autoGroupIDs = append(autoGroupIDs, destID)
				}
				break
			}
		}
	}

	reqBody := models.SetupKeyCreateRequest{
		Name:       key.Name,
		Type:       key.Type,
		ExpiresIn:  86400 * 30, // Default 30 days
		AutoGroups: autoGroupIDs,
		UsageLimit: key.UsageLimit,
		Ephemeral:  key.Ephemeral,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.DestClient.MakeRequest("POST", "/setup-keys", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// printMigrationSummary prints the migration summary
func (ctx *MigrateContext) printMigrationSummary() {
	fmt.Println("================================================")
	fmt.Println("Migration Summary")
	fmt.Println("================================================")
	fmt.Println()

	if len(ctx.Created) > 0 {
		fmt.Printf("✓ Created:  %d resources\n", len(ctx.Created))
		if ctx.Opts.Verbose {
			for _, res := range ctx.Created {
				fmt.Printf("    - %s\n", res)
			}
		}
	}

	if len(ctx.Updated) > 0 {
		fmt.Printf("✓ Updated:  %d resources\n", len(ctx.Updated))
		if ctx.Opts.Verbose {
			for _, res := range ctx.Updated {
				fmt.Printf("    - %s\n", res)
			}
		}
	}

	if len(ctx.Skipped) > 0 {
		fmt.Printf("⚠ Skipped:  %d resources\n", len(ctx.Skipped))
		if ctx.Opts.Verbose {
			for _, res := range ctx.Skipped {
				fmt.Printf("    - %s\n", res)
			}
		}
	}

	if len(ctx.Failed) > 0 {
		fmt.Printf("✗ Failed:   %d resources\n", len(ctx.Failed))
		fmt.Println()
		fmt.Println("Errors:")
		for i, msg := range ctx.Failed {
			fmt.Printf("  %d. %s\n", i+1, msg)
		}
	}

	fmt.Println()

	if ctx.Opts.DryRun {
		fmt.Println("This was a dry run. Use without --dry-run to apply changes.")
	} else {
		totalChanges := len(ctx.Created) + len(ctx.Updated)
		if totalChanges > 0 {
			fmt.Printf("Successfully migrated %d resources!\n", totalChanges)
		}
		if len(ctx.Failed) > 0 {
			fmt.Println("Some resources failed to migrate. Fix errors and re-run with --skip-existing")
		}
	}
}

// migrateAllPeers migrates all peers from source to destination
func migrateAllPeers(sourceClient, destClient *client.Client, opts MigrateOptions) error {
	fmt.Println()
	fmt.Println("Generating Peer Migration Commands...")
	fmt.Println("=====================================")
	fmt.Println()

	// Fetch all peers from source
	resp, err := sourceClient.MakeRequest("GET", "/peers", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch peers: %v", err)
	}
	defer resp.Body.Close()

	var peers []models.Peer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return fmt.Errorf("failed to decode peers: %v", err)
	}

	if len(peers) == 0 {
		fmt.Println("No peers found in source account.")
		return nil
	}

	fmt.Printf("Found %d peers to migrate.\n\n", len(peers))

	// Validate destination connection
	if err := validateConnection(destClient); err != nil {
		return fmt.Errorf("failed to connect to destination: %v", err)
	}

	// Collect all unique group names from all peers
	allGroupNames := make(map[string]bool)
	for _, peer := range peers {
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

	for i, peer := range peers {
		fmt.Printf("Peer %d/%d: %s\n", i+1, len(peers), peer.Name)

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

// PrintMigrateUsage displays help for the migrate command
func PrintMigrateUsage() {
	fmt.Println("Usage: netbird-manage migrate [options]")
	fmt.Println("\nMigrate peers and/or configuration between NetBird accounts.")
	fmt.Println("\nThis command supports two migration modes:")
	fmt.Println("  1. Peer Migration: Generates commands to move peers to a new account")
	fmt.Println("  2. Config Migration: Copies groups, policies, networks, routes, DNS, etc.")
	fmt.Println()
	fmt.Println("Required Flags:")
	fmt.Println("  --source-token <token>       API token for the source (exporting) account")
	fmt.Println("  --dest-token <token>         API token for the destination (importing) account")
	fmt.Println()
	fmt.Println("Migration Type (choose one or combine):")
	fmt.Println()
	fmt.Println("  Peer Migration:")
	fmt.Println("    --peer <peer-id>           Migrate a single peer by ID")
	fmt.Println("    --group <group-name>       Migrate all peers in a group")
	fmt.Println()
	fmt.Println("  Configuration Migration:")
	fmt.Println("    --config                   Migrate all configuration (groups, policies, networks,")
	fmt.Println("                               routes, DNS, posture checks, setup keys)")
	fmt.Println("    --all                      Migrate everything (config + all peers)")
	fmt.Println()
	fmt.Println("  Selective Configuration:")
	fmt.Println("    --groups                   Migrate only groups")
	fmt.Println("    --policies                 Migrate only policies")
	fmt.Println("    --networks                 Migrate only networks")
	fmt.Println("    --routes                   Migrate only routes")
	fmt.Println("    --dns                      Migrate only DNS nameserver groups")
	fmt.Println("    --posture-checks           Migrate only posture checks")
	fmt.Println("    --setup-keys               Migrate only setup keys")
	fmt.Println()
	fmt.Println("Configuration Options:")
	fmt.Println("  --skip-existing              Skip resources that already exist in destination")
	fmt.Println("  --update                     Update existing resources in destination")
	fmt.Println("  --dry-run                    Preview changes without applying them")
	fmt.Println("  --verbose                    Show detailed output")
	fmt.Println()
	fmt.Println("Peer Migration Options:")
	fmt.Println("  --source-url <url>           Source management URL (default: NetBird Cloud)")
	fmt.Println("  --dest-url <url>             Destination management URL (default: NetBird Cloud)")
	fmt.Println("  --create-groups              Create missing groups in destination (default: true)")
	fmt.Println("  --key-expiry <duration>      Setup key expiration: 1h, 24h, 7d (default: 24h)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println()
	fmt.Println("  # Migrate all configuration (dry-run first):")
	fmt.Println("  netbird-manage migrate \\")
	fmt.Println("    --source-token \"nbp_source...\" \\")
	fmt.Println("    --dest-token \"nbp_dest...\" \\")
	fmt.Println("    --config --dry-run")
	fmt.Println()
	fmt.Println("  # Migrate configuration and apply:")
	fmt.Println("  netbird-manage migrate \\")
	fmt.Println("    --source-token \"nbp_source...\" \\")
	fmt.Println("    --dest-token \"nbp_dest...\" \\")
	fmt.Println("    --config --skip-existing")
	fmt.Println()
	fmt.Println("  # Migrate everything (config + all peers):")
	fmt.Println("  netbird-manage migrate \\")
	fmt.Println("    --source-token \"nbp_source...\" \\")
	fmt.Println("    --dest-token \"nbp_dest...\" \\")
	fmt.Println("    --all")
	fmt.Println()
	fmt.Println("  # Migrate only policies and groups:")
	fmt.Println("  netbird-manage migrate \\")
	fmt.Println("    --source-token \"nbp_source...\" \\")
	fmt.Println("    --dest-token \"nbp_dest...\" \\")
	fmt.Println("    --groups --policies --skip-existing")
	fmt.Println()
	fmt.Println("  # Migrate a single peer:")
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
	fmt.Println("    --config")
	fmt.Println()
	fmt.Println("Recommended Migration Order:")
	fmt.Println("  1. Run with --config --dry-run to preview changes")
	fmt.Println("  2. Run with --config to migrate configuration")
	fmt.Println("  3. Run with --peer or --group to generate peer migration commands")
	fmt.Println("  4. Execute migration commands on each peer device")
	fmt.Println("  5. Clean up old peers from source account")
}
