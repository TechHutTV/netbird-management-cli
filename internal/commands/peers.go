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

// HandlePeersCommand routes peer-related commands using the flag package
func (s *Service) HandlePeersCommand(args []string) error {
	peerCmd := flag.NewFlagSet("peer", flag.ContinueOnError)
	peerCmd.SetOutput(os.Stderr)
	peerCmd.Usage = PrintPeerUsage

	listFlag := peerCmd.Bool("list", false, "List all peers")
	inspectFlag := peerCmd.String("inspect", "", "Inspect a peer by its ID")
	removeFlag := peerCmd.String("remove", "", "Remove a peer by its ID")
	removeBatchFlag := peerCmd.String("remove-batch", "", "Remove multiple peers (comma-separated IDs)")
	editFlag := peerCmd.String("edit", "", "Edit a peer by its ID (use with --add-group or --remove-group)")
	addGrpFlag := peerCmd.String("add-group", "", "Group to add to the peer (requires --edit)")
	rmGrpFlag := peerCmd.String("remove-group", "", "Group to remove from the peer (requires --edit)")

	updateFlag := peerCmd.String("update", "", "Update a peer by its ID (use with update flags)")
	renameFlag := peerCmd.String("rename", "", "New name for the peer (requires --update)")
	sshFlag := peerCmd.String("ssh-enabled", "", "Enable/disable SSH (true/false, requires --update)")
	loginExpFlag := peerCmd.String("login-expiration", "", "Enable/disable login expiration (true/false, requires --update)")
	inactivityExpFlag := peerCmd.String("inactivity-expiration", "", "Enable/disable inactivity expiration (true/false, requires --update)")
	approvalFlag := peerCmd.String("approval-required", "", "Enable/disable approval requirement (true/false, requires --update, cloud-only)")
	ipFlag := peerCmd.String("ip", "", "Set peer IP address (requires --update)")
	ipv6Flag := peerCmd.String("ipv6", "", "Set peer IPv6 address (requires --update)")

	tempAccessFlag := peerCmd.String("temporary-access", "", "Create a temporary access peer for the specified peer ID")
	tempNameFlag := peerCmd.String("name", "", "Name for the temporary access peer (requires --temporary-access)")
	wgPubKeyFlag := peerCmd.String("wg-pub-key", "", "WireGuard public key for the temporary access peer (requires --temporary-access)")
	rulesFlag := peerCmd.String("rules", "", "Comma-separated rules for the temporary access peer (requires --temporary-access)")

	accessiblePeersFlag := peerCmd.String("accessible-peers", "", "List peers accessible from the specified peer ID")
	filterNameFlag := peerCmd.String("filter-name", "", "Filter peers by name pattern (use with --list)")
	filterIPFlag := peerCmd.String("filter-ip", "", "Filter peers by IP pattern (use with --list)")
	outputFlag := peerCmd.String("output", "table", "Output format: table or json")

	if len(args) == 1 {
		PrintPeerUsage()
		return nil
	}

	if err := peerCmd.Parse(args[1:]); err != nil {
		return nil
	}

	if *listFlag {
		return s.listPeers(*filterNameFlag, *filterIPFlag, *outputFlag)
	}

	if *inspectFlag != "" {
		return s.inspectPeer(*inspectFlag, *outputFlag)
	}

	if *removeFlag != "" {
		return s.removePeerByID(*removeFlag)
	}

	if *removeBatchFlag != "" {
		return s.removePeersBatch(*removeBatchFlag)
	}

	if *accessiblePeersFlag != "" {
		return s.getAccessiblePeers(*accessiblePeersFlag, *outputFlag)
	}

	if *editFlag != "" {
		peerID := *editFlag
		if *addGrpFlag != "" {
			return s.modifyPeerGroup(peerID, *addGrpFlag, "add")
		}
		if *rmGrpFlag != "" {
			return s.modifyPeerGroup(peerID, *rmGrpFlag, "remove")
		}
		return fmt.Errorf("flag --edit requires --add-group or --remove-group")
	}

	if *updateFlag != "" {
		return s.handlePeerUpdate(*updateFlag, *renameFlag, *sshFlag, *loginExpFlag, *inactivityExpFlag, *approvalFlag, *ipFlag, *ipv6Flag)
	}

	if *tempAccessFlag != "" {
		if *tempNameFlag == "" || *wgPubKeyFlag == "" || *rulesFlag == "" {
			return fmt.Errorf("--temporary-access requires --name, --wg-pub-key, and --rules")
		}
		return s.createTemporaryAccessPeer(*tempAccessFlag, *tempNameFlag, *wgPubKeyFlag, helpers.SplitCommaList(*rulesFlag), *outputFlag)
	}

	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'peer' command.")
	PrintPeerUsage()
	return nil
}

