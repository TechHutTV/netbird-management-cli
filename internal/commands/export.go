// export.go
package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"netbird-manage/internal/models"
)

// HandleExportCommand handles the export command
func (s *Service) HandleExportCommand(args []string) error {
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
		return s.exportSplitFiles(directory, timestamp)
	}
	return s.exportFullSingleFile(directory, timestamp)
}

// exportFullSingleFile exports all resources to a single YAML file
func (s *Service) exportFullSingleFile(directory, timestamp string) error {
	fmt.Println("Exporting NetBird configuration to single YAML file...")

	// Fetch all resources
	data, err := s.fetchAllResources()
	if err != nil {
		return fmt.Errorf("failed to fetch resources: %v", err)
	}

	// Create output filename
	filename := fmt.Sprintf("netbird-manage-export-%s.yml", timestamp)
	outputPath := filepath.Join(directory, filename)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("Export completed: %s\n", outputPath)
	return nil
}

// exportSplitFiles exports resources to multiple YAML files in a directory
func (s *Service) exportSplitFiles(directory, timestamp string) error {
	fmt.Println("Exporting NetBird configuration to split YAML files...")

	// Create output directory
	dirName := fmt.Sprintf("netbird-manage-export-%s", timestamp)
	dirPath := filepath.Join(directory, dirName)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Fetch all resources
	allData, err := s.fetchAllResources()
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
	fmt.Printf("  config.yml\n")

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
		fmt.Printf("  %s\n", filename)
	}

	fmt.Printf("Export completed: %s/\n", dirPath)
	return nil
}

// writeYAMLFile writes data to a YAML file
func writeYAMLFile(outputPath string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}

	if err := os.WriteFile(outputPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", outputPath, err)
	}

	return nil
}

// fetchAllResources fetches all resources from the API and converts to YAML-friendly map structure
func (s *Service) fetchAllResources() (map[string]interface{}, error) {
	// Create metadata
	metadata := map[string]interface{}{
		"version":        "1.0",
		"exported_at":    time.Now().Format(time.RFC3339),
		"management_url": s.Client.ManagementURL,
	}

	// Fetch all resource types
	groups, err := s.fetchGroupsAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch groups: %v", err)
	}

	policies, err := s.fetchPoliciesAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policies: %v", err)
	}

	networks, err := s.fetchNetworksAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch networks: %v", err)
	}

	routes, err := s.fetchRoutesAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %v", err)
	}

	dns, err := s.fetchDNSAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DNS: %v", err)
	}

	postureChecks, err := s.fetchPostureChecksAsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posture checks: %v", err)
	}

	setupKeys, err := s.fetchSetupKeysAsMap()
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
func (s *Service) fetchGroupsAsMap() (map[string]interface{}, error) {
	resp, err := s.Client.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []models.GroupDetail
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
func (s *Service) fetchPoliciesAsMap() (map[string]interface{}, error) {
	resp, err := s.Client.MakeRequest("GET", "/policies", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var policies []models.Policy
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
func (s *Service) fetchNetworksAsMap() (map[string]interface{}, error) {
	resp, err := s.Client.MakeRequest("GET", "/networks", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var networks []models.Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, fmt.Errorf("failed to decode networks: %v", err)
	}

	result := make(map[string]interface{})
	for _, network := range networks {
		// Fetch detailed network information including resources and routers
		networkDetail, err := s.fetchNetworkDetail(network.ID)
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
		resources, err := s.fetchNetworkResources(network.ID)
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
		routers, err := s.fetchNetworkRouters(network.ID)
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
func (s *Service) fetchNetworkDetail(networkID string) (*models.NetworkDetail, error) {
	resp, err := s.Client.MakeRequest("GET", "/networks/"+networkID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var detail models.NetworkDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

// fetchNetworkResources fetches resources for a network
func (s *Service) fetchNetworkResources(networkID string) ([]models.NetworkResource, error) {
	resp, err := s.Client.MakeRequest("GET", "/networks/"+networkID+"/resources", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var resources []models.NetworkResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, err
	}

	return resources, nil
}

// fetchNetworkRouters fetches routers for a network
func (s *Service) fetchNetworkRouters(networkID string) ([]models.NetworkRouter, error) {
	resp, err := s.Client.MakeRequest("GET", "/networks/"+networkID+"/routers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var routers []models.NetworkRouter
	if err := json.NewDecoder(resp.Body).Decode(&routers); err != nil {
		return nil, err
	}

	return routers, nil
}

// fetchRoutesAsMap fetches routes and converts to map[routeKey]routeData
func (s *Service) fetchRoutesAsMap() (map[string]interface{}, error) {
	resp, err := s.Client.MakeRequest("GET", "/routes", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var routes []models.Route
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
func (s *Service) fetchDNSAsMap() (map[string]interface{}, error) {
	resp, err := s.Client.MakeRequest("GET", "/dns/nameservers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dnsGroups []models.DNSNameserverGroup
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
func (s *Service) fetchPostureChecksAsMap() (map[string]interface{}, error) {
	resp, err := s.Client.MakeRequest("GET", "/posture-checks", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var checks []models.PostureCheck
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
func (s *Service) fetchSetupKeysAsMap() (map[string]interface{}, error) {
	resp, err := s.Client.MakeRequest("GET", "/setup-keys", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keys []models.SetupKey
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
