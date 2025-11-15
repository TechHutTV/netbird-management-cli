// policies.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// handlePoliciesCommand routes policy-related commands
func handlePoliciesCommand(client *Client, args []string) error {
	// Create a new flag set for the 'policy' command
	policyCmd := flag.NewFlagSet("policy", flag.ContinueOnError)
	policyCmd.SetOutput(os.Stderr)
	policyCmd.Usage = printPolicyUsage

	// Define the flags for the 'policy' command
	listFlag := policyCmd.Bool("list", false, "List all policies")
	inspectFlag := policyCmd.String("inspect", "", "Inspect a specific policy by ID")
	createFlag := policyCmd.String("create", "", "Create a new policy with the given name")
	deleteFlag := policyCmd.String("delete", "", "Delete a policy by ID")
	enableFlag := policyCmd.String("enable", "", "Enable a policy by ID")
	disableFlag := policyCmd.String("disable", "", "Disable a policy by ID")

	// List filtering flags
	enabledFilterFlag := policyCmd.Bool("enabled", false, "Filter to show only enabled policies")
	disabledFilterFlag := policyCmd.Bool("disabled", false, "Filter to show only disabled policies")
	nameFilterFlag := policyCmd.String("name", "", "Filter policies by name (contains)")

	// Create/edit flags
	descriptionFlag := policyCmd.String("description", "", "Policy description")
	enabledFlag := policyCmd.Bool("active", true, "Enable the policy (default: true)")

	// Rule management flags
	addRuleFlag := policyCmd.String("add-rule", "", "Add a rule to a policy (requires --policy-id)")
	editRuleFlag := policyCmd.String("edit-rule", "", "Edit a rule by name or ID (requires --policy-id)")
	removeRuleFlag := policyCmd.String("remove-rule", "", "Remove a rule by name or ID (requires --policy-id)")
	policyIDFlag := policyCmd.String("policy-id", "", "Target policy ID for rule operations")

	// Rule configuration flags
	ruleNameFlag := policyCmd.String("rule-name", "", "Rule name")
	ruleDescFlag := policyCmd.String("rule-description", "", "Rule description")
	actionFlag := policyCmd.String("action", "accept", "Rule action: accept or drop")
	protocolFlag := policyCmd.String("protocol", "all", "Protocol: tcp, udp, icmp, or all")
	sourcesFlag := policyCmd.String("sources", "", "Source group IDs or names (comma-separated)")
	destinationsFlag := policyCmd.String("destinations", "", "Destination group IDs or names (comma-separated)")
	portsFlag := policyCmd.String("ports", "", "Ports (comma-separated, e.g., 80,443,8080)")
	portRangeFlag := policyCmd.String("port-range", "", "Port range (e.g., 6000-6100)")
	bidirectionalFlag := policyCmd.Bool("bidirectional", false, "Apply rule in both directions")
	ruleEnabledFlag := policyCmd.Bool("rule-enabled", true, "Enable the rule (default: true)")

	// If no flags are provided (just 'netbird-manage policy'), show usage
	if len(args) == 1 {
		printPolicyUsage()
		return nil
	}

	// Parse the flags (all args *after* 'policy')
	if err := policyCmd.Parse(args[1:]); err != nil {
		return nil
	}

	// Handle the flags in priority order

	// Create policy
	if *createFlag != "" {
		return client.createPolicy(*createFlag, *descriptionFlag, *enabledFlag)
	}

	// Delete policy
	if *deleteFlag != "" {
		return client.deletePolicy(*deleteFlag)
	}

	// Enable policy
	if *enableFlag != "" {
		return client.togglePolicy(*enableFlag, true)
	}

	// Disable policy
	if *disableFlag != "" {
		return client.togglePolicy(*disableFlag, false)
	}

	// Add rule to policy
	if *addRuleFlag != "" {
		if *policyIDFlag == "" {
			return fmt.Errorf("--policy-id is required when adding a rule")
		}
		if *sourcesFlag == "" || *destinationsFlag == "" {
			return fmt.Errorf("--sources and --destinations are required when adding a rule")
		}
		return client.addRuleToPolicy(*policyIDFlag, *addRuleFlag, &RuleConfig{
			Description:   *ruleDescFlag,
			Action:        *actionFlag,
			Protocol:      *protocolFlag,
			Sources:       *sourcesFlag,
			Destinations:  *destinationsFlag,
			Ports:         *portsFlag,
			PortRange:     *portRangeFlag,
			Bidirectional: *bidirectionalFlag,
			Enabled:       *ruleEnabledFlag,
		})
	}

	// Edit rule
	if *editRuleFlag != "" {
		if *policyIDFlag == "" {
			return fmt.Errorf("--policy-id is required when editing a rule")
		}
		return client.editRule(*policyIDFlag, *editRuleFlag, &RuleConfig{
			Name:          *ruleNameFlag,
			Description:   *ruleDescFlag,
			Action:        *actionFlag,
			Protocol:      *protocolFlag,
			Sources:       *sourcesFlag,
			Destinations:  *destinationsFlag,
			Ports:         *portsFlag,
			PortRange:     *portRangeFlag,
			Bidirectional: *bidirectionalFlag,
			Enabled:       *ruleEnabledFlag,
		})
	}

	// Remove rule
	if *removeRuleFlag != "" {
		if *policyIDFlag == "" {
			return fmt.Errorf("--policy-id is required when removing a rule")
		}
		return client.removeRuleFromPolicy(*policyIDFlag, *removeRuleFlag)
	}

	// Inspect policy
	if *inspectFlag != "" {
		return client.inspectPolicy(*inspectFlag)
	}

	// List policies (with optional filtering)
	if *listFlag {
		filters := &PolicyFilters{
			EnabledOnly:  *enabledFilterFlag,
			DisabledOnly: *disabledFilterFlag,
			NameFilter:   *nameFilterFlag,
		}
		return client.listPolicies(filters)
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'policy' command.")
	printPolicyUsage()
	return nil
}

// PolicyFilters holds filtering options for listing policies
type PolicyFilters struct {
	EnabledOnly  bool
	DisabledOnly bool
	NameFilter   string
}

// RuleConfig holds configuration for creating/editing rules
type RuleConfig struct {
	Name          string
	Description   string
	Action        string
	Protocol      string
	Sources       string
	Destinations  string
	Ports         string
	PortRange     string
	Bidirectional bool
	Enabled       bool
}

// listPolicies implements the "policy --list" command
func (c *Client) listPolicies(filters *PolicyFilters) error {
	resp, err := c.makeRequest("GET", "/policies", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var policies []Policy
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return fmt.Errorf("failed to decode policies response: %v", err)
	}

	// Apply filters
	var filteredPolicies []Policy
	for _, pol := range policies {
		// Filter by enabled/disabled
		if filters.EnabledOnly && !pol.Enabled {
			continue
		}
		if filters.DisabledOnly && pol.Enabled {
			continue
		}

		// Filter by name
		if filters.NameFilter != "" && !strings.Contains(strings.ToLower(pol.Name), strings.ToLower(filters.NameFilter)) {
			continue
		}

		filteredPolicies = append(filteredPolicies, pol)
	}

	if len(filteredPolicies) == 0 {
		fmt.Println("No policies found.")
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tENABLED\tRULES\tDESCRIPTION")
	fmt.Fprintln(w, "--\t----\t-------\t-----\t-----------")

	for _, pol := range filteredPolicies {
		fmt.Fprintf(w, "%s\t%s\t%t\t%d\t%s\n",
			pol.ID,
			pol.Name,
			pol.Enabled,
			len(pol.Rules),
			pol.Description,
		)
		// Print rules
		for _, rule := range pol.Rules {
			ports := formatPorts(rule.Ports, rule.PortRanges)
			bidir := ""
			if rule.Bidirectional {
				bidir = " (bidirectional)"
			}
			fmt.Fprintf(w, "\t  -> %s\t%s\t%s%s\t%s -> %s\n",
				rule.Name,
				rule.Action,
				rule.Protocol,
				ports,
				getGroupNames(rule.Sources),
				getGroupNames(rule.Destinations)+bidir,
			)
		}
	}
	w.Flush()
	return nil
}

// inspectPolicy implements the "policy --inspect" command
func (c *Client) inspectPolicy(policyID string) error {
	resp, err := c.makeRequest("GET", "/policies/"+policyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return fmt.Errorf("failed to decode policy response: %v", err)
	}

	// Print detailed policy information
	fmt.Println("Policy Details:")
	fmt.Println("===============")
	fmt.Printf("ID:          %s\n", policy.ID)
	fmt.Printf("Name:        %s\n", policy.Name)
	fmt.Printf("Description: %s\n", policy.Description)
	fmt.Printf("Enabled:     %t\n", policy.Enabled)

	if len(policy.SourcePostureChecks) > 0 {
		fmt.Printf("Posture Checks: %s\n", strings.Join(policy.SourcePostureChecks, ", "))
	}

	fmt.Printf("\nRules (%d):\n", len(policy.Rules))
	fmt.Println("===========")

	if len(policy.Rules) == 0 {
		fmt.Println("  No rules defined")
		return nil
	}

	for i, rule := range policy.Rules {
		fmt.Printf("\n[%d] %s (ID: %s)\n", i+1, rule.Name, rule.ID)
		if rule.Description != "" {
			fmt.Printf("    Description:   %s\n", rule.Description)
		}
		fmt.Printf("    Enabled:       %t\n", rule.Enabled)
		fmt.Printf("    Action:        %s\n", rule.Action)
		fmt.Printf("    Protocol:      %s\n", rule.Protocol)
		fmt.Printf("    Bidirectional: %t\n", rule.Bidirectional)

		if len(rule.Ports) > 0 {
			fmt.Printf("    Ports:         %s\n", strings.Join(rule.Ports, ", "))
		}

		if len(rule.PortRanges) > 0 {
			var ranges []string
			for _, pr := range rule.PortRanges {
				ranges = append(ranges, fmt.Sprintf("%d-%d", pr.Start, pr.End))
			}
			fmt.Printf("    Port Ranges:   %s\n", strings.Join(ranges, ", "))
		}

		fmt.Printf("    Sources:       %s\n", getGroupNames(rule.Sources))
		fmt.Printf("    Destinations:  %s\n", getGroupNames(rule.Destinations))
	}

	fmt.Println()
	return nil
}

// createPolicy implements the "policy --create" command
func (c *Client) createPolicy(name, description string, enabled bool) error {
	reqBody := PolicyCreateRequest{
		Name:        name,
		Description: description,
		Enabled:     enabled,
		Rules:       []PolicyRule{},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := c.makeRequest("POST", "/policies", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var createdPolicy Policy
	if err := json.NewDecoder(resp.Body).Decode(&createdPolicy); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	fmt.Printf("Policy created successfully:\n")
	fmt.Printf("  ID:      %s\n", createdPolicy.ID)
	fmt.Printf("  Name:    %s\n", createdPolicy.Name)
	fmt.Printf("  Enabled: %t\n", createdPolicy.Enabled)
	return nil
}

// deletePolicy implements the "policy --delete" command
func (c *Client) deletePolicy(policyID string) error {
	resp, err := c.makeRequest("DELETE", "/policies/"+policyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Policy '%s' deleted successfully\n", policyID)
	return nil
}

// togglePolicy enables or disables a policy
func (c *Client) togglePolicy(policyID string, enable bool) error {
	// First, get the current policy
	resp, err := c.makeRequest("GET", "/policies/"+policyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return fmt.Errorf("failed to decode policy: %v", err)
	}

	// Update the enabled status
	policy.Enabled = enable

	// Send the update
	updateReq := PolicyUpdateRequest{
		Name:                policy.Name,
		Description:         policy.Description,
		Enabled:             policy.Enabled,
		Rules:               policy.Rules,
		SourcePostureChecks: policy.SourcePostureChecks,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp2, err := c.makeRequest("PUT", "/policies/"+policyID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	status := "enabled"
	if !enable {
		status = "disabled"
	}
	fmt.Printf("Policy '%s' %s successfully\n", policy.Name, status)
	return nil
}

// addRuleToPolicy implements the "policy --add-rule" command
func (c *Client) addRuleToPolicy(policyID, ruleName string, config *RuleConfig) error {
	// First, get the current policy
	resp, err := c.makeRequest("GET", "/policies/"+policyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return fmt.Errorf("failed to decode policy: %v", err)
	}

	// Build the new rule
	newRule, err := c.buildRuleFromConfig(ruleName, config)
	if err != nil {
		return err
	}

	// Add the rule to the policy
	policy.Rules = append(policy.Rules, *newRule)

	// Send the update
	updateReq := PolicyUpdateRequest{
		Name:                policy.Name,
		Description:         policy.Description,
		Enabled:             policy.Enabled,
		Rules:               policy.Rules,
		SourcePostureChecks: policy.SourcePostureChecks,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp2, err := c.makeRequest("PUT", "/policies/"+policyID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	fmt.Printf("Rule '%s' added to policy '%s' successfully\n", ruleName, policy.Name)
	return nil
}

// editRule implements the "policy --edit-rule" command
func (c *Client) editRule(policyID, ruleIdentifier string, config *RuleConfig) error {
	// First, get the current policy
	resp, err := c.makeRequest("GET", "/policies/"+policyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return fmt.Errorf("failed to decode policy: %v", err)
	}

	// Find the rule by name or ID
	ruleIndex := -1
	for i, rule := range policy.Rules {
		if rule.ID == ruleIdentifier || rule.Name == ruleIdentifier {
			ruleIndex = i
			break
		}
	}

	if ruleIndex == -1 {
		return fmt.Errorf("rule '%s' not found in policy", ruleIdentifier)
	}

	// Update rule fields (only update non-empty values)
	existingRule := &policy.Rules[ruleIndex]

	if config.Name != "" {
		existingRule.Name = config.Name
	}
	if config.Description != "" {
		existingRule.Description = config.Description
	}
	if config.Action != "" {
		existingRule.Action = config.Action
	}
	if config.Protocol != "" {
		existingRule.Protocol = config.Protocol
	}
	if config.Sources != "" {
		sourceGroups, err := c.resolveGroupIdentifiers(config.Sources)
		if err != nil {
			return err
		}
		existingRule.Sources = sourceGroups
	}
	if config.Destinations != "" {
		destGroups, err := c.resolveGroupIdentifiers(config.Destinations)
		if err != nil {
			return err
		}
		existingRule.Destinations = destGroups
	}
	if config.Ports != "" {
		existingRule.Ports = strings.Split(config.Ports, ",")
	}
	if config.PortRange != "" {
		portRange, err := parsePortRange(config.PortRange)
		if err != nil {
			return err
		}
		existingRule.PortRanges = []PortRange{*portRange}
	}
	existingRule.Bidirectional = config.Bidirectional
	existingRule.Enabled = config.Enabled

	// Send the update
	updateReq := PolicyUpdateRequest{
		Name:                policy.Name,
		Description:         policy.Description,
		Enabled:             policy.Enabled,
		Rules:               policy.Rules,
		SourcePostureChecks: policy.SourcePostureChecks,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp2, err := c.makeRequest("PUT", "/policies/"+policyID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	fmt.Printf("Rule '%s' updated successfully\n", existingRule.Name)
	return nil
}

// removeRuleFromPolicy implements the "policy --remove-rule" command
func (c *Client) removeRuleFromPolicy(policyID, ruleIdentifier string) error {
	// First, get the current policy
	resp, err := c.makeRequest("GET", "/policies/"+policyID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return fmt.Errorf("failed to decode policy: %v", err)
	}

	// Find and remove the rule by name or ID
	ruleIndex := -1
	var removedRuleName string
	for i, rule := range policy.Rules {
		if rule.ID == ruleIdentifier || rule.Name == ruleIdentifier {
			ruleIndex = i
			removedRuleName = rule.Name
			break
		}
	}

	if ruleIndex == -1 {
		return fmt.Errorf("rule '%s' not found in policy", ruleIdentifier)
	}

	// Remove the rule
	policy.Rules = append(policy.Rules[:ruleIndex], policy.Rules[ruleIndex+1:]...)

	// Send the update
	updateReq := PolicyUpdateRequest{
		Name:                policy.Name,
		Description:         policy.Description,
		Enabled:             policy.Enabled,
		Rules:               policy.Rules,
		SourcePostureChecks: policy.SourcePostureChecks,
	}

	bodyBytes, err := json.Marshal(updateReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	resp2, err := c.makeRequest("PUT", "/policies/"+policyID, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	fmt.Printf("Rule '%s' removed from policy '%s' successfully\n", removedRuleName, policy.Name)
	return nil
}

// buildRuleFromConfig creates a PolicyRule from RuleConfig
func (c *Client) buildRuleFromConfig(ruleName string, config *RuleConfig) (*PolicyRule, error) {
	// Validate required fields
	if config.Action != "accept" && config.Action != "drop" {
		return nil, fmt.Errorf("invalid action '%s': must be 'accept' or 'drop'", config.Action)
	}

	// Resolve source and destination groups
	sourceGroups, err := c.resolveGroupIdentifiers(config.Sources)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source groups: %v", err)
	}

	destGroups, err := c.resolveGroupIdentifiers(config.Destinations)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve destination groups: %v", err)
	}

	// Build the rule
	rule := &PolicyRule{
		Name:          ruleName,
		Description:   config.Description,
		Enabled:       config.Enabled,
		Action:        config.Action,
		Bidirectional: config.Bidirectional,
		Protocol:      config.Protocol,
		Sources:       sourceGroups,
		Destinations:  destGroups,
	}

	// Add ports if specified
	if config.Ports != "" {
		rule.Ports = strings.Split(config.Ports, ",")
	}

	// Add port range if specified
	if config.PortRange != "" {
		portRange, err := parsePortRange(config.PortRange)
		if err != nil {
			return nil, err
		}
		rule.PortRanges = []PortRange{*portRange}
	}

	return rule, nil
}

// resolveGroupIdentifiers converts group names or IDs to PolicyGroup objects
func (c *Client) resolveGroupIdentifiers(identifiers string) ([]PolicyGroup, error) {
	if identifiers == "" {
		return []PolicyGroup{}, nil
	}

	parts := strings.Split(identifiers, ",")
	var groups []PolicyGroup

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Try to find the group by ID or name
		group, err := c.getGroupByNameOrID(part)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve group '%s': %v", part, err)
		}

		groups = append(groups, PolicyGroup{
			ID:   group.ID,
			Name: group.Name,
		})
	}

	return groups, nil
}

// getGroupByNameOrID retrieves a group by name or ID
func (c *Client) getGroupByNameOrID(identifier string) (*GroupDetail, error) {
	// First, try to get it as an ID
	resp, err := c.makeRequest("GET", "/groups/"+identifier, nil)
	if err == nil {
		defer resp.Body.Close()
		var group GroupDetail
		if err := json.NewDecoder(resp.Body).Decode(&group); err == nil {
			return &group, nil
		}
	}

	// If that fails, try to find it by name
	return c.getGroupByName(identifier)
}

// parsePortRange parses a port range string like "6000-6100"
func parsePortRange(rangeStr string) (*PortRange, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid port range format '%s': expected 'start-end'", rangeStr)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start port '%s': %v", parts[0], err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end port '%s': %v", parts[1], err)
	}

	if start < 1 || start > 65535 || end < 1 || end > 65535 {
		return nil, fmt.Errorf("port numbers must be between 1 and 65535")
	}

	if start > end {
		return nil, fmt.Errorf("start port %d cannot be greater than end port %d", start, end)
	}

	return &PortRange{Start: start, End: end}, nil
}

// formatPorts formats ports and port ranges for display
func formatPorts(ports []string, portRanges []PortRange) string {
	if len(ports) == 0 && len(portRanges) == 0 {
		return ""
	}

	var parts []string
	if len(ports) > 0 {
		parts = append(parts, strings.Join(ports, ","))
	}

	for _, pr := range portRanges {
		parts = append(parts, fmt.Sprintf("%d-%d", pr.Start, pr.End))
	}

	return ":" + strings.Join(parts, ",")
}

// getGroupNames is a helper for formatting policy output
func getGroupNames(groups []PolicyGroup) string {
	if len(groups) == 0 {
		return "[All]"
	}
	var names []string
	for _, g := range groups {
		names = append(names, g.Name)
	}
	return strings.Join(names, ", ")
}

// printPolicyUsage prints usage information for the policy command
func printPolicyUsage() {
	fmt.Println("Usage: netbird-manage policy [options]")
	fmt.Println("\nPolicy Management:")
	fmt.Println("  --list                      List all policies")
	fmt.Println("  --list --enabled            List only enabled policies")
	fmt.Println("  --list --disabled           List only disabled policies")
	fmt.Println("  --list --name <filter>      Filter policies by name")
	fmt.Println("  --inspect <policy-id>       Show detailed policy information")
	fmt.Println("  --create <name>             Create a new policy")
	fmt.Println("      --description <text>        Policy description")
	fmt.Println("      --active <true|false>       Enable policy (default: true)")
	fmt.Println("  --delete <policy-id>        Delete a policy")
	fmt.Println("  --enable <policy-id>        Enable a policy")
	fmt.Println("  --disable <policy-id>       Disable a policy")
	fmt.Println("\nRule Management:")
	fmt.Println("  --add-rule <rule-name>      Add a rule to a policy")
	fmt.Println("      --policy-id <id>            Target policy (required)")
	fmt.Println("      --action <accept|drop>      Rule action (default: accept)")
	fmt.Println("      --protocol <tcp|udp|icmp|all> Protocol type (default: all)")
	fmt.Println("      --sources <groups>          Source group names/IDs (comma-separated)")
	fmt.Println("      --destinations <groups>     Destination group names/IDs (comma-separated)")
	fmt.Println("      --ports <ports>             Ports (e.g., 80,443,8080)")
	fmt.Println("      --port-range <range>        Port range (e.g., 6000-6100)")
	fmt.Println("      --bidirectional             Apply rule in both directions")
	fmt.Println("      --rule-description <text>   Rule description")
	fmt.Println("      --rule-enabled <true|false> Enable rule (default: true)")
	fmt.Println("  --edit-rule <rule-name|id>  Edit an existing rule")
	fmt.Println("      --policy-id <id>            Target policy (required)")
	fmt.Println("      --rule-name <new-name>      New rule name")
	fmt.Println("      [other rule options]        Update other rule properties")
	fmt.Println("  --remove-rule <rule-name|id> Remove a rule from policy")
	fmt.Println("      --policy-id <id>            Target policy (required)")
	fmt.Println("\nExamples:")
	fmt.Println("  # List all policies")
	fmt.Println("  netbird-manage policy --list")
	fmt.Println()
	fmt.Println("  # Create a new policy")
	fmt.Println("  netbird-manage policy --create \"dev-access\" --description \"Developer access policy\"")
	fmt.Println()
	fmt.Println("  # Inspect a policy")
	fmt.Println("  netbird-manage policy --inspect <policy-id>")
	fmt.Println()
	fmt.Println("  # Add a rule allowing TCP traffic on ports 80,443")
	fmt.Println("  netbird-manage policy --add-rule \"web-access\" \\")
	fmt.Println("    --policy-id <policy-id> \\")
	fmt.Println("    --action accept \\")
	fmt.Println("    --protocol tcp \\")
	fmt.Println("    --sources \"developers\" \\")
	fmt.Println("    --destinations \"web-servers\" \\")
	fmt.Println("    --ports \"80,443\" \\")
	fmt.Println("    --bidirectional")
	fmt.Println()
	fmt.Println("  # Add a rule with port range")
	fmt.Println("  netbird-manage policy --add-rule \"app-ports\" \\")
	fmt.Println("    --policy-id <policy-id> \\")
	fmt.Println("    --protocol tcp \\")
	fmt.Println("    --sources \"app-servers\" \\")
	fmt.Println("    --destinations \"database\" \\")
	fmt.Println("    --port-range \"6000-6100\"")
	fmt.Println()
	fmt.Println("  # Edit a rule")
	fmt.Println("  netbird-manage policy --edit-rule \"web-access\" \\")
	fmt.Println("    --policy-id <policy-id> \\")
	fmt.Println("    --ports \"80,443,8443\"")
	fmt.Println()
	fmt.Println("  # Remove a rule")
	fmt.Println("  netbird-manage policy --remove-rule \"web-access\" --policy-id <policy-id>")
	fmt.Println()
	fmt.Println("  # Disable a policy")
	fmt.Println("  netbird-manage policy --disable <policy-id>")
}
