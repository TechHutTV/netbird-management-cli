// groups.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// handleGroupsCommand routes group-related commands
func handleGroupsCommand(client *Client, args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "group", "groups":
		if len(args) != 1 {
			return fmt.Errorf("usage: netbird-manage group")
		}
		return client.listGroups()
	default:
		printUsage()
		return nil
	}
}

// listGroups implements the "group" command
func (c *Client) listGroups() error {
	resp, err := c.makeRequest("GET", "/groups", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var groups []GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return fmt.Errorf("failed to decode groups response: %v", err)
	}

	if len(groups) == 0 {
		fmt.Println("No groups found.")
		return nil
	}

	// Print a formatted table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPEERS\tRESOURCES\tISSUED BY")
	fmt.Fprintln(w, "--\t----\t-----\t---------\t---------")

	for _, g := range groups {
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
