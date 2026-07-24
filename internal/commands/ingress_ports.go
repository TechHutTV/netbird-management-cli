// ingress_ports.go - Ingress port and peer management (Cloud-only features)
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// HandleIngressPortsCommand routes ingress port allocation commands
func (s *Service) HandleIngressPortsCommand(args []string) error {
	// Create a new flag set for the 'ingress-port' command
	ingressPortCmd := flag.NewFlagSet("ingress-port", flag.ContinueOnError)
	ingressPortCmd.SetOutput(os.Stderr)          // Send errors to stderr
	ingressPortCmd.Usage = PrintIngressPortUsage // Set our custom usage function

	// Query flags
	listFlag := ingressPortCmd.Bool("list", false, "List port allocations for a peer (requires --peer)")
	inspectFlag := ingressPortCmd.String("inspect", "", "Inspect a port allocation by its ID (requires --peer)")
	filterNameFlag := ingressPortCmd.String("filter-name", "", "Filter allocations by name (server-side)")

	// Modification flags
	createFlag := ingressPortCmd.Bool("create", false, "Create port allocation (requires --peer and --name)")
	updateFlag := ingressPortCmd.String("update", "", "Update port allocation by its ID (requires --peer)")
	deleteFlag := ingressPortCmd.String("delete", "", "Delete port allocation by its ID (requires --peer)")

	// Port allocation parameters
	peerFlag := ingressPortCmd.String("peer", "", "Peer ID (required for all operations)")
	nameFlag := ingressPortCmd.String("name", "", "Allocation name")
	enabledFlag := ingressPortCmd.String("enabled", "", "Enable/disable the allocation (true/false)")
	portRangesFlag := ingressPortCmd.String("port-ranges", "", "Comma-separated port ranges as start-end:protocol (e.g., 80:tcp,1000-2000:udp,443:tcp/udp)")
	directPortsFlag := ingressPortCmd.String("direct-ports", "", "Direct port mapping as count:protocol (e.g., 3:tcp)")

	// Output format
	outputFlag := ingressPortCmd.String("output", "table", "Output format: table or json")

	// If no flags are provided (just 'netbird-manage ingress-port'), show usage
	if len(args) == 1 {
		PrintIngressPortUsage()
		return nil
	}

	// Parse the flags (all args *after* 'ingress-port')
	if err := ingressPortCmd.Parse(args[1:]); err != nil {
		// The flag package will print an error, so we just return
		return nil
	}

	// Handle the flags
	if *listFlag {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --list")
		}
		return s.listIngressPorts(*peerFlag, *filterNameFlag, *outputFlag)
	}

	if *inspectFlag != "" {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --inspect")
		}
		return s.inspectIngressPort(*peerFlag, *inspectFlag, *outputFlag)
	}

	if *createFlag {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --create")
		}
		if *nameFlag == "" {
			return fmt.Errorf("--name is required for --create")
		}
		if *portRangesFlag == "" && *directPortsFlag == "" {
			return fmt.Errorf("--port-ranges or --direct-ports is required for --create")
		}

		req := models.IngressPortAllocationRequest{
			Name:    *nameFlag,
			Enabled: true,
		}
		if *enabledFlag != "" {
			enabled, err := strconv.ParseBool(*enabledFlag)
			if err != nil {
				return fmt.Errorf("invalid value for --enabled: %v", err)
			}
			req.Enabled = enabled
		}
		if *portRangesFlag != "" {
			ranges, err := parseIngressPortRanges(*portRangesFlag)
			if err != nil {
				return err
			}
			req.PortRanges = ranges
		}
		if *directPortsFlag != "" {
			directPort, err := parseIngressDirectPort(*directPortsFlag)
			if err != nil {
				return err
			}
			req.DirectPort = directPort
		}

		return s.createIngressPort(*peerFlag, req)
	}

	if *updateFlag != "" {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --update")
		}
		return s.updateIngressPort(*peerFlag, *updateFlag, *nameFlag, *enabledFlag, *portRangesFlag, *directPortsFlag)
	}

	if *deleteFlag != "" {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --delete")
		}
		return s.deleteIngressPort(*peerFlag, *deleteFlag)
	}

	// If no valid flags are provided, show usage
	ingressPortCmd.Usage()
	return nil
}

