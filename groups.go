// groups.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

// handleGroupsCommand routes group-related commands
func handleGroupsCommand(client *Client, args []string) error {
	// Create a new flag set for the 'group' command
	groupCmd := flag.NewFlagSet("group", flag.ContinueOnError)
	groupCmd.SetOutput(os.Stderr)    // Send errors to stderr
	groupCmd.Usage = printGroupUsage // Set our custom usage function

	// Query flags
	listFlag := groupCmd.Bool("list", false, "List all groups")
	inspectFlag := groupCmd.String("inspect", "", "Inspect a group by its ID")
	filterNameFlag := groupCmd.String("filter-name", "", "Filter groups by name pattern (use with --list)")

	// Modification flags
	createFlag := groupCmd.String("create", "", "Create a new group")
	deleteFlag := groupCmd.String("delete", "", "Delete a group by its ID")
	renameFlag := groupCmd.String("rename", "", "Rename a group (requires --new-name)")
	newNameFlag := groupCmd.String("new-name", "", "New name for the group (requires --rename)")

	// Bulk peer management flags
	addPeersFlag := groupCmd.String("add-peers", "", "Add peers to a group (requires --peers)")
	removePeersFlag := groupCmd.String("remove-peers", "", "Remove peers from a group (requires --peers)")
	peersFlag := groupCmd.String("peers", "", "Comma-separated list of peer IDs")

	// If no flags are provided (just 'netbird-manage group'), show usage
	if len(args) == 1 {
		printGroupUsage()
		return nil
	}

	// Parse the flags (all args *after* 'group')
	if err := groupCmd.Parse(args[1:]); err != nil {
		// The flag package will print an error, so we just return
		return nil
	}

	// Handle the flags
	if *listFlag {
		return client.listGroups(*filterNameFlag)
	}

	if *inspectFlag != "" {
		return client.inspectGroup(*inspectFlag)
	}

	if *createFlag != "" {
		// Parse peer IDs if provided
		var peerIDs []string
		if *peersFlag != "" {
			peerIDs = splitCommaList(*peersFlag)
		}
		return client.createGroup(*createFlag, peerIDs)
	}

	if *deleteFlag != "" {
		return client.deleteGroup(*deleteFlag)
	}

	if *renameFlag != "" {
		if *newNameFlag == "" {
			return fmt.Errorf("--new-name is required with --rename")
		}
		return client.renameGroup(*renameFlag, *newNameFlag)
	}

	if *addPeersFlag != "" {
		if *peersFlag == "" {
			return fmt.Errorf("--peers is required with --add-peers")
		}
		peerIDs := splitCommaList(*peersFlag)
		return client.addPeersToGroup(*addPeersFlag, peerIDs)
	}

	if *removePeersFlag != "" {
		if *peersFlag == "" {
			return fmt.Errorf("--peers is required with --remove-peers")
		}
		peerIDs := splitCommaList(*peersFlag)
		return client.removePeersFromGroup(*removePeersFlag, peerIDs)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'group' command.")
	printGroupUsage()
	return nil
}

// listGroups implements the "group" command with optional filtering
func (c *Client) listGroups(filterName string) error {
	resp, err := c.makeRequest("GET", "/groups", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var groups []GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return fmt.Errorf("failed to decode groups response: %v", err)
	}

	// Apply filter if provided
	var filteredGroups []GroupDetail
	for _, group := range groups {
		if filterName != "" && !matchesPattern(group.Name, filterName) {
			continue
		}
		filteredGroups = append(filteredGroups, group)
	}

	if len(filteredGroups) == 0 {
		if filterName != "" {
			fmt.Println("No groups found matching the specified filter.")
		} else {
			fmt.Println("No groups found.")
		}
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPEERS\tRESOURCES\tISSUED BY")
	fmt.Fprintln(w, "--\t----\t-----\t---------\t---------")

	for _, g := range filteredGroups {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
			g.ID,
			g.Name,
			g.PeersCount,
			g.ResourcesCount,
			g.Issued,
		)
	}
	w.Flush()
	return nil
}

// getGroupByName finds a group by its name
func (c *Client) getGroupByName(name string) (*GroupDetail, error) {
	resp, err := c.makeRequest("GET", "/groups", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("failed to decode groups response: %v", err)
	}

	for _, group := range groups {
		if group.Name == name {
			// Now we need the full group details, which includes the list of peers.
			// The list view might not be enough, so we fetch the specific group.
			return c.getGroupByID(group.ID)
		}
	}

	return nil, fmt.Errorf("no group found with name: %s", name)
}

// getGroupByID finds a group by its ID
func (c *Client) getGroupByID(id string) (*GroupDetail, error) {
	endpoint := "/groups/" + id
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var group GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode group response: %v", err)
	}
	return &group, nil
}

