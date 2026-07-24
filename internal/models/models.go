// Package models defines all data types for the NetBird Management CLI
package models

// Config holds the client configuration
type Config struct {
	Token         string `json:"token"`
	ManagementURL string `json:"management_url"`
}

// Peer represents a single NetBird peer (from peers.mdx)
type Peer struct {
	ID                          string          `json:"id"`
	Name                        string          `json:"name"`
	IP                          string          `json:"ip"`
	IPv6                        string          `json:"ipv6,omitempty"`
	ConnectionIP                string          `json:"connection_ip,omitempty"`
	Connected                   bool            `json:"connected"`
	LastSeen                    string          `json:"last_seen"`
	OS                          string          `json:"os"`
	KernelVersion               string          `json:"kernel_version,omitempty"`
	GeonameID                   int             `json:"geoname_id,omitempty"`
	Version                     string          `json:"version"`
	UIVersion                   string          `json:"ui_version,omitempty"`
	Groups                      []PolicyGroup   `json:"groups"` // This uses the simplified group object
	Hostname                    string          `json:"hostname"`
	DNSLabel                    string          `json:"dns_label,omitempty"`
	ExtraDNSLabels              []string        `json:"extra_dns_labels,omitempty"`
	UserID                      string          `json:"user_id,omitempty"`
	SSHEnabled                  bool            `json:"ssh_enabled"`
	LoginExpirationEnabled      bool            `json:"login_expiration_enabled"`
	LoginExpired                bool            `json:"login_expired,omitempty"`
	LastLogin                   string          `json:"last_login,omitempty"`
	InactivityExpirationEnabled bool            `json:"inactivity_expiration_enabled"`
	ApprovalRequired            *bool           `json:"approval_required,omitempty"`  // Optional, cloud-only
	DisapprovalReason           string          `json:"disapproval_reason,omitempty"` // Cloud-only
	CountryCode                 string          `json:"country_code,omitempty"`
	CityName                    string          `json:"city_name,omitempty"`
	SerialNumber                string          `json:"serial_number,omitempty"`
	Ephemeral                   bool            `json:"ephemeral,omitempty"`
	CreatedAt                   string          `json:"created_at,omitempty"`
	AccessiblePeersCount        int             `json:"accessible_peers_count,omitempty"`
	LocalFlags                  *PeerLocalFlags `json:"local_flags,omitempty"`
}

// PeerLocalFlags represents the local client flags reported by a peer
type PeerLocalFlags struct {
	RosenpassEnabled      bool `json:"rosenpass_enabled"`
	RosenpassPermissive   bool `json:"rosenpass_permissive"`
	ServerSSHAllowed      bool `json:"server_ssh_allowed"`
	DisableClientRoutes   bool `json:"disable_client_routes"`
	DisableServerRoutes   bool `json:"disable_server_routes"`
	DisableDNS            bool `json:"disable_dns"`
	DisableFirewall       bool `json:"disable_firewall"`
	BlockLANAccess        bool `json:"block_lan_access"`
	BlockInbound          bool `json:"block_inbound"`
	LazyConnectionEnabled bool `json:"lazy_connection_enabled"`
}

// PeerUpdateRequest represents the request body for updating a peer
type PeerUpdateRequest struct {
	Name                        string `json:"name"`
	SSHEnabled                  bool   `json:"ssh_enabled"`
	LoginExpirationEnabled      bool   `json:"login_expiration_enabled"`
	InactivityExpirationEnabled bool   `json:"inactivity_expiration_enabled"`
	ApprovalRequired            *bool  `json:"approval_required,omitempty"`
	IP                          string `json:"ip,omitempty"`
	IPv6                        string `json:"ipv6,omitempty"`
}

// TemporaryAccessRequest represents the request body for POST /peers/{peerId}/temporary-access
type TemporaryAccessRequest struct {
	Name     string   `json:"name"`
	WGPubKey string   `json:"wg_pub_key"`
	Rules    []string `json:"rules"`
}

// TemporaryAccessPeer represents the response for a temporary access peer
type TemporaryAccessPeer struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Rules []string `json:"rules"`
}

// PolicyGroup represents the simplified group object found inside other resources (like Peer)
type PolicyGroup struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	PeersCount     int    `json:"peers_count,omitempty"`
	ResourcesCount int    `json:"resources_count,omitempty"`
	Issued         string `json:"issued,omitempty"`
}