func (s *Service) handlePeerUpdate(peerID, rename, ssh, loginExp, inactivityExp, approval, ip, ipv6 string) error {
	peer, err := s.getPeerByID(peerID)
	if err != nil {
		return fmt.Errorf("failed to get peer: %v", err)
	}

	updateReq := models.PeerUpdateRequest{
		Name:                        peer.Name,
		SSHEnabled:                  peer.SSHEnabled,
		LoginExpirationEnabled:      peer.LoginExpirationEnabled,
		InactivityExpirationEnabled: peer.InactivityExpirationEnabled,
	}

	changes := make([]string, 0, 6)

	if rename != "" {
		updateReq.Name = rename
		changes = append(changes, fmt.Sprintf("name: %s -> %s", peer.Name, rename))
	}

	if ssh != "" {
		sshBool, err := strconv.ParseBool(ssh)
		if err != nil {
			return fmt.Errorf("invalid value for --ssh-enabled: %s (must be true or false)", ssh)
		}
		updateReq.SSHEnabled = sshBool
		changes = append(changes, fmt.Sprintf("ssh_enabled: %t -> %t", peer.SSHEnabled, sshBool))
	}

	if loginExp != "" {
		loginExpBool, err := strconv.ParseBool(loginExp)
		if err != nil {
			return fmt.Errorf("invalid value for --login-expiration: %s (must be true or false)", loginExp)
		}
		updateReq.LoginExpirationEnabled = loginExpBool
		changes = append(changes, fmt.Sprintf("login_expiration_enabled: %t -> %t", peer.LoginExpirationEnabled, loginExpBool))
	}

	if inactivityExp != "" {
		inactivityExpBool, err := strconv.ParseBool(inactivityExp)
		if err != nil {
			return fmt.Errorf("invalid value for --inactivity-expiration: %s (must be true or false)", inactivityExp)
		}
		updateReq.InactivityExpirationEnabled = inactivityExpBool
		changes = append(changes, fmt.Sprintf("inactivity_expiration_enabled: %t -> %t", peer.InactivityExpirationEnabled, inactivityExpBool))
	}

	if approval != "" {
		approvalBool, err := strconv.ParseBool(approval)
		if err != nil {
			return fmt.Errorf("invalid value for --approval-required: %s (must be true or false)", approval)
		}
		updateReq.ApprovalRequired = &approvalBool
		if peer.ApprovalRequired != nil {
			changes = append(changes, fmt.Sprintf("approval_required: %t -> %t", *peer.ApprovalRequired, approvalBool))
		} else {
			changes = append(changes, fmt.Sprintf("approval_required: (not set) -> %t", approvalBool))
		}
	}

	if ip != "" {
		updateReq.IP = ip
		changes = append(changes, fmt.Sprintf("ip: %s -> %s", peer.IP, ip))
	}

	if ipv6 != "" {
		updateReq.IPv6 = ipv6
		changes = append(changes, fmt.Sprintf("ipv6: %s -> %s", peer.IPv6, ipv6))
	}

	if len(changes) == 0 {
		return fmt.Errorf("no update flags provided (use --rename, --ssh-enabled, --login-expiration, --inactivity-expiration, --approval-required, --ip, or --ipv6)")
	}

	fmt.Printf("Updating peer %s (%s):\n", peer.Name, peerID)
	for _, change := range changes {
		fmt.Printf("  - %s\n", change)
	}

	return s.updatePeer(peerID, updateReq)
}