// parseIngressPortRanges parses "80:tcp,1000-2000:udp,443:tcp/udp" into port range objects
func parseIngressPortRanges(spec string) ([]models.IngressPortRange, error) {
	var ranges []models.IngressPortRange
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		portPart, protocol, found := strings.Cut(part, ":")
		if !found {
			return nil, fmt.Errorf("invalid port range %q: expected start-end:protocol (e.g., 80:tcp or 1000-2000:udp)", part)
		}
		if err := validateIngressProtocol(protocol); err != nil {
			return nil, err
		}

		startStr, endStr, isRange := strings.Cut(portPart, "-")
		start, err := strconv.Atoi(startStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port in range %q: %v", part, err)
		}
		end := start
		if isRange {
			end, err = strconv.Atoi(endStr)
			if err != nil {
				return nil, fmt.Errorf("invalid port in range %q: %v", part, err)
			}
		}
		if start < 1 || start > 65535 || end < 1 || end > 65535 || end < start {
			return nil, fmt.Errorf("invalid port range %q: ports must be 1-65535 and end >= start", part)
		}

		ranges = append(ranges, models.IngressPortRange{
			Start:    start,
			End:      end,
			Protocol: protocol,
		})
	}
	if len(ranges) == 0 {
		return nil, fmt.Errorf("no valid port ranges in %q", spec)
	}
	return ranges, nil
}

// parseIngressDirectPort parses "3:tcp" into a direct port request
func parseIngressDirectPort(spec string) (*models.IngressDirectPort, error) {
	countStr, protocol, found := strings.Cut(strings.TrimSpace(spec), ":")
	if !found {
		return nil, fmt.Errorf("invalid direct port %q: expected count:protocol (e.g., 3:tcp)", spec)
	}
	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 {
		return nil, fmt.Errorf("invalid direct port count in %q", spec)
	}
	if err := validateIngressProtocol(protocol); err != nil {
		return nil, err
	}
	return &models.IngressDirectPort{Count: count, Protocol: protocol}, nil
}

// validateIngressProtocol checks a protocol value against the API enum
func validateIngressProtocol(protocol string) error {
	switch protocol {
	case "tcp", "udp", "tcp/udp":
		return nil
	}
	return fmt.Errorf("invalid protocol %q: must be tcp, udp, or tcp/udp", protocol)
}

// formatPortRangeMappings renders port range mappings as a compact string
func formatPortRangeMappings(mappings []models.PortRangeMapping) string {
	if len(mappings) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(mappings))
	for _, m := range mappings {
		translated := fmt.Sprintf("%d", m.TranslatedStart)
		if m.TranslatedEnd != m.TranslatedStart {
			translated = fmt.Sprintf("%d-%d", m.TranslatedStart, m.TranslatedEnd)
		}
		ingress := fmt.Sprintf("%d", m.IngressStart)
		if m.IngressEnd != m.IngressStart {
			ingress = fmt.Sprintf("%d-%d", m.IngressStart, m.IngressEnd)
		}
		parts = append(parts, fmt.Sprintf("%s->%s/%s", ingress, translated, m.Protocol))
	}
	return strings.Join(parts, ", ")
}

