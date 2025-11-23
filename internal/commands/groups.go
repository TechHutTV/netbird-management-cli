package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"netbird-manage/internal/helpers"
	"netbird-manage/internal/models"
)

// HandleGroupsCommand routes group-related commands
func (s *Service) HandleGroupsCommand(args []string) error {
	groupCmd := flag.NewFlagSet("group", flag.ContinueOnError)
	groupCmd.SetOutput(os.Stderr)
	groupCmd.Usage = PrintGroupUsage

	listFlag := groupCmd.Bool("list", false, "List all groups")
	inspectFlag := groupCmd.String("inspect", "", "Inspect a group by its ID")
	filterNameFlag := groupCmd.String("filter-name", "", "Filter groups by name pattern (use with --list)")

	createFlag := groupCmd.String("create", "", "Create a new group")
	deleteFlag := groupCmd.String("delete", "", "Delete a group by its ID")
	deleteBatchFlag := groupCmd.String("delete-batch", "", "Delete multiple groups (comma-separated IDs)")
	renameFlag := groupCmd.String("rename", "", "Rename a group (requires --new-name)")
	newNameFlag := groupCmd.String("new-name", "", "New name for the group (requires --rename)")

	addPeersFlag := groupCmd.String("add-peers", "", "Add peers to a group (requires --peers)")
	removePeersFlag := groupCmd.String("remove-peers", "", "Remove peers from a group (requires --peers)")
	peersFlag := groupCmd.String("peers", "", "Comma-separated list of peer IDs")

	deleteUnusedFlag := groupCmd.Bool("delete-unused", false, "Delete all unused groups (not referenced anywhere)")

	if len(args) == 1 {
		PrintGroupUsage()
		return nil
	}

	if err := groupCmd.Parse(args[1:]); err != nil {
		return nil
	}

	if *listFlag {
		return s.listGroups(*filterNameFlag)
	}

	if *inspectFlag != "" {
		return s.inspectGroup(*inspectFlag)
	}

	if *createFlag != "" {
		var peerIDs []string
		if *peersFlag != "" {
			peerIDs = helpers.SplitCommaList(*peersFlag)
		}
		return s.createGroup(*createFlag, peerIDs)
	}

	if *deleteFlag != "" {
		return s.deleteGroup(*deleteFlag)
	}

	if *deleteBatchFlag != "" {
		return s.deleteGroupsBatch(*deleteBatchFlag)
	}

	if *renameFlag != "" {
		if *newNameFlag == "" {
			return fmt.Errorf("--new-name is required with --rename")
		}
		return s.renameGroup(*renameFlag, *newNameFlag)
	}

	if *addPeersFlag != "" {
		if *peersFlag == "" {
			return fmt.Errorf("--peers is required with --add-peers")
		}
		peerIDs := helpers.SplitCommaList(*peersFlag)
		return s.addPeersToGroup(*addPeersFlag, peerIDs)
	}

	if *removePeersFlag != "" {
		if *peersFlag == "" {
			return fmt.Errorf("--peers is required with --remove-peers")
		}
		peerIDs := helpers.SplitCommaList(*peersFlag)
		return s.removePeersFromGroup(*removePeersFlag, peerIDs)
	}

	if *deleteUnusedFlag {
		return s.deleteUnusedGroups()
	}

	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'group' command.")
	PrintGroupUsage()
	return nil
}

func (s *Service) listGroups(filterName string) error {
	resp, err := s.Client.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var groups []models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return fmt.Errorf("failed to decode groups response: %v", err)
	}

	var filteredGroups []models.GroupDetail
	for _, group := range groups {
		if filterName != "" && !helpers.MatchesPattern(group.Name, filterName) {
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

func (s *Service) getGroupByName(name string) (*models.GroupDetail, error) {
	resp, err := s.Client.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var groups []models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("failed to decode groups response: %v", err)
	}

	for _, group := range groups {
		if group.Name == name {
			return s.getGroupByID(group.ID)
		}
	}

	return nil, fmt.Errorf("no group found with name: %s", name)
}

func (s *Service) getGroupByID(id string) (*models.GroupDetail, error) {
	endpoint := "/groups/" + id
	resp, err := s.Client.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var group models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode group response: %v", err)
	}
	return &group, nil
}

