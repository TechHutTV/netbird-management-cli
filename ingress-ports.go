// ingress-ports.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
)

// handleIngressPortsCommand routes ingress port allocation commands
func handleIngressPortsCommand(client *Client, args []string) error {
	// Create a new flag set for the 'ingress-port' command
	ingressPortCmd := flag.NewFlagSet("ingress-port", flag.ContinueOnError)
	ingressPortCmd.SetOutput(os.Stderr)           // Send errors to stderr
	ingressPortCmd.Usage = printIngressPortUsage // Set our custom usage function

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

	// If no flags are provided (just 'netbird-manage ingress-port'), show usage
	if len(args) == 1 {
		printIngressPortUsage()
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
		return client.listIngressPorts(*peerFlag)
	}

	if *inspectFlag != "" {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --inspect")
		}
		return client.inspectIngressPort(*peerFlag, *inspectFlag)
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

		req := IngressPortCreateRequest{
			TargetPort:  *targetPortFlag,
			Protocol:    *protocolFlag,
			Description: *descriptionFlag,
		}
		return client.createIngressPort(*peerFlag, req)
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

		req := IngressPortUpdateRequest{
			TargetPort:  *targetPortFlag,
			Protocol:    *protocolFlag,
			Description: *descriptionFlag,
		}
		return client.updateIngressPort(*peerFlag, *updateFlag, req)
	}

	if *deleteFlag != "" {
		if *peerFlag == "" {
			return fmt.Errorf("--peer is required for --delete")
		}
		return client.deleteIngressPort(*peerFlag, *deleteFlag)
	}

	// If no valid flags are provided, show usage
	ingressPortCmd.Usage()
	return nil
}

