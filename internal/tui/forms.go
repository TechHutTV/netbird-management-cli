package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// formCompleteMsg is sent when a form completes and an API call succeeds
type formCompleteMsg struct {
	message string
}

// ─── Group Create ───────────────────────────────────────────────────

type groupFormData struct {
	name string
}

func newGroupCreateForm(data *groupFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Group Name").
				Placeholder("e.g. web-servers").
				Value(&data.name),
		),
	)
}

func submitGroupCreate(c *client.Client, data groupFormData) tea.Cmd {
	return func() tea.Msg {
		req := models.GroupPutRequest{Name: data.name}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create group"}
		}
		resp, err := c.MakeRequest("POST", "/groups", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create group"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Group '%s' created", data.name)}
	}
}

// ─── Peer Edit ─────────────────────────────────────────────────────

type peerEditFormData struct {
	name                        string
	sshEnabled                  bool
	loginExpirationEnabled      bool
	inactivityExpirationEnabled bool
}

func newPeerEditForm(data *peerEditFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&data.name),
			huh.NewConfirm().Title("SSH Enabled").Value(&data.sshEnabled),
			huh.NewConfirm().Title("Login Expiration Enabled").Value(&data.loginExpirationEnabled),
			huh.NewConfirm().Title("Inactivity Expiration Enabled").Value(&data.inactivityExpirationEnabled),
		),
	)
}

func submitPeerEdit(c *client.Client, peerID string, data peerEditFormData) tea.Cmd {
	req := models.PeerUpdateRequest{
		Name:                        data.name,
		SSHEnabled:                  data.sshEnabled,
		LoginExpirationEnabled:      data.loginExpirationEnabled,
		InactivityExpirationEnabled: data.inactivityExpirationEnabled,
	}
	return UpdatePeer(c, peerID, req)
}

// ─── User Edit ─────────────────────────────────────────────────────

type userEditFormData struct {
	role       string
	autoGroups string
}

func newUserEditForm(data *userEditFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Role").
				Options(
					huh.NewOption("Admin", "admin"),
					huh.NewOption("User", "user"),
					huh.NewOption("Owner", "owner"),
				).
				Value(&data.role),
			huh.NewInput().Title("Auto Groups (comma-separated IDs)").Value(&data.autoGroups),
		),
	)
}

func submitUserEdit(c *client.Client, userID string, data userEditFormData) tea.Cmd {
	req := models.UserUpdateRequest{
		Role:       data.role,
		AutoGroups: splitTrim(data.autoGroups),
	}
	return UpdateUser(c, userID, req)
}

// ─── Group Edit (submit only, reuses newGroupCreateForm) ───────────

func submitGroupEdit(c *client.Client, groupID string, data groupFormData) tea.Cmd {
	return UpdateGroup(c, groupID, data.name)
}

// ─── Setup Key Create ───────────────────────────────────────────────

type setupKeyFormData struct {
	name       string
	keyType    string
	expiresIn  string
	usageLimit string
	ephemeral  bool
}

func newSetupKeyCreateForm(data *setupKeyFormData) *huh.Form {
	data.keyType = "one-off"
	data.expiresIn = "7d"
	data.usageLimit = "0"

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Key Name").
				Placeholder("e.g. onboarding-key").
				Value(&data.name),
			huh.NewSelect[string]().
				Title("Type").
				Options(
					huh.NewOption("One-Off", "one-off"),
					huh.NewOption("Reusable", "reusable"),
				).
				Value(&data.keyType),
			huh.NewInput().
				Title("Expires In").
				Description("Duration: 1d, 7d, 30d, 1y").
				Placeholder("7d").
				Value(&data.expiresIn),
			huh.NewInput().
				Title("Usage Limit").
				Description("0 = unlimited").
				Placeholder("0").
				Value(&data.usageLimit),
			huh.NewConfirm().
				Title("Ephemeral").
				Description("Auto-remove peers when offline").
				Value(&data.ephemeral),
		),
	)
}

