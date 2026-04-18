package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	tea "charm.land/bubbletea/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// FetchPeerConnections returns a tea.Cmd that loads per-peer connection info from the local daemon
func FetchPeerConnections(daemon *DaemonClient) tea.Cmd {
	return func() tea.Msg {
		if daemon == nil {
			return PeerConnectionMsg{Err: fmt.Errorf("daemon not connected")}
		}
		conns, err := daemon.PeerConnections()
		return PeerConnectionMsg{Connections: conns, Err: err}
	}
}

// fetchPeers returns a tea.Cmd that loads all peers
func FetchPeers(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/peers", nil)
		if err != nil {
			return PeersLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var peers []models.Peer
		if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
			return PeersLoadedMsg{Err: fmt.Errorf("decode peers: %w", err)}
		}
		return PeersLoadedMsg{Peers: peers}
	}
}

// fetchGroups returns a tea.Cmd that loads all groups
func FetchGroups(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/groups", nil)
		if err != nil {
			return GroupsLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var groups []models.PolicyGroup
		if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
			return GroupsLoadedMsg{Err: fmt.Errorf("decode groups: %w", err)}
		}
		return GroupsLoadedMsg{Groups: groups}
	}
}

// fetchNetworks returns a tea.Cmd that loads all networks
func FetchNetworks(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/networks", nil)
		if err != nil {
			return NetworksLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var networks []models.Network
		if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
			return NetworksLoadedMsg{Err: fmt.Errorf("decode networks: %w", err)}
		}
		return NetworksLoadedMsg{Networks: networks}
	}
}

// fetchPolicies returns a tea.Cmd that loads all policies
func FetchPolicies(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/policies", nil)
		if err != nil {
			return PoliciesLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var policies []models.Policy
		if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
			return PoliciesLoadedMsg{Err: fmt.Errorf("decode policies: %w", err)}
		}
		return PoliciesLoadedMsg{Policies: policies}
	}
}

// fetchRoutes returns a tea.Cmd that loads all routes
func FetchRoutes(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/routes", nil)
		if err != nil {
			return RoutesLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var routes []models.Route
		if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
			return RoutesLoadedMsg{Err: fmt.Errorf("decode routes: %w", err)}
		}
		return RoutesLoadedMsg{Routes: routes}
	}
}

// fetchSetupKeys returns a tea.Cmd that loads all setup keys
func FetchSetupKeys(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/setup-keys", nil)
		if err != nil {
			return SetupKeysLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var keys []models.SetupKey
		if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
			return SetupKeysLoadedMsg{Err: fmt.Errorf("decode setup keys: %w", err)}
		}
		return SetupKeysLoadedMsg{Keys: keys}
	}
}

// fetchUsers returns a tea.Cmd that loads all users
func FetchUsers(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/users", nil)
		if err != nil {
			return UsersLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var users []models.User
		if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
			return UsersLoadedMsg{Err: fmt.Errorf("decode users: %w", err)}
		}
		return UsersLoadedMsg{Users: users}
	}
}

// fetchDNSGroups returns a tea.Cmd that loads all DNS nameserver groups
func FetchDNSGroups(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/dns/nameservers", nil)
		if err != nil {
			return DNSGroupsLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var groups []models.DNSNameserverGroup
		if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
			return DNSGroupsLoadedMsg{Err: fmt.Errorf("decode dns groups: %w", err)}
		}
		return DNSGroupsLoadedMsg{Groups: groups}
	}
}

// fetchPostureChecks returns a tea.Cmd that loads all posture checks
func FetchPostureChecks(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/posture-checks", nil)
		if err != nil {
			return PostureChecksLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var checks []models.PostureCheck
		if err := json.NewDecoder(resp.Body).Decode(&checks); err != nil {
			return PostureChecksLoadedMsg{Err: fmt.Errorf("decode posture checks: %w", err)}
		}
		return PostureChecksLoadedMsg{Checks: checks}
	}
}

// fetchEvents returns a tea.Cmd that loads audit events
func FetchEvents(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/events/audit", nil)
		if err != nil {
			return EventsLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var events []models.AuditEvent
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			return EventsLoadedMsg{Err: fmt.Errorf("decode events: %w", err)}
		}
		return EventsLoadedMsg{Events: events}
	}
}

