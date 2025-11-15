// peers.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

// handlePeersCommand routes peer-related commands using the flag package
func handlePeersCommand(client *Client, args []string) error {
	// Create a new flag set for the 'peer' command
	peerCmd := flag.NewFlagSet("peer", flag.ContinueOnError)
	peerCmd.SetOutput(os.Stderr) // Send errors to stderr
	peerCmd.Usage = printPeerUsage // Set our custom usage function

	// Define the flags for the 'peer' command
	listFlag := peerCmd.Bool("list", false, "List all peers")
	inspectFlag := peerCmd.String("inspect", "", "Inspect a peer by its ID")
	removeFlag := peerCmd.String("remove", "", "Remove a peer by its ID")
	editFlag := peerCmd.String("edit", "", "Edit a peer by its ID (use with --add-group or --remove-group)")
	addGrpFlag := peerCmd.String("add-group", "", "Group to add to the peer (requires --edit)")
	rmGrpFlag := peerCmd.String("remove-group", "", "Group to remove from the peer (requires --edit)")

	// If no flags are provided (just 'netbird-manage peer'), show usage
	if len(args) == 1 {
		printPeerUsage()
		return nil
	}

	// Parse the flags (all args *after* 'peer')
	if err := peerCmd.Parse(args[1:]); err != nil {
		// The flag package will print an error, so we just return
		return nil
	}

	// Handle the flags
	if *listFlag {
		return client.listPeers()
	}

	if *inspectFlag != "" {
		return client.inspectPeer(*inspectFlag)
	}

	if *removeFlag != "" {
		return client.removePeerByID(*removeFlag)
	}

	if *editFlag != "" {
		peerID := *editFlag
		if *addGrpFlag != "" {
			return client.modifyPeerGroup(peerID, *addGrpFlag, "add")
		}
		if *rmGrpFlag != "" {
			return client.modifyPeerGroup(peerID, *rmGrpFlag, "remove")
		}
		return fmt.Errorf("flag --edit requires --add-group or --remove-group")
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'peer' command.")
	printPeerUsage()
	return nil
}

// listPeers implements the "peer --list" command
func (c *Client) listPeers() error {
	resp, err := c.makeRequest("GET", "/peers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peers []Peer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return fmt.Errorf("failed to decode peers response: %v", err)
	}

	if len(peers) == 0 {
		fmt.Println("No peers found in your network.")
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tIP\tCONNECTED\tOS\tVERSION\tHOSTNAME")
	fmt.Fprintln(w, "--\t----\t--\t---------\t--\t-------\t--------")

	for _, peer := range peers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\t%s\t%s\n",
			peer.ID,
			peer.Name,
			peer.IP,
			peer.Connected,
			formatOS(peer.OS),
			peer.Version,
			peer.Hostname,
		)
	}
	w.Flush()
	return nil
}

// getPeerByID finds a peer by its ID
func (c *Client) getPeerByID(peerID string) (*Peer, error) {
	endpoint := "/peers/" + peerID
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var peer Peer
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		return nil, fmt.Errorf("failed to decode peer response: %v", err)
	}
	return &peer, nil
}

// removePeerByID implements the "peer --remove <id>" command
func (c *Client) removePeerByID(peerID string) error {
	// First, let's confirm the peer exists to give a better error message
	peer, err := c.getPeerByID(peerID)
	if err != nil {
		return fmt.Errorf("cannot remove peer: %v", err)
	}

	fmt.Printf("Removing peer '%s' (ID: %s)...\n", peer.Name, peer.ID)
	endpoint := "/peers/" + peer.ID
	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	fmt.Printf("Successfully removed peer '%s' (ID: %s)\n", peer.Name, peer.ID)
	return nil
}

// inspectPeer implements the "peer --inspect <id>" command
func (c *Client) inspectPeer(peerID string) error {
	peer, err := c.getPeerByID(peerID)
	if err != nil {
		return err
	}

	fmt.Printf("Inspecting Peer: %s (%s)\n", peer.Name, peer.ID)
	fmt.Println("---------------------------------")
	fmt.Printf("  IP:          %s\n", peer.IP)
	fmt.Printf("  Hostname:    %s\n", peer.Hostname)
	fmt.Printf("  OS:          %s\n", formatOS(peer.OS))
	fmt.Printf("  Version:     %s\n", peer.Version)
	fmt.Printf("  Connected:   %t\n", peer.Connected)
	fmt.Printf("  Last Seen:   %s\n", peer.LastSeen)

	// List groups
	if len(peer.Groups) > 0 {
		fmt.Println("  Groups:")
		for _, group := range peer.Groups {
			fmt.Printf("    - %s (%s)\n", group.Name, group.ID)
		}
	} else {
		fmt.Println("  Groups:      None")
	}
	return nil
}

// modifyPeerGroup adds or removes a peer from a group.
// This is an "edit group" operation under the hood.
func (c *Client) modifyPeerGroup(peerID, groupID, action string) error {
	// If no group ID is provided, list available groups
	if groupID == "" {
		fmt.Println("Error: No group ID specified.")
		fmt.Println("Listing available groups:")
		if err := c.listGroups(); err != nil {
			fmt.Fprintf(os.Stderr, "Could not list groups: %v\n", err)
		}
		return fmt.Errorf("missing <group-id> argument for --add-group or --remove-group")
	}

	// 1. Get the Group's full details by ID
	group, err := c.getGroupByID(groupID)
	if err != nil {
		return err
	}

	// 2. Check if peer exists (and is valid)
	if _, err := c.getPeerByID(peerID); err != nil {
		return fmt.Errorf("failed to verify peer: %v", err)
	}

	// 3. Prepare the new list of peer IDs
	var newPeerIDs []string
	peerFound := false
	for _, p := range group.Peers {
		if p.ID == peerID {
			peerFound = true
			if action == "remove" {
				continue // Skip this peer to remove it
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

	// 4. Prepare the list of resources (must be included in the PUT request)
	var resources []GroupResourcePutRequest
	for _, r := range group.Resources {
		resources = append(resources, GroupResourcePutRequest{ID: r.ID, Type: r.Type})
	}

	// 5. Create the PUT request body
	reqBody := GroupPutRequest{
		Name:      group.Name,
		Peers:     newPeerIDs,
		Resources: resources,
	}

	// 6. Send the PUT request to update the group
	if action == "add" {
		fmt.Printf("Adding peer %s to group %s (%s)...\n", peerID, group.Name, group.ID)
	} else {
		fmt.Printf("Removing peer %s from group %s (%s)...\n", peerID, group.Name, group.ID)
	}

	err = c.updateGroup(group.ID, reqBody)
	if err != nil {
		return fmt.Errorf("failed to update group: %v", err)
	}

	fmt.Println("Successfully updated group membership.")
	return nil
}
