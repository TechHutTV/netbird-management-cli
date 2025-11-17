// import.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ImportContext holds the state for an import operation
type ImportContext struct {
	Client *Client

	// Flags
	Apply         bool
	Update        bool
	SkipExisting  bool
	Force         bool
	Verbose       bool
	GroupsOnly    bool
	PoliciesOnly  bool
	NetworksOnly  bool
	RoutesOnly    bool
	DNSOnly       bool
	PostureOnly   bool
	SetupKeysOnly bool

	// State mappings (name -> ID)
	GroupNameToID        map[string]string
	PeerNameToID         map[string]string
	PolicyNameToID       map[string]string
	NetworkNameToID      map[string]string
	PostureCheckNameToID map[string]string

	// Existing resources
	ExistingGroups       map[string]*GroupDetail
	ExistingPolicies     map[string]*Policy
	ExistingNetworks     map[string]*Network
	ExistingRoutes       []Route
	ExistingDNS          map[string]*DNSNameserverGroup
	ExistingPosture      map[string]*PostureCheck
	ExistingSetupKeys    map[string]*SetupKey

	// Import results
	Created []string
	Updated []string
	Skipped []string
	Failed  []ImportError
}

// ImportError tracks errors during import
type ImportError struct {
	Resource string
	Error    error
}

// printImportUsage provides specific help for the 'import' command
func printImportUsage() {
	fmt.Println("Usage: netbird-manage import <file-or-directory> [flags]")
	fmt.Println("\nImport NetBird configuration from YAML files.")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <file-or-directory>           YAML file or directory to import")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --apply                       Actually apply changes (default is dry-run)")
	fmt.Println("  --update                      Update existing resources")
	fmt.Println("  --skip-existing               Skip resources that already exist")
	fmt.Println("  --force                       Create or update all resources (upsert)")
	fmt.Println("  --verbose                     Show detailed output")
	fmt.Println()
	fmt.Println("Selective Import:")
	fmt.Println("  --groups-only                 Import only groups")
	fmt.Println("  --policies-only               Import only policies")
	fmt.Println("  --networks-only               Import only networks")
	fmt.Println("  --routes-only                 Import only routes")
	fmt.Println("  --dns-only                    Import only DNS nameserver groups")
	fmt.Println("  --posture-only                Import only posture checks")
	fmt.Println("  --setup-keys-only             Import only setup keys")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  netbird-manage import config.yml                          # Dry-run (preview)")
	fmt.Println("  netbird-manage import config.yml --apply                  # Apply changes")
	fmt.Println("  netbird-manage import config.yml --apply --update         # Update existing")
	fmt.Println("  netbird-manage import config.yml --apply --skip-existing  # Skip conflicts")
	fmt.Println("  netbird-manage import config.yml --apply --force          # Create or update")
	fmt.Println("  netbird-manage import ./export-dir/ --apply               # Import from directory")
	fmt.Println("  netbird-manage import config.yml --apply --groups-only    # Only import groups")
	fmt.Println()
	fmt.Println("Conflict Resolution:")
	fmt.Println("  Default:        Fail on existing resources (safe)")
	fmt.Println("  --update:       Update existing resources with YAML values")
	fmt.Println("  --skip-existing: Silently skip resources that exist")
	fmt.Println("  --force:        Create new or update existing (upsert)")
}