func submitSetupKeyCreate(c *client.Client, data setupKeyFormData) tea.Cmd {
	return func() tea.Msg {
		expiresIn := parseDurationToSeconds(data.expiresIn)
		limit, _ := strconv.Atoi(data.usageLimit)

		req := models.SetupKeyCreateRequest{
			Name:       data.name,
			Type:       data.keyType,
			ExpiresIn:  expiresIn,
			UsageLimit: limit,
			Ephemeral:  data.ephemeral,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create setup key"}
		}
		resp, err := c.MakeRequest("POST", "/setup-keys", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create setup key"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Setup key '%s' created", data.name)}
	}
}

// ─── Route Create ───────────────────────────────────────────────────

type routeFormData struct {
	networkID   string
	network     string
	description string
	metric      string
	masquerade  bool
	enabled          bool
	selectedPeerGrps []string
	selectedDistGrps []string
}

func newRouteCreateForm(data *routeFormData, availableGroups map[string]string) *huh.Form {
	data.metric = "100"
	data.masquerade = true
	data.enabled = true

	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name, id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Network ID").
				Description("Route identifier").
				Placeholder("e.g. office-lan").
				Value(&data.networkID),
			huh.NewInput().
				Title("Network CIDR").
				Placeholder("e.g. 10.0.0.0/24").
				Value(&data.network),
			huh.NewInput().
				Title("Description").
				Placeholder("optional").
				Value(&data.description),
			huh.NewMultiSelect[string]().
				Title("Routing Peer Groups").
				Description("Peers that will route traffic").
				Options(options...).
				Value(&data.selectedPeerGrps),
			huh.NewMultiSelect[string]().
				Title("Distribution Groups").
				Description("Peers that will use this route").
				Options(options...).
				Value(&data.selectedDistGrps),
			huh.NewInput().
				Title("Metric").
				Description("1-9999, lower = higher priority").
				Placeholder("100").
				Value(&data.metric),
			huh.NewConfirm().
				Title("Masquerade (NAT)").
				Value(&data.masquerade),
			huh.NewConfirm().
				Title("Enabled").
				Value(&data.enabled),
		),
	)
}

func submitRouteCreate(c *client.Client, data routeFormData) tea.Cmd {
	return func() tea.Msg {
		metric, _ := strconv.Atoi(data.metric)
		req := models.RouteRequest{
			NetworkID:   data.networkID,
			Network:     data.network,
			Description: data.description,
			PeerGroups:  data.selectedPeerGrps,
			Groups:      data.selectedDistGrps,
			Metric:      metric,
			Masquerade:  data.masquerade,
			Enabled:     data.enabled,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create route"}
		}
		resp, err := c.MakeRequest("POST", "/routes", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create route"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Route '%s' created", data.networkID)}
	}
}

// ─── DNS Create ─────────────────────────────────────────────────────

type dnsFormData struct {
	name           string
	nameservers    string
	selectedGroups []string // group IDs selected via multi-select
	domains        string
	primary        bool
	searchDomain   bool
	description    string
}

// groupOption holds id+name for multi-select display
type groupOption struct {
	ID   string
	Name string
}

func (g groupOption) String() string { return g.Name }

// newDNSCreateForm builds a DNS create form with a group multi-select dropdown.
// availableGroups is a map of groupID → groupName from the API.
func newDNSCreateForm(data *dnsFormData, availableGroups map[string]string) *huh.Form {
	// Build options for the multi-select
	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name+" ("+id+")", id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("e.g. internal-dns").
				Value(&data.name),
			huh.NewInput().
				Title("Nameservers").
				Description("Comma-separated IPs (port 53 default)").
				Placeholder("e.g. 1.1.1.1, 8.8.8.8").
				Value(&data.nameservers),
			huh.NewMultiSelect[string]().
				Title("Distribution Groups").
				Description("Select groups to distribute this nameserver to (space to toggle)").
				Options(options...).
				Value(&data.selectedGroups),
			huh.NewInput().
				Title("Match Domains").
				Description("Comma-separated, leave empty for all").
				Placeholder("e.g. internal.example.com").
				Value(&data.domains),
			huh.NewInput().
				Title("Description").
				Placeholder("optional").
				Value(&data.description),
			huh.NewConfirm().
				Title("Primary DNS").
				Value(&data.primary),
		),
		// Search domains only makes sense when match domains are specified
		huh.NewGroup(
			huh.NewConfirm().
				Title("Search Domains Enabled").
				Description("Adds match domains to the OS search domain list").
				Value(&data.searchDomain),
		).WithHideFunc(func() bool {
			return strings.TrimSpace(data.domains) == ""
		}),
	)
}