// fetchAccounts returns a tea.Cmd that loads accounts
func FetchAccounts(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/accounts", nil)
		if err != nil {
			return AccountsLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var accounts []models.Account
		if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
			return AccountsLoadedMsg{Err: fmt.Errorf("decode accounts: %w", err)}
		}
		return AccountsLoadedMsg{Accounts: accounts}
	}
}

// FetchSettings returns a tea.Cmd that loads the first account (settings)
func FetchSettings(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/accounts", nil)
		if err != nil {
			return SettingsLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var accounts []models.Account
		if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
			return SettingsLoadedMsg{Err: fmt.Errorf("decode accounts: %w", err)}
		}
		if len(accounts) == 0 {
			return SettingsLoadedMsg{Err: fmt.Errorf("no accounts found")}
		}
		return SettingsLoadedMsg{Account: accounts[0]}
	}
}

// SaveSettings returns a tea.Cmd that PUTs updated account settings
func SaveSettings(c *client.Client, accountID string, settings models.AccountSettings) tea.Cmd {
	return func() tea.Msg {
		req := models.AccountUpdateRequest{
			Settings: settings,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return SettingsSavedMsg{Err: fmt.Errorf("marshal settings: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/accounts/"+url.PathEscape(accountID), bytes.NewReader(body))
		if err != nil {
			return SettingsSavedMsg{Err: err}
		}
		defer resp.Body.Close()

		var account models.Account
		if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
			return SettingsSavedMsg{Err: fmt.Errorf("decode response: %w", err)}
		}
		return SettingsSavedMsg{Account: account}
	}
}

// BulkAssignGroup adds or removes peers from a group via full PUT
func BulkAssignGroup(c *client.Client, group models.PolicyGroup, peerIDs []string, action string) tea.Cmd {
	return func() tea.Msg {
		// Fetch full group to get current peers
		resp, err := c.MakeRequest("GET", "/groups/"+url.PathEscape(group.ID), nil)
		if err != nil {
			return BulkGroupAssignMsg{Err: fmt.Errorf("fetch group: %w", err)}
		}
		defer resp.Body.Close()

		var detail struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Peers []struct {
				ID string `json:"id"`
			} `json:"peers"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
			return BulkGroupAssignMsg{Err: fmt.Errorf("decode group: %w", err)}
		}

		// Build updated peer list
		existingIDs := make(map[string]bool)
		for _, p := range detail.Peers {
			existingIDs[p.ID] = true
		}

		if action == "add" {
			for _, id := range peerIDs {
				existingIDs[id] = true
			}
		} else {
			for _, id := range peerIDs {
				delete(existingIDs, id)
			}
		}

		type peerRef struct {
			ID string `json:"id"`
		}
		updatedPeers := make([]peerRef, 0, len(existingIDs))
		for id := range existingIDs {
			updatedPeers = append(updatedPeers, peerRef{ID: id})
		}

		updateReq := struct {
			Name  string    `json:"name"`
			Peers []peerRef `json:"peers"`
		}{
			Name:  detail.Name,
			Peers: updatedPeers,
		}

		body, err := json.Marshal(updateReq)
		if err != nil {
			return BulkGroupAssignMsg{Err: fmt.Errorf("marshal update: %w", err)}
		}
		resp2, err := c.MakeRequest("PUT", "/groups/"+url.PathEscape(group.ID), bytes.NewReader(body))
		if err != nil {
			return BulkGroupAssignMsg{Err: err}
		}
		defer resp2.Body.Close()

		return BulkGroupAssignMsg{}
	}
}

// deletePeer returns a tea.Cmd that deletes a peer
func DeletePeer(c *client.Client, peerID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/peers/"+url.PathEscape(peerID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete peer"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Peer deleted successfully"}
	}
}

// DeleteGroup returns a tea.Cmd that deletes a group
func DeleteGroup(c *client.Client, groupID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/groups/"+url.PathEscape(groupID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete group"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Group deleted successfully"}
	}
}

// ToggleRoute enables or disables a route
func ToggleRoute(c *client.Client, route models.Route, enable bool) tea.Cmd {
	return func() tea.Msg {
		req := models.RouteRequest{
			Description:         route.Description,
			NetworkID:           route.NetworkID,
			Network:             route.Network,
			Domains:             route.Domains,
			Peer:                route.Peer,
			PeerGroups:          route.PeerGroups,
			Metric:              route.Metric,
			Masquerade:          route.Masquerade,
			Enabled:             enable,
			Groups:              route.Groups,
			AccessControlGroups: route.AccessControlGroups,
			KeepRoute:           route.KeepRoute,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "toggle route"}
		}
		resp, err := c.MakeRequest("PUT", "/routes/"+url.PathEscape(route.ID), bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "toggle route"}
		}
		defer resp.Body.Close()
		action := "disabled"
		if enable {
			action = "enabled"
		}
		return ToastMsg{Message: "Route " + action}
	}
}

// TogglePolicy enables or disables a policy
func TogglePolicy(c *client.Client, policy models.Policy, enable bool) tea.Cmd {
	return func() tea.Msg {
		rules := make([]models.PolicyRuleForWrite, len(policy.Rules))
		for i, r := range policy.Rules {
			sources := make([]string, len(r.Sources))
			for j, s := range r.Sources {
				sources[j] = s.ID
			}
			dests := make([]string, len(r.Destinations))
			for j, d := range r.Destinations {
				dests[j] = d.ID
			}
			rules[i] = models.PolicyRuleForWrite{
				ID:                  r.ID,
				Name:                r.Name,
				Description:         r.Description,
				Enabled:             r.Enabled,
				Action:              r.Action,
				Bidirectional:       r.Bidirectional,
				Protocol:            r.Protocol,
				Ports:               r.Ports,
				PortRanges:          r.PortRanges,
				Sources:             sources,
				Destinations:        dests,
				SourceResource:      r.SourceResource,
				DestinationResource: r.DestinationResource,
			}
		}

		req := models.PolicyUpdateRequest{
			Name:                policy.Name,
			Description:         policy.Description,
			Enabled:             enable,
			Rules:               rules,
			SourcePostureChecks: policy.SourcePostureChecks,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "toggle policy"}
		}
		resp, err := c.MakeRequest("PUT", "/policies/"+url.PathEscape(policy.ID), bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "toggle policy"}
		}
		defer resp.Body.Close()
		action := "disabled"
		if enable {
			action = "enabled"
		}
		return ToastMsg{Message: "Policy " + action}
	}
}

// RevokeSetupKey revokes or un-revokes a setup key
func RevokeSetupKey(c *client.Client, key models.SetupKey, revoke bool) tea.Cmd {
	return func() tea.Msg {
		req := models.SetupKeyUpdateRequest{
			Revoked:    revoke,
			AutoGroups: key.AutoGroups,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "revoke setup key"}
		}
		resp, err := c.MakeRequest("PUT", "/setup-keys/"+url.PathEscape(key.ID), bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "revoke setup key"}
		}
		defer resp.Body.Close()
		action := "un-revoked"
		if revoke {
			action = "revoked"
		}
		return ToastMsg{Message: "Setup key " + action}
	}
}

// ToggleUserBlock blocks or unblocks a user
func ToggleUserBlock(c *client.Client, user models.User, block bool) tea.Cmd {
	return func() tea.Msg {
		req := models.UserUpdateRequest{
			Role:       user.Role,
			AutoGroups: user.AutoGroups,
			IsBlocked:  block,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "toggle user block"}
		}
		resp, err := c.MakeRequest("PUT", "/users/"+url.PathEscape(user.ID), bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "toggle user block"}
		}
		defer resp.Body.Close()
		action := "unblocked"
		if block {
			action = "blocked"
		}
		return ToastMsg{Message: "User " + action}
	}
}

// FetchAccessiblePeers fetches peers accessible from a given peer
func FetchAccessiblePeers(c *client.Client, peerID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/peers/"+url.PathEscape(peerID)+"/accessible-peers", nil)
		if err != nil {
			return accessiblePeersLoadedMsg{err: err}
		}
		defer resp.Body.Close()

		var peers []models.Peer
		if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
			return accessiblePeersLoadedMsg{err: fmt.Errorf("decode accessible peers: %w", err)}
		}
		return accessiblePeersLoadedMsg{peers: peers}
	}
}

// accessiblePeersLoadedMsg carries accessible peers result
type accessiblePeersLoadedMsg struct {
	peers []models.Peer
	err   error
}

// UpdatePeer returns a tea.Cmd that updates a peer
func UpdatePeer(c *client.Client, peerID string, req models.PeerUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return PeerUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/peers/"+url.PathEscape(peerID), bytes.NewReader(body))
		if err != nil {
			return PeerUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		var peer models.Peer
		if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
			return PeerUpdatedMsg{Err: fmt.Errorf("decode peer: %w", err)}
		}
		return PeerUpdatedMsg{Peer: peer}
	}
}

// UpdateGroup returns a tea.Cmd that updates a group's name
func UpdateGroup(c *client.Client, groupID string, name string) tea.Cmd {
	return func() tea.Msg {
		req := struct {
			Name string `json:"name"`
		}{Name: name}
		body, err := json.Marshal(req)
		if err != nil {
			return GroupUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/groups/"+url.PathEscape(groupID), bytes.NewReader(body))
		if err != nil {
			return GroupUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return GroupUpdatedMsg{}
	}
}

// UpdatePolicy returns a tea.Cmd that updates a policy
func UpdatePolicy(c *client.Client, policyID string, req models.PolicyUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return PolicyUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/policies/"+url.PathEscape(policyID), bytes.NewReader(body))
		if err != nil {
			return PolicyUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return PolicyUpdatedMsg{}
	}
}

// UpdateRoute returns a tea.Cmd that updates a route
func UpdateRoute(c *client.Client, routeID string, req models.RouteRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return RouteUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/routes/"+url.PathEscape(routeID), bytes.NewReader(body))
		if err != nil {
			return RouteUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return RouteUpdatedMsg{}
	}
}

// UpdateDNSGroup returns a tea.Cmd that updates a DNS nameserver group
func UpdateDNSGroup(c *client.Client, groupID string, req models.DNSNameserverGroupRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return DNSUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/dns/nameservers/"+url.PathEscape(groupID), bytes.NewReader(body))
		if err != nil {
			return DNSUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return DNSUpdatedMsg{}
	}
}

// UpdateNetwork returns a tea.Cmd that updates a network
func UpdateNetwork(c *client.Client, networkID string, req models.NetworkUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return NetworkUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/networks/"+url.PathEscape(networkID), bytes.NewReader(body))
		if err != nil {
			return NetworkUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return NetworkUpdatedMsg{}
	}
}

// UpdateUser returns a tea.Cmd that updates a user
func UpdateUser(c *client.Client, userID string, req models.UserUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return UserUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/users/"+url.PathEscape(userID), bytes.NewReader(body))
		if err != nil {
			return UserUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return UserUpdatedMsg{}
	}
}

// UpdatePostureCheck returns a tea.Cmd that updates a posture check
func UpdatePostureCheck(c *client.Client, checkID string, req models.PostureCheckRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return PostureCheckUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/posture-checks/"+url.PathEscape(checkID), bytes.NewReader(body))
		if err != nil {
			return PostureCheckUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return PostureCheckUpdatedMsg{}
	}
}

// DeletePostureCheck returns a tea.Cmd that deletes a posture check
func DeletePostureCheck(c *client.Client, checkID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/posture-checks/"+url.PathEscape(checkID), nil)
		if err != nil {
			return PostureCheckDeletedMsg{Err: err}
		}
		defer resp.Body.Close()
		return PostureCheckDeletedMsg{}
	}
}

// ─── Reverse Proxy ──────────────────────────────────────────────────

// FetchReverseProxyData loads services + lookup maps (groups/peers + clusters)
// in a single Cmd so the page can render on first paint without extra round trips.
func FetchReverseProxyData(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		groupMap, err := fetchGroupNameMap(c)
		if err != nil {
			return ReverseProxiesLoadedMsg{Err: err}
		}
		peerMap, err := fetchPeerNameMap(c)
		if err != nil {
			return ReverseProxiesLoadedMsg{Err: err}
		}

		clusters, err := fetchReverseProxyClustersList(c)
		if err != nil {
			// Clusters endpoint may not always be populated — don't fail the whole load.
			clusters = nil
		}

		resp, err := c.MakeRequest("GET", "/reverse-proxies/services", nil)
		if err != nil {
			return ReverseProxiesLoadedMsg{Err: err}
		}
		defer resp.Body.Close()
		var services []models.ReverseProxyService
		if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
			return ReverseProxiesLoadedMsg{Err: fmt.Errorf("decode services: %w", err)}
		}

		return ReverseProxiesLoadedMsg{
			Services: services,
			Clusters: clusters,
			Groups:   groupMap,
			Peers:    peerMap,
		}
	}
}

func fetchGroupNameMap(c *client.Client) (map[string]string, error) {
	resp, err := c.MakeRequest("GET", "/groups", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var all []models.PolicyGroup
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil, fmt.Errorf("decode groups: %w", err)
	}
	m := make(map[string]string, len(all))
	for _, g := range all {
		m[g.ID] = g.Name
	}
	return m, nil
}

func fetchPeerNameMap(c *client.Client) (map[string]string, error) {
	resp, err := c.MakeRequest("GET", "/peers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var all []models.Peer
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		return nil, fmt.Errorf("decode peers: %w", err)
	}
	m := make(map[string]string, len(all))
	for _, p := range all {
		m[p.ID] = p.Name
	}
	return m, nil
}

func fetchReverseProxyClustersList(c *client.Client) ([]models.ReverseProxyCluster, error) {
	resp, err := c.MakeRequest("GET", "/reverse-proxies/clusters", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var clusters []models.ReverseProxyCluster
	if err := json.NewDecoder(resp.Body).Decode(&clusters); err != nil {
		return nil, fmt.Errorf("decode clusters: %w", err)
	}
	return clusters, nil
}

// CreateReverseProxy creates a new service.
func CreateReverseProxy(c *client.Client, req models.ReverseProxyCreateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return ReverseProxyUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("POST", "/reverse-proxies/services", bytes.NewReader(body))
		if err != nil {
			return ReverseProxyUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return ReverseProxyUpdatedMsg{}
	}
}

// UpdateReverseProxy PUTs a full service object.
func UpdateReverseProxy(c *client.Client, serviceID string, req models.ReverseProxyUpdateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return ReverseProxyUpdatedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/reverse-proxies/services/"+url.PathEscape(serviceID), bytes.NewReader(body))
		if err != nil {
			return ReverseProxyUpdatedMsg{Err: err}
		}
		defer resp.Body.Close()
		return ReverseProxyUpdatedMsg{}
	}
}

// DeleteReverseProxy removes a service.
func DeleteReverseProxy(c *client.Client, serviceID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/reverse-proxies/services/"+url.PathEscape(serviceID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete reverse proxy"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Reverse proxy deleted"}
	}
}

// ToggleReverseProxy flips enabled on a service, re-sending the full object per API rules.
func ToggleReverseProxy(c *client.Client, svc models.ReverseProxyService, enable bool) tea.Cmd {
	return func() tea.Msg {
		req := ReverseProxyUpdateFromService(svc)
		req.Enabled = enable
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "toggle reverse proxy"}
		}
		resp, err := c.MakeRequest("PUT", "/reverse-proxies/services/"+url.PathEscape(svc.ID), bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "toggle reverse proxy"}
		}
		defer resp.Body.Close()
		action := "disabled"
		if enable {
			action = "enabled"
		}
		return ToastMsg{Message: "Reverse proxy " + action}
	}
}

// ReverseProxyUpdateFromService builds the full PUT body from a fetched service.
// Exported so form submit handlers can modify fields and re-send.
func ReverseProxyUpdateFromService(svc models.ReverseProxyService) models.ReverseProxyUpdateRequest {
	return models.ReverseProxyUpdateRequest{
		Name:               svc.Name,
		ListenPort:         svc.ListenPort,
		ProxyCluster:       svc.ProxyCluster,
		Targets:            svc.Targets,
		Enabled:            svc.Enabled,
		PassHostHeader:     svc.PassHostHeader,
		RewriteRedirects:   svc.RewriteRedirects,
		Auth:               svc.Auth,
		AccessRestrictions: svc.AccessRestrictions,
	}
}

// FetchReverseProxyDomains lists custom domains attached to a service.
func FetchReverseProxyDomains(c *client.Client, serviceID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/reverse-proxies/services/"+url.PathEscape(serviceID)+"/domains", nil)
		if err != nil {
			return ReverseProxyDomainsLoadedMsg{ServiceID: serviceID, Err: err}
		}
		defer resp.Body.Close()
		var domains []models.ReverseProxyDomain
		if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
			return ReverseProxyDomainsLoadedMsg{ServiceID: serviceID, Err: fmt.Errorf("decode domains: %w", err)}
		}
		return ReverseProxyDomainsLoadedMsg{ServiceID: serviceID, Domains: domains}
	}
}

// CreateReverseProxyDomain attaches a custom domain to a service.
func CreateReverseProxyDomain(c *client.Client, serviceID string, req models.ReverseProxyDomainCreateRequest) tea.Cmd {
	return func() tea.Msg {
		body, err := json.Marshal(req)
		if err != nil {
			return ReverseProxyDomainChangedMsg{Err: fmt.Errorf("marshal request: %w", err)}
		}
		resp, err := c.MakeRequest("POST", "/reverse-proxies/services/"+url.PathEscape(serviceID)+"/domains", bytes.NewReader(body))
		if err != nil {
			return ReverseProxyDomainChangedMsg{Err: err}
		}
		defer resp.Body.Close()
		return ReverseProxyDomainChangedMsg{Message: "Domain added"}
	}
}

// ValidateReverseProxyDomain triggers DNS validation for a pending custom domain.
func ValidateReverseProxyDomain(c *client.Client, serviceID, domainID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/reverse-proxies/services/"+url.PathEscape(serviceID)+"/domains/"+url.PathEscape(domainID)+"/validate", nil)
		if err != nil {
			return ReverseProxyDomainChangedMsg{Err: err}
		}
		defer resp.Body.Close()
		return ReverseProxyDomainChangedMsg{Message: "Validation started"}
	}
}

// DeleteReverseProxyDomain removes a custom domain from a service.
func DeleteReverseProxyDomain(c *client.Client, serviceID, domainID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/reverse-proxies/services/"+url.PathEscape(serviceID)+"/domains/"+url.PathEscape(domainID), nil)
		if err != nil {
			return ReverseProxyDomainChangedMsg{Err: err}
		}
		defer resp.Body.Close()
		return ReverseProxyDomainChangedMsg{Message: "Domain removed"}
	}
}

// FetchReverseProxyEvents loads the proxy access log for a service within a date range.
func FetchReverseProxyEvents(c *client.Client, serviceID, startDate, endDate string, page, limit int) tea.Cmd {
	return func() tea.Msg {
		q := url.Values{}
		if serviceID != "" {
			q.Set("service_id", serviceID)
		}
		if startDate != "" {
			q.Set("start_date", startDate)
		}
		if endDate != "" {
			q.Set("end_date", endDate)
		}
		if page > 0 {
			q.Set("page", fmt.Sprintf("%d", page))
		}
		if limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", limit))
		}

		endpoint := "/events/proxy"
		if encoded := q.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}

		resp, err := c.MakeRequest("GET", endpoint, nil)
		if err != nil {
			return ReverseProxyEventsLoadedMsg{ServiceID: serviceID, Err: err}
		}
		defer resp.Body.Close()

		// Response may be a bare array or an object wrapping events/pagination —
		// decode into a tolerant shape first, then fall back to a bare slice.
		var wrapper struct {
			Data   []models.ReverseProxyEvent `json:"data"`
			Events []models.ReverseProxyEvent `json:"events"`
		}
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return ReverseProxyEventsLoadedMsg{ServiceID: serviceID, Err: fmt.Errorf("read response: %w", err)}
		}
		if err := json.Unmarshal(raw, &wrapper); err == nil {
			if len(wrapper.Data) > 0 {
				return ReverseProxyEventsLoadedMsg{ServiceID: serviceID, Events: wrapper.Data}
			}
			if len(wrapper.Events) > 0 {
				return ReverseProxyEventsLoadedMsg{ServiceID: serviceID, Events: wrapper.Events}
			}
		}
		var bare []models.ReverseProxyEvent
		if err := json.Unmarshal(raw, &bare); err != nil {
			return ReverseProxyEventsLoadedMsg{ServiceID: serviceID, Err: fmt.Errorf("decode events: %w", err)}
		}
		return ReverseProxyEventsLoadedMsg{ServiceID: serviceID, Events: bare}
	}
}
