// models.go
package main

// Config holds the client configuration
type Config struct {
	Token         string `json:"token"`
	ManagementURL string `json:"management_url"`
}

// Peer represents a single NetBird peer (from peers.mdx)
type Peer struct {
	ID                          string        `json:"id"`
	Name                        string        `json:"name"`
	IP                          string        `json:"ip"`
	Connected                   bool          `json:"connected"`
	LastSeen                    string        `json:"last_seen"`
	OS                          string        `json:"os"`
	Version                     string        `json:"version"`
	Groups                      []PolicyGroup `json:"groups"` // This uses the simplified group object
	Hostname                    string        `json:"hostname"`
	SSHEnabled                  bool          `json:"ssh_enabled"`
	LoginExpirationEnabled      bool          `json:"login_expiration_enabled"`
	InactivityExpirationEnabled bool          `json:"inactivity_expiration_enabled"`
	ApprovalRequired            *bool         `json:"approval_required,omitempty"` // Optional, cloud-only
}

// PeerUpdateRequest represents the request body for updating a peer
type PeerUpdateRequest struct {
	Name                        string `json:"name"`
	SSHEnabled                  bool   `json:"ssh_enabled"`
	LoginExpirationEnabled      bool   `json:"login_expiration_enabled"`
	InactivityExpirationEnabled bool   `json:"inactivity_expiration_enabled"`
	ApprovalRequired            *bool  `json:"approval_required,omitempty"`
	IP                          string `json:"ip,omitempty"`
}

// PolicyGroup represents the simplified group object found inside other resources (like Peer)
type PolicyGroup struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	PeersCount     int    `json:"peers_count,omitempty"`
	ResourcesCount int    `json:"resources_count,omitempty"`
}

// GroupDetail represents the full group object (from groups.mdx)
type GroupDetail struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	PeersCount     int             `json:"peers_count"`
	ResourcesCount int             `json:"resources_count"`
	Issued         string          `json:"issued"`
	Peers          []Peer          `json:"peers"` // Contains list of full Peer objects
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
	Routers           []string `json:"routers"` // Router IDs
	RoutingPeersCount int      `json:"routing_peers_count"`
	Resources         []string `json:"resources"` // Resource IDs
	Policies          []string `json:"policies"`  // Policy IDs
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
	ID                  string          `json:"id,omitempty"`
	Name                string          `json:"name"`
	Description         string          `json:"description,omitempty"`
	Enabled             bool            `json:"enabled"`
	Action              string          `json:"action"` // "accept" or "drop"
	Bidirectional       bool            `json:"bidirectional"`
	Protocol            string          `json:"protocol"` // tcp, udp, icmp, all
	Ports               []string        `json:"ports,omitempty"`
	PortRanges          []PortRange     `json:"port_ranges,omitempty"`
	Sources             []PolicyGroup   `json:"sources,omitempty"`
	Destinations        []PolicyGroup   `json:"destinations,omitempty"`
	SourceResource      *PolicyResource `json:"sourceResource,omitempty"`
	DestinationResource *PolicyResource `json:"destinationResource,omitempty"`
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
	Name                string               `json:"name"`
	Description         string               `json:"description,omitempty"`
	Enabled             bool                 `json:"enabled"`
	Rules               []PolicyRuleForWrite `json:"rules,omitempty"`
	SourcePostureChecks []string             `json:"source_posture_checks,omitempty"`
}

// PolicyUpdateRequest represents the request body for updating a policy
type PolicyUpdateRequest struct {
	Name                string               `json:"name"`
	Description         string               `json:"description,omitempty"`
	Enabled             bool                 `json:"enabled"`
	Rules               []PolicyRuleForWrite `json:"rules"`
	SourcePostureChecks []string             `json:"source_posture_checks,omitempty"`
}

// PolicyRuleForWrite represents a policy rule for create/update operations (uses string IDs instead of objects)
type PolicyRuleForWrite struct {
	Name                string          `json:"name"`
	Description         string          `json:"description,omitempty"`
	Enabled             bool            `json:"enabled"`
	Action              string          `json:"action"`
	Bidirectional       bool            `json:"bidirectional"`
	Protocol            string          `json:"protocol"`
	Ports               []string        `json:"ports,omitempty"`
	PortRanges          []PortRange     `json:"port_ranges,omitempty"`
	Sources             []string        `json:"sources,omitempty"`             // String IDs for updates
	Destinations        []string        `json:"destinations,omitempty"`        // String IDs for updates
	SourceResource      *PolicyResource `json:"sourceResource,omitempty"`
	DestinationResource *PolicyResource `json:"destinationResource,omitempty"`
}

