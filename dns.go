// dns.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// handleDNSCommand routes DNS-related commands
func handleDNSCommand(client *Client, args []string) error {
	dnsCmd := flag.NewFlagSet("dns", flag.ContinueOnError)
	dnsCmd.SetOutput(os.Stderr)
	dnsCmd.Usage = printDNSUsage

	// Query flags
	listFlag := dnsCmd.Bool("list", false, "List all DNS nameserver groups")
	inspectFlag := dnsCmd.String("inspect", "", "Inspect a DNS group by ID")
	filterName := dnsCmd.String("filter-name", "", "Filter by name pattern")
	primaryOnlyFlag := dnsCmd.Bool("primary-only", false, "Show only primary groups")
	enabledOnlyFlag := dnsCmd.Bool("enabled-only", false, "Show only enabled groups")
	getSettingsFlag := dnsCmd.Bool("get-settings", false, "Get DNS settings for the account")

	// Create flags
	createFlag := dnsCmd.String("create", "", "Create a new DNS nameserver group with the given name")
	nameserversFlag := dnsCmd.String("nameservers", "", "Nameservers (e.g., 8.8.8.8:53,1.1.1.1:53)")
	groupsFlag := dnsCmd.String("groups", "", "Target group IDs (comma-separated, required)")
	domainsFlag := dnsCmd.String("domains", "", "Match domains (comma-separated)")
	descriptionFlag := dnsCmd.String("description", "", "Description")
	searchDomainsFlag := dnsCmd.Bool("search-domains", false, "Enable search domains")
	primaryFlag := dnsCmd.Bool("primary", false, "Set as primary DNS")
	enabledFlag := dnsCmd.Bool("enabled", true, "Enable group (default)")
	disabledFlag := dnsCmd.Bool("disabled", false, "Disable group")

	// Update flags
	updateFlag := dnsCmd.String("update", "", "Update a DNS nameserver group by ID")

	// Delete flags
	deleteFlag := dnsCmd.String("delete", "", "Delete a DNS group by ID")

	// Toggle flags
	enableFlag := dnsCmd.String("enable", "", "Enable a DNS group by ID")
	disableFlag := dnsCmd.String("disable", "", "Disable a DNS group by ID")

	// Settings flags
	updateSettingsFlag := dnsCmd.Bool("update-settings", false, "Update DNS settings")
	disabledGroupsFlag := dnsCmd.String("disabled-groups", "", "Groups with disabled DNS management (comma-separated)")

	// If no flags provided, show usage
	if len(args) == 1 {
		printDNSUsage()
		return nil
	}

	// Parse the flags
	if err := dnsCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags in priority order

	// Get settings
	if *getSettingsFlag {
		return client.getDNSSettings()
	}

	// Update settings
	if *updateSettingsFlag {
		if *disabledGroupsFlag == "" {
			return fmt.Errorf("--disabled-groups is required when updating settings")
		}
		return client.updateDNSSettings(*disabledGroupsFlag)
	}

	// Create DNS group
	if *createFlag != "" {
		if *nameserversFlag == "" {
			return fmt.Errorf("--nameservers is required when creating a DNS group")
		}
		if *groupsFlag == "" {
			return fmt.Errorf("--groups is required when creating a DNS group")
		}

		enabled := *enabledFlag
		if *disabledFlag {
			enabled = false
		}

		return client.createDNSGroup(*createFlag, *nameserversFlag, *groupsFlag, *domainsFlag, *descriptionFlag, *searchDomainsFlag, *primaryFlag, enabled)
	}

	// Delete DNS group
	if *deleteFlag != "" {
		return client.deleteDNSGroup(*deleteFlag)
	}

	// Enable DNS group
	if *enableFlag != "" {
		return client.toggleDNSGroup(*enableFlag, true)
	}

	// Disable DNS group
	if *disableFlag != "" {
		return client.toggleDNSGroup(*disableFlag, false)
	}

	// Update DNS group
	if *updateFlag != "" {
		enabled := *enabledFlag
		if *disabledFlag {
			enabled = false
		}

		return client.updateDNSGroup(*updateFlag, *nameserversFlag, *groupsFlag, *domainsFlag, *descriptionFlag, *searchDomainsFlag, *primaryFlag, enabled)
	}

	// Inspect DNS group
	if *inspectFlag != "" {
		return client.inspectDNSGroup(*inspectFlag)
	}

	// List DNS groups
	if *listFlag {
		filters := &DNSFilters{
			NamePattern:  *filterName,
			PrimaryOnly:  *primaryOnlyFlag,
			EnabledOnly:  *enabledOnlyFlag,
		}
		return client.listDNSGroups(filters)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'dns' command.")
	printDNSUsage()
	return nil
}

// DNSFilters holds filtering options for listing DNS groups
type DNSFilters struct {
	NamePattern string
	PrimaryOnly bool
	EnabledOnly bool
}

// listDNSGroups implements the "dns --list" command
func (c *Client) listDNSGroups(filters *DNSFilters) error {
	resp, err := c.makeRequest("GET", "/dns/nameservers", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var groups []DNSNameserverGroup
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return fmt.Errorf("failed to decode DNS groups response: %v", err)
	}

	// Apply filters
	var filtered []DNSNameserverGroup
	for _, group := range groups {
		// Filter by name pattern
		if filters.NamePattern != "" && !matchesPattern(group.Name, filters.NamePattern) {
			continue
		}

		// Filter by primary
		if filters.PrimaryOnly && !group.Primary {
			continue
		}

		// Filter by enabled
		if filters.EnabledOnly && !group.Enabled {
			continue
		}

		filtered = append(filtered, group)
	}

	if len(filtered) == 0 {
		fmt.Println("No DNS nameserver groups found.")
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tNAMESERVERS\tGROUPS\tDOMAINS\tPRIMARY\tENABLED")
	fmt.Fprintln(w, "--\t----\t-----------\t------\t-------\t-------\t-------")

	for _, group := range filtered {
		nsCount := fmt.Sprintf("%d servers", len(group.Nameservers))
		groupCount := fmt.Sprintf("%d groups", len(group.Groups))
		domainCount := "-"
		if len(group.Domains) > 0 {
			domainCount = fmt.Sprintf("%d domains", len(group.Domains))
		}

		primaryStr := "No"
		if group.Primary {
			primaryStr = "Yes"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%t\n",
			group.ID,
			group.Name,
			nsCount,
			groupCount,
			domainCount,
			primaryStr,
			group.Enabled,
		)
	}

	w.Flush()
	fmt.Printf("\nTotal: %d DNS nameserver groups\n", len(filtered))
	return nil
}

// inspectDNSGroup implements the "dns --inspect" command
func (c *Client) inspectDNSGroup(groupID string) error {
	resp, err := c.makeRequest("GET", "/dns/nameservers/"+groupID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var group DNSNameserverGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return fmt.Errorf("failed to decode DNS group response: %v", err)
	}

	// Print detailed DNS group information
	fmt.Println("DNS Nameserver Group Details:")
	fmt.Println("=============================")
	fmt.Printf("ID:                    %s\n", group.ID)
	fmt.Printf("Name:                  %s\n", group.Name)
	fmt.Printf("Primary:               %t\n", group.Primary)
	fmt.Printf("Enabled:               %t\n", group.Enabled)
	fmt.Printf("Search Domains:        %t\n", group.SearchDomainsEnabled)

	if group.Description != "" {
		fmt.Printf("Description:           %s\n", group.Description)
	}

	fmt.Println()
	fmt.Println("Nameservers:")
	fmt.Println("------------")
	if len(group.Nameservers) > 0 {
		for i, ns := range group.Nameservers {
			fmt.Printf("  [%d] %s:%d (%s)\n", i+1, ns.IP, ns.Port, ns.NSType)
		}
	} else {
		fmt.Println("  None")
	}

	fmt.Println()
	fmt.Println("Target Groups:")
	fmt.Println("--------------")
	if len(group.Groups) > 0 {
		for _, groupID := range group.Groups {
			fmt.Printf("  - %s\n", groupID)
		}
	} else {
		fmt.Println("  None")
	}

	fmt.Println()
	fmt.Println("Match Domains:")
	fmt.Println("--------------")
	if len(group.Domains) > 0 {
		for _, domain := range group.Domains {
			fmt.Printf("  - %s\n", domain)
		}
	} else {
		fmt.Println("  All domains")
	}

	return nil
}

// createDNSGroup implements the "dns --create" command
func (c *Client) createDNSGroup(name, nameservers, groups, domains, description string, searchDomains, primary, enabled bool) error {
	// Parse nameservers
	nsList, err := parseNameservers(nameservers)
	if err != nil {
		return err
	}

	// Parse groups
	groupList := splitCommaList(groups)
	if len(groupList) == 0 {
		return fmt.Errorf("at least one group is required")
	}

	// Parse domains (optional)
	var domainList []string
	if domains != "" {
		domainList = splitCommaList(domains)
	}

	reqBody := DNSNameserverGroupRequest{
		Name:                 name,
		Description:          description,
		Nameservers:          nsList,
		Groups:               groupList,
		Domains:              domainList,
		SearchDomainsEnabled: searchDomains,
		Primary:              primary,
		Enabled:              enabled,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/dns/nameservers", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var createdGroup DNSNameserverGroup
	if err := json.NewDecoder(resp.Body).Decode(&createdGroup); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("✓ DNS nameserver group created successfully!\n")
	fmt.Printf("  ID:          %s\n", createdGroup.ID)
	fmt.Printf("  Name:        %s\n", createdGroup.Name)
	fmt.Printf("  Nameservers: %d\n", len(createdGroup.Nameservers))
	fmt.Printf("  Primary:     %t\n", createdGroup.Primary)
	fmt.Printf("  Enabled:     %t\n", createdGroup.Enabled)
	return nil
}

// updateDNSGroup implements the "dns --update" command
func (c *Client) updateDNSGroup(groupID, nameservers, groups, domains, description string, searchDomains, primary, enabled bool) error {
	// First, get the current group
	resp, err := c.makeRequest("GET", "/dns/nameservers/"+groupID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var currentGroup DNSNameserverGroup
	if err := json.NewDecoder(resp.Body).Decode(&currentGroup); err != nil {
		return fmt.Errorf("failed to decode current DNS group: %v", err)
	}

	// Build update request (update only provided fields)
	updateReq := DNSNameserverGroupRequest{
		Name:                 currentGroup.Name,
		Description:          currentGroup.Description,
		Nameservers:          currentGroup.Nameservers,
		Groups:               currentGroup.Groups,
		Domains:              currentGroup.Domains,
		SearchDomainsEnabled: searchDomains,
		Primary:              primary,
		Enabled:              enabled,
	}

	// Update fields if provided
	if description != "" {
		updateReq.Description = description
	}
	if nameservers != "" {
		nsList, err := parseNameservers(nameservers)
		if err != nil {
			return err
		}
		updateReq.Nameservers = nsList
	}
	if groups != "" {
		updateReq.Groups = splitCommaList(groups)
	}
	if domains != "" {
		updateReq.Domains = splitCommaList(domains)
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/dns/nameservers/"+groupID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ DNS nameserver group %s updated successfully\n", groupID)
	return nil
}

// deleteDNSGroup implements the "dns --delete" command
func (c *Client) deleteDNSGroup(groupID string) error {
	resp, err := c.makeRequest("DELETE", "/dns/nameservers/"+groupID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ DNS nameserver group %s deleted successfully\n", groupID)
	return nil
}

// toggleDNSGroup enables or disables a DNS group
func (c *Client) toggleDNSGroup(groupID string, enable bool) error {
	// First, get the current group
	resp, err := c.makeRequest("GET", "/dns/nameservers/"+groupID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var group DNSNameserverGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return fmt.Errorf("failed to decode DNS group: %v", err)
	}

	// Update the enabled status
	updateReq := DNSNameserverGroupRequest{
		Name:                 group.Name,
		Description:          group.Description,
		Nameservers:          group.Nameservers,
		Groups:               group.Groups,
		Domains:              group.Domains,
		SearchDomainsEnabled: group.SearchDomainsEnabled,
		Primary:              group.Primary,
		Enabled:              enable,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err = c.makeRequest("PUT", "/dns/nameservers/"+groupID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	status := "enabled"
	if !enable {
		status = "disabled"
	}
	fmt.Printf("✓ DNS nameserver group '%s' %s successfully\n", group.Name, status)
	return nil
}

// getDNSSettings implements the "dns --get-settings" command
func (c *Client) getDNSSettings() error {
	resp, err := c.makeRequest("GET", "/dns/settings", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var settings DNSSettings
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return fmt.Errorf("failed to decode DNS settings response: %v", err)
	}

	fmt.Println("DNS Settings:")
	fmt.Println("=============")
	fmt.Println()
	fmt.Println("Disabled Management Groups:")
	fmt.Println("---------------------------")
	if len(settings.DisabledManagementGroups) > 0 {
		for _, groupID := range settings.DisabledManagementGroups {
			fmt.Printf("  - %s\n", groupID)
		}
	} else {
		fmt.Println("  None (DNS management enabled for all groups)")
	}

	return nil
}

// updateDNSSettings implements the "dns --update-settings" command
func (c *Client) updateDNSSettings(disabledGroups string) error {
	// Parse disabled groups
	groupList := splitCommaList(disabledGroups)

	reqBody := DNSSettings{
		DisabledManagementGroups: groupList,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("PUT", "/dns/settings", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("✓ DNS settings updated successfully\n")
	if len(groupList) > 0 {
		fmt.Printf("  Disabled management for %d group(s)\n", len(groupList))
	} else {
		fmt.Printf("  DNS management enabled for all groups\n")
	}
	return nil
}

// parseNameservers parses a comma-separated list of nameservers
// Format: "8.8.8.8:53,1.1.1.1:53" or "8.8.8.8,1.1.1.1" (default port 53)
func parseNameservers(nameservers string) ([]Nameserver, error) {
	parts := strings.Split(nameservers, ",")
	var nsList []Nameserver

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		ns, err := parseNameserver(part)
		if err != nil {
			return nil, err
		}
		nsList = append(nsList, ns)
	}

	if len(nsList) == 0 {
		return nil, fmt.Errorf("at least one nameserver is required")
	}

	return nsList, nil
}

// parseNameserver parses a single nameserver string
// Format: "8.8.8.8:53" or "8.8.8.8" (default port 53, type udp)
func parseNameserver(ns string) (Nameserver, error) {
	var ip string
	var port int = 53
	var nsType string = "udp"

	// Check if port is specified
	if strings.Contains(ns, ":") {
		parts := strings.Split(ns, ":")
		if len(parts) != 2 {
			return Nameserver{}, fmt.Errorf("invalid nameserver format '%s': expected IP:port", ns)
		}

		ip = parts[0]
		portNum, err := strconv.Atoi(parts[1])
		if err != nil {
			return Nameserver{}, fmt.Errorf("invalid port in nameserver '%s': %v", ns, err)
		}
		if portNum < 1 || portNum > 65535 {
			return Nameserver{}, fmt.Errorf("port must be between 1 and 65535 (got %d)", portNum)
		}
		port = portNum
	} else {
		ip = ns
	}

	// Validate IP
	if err := validateDNSIP(ip); err != nil {
		return Nameserver{}, err
	}

	return Nameserver{
		IP:     ip,
		Port:   port,
		NSType: nsType,
	}, nil
}

// validateDNSIP validates a DNS server IP address
func validateDNSIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return nil
}

// printDNSUsage prints usage information for the dns command
func printDNSUsage() {
	fmt.Println("Usage: netbird-manage dns [options]")
	fmt.Println("\nManage DNS nameserver groups and settings.")
	fmt.Println("\nQuery Flags:")
	fmt.Println("  --list                              List all DNS nameserver groups")
	fmt.Println("    --filter-name <pattern>           Filter by name pattern")
	fmt.Println("    --primary-only                    Show only primary groups")
	fmt.Println("    --enabled-only                    Show only enabled groups")
	fmt.Println("  --inspect <group-id>                Show detailed DNS group information")
	fmt.Println("  --get-settings                      Get DNS settings for the account")
	fmt.Println()
	fmt.Println("Modification Flags:")
	fmt.Println("  --create <name>                     Create DNS nameserver group")
	fmt.Println("    --nameservers <ip:port,...>       Nameservers (e.g., 8.8.8.8:53,1.1.1.1:53)")
	fmt.Println("    --groups <id1,id2,...>            Target group IDs (required)")
	fmt.Println("    --domains <domain1,domain2,...>   Match domains (optional)")
	fmt.Println("    --description <desc>              Description (optional)")
	fmt.Println("    --search-domains                  Enable search domains")
	fmt.Println("    --primary                         Set as primary DNS")
	fmt.Println("    --enabled                         Enable group (default)")
	fmt.Println("    --disabled                        Disable group")
	fmt.Println()
	fmt.Println("  --update <group-id>                 Update DNS nameserver group")
	fmt.Println("    [same flags as create]            All fields optional")
	fmt.Println()
	fmt.Println("  --delete <group-id>                 Delete DNS nameserver group")
	fmt.Println()
	fmt.Println("  --enable <group-id>                 Enable DNS group")
	fmt.Println("  --disable <group-id>                Disable DNS group")
	fmt.Println()
	fmt.Println("  --update-settings                   Update DNS settings")
	fmt.Println("    --disabled-groups <id1,id2,...>   Groups with disabled DNS management")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List all DNS groups")
	fmt.Println("  netbird-manage dns --list")
	fmt.Println()
	fmt.Println("  # Create a DNS group with Google and Cloudflare DNS")
	fmt.Println("  netbird-manage dns --create \"corp-dns\" \\")
	fmt.Println("    --nameservers \"8.8.8.8:53,1.1.1.1:53\" \\")
	fmt.Println("    --groups <group-id> \\")
	fmt.Println("    --domains \"example.com,internal.local\" \\")
	fmt.Println("    --primary")
	fmt.Println()
	fmt.Println("  # Inspect a DNS group")
	fmt.Println("  netbird-manage dns --inspect <group-id>")
	fmt.Println()
	fmt.Println("  # Get DNS settings")
	fmt.Println("  netbird-manage dns --get-settings")
	fmt.Println()
	fmt.Println("  # Update DNS settings")
	fmt.Println("  netbird-manage dns --update-settings --disabled-groups <group-id>")
}
