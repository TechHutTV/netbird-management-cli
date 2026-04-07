package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

type dnsViewState int

const (
	dnsViewList dnsViewState = iota
	dnsViewDetail
	dnsViewForm
	dnsViewConfirm
	dnsViewEdit
	dnsViewEditConfirm
)

// DNSPage manages DNS nameserver groups list and detail views
type DNSPage struct {
	groups     []models.DNSNameserverGroup
	filtered   []models.DNSNameserverGroup
	cursor     int
	loading    bool
	err        error
	state      dnsViewState
	groupNames map[string]string // group ID → name lookup
	form       *huh.Form
	formData   dnsFormData
	focused    bool
	editForm   *huh.Form
	editData   dnsFormData
	editID     string
}

func NewDNSPage() *DNSPage {
	return &DNSPage{loading: true}
}

func (d *DNSPage) Title() string { return "DNS" }
func (d *DNSPage) CursorPosition() int { return d.cursor }
func (d *DNSPage) SetFocused(focused bool) { d.focused = focused }

// dnsDataLoadedMsg carries DNS groups + resolved group names
type dnsDataLoadedMsg struct {
	groups     []models.DNSNameserverGroup
	groupNames map[string]string
	err        error
}

func fetchDNSData(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		// Fetch groups first for the name lookup
		groupResp, err := c.MakeRequest("GET", "/groups", nil)
		if err != nil {
			return dnsDataLoadedMsg{err: err}
		}
		defer groupResp.Body.Close()

		var allGroups []models.PolicyGroup
		if err := jsonDecode(groupResp.Body, &allGroups); err != nil {
			return dnsDataLoadedMsg{err: err}
		}

		nameMap := make(map[string]string, len(allGroups))
		for _, g := range allGroups {
			nameMap[g.ID] = g.Name
		}

		// Fetch DNS nameserver groups
		dnsResp, err := c.MakeRequest("GET", "/dns/nameservers", nil)
		if err != nil {
			return dnsDataLoadedMsg{err: err}
		}
		defer dnsResp.Body.Close()

		var dnsGroups []models.DNSNameserverGroup
		if err := jsonDecode(dnsResp.Body, &dnsGroups); err != nil {
			return dnsDataLoadedMsg{err: err}
		}

		return dnsDataLoadedMsg{groups: dnsGroups, groupNames: nameMap}
	}
}

func (d *DNSPage) Init(c *client.Client) tea.Cmd {
	if len(d.groups) > 0 {
		return nil
	}
	d.loading = true
	d.err = nil
	return fetchDNSData(c)
}

func (d *DNSPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle edit confirm screen
	if d.state == dnsViewEditConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				d.state = dnsViewList
				return d, submitDNSEdit(c, d.editID, d.editData)
			case "n", "esc":
				d.state = dnsViewDetail
			}
		}
		return d, nil
	}

	// Delegate to edit form when active
	if d.state == dnsViewEdit && d.editForm != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			d.state = dnsViewDetail
			d.editForm = nil
			return d, nil
		}
		m, cmd := d.editForm.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			d.editForm = f
		}
		if d.editForm.State == huh.StateCompleted {
			d.state = dnsViewEditConfirm
			d.editForm = nil
			return d, nil
		}
		if d.editForm.State == huh.StateAborted {
			d.state = dnsViewDetail
			d.editForm = nil
		}
		return d, cmd
	}

	// Handle create confirm screen
	if d.state == dnsViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				d.state = dnsViewList
				return d, submitDNSCreate(c, d.formData)
			case "n", "esc":
				d.state = dnsViewList
			}
		}
		return d, nil
	}

	// Delegate to create form when active
	if d.state == dnsViewForm && d.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			d.state = dnsViewList
			d.form = nil
			return d, nil
		}
		m, cmd := d.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			d.form = f
		}
		if d.form.State == huh.StateCompleted {
			d.state = dnsViewConfirm
			d.form = nil
			return d, nil
		}
		if d.form.State == huh.StateAborted {
			d.state = dnsViewList
			d.form = nil
		}
		return d, cmd
	}

	switch msg := msg.(type) {
	case PageRefreshTickMsg:
		// Only refresh if in list view and not editing or in form
		if d.state == dnsViewList && d.form == nil && d.editForm == nil {
			d.loading = true
			return d, fetchDNSData(c)
		}
		return d, nil

	case DNSUpdatedMsg:
		if msg.Err != nil {
			d.err = msg.Err
			return d, nil
		}
		d.loading = true
		return d, fetchDNSData(c)

	case dnsDataLoadedMsg:
		d.loading = false
		if msg.err != nil {
			d.err = msg.err
			return d, nil
		}
		d.groups = msg.groups
		d.filtered = msg.groups
		d.groupNames = msg.groupNames
		return d, nil

	case formCompleteMsg:
		d.loading = true
		return d, fetchDNSData(c)

	case ToastMsg:
		d.loading = true
		return d, fetchDNSData(c)

	case APIErrorMsg:
		d.err = msg.Err
		return d, nil

	case tea.KeyPressMsg:
		return d.handleKey(msg, c)
	}

	return d, nil
}

