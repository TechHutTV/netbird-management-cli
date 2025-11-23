// ingress_ports.go - Ingress port and peer management (Cloud-only features)
package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
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

	// Modification flags
	createFlag := ingressPortCmd.Bool("create", false, "Create port allocation (requires --peer and --target-port)")
	updateFlag := ingressPortCmd.String("update", "", "Update port allocation by its ID (requires --peer)")
	deleteFlag := ingressPortCmd.String("delete", "", "Delete port allocation by its ID (requires --peer)")

	// Port allocation parameters
	peerFlag := ingressPortCmd.String("peer", "", "Peer ID (required for all operations)")
	targetPortFlag := ingressPortCmd.Int("target-port", 0, "Target port to forward (1-65535)")
	protocolFlag := ingressPortCmd.String("protocol", "tcp", "Protocol (tcp or udp)")
	descriptionFlag := ingressPortCmd.String("description", "", "Port allocation description")

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
		return s.listIngressPorts(*peerFlag, *outputFlag)
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
		if *targetPortFlag == 0 {
			return fmt.Errorf("--target-port is required for --create")
		}
		if *targetPortFlag < 1 || *targetPortFlag > 65535 {
			return fmt.Errorf("--target-port must be between 1 and 65535")
		}

		req := models.IngressPortCreateRequest{
			TargetPort:  *targetPortFlag,
			Protocol:    *protocolFlag,
			Description: *descriptionFlag,
		}
		return s.createIngressPort(*peerFlag, req)
	}

	if *updateFlag != "" {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --update")
		}
		if *targetPortFlag == 0 {
			return fmt.Errorf("--target-port is required for --update")
		}
		if *targetPortFlag < 1 || *targetPortFlag > 65535 {
			return fmt.Errorf("--target-port must be between 1 and 65535")
		}

		req := models.IngressPortUpdateRequest{
			TargetPort:  *targetPortFlag,
			Protocol:    *protocolFlag,
			Description: *descriptionFlag,
		}
		return s.updateIngressPort(*peerFlag, *updateFlag, req)
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
	createFlag := ingressPeerCmd.Bool("create", false, "Create ingress peer (requires --name)")
	updateFlag := ingressPeerCmd.String("update", "", "Update ingress peer by its ID")
	deleteFlag := ingressPeerCmd.String("delete", "", "Delete ingress peer by its ID")

	// Ingress peer parameters
	nameFlag := ingressPeerCmd.String("name", "", "Ingress peer name")
	locationFlag := ingressPeerCmd.String("location", "", "Geographic location")
	enabledFlag := ingressPeerCmd.String("enabled", "", "Enable/disable ingress peer (true/false)")

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
		if *nameFlag == "" {
			return fmt.Errorf("--name is required for --create")
		}

		req := models.IngressPeerCreateRequest{
			Name:     *nameFlag,
			Location: *locationFlag,
		}

		// Parse enabled flag if provided
		if *enabledFlag != "" {
			enabled, err := strconv.ParseBool(*enabledFlag)
			if err != nil {
				return fmt.Errorf("invalid value for --enabled: %v", err)
			}
			req.Enabled = enabled
		} else {
			req.Enabled = true // Default to enabled
		}

		return s.createIngressPeer(req)
	}

	if *updateFlag != "" {
		req := models.IngressPeerUpdateRequest{
			Name:     *nameFlag,
			Location: *locationFlag,
		}

		// Parse enabled flag if provided
		if *enabledFlag != "" {
			enabled, err := strconv.ParseBool(*enabledFlag)
			if err != nil {
				return fmt.Errorf("invalid value for --enabled: %v", err)
			}
			req.Enabled = &enabled
		}

		return s.updateIngressPeer(*updateFlag, req)
	}

	if *deleteFlag != "" {
		return s.deleteIngressPeer(*deleteFlag)
	}

	// If no valid flags are provided, show usage
	ingressPeerCmd.Usage()
	return nil
}

