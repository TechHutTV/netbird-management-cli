// models.go
package main

// Config holds the client configuration
type Config struct {
	Token         string `json:"token"`
	ManagementURL string `json:"management_url"`
}

// Peer represents a single NetBird peer (from peers.mdx)
type Peer struct {
	ID                        string        `json:"id"`
	Name                      string        `json:"name"`
	IP                        string        `json:"ip"`
	Connected                 bool          `json:"connected"`
	LastSeen                  string        `json:"last_seen"`
	OS                        string        `json:"os"`
	Version                   string        `json:"version"`
	Groups                    []PolicyGroup `json:"groups"` // This uses the simplified group object
	Hostname                  string        `json:"hostname"`
	SSHEnabled                bool          `json:"ssh_enabled"`
	LoginExpirationEnabled    bool          `json:"login_expiration_enabled"`
	InactivityExpirationEnabled bool        `json:"inactivity_expiration_enabled"`
	ApprovalRequired          *bool         `json:"approval_required,omitempty"` // Optional, cloud-only
}

// PeerUpdateRequest represents the request body for updating a peer
type PeerUpdateRequest struct {
	Name                      string `json:"name"`
	SSHEnabled                bool   `json:"ssh_enabled"`
	LoginExpirationEnabled    bool   `json:"login_expiration_enabled"`
	InactivityExpirationEnabled bool   `json:"inactivity_expiration_enabled"`
	ApprovalRequired          *bool  `json:"approval_required,omitempty"`
	IP                        string `json:"ip,omitempty"`
}

// PolicyGroup represents the simplified group object found inside other resources (like Peer)
type PolicyGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GroupDetail represents the full group object (from groups.mdx)
type GroupDetail struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	PeersCount     int                     `json:"peers_count"`
	ResourcesCount int                     `json:"resources_count"`
	Issued         string                  `json:"issued"`
	Peers          []Peer                  `json:"peers"` // Contains list of full Peer objects
	Resources      []GroupResource `json:"resources"`
}

// GroupResource represents a resource in a group's details
type GroupResource struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// GroupPutRequest is the structure needed to update a group
type GroupPutRequest struct {
	Name      string                    `json:"name"`
	Peers     []string                  `json:"peers"` // List of Peer IDs
	Resources []GroupResourcePutRequest `json:"resources"`
}

// GroupResourcePutRequest is the simplified resource struct for PUT requests
type GroupResourcePutRequest struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Network represents a single network (from networks.mdx)
type Network struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Routers           []string `json:"routers"`
	RoutingPeersCount int      `json:"routing_peers_count"`
	Resources         []string `json:"resources"`
	Policies          []string `json:"policies"`
	Description       string   `json:"description"`
}

// Policy represents an access control policy (from policies.mdx)
type Policy struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Enabled     bool         `json:"enabled"`
	Rules       []PolicyRule `json:"rules"`
}

// PolicyRule is a rule within a policy
type PolicyRule struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Enabled      bool          `json:"enabled"`
	Action       string        `json:"action"` // "accept" or "drop"
	Protocol     string        `json:"protocol"`
	Sources      []PolicyGroup `json:"sources"`
	Destinations []PolicyGroup `json:"destinations"`
}