// SetupKey represents a setup key for peer registration
type SetupKey struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Expires             string   `json:"expires"`
	Type                string   `json:"type"` // "one-off" or "reusable"
	Valid               bool     `json:"valid"`
	Revoked             bool     `json:"revoked"`
	UsedTimes           int      `json:"used_times"`
	LastUsed            string   `json:"last_used"`
	State               string   `json:"state"`
	AutoGroups          []string `json:"auto_groups"`
	UpdatedAt           string   `json:"updated_at"`
	UsageLimit          int      `json:"usage_limit"`
	Ephemeral           bool     `json:"ephemeral"`
	AllowExtraDNSLabels bool     `json:"allow_extra_dns_labels"`
	Key                 string   `json:"key,omitempty"` // Only in create response
}

// SetupKeyCreateRequest represents the request body for creating a setup key
type SetupKeyCreateRequest struct {
	Name                string   `json:"name"`
	Type                string   `json:"type"` // "one-off" or "reusable"
	ExpiresIn           int      `json:"expires_in"`
	AutoGroups          []string `json:"auto_groups"`
	UsageLimit          int      `json:"usage_limit"`
	Ephemeral           bool     `json:"ephemeral,omitempty"`
	AllowExtraDNSLabels bool     `json:"allow_extra_dns_labels,omitempty"`
}

// SetupKeyUpdateRequest represents the request body for updating a setup key
type SetupKeyUpdateRequest struct {
	Revoked    bool     `json:"revoked"`
	AutoGroups []string `json:"auto_groups"`
}

// User represents a NetBird user account
type User struct {
	ID            string          `json:"id"`
	Email         string          `json:"email"`
	Name          string          `json:"name"`
	Role          string          `json:"role"`
	Status        string          `json:"status"`
	LastLogin     string          `json:"last_login"`
	AutoGroups    []string        `json:"auto_groups"`
	IsServiceUser bool            `json:"is_service_user"`
	IsBlocked     bool            `json:"is_blocked"`
	Permissions   UserPermissions `json:"permissions"`
}

// UserPermissions represents user permission settings
type UserPermissions struct {
	DashboardView string `json:"dashboard_view"`
}

// UserCreateRequest represents the request body for creating/inviting a user
type UserCreateRequest struct {
	Email         string   `json:"email,omitempty"`
	Name          string   `json:"name,omitempty"`
	Role          string   `json:"role"`
	AutoGroups    []string `json:"auto_groups"`
	IsServiceUser bool     `json:"is_service_user"`
}

// UserUpdateRequest represents the request body for updating a user
type UserUpdateRequest struct {
	Role       string   `json:"role"`
	AutoGroups []string `json:"auto_groups"`
	IsBlocked  bool     `json:"is_blocked"`
}