func submitDNSCreate(c *client.Client, data dnsFormData) tea.Cmd {
	return func() tea.Msg {
		nsList := buildNameservers(data.nameservers)
		var domains []string
		if data.domains != "" {
			domains = splitTrim(data.domains)
		}

		req := models.DNSNameserverGroupRequest{
			Name:                 data.name,
			Description:          data.description,
			Nameservers:          nsList,
			Groups:               data.selectedGroups,
			Domains:              domains,
			SearchDomainsEnabled: data.searchDomain,
			Primary:              data.primary,
			Enabled:              true,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create dns nameserver"}
		}
		resp, err := c.MakeRequest("POST", "/dns/nameservers", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create dns nameserver"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("DNS nameserver '%s' created", data.name)}
	}
}

// ─── Network Create ─────────────────────────────────────────────────

type networkFormData struct {
	name        string
	description string
}

func newNetworkCreateForm(data *networkFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Network Name").
				Placeholder("e.g. office-network").
				Value(&data.name),
			huh.NewInput().
				Title("Description").
				Placeholder("optional").
				Value(&data.description),
		),
	)
}

func submitNetworkCreate(c *client.Client, data networkFormData) tea.Cmd {
	return func() tea.Msg {
		req := models.NetworkCreateRequest{
			Name:        data.name,
			Description: data.description,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create network"}
		}
		resp, err := c.MakeRequest("POST", "/networks", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create network"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Network '%s' created", data.name)}
	}
}

// ─── Network Edit (submit only, reuses newNetworkCreateForm) ──────────────────

func submitNetworkEdit(c *client.Client, networkID string, data networkFormData) tea.Cmd {
	req := models.NetworkUpdateRequest{
		Name:        data.name,
		Description: data.description,
	}
	return UpdateNetwork(c, networkID, req)
}

// ─── Network Resource Add ───────────────────────────────────────────

type networkResourceFormData struct {
	name           string
	address        string
	description    string
	selectedGroups []string
}

func newNetworkResourceForm(data *networkResourceFormData, availableGroups map[string]string) *huh.Form {
	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name, id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Resource Name").
				Placeholder("e.g. web-server").
				Value(&data.name),
			huh.NewInput().
				Title("Address").
				Description("IP (1.1.1.1), subnet (10.0.0.0/24), or domain (*.example.com)").
				Placeholder("e.g. 192.168.1.0/24").
				Value(&data.address),
			huh.NewInput().
				Title("Description").
				Placeholder("optional").
				Value(&data.description),
			huh.NewMultiSelect[string]().
				Title("Groups").
				Description("Select groups for this resource (space to toggle)").
				Options(options...).
				Value(&data.selectedGroups),
		),
	)
}

func submitNetworkResource(c *client.Client, networkID string, data networkResourceFormData) tea.Cmd {
	return func() tea.Msg {
		req := models.NetworkResourceRequest{
			Name:        data.name,
			Address:     data.address,
			Description: data.description,
			Groups:      data.selectedGroups,
			Enabled:     true,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "add resource"}
		}
		resp, err := c.MakeRequest("POST", "/networks/"+url.PathEscape(networkID)+"/resources", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "add resource"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: fmt.Sprintf("Resource '%s' added", data.name)}
	}
}

// ─── Network Router Add ─────────────────────────────────────────────

type networkRouterFormData struct {
	selectedPeerGroups []string
	metric             string
	masquerade         bool
}