// handleImportCommand handles the import command
func handleImportCommand(client *Client, args []string) error {
	importCmd := flag.NewFlagSet("import", flag.ContinueOnError)
	importCmd.SetOutput(os.Stderr)

	applyFlag := importCmd.Bool("apply", false, "Actually apply changes (default is dry-run)")
	updateFlag := importCmd.Bool("update", false, "Update existing resources")
	skipFlag := importCmd.Bool("skip-existing", false, "Skip resources that already exist")
	forceFlag := importCmd.Bool("force", false, "Create or update all resources (upsert)")
	verboseFlag := importCmd.Bool("verbose", false, "Show detailed output")

	groupsOnlyFlag := importCmd.Bool("groups-only", false, "Import only groups")
	policiesOnlyFlag := importCmd.Bool("policies-only", false, "Import only policies")
	networksOnlyFlag := importCmd.Bool("networks-only", false, "Import only networks")
	routesOnlyFlag := importCmd.Bool("routes-only", false, "Import only routes")
	dnsOnlyFlag := importCmd.Bool("dns-only", false, "Import only DNS nameserver groups")
	postureOnlyFlag := importCmd.Bool("posture-only", false, "Import only posture checks")
	setupKeysOnlyFlag := importCmd.Bool("setup-keys-only", false, "Import only setup keys")

	if err := importCmd.Parse(args[1:]); err != nil {
		return err
	}

	// Get file/directory argument
	remainingArgs := importCmd.Args()
	if len(remainingArgs) == 0 {
		printImportUsage()
		return fmt.Errorf("missing required argument: file or directory")
	}

	path := remainingArgs[0]

	// Create import context
	ctx := &ImportContext{
		Client:               client,
		Apply:                *applyFlag,
		Update:               *updateFlag,
		SkipExisting:         *skipFlag,
		Force:                *forceFlag,
		Verbose:              *verboseFlag,
		GroupsOnly:           *groupsOnlyFlag,
		PoliciesOnly:         *policiesOnlyFlag,
		NetworksOnly:         *networksOnlyFlag,
		RoutesOnly:           *routesOnlyFlag,
		DNSOnly:              *dnsOnlyFlag,
		PostureOnly:          *postureOnlyFlag,
		SetupKeysOnly:        *setupKeysOnlyFlag,
		GroupNameToID:        make(map[string]string),
		PeerNameToID:         make(map[string]string),
		PolicyNameToID:       make(map[string]string),
		NetworkNameToID:      make(map[string]string),
		PostureCheckNameToID: make(map[string]string),
		ExistingGroups:       make(map[string]*GroupDetail),
		ExistingPolicies:     make(map[string]*Policy),
		ExistingNetworks:     make(map[string]*Network),
		ExistingDNS:          make(map[string]*DNSNameserverGroup),
		ExistingPosture:      make(map[string]*PostureCheck),
		ExistingSetupKeys:    make(map[string]*SetupKey),
	}

	// Validate conflict resolution flags
	flagCount := 0
	if ctx.Update {
		flagCount++
	}
	if ctx.SkipExisting {
		flagCount++
	}
	if ctx.Force {
		flagCount++
	}
	if flagCount > 1 {
		return fmt.Errorf("cannot use --update, --skip-existing, and --force together")
	}

	// Show mode
	if !ctx.Apply {
		fmt.Println("🔍 Import Preview (Dry Run)")
		fmt.Println("================================================")
		fmt.Println()
	} else {
		fmt.Println("▶ Importing NetBird configuration...")
		fmt.Println()
	}

	// Step 1: Parse YAML file(s)
	yamlData, err := loadYAMLData(path)
	if err != nil {
		return fmt.Errorf("failed to load YAML: %v", err)
	}

	// Step 2: Fetch current state from API
	if err := ctx.fetchCurrentState(); err != nil {
		return fmt.Errorf("failed to fetch current state: %v", err)
	}

	// Step 3: Import resources in dependency order
	if err := ctx.importResources(yamlData); err != nil {
		return err
	}

	// Step 4: Print summary
	ctx.printSummary()

	return nil
}

// loadYAMLData loads YAML from a file or directory
func loadYAMLData(path string) (map[string]interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %v", err)
	}

	if info.IsDir() {
		return loadYAMLFromDirectory(path)
	}
	return loadYAMLFromFile(path)
}

// loadYAMLFromFile loads YAML from a single file
func loadYAMLFromFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid YAML syntax: %v", err)
	}

	return result, nil
}

// loadYAMLFromDirectory loads YAML from split files in a directory
func loadYAMLFromDirectory(dirPath string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Load config.yml to get import order
	configPath := filepath.Join(dirPath, "config.yml")
	configData, err := loadYAMLFromFile(configPath)
	if err != nil {
		// If no config.yml, use default order
		return loadDefaultDirectoryOrder(dirPath)
	}

	// Get import order from config
	importOrder, ok := configData["import_order"].([]interface{})
	if !ok {
		return loadDefaultDirectoryOrder(dirPath)
	}

	// Load files in specified order
	for _, fileInterface := range importOrder {
		filename, ok := fileInterface.(string)
		if !ok {
			continue
		}

		filePath := filepath.Join(dirPath, filename)
		fileData, err := loadYAMLFromFile(filePath)
		if err != nil {
			// Skip missing files
			continue
		}

		// Merge file data into result
		for key, value := range fileData {
			result[key] = value
		}
	}

	return result, nil
}

// loadDefaultDirectoryOrder loads files in default dependency order
func loadDefaultDirectoryOrder(dirPath string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	defaultOrder := []string{
		"groups.yml",
		"posture-checks.yml",
		"policies.yml",
		"routes.yml",
		"dns.yml",
		"networks.yml",
		"setup-keys.yml",
	}

	for _, filename := range defaultOrder {
		filePath := filepath.Join(dirPath, filename)
		if _, err := os.Stat(filePath); err != nil {
			// Skip missing files
			continue
		}

		fileData, err := loadYAMLFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %v", filename, err)
		}

		// Merge file data into result
		for key, value := range fileData {
			result[key] = value
		}
	}

	return result, nil
}

