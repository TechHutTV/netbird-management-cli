// export.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// printExportUsage provides specific help for the 'export' command
func printExportUsage() {
	fmt.Println("Usage: netbird-manage export [flags] [directory]")
	fmt.Println("\nExport NetBird configuration to YAML files for GitOps workflows.")
	fmt.Println("\nFlags:")
	fmt.Println("  --full                        Export to a single YAML file (default)")
	fmt.Println("  --split                       Export to multiple YAML files in a directory")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  [directory]                   Target directory (optional, defaults to current directory)")
	fmt.Println()
	fmt.Println("Output Format:")
	fmt.Println("  Single file:  netbird-manage-export-YYMMDD.yml")
	fmt.Println("  Split files:  netbird-manage-export-YYMMDD/")
	fmt.Println("                  ├── config.yml")
	fmt.Println("                  ├── groups.yml")
	fmt.Println("                  ├── policies.yml")
	fmt.Println("                  ├── networks.yml")
	fmt.Println("                  ├── routes.yml")
	fmt.Println("                  ├── dns.yml")
	fmt.Println("                  ├── posture-checks.yml")
	fmt.Println("                  └── setup-keys.yml")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  netbird-manage export                    # Export to single file in current directory")
	fmt.Println("  netbird-manage export --full ./backups   # Export to single file in ./backups")
	fmt.Println("  netbird-manage export --split            # Export to split files in current directory")
	fmt.Println("  netbird-manage export --split ~/exports  # Export to split files in ~/exports")
}

// handleExportCommand handles the export command
func handleExportCommand(client *Client, args []string) error {
	exportCmd := flag.NewFlagSet("export", flag.ContinueOnError)
	exportCmd.SetOutput(os.Stderr)

	fullFlag := exportCmd.Bool("full", false, "Export to a single YAML file (default if neither flag specified)")
	splitFlag := exportCmd.Bool("split", false, "Export to multiple YAML files in a directory")

	if err := exportCmd.Parse(args[1:]); err != nil {
		return err
	}

	// Get optional directory argument
	directory := "."
	remainingArgs := exportCmd.Args()
	if len(remainingArgs) > 0 {
		directory = remainingArgs[0]
	}

	// Default to full export if neither flag specified
	useSplitMode := *splitFlag
	if !*fullFlag && !*splitFlag {
		useSplitMode = false // Default to single file
	}

	// Generate timestamp for filename/directory
	timestamp := time.Now().Format("060102") // YYMMDD format

	if useSplitMode {
		return exportSplitFiles(client, directory, timestamp)
	}
	return exportFullSingleFile(client, directory, timestamp)
}

// exportFullSingleFile exports all resources to a single YAML file
func exportFullSingleFile(client *Client, directory, timestamp string) error {
	fmt.Println("Exporting NetBird configuration to single YAML file...")

	// Fetch all resources
	data, err := fetchAllResources(client)
	if err != nil {
		return fmt.Errorf("failed to fetch resources: %v", err)
	}

	// Create output filename
	filename := fmt.Sprintf("netbird-manage-export-%s.yml", timestamp)
	filepath := filepath.Join(directory, filename)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("✓ Export completed: %s\n", filepath)
	return nil
}