// GroupPeer represents the minimal peer object returned inside a group ({id, name} only)
type GroupPeer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GroupDetail represents the full group object (from groups.mdx)
type GroupDetail struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	PeersCount     int             `json:"peers_count"`
	ResourcesCount int             `json:"resources_count"`
	Issued         string          `json:"issued"`
	Peers          []GroupPeer     `json:"peers"` // API returns only {id, name} here
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
	ID                  string              `json:"id,omitempty"`
	Name                string              `json:"name"`
	Description         string              `json:"description,omitempty"`
	Enabled             bool                `json:"enabled"`
	Action              string              `json:"action"` // "accept" or "drop"
	Bidirectional       bool                `json:"bidirectional"`
	Protocol            string              `json:"protocol"` // tcp, udp, icmp, all
	Ports               []string            `json:"ports,omitempty"`
	PortRanges          []PortRange         `json:"port_ranges,omitempty"`
	AuthorizedGroups    map[string][]string `json:"authorized_groups,omitempty"` // Map of user group IDs to local users
	Sources             []PolicyGroup       `json:"sources,omitempty"`
	Destinations        []PolicyGroup       `json:"destinations,omitempty"`
	SourceResource      *PolicyResource     `json:"sourceResource,omitempty"`
	DestinationResource *PolicyResource     `json:"destinationResource,omitempty"`
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
	ID                  string              `json:"id,omitempty"` // Include ID for updates, omit for creates
	Name                string              `json:"name"`
	Description         string              `json:"description,omitempty"`
	Enabled             bool                `json:"enabled"`
	Action              string              `json:"action"`
	Bidirectional       bool                `json:"bidirectional"`
	Protocol            string              `json:"protocol"`
	Ports               []string            `json:"ports,omitempty"`
	PortRanges          []PortRange         `json:"port_ranges,omitempty"`
	AuthorizedGroups    map[string][]string `json:"authorized_groups,omitempty"` // Map of user group IDs to local users
	Sources             []string            `json:"sources,omitempty"`           // String IDs for updates
	Destinations        []string            `json:"destinations,omitempty"`      // String IDs for updates
	SourceResource      *PolicyResource     `json:"sourceResource,omitempty"`
	DestinationResource *PolicyResource     `json:"destinationResource,omitempty"`
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
	ID              string          `json:"id"`
	Email           string          `json:"email"`
	Name            string          `json:"name"`
	Role            string          `json:"role"`
	Status          string          `json:"status"`
	LastLogin       string          `json:"last_login"`
	AutoGroups      []string        `json:"auto_groups"`
	IsCurrent       bool            `json:"is_current,omitempty"`
	IsServiceUser   bool            `json:"is_service_user"`
	IsBlocked       bool            `json:"is_blocked"`
	PendingApproval bool            `json:"pending_approval,omitempty"`
	Issued          string          `json:"issued,omitempty"`
	IdpID           string          `json:"idp_id,omitempty"`
	Permissions     UserPermissions `json:"permissions"`
}

// UserPermissions represents user permission settings
type UserPermissions struct {
	IsRestricted bool                        `json:"is_restricted"`
	Modules      map[string]ModulePermission `json:"modules,omitempty"` // Module name -> allowed operations
}

// ModulePermission represents per-module CRUD permissions
type ModulePermission struct {
	Read   bool `json:"read"`
	Create bool `json:"create"`
	Update bool `json:"update"`
	Delete bool `json:"delete"`
}

// UserInvite represents a pending user invitation (from /users/invites)
type UserInvite struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	AutoGroups  []string `json:"auto_groups"`
	ExpiresAt   string   `json:"expires_at"`
	CreatedAt   string   `json:"created_at"`
	Expired     bool     `json:"expired"`
	InviteToken string   `json:"invite_token,omitempty"`
}

// UserInviteCreateRequest represents the request body for POST /users/invites
type UserInviteCreateRequest struct {
	Email      string   `json:"email"`
	Name       string   `json:"name"`
	Role       string   `json:"role"`
	AutoGroups []string `json:"auto_groups"`
	ExpiresIn  int      `json:"expires_in,omitempty"` // Seconds
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
	ID                  string   `json:"id"`
	NetworkID           string   `json:"network_id"`
	Network             string   `json:"network"`      // CIDR notation (e.g., "10.0.0.0/16")
	NetworkType         string   `json:"network_type"` // "IPv4", "IPv6", or "Domain"
	Domains             []string `json:"domains,omitempty"`
	Peer                string   `json:"peer,omitempty"`
	PeerGroups          []string `json:"peer_groups,omitempty"`
	Metric              int      `json:"metric"`
	Masquerade          bool     `json:"masquerade"`
	Enabled             bool     `json:"enabled"`
	Groups              []string `json:"groups"`
	AccessControlGroups []string `json:"access_control_groups,omitempty"`
	Description         string   `json:"description,omitempty"`
	KeepRoute           bool     `json:"keep_route"`
	SkipAutoApply       bool     `json:"skip_auto_apply,omitempty"` // Exit node route skips auto-application
}

