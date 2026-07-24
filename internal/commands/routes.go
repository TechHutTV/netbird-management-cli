// routes.go
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

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// RouteFilters holds filtering options for listing routes
type RouteFilters struct {
	NetworkPattern string
	PeerID         string
	EnabledOnly    bool
	DisabledOnly   bool
}

// HandleRoutesCommand routes route-related commands
func (s *Service) HandleRoutesCommand(args []string) error {
	routeCmd := flag.NewFlagSet("route", flag.ContinueOnError)
	routeCmd.SetOutput(os.Stderr)
	routeCmd.Usage = PrintRouteUsage

	// Query flags
	listFlag := routeCmd.Bool("list", false, "List all routes")
	inspectFlag := routeCmd.String("inspect", "", "Inspect a route by ID")
	filterNetwork := routeCmd.String("filter-network", "", "Filter by network CIDR pattern")
	filterPeer := routeCmd.String("filter-peer", "", "Filter by routing peer ID")
	enabledOnlyFlag := routeCmd.Bool("enabled-only", false, "Show only enabled routes")
	disabledOnlyFlag := routeCmd.Bool("disabled-only", false, "Show only disabled routes")

	// Create flags
	createFlag := routeCmd.String("create", "", "Create a new route with the given network CIDR or comma-separated domain list")
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
	domainsFlag := routeCmd.String("domains", "", "Domain list for domain-based routes (comma-separated, use with --update)")
	keepRouteFlag := routeCmd.String("keep-route", "", "Keep routes for resolved domain IPs (true/false)")
	aclGroupsFlag := routeCmd.String("access-control-groups", "", "Access control group IDs (comma-separated)")
	skipAutoApplyFlag := routeCmd.String("skip-auto-apply", "", "Skip auto-applying this exit node route on clients (true/false)")

	// Update flags
	updateFlag := routeCmd.String("update", "", "Update a route by ID")

	// Delete flags
	deleteFlag := routeCmd.String("delete", "", "Delete a route by ID")

	// Toggle flags
	enableFlag := routeCmd.String("enable", "", "Enable a route by ID")
	disableFlag := routeCmd.String("disable", "", "Disable a route by ID")

	// Output flags
	outputFlag := routeCmd.String("output", "table", "Output format: table or json")

	// If no flags provided, show usage
	if len(args) == 1 {
		PrintRouteUsage()
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

		keepRoute, err := parseOptionalBool(*keepRouteFlag, "keep-route")
		if err != nil {
			return err
		}
		skipAutoApply, err := parseOptionalBool(*skipAutoApplyFlag, "skip-auto-apply")
		if err != nil {
			return err
		}

		return s.createRoute(*createFlag, *networkIDFlag, *descriptionFlag, *peerFlag, *peerGroupsFlag, *metricFlag, masquerade, enabled, *groupsFlag, *aclGroupsFlag, keepRoute, skipAutoApply)
	}

	// Delete route
	if *deleteFlag != "" {
		return s.deleteRoute(*deleteFlag)
	}

	// Enable route
	if *enableFlag != "" {
		return s.toggleRoute(*enableFlag, true)
	}

	// Disable route
	if *disableFlag != "" {
		return s.toggleRoute(*disableFlag, false)
	}

	// Update route
	if *updateFlag != "" {
		// Determine if masquerade flags were explicitly set
		var masqueradePtr *bool
		if *masqueradeFlag {
			val := true
			masqueradePtr = &val
		} else if *noMasqueradeFlag {
			val := false
			masqueradePtr = &val
		}

		// Determine if enabled flags were explicitly set
		var enabledPtr *bool
		if *disabledFlag {
			val := false
			enabledPtr = &val
		} else if *enabledFlag && !*disabledFlag {
			// Only set enabled if explicitly passed (not just default)
			// This is tricky with boolean flags, so we check if disabled is false
			// A better approach would be to track if the flag was actually set
			// For now, we won't auto-set enabled unless explicitly requested
			enabledPtr = nil
		}

		keepRoutePtr, err := parseOptionalBool(*keepRouteFlag, "keep-route")
		if err != nil {
			return err
		}
		skipAutoApplyPtr, err := parseOptionalBool(*skipAutoApplyFlag, "skip-auto-apply")
		if err != nil {
			return err
		}

		return s.updateRoute(*updateFlag, *networkIDFlag, *descriptionFlag, *peerFlag, *peerGroupsFlag, *metricFlag, masqueradePtr, enabledPtr, *groupsFlag, *aclGroupsFlag, *domainsFlag, keepRoutePtr, skipAutoApplyPtr)
	}

	// Inspect route
	if *inspectFlag != "" {
		return s.inspectRoute(*inspectFlag, *outputFlag)
	}

	// List routes
	if *listFlag {
		filters := &RouteFilters{
			NetworkPattern: *filterNetwork,
			PeerID:         *filterPeer,
			EnabledOnly:    *enabledOnlyFlag,
			DisabledOnly:   *disabledOnlyFlag,
		}
		return s.listRoutes(filters, *outputFlag)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'route' command.")
	PrintRouteUsage()
	return nil
}

// listRoutes implements the "route --list" command
func (s *Service) listRoutes(filters *RouteFilters, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/routes", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var routes []models.Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return fmt.Errorf("failed to decode routes response: %v", err)
	}

	// Apply filters
	var filtered []models.Route
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
func (s *Service) inspectRoute(routeID string, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var route models.Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return fmt.Errorf("failed to decode route response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(route, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Print detailed route information
	fmt.Println("Route Details:")
	fmt.Println("==============")
	fmt.Printf("ID:             %s\n", route.ID)
	if len(route.Domains) > 0 {
		fmt.Printf("Domains:        %s\n", strings.Join(route.Domains, ", "))
	} else {
		fmt.Printf("Network:        %s\n", route.Network)
	}
	fmt.Printf("Network Type:   %s\n", route.NetworkType)
	fmt.Printf("Network ID:     %s\n", route.NetworkID)
	fmt.Printf("Metric:         %d (lower = higher priority)\n", route.Metric)
	fmt.Printf("Masquerade:     %t\n", route.Masquerade)
	fmt.Printf("Enabled:        %t\n", route.Enabled)
	fmt.Printf("Keep Route:     %t\n", route.KeepRoute)
	fmt.Printf("Skip Auto Apply: %t\n", route.SkipAutoApply)

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

	if len(route.AccessControlGroups) > 0 {
		fmt.Println()
		fmt.Println("Access Control Groups:")
		fmt.Println("----------------------")
		for _, groupID := range route.AccessControlGroups {
			fmt.Printf("  - %s\n", groupID)
		}
	}

	return nil
}

// parseOptionalBool parses an optional true/false string flag, returning nil when unset
func parseOptionalBool(value, flagName string) (*bool, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid value for --%s: %v", flagName, err)
	}
	return &parsed, nil
}

// parseDomainList validates a comma-separated domain list for domain-based routes
func parseDomainList(spec string) ([]string, error) {
	domains := helpers.SplitCommaList(spec)
	if len(domains) == 0 {
		return nil, fmt.Errorf("no domains provided")
	}
	if len(domains) > 32 {
		return nil, fmt.Errorf("at most 32 domains are allowed (got %d)", len(domains))
	}
	for _, domain := range domains {
		if !strings.Contains(domain, ".") || strings.ContainsAny(domain, " /") {
			return nil, fmt.Errorf("invalid domain %q", domain)
		}
	}
	return domains, nil
}

// createRoute implements the "route --create" command.
// The network argument accepts a CIDR (network route) or a comma-separated
// domain list (domain-based route).
func (s *Service) createRoute(network, networkID, description, peer, peerGroups string, metric int, masquerade, enabled bool, groups, aclGroups string, keepRoute, skipAutoApply *bool) error {
	// Decide between a CIDR route and a domain-based route
	var domains []string
	if err := validateCIDR(network); err != nil {
		parsedDomains, domainErr := parseDomainList(network)
		if domainErr != nil {
			return fmt.Errorf("--create must be a network CIDR or a comma-separated domain list: %v", err)
		}
		domains = parsedDomains
		network = ""
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
	groupList := helpers.SplitCommaList(groups)
	if len(groupList) == 0 {
		return fmt.Errorf("at least one group is required")
	}

	// Parse peer groups if provided
	var peerGroupList []string
	if peerGroups != "" {
		peerGroupList = helpers.SplitCommaList(peerGroups)
	}

	reqBody := models.RouteRequest{
		Description: description,
		NetworkID:   networkID,
		Network:     network,
		Domains:     domains,
		Peer:        peer,
		PeerGroups:  peerGroupList,
		Metric:      metric,
		Masquerade:  masquerade,
		Enabled:     enabled,
		Groups:      groupList,
	}
	if aclGroups != "" {
		reqBody.AccessControlGroups = helpers.SplitCommaList(aclGroups)
	}
	if keepRoute != nil {
		reqBody.KeepRoute = *keepRoute
	}
	if skipAutoApply != nil {
		reqBody.SkipAutoApply = *skipAutoApply
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/routes", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var createdRoute models.Route
	if err := json.NewDecoder(resp.Body).Decode(&createdRoute); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Route created successfully!\n")
	fmt.Printf("  ID:         %s\n", createdRoute.ID)
	if len(createdRoute.Domains) > 0 {
		fmt.Printf("  Domains:    %s (%s)\n", strings.Join(createdRoute.Domains, ", "), createdRoute.NetworkType)
	} else {
		fmt.Printf("  Network:    %s (%s)\n", createdRoute.Network, createdRoute.NetworkType)
	}
	fmt.Printf("  Metric:     %d\n", createdRoute.Metric)
	fmt.Printf("  Masquerade: %t\n", createdRoute.Masquerade)
	fmt.Printf("  Enabled:    %t\n", createdRoute.Enabled)
	return nil
}

// updateRoute implements the "route --update" command
func (s *Service) updateRoute(routeID, networkID, description, peer, peerGroups string, metric int, masquerade, enabled *bool, groups, aclGroups, domains string, keepRoute, skipAutoApply *bool) error {
	// First, get the current route
	resp, err := s.Client.MakeRequest("GET", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentRoute models.Route
	if err := json.NewDecoder(resp.Body).Decode(&currentRoute); err != nil {
		return fmt.Errorf("failed to decode current route: %v", err)
	}

	// Build update request (update only provided fields)
	updateReq := models.RouteRequest{
		Description:         currentRoute.Description,
		NetworkID:           currentRoute.NetworkID,
		Network:             currentRoute.Network,
		Domains:             currentRoute.Domains,
		Peer:                currentRoute.Peer,
		PeerGroups:          currentRoute.PeerGroups,
		Metric:              currentRoute.Metric,
		Masquerade:          currentRoute.Masquerade,
		Enabled:             currentRoute.Enabled,
		Groups:              currentRoute.Groups,
		AccessControlGroups: currentRoute.AccessControlGroups,
		KeepRoute:           currentRoute.KeepRoute,
		SkipAutoApply:       currentRoute.SkipAutoApply,
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
		updateReq.PeerGroups = helpers.SplitCommaList(peerGroups)
		updateReq.Peer = ""
	}
	if metric != 100 { // Only update if not default
		if metric < 1 || metric > 9999 {
			return fmt.Errorf("metric must be between 1 and 9999 (got %d)", metric)
		}
		updateReq.Metric = metric
	}
	if groups != "" {
		updateReq.Groups = helpers.SplitCommaList(groups)
	}
	if aclGroups != "" {
		updateReq.AccessControlGroups = helpers.SplitCommaList(aclGroups)
	}
	if domains != "" {
		domainList, err := parseDomainList(domains)
		if err != nil {
			return err
		}
		updateReq.Domains = domainList
		updateReq.Network = ""
	}
	// Update masquerade if explicitly provided
	if masquerade != nil {
		updateReq.Masquerade = *masquerade
	}
	// Update enabled if explicitly provided
	if enabled != nil {
		updateReq.Enabled = *enabled
	}
	if keepRoute != nil {
		updateReq.KeepRoute = *keepRoute
	}
	if skipAutoApply != nil {
		updateReq.SkipAutoApply = *skipAutoApply
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = s.Client.MakeRequest("PUT", "/routes/"+routeID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Route %s updated successfully\n", routeID)
	return nil
}

// deleteRoute implements the "route --delete" command
func (s *Service) deleteRoute(routeID string) error {
	// Fetch route details first
	resp, err := s.Client.MakeRequest("GET", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	var route models.Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode route: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Network": route.Network,
		"Metric":  fmt.Sprintf("%d", route.Metric),
		"Enabled": fmt.Sprintf("%v", route.Enabled),
	}
	if route.Description != "" {
		details["Description"] = route.Description
	}
	if route.Peer != "" {
		details["Peer"] = route.Peer
	} else if len(route.PeerGroups) > 0 {
		details["Peer Groups"] = fmt.Sprintf("%d groups", len(route.PeerGroups))
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("route", "", routeID, details) {
		return nil // User cancelled
	}

	resp, err = s.Client.MakeRequest("DELETE", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Route %s deleted successfully\n", routeID)
	return nil
}

// toggleRoute enables or disables a route
func (s *Service) toggleRoute(routeID string, enable bool) error {
	// First, get the current route
	resp, err := s.Client.MakeRequest("GET", "/routes/"+routeID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var route models.Route
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return fmt.Errorf("failed to decode route: %v", err)
	}

	// Update the enabled status
	route.Enabled = enable

	updateReq := models.RouteRequest{
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

	resp, err = s.Client.MakeRequest("PUT", "/routes/"+routeID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	status := "enabled"
	if !enable {
		status = "disabled"
	}
	fmt.Printf("Route %s %s successfully\n", routeID, status)
	return nil
}