// updateGroup sends a PUT request to update a group
func (c *Client) updateGroup(id string, reqBody GroupPutRequest) error {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal group update request: %v", err)
	}

	endpoint := "/groups/" + id
	resp, err := c.makeRequest("PUT", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// inspectGroup displays detailed information about a specific group
func (c *Client) inspectGroup(groupID string) error {
	group, err := c.getGroupByID(groupID)
	if err != nil {
		return err
	}

	fmt.Printf("Group: %s (%s)\n", group.Name, group.ID)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  Peers Count:     %d\n", group.PeersCount)
	fmt.Printf("  Resources Count: %d\n", group.ResourcesCount)
	fmt.Printf("  Issued By:       %s\n", group.Issued)

	// List peers in the group
	if len(group.Peers) > 0 {
		fmt.Println("\n  Peers:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "    ID\tNAME\tIP\tCONNECTED")
		fmt.Fprintln(w, "    --\t----\t--\t---------")
		for _, peer := range group.Peers {
			fmt.Fprintf(w, "    %s\t%s\t%s\t%t\n",
				peer.ID,
				peer.Name,
				peer.IP,
				peer.Connected,
			)
		}
		w.Flush()
	} else {
		fmt.Println("\n  Peers:           None")
	}

	// List resources
	if len(group.Resources) > 0 {
		fmt.Println("\n  Resources:")
		for _, resource := range group.Resources {
			fmt.Printf("    - %s (Type: %s)\n", resource.ID, resource.Type)
		}
	} else {
		fmt.Println("\n  Resources:       None")
	}

	return nil
}

// createGroup creates a new group
func (c *Client) createGroup(name string, peerIDs []string) error {
	reqBody := GroupPutRequest{
		Name:      name,
		Peers:     peerIDs,
		Resources: []GroupResourcePutRequest{},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal create group request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/groups", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var createdGroup GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&createdGroup); err != nil {
		return fmt.Errorf("failed to decode created group response: %v", err)
	}

	fmt.Printf("Successfully created group '%s' (ID: %s)\n", createdGroup.Name, createdGroup.ID)
	if len(peerIDs) > 0 {
		fmt.Printf("Added %d peer(s) to the group\n", len(peerIDs))
	}
	return nil
}

// deleteGroup deletes a group by ID
func (c *Client) deleteGroup(groupID string) error {
	// First, get group details to show what's being deleted
	group, err := c.getGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}

	fmt.Printf("Deleting group '%s' (ID: %s)...\n", group.Name, group.ID)
	if group.PeersCount > 0 || group.ResourcesCount > 0 {
		fmt.Printf("Warning: This group has %d peer(s) and %d resource(s)\n", group.PeersCount, group.ResourcesCount)
	}

	endpoint := "/groups/" + groupID
	resp, err := c.makeRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully deleted group '%s'\n", group.Name)
	return nil
}

// resolveGroupIdentifier resolves a group name or ID to an ID
func (c *Client) resolveGroupIdentifier(identifier string) (string, error) {
	// First, try to get it as an ID
	group, err := c.getGroupByID(identifier)
	if err == nil {
		return group.ID, nil
	}

	// If that fails, try to find by name
	group, err = c.getGroupByName(identifier)
	if err != nil {
		return "", fmt.Errorf("group '%s' not found (tried as both ID and name)", identifier)
	}

	return group.ID, nil
}

// renameGroup renames an existing group (accepts group ID or name)
func (c *Client) renameGroup(groupIdentifier, newName string) error {
	// Resolve group identifier to ID
	groupID, err := c.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	// Get current group state
	group, err := c.getGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}

	oldName := group.Name

	// Prepare updated group with new name
	var peerIDs []string
	for _, peer := range group.Peers {
		peerIDs = append(peerIDs, peer.ID)
	}

	var resources []GroupResourcePutRequest
	for _, r := range group.Resources {
		resources = append(resources, GroupResourcePutRequest{ID: r.ID, Type: r.Type})
	}

	reqBody := GroupPutRequest{
		Name:      newName,
		Peers:     peerIDs,
		Resources: resources,
	}

	fmt.Printf("Renaming group '%s' to '%s'...\n", oldName, newName)

	if err := c.updateGroup(groupID, reqBody); err != nil {
		return fmt.Errorf("failed to rename group: %v", err)
	}

	fmt.Printf("Successfully renamed group from '%s' to '%s'\n", oldName, newName)
	return nil
}