// RouteRequest represents the request body for creating/updating a route
type RouteRequest struct {
	Description         string   `json:"description,omitempty"`
	NetworkID           string   `json:"network_id"`
	Network             string   `json:"network,omitempty"` // CIDR (use Network OR Domains)
	Domains             []string `json:"domains,omitempty"` // Domain-based routing (use OR Network)
	Peer                string   `json:"peer,omitempty"`
	PeerGroups          []string `json:"peer_groups,omitempty"`
	Metric              int      `json:"metric"`
	Masquerade          bool     `json:"masquerade"`
	Enabled             bool     `json:"enabled"`
	Groups              []string `json:"groups"`
	AccessControlGroups []string `json:"access_control_groups,omitempty"`
	KeepRoute           bool     `json:"keep_route"`
	SkipAutoApply       bool     `json:"skip_auto_apply,omitempty"`
}

// DNSNameserverGroup represents a DNS nameserver group
type DNSNameserverGroup struct {
	ID                   string       `json:"id"`
	Name                 string       `json:"name"`
	Description          string       `json:"description,omitempty"`
	Nameservers          []Nameserver `json:"nameservers"`
	Groups               []string     `json:"groups"`
	Domains              []string     `json:"domains,omitempty"`
	SearchDomainsEnabled bool         `json:"search_domains_enabled"`
	Primary              bool         `json:"primary"`
	Enabled              bool         `json:"enabled"`
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
	Description          string       `json:"description"` // Required by the API (may be empty)
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
	NBVersionCheck        *NBVersionCheck        `json:"nb_version_check,omitempty"`
	OSVersionCheck        *OSVersionCheck        `json:"os_version_check,omitempty"`
	GeoLocationCheck      *GeoLocationCheck      `json:"geo_location_check,omitempty"`
	PeerNetworkRangeCheck *PeerNetworkRangeCheck `json:"peer_network_range_check,omitempty"`
	ProcessCheck          *ProcessCheck          `json:"process_check,omitempty"`
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
	Description string                 `json:"description"` // Required by the API (may be empty)
	Checks      PostureCheckDefinition `json:"checks"`
}

// AuditEvent represents an audit log entry
type AuditEvent struct {
	ID             string                 `json:"id"`
	Timestamp      string                 `json:"timestamp"`
	Activity       string                 `json:"activity"`
	ActivityCode   string                 `json:"activity_code"`
	InitiatorID    string                 `json:"initiator_id"`
	InitiatorName  string                 `json:"initiator_name"`
	InitiatorEmail string                 `json:"initiator_email"`
	TargetID       string                 `json:"target_id"`
	Meta           map[string]interface{} `json:"meta"`
}

// TrafficEvent represents a network traffic event (flow)
type TrafficEvent struct {
	FlowID      string            `json:"flow_id"`
	ReporterID  string            `json:"reporter_id"`
	Source      TrafficEndpoint   `json:"source"`
	Destination TrafficEndpoint   `json:"destination"`
	User        *TrafficUser      `json:"user,omitempty"`
	Policy      *TrafficPolicy    `json:"policy,omitempty"`
	ICMP        *TrafficICMP      `json:"icmp,omitempty"`
	Protocol    int               `json:"protocol"`
	Direction   string            `json:"direction"`
	RxBytes     int64             `json:"rx_bytes"`
	RxPackets   int64             `json:"rx_packets"`
	TxBytes     int64             `json:"tx_bytes"`
	TxPackets   int64             `json:"tx_packets"`
	Events      []TrafficSubEvent `json:"events,omitempty"`
}

// TrafficEndpoint represents the source or destination of a traffic event
type TrafficEndpoint struct {
	ID          string              `json:"id"`
	Type        string              `json:"type"`
	Name        string              `json:"name"`
	GeoLocation *TrafficGeoLocation `json:"geo_location,omitempty"`
	OS          string              `json:"os,omitempty"`
	Address     string              `json:"address"`
	DNSLabel    string              `json:"dns_label,omitempty"`
}