// fetchCurrentState fetches all existing resources from API
func (ctx *ImportContext) fetchCurrentState() error {
	if ctx.Verbose {
		fmt.Println("📡 Fetching current state from API...")
	}

	// Fetch groups
	resp, err := ctx.Client.makeRequest("GET", "/groups", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch groups: %v", err)
	}
	defer resp.Body.Close()

	var groups []GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return fmt.Errorf("failed to decode groups: %v", err)
	}

	for _, group := range groups {
		ctx.GroupNameToID[group.Name] = group.ID
		groupCopy := group
		ctx.ExistingGroups[group.Name] = &groupCopy

		// Build peer name to ID mapping
		for _, peer := range group.Peers {
			ctx.PeerNameToID[peer.Name] = peer.ID
		}
	}

	// Fetch policies
	resp, err = ctx.Client.makeRequest("GET", "/policies", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch policies: %v", err)
	}
	defer resp.Body.Close()

	var policies []Policy
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return fmt.Errorf("failed to decode policies: %v", err)
	}

	for _, policy := range policies {
		ctx.PolicyNameToID[policy.Name] = policy.ID
		policyCopy := policy
		ctx.ExistingPolicies[policy.Name] = &policyCopy
	}

	// Fetch networks
	resp, err = ctx.Client.makeRequest("GET", "/networks", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch networks: %v", err)
	}
	defer resp.Body.Close()

	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return fmt.Errorf("failed to decode networks: %v", err)
	}

	for _, network := range networks {
		ctx.NetworkNameToID[network.Name] = network.ID
		networkCopy := network
		ctx.ExistingNetworks[network.Name] = &networkCopy
	}

	// Fetch routes
	resp, err = ctx.Client.makeRequest("GET", "/routes", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch routes: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&ctx.ExistingRoutes); err != nil {
		return fmt.Errorf("failed to decode routes: %v", err)
	}

	// Fetch DNS
	resp, err = ctx.Client.makeRequest("GET", "/dns/nameservers", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch DNS: %v", err)
	}
	defer resp.Body.Close()

	var dnsGroups []DNSNameserverGroup
	if err := json.NewDecoder(resp.Body).Decode(&dnsGroups); err != nil {
		return fmt.Errorf("failed to decode DNS: %v", err)
	}

	for _, dns := range dnsGroups {
		dnsCopy := dns
		ctx.ExistingDNS[dns.Name] = &dnsCopy
	}

	// Fetch posture checks
	resp, err = ctx.Client.makeRequest("GET", "/posture-checks", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch posture checks: %v", err)
	}
	defer resp.Body.Close()

	var checks []PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
		return fmt.Errorf("failed to decode posture checks: %v", err)
	}

	for _, check := range checks {
		ctx.PostureCheckNameToID[check.Name] = check.ID
		checkCopy := check
		ctx.ExistingPosture[check.Name] = &checkCopy
	}

	// Fetch setup keys
	resp, err = ctx.Client.makeRequest("GET", "/setup-keys", nil)
	if err != nil {
		return fmt.Errorf("failed to fetch setup keys: %v", err)
	}
	defer resp.Body.Close()

	var keys []SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return fmt.Errorf("failed to decode setup keys: %v", err)
	}

	for _, key := range keys {
		keyCopy := key
		ctx.ExistingSetupKeys[key.Name] = &keyCopy
	}

	if ctx.Verbose {
		fmt.Printf("  Found: %d groups, %d policies, %d networks, %d routes, %d DNS, %d posture checks, %d setup keys\n",
			len(ctx.ExistingGroups), len(ctx.ExistingPolicies), len(ctx.ExistingNetworks),
			len(ctx.ExistingRoutes), len(ctx.ExistingDNS), len(ctx.ExistingPosture), len(ctx.ExistingSetupKeys))
		fmt.Println()
	}

	return nil
}

// importResources imports all resources in dependency order
func (ctx *ImportContext) importResources(data map[string]interface{}) error {
	// Import in dependency order
	if !ctx.skipResourceType("groups") {
		if err := ctx.importGroups(data); err != nil {
			return err
		}
	}

	if !ctx.skipResourceType("posture") {
		if err := ctx.importPostureChecks(data); err != nil {
			return err
		}
	}

	if !ctx.skipResourceType("policies") {
		if err := ctx.importPolicies(data); err != nil {
			return err
		}
	}

	if !ctx.skipResourceType("routes") {
		if err := ctx.importRoutes(data); err != nil {
			return err
		}
	}

	if !ctx.skipResourceType("dns") {
		if err := ctx.importDNS(data); err != nil {
			return err
		}
	}

	if !ctx.skipResourceType("networks") {
		if err := ctx.importNetworks(data); err != nil {
			return err
		}
	}

	if !ctx.skipResourceType("setup-keys") {
		if err := ctx.importSetupKeys(data); err != nil {
			return err
		}
	}

	return nil
}

