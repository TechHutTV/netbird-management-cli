package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// ─── Zone Form Data ──────────────────────────────────────────────────

type zoneFormData struct {
	name               string
	domain             string
	enabled            bool
	enableSearchDomain bool
	selectedGroups     []string
}

type recordFormData struct {
	name       string
	recordType string
	content    string
	ttl        string
}

// ─── Zone Forms ──────────────────────────────────────────────────────

func newZoneCreateForm(data *zoneFormData, availableGroups map[string]string) *huh.Form {
	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name, id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("e.g. Internal Zone").
				Value(&data.name),
			huh.NewInput().
				Title("Domain").
				Placeholder("e.g. company.internal").
				Value(&data.domain),
			huh.NewConfirm().Title("Enabled").Value(&data.enabled),
			huh.NewConfirm().
				Title("Search Domain").
				Description("Add this domain to the DNS search list").
				Value(&data.enableSearchDomain),
			huh.NewMultiSelect[string]().
				Title("Distribution Groups").
				Description("Peers in these groups will use this zone").
				Options(options...).
				Value(&data.selectedGroups),
		),
	)
}

func newZoneEditForm(data *zoneFormData, availableGroups map[string]string) *huh.Form {
	options := make([]huh.Option[string], 0, len(availableGroups))
	for id, name := range availableGroups {
		options = append(options, huh.NewOption(name, id))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&data.name),
			huh.NewInput().Title("Domain").Value(&data.domain),
			huh.NewConfirm().Title("Enabled").Value(&data.enabled),
			huh.NewConfirm().
				Title("Search Domain").
				Description("Add this domain to the DNS search list").
				Value(&data.enableSearchDomain),
			huh.NewMultiSelect[string]().
				Title("Distribution Groups").
				Options(options...).
				Value(&data.selectedGroups),
		),
	)
}

func newRecordForm(data *recordFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Record Name").
				Description("FQDN, e.g. server.company.internal").
				Placeholder("server.company.internal").
				Value(&data.name),
			huh.NewSelect[string]().
				Title("Type").
				Options(
					huh.NewOption("A (IPv4)", "A"),
					huh.NewOption("AAAA (IPv6)", "AAAA"),
					huh.NewOption("CNAME (Alias)", "CNAME"),
				).
				Value(&data.recordType),
			huh.NewInput().
				Title("Content").
				Description("IP address (A/AAAA) or domain (CNAME)").
				Placeholder("e.g. 192.168.1.1").
				Value(&data.content),
			huh.NewInput().
				Title("TTL (seconds)").
				Placeholder("300").
				Value(&data.ttl),
		),
	)
}

// ─── Zone API ────────────────────────────────────────────────────────

func submitZoneCreate(c *client.Client, data zoneFormData) tea.Cmd {
	return func() tea.Msg {
		req := models.DNSZoneRequest{
			Name:               data.name,
			Domain:             data.domain,
			Enabled:            data.enabled,
			EnableSearchDomain: data.enableSearchDomain,
			DistributionGroups: data.selectedGroups,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create zone"}
		}
		resp, err := c.MakeRequest("POST", "/dns/zones", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create zone"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Zone '%s' created", data.domain)}
	}
}

func submitZoneEdit(c *client.Client, zoneID string, data zoneFormData) tea.Cmd {
	return func() tea.Msg {
		req := models.DNSZoneRequest{
			Name:               data.name,
			Domain:             data.domain,
			Enabled:            data.enabled,
			EnableSearchDomain: data.enableSearchDomain,
			DistributionGroups: data.selectedGroups,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "update zone"}
		}
		resp, err := c.MakeRequest("PUT", "/dns/zones/"+url.PathEscape(zoneID), bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "update zone"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Zone '%s' updated", data.domain)}
	}
}

func submitRecordCreate(c *client.Client, zoneID string, data recordFormData) tea.Cmd {
	return func() tea.Msg {
		ttl := 300
		if v, err := strconv.Atoi(data.ttl); err == nil && v > 0 {
			ttl = v
		}
		req := models.DNSRecordRequest{
			Name:    data.name,
			Type:    data.recordType,
			Content: data.content,
			TTL:     ttl,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create record"}
		}
		resp, err := c.MakeRequest("POST", "/dns/zones/"+url.PathEscape(zoneID)+"/records", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create record"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Record '%s' created", data.name)}
	}
}

func deleteZone(c *client.Client, zoneID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/dns/zones/"+url.PathEscape(zoneID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete zone"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Zone deleted"}
	}
}

func toggleZone(c *client.Client, zone models.DNSZone, enable bool) tea.Cmd {
	return func() tea.Msg {
		req := models.DNSZoneRequest{
			Name:               zone.Name,
			Domain:             zone.Domain,
			Enabled:            enable,
			EnableSearchDomain: zone.EnableSearchDomain,
			DistributionGroups: zone.DistributionGroups,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "toggle zone"}
		}
		resp, err := c.MakeRequest("PUT", "/dns/zones/"+url.PathEscape(zone.ID), bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "toggle zone"}
		}
		defer resp.Body.Close()
		action := "disabled"
		if enable {
			action = "enabled"
		}
		return ToastMsg{Message: "Zone " + action}
	}
}