// TrafficGeoLocation represents the geo location of a traffic endpoint
type TrafficGeoLocation struct {
	CityName    string `json:"city_name,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

// TrafficUser represents the user associated with a traffic event
type TrafficUser struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// TrafficPolicy represents the policy that allowed a traffic event
type TrafficPolicy struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// TrafficICMP represents ICMP type/code info for a traffic event
type TrafficICMP struct {
	Type int `json:"type"`
	Code int `json:"code"`
}

// TrafficSubEvent represents a lifecycle event within a traffic flow
type TrafficSubEvent struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
}

// AuditEventFilters for filtering audit events
type AuditEventFilters struct {
	UserID       string
	TargetID     string
	ActivityCode string
	StartDate    string
	EndDate      string
	Search       string
}

// TrafficEventFilters for filtering traffic events
type TrafficEventFilters struct {
	Page           int
	PageSize       int
	UserID         string
	ReporterID     string
	Protocol       int
	Type           string
	ConnectionType string
	Direction      string
	Search         string
	StartDate      string
	EndDate        string
}

// TrafficEventResponse for paginated traffic events
type TrafficEventResponse struct {
	Data         []TrafficEvent `json:"data"`
	Page         int            `json:"page"`
	PageSize     int            `json:"page_size"`
	TotalRecords int            `json:"total_records"`
	TotalPages   int            `json:"total_pages"`
}

// ProxyEvent represents a reverse proxy access log entry (from /events/proxy)
type ProxyEvent struct {
	ID              string            `json:"id"`
	ServiceID       string            `json:"service_id"`
	Timestamp       string            `json:"timestamp"`
	Method          string            `json:"method"`
	Host            string            `json:"host"`
	Path            string            `json:"path"`
	DurationMS      int               `json:"duration_ms"`
	StatusCode      int               `json:"status_code"`
	SourceIP        string            `json:"source_ip"`
	Reason          string            `json:"reason,omitempty"`
	UserID          string            `json:"user_id,omitempty"`
	AuthMethodUsed  string            `json:"auth_method_used,omitempty"`
	CountryCode     string            `json:"country_code,omitempty"`
	CityName        string            `json:"city_name,omitempty"`
	SubdivisionCode string            `json:"subdivision_code,omitempty"`
	BytesUpload     int64             `json:"bytes_upload"`
	BytesDownload   int64             `json:"bytes_download"`
	Protocol        string            `json:"protocol,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// ProxyEventResponse for paginated reverse proxy access logs
type ProxyEventResponse struct {
	Data         []ProxyEvent `json:"data"`
	Page         int          `json:"page"`
	PageSize     int          `json:"page_size"`
	TotalRecords int          `json:"total_records"`
	TotalPages   int          `json:"total_pages"`
}

// ProxyEventFilters for filtering reverse proxy access logs
type ProxyEventFilters struct {
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
	Search     string
	SourceIP   string
	Host       string
	Path       string
	UserID     string
	UserEmail  string
	UserName   string
	Method     string
	Status     string
	StatusCode string
	StartDate  string
	EndDate    string
}

// City represents a city location
type City struct {
	GeonameID int    `json:"geoname_id"`
	CityName  string `json:"city_name"`
}

// Account represents a NetBird account
type Account struct {
	ID             string             `json:"id"`
	Settings       AccountSettings    `json:"settings"`
	Domain         string             `json:"domain"`
	DomainCategory string             `json:"domain_category,omitempty"`
	CreatedBy      string             `json:"created_by"`
	CreatedAt      string             `json:"created_at"`
	Onboarding     *AccountOnboarding `json:"onboarding,omitempty"`
}

// AccountSettings contains account-wide configuration
type AccountSettings struct {
	PeerLoginExpirationEnabled      bool                  `json:"peer_login_expiration_enabled"`
	PeerLoginExpiration             int                   `json:"peer_login_expiration"` // Seconds
	PeerInactivityExpirationEnabled bool                  `json:"peer_inactivity_expiration_enabled"`
	PeerInactivityExpiration        int                   `json:"peer_inactivity_expiration"` // Seconds
	DNSDomain                       string                `json:"dns_domain,omitempty"`
	NetworkRange                    string                `json:"network_range,omitempty"`
	NetworkRangeV6                  string                `json:"network_range_v6,omitempty"`
	RoutingPeerDNSResolutionEnabled bool                  `json:"routing_peer_dns_resolution_enabled"`
	JWTGroupsEnabled                bool                  `json:"jwt_groups_enabled"`
	JWTGroupsClaimName              string                `json:"jwt_groups_claim_name,omitempty"`
	JWTAllowGroups                  []string              `json:"jwt_allow_groups,omitempty"`
	GroupsPropagationEnabled        bool                  `json:"groups_propagation_enabled"`
	RegularUsersViewBlocked         bool                  `json:"regular_users_view_blocked"`
	PeerExposeEnabled               bool                  `json:"peer_expose_enabled"`
	PeerExposeGroups                []string              `json:"peer_expose_groups"`
	LazyConnectionEnabled           bool                  `json:"lazy_connection_enabled"`
	AutoUpdateVersion               string                `json:"auto_update_version,omitempty"`
	AutoUpdateAlways                bool                  `json:"auto_update_always"`
	MetricsPushEnabled              bool                  `json:"metrics_push_enabled"`
	EmbeddedIdpEnabled              bool                  `json:"embedded_idp_enabled,omitempty"` // Read-only
	LocalAuthDisabled               bool                  `json:"local_auth_disabled,omitempty"`  // Read-only
	LocalMFAEnabled                 bool                  `json:"local_mfa_enabled"`
	IPv6EnabledGroups               []string              `json:"ipv6_enabled_groups,omitempty"`
	Extra                           *AccountSettingsExtra `json:"extra,omitempty"` // Cloud-only
}

// AccountSettingsExtra contains Cloud-only extra account settings
type AccountSettingsExtra struct {
	PeerApprovalEnabled                bool     `json:"peer_approval_enabled"`
	UserApprovalRequired               bool     `json:"user_approval_required"`
	NetworkTrafficLogsEnabled          bool     `json:"network_traffic_logs_enabled"`
	NetworkTrafficLogsGroups           []string `json:"network_traffic_logs_groups"`
	NetworkTrafficPacketCounterEnabled bool     `json:"network_traffic_packet_counter_enabled"`
}

// AccountOnboarding tracks signup and onboarding progress
type AccountOnboarding struct {
	SignupFormPending     bool `json:"signup_form_pending"`
	OnboardingFlowPending bool `json:"onboarding_flow_pending"`
}

// AccountUpdateRequest for PUT /accounts/{id}
type AccountUpdateRequest struct {
	Settings   AccountSettings    `json:"settings"`
	Onboarding *AccountOnboarding `json:"onboarding,omitempty"`
}

// IngressPortAllocation represents an ingress port allocation for a peer
type IngressPortAllocation struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	IngressPeerID     string             `json:"ingress_peer_id,omitempty"`
	Region            string             `json:"region,omitempty"`
	Enabled           bool               `json:"enabled"`
	IngressIP         string             `json:"ingress_ip,omitempty"`
	PortRangeMappings []PortRangeMapping `json:"port_range_mappings,omitempty"`
}