func newNetworkRouterForm(data *networkRouterFormData, availableGroups map[string]string) *huh.Form {
	data.metric = "100"
	data.masquerade = true

	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name, id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Peer Groups").
				Description("Select groups containing routing peers (space to toggle)").
				Options(options...).
				Value(&data.selectedPeerGroups),
			huh.NewInput().
				Title("Metric").
				Description("1-9999, lower = higher priority").
				Placeholder("100").
				Value(&data.metric),
			huh.NewConfirm().
				Title("Masquerade (NAT)").
				Value(&data.masquerade),
		),
	)
}

func submitNetworkRouter(c *client.Client, networkID string, data networkRouterFormData) tea.Cmd {
	return func() tea.Msg {
		metric, _ := strconv.Atoi(data.metric)
		req := models.NetworkRouterRequest{
			PeerGroups: data.selectedPeerGroups,
			Metric:     metric,
			Masquerade: data.masquerade,
			Enabled:    true,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "add routing peer"}
		}
		resp, err := c.MakeRequest("POST", "/networks/"+url.PathEscape(networkID)+"/routers", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "add routing peer"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Routing peer added"}
	}
}

// ─── Policy Create ──────────────────────────────────────────────────

type policyFormData struct {
	name           string
	description    string
	protocol       string
	action         string
	ports          string
	selectedSrc    []string // group IDs via multi-select
	selectedDst    []string // group IDs via multi-select
}

func newPolicyCreateForm(data *policyFormData, availableGroups map[string]string) *huh.Form {
	data.protocol = "all"
	data.action = "accept"

	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name, id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Policy Name").
				Placeholder("e.g. allow-web-servers").
				Value(&data.name),
			huh.NewInput().
				Title("Description").
				Placeholder("optional").
				Value(&data.description),
			huh.NewSelect[string]().
				Title("Protocol").
				Options(
					huh.NewOption("All", "all"),
					huh.NewOption("TCP", "tcp"),
					huh.NewOption("UDP", "udp"),
					huh.NewOption("ICMP", "icmp"),
				).
				Value(&data.protocol),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Accept", "accept"),
					huh.NewOption("Drop", "drop"),
				).
				Value(&data.action),
			huh.NewInput().
				Title("Ports").
				Description("Comma-separated, e.g. 80,443 (leave empty for all)").
				Placeholder("80, 443").
				Value(&data.ports),
			huh.NewMultiSelect[string]().
				Title("Source Groups").
				Options(options...).
				Value(&data.selectedSrc),
			huh.NewMultiSelect[string]().
				Title("Destination Groups").
				Options(options...).
				Value(&data.selectedDst),
		),
	)
}

func submitPolicyCreate(c *client.Client, data policyFormData) tea.Cmd {
	return func() tea.Msg {
		var ports []string
		if data.ports != "" {
			ports = splitTrim(data.ports)
		}

		rule := models.PolicyRuleForWrite{
			Name:          data.name,
			Enabled:       true,
			Action:        data.action,
			Protocol:      data.protocol,
			Ports:         ports,
			Sources:       data.selectedSrc,
			Destinations:  data.selectedDst,
			Bidirectional: true,
		}

		req := models.PolicyCreateRequest{
			Name:        data.name,
			Description: data.description,
			Enabled:     true,
			Rules:       []models.PolicyRuleForWrite{rule},
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create policy"}
		}
		resp, err := c.MakeRequest("POST", "/policies", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create policy"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Policy '%s' created", data.name)}
	}
}

// ─── User Invite ────────────────────────────────────────────────────

type userInviteFormData struct {
	email      string
	name       string
	role       string
	autoGroups string
	isService  bool
}

func newUserInviteForm(data *userInviteFormData) *huh.Form {
	data.role = "user"

	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Service User").
				Description("Create as service user (no email invite)").
				Value(&data.isService),
			huh.NewInput().
				Title("Name").
				Placeholder("e.g. John Doe").
				Value(&data.name),
			huh.NewInput().
				Title("Email").
				Description("Required for regular users").
				Placeholder("user@example.com").
				Value(&data.email),
			huh.NewSelect[string]().
				Title("Role").
				Options(
					huh.NewOption("User", "user"),
					huh.NewOption("Admin", "admin"),
				).
				Value(&data.role),
			huh.NewInput().
				Title("Auto Groups").
				Description("Comma-separated group IDs (optional)").
				Placeholder("group-id").
				Value(&data.autoGroups),
		),
	)
}