func (s *Service) listPeers(filterName, filterIP, outputFormat string) error {
	// Build query parameters for server-side filtering
	params := url.Values{}
	if filterName != "" {
		params.Add("name", filterName)
	}
	if filterIP != "" {
		params.Add("ip", filterIP)
	}

	endpoint := "/peers"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peers []models.Peer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return fmt.Errorf("failed to decode peers response: %v", err)
	}

	// Apply additional local filtering for pattern matching (server does exact match)
	var filteredPeers []models.Peer
	for _, peer := range peers {
		if filterName != "" && !helpers.MatchesPattern(peer.Name, filterName) {
			continue
		}
		if filterIP != "" && !helpers.MatchesPattern(peer.IP, filterIP) {
			continue
		}
		filteredPeers = append(filteredPeers, peer)
	}

	if len(filteredPeers) == 0 {
		if filterName != "" || filterIP != "" {
			fmt.Println("No peers found matching the specified filters.")
		} else {
			fmt.Println("No peers found in your network.")
		}
		return nil
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(filteredPeers, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output (default)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tIP\tCONNECTED\tOS\tVERSION\tHOSTNAME")
	fmt.Fprintln(w, "--\t----\t--\t---------\t--\t-------\t--------")

	for _, peer := range filteredPeers {
		connectedStatus := "Offline"
		if peer.Connected {
			connectedStatus = "Online"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			peer.ID,
			peer.Name,
			peer.IP,
			connectedStatus,
			helpers.FormatOS(peer.OS),
			peer.Version,
			peer.Hostname,
		)
	}
	w.Flush()
	return nil
}

func (s *Service) getPeerByID(peerID string) (*models.Peer, error) {
	endpoint := "/peers/" + peerID
	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var peer models.Peer
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		return nil, fmt.Errorf("failed to decode peer response: %v", err)
	}
	return &peer, nil
}

func (s *Service) removePeerByID(peerID string) error {
	peer, err := s.getPeerByID(peerID)
	if err != nil {
		return fmt.Errorf("cannot remove peer: %v", err)
	}

	details := map[string]string{
		"IP":        peer.IP,
		"Hostname":  peer.Hostname,
		"OS":        helpers.FormatOS(peer.OS),
		"Connected": fmt.Sprintf("%t", peer.Connected),
	}

	if len(peer.Groups) > 0 {
		groupNames := make([]string, len(peer.Groups))
		for i, g := range peer.Groups {
			groupNames[i] = g.Name
		}
		details["Groups"] = fmt.Sprintf("%d (%s)", len(peer.Groups), strings.Join(groupNames, ", "))
	} else {
		details["Groups"] = "None"
	}

	if !helpers.ConfirmSingleDeletion("peer", peer.Name, peer.ID, details) {
		return nil
	}

	fmt.Printf("Removing peer '%s' (ID: %s)...\n", peer.Name, peer.ID)
	endpoint := "/peers/" + peer.ID
	resp, err := s.Client.MakeRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	fmt.Printf("Successfully removed peer '%s' (ID: %s)\n", peer.Name, peer.ID)
	return nil
}

func (s *Service) removePeersBatch(idList string) error {
	peerIDs := helpers.SplitCommaList(idList)
	if len(peerIDs) == 0 {
		return fmt.Errorf("no peer IDs provided")
	}

	peers := make([]*models.Peer, 0, len(peerIDs))
	itemList := make([]string, 0, len(peerIDs))

	fmt.Println("Fetching peer details...")
	for _, id := range peerIDs {
		peer, err := s.getPeerByID(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping %s: %v\n", id, err)
			continue
		}
		peers = append(peers, peer)
		itemList = append(itemList, fmt.Sprintf("%s (ID: %s, IP: %s)", peer.Name, peer.ID, peer.IP))
	}

	if len(peers) == 0 {
		return fmt.Errorf("no valid peers found to remove")
	}

	if !helpers.ConfirmBulkDeletion("peers", itemList, len(peers)) {
		return nil
	}

	var succeeded, failed int
	for i, peer := range peers {
		fmt.Printf("[%d/%d] Removing peer '%s'... ", i+1, len(peers), peer.Name)

		endpoint := "/peers/" + peer.ID
		resp, err := s.Client.MakeRequest("DELETE", endpoint, nil)
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			failed++
			continue
		}
		resp.Body.Close()
		fmt.Println("Done")
		succeeded++
	}

	fmt.Println()
	if failed > 0 {
		fmt.Fprintf(os.Stderr, "Warning: Completed: %d succeeded, %d failed\n", succeeded, failed)
	} else {
		fmt.Printf("All %d peers removed successfully\n", succeeded)
	}

	return nil
}