// addPeersToGroup adds multiple peers to a group at once (accepts group ID or name)
func (c *Client) addPeersToGroup(groupIdentifier string, peerIDs []string) error {
	// Resolve group identifier to ID
	groupID, err := c.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	// Get current group state
	group, err := c.getGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}

	// Build new peer list (existing + new)
	newPeerIDs := make([]string, 0, len(group.Peers)+len(peerIDs))
	existingPeerMap := make(map[string]bool, len(group.Peers))

	// Add existing peers
	for _, peer := range group.Peers {
		newPeerIDs = append(newPeerIDs, peer.ID)
		existingPeerMap[peer.ID] = true
	}

	// Add new peers (skip duplicates)
	addedCount := 0
	for _, peerID := range peerIDs {
		if !existingPeerMap[peerID] {
			newPeerIDs = append(newPeerIDs, peerID)
			addedCount++
		}
	}

	if addedCount == 0 {
		fmt.Println("All specified peers are already in the group")
		return nil
	}

	// Prepare resources
	var resources []GroupResourcePutRequest
	for _, r := range group.Resources {
		resources = append(resources, GroupResourcePutRequest{ID: r.ID, Type: r.Type})
	}

	reqBody := GroupPutRequest{
		Name:      group.Name,
		Peers:     newPeerIDs,
		Resources: resources,
	}

	fmt.Printf("Adding %d peer(s) to group '%s'...\n", addedCount, group.Name)

	if err := c.updateGroup(groupID, reqBody); err != nil {
		return fmt.Errorf("failed to add peers: %v", err)
	}

	fmt.Printf("Successfully added %d peer(s) to group '%s'\n", addedCount, group.Name)
	return nil
}

// removePeersFromGroup removes multiple peers from a group at once (accepts group ID or name)
func (c *Client) removePeersFromGroup(groupIdentifier string, peerIDs []string) error {
	// Resolve group identifier to ID
	groupID, err := c.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	// Get current group state
	group, err := c.getGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}

	// Build map of peers to remove for efficient lookup
	removeMap := make(map[string]bool, len(peerIDs))
	for _, peerID := range peerIDs {
		removeMap[peerID] = true
	}

	// Build new peer list (exclude removed peers)
	newPeerIDs := make([]string, 0, len(group.Peers))
	removedCount := 0

	for _, peer := range group.Peers {
		if removeMap[peer.ID] {
			removedCount++
		} else {
			newPeerIDs = append(newPeerIDs, peer.ID)
		}
	}

	if removedCount == 0 {
		fmt.Println("None of the specified peers are in the group")
		return nil
	}

	// Prepare resources
	var resources []GroupResourcePutRequest
	for _, r := range group.Resources {
		resources = append(resources, GroupResourcePutRequest{ID: r.ID, Type: r.Type})
	}

	reqBody := GroupPutRequest{
		Name:      group.Name,
		Peers:     newPeerIDs,
		Resources: resources,
	}

	fmt.Printf("Removing %d peer(s) from group '%s'...\n", removedCount, group.Name)

	if err := c.updateGroup(groupID, reqBody); err != nil {
		return fmt.Errorf("failed to remove peers: %v", err)
	}

	fmt.Printf("Successfully removed %d peer(s) from group '%s'\n", removedCount, group.Name)
	return nil
}
