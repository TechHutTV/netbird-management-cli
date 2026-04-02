package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
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