func (s *Service) inspectPeer(peerID, outputFormat string) error {
	peer, err := s.getPeerByID(peerID)
	if err != nil {
		return err
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

	// Table output (default)
	fmt.Printf("Inspecting Peer: %s (%s)\n", peer.Name, peer.ID)
	fmt.Println("---------------------------------")
	fmt.Printf("  IP:          %s\n", peer.IP)
	if peer.IPv6 != "" {
		fmt.Printf("  IPv6:        %s\n", peer.IPv6)
	}
	if peer.ConnectionIP != "" {
		fmt.Printf("  Public IP:   %s\n", peer.ConnectionIP)
	}
	fmt.Printf("  Hostname:    %s\n", peer.Hostname)
	if peer.DNSLabel != "" {
		fmt.Printf("  DNS Label:   %s\n", peer.DNSLabel)
	}
	if len(peer.ExtraDNSLabels) > 0 {
		fmt.Printf("  Extra DNS:   %s\n", strings.Join(peer.ExtraDNSLabels, ", "))
	}
	fmt.Printf("  OS:          %s\n", helpers.FormatOS(peer.OS))
	if peer.KernelVersion != "" {
		fmt.Printf("  Kernel:      %s\n", peer.KernelVersion)
	}
	fmt.Printf("  Version:     %s\n", peer.Version)
	if peer.UIVersion != "" {
		fmt.Printf("  UI Version:  %s\n", peer.UIVersion)
	}
	if peer.SerialNumber != "" {
		fmt.Printf("  Serial No:   %s\n", peer.SerialNumber)
	}
	fmt.Printf("  Connected:   %t\n", peer.Connected)
	fmt.Printf("  Last Seen:   %s\n", peer.LastSeen)
	if peer.CountryCode != "" {
		location := peer.CountryCode
		if peer.CityName != "" {
			location = peer.CityName + ", " + peer.CountryCode
		}
		fmt.Printf("  Location:    %s\n", location)
	}
	if peer.UserID != "" {
		fmt.Printf("  User ID:     %s\n", peer.UserID)
	}
	fmt.Printf("  SSH Enabled: %t\n", peer.SSHEnabled)
	fmt.Printf("  Login Exp:   %t (expired: %t)\n", peer.LoginExpirationEnabled, peer.LoginExpired)
	if peer.LastLogin != "" {
		fmt.Printf("  Last Login:  %s\n", peer.LastLogin)
	}
	if peer.Ephemeral {
		fmt.Printf("  Ephemeral:   true\n")
	}
	if peer.ApprovalRequired != nil {
		fmt.Printf("  Approval:    required=%t\n", *peer.ApprovalRequired)
		if peer.DisapprovalReason != "" {
			fmt.Printf("  Disapproval: %s\n", peer.DisapprovalReason)
		}
	}
	if peer.CreatedAt != "" {
		fmt.Printf("  Created At:  %s\n", peer.CreatedAt)
	}

	if len(peer.Groups) > 0 {
		fmt.Println("  Groups:")
		for _, group := range peer.Groups {
			fmt.Printf("    - %s (%s)\n", group.Name, group.ID)
		}
	} else {
		fmt.Println("  Groups:      None")
	}

	if peer.LocalFlags != nil {
		flags := peer.LocalFlags
		fmt.Println("  Local Flags:")
		fmt.Printf("    Rosenpass Enabled:     %t (permissive: %t)\n", flags.RosenpassEnabled, flags.RosenpassPermissive)
		fmt.Printf("    Server SSH Allowed:    %t\n", flags.ServerSSHAllowed)
		fmt.Printf("    Disable Client Routes: %t\n", flags.DisableClientRoutes)
		fmt.Printf("    Disable Server Routes: %t\n", flags.DisableServerRoutes)
		fmt.Printf("    Disable DNS:           %t\n", flags.DisableDNS)
		fmt.Printf("    Disable Firewall:      %t\n", flags.DisableFirewall)
		fmt.Printf("    Block LAN Access:      %t\n", flags.BlockLANAccess)
		fmt.Printf("    Block Inbound:         %t\n", flags.BlockInbound)
		fmt.Printf("    Lazy Connection:       %t\n", flags.LazyConnectionEnabled)
	}
	return nil
}

// createTemporaryAccessPeer creates a temporary access peer via POST /peers/{peerId}/temporary-access
func (s *Service) createTemporaryAccessPeer(peerID, name, wgPubKey string, rules []string, outputFormat string) error {
	req := models.TemporaryAccessRequest{
		Name:     name,
		WGPubKey: wgPubKey,
		Rules:    rules,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/peers/"+peerID+"/temporary-access", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tempPeer models.TemporaryAccessPeer
	if err := json.NewDecoder(resp.Body).Decode(&tempPeer); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(tempPeer, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	fmt.Println("Temporary access peer created successfully")
	fmt.Printf("  ID:    %s\n", tempPeer.ID)
	fmt.Printf("  Name:  %s\n", tempPeer.Name)
	fmt.Printf("  Rules: %s\n", strings.Join(tempPeer.Rules, ", "))
	return nil
}

func (s *Service) modifyPeerGroup(peerID, groupIdentifier, action string) error {
	if groupIdentifier == "" {
		fmt.Println("Error: No group identifier specified.")
		fmt.Println("Listing available groups:")
		if err := s.listGroups("", "table"); err != nil {
			fmt.Fprintf(os.Stderr, "Could not list groups: %v\n", err)
		}
		return fmt.Errorf("missing <group-id> or <group-name> argument for --add-group or --remove-group")
	}

	groupID, err := s.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	group, err := s.getGroupByID(groupID)
	if err != nil {
		return err
	}

	if _, err := s.getPeerByID(peerID); err != nil {
		return fmt.Errorf("failed to verify peer: %v", err)
	}

	var newPeerIDs []string
	peerFound := false
	for _, p := range group.Peers {
		if p.ID == peerID {
			peerFound = true
			if action == "remove" {
				continue
			}
		}
		newPeerIDs = append(newPeerIDs, p.ID)
	}

	if action == "add" && !peerFound {
		newPeerIDs = append(newPeerIDs, peerID)
	}

	if action == "add" && peerFound {
		fmt.Printf("Peer %s is already in group %s (%s).\n", peerID, group.Name, group.ID)
		return nil
	}
	if action == "remove" && !peerFound {
		fmt.Printf("Peer %s is not in group %s (%s).\n", peerID, group.Name, group.ID)
		return nil
	}

	var resources []models.GroupResourcePutRequest
	for _, r := range group.Resources {
		resources = append(resources, models.GroupResourcePutRequest{ID: r.ID, Type: r.Type})
	}

	reqBody := models.GroupPutRequest{
		Name:      group.Name,
		Peers:     newPeerIDs,
		Resources: resources,
	}

	if action == "add" {
		fmt.Printf("Adding peer %s to group %s (%s)...\n", peerID, group.Name, group.ID)
	} else {
		fmt.Printf("Removing peer %s from group %s (%s)...\n", peerID, group.Name, group.ID)
	}

	err = s.updateGroup(group.ID, reqBody)
	if err != nil {
		return fmt.Errorf("failed to update group: %v", err)
	}

	fmt.Println("Successfully updated group membership.")
	return nil
}

func (s *Service) updatePeer(peerID string, updates models.PeerUpdateRequest) error {
	if updates.IP != "" {
		if err := helpers.ValidateNetBirdIP(updates.IP); err != nil {
			return err
		}
	}

	payload, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %v", err)
	}

	endpoint := "/peers/" + peerID
	resp, err := s.Client.MakeRequest("PUT", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully updated peer %s\n", peerID)
	return nil
}

func (s *Service) getAccessiblePeers(peerID, outputFormat string) error {
	endpoint := "/peers/" + peerID + "/accessible-peers"
	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var accessiblePeers []models.Peer
	if err := json.NewDecoder(resp.Body).Decode(&accessiblePeers); err != nil {
		return fmt.Errorf("failed to decode accessible peers response: %v", err)
	}

	if len(accessiblePeers) == 0 {
		if outputFormat == "json" {
			fmt.Println("[]")
		} else {
			fmt.Println("This peer cannot access any other peers.")
		}
		return nil
	}

	// JSON output
	if outputFormat == "json" {
		output, err := json.MarshalIndent(accessiblePeers, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(output))
		return nil
	}

	// Table output (default)
	fmt.Printf("Peers accessible from %s:\n\n", peerID)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tIP\tCONNECTED\tOS\tHOSTNAME")
	fmt.Fprintln(w, "--\t----\t--\t---------\t--\t--------")

	for _, peer := range accessiblePeers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\t%s\n",
			peer.ID,
			peer.Name,
			peer.IP,
			peer.Connected,
			helpers.FormatOS(peer.OS),
			peer.Hostname,
		)
	}
	w.Flush()
	return nil
}