// exportSplitFiles exports resources to multiple YAML files in a directory
func exportSplitFiles(client *Client, directory, timestamp string) error {
	fmt.Println("Exporting NetBird configuration to split YAML files...")

	// Create output directory
	dirName := fmt.Sprintf("netbird-manage-export-%s", timestamp)
	dirPath := filepath.Join(directory, dirName)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Fetch all resources
	allData, err := fetchAllResources(client)
	if err != nil {
		return fmt.Errorf("failed to fetch resources: %v", err)
	}

	// Extract metadata for config.yml
	metadata := allData["metadata"]
	configData := map[string]interface{}{
		"metadata": metadata,
		"import_order": []string{
			"groups.yml",
			"posture-checks.yml",
			"policies.yml",
			"routes.yml",
			"dns.yml",
			"networks.yml",
			"setup-keys.yml",
		},
	}

	// Write config.yml
	if err := writeYAMLFile(filepath.Join(dirPath, "config.yml"), configData); err != nil {
		return err
	}
	fmt.Printf("  ✓ config.yml\n")

	// Write individual resource files
	files := map[string]string{
		"groups.yml":         "groups",
		"posture-checks.yml": "posture_checks",
		"policies.yml":       "policies",
		"routes.yml":         "routes",
		"dns.yml":            "dns",
		"networks.yml":       "networks",
		"setup-keys.yml":     "setup_keys",
	}

	for filename, key := range files {
		fileData := map[string]interface{}{
			key: allData[key],
		}
		if err := writeYAMLFile(filepath.Join(dirPath, filename), fileData); err != nil {
			return err
		}
		fmt.Printf("  ✓ %s\n", filename)
	}

	fmt.Printf("✓ Export completed: %s/\n", dirPath)
	return nil
}

// writeYAMLFile writes data to a YAML file
func writeYAMLFile(filepath string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}

	if err := os.WriteFile(filepath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", filepath, err)
	}

	return nil
}

// fetchAllResources fetches all resources from the API and converts to YAML-friendly map structure
func fetchAllResources(client *Client) (map[string]interface{}, error) {
	// Create metadata
	metadata := map[string]interface{}{
		"version":        "1.0",
		"exported_at":    time.Now().Format(time.RFC3339),
		"management_url": client.ManagementURL,
	}

	// Fetch all resource types
	groups, err := fetchGroupsAsMap(client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch groups: %v", err)
	}

	policies, err := fetchPoliciesAsMap(client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policies: %v", err)
	}

	networks, err := fetchNetworksAsMap(client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch networks: %v", err)
	}

	routes, err := fetchRoutesAsMap(client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %v", err)
	}

	dns, err := fetchDNSAsMap(client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DNS: %v", err)
	}

	postureChecks, err := fetchPostureChecksAsMap(client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posture checks: %v", err)
	}

	setupKeys, err := fetchSetupKeysAsMap(client)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch setup keys: %v", err)
	}

	// Combine all resources
	return map[string]interface{}{
		"metadata":       metadata,
		"groups":         groups,
		"policies":       policies,
		"networks":       networks,
		"routes":         routes,
		"dns":            dns,
		"posture_checks": postureChecks,
		"setup_keys":     setupKeys,
	}, nil
}

// fetchGroupsAsMap fetches groups and converts to map[groupName]groupData
func fetchGroupsAsMap(client *Client) (map[string]interface{}, error) {
	resp, err := client.makeRequest("GET", "/groups", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("failed to decode groups: %v", err)
	}

	result := make(map[string]interface{})
	for _, group := range groups {
		// Extract peer names
		peerNames := make([]string, len(group.Peers))
		for i, peer := range group.Peers {
			peerNames[i] = peer.Name
		}

		result[group.Name] = map[string]interface{}{
			"description": fmt.Sprintf("Group with %d peers", group.PeersCount),
			"peers":       peerNames,
		}
	}

	return result, nil
}

