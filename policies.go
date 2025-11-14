// policies.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// handlePoliciesCommand routes policy-related commands
func handlePoliciesCommand(client *Client, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: netbird-manage policy")
	}

	switch args[0] {
	case "policy":
		return client.listPolicies()
	default:
		printUsage()
		return nil
	}
}

// listPolicies implements the "policy" command
func (c *Client) listPolicies() error {
	resp, err := c.makeRequest("GET", "/policies", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var policies []Policy
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return fmt.Errorf("failed to decode policies response: %v", err)
	}

	if len(policies) == 0 {
		fmt.Println("No policies found.")
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tENABLED\tRULES\tDESCRIPTION")
	fmt.Fprintln(w, "--\t----\t-------\t-----\t-----------")

	for _, pol := range policies {
		fmt.Fprintf(w, "%s\t%s\t%t\t%d\t%s\n",
			pol.ID,
			pol.Name,
			pol.Enabled,
			len(pol.Rules),
			pol.Description,
		)
		// Optionally print rules
		for _, rule := range pol.Rules {
			fmt.Fprintf(w, "\t  -> %s\t%s\t(%s)\t%s -> %s\n",
				rule.ID,
				rule.Action,
				rule.Protocol,
				getGroupNames(rule.Sources),
				getGroupNames(rule.Destinations),
			)
		}
	}
	w.Flush()
	return nil
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