// handleIngressPeersCommand routes ingress peer management commands
func handleIngressPeersCommand(client *Client, args []string) error {
	// Create a new flag set for the 'ingress-peer' command
	ingressPeerCmd := flag.NewFlagSet("ingress-peer", flag.ContinueOnError)
	ingressPeerCmd.SetOutput(os.Stderr)           // Send errors to stderr
	ingressPeerCmd.Usage = printIngressPeerUsage // Set our custom usage function

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

	// If no flags are provided (just 'netbird-manage ingress-peer'), show usage
	if len(args) == 1 {
		printIngressPeerUsage()
		return nil
	}

	// Parse the flags (all args *after* 'ingress-peer')
	if err := ingressPeerCmd.Parse(args[1:]); err != nil {
		// The flag package will print an error, so we just return
		return nil
	}

	// Handle the flags
	if *listFlag {
		return client.listIngressPeers()
	}

	if *inspectFlag != "" {
		return client.inspectIngressPeer(*inspectFlag)
	}

	if *createFlag {
		if *nameFlag == "" {
			return fmt.Errorf("--name is required for --create")
		}

		req := IngressPeerCreateRequest{
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

		return client.createIngressPeer(req)
	}

	if *updateFlag != "" {
		req := IngressPeerUpdateRequest{
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

		return client.updateIngressPeer(*updateFlag, req)
	}

	if *deleteFlag != "" {
		return client.deleteIngressPeer(*deleteFlag)
	}

	// If no valid flags are provided, show usage
	ingressPeerCmd.Usage()
	return nil
}

// listIngressPorts lists all port allocations for a peer
func (c *Client) listIngressPorts(peerID string) error {
	resp, err := c.makeRequest("GET", "/peers/"+peerID+"/ingress/ports", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var allocations []IngressPortAllocation
	if err := json.NewDecoder(resp.Body).Decode(&allocations); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(allocations) == 0 {
		fmt.Println("No ingress port allocations found for this peer")
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
func (c *Client) inspectIngressPort(peerID, allocationID string) error {
	resp, err := c.makeRequest("GET", "/peers/"+peerID+"/ingress/ports/"+allocationID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var allocation IngressPortAllocation
	if err := json.NewDecoder(resp.Body).Decode(&allocation); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
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
func (c *Client) createIngressPort(peerID string, req IngressPortCreateRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/peers/"+peerID+"/ingress/ports", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var allocation IngressPortAllocation
	if err := json.NewDecoder(resp.Body).Decode(&allocation); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Ingress port allocation created successfully\n")
	fmt.Printf("Allocation ID:  %s\n", allocation.ID)
	fmt.Printf("Target Port:    %d\n", allocation.TargetPort)
	fmt.Printf("Public Port:    %d\n", allocation.PublicPort)
	fmt.Printf("Protocol:       %s\n", allocation.Protocol)

	return nil
}

// updateIngressPort updates an existing port allocation
func (c *Client) updateIngressPort(peerID, allocationID string, req IngressPortUpdateRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("PUT", "/peers/"+peerID+"/ingress/ports/"+allocationID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Ingress port allocation %s updated successfully\n", allocationID)
	return nil
}

// deleteIngressPort deletes a port allocation
func (c *Client) deleteIngressPort(peerID, allocationID string) error {
	// Fetch port allocation details first
	resp, err := c.makeRequest("GET", "/peers/"+peerID+"/ingress/ports/"+allocationID, nil)
	if err != nil {
		return err
	}
	var allocation IngressPortAllocation
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
	if !confirmSingleDeletion("ingress port allocation", "", allocationID, details) {
		return nil // User cancelled
	}

	resp, err = c.makeRequest("DELETE", "/peers/"+peerID+"/ingress/ports/"+allocationID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Ingress port allocation %s deleted successfully\n", allocationID)
	return nil
}

// listIngressPeers lists all ingress peers
func (c *Client) listIngressPeers() error {
	resp, err := c.makeRequest("GET", "/ingress/peers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peers []IngressPeer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if len(peers) == 0 {
		fmt.Println("No ingress peers found")
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
func (c *Client) inspectIngressPeer(ingressPeerID string) error {
	resp, err := c.makeRequest("GET", "/ingress/peers/"+ingressPeerID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peer IngressPeer
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
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
func (c *Client) createIngressPeer(req IngressPeerCreateRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/ingress/peers", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peer IngressPeer
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ Ingress peer created successfully\n")
	fmt.Printf("Ingress Peer ID: %s\n", peer.ID)
	fmt.Printf("Name:            %s\n", peer.Name)
	fmt.Printf("Location:        %s\n", peer.Location)
	fmt.Printf("Enabled:         %t\n", peer.Enabled)

	return nil
}

// updateIngressPeer updates an existing ingress peer
func (c *Client) updateIngressPeer(ingressPeerID string, req IngressPeerUpdateRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("PUT", "/ingress/peers/"+ingressPeerID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Ingress peer %s updated successfully\n", ingressPeerID)
	return nil
}

// deleteIngressPeer deletes an ingress peer
func (c *Client) deleteIngressPeer(ingressPeerID string) error {
	// Fetch ingress peer details first
	resp, err := c.makeRequest("GET", "/ingress/peers/"+ingressPeerID, nil)
	if err != nil {
		return err
	}
	var peer IngressPeer
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
	if !confirmSingleDeletion("ingress peer", peer.Name, ingressPeerID, details) {
		return nil // User cancelled
	}

	resp, err = c.makeRequest("DELETE", "/ingress/peers/"+ingressPeerID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Ingress peer %s deleted successfully\n", ingressPeerID)
	return nil
}

// printIngressPortUsage prints the usage information for the ingress-port command
func printIngressPortUsage() {
	fmt.Println("⚠️  CLOUD-ONLY FEATURE: Ingress ports are only available on NetBird Cloud")
	fmt.Println("\nUsage: netbird-manage ingress-port [options]")
	fmt.Println("\nPort Allocation Commands:")
	fmt.Println("  --list --peer <peer-id>                List port allocations for peer")
	fmt.Println("  --inspect <id> --peer <peer-id>        Show port allocation details")
	fmt.Println("  --create --peer <peer-id>              Create port allocation (requires --target-port)")
	fmt.Println("  --update <id> --peer <peer-id>         Update port allocation")
	fmt.Println("  --delete <id> --peer <peer-id>         Delete port allocation")
	fmt.Println("\nPort Allocation Flags:")
	fmt.Println("  --peer <peer-id>           Target peer ID (required)")
	fmt.Println("  --target-port <port>       Port to forward (1-65535)")
	fmt.Println("  --protocol <tcp|udp>       Protocol (default: tcp)")
	fmt.Println("  --description <desc>       Port allocation description")
	fmt.Println("\nExamples:")
	fmt.Println("  netbird-manage ingress-port --list --peer abc-123")
	fmt.Println("  netbird-manage ingress-port --create --peer abc-123 --target-port 8080 --description \"Web Server\"")
	fmt.Println("  netbird-manage ingress-port --update def-456 --peer abc-123 --target-port 8443")
	fmt.Println("  netbird-manage ingress-port --delete def-456 --peer abc-123")
}

// printIngressPeerUsage prints the usage information for the ingress-peer command
func printIngressPeerUsage() {
	fmt.Println("⚠️  CLOUD-ONLY FEATURE: Ingress peers are only available on NetBird Cloud")
	fmt.Println("\nUsage: netbird-manage ingress-peer [options]")
	fmt.Println("\nIngress Peer Commands:")
	fmt.Println("  --list                     List all ingress peers")
	fmt.Println("  --inspect <id>             Show ingress peer details")
	fmt.Println("  --create                   Create ingress peer (requires --name)")
	fmt.Println("  --update <id>              Update ingress peer")
	fmt.Println("  --delete <id>              Delete ingress peer")
	fmt.Println("\nIngress Peer Flags:")
	fmt.Println("  --name <name>              Ingress peer name")
	fmt.Println("  --location <location>      Geographic location")
	fmt.Println("  --enabled <true|false>     Enable/disable ingress peer")
	fmt.Println("\nExamples:")
	fmt.Println("  netbird-manage ingress-peer --list")
	fmt.Println("  netbird-manage ingress-peer --create --name \"US West\" --location us-west-1")
	fmt.Println("  netbird-manage ingress-peer --update abc-123 --enabled false")
	fmt.Println("  netbird-manage ingress-peer --delete abc-123")
}