func submitUserInvite(c *client.Client, data userInviteFormData) tea.Cmd {
	return func() tea.Msg {
		var autoGroups []string
		if data.autoGroups != "" {
			autoGroups = splitTrim(data.autoGroups)
		}

		req := models.UserCreateRequest{
			Email:         data.email,
			Name:          data.name,
			Role:          data.role,
			AutoGroups:    autoGroups,
			IsServiceUser: data.isService,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create user"}
		}
		resp, err := c.MakeRequest("POST", "/users", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create user"}
		}
		defer resp.Body.Close()
		label := "User invited"
		if data.isService {
			label = "Service user created"
		}
		return formCompleteMsg{message: fmt.Sprintf("%s: %s", label, data.name)}
	}
}

// ─── Policy Edit ───────────────────────────────────────────────────

func newPolicyEditForm(data *policyFormData, availableGroups map[string]string) *huh.Form {
	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name, id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Policy Name").
				Value(&data.name),
			huh.NewInput().
				Title("Description").
				Value(&data.description),
			huh.NewSelect[string]().
				Title("Protocol").
				Options(
					huh.NewOption("All", "all"),
					huh.NewOption("TCP", "tcp"),
					huh.NewOption("UDP", "udp"),
					huh.NewOption("ICMP", "icmp"),
				).
				Value(&data.protocol),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Accept", "accept"),
					huh.NewOption("Drop", "drop"),
				).
				Value(&data.action),
			huh.NewInput().
				Title("Ports").
				Description("Comma-separated, e.g. 80,443 (leave empty for all)").
				Value(&data.ports),
			huh.NewMultiSelect[string]().
				Title("Source Groups").
				Options(options...).
				Value(&data.selectedSrc),
			huh.NewMultiSelect[string]().
				Title("Destination Groups").
				Options(options...).
				Value(&data.selectedDst),
		),
	)
}

func submitPolicyEdit(c *client.Client, policyID string, data policyFormData) tea.Cmd {
	var ports []string
	if data.ports != "" {
		ports = splitTrim(data.ports)
	}

	rules := []models.PolicyRuleForWrite{{
		Name:          data.name + "-rule",
		Enabled:       true,
		Action:        data.action,
		Bidirectional: true,
		Protocol:      data.protocol,
		Ports:         ports,
		Sources:       data.selectedSrc,
		Destinations:  data.selectedDst,
	}}
	req := models.PolicyUpdateRequest{
		Name:        data.name,
		Description: data.description,
		Enabled:     true,
		Rules:       rules,
	}
	return UpdatePolicy(c, policyID, req)
}

// ─── Route Edit ────────────────────────────────────────────────────

func newRouteEditForm(data *routeFormData, availableGroups map[string]string) *huh.Form {
	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name, id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Network ID").
				Description("Route identifier").
				Value(&data.networkID),
			huh.NewInput().
				Title("Network CIDR").
				Value(&data.network),
			huh.NewInput().
				Title("Description").
				Value(&data.description),
			huh.NewMultiSelect[string]().
				Title("Routing Peer Groups").
				Description("Peers that will route traffic").
				Options(options...).
				Value(&data.selectedPeerGrps),
			huh.NewMultiSelect[string]().
				Title("Distribution Groups").
				Description("Peers that will use this route").
				Options(options...).
				Value(&data.selectedDistGrps),
			huh.NewInput().
				Title("Metric").
				Description("1-9999, lower = higher priority").
				Value(&data.metric),
			huh.NewConfirm().
				Title("Masquerade (NAT)").
				Value(&data.masquerade),
			huh.NewConfirm().
				Title("Enabled").
				Value(&data.enabled),
		),
	)
}