// PortRangeMapping maps a translated (peer) port range to an ingress port range
type PortRangeMapping struct {
	TranslatedStart int    `json:"translated_start"`
	TranslatedEnd   int    `json:"translated_end"`
	IngressStart    int    `json:"ingress_start"`
	IngressEnd      int    `json:"ingress_end"`
	Protocol        string `json:"protocol"` // "tcp", "udp", or "tcp/udp"
}

// IngressPortRange represents a requested port range for an allocation
type IngressPortRange struct {
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Protocol string `json:"protocol"` // "tcp", "udp", or "tcp/udp"
}

// IngressDirectPort requests direct (1:1) port mappings for an allocation
type IngressDirectPort struct {
	Count    int    `json:"count"`
	Protocol string `json:"protocol"` // "tcp", "udp", or "tcp/udp"
}

// IngressPortAllocationRequest for POST/PUT /peers/{id}/ingress/ports[/{allocationId}]
type IngressPortAllocationRequest struct {
	Name       string             `json:"name"`
	Enabled    bool               `json:"enabled"`
	PortRanges []IngressPortRange `json:"port_ranges,omitempty"`
	DirectPort *IngressDirectPort `json:"direct_port,omitempty"`
}

// IngressPeer represents a peer acting as an ingress gateway
type IngressPeer struct {
	ID             string                `json:"id"`
	PeerID         string                `json:"peer_id"`
	IngressIP      string                `json:"ingress_ip,omitempty"`
	AvailablePorts *IngressPeerPortsInfo `json:"available_ports,omitempty"`
	Enabled        bool                  `json:"enabled"`
	Connected      bool                  `json:"connected"`
	Fallback       bool                  `json:"fallback"`
	Region         string                `json:"region,omitempty"`
}