// PersonalAccessToken represents a personal access token
type PersonalAccessToken struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ExpirationDate string `json:"expiration_date"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      string `json:"created_at"`
	LastUsed       string `json:"last_used"`
}

// TokenCreateRequest represents the request body for creating a token
type TokenCreateRequest struct {
	Name      string `json:"name"`
	ExpiresIn int    `json:"expires_in"` // Days (1-365)
}

// TokenCreateResponse represents the response when creating a token
type TokenCreateResponse struct {
	PlainToken          string              `json:"plain_token"`
	PersonalAccessToken PersonalAccessToken `json:"personal_access_token"`
}

// Route represents a network route
type Route struct {
	ID          string   `json:"id"`
	NetworkID   string   `json:"network_id"`
	Network     string   `json:"network"`      // CIDR notation (e.g., "10.0.0.0/16")
	NetworkType string   `json:"network_type"` // "IPv4" or "IPv6"
	Peer        string   `json:"peer,omitempty"`
	PeerGroups  []string `json:"peer_groups,omitempty"`
	Metric      int      `json:"metric"`
	Masquerade  bool     `json:"masquerade"`
	Enabled     bool     `json:"enabled"`
	Groups      []string `json:"groups"`
	Description string   `json:"description,omitempty"`
}

// RouteRequest represents the request body for creating/updating a route
type RouteRequest struct {
	Description string   `json:"description,omitempty"`
	NetworkID   string   `json:"network_id"`
	Network     string   `json:"network"` // CIDR
	Peer        string   `json:"peer,omitempty"`
	PeerGroups  []string `json:"peer_groups,omitempty"`
	Metric      int      `json:"metric"`
	Masquerade  bool     `json:"masquerade"`
	Enabled     bool     `json:"enabled"`
	Groups      []string `json:"groups"`
}

// DNSNameserverGroup represents a DNS nameserver group
type DNSNameserverGroup struct {
	ID                   string        `json:"id"`
	Name                 string        `json:"name"`
	Description          string        `json:"description,omitempty"`
	Nameservers          []Nameserver  `json:"nameservers"`
	Groups               []string      `json:"groups"`
	Domains              []string      `json:"domains,omitempty"`
	SearchDomainsEnabled bool          `json:"search_domains_enabled"`
	Primary              bool          `json:"primary"`
	Enabled              bool          `json:"enabled"`
}

// Nameserver represents a DNS nameserver
type Nameserver struct {
	IP     string `json:"ip"`
	NSType string `json:"ns_type"` // "udp" or "tcp"
	Port   int    `json:"port"`
}

// DNSNameserverGroupRequest represents the request body for creating/updating a DNS nameserver group
type DNSNameserverGroupRequest struct {
	Name                 string       `json:"name"`
	Description          string       `json:"description,omitempty"`
	Nameservers          []Nameserver `json:"nameservers"`
	Groups               []string     `json:"groups"`
	Domains              []string     `json:"domains,omitempty"`
	SearchDomainsEnabled bool         `json:"search_domains_enabled"`
	Primary              bool         `json:"primary"`
	Enabled              bool         `json:"enabled"`
}

// DNSSettings represents DNS settings for the account
type DNSSettings struct {
	DisabledManagementGroups []string `json:"disabled_management_groups"`
}

// PostureCheck represents a device posture check
type PostureCheck struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Checks      PostureCheckDefinition `json:"checks"`
}

// PostureCheckDefinition contains the actual check definitions
type PostureCheckDefinition struct {
	NBVersionCheck         *NBVersionCheck         `json:"nb_version_check,omitempty"`
	OSVersionCheck         *OSVersionCheck         `json:"os_version_check,omitempty"`
	GeoLocationCheck       *GeoLocationCheck       `json:"geo_location_check,omitempty"`
	PeerNetworkRangeCheck  *PeerNetworkRangeCheck  `json:"peer_network_range_check,omitempty"`
	ProcessCheck           *ProcessCheck           `json:"process_check,omitempty"`
}

// NBVersionCheck checks NetBird version
type NBVersionCheck struct {
	MinVersion string `json:"min_version"`
}

// OSVersionCheck checks operating system version
type OSVersionCheck struct {
	Android *MinVersionConfig       `json:"android,omitempty"`
	Darwin  *MinVersionConfig       `json:"darwin,omitempty"`
	IOS     *MinVersionConfig       `json:"ios,omitempty"`
	Linux   *MinKernelVersionConfig `json:"linux,omitempty"`
	Windows *MinKernelVersionConfig `json:"windows,omitempty"`
}

// MinVersionConfig represents minimum version configuration
type MinVersionConfig struct {
	MinVersion string `json:"min_version"`
}

// MinKernelVersionConfig represents minimum kernel version configuration
type MinKernelVersionConfig struct {
	MinKernelVersion string `json:"min_kernel_version"`
}

// GeoLocationCheck checks geographic location
type GeoLocationCheck struct {
	Locations []Location `json:"locations"`
	Action    string     `json:"action"` // "allow" or "deny"
}

// Location represents a geographic location
type Location struct {
	CountryCode string `json:"country_code"` // ISO 3166-1 alpha-2
	CityName    string `json:"city_name,omitempty"`
}

// PeerNetworkRangeCheck checks peer network ranges
type PeerNetworkRangeCheck struct {
	Ranges []string `json:"ranges"` // CIDR ranges
	Action string   `json:"action"` // "allow" or "deny"
}

// ProcessCheck checks for running processes
type ProcessCheck struct {
	Processes []Process `json:"processes"`
}

// Process represents a process to check for
type Process struct {
	LinuxPath   string `json:"linux_path,omitempty"`
	MacPath     string `json:"mac_path,omitempty"`
	WindowsPath string `json:"windows_path,omitempty"`
}

// PostureCheckRequest represents the request body for creating/updating a posture check
type PostureCheckRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Checks      PostureCheckDefinition `json:"checks"`
}