// fetchPoliciesAsMap fetches policies and converts to map[policyName]policyData
func fetchPoliciesAsMap(client *Client) (map[string]interface{}, error) {
	resp, err := client.makeRequest("GET", "/policies", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var policies []Policy
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return nil, fmt.Errorf("failed to decode policies: %v", err)
	}

	result := make(map[string]interface{})
	for _, policy := range policies {
		// Convert rules array to map[ruleName]ruleData
		rules := make(map[string]interface{})
		for _, rule := range policy.Rules {
			// Convert source/destination PolicyGroups to string names
			sourceNames := make([]string, len(rule.Sources))
			for i, src := range rule.Sources {
				sourceNames[i] = src.Name
			}

			destNames := make([]string, len(rule.Destinations))
			for i, dest := range rule.Destinations {
				destNames[i] = dest.Name
			}

			ruleData := map[string]interface{}{
				"description":   rule.Description,
				"enabled":       rule.Enabled,
				"action":        rule.Action,
				"bidirectional": rule.Bidirectional,
				"protocol":      rule.Protocol,
			}

			if len(rule.Ports) > 0 {
				ruleData["ports"] = rule.Ports
			}

			if len(rule.PortRanges) > 0 {
				ruleData["port_ranges"] = rule.PortRanges
			}

			if len(sourceNames) > 0 {
				ruleData["sources"] = sourceNames
			}

			if len(destNames) > 0 {
				ruleData["destinations"] = destNames
			}

			if rule.SourceResource != nil {
				ruleData["source_resource"] = rule.SourceResource
			}

			if rule.DestinationResource != nil {
				ruleData["destination_resource"] = rule.DestinationResource
			}

			rules[rule.Name] = ruleData
		}

		policyData := map[string]interface{}{
			"description": policy.Description,
			"enabled":     policy.Enabled,
			"rules":       rules,
		}

		if len(policy.SourcePostureChecks) > 0 {
			policyData["source_posture_checks"] = policy.SourcePostureChecks
		}

		result[policy.Name] = policyData
	}

	return result, nil
}

// fetchNetworksAsMap fetches networks and converts to map[networkName]networkData
func fetchNetworksAsMap(client *Client) (map[string]interface{}, error) {
	resp, err := client.makeRequest("GET", "/networks", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, fmt.Errorf("failed to decode networks: %v", err)
	}

	result := make(map[string]interface{})
	for _, network := range networks {
		// Fetch detailed network information including resources and routers
		networkDetail, err := fetchNetworkDetail(client, network.ID)
		if err != nil {
			// If we can't get details, use basic info
			result[network.Name] = map[string]interface{}{
				"description": network.Description,
				"policies":    network.Policies,
			}
			continue
		}

		networkData := map[string]interface{}{
			"description": networkDetail.Description,
		}

		// Fetch and add resources
		resources, err := fetchNetworkResources(client, network.ID)
		if err == nil && len(resources) > 0 {
			resourcesMap := make(map[string]interface{})
			for _, resource := range resources {
				groupNames := make([]string, len(resource.Groups))
				for i, group := range resource.Groups {
					groupNames[i] = group.Name
				}

				resourcesMap[resource.Name] = map[string]interface{}{
					"type":        resource.Type,
					"address":     resource.Address,
					"enabled":     resource.Enabled,
					"description": resource.Description,
					"groups":      groupNames,
				}
			}
			networkData["resources"] = resourcesMap
		}

		// Fetch and add routers
		routers, err := fetchNetworkRouters(client, network.ID)
		if err == nil && len(routers) > 0 {
			routersMap := make(map[string]interface{})
			for i, router := range routers {
				routerName := fmt.Sprintf("router-%d", i+1)
				routerData := map[string]interface{}{
					"metric":     router.Metric,
					"masquerade": router.Masquerade,
					"enabled":    router.Enabled,
				}

				if router.Peer != "" {
					routerData["peer"] = router.Peer
				}

				if len(router.PeerGroups) > 0 {
					routerData["peer_groups"] = router.PeerGroups
				}

				routersMap[routerName] = routerData
			}
			networkData["routers"] = routersMap
		}

		if len(network.Policies) > 0 {
			networkData["policies"] = network.Policies
		}

		result[network.Name] = networkData
	}

	return result, nil
}

