package tui

import (
	"netbird-manage/internal/models"

	"github.com/netbirdio/netbird/client/proto"
)

// Navigation messages

// NavSelectMsg is sent when a navigation item is selected
type NavSelectMsg struct {
	Section Section
}

// API response messages

// PeersLoadedMsg carries the result of fetching peers
type PeersLoadedMsg struct {
	Peers []models.Peer
	Err   error
}

// GroupsLoadedMsg carries the result of fetching groups
type GroupsLoadedMsg struct {
	Groups []models.PolicyGroup
	Err    error
}

// NetworksLoadedMsg carries the result of fetching networks
type NetworksLoadedMsg struct {
	Networks []models.Network
	Err      error
}

// PoliciesLoadedMsg carries the result of fetching policies
type PoliciesLoadedMsg struct {
	Policies []models.Policy
	Err      error
}

// RoutesLoadedMsg carries the result of fetching routes
type RoutesLoadedMsg struct {
	Routes []models.Route
	Err    error
}

// SetupKeysLoadedMsg carries the result of fetching setup keys
type SetupKeysLoadedMsg struct {
	Keys []models.SetupKey
	Err  error
}

// UsersLoadedMsg carries the result of fetching users
type UsersLoadedMsg struct {
	Users []models.User
	Err   error
}

// DNSGroupsLoadedMsg carries the result of fetching DNS groups
type DNSGroupsLoadedMsg struct {
	Groups []models.DNSNameserverGroup
	Err    error
}

// PostureChecksLoadedMsg carries the result of fetching posture checks
type PostureChecksLoadedMsg struct {
	Checks []models.PostureCheck
	Err    error
}

// EventsLoadedMsg carries the result of fetching events
type EventsLoadedMsg struct {
	Events []models.AuditEvent
	Err    error
}

// AccountsLoadedMsg carries the result of fetching accounts
type AccountsLoadedMsg struct {
	Accounts []models.Account
	Err      error
}

// SettingsLoadedMsg carries the result of fetching account settings
type SettingsLoadedMsg struct {
	Account models.Account
	Err     error
}

// SettingsSavedMsg carries the result of saving account settings
type SettingsSavedMsg struct {
	Account models.Account
	Err     error
}

// APIErrorMsg is a generic API error
type APIErrorMsg struct {
	Err     error
	Context string
}

// ToastMsg triggers a toast notification
type ToastMsg struct {
	Message string
	IsError bool
}

// DaemonStatusMsg carries daemon status + config
type DaemonStatusMsg struct {
	Status *proto.StatusResponse
	Config *proto.GetConfigResponse
	Err    error
}

// DaemonTickMsg triggers a daemon status refresh
type DaemonTickMsg struct{}

// PeersTickMsg triggers an auto-refresh of the peers list
type PeersTickMsg struct{}

// EventsTickMsg triggers an auto-refresh of the events list
type EventsTickMsg struct{}

// ConnectionTickMsg triggers an auto-refresh of peer connection detail
type ConnectionTickMsg struct{}

// PeerConnectionInfo holds daemon-reported connection details for a single peer
type PeerConnectionInfo struct {
	ConnType       string // "P2P", "Relayed", or "Disconnected"
	RemoteEndpoint string
	LocalICEType   string
	RemoteICEType  string
	Latency        string
	BytesSent      int64
	BytesReceived  int64
	LastHandshake  string
	RelayAddress   string
}

// PeerConnectionMsg carries per-peer connection info from the local daemon
type PeerConnectionMsg struct {
	Connections map[string]PeerConnectionInfo // keyed by NetBird IP
	Err         error
}

// Edit/update result messages

// PeerUpdatedMsg carries the result of updating a peer
type PeerUpdatedMsg struct {
	Peer models.Peer
	Err  error
}

// GroupUpdatedMsg carries the result of updating a group
type GroupUpdatedMsg struct {
	Err error
}

// PolicyUpdatedMsg carries the result of updating a policy
type PolicyUpdatedMsg struct {
	Err error
}

// RouteUpdatedMsg carries the result of updating a route
type RouteUpdatedMsg struct {
	Err error
}

// DNSUpdatedMsg carries the result of updating a DNS nameserver group
type DNSUpdatedMsg struct {
	Err error
}

// NetworkUpdatedMsg carries the result of updating a network
type NetworkUpdatedMsg struct {
	Err error
}

// UserUpdatedMsg carries the result of updating a user
type UserUpdatedMsg struct {
	Err error
}

// PostureCheckUpdatedMsg carries the result of updating a posture check
type PostureCheckUpdatedMsg struct {
	Err error
}

// PostureCheckDeletedMsg carries the result of deleting a posture check
type PostureCheckDeletedMsg struct {
	Err error
}

// BulkGroupAssignMsg carries the result of bulk group assignment
type BulkGroupAssignMsg struct {
	Err error
}

// DashboardCountsMsg carries management API counts for the dashboard
type DashboardCountsMsg struct {
	PeersOnline int
	PeersTotal  int
	Groups      int
	Networks    int
	Policies    int
	Routes      int
	SetupKeys   int
	DNSGroups   int
	Posture     int
	AccountDomain string
	Err         error
}