// listIngressPorts lists all port allocations for a peer
func (s *Service) listIngressPorts(peerID string, outputFormat string) error {
	resp, err := s.Client.MakeRequest("GET", "/peers/"+peerID+"/ingress/ports", nil)
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
	fmt.Fprintln(w, "ALLOCATION ID\tTARGET PORT\tPUBLIC PORT\tPROTOCOL\tDESCRIPTION")
	fmt.Fprintln(w, "-------------\t-----------\t-----------\t--------\t-----------")

	for _, allocation := range allocations {
		desc := allocation.Description
		if desc == "" {
			desc = "-"
		}
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\n",
			allocation.ID,
			allocation.TargetPort,
			allocation.PublicPort,
			allocation.Protocol,
			desc,
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
	fmt.Printf("Allocation ID:  %s\n", allocation.ID)
	fmt.Printf("Peer ID:        %s\n", allocation.PeerID)
	fmt.Printf("Target Port:    %d\n", allocation.TargetPort)
	fmt.Printf("Public Port:    %d\n", allocation.PublicPort)
	fmt.Printf("Protocol:       %s\n", allocation.Protocol)
	fmt.Printf("Description:    %s\n", allocation.Description)
	if allocation.IngressPeer != "" {
		fmt.Printf("Ingress Peer:   %s\n", allocation.IngressPeer)
	}
	if allocation.CreatedAt != "" {
		fmt.Printf("Created At:     %s\n", allocation.CreatedAt)
	}
	if allocation.UpdatedAt != "" {
		fmt.Printf("Updated At:     %s\n", allocation.UpdatedAt)
	}

	return nil
}

// createIngressPort creates a new port allocation
func (s *Service) createIngressPort(peerID string, req models.IngressPortCreateRequest) error {
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
	fmt.Printf("Allocation ID:  %s\n", allocation.ID)
	fmt.Printf("Target Port:    %d\n", allocation.TargetPort)
	fmt.Printf("Public Port:    %d\n", allocation.PublicPort)
	fmt.Printf("Protocol:       %s\n", allocation.Protocol)

	return nil
}

// updateIngressPort updates an existing port allocation
func (s *Service) updateIngressPort(peerID, allocationID string, req models.IngressPortUpdateRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("PUT", "/peers/"+peerID+"/ingress/ports/"+allocationID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
		"Target Port": fmt.Sprintf("%d", allocation.TargetPort),
		"Public Port": fmt.Sprintf("%d", allocation.PublicPort),
		"Protocol":    allocation.Protocol,
	}
	if allocation.Description != "" {
		details["Description"] = allocation.Description
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("ingress port allocation", "", allocationID, details) {
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
	fmt.Fprintln(w, "INGRESS PEER ID\tNAME\tLOCATION\tHOSTNAME\tENABLED")
	fmt.Fprintln(w, "---------------\t----\t--------\t--------\t-------")

	for _, peer := range peers {
		location := peer.Location
		if location == "" {
			location = "-"
		}
		hostname := peer.Hostname
		if hostname == "" {
			hostname = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\n",
			peer.ID,
			peer.Name,
			location,
			hostname,
			peer.Enabled,
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
	fmt.Printf("Name:            %s\n", peer.Name)
	fmt.Printf("Location:        %s\n", peer.Location)
	fmt.Printf("Hostname:        %s\n", peer.Hostname)
	fmt.Printf("Enabled:         %t\n", peer.Enabled)
	if peer.CreatedAt != "" {
		fmt.Printf("Created At:      %s\n", peer.CreatedAt)
	}
	if peer.UpdatedAt != "" {
		fmt.Printf("Updated At:      %s\n", peer.UpdatedAt)
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
	fmt.Printf("Name:            %s\n", peer.Name)
	fmt.Printf("Location:        %s\n", peer.Location)
	fmt.Printf("Enabled:         %t\n", peer.Enabled)

	return nil
}

// updateIngressPeer updates an existing ingress peer
func (s *Service) updateIngressPeer(ingressPeerID string, req models.IngressPeerUpdateRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("PUT", "/ingress/peers/"+ingressPeerID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
		"Enabled": fmt.Sprintf("%v", peer.Enabled),
	}
	if peer.Location != "" {
		details["Location"] = peer.Location
	}
	if peer.Hostname != "" {
		details["Hostname"] = peer.Hostname
	}

	// Ask for confirmation
	if !helpers.ConfirmSingleDeletion("ingress peer", peer.Name, ingressPeerID, details) {
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