// fetchNetworkDetail fetches detailed network information
func fetchNetworkDetail(client *Client, networkID string) (*NetworkDetail, error) {
	resp, err := client.makeRequest("GET", "/networks/"+networkID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var detail NetworkDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

// fetchNetworkResources fetches resources for a network
func fetchNetworkResources(client *Client, networkID string) ([]NetworkResource, error) {
	resp, err := client.makeRequest("GET", "/networks/"+networkID+"/resources", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var resources []NetworkResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, err
	}

	return resources, nil
}

// fetchNetworkRouters fetches routers for a network
func fetchNetworkRouters(client *Client, networkID string) ([]NetworkRouter, error) {
	resp, err := client.makeRequest("GET", "/networks/"+networkID+"/routers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var routers []NetworkRouter
	if err := json.NewDecoder(resp.Body).Decode(&routers); err != nil {
		return nil, err
	}

	return routers, nil
}

// fetchRoutesAsMap fetches routes and converts to map[routeKey]routeData
func fetchRoutesAsMap(client *Client) (map[string]interface{}, error) {
	resp, err := client.makeRequest("GET", "/routes", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, fmt.Errorf("failed to decode routes: %v", err)
	}

	result := make(map[string]interface{})
	for i, route := range routes {
		// Use network CIDR as key, or generate unique key if duplicate
		routeKey := fmt.Sprintf("route-%d", i+1)
		if route.Description != "" {
			// Create a simple key from description (remove spaces, lowercase)
			routeKey = route.Description
		}

		routeData := map[string]interface{}{
			"description": route.Description,
			"network":     route.Network,
			"metric":      route.Metric,
			"masquerade":  route.Masquerade,
			"enabled":     route.Enabled,
			"groups":      route.Groups,
		}

		if route.Peer != "" {
			routeData["peer"] = route.Peer
		}

		if len(route.PeerGroups) > 0 {
			routeData["peer_groups"] = route.PeerGroups
		}

		result[routeKey] = routeData
	}

	return result, nil
}

// fetchDNSAsMap fetches DNS nameserver groups and converts to map[dnsGroupName]dnsData
func fetchDNSAsMap(client *Client) (map[string]interface{}, error) {
	resp, err := client.makeRequest("GET", "/dns/nameservers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dnsGroups []DNSNameserverGroup
	if err := json.NewDecoder(resp.Body).Decode(&dnsGroups); err != nil {
		return nil, fmt.Errorf("failed to decode DNS groups: %v", err)
	}

	result := make(map[string]interface{})
	for _, dns := range dnsGroups {
		result[dns.Name] = map[string]interface{}{
			"description":            dns.Description,
			"nameservers":            dns.Nameservers,
			"groups":                 dns.Groups,
			"domains":                dns.Domains,
			"search_domains_enabled": dns.SearchDomainsEnabled,
			"primary":                dns.Primary,
			"enabled":                dns.Enabled,
		}
	}

	return result, nil
}

// fetchPostureChecksAsMap fetches posture checks and converts to map[checkName]checkData
func fetchPostureChecksAsMap(client *Client) (map[string]interface{}, error) {
	resp, err := client.makeRequest("GET", "/posture-checks", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var checks []PostureCheck
	if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
		return nil, fmt.Errorf("failed to decode posture checks: %v", err)
	}

	result := make(map[string]interface{})
	for _, check := range checks {
		result[check.Name] = map[string]interface{}{
			"description": check.Description,
			"checks":      check.Checks,
		}
	}

	return result, nil
}

// fetchSetupKeysAsMap fetches setup keys and converts to map[keyName]keyData
func fetchSetupKeysAsMap(client *Client) (map[string]interface{}, error) {
	resp, err := client.makeRequest("GET", "/setup-keys", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keys []SetupKey
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode setup keys: %v", err)
	}

	result := make(map[string]interface{})
	for _, key := range keys {
		// Calculate expires_in from expires timestamp (approximate)
		expiresIn := 30 // Default 30 days if we can't calculate

		keyData := map[string]interface{}{
			"description": fmt.Sprintf("Type: %s, State: %s", key.Type, key.State),
			"type":        key.Type,
			"expires_in":  expiresIn,
			"auto_groups": key.AutoGroups,
			"usage_limit": key.UsageLimit,
			"ephemeral":   key.Ephemeral,
		}

		result[key.Name] = keyData
	}

	return result, nil
}
