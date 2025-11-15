// networks.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

// handleNetworkCommand routes network-related commands
func handleNetworkCommand(client *Client, args []string) error {
	// Create a new flag set for the 'network' command
	networkCmd := flag.NewFlagSet("network", flag.ContinueOnError)
	networkCmd.SetOutput(os.Stderr) // Send errors to stderr
	networkCmd.Usage = printNetworkUsage // Set our custom usage function

	// Define the flags for the 'network' command
	listFlag := networkCmd.Bool("list", false, "List all networks")

	// If no flags are provided (just 'netbird-manage network'), show usage
	if len(args) == 1 {
		printNetworkUsage()
		return nil
	}

	// Parse the flags (all args *after* 'network')
	if err := networkCmd.Parse(args[1:]); err != nil {
		// The flag package will print an error, so we just return
		return nil
	}

	// Handle the flags
	if *listFlag {
		return client.listNetworks()
	}

	// If no known flag was used
	fmt.Fprintln(os.Stderr, "Error: Invalid or missing flags for 'network' command.")
	printNetworkUsage()
	return nil
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
