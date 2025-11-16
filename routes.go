// routes.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"text/tabwriter"
)

// handleRoutesCommand routes route-related commands
func handleRoutesCommand(client *Client, args []string) error {
	routeCmd := flag.NewFlagSet("route", flag.ContinueOnError)
	routeCmd.SetOutput(os.Stderr)
	routeCmd.Usage = printRouteUsage

	// Query flags
	listFlag := routeCmd.Bool("list", false, "List all routes")
	inspectFlag := routeCmd.String("inspect", "", "Inspect a route by ID")
	filterNetwork := routeCmd.String("filter-network", "", "Filter by network CIDR pattern")
	filterPeer := routeCmd.String("filter-peer", "", "Filter by routing peer ID")
	enabledOnlyFlag := routeCmd.Bool("enabled-only", false, "Show only enabled routes")
	disabledOnlyFlag := routeCmd.Bool("disabled-only", false, "Show only disabled routes")

	// Create flags
	createFlag := routeCmd.String("create", "", "Create a new route with the given network CIDR")
	networkIDFlag := routeCmd.String("network-id", "", "Target network ID (required for create)")
	descriptionFlag := routeCmd.String("description", "", "Route description")
	peerFlag := routeCmd.String("peer", "", "Single routing peer ID (use OR --peer-groups)")
	peerGroupsFlag := routeCmd.String("peer-groups", "", "Peer group IDs (comma-separated, use OR --peer)")
	metricFlag := routeCmd.Int("metric", 100, "Route metric/priority (1-9999, lower = higher priority)")
	masqueradeFlag := routeCmd.Bool("masquerade", false, "Enable masquerading (NAT)")
	noMasqueradeFlag := routeCmd.Bool("no-masquerade", false, "Disable masquerading")
	groupsFlag := routeCmd.String("groups", "", "Access group IDs (comma-separated, required for create)")
	enabledFlag := routeCmd.Bool("enabled", true, "Enable route")
	disabledFlag := routeCmd.Bool("disabled", false, "Disable route")

	// Update flags
	updateFlag := routeCmd.String("update", "", "Update a route by ID")

	// Delete flags
	deleteFlag := routeCmd.String("delete", "", "Delete a route by ID")

	// Toggle flags
	enableFlag := routeCmd.String("enable", "", "Enable a route by ID")
	disableFlag := routeCmd.String("disable", "", "Disable a route by ID")

	// If no flags provided, show usage
	if len(args) == 1 {
		printRouteUsage()
		return nil
	}

	// Parse the flags
	if err := routeCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags in priority order

	// Create route
	if *createFlag != "" {
		if *networkIDFlag == "" {
			return fmt.Errorf("--network-id is required when creating a route")
		}
		if *groupsFlag == "" {
			return fmt.Errorf("--groups is required when creating a route")
		}

		masquerade := *masqueradeFlag
		if *noMasqueradeFlag {
			masquerade = false
		}

		enabled := *enabledFlag
		if *disabledFlag {
			enabled = false
		}

		return client.createRoute(*createFlag, *networkIDFlag, *descriptionFlag, *peerFlag, *peerGroupsFlag, *metricFlag, masquerade, enabled, *groupsFlag)
	}

	// Delete route
	if *deleteFlag != "" {
		return client.deleteRoute(*deleteFlag)
	}

	// Enable route
	if *enableFlag != "" {
		return client.toggleRoute(*enableFlag, true)
	}

	// Disable route
	if *disableFlag != "" {
		return client.toggleRoute(*disableFlag, false)
	}

	// Update route
	if *updateFlag != "" {
		masquerade := *masqueradeFlag
		if *noMasqueradeFlag {
			masquerade = false
		}

		enabled := *enabledFlag
		if *disabledFlag {
			enabled = false
		}

		return client.updateRoute(*updateFlag, *networkIDFlag, *descriptionFlag, *peerFlag, *peerGroupsFlag, *metricFlag, masquerade, enabled, *groupsFlag)
	}

	// Inspect route
	if *inspectFlag != "" {
		return client.inspectRoute(*inspectFlag)
	}

	// List routes
	if *listFlag {
		filters := &RouteFilters{
			NetworkPattern: *filterNetwork,
			PeerID:         *filterPeer,
			EnabledOnly:    *enabledOnlyFlag,
			DisabledOnly:   *disabledOnlyFlag,
		}
		return client.listRoutes(filters)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'route' command.")
	printRouteUsage()
	return nil
}

// RouteFilters holds filtering options for listing routes
type RouteFilters struct {
	NetworkPattern string
	PeerID         string
	EnabledOnly    bool
	DisabledOnly   bool
}

// listRoutes implements the "route --list" command
func (c *Client) listRoutes(filters *RouteFilters) error {
	resp, err := c.makeRequest("GET", "/routes", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return fmt.Errorf("failed to decode routes response: %v", err)
	}

	// Apply filters
	var filtered []Route
	for _, route := range routes {
		// Filter by network pattern
		if filters.NetworkPattern != "" && !strings.Contains(strings.ToLower(route.Network), strings.ToLower(filters.NetworkPattern)) {
			continue
		}

		// Filter by peer ID
		if filters.PeerID != "" && route.Peer != filters.PeerID {
			continue
		}

		// Filter by enabled/disabled
		if filters.EnabledOnly && !route.Enabled {
			continue
		}
		if filters.DisabledOnly && route.Enabled {
			continue
		}

		filtered = append(filtered, route)
	}

	if len(filtered) == 0 {
		fmt.Println("No routes found.")
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNETWORK\tTYPE\tMETRIC\tPEER/GROUPS\tMASQ\tENABLED\tGROUPS")
	fmt.Fprintln(w, "--\t-------\t----\t------\t-----------\t----\t-------\t------")

	for _, route := range filtered {
		peerInfo := "-"
		if route.Peer != "" {
			peerInfo = fmt.Sprintf("peer:%s", route.Peer[:8])
		} else if len(route.PeerGroups) > 0 {
			peerInfo = fmt.Sprintf("%d groups", len(route.PeerGroups))
		}

		masqStr := "No"
		if route.Masquerade {
			masqStr = "Yes"
		}

		groupsStr := fmt.Sprintf("%d groups", len(route.Groups))

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%t\t%s\n",
			route.ID,
			route.Network,
			route.NetworkType,
			route.Metric,
			peerInfo,
			masqStr,
			route.Enabled,
			groupsStr,
		)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d routes\n", len(filtered))
	return nil
}

// inspectRoute implements the "route --inspect" command
func (c *Client) inspectRoute(routeID string) error {
	resp, err := c.makeRequest("GET", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var route Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return fmt.Errorf("failed to decode route response: %v", err)
	}

	// Print detailed route information
	fmt.Println("Route Details:")
	fmt.Println("==============")
	fmt.Printf("ID:             %s\n", route.ID)
	fmt.Printf("Network:        %s\n", route.Network)
	fmt.Printf("Network Type:   %s\n", route.NetworkType)
	fmt.Printf("Network ID:     %s\n", route.NetworkID)
	fmt.Printf("Metric:         %d (lower = higher priority)\n", route.Metric)
	fmt.Printf("Masquerade:     %t\n", route.Masquerade)
	fmt.Printf("Enabled:        %t\n", route.Enabled)

	if route.Description != "" {
		fmt.Printf("Description:    %s\n", route.Description)
	}

	fmt.Println()
	fmt.Println("Routing Configuration:")
	fmt.Println("----------------------")
	if route.Peer != "" {
		fmt.Printf("Routing Peer:   %s\n", route.Peer)
	} else if len(route.PeerGroups) > 0 {
		fmt.Printf("Peer Groups:    %s\n", strings.Join(route.PeerGroups, ", "))
	} else {
		fmt.Printf("Routing Peers:  None configured\n")
	}

	fmt.Println()
	fmt.Println("Access Groups:")
	fmt.Println("--------------")
	if len(route.Groups) > 0 {
		for _, groupID := range route.Groups {
			fmt.Printf("  - %s\n", groupID)
		}
	} else {
		fmt.Println("  None")
	}

	return nil
}

// createRoute implements the "route --create" command
func (c *Client) createRoute(network, networkID, description, peer, peerGroups string, metric int, masquerade, enabled bool, groups string) error {
	// Validate network CIDR
	if err := validateCIDR(network); err != nil {
		return err
	}

	// Validate metric range
	if metric < 1 || metric > 9999 {
		return fmt.Errorf("metric must be between 1 and 9999 (got %d)", metric)
	}

	// Validate peer vs peer groups (mutually exclusive)
	if peer != "" && peerGroups != "" {
		return fmt.Errorf("cannot specify both --peer and --peer-groups (use one or the other)")
	}

	// Parse groups
	groupList := splitCommaList(groups)
	if len(groupList) == 0 {
		return fmt.Errorf("at least one group is required")
	}

	// Parse peer groups if provided
	var peerGroupList []string
	if peerGroups != "" {
		peerGroupList = splitCommaList(peerGroups)
	}

	reqBody := RouteRequest{
		Description: description,
		NetworkID:   networkID,
		Network:     network,
		Peer:        peer,
		PeerGroups:  peerGroupList,
		Metric:      metric,
		Masquerade:  masquerade,
		Enabled:     enabled,
		Groups:      groupList,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/routes", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var createdRoute Route
	if err := json.NewDecoder(resp.Body).Decode(&createdRoute); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Route created successfully!\n")
	fmt.Printf("  ID:         %s\n", createdRoute.ID)
	fmt.Printf("  Network:    %s (%s)\n", createdRoute.Network, createdRoute.NetworkType)
	fmt.Printf("  Metric:     %d\n", createdRoute.Metric)
	fmt.Printf("  Masquerade: %t\n", createdRoute.Masquerade)
	fmt.Printf("  Enabled:    %t\n", createdRoute.Enabled)
	return nil
}

// updateRoute implements the "route --update" command
func (c *Client) updateRoute(routeID, networkID, description, peer, peerGroups string, metric int, masquerade, enabled bool, groups string) error {
	// First, get the current route
	resp, err := c.makeRequest("GET", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentRoute Route
	if err := json.NewDecoder(resp.Body).Decode(&currentRoute); err != nil {
		return fmt.Errorf("failed to decode current route: %v", err)
	}

	// Build update request (update only provided fields)
	updateReq := RouteRequest{
		Description: currentRoute.Description,
		NetworkID:   currentRoute.NetworkID,
		Network:     currentRoute.Network,
		Peer:        currentRoute.Peer,
		PeerGroups:  currentRoute.PeerGroups,
		Metric:      currentRoute.Metric,
		Masquerade:  masquerade,
		Enabled:     enabled,
		Groups:      currentRoute.Groups,
	}

	// Update fields if provided
	if networkID != "" {
		updateReq.NetworkID = networkID
	}
	if description != "" {
		updateReq.Description = description
	}
	if peer != "" {
		updateReq.Peer = peer
		updateReq.PeerGroups = nil
	}
	if peerGroups != "" {
		updateReq.PeerGroups = splitCommaList(peerGroups)
		updateReq.Peer = ""
	}
	if metric != 100 { // Only update if not default
		if metric < 1 || metric > 9999 {
			return fmt.Errorf("metric must be between 1 and 9999 (got %d)", metric)
		}
		updateReq.Metric = metric
	}
	if groups != "" {
		updateReq.Groups = splitCommaList(groups)
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/routes/"+routeID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Route %s updated successfully\n", routeID)
	return nil
}

// deleteRoute implements the "route --delete" command
func (c *Client) deleteRoute(routeID string) error {
	resp, err := c.makeRequest("DELETE", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Route %s deleted successfully\n", routeID)
	return nil
}

// toggleRoute enables or disables a route
func (c *Client) toggleRoute(routeID string, enable bool) error {
	// First, get the current route
	resp, err := c.makeRequest("GET", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var route Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return fmt.Errorf("failed to decode route: %v", err)
	}

	// Update the enabled status
	route.Enabled = enable

	updateReq := RouteRequest{
		Description: route.Description,
		NetworkID:   route.NetworkID,
		Network:     route.Network,
		Peer:        route.Peer,
		PeerGroups:  route.PeerGroups,
		Metric:      route.Metric,
		Masquerade:  route.Masquerade,
		Enabled:     enable,
		Groups:      route.Groups,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/routes/"+routeID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	status := "enabled"
	if !enable {
		status = "disabled"
	}
	fmt.Printf("✓ Route %s %s successfully\n", routeID, status)
	return nil
}

// validateCIDR validates a CIDR notation network address
func validateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR notation '%s': %v", cidr, err)
	}
	return nil
}

// printRouteUsage prints usage information for the route command
func printRouteUsage() {
	fmt.Println("Usage: netbird-manage route [options]")
	fmt.Println("\nManage network routes and routing configuration.")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                              List all routes")
	fmt.Println("    --filter-network <pattern>        Filter by network CIDR pattern")
	fmt.Println("    --filter-peer <peer-id>           Filter by routing peer ID")
	fmt.Println("    --enabled-only                    Show only enabled routes")
	fmt.Println("    --disabled-only                   Show only disabled routes")
	fmt.Println("  --inspect <route-id>                Show detailed route information")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --create <network-cidr>             Create a new route")
	fmt.Println("    --network-id <id>                 Target network ID (required)")
	fmt.Println("    --description <desc>              Route description")
	fmt.Println("    --peer <peer-id>                  Single routing peer (use OR --peer-groups)")
	fmt.Println("    --peer-groups <id1,id2,...>       Peer groups as routers (use OR --peer)")
	fmt.Println("    --metric <1-9999>                 Route priority (default: 100, lower = higher priority)")
	fmt.Println("    --masquerade                      Enable masquerading (NAT)")
	fmt.Println("    --no-masquerade                   Disable masquerading (default)")
	fmt.Println("    --groups <id1,id2,...>            Access group IDs (required)")
	fmt.Println("    --enabled                         Enable route (default)")
	fmt.Println("    --disabled                        Disable route")
	fmt.Println()
	fmt.Println("  --update <route-id>                 Update an existing route")
	fmt.Println("    [same flags as create]            All fields optional")
	fmt.Println()
	fmt.Println("  --delete <route-id>                 Delete a route")
	fmt.Println()
	fmt.Println("  --enable <route-id>                 Enable a route")
	fmt.Println("  --disable <route-id>                Disable a route")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List all routes")
	fmt.Println("  netbird-manage route --list")
	fmt.Println()
	fmt.Println("  # Create a route for 10.0.0.0/16 network")
	fmt.Println("  netbird-manage route --create \"10.0.0.0/16\" \\")
	fmt.Println("    --network-id <network-id> \\")
	fmt.Println("    --peer <peer-id> \\")
	fmt.Println("    --groups <group-id> \\")
	fmt.Println("    --metric 100 \\")
	fmt.Println("    --masquerade")
	fmt.Println()
	fmt.Println("  # Inspect a route")
	fmt.Println("  netbird-manage route --inspect <route-id>")
	fmt.Println()
	fmt.Println("  # Update route metric")
	fmt.Println("  netbird-manage route --update <route-id> --metric 50")
	fmt.Println()
	fmt.Println("  # Disable a route")
	fmt.Println("  netbird-manage route --disable <route-id>")
}