func (d *DNSPage) View(width, height int) string {
	if d.loading {
		return loadingStyle.Render("  Loading DNS nameserver groups...")
	}
	if d.err != nil {
		return errorStyle.Render("  Error: " + d.err.Error())
	}

	if d.state == dnsViewEditConfirm {
		return RenderConfirm("Confirm: Edit DNS Nameserver", []ConfirmField{
			{Label: "Name", Value: d.editData.name},
			{Label: "Nameservers", Value: d.editData.nameservers},
			{Label: "Groups", Value: d.resolveGroupNames(d.editData.selectedGroups)},
			{Label: "Match Domains", Value: d.editData.domains},
			{Label: "Description", Value: d.editData.description},
			{Label: "Primary", Value: fmt.Sprintf("%v", d.editData.primary)},
			{Label: "Search Domains", Value: fmt.Sprintf("%v", d.editData.searchDomain)},
		})
	}
	if d.state == dnsViewEdit && d.editForm != nil {
		return pageTitleStyle.Render("Edit DNS Nameserver") + "\n\n" + d.editForm.View()
	}
	if d.state == dnsViewConfirm {
		return RenderConfirm("Confirm: Create DNS Nameserver", []ConfirmField{
			{Label: "Name", Value: d.formData.name},
			{Label: "Nameservers", Value: d.formData.nameservers},
			{Label: "Groups", Value: d.resolveGroupNames(d.formData.selectedGroups)},
			{Label: "Match Domains", Value: d.formData.domains},
			{Label: "Description", Value: d.formData.description},
			{Label: "Primary", Value: fmt.Sprintf("%v", d.formData.primary)},
			{Label: "Search Domains", Value: fmt.Sprintf("%v", d.formData.searchDomain)},
		})
	}
	if d.state == dnsViewForm && d.form != nil {
		return pageTitleStyle.Render("Create DNS Nameserver") + "\n\n" + d.form.View()
	}
	if d.state == dnsViewDetail {
		return d.viewDetail(width)
	}
	return d.viewList(width, height)
}

func (d *DNSPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if d.state == dnsViewDetail {
		switch key {
		case "esc", "backspace", "q":
			d.state = dnsViewList
		case "e":
			if d.cursor < len(d.filtered) {
				group := d.filtered[d.cursor]
				d.editID = group.ID
				d.editData = dnsFormData{
					name:           group.Name,
					description:    group.Description,
					nameservers:    formatNameservers(group.Nameservers),
					selectedGroups: group.Groups,
					domains:        strings.Join(group.Domains, ", "),
					primary:        group.Primary,
					searchDomain:   group.SearchDomainsEnabled,
				}
				d.editForm = newDNSEditForm(&d.editData, d.groupNames)
				d.state = dnsViewEdit
				return d, d.editForm.Init()
			}
		}
		return d, nil
	}

	switch key {
	case "up", "k":
		if d.cursor > 0 {
			d.cursor--
		}
	case "down", "j":
		if d.cursor < len(d.filtered)-1 {
			d.cursor++
		}
	case "enter":
		if len(d.filtered) > 0 {
			d.state = dnsViewDetail
		}
	case "r":
		d.loading = true
		return d, fetchDNSData(c)
	case "c":
		d.formData = dnsFormData{}
		d.form = newDNSCreateForm(&d.formData, d.groupNames)
		d.state = dnsViewForm
		return d, d.form.Init()
	case "d":
		if len(d.filtered) > 0 {
			group := d.filtered[d.cursor]
			return d, deleteDNSGroup(c, group.ID)
		}
	case "t":
		if len(d.filtered) > 0 {
			group := d.filtered[d.cursor]
			return d, toggleDNSGroup(c, group, !group.Enabled)
		}
	}

	return d, nil
}

func (d *DNSPage) resolveGroupNames(ids []string) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := d.groupNames[id]; ok {
			names = append(names, name)
		} else {
			names = append(names, id)
		}
	}
	return strings.Join(names, ", ")
}