// skipResourceType checks if a resource type should be skipped
func (ctx *ImportContext) skipResourceType(resourceType string) bool {
	// If no selective flags, import all
	if !ctx.GroupsOnly && !ctx.PoliciesOnly && !ctx.NetworksOnly &&
		!ctx.RoutesOnly && !ctx.DNSOnly && !ctx.PostureOnly && !ctx.SetupKeysOnly {
		return false
	}

	// Check selective flags
	switch resourceType {
	case "groups":
		return !ctx.GroupsOnly
	case "policies":
		return !ctx.PoliciesOnly
	case "networks":
		return !ctx.NetworksOnly
	case "routes":
		return !ctx.RoutesOnly
	case "dns":
		return !ctx.DNSOnly
	case "posture":
		return !ctx.PostureOnly
	case "setup-keys":
		return !ctx.SetupKeysOnly
	default:
		return true
	}
}

// importGroups imports group resources
func (ctx *ImportContext) importGroups(data map[string]interface{}) error {
	groupsData, ok := data["groups"].(map[string]interface{})
	if !ok {
		return nil // No groups to import
	}

	fmt.Println("📦 Groups:")

	for groupName, groupDataInterface := range groupsData {
		groupData, ok := groupDataInterface.(map[string]interface{})
		if !ok {
			ctx.addError("Group "+groupName, fmt.Errorf("invalid group data"))
			continue
		}

		if err := ctx.importGroup(groupName, groupData); err != nil {
			ctx.addError("Group "+groupName, err)
		}
	}

	fmt.Println()
	return nil
}

// importGroup imports a single group
func (ctx *ImportContext) importGroup(name string, data map[string]interface{}) error {
	// Check if group exists
	existing, exists := ctx.ExistingGroups[name]

	// Handle conflict
	if exists {
		if ctx.SkipExisting {
			fmt.Printf("  ⚠ SKIP     %s (already exists)\n", name)
			ctx.Skipped = append(ctx.Skipped, "Group "+name)
			return nil
		}

		if !ctx.Update && !ctx.Force {
			fmt.Printf("  ✗ CONFLICT %s (already exists, use --update or --skip-existing)\n", name)
			return fmt.Errorf("group already exists")
		}

		// Update existing group
		if ctx.Apply {
			if err := ctx.updateGroup(name, existing.ID, data); err != nil {
				fmt.Printf("  ✗ FAILED   %s (%v)\n", name, err)
				return err
			}
			fmt.Printf("  ✓ UPDATED  %s\n", name)
			ctx.Updated = append(ctx.Updated, "Group "+name)
		} else {
			fmt.Printf("  ✓ UPDATE   %s (would update)\n", name)
		}
		return nil
	}

	// Create new group
	if ctx.Apply {
		if err := ctx.createGroup(name, data); err != nil {
			fmt.Printf("  ✗ FAILED   %s (%v)\n", name, err)
			return err
		}
		fmt.Printf("  ✓ CREATED  %s\n", name)
		ctx.Created = append(ctx.Created, "Group "+name)
	} else {
		fmt.Printf("  ✓ CREATE   %s (would create)\n", name)
	}

	return nil
}