// HandleIngressPeersCommand routes ingress peer management commands
func (s *Service) HandleIngressPeersCommand(args []string) error {
	// Create a new flag set for the 'ingress-peer' command
	ingressPeerCmd := flag.NewFlagSet("ingress-peer", flag.ContinueOnError)
	ingressPeerCmd.SetOutput(os.Stderr)          // Send errors to stderr
	ingressPeerCmd.Usage = PrintIngressPeerUsage // Set our custom usage function

	// Query flags
	listFlag := ingressPeerCmd.Bool("list", false, "List all ingress peers")
	inspectFlag := ingressPeerCmd.String("inspect", "", "Inspect an ingress peer by its ID")

	// Modification flags
	createFlag := ingressPeerCmd.Bool("create", false, "Convert a peer into an ingress peer (requires --peer)")
	updateFlag := ingressPeerCmd.String("update", "", "Update ingress peer by its ID")
	deleteFlag := ingressPeerCmd.String("delete", "", "Delete ingress peer by its ID")

	// Ingress peer parameters
	peerFlag := ingressPeerCmd.String("peer", "", "Peer ID to convert into an ingress peer (for --create)")
	enabledFlag := ingressPeerCmd.String("enabled", "", "Enable/disable ingress peer (true/false)")
	fallbackFlag := ingressPeerCmd.String("fallback", "", "Mark as fallback ingress peer (true/false)")

	// Output format
	outputFlag := ingressPeerCmd.String("output", "table", "Output format: table or json")

	// If no flags are provided (just 'netbird-manage ingress-peer'), show usage
	if len(args) == 1 {
		PrintIngressPeerUsage()
		return nil
	}

	// Parse the flags (all args *after* 'ingress-peer')
	if err := ingressPeerCmd.Parse(args[1:]); err != nil {
		// The flag package will print an error, so we just return
		return nil
	}

	// Handle the flags
	if *listFlag {
		return s.listIngressPeers(*outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectIngressPeer(*inspectFlag, *outputFlag)
	}

	if *createFlag {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --create")
		}

		req := models.IngressPeerCreateRequest{
			PeerID:  *peerFlag,
			Enabled: true,
		}
		if err := applyIngressPeerBoolFlags(*enabledFlag, *fallbackFlag, &req.Enabled, &req.Fallback); err != nil {
			return err
		}

		return s.createIngressPeer(req)
	}

	if *updateFlag != "" {
		return s.updateIngressPeer(*updateFlag, *enabledFlag, *fallbackFlag)
	}

	if *deleteFlag != "" {
		return s.deleteIngressPeer(*deleteFlag)
	}

	// If no valid flags are provided, show usage
	ingressPeerCmd.Usage()
	return nil
}

// applyIngressPeerBoolFlags parses the optional --enabled and --fallback values
func applyIngressPeerBoolFlags(enabledValue, fallbackValue string, enabled, fallback *bool) error {
	if enabledValue != "" {
		parsed, err := strconv.ParseBool(enabledValue)
		if err != nil {
			return fmt.Errorf("invalid value for --enabled: %v", err)
		}
		*enabled = parsed
	}
	if fallbackValue != "" {
		parsed, err := strconv.ParseBool(fallbackValue)
		if err != nil {
			return fmt.Errorf("invalid value for --fallback: %v", err)
		}
		*fallback = parsed
	}
	return nil
}