// IngressPeerPortsInfo reports available ports on an ingress peer
type IngressPeerPortsInfo struct {
	TCP int `json:"tcp"`
	UDP int `json:"udp"`
}

// IngressPeerCreateRequest for POST /ingress/peers
type IngressPeerCreateRequest struct {
	PeerID   string `json:"peer_id"`
	Enabled  bool   `json:"enabled"`
	Fallback bool   `json:"fallback"`
}

// IngressPeerUpdateRequest for PUT /ingress/peers/{id}
type IngressPeerUpdateRequest struct {
	Enabled  bool `json:"enabled"`
	Fallback bool `json:"fallback"`
}

// DNSZone represents a custom DNS zone (from dns-zones.mdx)
type DNSZone struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	Domain             string      `json:"domain"`
	Enabled            bool        `json:"enabled"`
	EnableSearchDomain bool        `json:"enable_search_domain"`
	DistributionGroups []string    `json:"distribution_groups"`
	Records            []DNSRecord `json:"records,omitempty"`
}

// DNSZoneRequest for POST/PUT /dns/zones[/{zoneId}]
type DNSZoneRequest struct {
	Name               string   `json:"name"`
	Domain             string   `json:"domain"`
	Enabled            bool     `json:"enabled"`
	EnableSearchDomain bool     `json:"enable_search_domain"`
	DistributionGroups []string `json:"distribution_groups"`
}

// DNSRecord represents a record in a custom DNS zone
type DNSRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"` // "A", "AAAA", or "CNAME"
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// DNSRecordRequest for POST/PUT /dns/zones/{zoneId}/records[/{recordId}]
type DNSRecordRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // "A", "AAAA", or "CNAME"
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// PeerJob represents an asynchronous job run against a peer (from jobs.mdx)
type PeerJob struct {
	ID           string      `json:"id"`
	CreatedAt    string      `json:"created_at"`
	CompletedAt  string      `json:"completed_at,omitempty"`
	TriggeredBy  string      `json:"triggered_by,omitempty"`
	Status       string      `json:"status"`
	FailedReason string      `json:"failed_reason,omitempty"`
	Workload     JobWorkload `json:"workload"`
}

// JobWorkload describes the work a peer job performs (currently only "bundle")
type JobWorkload struct {
	Type       string               `json:"type"` // e.g. "bundle"
	Parameters *BundleJobParameters `json:"parameters,omitempty"`
	Result     *JobWorkloadResult   `json:"result,omitempty"`
}

// BundleJobParameters configures a debug bundle collection job
type BundleJobParameters struct {
	BundleFor     bool `json:"bundle_for"`
	BundleForTime int  `json:"bundle_for_time"`
	LogFileCount  int  `json:"log_file_count"`
	Anonymize     bool `json:"anonymize"`
}

// JobWorkloadResult holds the output of a completed job
type JobWorkloadResult struct {
	UploadKey string `json:"upload_key,omitempty"`
}

// PeerJobCreateRequest for POST /peers/{peerId}/jobs
type PeerJobCreateRequest struct {
	Workload JobWorkload `json:"workload"`
}

// NotificationChannel represents a notification channel (from notifications.mdx)
type NotificationChannel struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"` // "email" or "webhook"
	Target     *NotificationTarget `json:"target,omitempty"`
	EventTypes []string            `json:"event_types"`
	Enabled    bool                `json:"enabled"`
}

// NotificationTarget holds the type-dependent target of a notification channel.
// For type "email" only Emails is set; for type "webhook" URL (and optionally Headers) are set.
type NotificationTarget struct {
	Emails  []string          `json:"emails,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"` // Write-only; masked in responses
}

// NotificationChannelRequest for POST/PUT /integrations/notifications/channels[/{channelId}]
type NotificationChannelRequest struct {
	Type       string              `json:"type"` // "email" or "webhook"
	Target     *NotificationTarget `json:"target,omitempty"`
	EventTypes []string            `json:"event_types"`
	Enabled    bool                `json:"enabled"`
}
