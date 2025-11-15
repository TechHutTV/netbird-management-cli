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

// NetworkDetail represents the full network object (GET /networks/{id} returns IDs as strings)
type NetworkDetail struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Routers           []string `json:"routers"`           // Router IDs
	RoutingPeersCount int      `json:"routing_peers_count"`
	Resources         []string `json:"resources"`         // Resource IDs
	Policies          []string `json:"policies"`          // Policy IDs
}

// NetworkResource represents a resource within a network (host, subnet, or domain)
type NetworkResource struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Address     string        `json:"address"` // IP (1.1.1.1 or 1.1.1.1/32), subnet (192.168.0.0/24), or domain (*.example.com)
	Enabled     bool          `json:"enabled"`
	Groups      []PolicyGroup `json:"groups"` // Group objects with id and name
}

// NetworkRouter represents a routing peer in a network
type NetworkRouter struct {
	ID         string   `json:"id"`
	Peer       string   `json:"peer,omitempty"`        // Single peer ID (mutually exclusive with peer_groups)
	PeerGroups []string `json:"peer_groups,omitempty"` // Peer group IDs (mutually exclusive with peer)
	Metric     int      `json:"metric"`                // 1-9999, lower = higher priority
	Masquerade bool     `json:"masquerade"`            // Enable NAT
	Enabled    bool     `json:"enabled"`
}

// NetworkCreateRequest represents the request body for creating a network
type NetworkCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// NetworkUpdateRequest represents the request body for updating a network
type NetworkUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NetworkResourceRequest represents the request body for creating/updating a network resource
type NetworkResourceRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Address     string   `json:"address"`
	Enabled     bool     `json:"enabled"`
	Groups      []string `json:"groups"`
}

// NetworkRouterRequest represents the request body for creating/updating a network router
type NetworkRouterRequest struct {
	Peer       string   `json:"peer,omitempty"`        // Single peer ID
	PeerGroups []string `json:"peer_groups,omitempty"` // Peer group IDs
	Metric     int      `json:"metric"`
	Masquerade bool     `json:"masquerade"`
	Enabled    bool     `json:"enabled"`
}

// Policy represents an access control policy (from policies.mdx)
type Policy struct {
	ID                  string       `json:"id"`
	Name                string       `json:"name"`
	Description         string       `json:"description"`
	Enabled             bool         `json:"enabled"`
	Rules               []PolicyRule `json:"rules"`
	SourcePostureChecks []string     `json:"source_posture_checks,omitempty"`
}

// PolicyRule is a rule within a policy
type PolicyRule struct {
	ID                  string            `json:"id,omitempty"`
	Name                string            `json:"name"`
	Description         string            `json:"description,omitempty"`
	Enabled             bool              `json:"enabled"`
	Action              string            `json:"action"` // "accept" or "drop"
	Bidirectional       bool              `json:"bidirectional"`
	Protocol            string            `json:"protocol"` // tcp, udp, icmp, all
	Ports               []string          `json:"ports,omitempty"`
	PortRanges          []PortRange       `json:"port_ranges,omitempty"`
	Sources             []PolicyGroup     `json:"sources,omitempty"`
	Destinations        []PolicyGroup     `json:"destinations,omitempty"`
	SourceResource      *PolicyResource   `json:"sourceResource,omitempty"`
	DestinationResource *PolicyResource   `json:"destinationResource,omitempty"`
}

// PortRange represents a port range for policy rules
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// PolicyResource represents a resource in policy rules
type PolicyResource struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// PolicyCreateRequest represents the request body for creating a policy
type PolicyCreateRequest struct {
	Name                string       `json:"name"`
	Description         string       `json:"description,omitempty"`
	Enabled             bool         `json:"enabled"`
	Rules               []PolicyRule `json:"rules,omitempty"`
	SourcePostureChecks []string     `json:"source_posture_checks,omitempty"`
}

// PolicyUpdateRequest represents the request body for updating a policy
type PolicyUpdateRequest struct {
	Name                string       `json:"name"`
	Description         string       `json:"description"`
	Enabled             bool         `json:"enabled"`
	Rules               []PolicyRule `json:"rules"`
	SourcePostureChecks []string     `json:"source_posture_checks,omitempty"`
}
