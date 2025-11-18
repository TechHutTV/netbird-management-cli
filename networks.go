// networks.go
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

// handleNetworkCommand routes network-related commands
func handleNetworkCommand(client *Client, args []string) error {
	// Create a new flag set for the 'network' command
	networkCmd := flag.NewFlagSet("network", flag.ContinueOnError)
	networkCmd.SetOutput(os.Stderr)      // Send errors to stderr
	networkCmd.Usage = printNetworkUsage // Set our custom usage function

	// Query flags
	listFlag := networkCmd.Bool("list", false, "List all networks")
	filterName := networkCmd.String("filter-name", "", "Filter networks by name (supports wildcards)")
	inspectFlag := networkCmd.String("inspect", "", "Inspect a specific network by ID")

	// Network CRUD flags
	createFlag := networkCmd.String("create", "", "Create a new network")
	deleteFlag := networkCmd.String("delete", "", "Delete a network by ID")
	renameFlag := networkCmd.String("rename", "", "Rename a network by ID")
	updateFlag := networkCmd.String("update", "", "Update a network by ID")
	newName := networkCmd.String("new-name", "", "New name for network (use with --rename)")
	description := networkCmd.String("description", "", "Network description")

	// Resource management flags
	listResourcesFlag := networkCmd.String("list-resources", "", "List all resources in a network")
	inspectResourceFlag := networkCmd.Bool("inspect-resource", false, "Inspect a resource (requires --network-id and --resource-id)")
	addResourceFlag := networkCmd.String("add-resource", "", "Add a resource to a network by ID")
	updateResourceFlag := networkCmd.Bool("update-resource", false, "Update a resource (requires --network-id and --resource-id)")
	removeResourceFlag := networkCmd.Bool("remove-resource", false, "Remove a resource (requires --network-id and --resource-id)")

	// Resource-specific flags
	networkID := networkCmd.String("network-id", "", "Network ID (for resource/router operations)")
	resourceID := networkCmd.String("resource-id", "", "Resource ID")
	resourceName := networkCmd.String("name", "", "Resource/Router name")
	address := networkCmd.String("address", "", "Resource address (IP, subnet, or domain)")
	groups := networkCmd.String("groups", "", "Comma-separated group IDs")
	enabled := networkCmd.Bool("enabled", true, "Enable resource/router (default: true)")
	disabled := networkCmd.Bool("disabled", false, "Disable resource/router")

	// Router management flags
	listRoutersFlag := networkCmd.String("list-routers", "", "List all routers in a network")
	listAllRoutersFlag := networkCmd.Bool("list-all-routers", false, "List all routers across all networks")
	inspectRouterFlag := networkCmd.Bool("inspect-router", false, "Inspect a router (requires --network-id and --router-id)")
	addRouterFlag := networkCmd.String("add-router", "", "Add a router to a network by ID")
	updateRouterFlag := networkCmd.Bool("update-router", false, "Update a router (requires --network-id and --router-id)")
	removeRouterFlag := networkCmd.Bool("remove-router", false, "Remove a router (requires --network-id and --router-id)")

	// Router-specific flags
	routerID := networkCmd.String("router-id", "", "Router ID")
	peer := networkCmd.String("peer", "", "Single peer ID for router")
	peerGroups := networkCmd.String("peer-groups", "", "Comma-separated peer group IDs for router")
	metric := networkCmd.Int("metric", 100, "Route metric (1-9999, lower = higher priority)")
	masquerade := networkCmd.Bool("masquerade", false, "Enable masquerading (NAT)")
	noMasquerade := networkCmd.Bool("no-masquerade", false, "Disable masquerading")

	// If no flags are provided (just 'netbird-manage network'), show usage
	if len(args) == 1 {
		printNetworkUsage()
		return nil
	}

	// Parse the flags (all args *after* 'network')
	if err := networkCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle network CRUD operations
	if *createFlag != "" {
		return client.createNetwork(*createFlag, *description)
	}
	if *deleteFlag != "" {
		return client.deleteNetwork(*deleteFlag)
	}
	if *renameFlag != "" {
		if *newName == "" {
			fmt.Fprintln(os.Stderr, "Error: --new-name is required with --rename")
			return nil
		}
		return client.renameNetwork(*renameFlag, *newName)
	}
	if *updateFlag != "" {
		return client.updateNetworkDescription(*updateFlag, *description)
	}
	if *inspectFlag != "" {
		return client.inspectNetwork(*inspectFlag)
	}

	// Handle resource operations
	if *listResourcesFlag != "" {
		return client.listNetworkResources(*listResourcesFlag)
	}
	if *inspectResourceFlag {
		if *networkID == "" || *resourceID == "" {
			fmt.Fprintln(os.Stderr, "Error: --network-id and --resource-id are required")
			return nil
		}
		return client.inspectNetworkResource(*networkID, *resourceID)
	}
	if *addResourceFlag != "" {
		if *resourceName == "" || *address == "" || *groups == "" {
			fmt.Fprintln(os.Stderr, "Error: --name, --address, and --groups are required")
			return nil
		}
		enabledVal := *enabled && !*disabled
		return client.addNetworkResource(*addResourceFlag, *resourceName, *address, *description, *groups, enabledVal)
	}
	if *updateResourceFlag {
		if *networkID == "" || *resourceID == "" {
			fmt.Fprintln(os.Stderr, "Error: --network-id and --resource-id are required")
			return nil
		}
		enabledVal := *enabled && !*disabled
		return client.updateNetworkResource(*networkID, *resourceID, *resourceName, *address, *description, *groups, enabledVal)
	}
	if *removeResourceFlag {
		if *networkID == "" || *resourceID == "" {
			fmt.Fprintln(os.Stderr, "Error: --network-id and --resource-id are required")
			return nil
		}
		return client.removeNetworkResource(*networkID, *resourceID)
	}

	// Handle router operations
	if *listAllRoutersFlag {
		return client.listAllRouters()
	}
	if *listRoutersFlag != "" {
		return client.listNetworkRouters(*listRoutersFlag)
	}
	if *inspectRouterFlag {
		if *networkID == "" || *routerID == "" {
			fmt.Fprintln(os.Stderr, "Error: --network-id and --router-id are required")
			return nil
		}
		return client.inspectNetworkRouter(*networkID, *routerID)
	}
	if *addRouterFlag != "" {
		if *peer == "" && *peerGroups == "" {
			fmt.Fprintln(os.Stderr, "Error: Either --peer or --peer-groups is required")
			return nil
		}
		if *peer != "" && *peerGroups != "" {
			fmt.Fprintln(os.Stderr, "Error: Cannot use both --peer and --peer-groups together")
			return nil
		}
		masqueradeVal := *masquerade
		if *noMasquerade {
			masqueradeVal = false
		}
		enabledVal := *enabled && !*disabled
		return client.addNetworkRouter(*addRouterFlag, *peer, *peerGroups, *metric, masqueradeVal, enabledVal)
	}
	if *updateRouterFlag {
		if *networkID == "" || *routerID == "" {
			fmt.Fprintln(os.Stderr, "Error: --network-id and --router-id are required")
			return nil
		}
		if *peer != "" && *peerGroups != "" {
			fmt.Fprintln(os.Stderr, "Error: Cannot use both --peer and --peer-groups together")
			return nil
		}
		masqueradeVal := *masquerade
		if *noMasquerade {
			masqueradeVal = false
		}
		enabledVal := *enabled && !*disabled
		return client.updateNetworkRouter(*networkID, *routerID, *peer, *peerGroups, *metric, masqueradeVal, enabledVal)
	}
	if *removeRouterFlag {
		if *networkID == "" || *routerID == "" {
			fmt.Fprintln(os.Stderr, "Error: --network-id and --router-id are required")
			return nil
		}
		return client.removeNetworkRouter(*networkID, *routerID)
	}

	// Handle list with optional filter
	if *listFlag {
		return client.listNetworks(*filterName)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'network' command.")
	printNetworkUsage()
	return nil
}

// ========== Network CRUD Operations ==========

// listNetworks lists all networks with optional name filtering
func (c *Client) listNetworks(filterName string) error {
	resp, err := c.makeRequest("GET", "/networks", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return fmt.Errorf("failed to decode networks response: %v", err)
	}

	// Apply filter if provided
	if filterName != "" {
		var filtered []Network
		for _, net := range networks {
			if matchesPattern(net.Name, filterName) {
				filtered = append(filtered, net)
			}
		}
		networks = filtered
	}

	if len(networks) == 0 {
		if filterName != "" {
			fmt.Println("No networks found matching the specified filter.")
		} else {
			fmt.Println("No networks found.")
		}
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tROUTERS\tRESOURCES\tPOLICIES\tDESCRIPTION")
	fmt.Fprintln(w, "--\t----\t-------\t---------\t--------\t-----------")

	for _, net := range networks {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%s\n",
			net.ID,
			net.Name,
			net.RoutingPeersCount,
			len(net.Resources),
			len(net.Policies),
			net.Description,
		)
	}
	w.Flush()
	return nil
}

// inspectNetwork shows detailed information about a specific network
func (c *Client) inspectNetwork(networkID string) error {
	// Fetch basic network details
	resp, err := c.makeRequest("GET", "/networks/"+networkID, nil)
	if err != nil {
		return err
	}
	var network NetworkDetail
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode network response: %v", err)
	}
	resp.Body.Close()

	// Fetch full router details
	var routers []NetworkRouter
	if len(network.Routers) > 0 {
		resp, err = c.makeRequest("GET", "/networks/"+networkID+"/routers", nil)
		if err != nil {
			return err
		}
		if err := json.NewDecoder(resp.Body).Decode(&routers); err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to decode routers response: %v", err)
		}
		resp.Body.Close()
	}

	// Fetch full resource details
	var resources []NetworkResource
	if len(network.Resources) > 0 {
		resp, err = c.makeRequest("GET", "/networks/"+networkID+"/resources", nil)
		if err != nil {
			return err
		}
		if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to decode resources response: %v", err)
		}
		resp.Body.Close()
	}

	// Display network information
	fmt.Printf("Network: %s (%s)\n", network.Name, network.ID)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("  Description:     %s\n", network.Description)
	fmt.Printf("  Routers:         %d\n", len(routers))
	fmt.Printf("  Resources:       %d\n", len(resources))
	fmt.Printf("  Policies:        %d\n\n", len(network.Policies))

	// Display routers
	if len(routers) > 0 {
		fmt.Println("  Routers:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "    ID\tPEER/GROUPS\tMETRIC\tMASQUERADE\tENABLED")
		fmt.Fprintln(w, "    --\t-----------\t------\t----------\t-------")
		for _, router := range routers {
			peerInfo := router.Peer
			if len(router.PeerGroups) > 0 {
				peerInfo = fmt.Sprintf("Groups: %s", strings.Join(router.PeerGroups, ", "))
			}
			fmt.Fprintf(w, "    %s\t%s\t%d\t%v\t%v\n",
				router.ID,
				peerInfo,
				router.Metric,
				router.Masquerade,
				router.Enabled,
			)
		}
		w.Flush()
		fmt.Println()
	} else {
		fmt.Println("  Routers:         None")
	}

	// Display resources
	if len(resources) > 0 {
		fmt.Println("  Resources:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "    ID\tNAME\tADDRESS\tTYPE\tGROUPS\tENABLED")
		fmt.Fprintln(w, "    --\t----\t-------\t----\t------\t-------")
		for _, resource := range resources {
			// Extract group names from PolicyGroup objects
			groupNames := make([]string, len(resource.Groups))
			for i, group := range resource.Groups {
				groupNames[i] = group.Name
			}
			groupsStr := strings.Join(groupNames, ", ")
			if groupsStr == "" {
				groupsStr = "None"
			}
			fmt.Fprintf(w, "    %s\t%s\t%s\t%s\t%s\t%v\n",
				resource.ID,
				resource.Name,
				resource.Address,
				resource.Type,
				groupsStr,
				resource.Enabled,
			)
		}
		w.Flush()
	} else {
		fmt.Println("  Resources:       None")
	}

	return nil
}

// createNetwork creates a new network
func (c *Client) createNetwork(name, description string) error {
	reqBody := NetworkCreateRequest{
		Name:        name,
		Description: description,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/networks", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var network Network
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Successfully created network '%s' (ID: %s)\n", network.Name, network.ID)
	return nil
}

// deleteNetwork deletes a network by ID
func (c *Client) deleteNetwork(networkID string) error {
	// Fetch network details first to show what we're deleting
	resp, err := c.makeRequest("GET", "/networks/"+networkID, nil)
	if err != nil {
		return err
	}
	var network Network
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode network response: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Routers":   fmt.Sprintf("%d", len(network.Routers)),
		"Resources": fmt.Sprintf("%d", len(network.Resources)),
		"Policies":  fmt.Sprintf("%d", len(network.Policies)),
	}
	if network.Description != "" {
		details["Description"] = network.Description
	}

	// Ask for confirmation
	if !confirmSingleDeletion("network", network.Name, networkID, details) {
		return nil // User cancelled
	}

	resp, err = c.makeRequest("DELETE", "/networks/"+networkID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully deleted network '%s'\n", network.Name)
	return nil
}

// renameNetwork renames a network
func (c *Client) renameNetwork(networkID, newName string) error {
	// Get existing network details
	resp, err := c.makeRequest("GET", "/networks/"+networkID, nil)
	if err != nil {
		return err
	}
	var network Network
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode network: %v", err)
	}
	resp.Body.Close()

	// Update with new name
	reqBody := NetworkUpdateRequest{
		Name:        newName,
		Description: network.Description,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/networks/"+networkID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully renamed network from '%s' to '%s'\n", network.Name, newName)
	return nil
}

// updateNetworkDescription updates a network's description
func (c *Client) updateNetworkDescription(networkID, description string) error {
	// Get existing network details
	resp, err := c.makeRequest("GET", "/networks/"+networkID, nil)
	if err != nil {
		return err
	}
	var network Network
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode network: %v", err)
	}
	resp.Body.Close()

	// Update with new description
	reqBody := NetworkUpdateRequest{
		Name:        network.Name,
		Description: description,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/networks/"+networkID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully updated description for network '%s'\n", network.Name)
	return nil
}