// listIngressPorts lists all port allocations for a peer
func (s *Service) listIngressPorts(peerID, filterName, outputFormat string) error {
	endpoint := "/peers/" + peerID + "/ingress/ports"
	if filterName != "" {
		endpoint += "?name=" + url.QueryEscape(filterName)
	}
	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var allocations []models.IngressPortAllocation
	if err := json.NewDecoder(resp.Body).Decode(&allocations); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(allocations) == 0 {
		fmt.Println("No ingress port allocations found for this peer")
		return nil
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(allocations, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ALLOCATION ID\tNAME\tENABLED\tINGRESS IP\tREGION\tPORT MAPPINGS")
	fmt.Fprintln(w, "-------------\t----\t-------\t----------\t------\t-------------")

	for _, allocation := range allocations {
		region := allocation.Region
		if region == "" {
			region = "-"
		}
		ingressIP := allocation.IngressIP
		if ingressIP == "" {
			ingressIP = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%t\t%s\t%s\t%s\n",
			allocation.ID,
			allocation.Name,
			allocation.Enabled,
			ingressIP,
			region,
			formatPortRangeMappings(allocation.PortRangeMappings),
		)
	}
	w.Flush()

	return nil
}

// inspectIngressPort shows detailed information about a port allocation
func (s *Service) inspectIngressPort(peerID, allocationID string, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/peers/"+peerID+"/ingress/ports/"+allocationID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var allocation models.IngressPortAllocation
	if err := json.NewDecoder(resp.Body).Decode(&allocation); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(allocation, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Display allocation details
	fmt.Printf("Allocation ID:   %s\n", allocation.ID)
	fmt.Printf("Name:            %s\n", allocation.Name)
	fmt.Printf("Enabled:         %t\n", allocation.Enabled)
	fmt.Printf("Ingress Peer ID: %s\n", allocation.IngressPeerID)
	fmt.Printf("Ingress IP:      %s\n", allocation.IngressIP)
	fmt.Printf("Region:          %s\n", allocation.Region)
	fmt.Printf("Port Mappings:   %s\n", formatPortRangeMappings(allocation.PortRangeMappings))

	return nil
}

// createIngressPort creates a new port allocation
func (s *Service) createIngressPort(peerID string, req models.IngressPortAllocationRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/peers/"+peerID+"/ingress/ports", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var allocation models.IngressPortAllocation
	if err := json.NewDecoder(resp.Body).Decode(&allocation); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Ingress port allocation created successfully\n")
	fmt.Printf("Allocation ID:   %s\n", allocation.ID)
	fmt.Printf("Name:            %s\n", allocation.Name)
	fmt.Printf("Ingress IP:      %s\n", allocation.IngressIP)
	fmt.Printf("Port Mappings:   %s\n", formatPortRangeMappings(allocation.PortRangeMappings))

	return nil
}

// updateIngressPort updates an existing port allocation, preserving unset fields
func (s *Service) updateIngressPort(peerID, allocationID, name, enabledValue, portRangesSpec, directPortsSpec string) error {
	// Fetch current allocation so unset flags keep their values
	resp, err := s.Client.MakeRequest("GET", "/peers/"+peerID+"/ingress/ports/"+allocationID, nil)
	if err != nil {
		return err
	}
	var current models.IngressPortAllocation
	if err := json.NewDecoder(resp.Body).Decode(&current); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode port allocation: %v", err)
	}
	resp.Body.Close()

	req := models.IngressPortAllocationRequest{
		Name:    current.Name,
		Enabled: current.Enabled,
	}
	if name != "" {
		req.Name = name
	}
	if enabledValue != "" {
		enabled, err := strconv.ParseBool(enabledValue)
		if err != nil {
			return fmt.Errorf("invalid value for --enabled: %v", err)
		}
		req.Enabled = enabled
	}
	if portRangesSpec != "" {
		ranges, err := parseIngressPortRanges(portRangesSpec)
		if err != nil {
			return err
		}
		req.PortRanges = ranges
	} else {
		// Preserve current ranges (translated side is the peer's requested range)
		for _, m := range current.PortRangeMappings {
			req.PortRanges = append(req.PortRanges, models.IngressPortRange{
				Start:    m.TranslatedStart,
				End:      m.TranslatedEnd,
				Protocol: m.Protocol,
			})
		}
	}
	if directPortsSpec != "" {
		directPort, err := parseIngressDirectPort(directPortsSpec)
		if err != nil {
			return err
		}
		req.DirectPort = directPort
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	updateResp, err := s.Client.MakeRequest("PUT", "/peers/"+peerID+"/ingress/ports/"+allocationID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer updateResp.Body.Close()

	fmt.Printf("Ingress port allocation %s updated successfully\n", allocationID)
	return nil
}

// deleteIngressPort deletes a port allocation
func (s *Service) deleteIngressPort(peerID, allocationID string) error {
	// Fetch port allocation details first
	resp, err := s.Client.MakeRequest("GET", "/peers/"+peerID+"/ingress/ports/"+allocationID, nil)
	if err != nil {
		return err
	}
	var allocation models.IngressPortAllocation
	if err := json.NewDecoder(resp.Body).Decode(&allocation); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode port allocation: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Enabled":       fmt.Sprintf("%t", allocation.Enabled),
		"Port Mappings": formatPortRangeMappings(allocation.PortRangeMappings),
	}
	if allocation.IngressIP != "" {
		details["Ingress IP"] = allocation.IngressIP
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("ingress port allocation", allocation.Name, allocationID, details) {
		return nil // User cancelled
	}

	resp, err = s.Client.MakeRequest("DELETE", "/peers/"+peerID+"/ingress/ports/"+allocationID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Ingress port allocation %s deleted successfully\n", allocationID)
	return nil
}

// listIngressPeers lists all ingress peers
func (s *Service) listIngressPeers(outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/ingress/peers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peers []models.IngressPeer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(peers) == 0 {
		fmt.Println("No ingress peers found")
		return nil
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(peers, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "INGRESS PEER ID\tPEER ID\tINGRESS IP\tREGION\tENABLED\tCONNECTED\tFALLBACK")
	fmt.Fprintln(w, "---------------\t-------\t----------\t------\t-------\t---------\t--------")

	for _, peer := range peers {
		region := peer.Region
		if region == "" {
			region = "-"
		}
		ingressIP := peer.IngressIP
		if ingressIP == "" {
			ingressIP = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%t\t%t\n",
			peer.ID,
			peer.PeerID,
			ingressIP,
			region,
			peer.Enabled,
			peer.Connected,
			peer.Fallback,
		)
	}
	w.Flush()

	return nil
}

// inspectIngressPeer shows detailed information about an ingress peer
func (s *Service) inspectIngressPeer(ingressPeerID string, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/ingress/peers/"+ingressPeerID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peer models.IngressPeer
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(peer, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Display ingress peer details
	fmt.Printf("Ingress Peer ID: %s\n", peer.ID)
	fmt.Printf("Peer ID:         %s\n", peer.PeerID)
	fmt.Printf("Ingress IP:      %s\n", peer.IngressIP)
	fmt.Printf("Region:          %s\n", peer.Region)
	fmt.Printf("Enabled:         %t\n", peer.Enabled)
	fmt.Printf("Connected:       %t\n", peer.Connected)
	fmt.Printf("Fallback:        %t\n", peer.Fallback)
	if peer.AvailablePorts != nil {
		fmt.Printf("Available Ports: tcp=%d udp=%d\n", peer.AvailablePorts.TCP, peer.AvailablePorts.UDP)
	}

	return nil
}

// createIngressPeer creates a new ingress peer
func (s *Service) createIngressPeer(req models.IngressPeerCreateRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/ingress/peers", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peer models.IngressPeer
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Ingress peer created successfully\n")
	fmt.Printf("Ingress Peer ID: %s\n", peer.ID)
	fmt.Printf("Peer ID:         %s\n", peer.PeerID)
	fmt.Printf("Enabled:         %t\n", peer.Enabled)
	fmt.Printf("Fallback:        %t\n", peer.Fallback)

	return nil
}

// updateIngressPeer updates an existing ingress peer, preserving unset fields
func (s *Service) updateIngressPeer(ingressPeerID, enabledValue, fallbackValue string) error {
	// Fetch current state so unset flags keep their values
	resp, err := s.Client.MakeRequest("GET", "/ingress/peers/"+ingressPeerID, nil)
	if err != nil {
		return err
	}
	var current models.IngressPeer
	if err := json.NewDecoder(resp.Body).Decode(&current); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode ingress peer: %v", err)
	}
	resp.Body.Close()

	req := models.IngressPeerUpdateRequest{
		Enabled:  current.Enabled,
		Fallback: current.Fallback,
	}
	if err := applyIngressPeerBoolFlags(enabledValue, fallbackValue, &req.Enabled, &req.Fallback); err != nil {
		return err
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	updateResp, err := s.Client.MakeRequest("PUT", "/ingress/peers/"+ingressPeerID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer updateResp.Body.Close()

	fmt.Printf("Ingress peer %s updated successfully\n", ingressPeerID)
	return nil
}

// deleteIngressPeer deletes an ingress peer
func (s *Service) deleteIngressPeer(ingressPeerID string) error {
	// Fetch ingress peer details first
	resp, err := s.Client.MakeRequest("GET", "/ingress/peers/"+ingressPeerID, nil)
	if err != nil {
		return err
	}
	var peer models.IngressPeer
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode ingress peer: %v", err)
	}
	resp.Body.Close()

	// Build details map
	details := map[string]string{
		"Peer ID":  peer.PeerID,
		"Enabled":  fmt.Sprintf("%t", peer.Enabled),
		"Fallback": fmt.Sprintf("%t", peer.Fallback),
	}
	if peer.IngressIP != "" {
		details["Ingress IP"] = peer.IngressIP
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("ingress peer", peer.PeerID, ingressPeerID, details) {
		return nil // User cancelled
	}

	resp, err = s.Client.MakeRequest("DELETE", "/ingress/peers/"+ingressPeerID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Ingress peer %s deleted successfully\n", ingressPeerID)
	return nil
}