func (s *Service) updateGroup(id string, reqBody models.GroupPutRequest) error {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal group update request: %v", err)
	}

	endpoint := "/groups/" + id
	resp, err := s.Client.MakeRequest("PUT", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *Service) inspectGroup(groupIdentifier string) error {
	groupID, err := s.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	group, err := s.getGroupByID(groupID)
	if err != nil {
		return err
	}

	fmt.Printf("Group: %s (%s)\n", group.Name, group.ID)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("  Peers Count:     %d\n", group.PeersCount)
	fmt.Printf("  Resources Count: %d\n", group.ResourcesCount)
	fmt.Printf("  Issued By:       %s\n", group.Issued)

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

func (s *Service) createGroup(name string, peerIDs []string) error {
	reqBody := models.GroupPutRequest{
		Name:      name,
		Peers:     peerIDs,
		Resources: []models.GroupResourcePutRequest{},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal create group request: %v", err)
	}

	resp, err := s.Client.MakeRequest("POST", "/groups", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var createdGroup models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&createdGroup); err != nil {
		return fmt.Errorf("failed to decode created group response: %v", err)
	}

	fmt.Printf("Successfully created group '%s' (ID: %s)\n", createdGroup.Name, createdGroup.ID)
	if len(peerIDs) > 0 {
		fmt.Printf("Added %d peer(s) to the group\n", len(peerIDs))
	}
	return nil
}

func (s *Service) deleteGroup(groupIdentifier string) error {
	groupID, err := s.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	group, err := s.getGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}

	details := map[string]string{
		"Peers":     fmt.Sprintf("%d", group.PeersCount),
		"Resources": fmt.Sprintf("%d", group.ResourcesCount),
	}

	if !helpers.ConfirmSingleDeletion("group", group.Name, group.ID, details) {
		return nil
	}

	fmt.Printf("Deleting group '%s' (ID: %s)...\n", group.Name, group.ID)

	endpoint := "/groups/" + groupID
	resp, err := s.Client.MakeRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Successfully deleted group '%s'\n", group.Name)
	return nil
}

func (s *Service) deleteGroupsBatch(idList string) error {
	groupIDs := helpers.SplitCommaList(idList)
	if len(groupIDs) == 0 {
		return fmt.Errorf("no group IDs provided")
	}

	groups := make([]*models.GroupDetail, 0, len(groupIDs))
	itemList := make([]string, 0, len(groupIDs))

	fmt.Println("Fetching group details...")
	for _, id := range groupIDs {
		resolvedID, err := s.resolveGroupIdentifier(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping %s: %v\n", id, err)
			continue
		}

		group, err := s.getGroupByID(resolvedID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Skipping %s: %v\n", id, err)
			continue
		}
		groups = append(groups, group)
		itemList = append(itemList, fmt.Sprintf("%s (ID: %s, Peers: %d, Resources: %d)",
			group.Name, group.ID, group.PeersCount, group.ResourcesCount))
	}

	if len(groups) == 0 {
		return fmt.Errorf("no valid groups found to delete")
	}

	if !helpers.ConfirmBulkDeletion("groups", itemList, len(groups)) {
		return nil
	}

	var succeeded, failed int
	for i, group := range groups {
		fmt.Printf("[%d/%d] Deleting group '%s'... ", i+1, len(groups), group.Name)

		endpoint := "/groups/" + group.ID
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
		fmt.Printf("All %d groups deleted successfully\n", succeeded)
	}

	return nil
}

func (s *Service) resolveGroupIdentifier(identifier string) (string, error) {
	group, err := s.getGroupByID(identifier)
	if err == nil {
		return group.ID, nil
	}

	group, err = s.getGroupByName(identifier)
	if err != nil {
		return "", fmt.Errorf("group '%s' not found (tried as both ID and name)", identifier)
	}

	return group.ID, nil
}

