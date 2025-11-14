// networks.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// handleNetworksCommand routes network-related commands
func handleNetworksCommand(client *Client, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: netbird-manage networks")
	}

	switch args[0] {
	case "networks":
		return client.listNetworks()
	default:
		printUsage()
		return nil
	}
}

// listNetworks implements the "networks" command
func (c *Client) listNetworks() error {
	resp, err := c.makeRequest("GET", "/networks", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return fmt.Errorf("failed to decode networks response: %v", err)
	}

	if len(networks) == 0 {
		fmt.Println("No networks found.")
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tROUTERS\tRESOURCES\tPOLICIES\tDESCRIPTION")
	fmt.Fprintln(w, "--\t----\t-------\t---------\t--------\t-----------")

	for _, net := range networks {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%s\n",
			net.ID,
			net.Name,
			net.RoutingPeersCount,
			len(net.Resources),
			len(net.Policies),
			net.Description,
		)
	}
	w.Flush()
	return nil
}
