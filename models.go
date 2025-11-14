// models.go
package main

// Peer represents a single NetBird peer (from peers.mdx)
// This struct now includes Groups for the --inspect command
type Peer struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	IP        string        `json:"ip"`
	Connected bool          `json:"connected"`
	LastSeen  string        `json:"last_seen"`
	OS        string        `json:"os"`
	Version   string        `json:"version"`
	Hostname  string        `json:"hostname"`
	Groups    []PolicyGroup `json:"groups"` // Used for --inspect
}

// User represents a NetBird user (from users.mdx)
// Still needed for the config.go test function, even if not a primary command.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// Network represents a NetBird Network (from networks.mdx)
type Network struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Routers           []string `json:"routers"`
	Resources         []string `json:"resources"`
	Policies          []string `json:"policies"`
	RoutingPeersCount int      `json:"routing_peers_count"`
}

// Policy represents an access control policy (from policies.mdx)
type Policy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Rules       []Rule `json:"rules"`
}

// Rule is a single rule within a policy
type Rule struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Enabled      bool          `json:"enabled"`
	Action       string        `json:"action"`
	Protocol     string        `json:"protocol"`
	Sources      []PolicyGroup `json:"sources"`
	Destinations []PolicyGroup `json:"destinations"`
	Ports        []string      `json:"ports"`
}

// PolicyGroup is a lightweight representation of a group used in policy rules
type PolicyGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GroupDetail represents the full group object from GET /api/groups
type GroupDetail struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	PeersCount     int    `json:"peers_count"`
	ResourcesCount int    `json:"resources_count"`
	Issued         string `json:"issued"`
	Peers          []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"peers"`
	Resources []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"resources"`
}

// GroupPutRequest is the object for PUT /api/groups/{id}
type GroupPutRequest struct {
	Name      string                  `json:"name"`
	Peers     []string                `json:"peers"`
	Resources []GroupResourcePutRequest `json:"resources"`
}

// GroupResourcePutRequest is the resource object for the PUT request
type GroupResourcePutRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