func (s *Service) resolveMultipleGroupIdentifiers(identifiers []string) ([]string, error) {
	if len(identifiers) == 0 {
		return []string{}, nil
	}

	resolvedIDs := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		if identifier == "" {
			continue
		}
		id, err := s.resolveGroupIdentifier(identifier)
		if err != nil {
			return nil, err
		}
		resolvedIDs = append(resolvedIDs, id)
	}
	return resolvedIDs, nil
}

func (s *Service) renameGroup(groupIdentifier, newName string) error {
	groupID, err := s.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	group, err := s.getGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}

	oldName := group.Name

	var peerIDs []string
	for _, peer := range group.Peers {
		peerIDs = append(peerIDs, peer.ID)
	}

	var resources []models.GroupResourcePutRequest
	for _, r := range group.Resources {
		resources = append(resources, models.GroupResourcePutRequest{ID: r.ID, Type: r.Type})
	}

	reqBody := models.GroupPutRequest{
		Name:      newName,
		Peers:     peerIDs,
		Resources: resources,
	}

	fmt.Printf("Renaming group '%s' to '%s'...\n", oldName, newName)

	if err := s.updateGroup(groupID, reqBody); err != nil {
		return fmt.Errorf("failed to rename group: %v", err)
	}

	fmt.Printf("Successfully renamed group from '%s' to '%s'\n", oldName, newName)
	return nil
}

func (s *Service) addPeersToGroup(groupIdentifier string, peerIDs []string) error {
	groupID, err := s.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	group, err := s.getGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}

	newPeerIDs := make([]string, 0, len(group.Peers)+len(peerIDs))
	existingPeerMap := make(map[string]bool, len(group.Peers))

	for _, peer := range group.Peers {
		newPeerIDs = append(newPeerIDs, peer.ID)
		existingPeerMap[peer.ID] = true
	}

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

	var resources []models.GroupResourcePutRequest
	for _, r := range group.Resources {
		resources = append(resources, models.GroupResourcePutRequest{ID: r.ID, Type: r.Type})
	}

	reqBody := models.GroupPutRequest{
		Name:      group.Name,
		Peers:     newPeerIDs,
		Resources: resources,
	}

	fmt.Printf("Adding %d peer(s) to group '%s'...\n", addedCount, group.Name)

	if err := s.updateGroup(groupID, reqBody); err != nil {
		return fmt.Errorf("failed to add peers: %v", err)
	}

	fmt.Printf("Successfully added %d peer(s) to group '%s'\n", addedCount, group.Name)
	return nil
}

func (s *Service) removePeersFromGroup(groupIdentifier string, peerIDs []string) error {
	groupID, err := s.resolveGroupIdentifier(groupIdentifier)
	if err != nil {
		return err
	}

	group, err := s.getGroupByID(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}

	removeMap := make(map[string]bool, len(peerIDs))
	for _, peerID := range peerIDs {
		removeMap[peerID] = true
	}

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

	var resources []models.GroupResourcePutRequest
	for _, r := range group.Resources {
		resources = append(resources, models.GroupResourcePutRequest{ID: r.ID, Type: r.Type})
	}

	reqBody := models.GroupPutRequest{
		Name:      group.Name,
		Peers:     newPeerIDs,
		Resources: resources,
	}

	fmt.Printf("Removing %d peer(s) from group '%s'...\n", removedCount, group.Name)

	if err := s.updateGroup(groupID, reqBody); err != nil {
		return fmt.Errorf("failed to remove peers: %v", err)
	}

	fmt.Printf("Successfully removed %d peer(s) from group '%s'\n", removedCount, group.Name)
	return nil
}