func submitRouteEdit(c *client.Client, route models.Route, data routeFormData) tea.Cmd {
	metric := 9999
	if m, err := strconv.Atoi(data.metric); err == nil {
		metric = m
	}
	req := models.RouteRequest{
		Description:         data.description,
		NetworkID:           data.networkID,
		Network:             data.network,
		Domains:             route.Domains,
		Peer:                route.Peer,
		PeerGroups:          data.selectedPeerGrps,
		Metric:              metric,
		Masquerade:          data.masquerade,
		Enabled:             data.enabled,
		Groups:              data.selectedDistGrps,
		AccessControlGroups: route.AccessControlGroups,
		KeepRoute:           route.KeepRoute,
	}
	return UpdateRoute(c, route.ID, req)
}

// ─── DNS Edit ──────────────────────────────────────────────────────

func newDNSEditForm(data *dnsFormData, availableGroups map[string]string) *huh.Form {
	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name+" ("+id+")", id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Value(&data.name),
			huh.NewInput().
				Title("Nameservers").
				Description("Comma-separated IPs (port 53 default)").
				Value(&data.nameservers),
			huh.NewMultiSelect[string]().
				Title("Distribution Groups").
				Description("Select groups to distribute this nameserver to (space to toggle)").
				Options(options...).
				Value(&data.selectedGroups),
			huh.NewInput().
				Title("Match Domains").
				Description("Comma-separated, leave empty for all").
				Value(&data.domains),
			huh.NewInput().
				Title("Description").
				Value(&data.description),
			huh.NewConfirm().
				Title("Primary DNS").
				Value(&data.primary),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Search Domains Enabled").
				Description("Adds match domains to the OS search domain list").
				Value(&data.searchDomain),
		).WithHideFunc(func() bool {
			return strings.TrimSpace(data.domains) == ""
		}),
	)
}

func submitDNSEdit(c *client.Client, groupID string, data dnsFormData) tea.Cmd {
	var domains []string
	if data.domains != "" {
		domains = splitTrim(data.domains)
	}

	req := models.DNSNameserverGroupRequest{
		Name:                 data.name,
		Description:          data.description,
		Nameservers:          buildNameservers(data.nameservers),
		Groups:               data.selectedGroups,
		Domains:              domains,
		SearchDomainsEnabled: data.searchDomain,
		Primary:              data.primary,
		Enabled:              true,
	}
	return UpdateDNSGroup(c, groupID, req)
}

// formatNameservers converts a slice of Nameserver to a comma-separated IP string
func formatNameservers(servers []models.Nameserver) string {
	parts := make([]string, 0, len(servers))
	for _, ns := range servers {
		if ns.Port != 53 {
			parts = append(parts, fmt.Sprintf("%s:%d", ns.IP, ns.Port))
		} else {
			parts = append(parts, ns.IP)
		}
	}
	return strings.Join(parts, ", ")
}

// extractGroupIDs extracts IDs from a slice of PolicyGroup
func extractGroupIDs(groups []models.PolicyGroup) []string {
	ids := make([]string, len(groups))
	for i, g := range groups {
		ids[i] = g.ID
	}
	return ids
}

// ─── Helpers ────────────────────────────────────────────────────────

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func parseDurationToSeconds(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 7 * 86400 // default 7 days
	}
	multiplier := 1
	if strings.HasSuffix(s, "d") {
		multiplier = 86400
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "w") {
		multiplier = 7 * 86400
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "y") {
		multiplier = 365 * 86400
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "h") {
		multiplier = 3600
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "m") {
		multiplier = 60
		s = s[:len(s)-1]
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 7 * 86400
	}
	return n * multiplier
}

func buildNameservers(input string) []models.Nameserver {
	parts := splitTrim(input)
	nsList := make([]models.Nameserver, 0, len(parts))
	for _, p := range parts {
		ip := p
		port := 53
		if strings.Contains(p, ":") {
			split := strings.SplitN(p, ":", 2)
			ip = split[0]
			if n, err := strconv.Atoi(split[1]); err == nil {
				port = n
			}
		}
		nsList = append(nsList, models.Nameserver{
			IP:     ip,
			NSType: "udp",
			Port:   port,
		})
	}
	return nsList
}