// createGroup creates a new group via API
func (ctx *ImportContext) createGroup(name string, data map[string]interface{}) error {
	// Resolve peer names to IDs
	peerIDs, err := ctx.resolvePeerNames(data["peers"])
	if err != nil {
		return fmt.Errorf("failed to resolve peers: %v", err)
	}

	reqBody := map[string]interface{}{
		"name":  name,
		"peers": peerIDs,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.Client.makeRequest("POST", "/groups", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	// Add to context for future references
	var createdGroup GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&createdGroup); err == nil {
		ctx.GroupNameToID[name] = createdGroup.ID
		ctx.ExistingGroups[name] = &createdGroup
	}

	return nil
}

// updateGroup updates an existing group
func (ctx *ImportContext) updateGroup(name, groupID string, data map[string]interface{}) error {
	// Resolve peer names to IDs
	peerIDs, err := ctx.resolvePeerNames(data["peers"])
	if err != nil {
		return fmt.Errorf("failed to resolve peers: %v", err)
	}

	// Get existing group to preserve resources
	existing := ctx.ExistingGroups[name]
	resources := []GroupResourcePutRequest{}
	for _, res := range existing.Resources {
		resources = append(resources, GroupResourcePutRequest{
			ID:   res.ID,
			Type: res.Type,
		})
	}

	reqBody := GroupPutRequest{
		Name:      name,
		Peers:     peerIDs,
		Resources: resources,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.Client.makeRequest("PUT", "/groups/"+groupID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// resolvePeerNames resolves peer names to IDs
func (ctx *ImportContext) resolvePeerNames(peersInterface interface{}) ([]string, error) {
	if peersInterface == nil {
		return []string{}, nil
	}

	peersList, ok := peersInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid peers format")
	}

	peerIDs := []string{}
	for _, peerInterface := range peersList {
		peerName, ok := peerInterface.(string)
		if !ok {
			continue
		}

		peerID, exists := ctx.PeerNameToID[peerName]
		if !exists {
			return nil, fmt.Errorf("peer '%s' not found (peers must be registered first)", peerName)
		}

		peerIDs = append(peerIDs, peerID)
	}

	return peerIDs, nil
}

// importPolicies imports policy resources
func (ctx *ImportContext) importPolicies(data map[string]interface{}) error {
	policiesData, ok := data["policies"].(map[string]interface{})
	if !ok {
		return nil // No policies to import
	}

	fmt.Println("🔐 Policies:")

	for policyName, policyDataInterface := range policiesData {
		policyData, ok := policyDataInterface.(map[string]interface{})
		if !ok {
			ctx.addError("Policy "+policyName, fmt.Errorf("invalid policy data"))
			continue
		}

		if err := ctx.importPolicy(policyName, policyData); err != nil {
			ctx.addError("Policy "+policyName, err)
		}
	}

	fmt.Println()
	return nil
}

// importPolicy imports a single policy
func (ctx *ImportContext) importPolicy(name string, data map[string]interface{}) error {
	// Check if policy exists
	_, exists := ctx.ExistingPolicies[name]

	// Handle conflict
	if exists {
		if ctx.SkipExisting {
			fmt.Printf("  ⚠ SKIP     %s (already exists)\n", name)
			ctx.Skipped = append(ctx.Skipped, "Policy "+name)
			return nil
		}

		if !ctx.Update && !ctx.Force {
			fmt.Printf("  ✗ CONFLICT %s (already exists, use --update or --skip-existing)\n", name)
			return fmt.Errorf("policy already exists")
		}

		// Update existing policy
		if ctx.Apply {
			if err := ctx.updatePolicy(name, data); err != nil {
				fmt.Printf("  ✗ FAILED   %s (%v)\n", name, err)
				return err
			}
			fmt.Printf("  ✓ UPDATED  %s\n", name)
			ctx.Updated = append(ctx.Updated, "Policy "+name)
		} else {
			fmt.Printf("  ✓ UPDATE   %s (would update)\n", name)
		}
		return nil
	}

	// Create new policy
	if ctx.Apply {
		if err := ctx.createPolicy(name, data); err != nil {
			fmt.Printf("  ✗ FAILED   %s (%v)\n", name, err)
			return err
		}
		fmt.Printf("  ✓ CREATED  %s\n", name)
		ctx.Created = append(ctx.Created, "Policy "+name)
	} else {
		fmt.Printf("  ✓ CREATE   %s (would create)\n", name)
	}

	return nil
}

// createPolicy creates a new policy
func (ctx *ImportContext) createPolicy(name string, data map[string]interface{}) error {
	description, _ := data["description"].(string)
	enabled, _ := data["enabled"].(bool)

	// Convert rules
	rules, err := ctx.convertPolicyRules(data["rules"])
	if err != nil {
		return fmt.Errorf("failed to convert rules: %v", err)
	}

	reqBody := PolicyCreateRequest{
		Name:        name,
		Description: description,
		Enabled:     enabled,
		Rules:       rules,
	}

	// Add source posture checks if present
	if postureChecks, ok := data["source_posture_checks"].([]interface{}); ok {
		for _, pc := range postureChecks {
			if pcStr, ok := pc.(string); ok {
				reqBody.SourcePostureChecks = append(reqBody.SourcePostureChecks, pcStr)
			}
		}
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.Client.makeRequest("POST", "/policies", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	// Add to context
	var createdPolicy Policy
	if err := json.NewDecoder(resp.Body).Decode(&createdPolicy); err == nil {
		ctx.PolicyNameToID[name] = createdPolicy.ID
		ctx.ExistingPolicies[name] = &createdPolicy
	}

	return nil
}

// updatePolicy updates an existing policy
func (ctx *ImportContext) updatePolicy(name string, data map[string]interface{}) error {
	policyID := ctx.PolicyNameToID[name]
	description, _ := data["description"].(string)
	enabled, _ := data["enabled"].(bool)

	// Convert rules
	rules, err := ctx.convertPolicyRules(data["rules"])
	if err != nil {
		return fmt.Errorf("failed to convert rules: %v", err)
	}

	reqBody := PolicyUpdateRequest{
		Name:        name,
		Description: description,
		Enabled:     enabled,
		Rules:       rules,
	}

	// Add source posture checks if present
	if postureChecks, ok := data["source_posture_checks"].([]interface{}); ok {
		for _, pc := range postureChecks {
			if pcStr, ok := pc.(string); ok {
				reqBody.SourcePostureChecks = append(reqBody.SourcePostureChecks, pcStr)
			}
		}
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.Client.makeRequest("PUT", "/policies/"+policyID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// convertPolicyRules converts YAML rules to API format
func (ctx *ImportContext) convertPolicyRules(rulesInterface interface{}) ([]PolicyRuleForWrite, error) {
	if rulesInterface == nil {
		return []PolicyRuleForWrite{}, nil
	}

	rulesMap, ok := rulesInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid rules format")
	}

	rules := []PolicyRuleForWrite{}
	for ruleName, ruleDataInterface := range rulesMap {
		ruleData, ok := ruleDataInterface.(map[string]interface{})
		if !ok {
			continue
		}

		rule := PolicyRuleForWrite{
			Name:          ruleName,
			Description:   getString(ruleData, "description"),
			Enabled:       getBool(ruleData, "enabled"),
			Action:        getString(ruleData, "action"),
			Bidirectional: getBool(ruleData, "bidirectional"),
			Protocol:      getString(ruleData, "protocol"),
		}

		// Convert ports
		if ports, ok := ruleData["ports"].([]interface{}); ok {
			for _, port := range ports {
				if portStr, ok := port.(string); ok {
					rule.Ports = append(rule.Ports, portStr)
				}
			}
		}

		// Convert port ranges
		if portRanges, ok := ruleData["port_ranges"].([]interface{}); ok {
			for _, pr := range portRanges {
				if prMap, ok := pr.(map[string]interface{}); ok {
					start, _ := prMap["start"].(int)
					end, _ := prMap["end"].(int)
					rule.PortRanges = append(rule.PortRanges, PortRange{
						Start: start,
						End:   end,
					})
				}
			}
		}

		// Resolve source group names to IDs
		if sources, ok := ruleData["sources"].([]interface{}); ok {
			for _, src := range sources {
				if srcName, ok := src.(string); ok {
					srcID, exists := ctx.GroupNameToID[srcName]
					if !exists {
						return nil, fmt.Errorf("source group '%s' not found", srcName)
					}
					rule.Sources = append(rule.Sources, srcID)
				}
			}
		}

		// Resolve destination group names to IDs
		if dests, ok := ruleData["destinations"].([]interface{}); ok {
			for _, dest := range dests {
				if destName, ok := dest.(string); ok {
					destID, exists := ctx.GroupNameToID[destName]
					if !exists {
						return nil, fmt.Errorf("destination group '%s' not found", destName)
					}
					rule.Destinations = append(rule.Destinations, destID)
				}
			}
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// Helper functions to safely get values from maps
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	return 0
}

// Stub implementations for other resource types (simplified for now)
func (ctx *ImportContext) importPostureChecks(data map[string]interface{}) error {
	// TODO: Implement posture checks import
	return nil
}

func (ctx *ImportContext) importRoutes(data map[string]interface{}) error {
	// TODO: Implement routes import
	return nil
}

func (ctx *ImportContext) importDNS(data map[string]interface{}) error {
	// TODO: Implement DNS import
	return nil
}

func (ctx *ImportContext) importNetworks(data map[string]interface{}) error {
	networksData, ok := data["networks"].(map[string]interface{})
	if !ok {
		return nil // No networks to import
	}

	fmt.Println("🌐 Networks:")

	for networkName, networkDataInterface := range networksData {
		networkData, ok := networkDataInterface.(map[string]interface{})
		if !ok {
			ctx.addError("Network "+networkName, fmt.Errorf("invalid network data"))
			continue
		}

		if err := ctx.importNetwork(networkName, networkData); err != nil {
			ctx.addError("Network "+networkName, err)
		}
	}

	fmt.Println()
	return nil
}

// importNetwork imports a single network with its resources and routers
func (ctx *ImportContext) importNetwork(name string, data map[string]interface{}) error {
	// Check if network exists
	existing, exists := ctx.ExistingNetworks[name]

	// Handle conflict
	if exists {
		if ctx.SkipExisting {
			fmt.Printf("  ⚠ SKIP     %s (already exists)\n", name)
			ctx.Skipped = append(ctx.Skipped, "Network "+name)
			return nil
		}

		if !ctx.Update && !ctx.Force {
			fmt.Printf("  ✗ CONFLICT %s (already exists, use --update or --skip-existing)\n", name)
			return fmt.Errorf("network already exists")
		}

		// Update existing network
		if ctx.Apply {
			if err := ctx.updateNetwork(name, existing.ID, data); err != nil {
				fmt.Printf("  ✗ FAILED   %s (%v)\n", name, err)
				return err
			}
			fmt.Printf("  ✓ UPDATED  %s\n", name)
			ctx.Updated = append(ctx.Updated, "Network "+name)
		} else {
			fmt.Printf("  ✓ UPDATE   %s (would update)\n", name)
		}
		return nil
	}

	// Create new network
	if ctx.Apply {
		if err := ctx.createNetwork(name, data); err != nil {
			fmt.Printf("  ✗ FAILED   %s (%v)\n", name, err)
			return err
		}
		fmt.Printf("  ✓ CREATED  %s\n", name)
		ctx.Created = append(ctx.Created, "Network "+name)
	} else {
		fmt.Printf("  ✓ CREATE   %s (would create)\n", name)
	}

	return nil
}

// createNetwork creates a new network with resources and routers
func (ctx *ImportContext) createNetwork(name string, data map[string]interface{}) error {
	description, _ := data["description"].(string)

	// Create the network first
	reqBody := NetworkCreateRequest{
		Name:        name,
		Description: description,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.Client.makeRequest("POST", "/networks", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	var createdNetwork Network
	if err := json.NewDecoder(resp.Body).Decode(&createdNetwork); err != nil {
		return fmt.Errorf("failed to decode network: %v", err)
	}

	// Add to context
	ctx.NetworkNameToID[name] = createdNetwork.ID
	ctx.ExistingNetworks[name] = &createdNetwork

	// Now add resources and routers
	if err := ctx.addNetworkResources(createdNetwork.ID, data); err != nil {
		return fmt.Errorf("failed to add resources: %v", err)
	}

	if err := ctx.addNetworkRouters(createdNetwork.ID, data); err != nil {
		return fmt.Errorf("failed to add routers: %v", err)
	}

	return nil
}

// updateNetwork updates an existing network
func (ctx *ImportContext) updateNetwork(name, networkID string, data map[string]interface{}) error {
	description, _ := data["description"].(string)

	// Update network metadata
	reqBody := NetworkUpdateRequest{
		Name:        name,
		Description: description,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := ctx.Client.makeRequest("PUT", "/networks/"+networkID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	// Update resources and routers
	if err := ctx.addNetworkResources(networkID, data); err != nil {
		return fmt.Errorf("failed to add resources: %v", err)
	}

	if err := ctx.addNetworkRouters(networkID, data); err != nil {
		return fmt.Errorf("failed to add routers: %v", err)
	}

	return nil
}

// addNetworkResources adds resources to a network
func (ctx *ImportContext) addNetworkResources(networkID string, data map[string]interface{}) error {
	resourcesData, ok := data["resources"].(map[string]interface{})
	if !ok || len(resourcesData) == 0 {
		return nil // No resources to add
	}

	for resourceName, resourceDataInterface := range resourcesData {
		resourceData, ok := resourceDataInterface.(map[string]interface{})
		if !ok {
			continue
		}

		address, _ := resourceData["address"].(string)
		description, _ := resourceData["description"].(string)
		enabled := getBool(resourceData, "enabled")
		resourceType, _ := resourceData["type"].(string)

		// Resolve group names to IDs
		var groupIDs []string
		if groupsInterface, ok := resourceData["groups"].([]interface{}); ok {
			for _, groupInterface := range groupsInterface {
				if groupName, ok := groupInterface.(string); ok {
					groupID, exists := ctx.GroupNameToID[groupName]
					if !exists {
						return fmt.Errorf("group '%s' not found for resource '%s'", groupName, resourceName)
					}
					groupIDs = append(groupIDs, groupID)
				}
			}
		}

		if len(groupIDs) == 0 {
			return fmt.Errorf("resource '%s' must have at least one group", resourceName)
		}

		if address == "" {
			return fmt.Errorf("resource '%s' must have an address", resourceName)
		}

		// Set type to subnet if not specified
		if resourceType == "" {
			resourceType = "subnet"
		}

		// Create the resource
		resourceReq := NetworkResourceRequest{
			Name:        resourceName,
			Description: description,
			Address:     address,
			Enabled:     enabled,
			Groups:      groupIDs,
		}

		bodyBytes, _ := json.Marshal(resourceReq)
		resp, err := ctx.Client.makeRequest("POST", "/networks/"+networkID+"/resources", bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create resource '%s': %v", resourceName, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("failed to create resource '%s': %s", resourceName, resp.Status)
		}
	}

	return nil
}

// addNetworkRouters adds routers to a network
func (ctx *ImportContext) addNetworkRouters(networkID string, data map[string]interface{}) error {
	routersData, ok := data["routers"].(map[string]interface{})
	if !ok || len(routersData) == 0 {
		return nil // No routers to add
	}

	for routerName, routerDataInterface := range routersData {
		routerData, ok := routerDataInterface.(map[string]interface{})
		if !ok {
			continue
		}

		peer, _ := routerData["peer"].(string)
		metric := getInt(routerData, "metric")
		if metric == 0 {
			metric = 100 // Default metric
		}
		masquerade := getBool(routerData, "masquerade")
		enabled := getBool(routerData, "enabled")

		// Resolve peer groups if present
		var peerGroups []string
		if peerGroupsInterface, ok := routerData["peer_groups"].([]interface{}); ok {
			for _, pgInterface := range peerGroupsInterface {
				if pgName, ok := pgInterface.(string); ok {
					pgID, exists := ctx.GroupNameToID[pgName]
					if !exists {
						return fmt.Errorf("peer group '%s' not found for router '%s'", pgName, routerName)
					}
					peerGroups = append(peerGroups, pgID)
				}
			}
		}

		// Must have either peer or peer_groups
		if peer == "" && len(peerGroups) == 0 {
			return fmt.Errorf("router '%s' must have either a peer or peer_groups", routerName)
		}

		// Create the router
		routerReq := NetworkRouterRequest{
			Peer:       peer,
			PeerGroups: peerGroups,
			Metric:     metric,
			Masquerade: masquerade,
			Enabled:    enabled,
		}

		bodyBytes, _ := json.Marshal(routerReq)
		resp, err := ctx.Client.makeRequest("POST", "/networks/"+networkID+"/routers", bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("failed to create router '%s': %v", routerName, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("failed to create router '%s': %s", routerName, resp.Status)
		}
	}

	return nil
}

func (ctx *ImportContext) importSetupKeys(data map[string]interface{}) error {
	// TODO: Implement setup keys import
	return nil
}

// addError adds an error to the import context
func (ctx *ImportContext) addError(resource string, err error) {
	ctx.Failed = append(ctx.Failed, ImportError{
		Resource: resource,
		Error:    err,
	})
}

// printSummary prints the import summary
func (ctx *ImportContext) printSummary() {
	fmt.Println("================================================")
	fmt.Println("📊 Import Summary")
	fmt.Println("================================================")
	fmt.Println()

	if len(ctx.Created) > 0 {
		fmt.Printf("✓ Created:  %d resources\n", len(ctx.Created))
		if ctx.Verbose {
			for _, res := range ctx.Created {
				fmt.Printf("    - %s\n", res)
			}
		}
	}

	if len(ctx.Updated) > 0 {
		fmt.Printf("✓ Updated:  %d resources\n", len(ctx.Updated))
		if ctx.Verbose {
			for _, res := range ctx.Updated {
				fmt.Printf("    - %s\n", res)
			}
		}
	}

	if len(ctx.Skipped) > 0 {
		fmt.Printf("⚠ Skipped:  %d resources\n", len(ctx.Skipped))
		if ctx.Verbose {
			for _, res := range ctx.Skipped {
				fmt.Printf("    - %s\n", res)
			}
		}
	}

	if len(ctx.Failed) > 0 {
		fmt.Printf("✗ Failed:   %d resources\n", len(ctx.Failed))
		fmt.Println()
		fmt.Println("Errors:")
		for i, fail := range ctx.Failed {
			fmt.Printf("  %d. %s: %v\n", i+1, fail.Resource, fail.Error)
		}
	}

	fmt.Println()

	if !ctx.Apply {
		fmt.Println("▶ This was a dry run. Use --apply to execute these changes.")
	} else {
		totalChanges := len(ctx.Created) + len(ctx.Updated)
		if totalChanges > 0 {
			fmt.Printf("✓ Successfully applied %d changes!\n", totalChanges)
		}
		if len(ctx.Failed) > 0 {
			fmt.Println("⚠ Some resources failed to import. Fix errors and re-run with --skip-existing")
		}
	}
}