func (s *Service) deleteUnusedGroups() error {
	fmt.Println("Scanning for unused groups...")

	resp, err := s.Client.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var groups []models.GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return fmt.Errorf("failed to decode groups: %v", err)
	}

	if len(groups) == 0 {
		fmt.Println("No groups found.")
		return nil
	}

	policies, setupKeys, routes, dnsGroups, users, err := s.getAllGroupDependencies()
	if err != nil {
		return fmt.Errorf("failed to get dependencies: %v", err)
	}

	referencedGroups := make(map[string]bool)

	for _, policy := range policies {
		for _, rule := range policy.Rules {
			for _, src := range rule.Sources {
				referencedGroups[src.ID] = true
			}
			for _, dest := range rule.Destinations {
				referencedGroups[dest.ID] = true
			}
		}
	}

	for _, key := range setupKeys {
		for _, groupID := range key.AutoGroups {
			referencedGroups[groupID] = true
		}
	}

	for _, route := range routes {
		for _, groupID := range route.Groups {
			referencedGroups[groupID] = true
		}
	}

	for _, dnsGroup := range dnsGroups {
		for _, groupID := range dnsGroup.Groups {
			referencedGroups[groupID] = true
		}
	}

	for _, user := range users {
		for _, groupID := range user.AutoGroups {
			referencedGroups[groupID] = true
		}
	}

	var unusedGroups []models.GroupDetail
	for _, group := range groups {
		if group.PeersCount == 0 && group.ResourcesCount == 0 && !referencedGroups[group.ID] {
			unusedGroups = append(unusedGroups, group)
		}
	}

	if len(unusedGroups) == 0 {
		fmt.Println("No unused groups found. All groups are in use.")
		return nil
	}

	groupList := make([]string, len(unusedGroups))
	for i, group := range unusedGroups {
		groupList[i] = fmt.Sprintf("%s (ID: %s)", group.Name, group.ID)
	}

	if !helpers.ConfirmBulkDeletion("groups", groupList, len(unusedGroups)) {
		return nil
	}

	fmt.Printf("\nDeleting %d group(s)...\n", len(unusedGroups))
	successCount := 0
	failCount := 0

	for _, group := range unusedGroups {
		endpoint := "/groups/" + group.ID
		resp, err := s.Client.MakeRequest("DELETE", endpoint, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete '%s' (%s): %v\n", group.Name, group.ID, err)
			failCount++
			continue
		}
		resp.Body.Close()

		fmt.Printf("Deleted '%s' (%s)\n", group.Name, group.ID)
		successCount++
	}

	fmt.Printf("\nDeletion complete: %d successful, %d failed\n", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("failed to delete %d group(s)", failCount)
	}

	return nil
}

func (s *Service) getAllGroupDependencies() ([]models.Policy, []models.SetupKey, []models.Route, []models.DNSNameserverGroup, []models.User, error) {
	var policies []models.Policy
	var setupKeys []models.SetupKey
	var routes []models.Route
	var dnsGroups []models.DNSNameserverGroup
	var users []models.User

	resp, err := s.Client.MakeRequest("GET", "/policies", nil)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get policies: %v", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		resp.Body.Close()
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to decode policies: %v", err)
	}
	resp.Body.Close()

	resp, err = s.Client.MakeRequest("GET", "/setup-keys", nil)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get setup keys: %v", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&setupKeys); err != nil {
		resp.Body.Close()
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to decode setup keys: %v", err)
	}
	resp.Body.Close()

	resp, err = s.Client.MakeRequest("GET", "/routes", nil)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get routes: %v", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		resp.Body.Close()
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to decode routes: %v", err)
	}
	resp.Body.Close()

	resp, err = s.Client.MakeRequest("GET", "/dns/nameservers", nil)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get DNS groups: %v", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&dnsGroups); err != nil {
		resp.Body.Close()
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to decode DNS groups: %v", err)
	}
	resp.Body.Close()

	resp, err = s.Client.MakeRequest("GET", "/users", nil)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to get users: %v", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		resp.Body.Close()
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to decode users: %v", err)
	}
	resp.Body.Close()

	return policies, setupKeys, routes, dnsGroups, users, nil
}