func toggleDNSGroup(c *client.Client, group models.DNSNameserverGroup, enable bool) tea.Cmd {
	return func() tea.Msg {
		req := models.DNSNameserverGroupRequest{
			Name:                 group.Name,
			Description:          group.Description,
			Nameservers:          group.Nameservers,
			Groups:               group.Groups,
			Domains:              group.Domains,
			SearchDomainsEnabled: group.SearchDomainsEnabled,
			Primary:              group.Primary,
			Enabled:              enable,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "toggle dns group"}
		}
		resp, err := c.MakeRequest("PUT", "/dns/nameservers/"+url.PathEscape(group.ID), bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "toggle dns group"}
		}
		defer resp.Body.Close()
		action := "disabled"
		if enable {
			action = "enabled"
		}
		return ToastMsg{Message: "DNS group " + action}
	}
}

func (d *DNSPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("DNS Nameserver Groups (%d)", len(d.filtered))) + "\n")

	if len(d.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No DNS nameserver groups found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if d.cursor >= maxRows {
		offset = d.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(d.filtered) {
		end = len(d.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		group := d.filtered[i]

		matchDomains := "All"
		if len(group.Domains) > 0 {
			matchDomains = strings.Join(group.Domains, ", ")
		}

		searchDomains := "No"
		if group.SearchDomainsEnabled {
			searchDomains = "Yes"
		}

		enabled := "No"
		if group.Enabled {
			enabled = "Yes"
		}

		nsIPs := make([]string, len(group.Nameservers))
		for j, ns := range group.Nameservers {
			nsIPs[j] = ns.IP
		}

		groupNameList := make([]string, 0, len(group.Groups))
		for _, gid := range group.Groups {
			if name, ok := d.groupNames[gid]; ok {
				groupNameList = append(groupNameList, name)
			} else {
				groupNameList = append(groupNameList, gid)
			}
		}
		distGroups := strings.Join(groupNameList, ", ")
		if len(groupNameList) == 0 {
			distGroups = "-"
		}

		rows = append(rows, []string{
			group.Name,
			strings.Join(nsIPs, ", "),
			distGroups,
			matchDomains,
			searchDomains,
			enabled,
		})
	}

	tw := width
	if tw > 120 {
		tw = 120
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "NAMESERVERS", "GROUPS", "MATCH DOMAINS", "SEARCH", "ENABLED").
		Rows(rows...).
		Width(tw).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return tableHeaderStyle
			}

			dataIdx := row + offset
			base := tableCellStyle
			if row%2 == 0 {
				base = tableDimCellStyle
			}

			if dataIdx == d.cursor && d.focused {
				base = tableSelectedStyle
			}

			// Enabled column coloring
			if col == 5 && row >= 0 && row < len(rows) {
				if rows[row][5] == "Yes" {
					return base.Foreground(colorSuccess)
				}
				return base.Foreground(colorDanger)
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  c: create  t: toggle  d: delete  r: refresh", d.cursor+1, len(d.filtered))))

	return b.String()
}

func (d *DNSPage) viewDetail(width int) string {
	if d.cursor >= len(d.filtered) {
		return "No DNS group selected"
	}

	group := d.filtered[d.cursor]
	var b strings.Builder

	// Title with enabled badge
	badge := disabledStyle.Render(" DISABLED ")
	if group.Enabled {
		badge = enabledStyle.Render(" ENABLED ")
	}

	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("  %s  %s", group.Name, badge)))
	b.WriteString("\n\n")

	nsLines := make([]string, len(group.Nameservers))
	for i, ns := range group.Nameservers {
		nsLines[i] = fmt.Sprintf("%s:%d (%s)", ns.IP, ns.Port, ns.NSType)
	}
	nsStr := "None"
	if len(nsLines) > 0 {
		nsStr = strings.Join(nsLines, ", ")
	}

	domainsStr := "All domains"
	if len(group.Domains) > 0 {
		domainsStr = strings.Join(group.Domains, ", ")
	}

	detailGroupNames := make([]string, 0, len(group.Groups))
	for _, gid := range group.Groups {
		if name, ok := d.groupNames[gid]; ok {
			detailGroupNames = append(detailGroupNames, name)
		} else {
			detailGroupNames = append(detailGroupNames, gid)
		}
	}
	groupsStr := "None"
	if len(detailGroupNames) > 0 {
		groupsStr = strings.Join(detailGroupNames, ", ")
	}

	fields := []struct{ label, value string }{
		{"ID", group.ID},
		{"Name", group.Name},
		{"Description", group.Description},
		{"Primary", fmt.Sprintf("%v", group.Primary)},
		{"Enabled", fmt.Sprintf("%v", group.Enabled)},
		{"Search Domains", fmt.Sprintf("%v", group.SearchDomainsEnabled)},
		{"Nameservers", nsStr},
		{"Target Groups", groupsStr},
		{"Match Domains", domainsStr},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  e: edit  d: delete  r: refresh"))

	return b.String()
}

func deleteDNSGroup(c *client.Client, groupID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/dns/nameservers/"+url.PathEscape(groupID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete dns group"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "DNS group deleted"}
	}
}