// ========== Network Resources Management ==========

// listNetworkResources lists all resources in a network
func (c *Client) listNetworkResources(networkID string) error {
	resp, err := c.makeRequest("GET", "/networks/"+networkID+"/resources", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var resources []NetworkResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return fmt.Errorf("failed to decode resources response: %v", err)
	}

	if len(resources) == 0 {
		fmt.Println("No resources found in this network.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tADDRESS\tTYPE\tGROUPS\tENABLED")
	fmt.Fprintln(w, "--\t----\t-------\t----\t------\t-------")

	for _, resource := range resources {
		// Extract group names from PolicyGroup objects
		groupNames := make([]string, len(resource.Groups))
		for i, group := range resource.Groups {
			groupNames[i] = group.Name
		}
		groupsStr := strings.Join(groupNames, ", ")
		if groupsStr == "" {
			groupsStr = "None"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%v\n",
			resource.ID,
			resource.Name,
			resource.Address,
			resource.Type,
			groupsStr,
			resource.Enabled,
		)
	}
	w.Flush()
	return nil
}

// inspectNetworkResource shows detailed information about a resource
func (c *Client) inspectNetworkResource(networkID, resourceID string) error {
	resp, err := c.makeRequest("GET", "/networks/"+networkID+"/resources/"+resourceID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var resource NetworkResource
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return fmt.Errorf("failed to decode resource response: %v", err)
	}

	// Extract group names from PolicyGroup objects
	groupNames := make([]string, len(resource.Groups))
	for i, group := range resource.Groups {
		groupNames[i] = fmt.Sprintf("%s (%s)", group.Name, group.ID)
	}

	fmt.Printf("Resource: %s (%s)\n", resource.Name, resource.ID)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("  Type:        %s\n", resource.Type)
	fmt.Printf("  Address:     %s\n", resource.Address)
	fmt.Printf("  Description: %s\n", resource.Description)
	fmt.Printf("  Enabled:     %v\n", resource.Enabled)
	fmt.Printf("  Groups:      %s\n", strings.Join(groupNames, ", "))

	return nil
}

// addNetworkResource adds a resource to a network
func (c *Client) addNetworkResource(networkID, name, address, description, groupsStr string, enabled bool) error {
	// Validate address format
	if err := validateNetworkAddress(address); err != nil {
		return err
	}

	groupIDs := splitCommaList(groupsStr)
	if len(groupIDs) == 0 {
		return fmt.Errorf("at least one group ID is required")
	}

	reqBody := NetworkResourceRequest{
		Name:        name,
		Address:     address,
		Description: description,
		Enabled:     enabled,
		Groups:      groupIDs,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/networks/"+networkID+"/resources", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var resource NetworkResource
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Successfully added resource '%s' (ID: %s) to network\n", resource.Name, resource.ID)
	return nil
}

// updateNetworkResource updates a resource in a network
func (c *Client) updateNetworkResource(networkID, resourceID, name, address, description, groupsStr string, enabled bool) error {
	// Get existing resource
	resp, err := c.makeRequest("GET", "/networks/"+networkID+"/resources/"+resourceID, nil)
	if err != nil {
		return err
	}
	var resource NetworkResource
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode resource: %v", err)
	}
	resp.Body.Close()

	// Extract current group IDs
	currentGroupIDs := make([]string, len(resource.Groups))
	for i, group := range resource.Groups {
		currentGroupIDs[i] = group.ID
	}

	// Update fields if provided
	if name != "" {
		resource.Name = name
	}
	if address != "" {
		if err := validateNetworkAddress(address); err != nil {
			return err
		}
		resource.Address = address
	}
	if description != "" {
		resource.Description = description
	}

	// Use new groups if provided, otherwise keep current
	var groupIDs []string
	if groupsStr != "" {
		groupIDs = splitCommaList(groupsStr)
	} else {
		groupIDs = currentGroupIDs
	}

	resource.Enabled = enabled

	reqBody := NetworkResourceRequest{
		Name:        resource.Name,
		Address:     resource.Address,
		Description: resource.Description,
		Enabled:     resource.Enabled,
		Groups:      groupIDs,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/networks/"+networkID+"/resources/"+resourceID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully updated resource '%s'\n", resource.Name)
	return nil
}

// removeNetworkResource removes a resource from a network
func (c *Client) removeNetworkResource(networkID, resourceID string) error {
	// Fetch resource details first
	resp, err := c.makeRequest("GET", "/networks/"+networkID+"/resources/"+resourceID, nil)
	if err != nil {
		return err
	}
	var resource NetworkResource
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode resource: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Address": resource.Address,
		"Type":    resource.Type,
		"Enabled": fmt.Sprintf("%v", resource.Enabled),
	}
	if resource.Description != "" {
		details["Description"] = resource.Description
	}

	// Ask for confirmation
	if !confirmSingleDeletion("network resource", resource.Name, resourceID, details) {
		return nil // User cancelled
	}

	resp, err = c.makeRequest("DELETE", "/networks/"+networkID+"/resources/"+resourceID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully removed resource from network\n")
	return nil
}

// ========== Network Routers Management ==========

// listAllRouters lists all routers across all networks
func (c *Client) listAllRouters() error {
	resp, err := c.makeRequest("GET", "/networks/routers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var routers []NetworkRouter
	if err := json.NewDecoder(resp.Body).Decode(&routers); err != nil {
		return fmt.Errorf("failed to decode routers response: %v", err)
	}

	if len(routers) == 0 {
		fmt.Println("No routers found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tPEER/GROUPS\tMETRIC\tMASQUERADE\tENABLED")
	fmt.Fprintln(w, "--\t-----------\t------\t----------\t-------")

	for _, router := range routers {
		peerInfo := router.Peer
		if len(router.PeerGroups) > 0 {
			peerInfo = fmt.Sprintf("Groups: %s", strings.Join(router.PeerGroups, ", "))
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%v\t%v\n",
			router.ID,
			peerInfo,
			router.Metric,
			router.Masquerade,
			router.Enabled,
		)
	}
	w.Flush()
	return nil
}

// listNetworkRouters lists all routers in a specific network
func (c *Client) listNetworkRouters(networkID string) error {
	resp, err := c.makeRequest("GET", "/networks/"+networkID+"/routers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var routers []NetworkRouter
	if err := json.NewDecoder(resp.Body).Decode(&routers); err != nil {
		return fmt.Errorf("failed to decode routers response: %v", err)
	}

	if len(routers) == 0 {
		fmt.Println("No routers found in this network.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tPEER/GROUPS\tMETRIC\tMASQUERADE\tENABLED")
	fmt.Fprintln(w, "--\t-----------\t------\t----------\t-------")

	for _, router := range routers {
		peerInfo := router.Peer
		if len(router.PeerGroups) > 0 {
			peerInfo = fmt.Sprintf("Groups: %s", strings.Join(router.PeerGroups, ", "))
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%v\t%v\n",
			router.ID,
			peerInfo,
			router.Metric,
			router.Masquerade,
			router.Enabled,
		)
	}
	w.Flush()
	return nil
}

// inspectNetworkRouter shows detailed information about a router
func (c *Client) inspectNetworkRouter(networkID, routerID string) error {
	resp, err := c.makeRequest("GET", "/networks/"+networkID+"/routers/"+routerID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var router NetworkRouter
	if err := json.NewDecoder(resp.Body).Decode(&router); err != nil {
		return fmt.Errorf("failed to decode router response: %v", err)
	}

	fmt.Printf("Router: %s\n", router.ID)
	fmt.Println(strings.Repeat("-", 50))
	if router.Peer != "" {
		fmt.Printf("  Peer:        %s\n", router.Peer)
	}
	if len(router.PeerGroups) > 0 {
		fmt.Printf("  Peer Groups: %s\n", strings.Join(router.PeerGroups, ", "))
	}
	fmt.Printf("  Metric:      %d\n", router.Metric)
	fmt.Printf("  Masquerade:  %v\n", router.Masquerade)
	fmt.Printf("  Enabled:     %v\n", router.Enabled)

	return nil
}

// addNetworkRouter adds a router to a network
func (c *Client) addNetworkRouter(networkID, peer, peerGroupsStr string, metric int, masquerade, enabled bool) error {
	// Validate metric range
	if metric < 1 || metric > 9999 {
		return fmt.Errorf("metric must be between 1 and 9999")
	}

	var peerGroups []string
	if peerGroupsStr != "" {
		peerGroups = splitCommaList(peerGroupsStr)
	}

	reqBody := NetworkRouterRequest{
		Peer:       peer,
		PeerGroups: peerGroups,
		Metric:     metric,
		Masquerade: masquerade,
		Enabled:    enabled,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/networks/"+networkID+"/routers", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var router NetworkRouter
	if err := json.NewDecoder(resp.Body).Decode(&router); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Successfully added router (ID: %s) to network\n", router.ID)
	return nil
}

// updateNetworkRouter updates a router in a network
func (c *Client) updateNetworkRouter(networkID, routerID, peer, peerGroupsStr string, metric int, masquerade, enabled bool) error {
	// Get existing router
	resp, err := c.makeRequest("GET", "/networks/"+networkID+"/routers/"+routerID, nil)
	if err != nil {
		return err
	}
	var router NetworkRouter
	if err := json.NewDecoder(resp.Body).Decode(&router); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode router: %v", err)
	}
	resp.Body.Close()

	// Validate metric range
	if metric < 1 || metric > 9999 {
		return fmt.Errorf("metric must be between 1 and 9999")
	}

	// Update fields
	if peer != "" {
		router.Peer = peer
		router.PeerGroups = nil // Clear peer groups when using single peer
	}
	if peerGroupsStr != "" {
		router.PeerGroups = splitCommaList(peerGroupsStr)
		router.Peer = "" // Clear peer when using peer groups
	}
	router.Metric = metric
	router.Masquerade = masquerade
	router.Enabled = enabled

	reqBody := NetworkRouterRequest{
		Peer:       router.Peer,
		PeerGroups: router.PeerGroups,
		Metric:     router.Metric,
		Masquerade: router.Masquerade,
		Enabled:    router.Enabled,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/networks/"+networkID+"/routers/"+routerID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully updated router %s\n", routerID)
	return nil
}

// removeNetworkRouter removes a router from a network
func (c *Client) removeNetworkRouter(networkID, routerID string) error {
	// Fetch router details first
	resp, err := c.makeRequest("GET", "/networks/"+networkID+"/routers/"+routerID, nil)
	if err != nil {
		return err
	}
	var router NetworkRouter
	if err := json.NewDecoder(resp.Body).Decode(&router); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode router: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Metric":     fmt.Sprintf("%d", router.Metric),
		"Masquerade": fmt.Sprintf("%v", router.Masquerade),
		"Enabled":    fmt.Sprintf("%v", router.Enabled),
	}
	if router.Peer != "" {
		details["Peer"] = router.Peer
	} else if len(router.PeerGroups) > 0 {
		details["Peer Groups"] = strings.Join(router.PeerGroups, ", ")
	}

	// Ask for confirmation
	if !confirmSingleDeletion("network router", "", routerID, details) {
		return nil // User cancelled
	}

	resp, err = c.makeRequest("DELETE", "/networks/"+networkID+"/routers/"+routerID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully removed router from network\n")
	return nil
}
