package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// rpDomainsState represents the sub-state of the custom-domains screen.
type rpDomainsState int

const (
	rpDomainsList rpDomainsState = iota
	rpDomainsAddForm
)

// rpDomainsView holds the mutable state of the custom-domains sub-view.
type rpDomainsView struct {
	serviceID string
	domains   []models.ReverseProxyDomain
	cursor    int
	loading   bool
	err       error

	state    rpDomainsState
	form     *huh.Form
	formData rpDomainFormData
	focused  bool
}

func (d *rpDomainsView) setServiceID(id string) {
	if d.serviceID != id {
		d.domains = nil
		d.cursor = 0
	}
	d.serviceID = id
	d.state = rpDomainsList
	d.form = nil
}

func (d *rpDomainsView) load(c *client.Client) tea.Cmd {
	d.loading = true
	d.err = nil
	return FetchReverseProxyDomains(c, d.serviceID)
}

// handleDomainsMsg processes domain-related messages; returns true if the message was consumed.
func (d *rpDomainsView) handleDomainsMsg(msg tea.Msg, c *client.Client) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case ReverseProxyDomainsLoadedMsg:
		if msg.ServiceID != d.serviceID {
			return nil, false
		}
		d.loading = false
		if msg.Err != nil {
			d.err = msg.Err
			return nil, true
		}
		d.domains = msg.Domains
		if d.cursor >= len(d.domains) {
			d.cursor = 0
		}
		return nil, true

	case ReverseProxyDomainChangedMsg:
		if msg.Err != nil {
			d.err = msg.Err
			return nil, true
		}
		// After change: back to list, reload.
		d.state = rpDomainsList
		d.form = nil
		return d.load(c), true
	}
	return nil, false
}

// update processes a bubbletea message for the domains sub-view.
// Returns: command, whether the sub-view wants to close (user pressed esc at the list level).
func (d *rpDomainsView) update(msg tea.Msg, c *client.Client) (tea.Cmd, bool) {
	if cmd, handled := d.handleDomainsMsg(msg, c); handled {
		return cmd, false
	}

	if d.state == rpDomainsAddForm && d.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			d.state = rpDomainsList
			d.form = nil
			return nil, false
		}
		m, cmd := d.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			d.form = f
		}
		if d.form.State == huh.StateCompleted {
			data := d.formData
			d.state = rpDomainsList
			d.form = nil
			return CreateReverseProxyDomain(c, d.serviceID, models.ReverseProxyDomainCreateRequest{
				Domain:        strings.TrimSpace(data.domain),
				TargetCluster: strings.TrimSpace(data.targetCluster),
			}), false
		}
		if d.form.State == huh.StateAborted {
			d.state = rpDomainsList
			d.form = nil
		}
		return cmd, false
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, false
	}

	switch keyMsg.String() {
	case "esc":
		return nil, true
	case "up", "k":
		if d.cursor > 0 {
			d.cursor--
		}
	case "down", "j":
		if d.cursor < len(d.domains)-1 {
			d.cursor++
		}
	case "r":
		return d.load(c), false
	case "a", "c":
		d.formData = rpDomainFormData{}
		d.form = newRPDomainForm(&d.formData)
		d.state = rpDomainsAddForm
		return d.form.Init(), false
	case "v":
		if len(d.domains) > 0 {
			dom := d.domains[d.cursor]
			return ValidateReverseProxyDomain(c, d.serviceID, dom.ID), false
		}
	case "d":
		if len(d.domains) > 0 {
			dom := d.domains[d.cursor]
			return DeleteReverseProxyDomain(c, d.serviceID, dom.ID), false
		}
	}
	return nil, false
}

func (d *rpDomainsView) view(width int) string {
	if d.state == rpDomainsAddForm && d.form != nil {
		return pageTitleStyle.Render("Add Custom Domain") + "\n\n" + d.form.View()
	}

	var b strings.Builder
	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Custom Domains (%d)", len(d.domains))) + "\n")

	if d.loading {
		b.WriteString(loadingStyle.Render("  Loading...") + "\n")
		return b.String()
	}
	if d.err != nil {
		b.WriteString(errorStyle.Render("  Error: "+d.err.Error()) + "\n")
	}

	if len(d.domains) == 0 {
		b.WriteString(dimHintStyle.Render("  No custom domains attached.") + "\n")
	} else {
		rows := make([][]string, 0, len(d.domains))
		for _, dom := range d.domains {
			status := "unvalidated"
			if dom.Validated {
				status = "validated"
			}
			rows = append(rows, []string{dom.Domain, defaultStr(dom.Type, "-"), defaultStr(dom.TargetCluster, "-"), status})
		}

		tw := width
		if tw > 100 {
			tw = 100
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(tableBorderStyle).
			Headers("DOMAIN", "TYPE", "CLUSTER", "STATUS").
			Rows(rows...).
			Width(tw).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return tableHeaderStyle
				}
				base := tableCellStyle
				if row%2 == 0 {
					base = tableDimCellStyle
				}
				if row == d.cursor && d.focused {
					return tableSelectedStyle
				}
				if col == 3 && row < len(rows) {
					if rows[row][3] == "validated" {
						return base.Foreground(colorSuccess)
					}
					return base.Foreground(colorYellow)
				}
				return base
			})
		b.WriteString(t.Render() + "\n")
	}

	b.WriteString(dimHintStyle.Render("  esc: back  a: add  v: validate  d: delete  r: refresh"))
	return b.String()
}
