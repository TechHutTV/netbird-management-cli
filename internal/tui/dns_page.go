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
	// Zone-specific states
	dnsViewZoneDetail
	dnsViewZoneForm
	dnsViewZoneConfirm
	dnsViewZoneEdit
	dnsViewZoneEditConfirm
	dnsViewRecordForm
	dnsViewRecordConfirm
)

// dnsSection tracks which section the cursor is in on the list view
type dnsSection int

const (
	dnsSectionNameservers dnsSection = iota
	dnsSectionZones
)

// DNSPage manages DNS nameserver groups and zones on a single page
type DNSPage struct {
	// Nameservers
	groups     []models.DNSNameserverGroup
	filtered   []models.DNSNameserverGroup
	cursor     int
	loading    bool
	err        error
	state      dnsViewState
	groupNames map[string]string
	form       *huh.Form
	formData   dnsFormData
	focused    bool
	editForm   *huh.Form
	editData   dnsFormData
	editID     string
	search     string
	searching  bool

	// Zones
	zones       []models.DNSZone
	zoneCursor  int
	zoneForm    *huh.Form
	zoneData    zoneFormData
	zoneEditID  string
	recForm     *huh.Form
	recData     recordFormData
	recZoneID   string

	// Which section is active in list view
	section dnsSection
}

func NewDNSPage() *DNSPage {
	return &DNSPage{loading: true}
}

func (d *DNSPage) Title() string { return "DNS" }
func (d *DNSPage) CursorPosition() int {
	if d.state != dnsViewList {
		return 1
	}
	if d.section == dnsSectionZones {
		return d.zoneCursor
	}
	return d.cursor
}
func (d *DNSPage) SetFocused(focused bool) { d.focused = focused }

// AcceptTab accepts focus into the page on the current section.
func (d *DNSPage) AcceptTab() bool {
	return d.state == dnsViewList
}

// CycleTab advances nameservers → zones, then returns false to go back to nav.
func (d *DNSPage) CycleTab() bool {
	if d.state != dnsViewList {
		return false // don't cycle when in a form/detail
	}
	if d.section == dnsSectionNameservers {
		d.section = dnsSectionZones
		return true
	}
	// At zones already — let tab go back to nav
	d.section = dnsSectionNameservers
	return false
}

// dnsDataLoadedMsg carries nameservers + zones + group names
type dnsDataLoadedMsg struct {
	groups     []models.DNSNameserverGroup
	zones      []models.DNSZone
	groupNames map[string]string
	err        error
}

func fetchDNSData(c *client.Client) tea.Cmd {
	return func() tea.Msg {
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

		dnsResp, err := c.MakeRequest("GET", "/dns/nameservers", nil)
		if err != nil {
			return dnsDataLoadedMsg{err: err}
		}
		defer dnsResp.Body.Close()
		var dnsGroups []models.DNSNameserverGroup
		if err := jsonDecode(dnsResp.Body, &dnsGroups); err != nil {
			return dnsDataLoadedMsg{err: err}
		}

		// Zones may 404 on older API versions
		var zones []models.DNSZone
		if zonesResp, err := c.MakeRequest("GET", "/dns/zones", nil); err == nil {
			defer zonesResp.Body.Close()
			_ = jsonDecode(zonesResp.Body, &zones)
		}

		return dnsDataLoadedMsg{groups: dnsGroups, zones: zones, groupNames: nameMap}
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
	// Form delegation for all form states
	switch d.state {
	case dnsViewEditConfirm:
		return d.handleConfirm(msg, func() tea.Cmd { return submitDNSEdit(c, d.editID, d.editData) }, dnsViewDetail)
	case dnsViewConfirm:
		return d.handleConfirm(msg, func() tea.Cmd { return submitDNSCreate(c, d.formData) }, dnsViewList)
	case dnsViewZoneEditConfirm:
		return d.handleConfirm(msg, func() tea.Cmd { return submitZoneEdit(c, d.zoneEditID, d.zoneData) }, dnsViewList)
	case dnsViewZoneConfirm:
		return d.handleConfirm(msg, func() tea.Cmd { return submitZoneCreate(c, d.zoneData) }, dnsViewList)
	case dnsViewRecordConfirm:
		return d.handleConfirm(msg, func() tea.Cmd { return submitRecordCreate(c, d.recZoneID, d.recData) }, dnsViewList)

	case dnsViewEdit:
		return d.handleForm(msg, &d.editForm, dnsViewEditConfirm, dnsViewDetail)
	case dnsViewForm:
		return d.handleForm(msg, &d.form, dnsViewConfirm, dnsViewList)
	case dnsViewZoneEdit:
		return d.handleForm(msg, &d.zoneForm, dnsViewZoneEditConfirm, dnsViewList)
	case dnsViewZoneForm:
		return d.handleForm(msg, &d.zoneForm, dnsViewZoneConfirm, dnsViewList)
	case dnsViewRecordForm:
		return d.handleForm(msg, &d.recForm, dnsViewRecordConfirm, dnsViewList)
	}

	switch msg := msg.(type) {
	case PageRefreshTickMsg:
		if d.state == dnsViewList {
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
		d.zones = msg.zones
		d.groupNames = msg.groupNames
		d.applyFilter()
		return d, nil

	case formCompleteMsg, ToastMsg:
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

// handleConfirm handles y/n on any confirm screen
func (d *DNSPage) handleConfirm(msg tea.Msg, onYes func() tea.Cmd, cancelState dnsViewState) (Page, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "y":
			d.state = dnsViewList
			return d, onYes()
		case "n", "esc":
			d.state = cancelState
		}
	}
	return d, nil
}

// handleForm delegates to an active huh.Form
func (d *DNSPage) handleForm(msg tea.Msg, form **huh.Form, doneState, cancelState dnsViewState) (Page, tea.Cmd) {
	if *form == nil {
		return d, nil
	}
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
		d.state = cancelState
		*form = nil
		return d, nil
	}
	m, cmd := (*form).Update(msg)
	if f, ok := m.(*huh.Form); ok {
		*form = f
	}
	if (*form).State == huh.StateCompleted {
		d.state = doneState
		*form = nil
		return d, nil
	}
	if (*form).State == huh.StateAborted {
		d.state = cancelState
		*form = nil
	}
	return d, cmd
}

func (d *DNSPage) View(width, height int) string {
	if d.loading {
		return loadingStyle.Render("  Loading DNS...")
	}
	if d.err != nil {
		return errorStyle.Render("  Error: " + d.err.Error())
	}

	switch d.state {
	// Nameserver forms/details
	case dnsViewEditConfirm:
		return RenderConfirm("Confirm: Edit DNS Nameserver", []ConfirmField{
			{Label: "Name", Value: d.editData.name},
			{Label: "Nameservers", Value: d.editData.nameservers},
			{Label: "Groups", Value: resolveGroupNames(d.editData.selectedGroups, d.groupNames)},
			{Label: "Match Domains", Value: d.editData.domains},
			{Label: "Description", Value: d.editData.description},
			{Label: "Primary", Value: fmt.Sprintf("%v", d.editData.primary)},
			{Label: "Search Domains", Value: fmt.Sprintf("%v", d.editData.searchDomain)},
		})
	case dnsViewEdit:
		return pageTitleStyle.Render("Edit DNS Nameserver") + "\n\n" + d.editForm.View()
	case dnsViewConfirm:
		return RenderConfirm("Confirm: Create DNS Nameserver", []ConfirmField{
			{Label: "Name", Value: d.formData.name},
			{Label: "Nameservers", Value: d.formData.nameservers},
			{Label: "Groups", Value: resolveGroupNames(d.formData.selectedGroups, d.groupNames)},
			{Label: "Match Domains", Value: d.formData.domains},
			{Label: "Description", Value: d.formData.description},
			{Label: "Primary", Value: fmt.Sprintf("%v", d.formData.primary)},
			{Label: "Search Domains", Value: fmt.Sprintf("%v", d.formData.searchDomain)},
		})
	case dnsViewForm:
		return pageTitleStyle.Render("Create DNS Nameserver") + "\n\n" + d.form.View()
	case dnsViewDetail:
		return d.viewNSDetail(width)

	// Zone forms/details
	case dnsViewZoneEditConfirm:
		return RenderConfirm("Confirm: Edit Zone", []ConfirmField{
			{Label: "Name", Value: d.zoneData.name},
			{Label: "Domain", Value: d.zoneData.domain},
			{Label: "Enabled", Value: fmt.Sprintf("%v", d.zoneData.enabled)},
			{Label: "Search Domain", Value: fmt.Sprintf("%v", d.zoneData.enableSearchDomain)},
			{Label: "Groups", Value: resolveGroupNames(d.zoneData.selectedGroups, d.groupNames)},
		})
	case dnsViewZoneEdit:
		return pageTitleStyle.Render("Edit Zone") + "\n\n" + d.zoneForm.View()
	case dnsViewZoneConfirm:
		return RenderConfirm("Confirm: Create Zone", []ConfirmField{
			{Label: "Name", Value: d.zoneData.name},
			{Label: "Domain", Value: d.zoneData.domain},
			{Label: "Enabled", Value: fmt.Sprintf("%v", d.zoneData.enabled)},
			{Label: "Search Domain", Value: fmt.Sprintf("%v", d.zoneData.enableSearchDomain)},
			{Label: "Groups", Value: resolveGroupNames(d.zoneData.selectedGroups, d.groupNames)},
		})
	case dnsViewZoneForm:
		return pageTitleStyle.Render("Create Zone") + "\n\n" + d.zoneForm.View()
	case dnsViewRecordConfirm:
		return RenderConfirm("Confirm: Create Record", []ConfirmField{
			{Label: "Name", Value: d.recData.name},
			{Label: "Type", Value: d.recData.recordType},
			{Label: "Content", Value: d.recData.content},
			{Label: "TTL", Value: d.recData.ttl},
		})
	case dnsViewRecordForm:
		return pageTitleStyle.Render("Add DNS Record") + "\n\n" + d.recForm.View()
	}

	return d.viewList(width, height)
}

func (d *DNSPage) applyFilter() {
	if d.search == "" {
		d.filtered = d.groups
	} else {
		filtered := make([]models.DNSNameserverGroup, 0)
		for _, g := range d.groups {
			if matchesQuery(d.search, g.Name, g.Description) {
				filtered = append(filtered, g)
			}
		}
		d.filtered = filtered
	}
	d.cursor = clampCursor(d.cursor, len(d.filtered))
	d.zoneCursor = clampCursor(d.zoneCursor, len(d.zones))
}

func (d *DNSPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	// NS detail view
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

	// Zone detail is inline — no separate detail view

	// Search
	if d.searching {
		newSearch, still, changed := handleSearchKey(key, d.search)
		d.search = newSearch
		d.searching = still
		if changed || !still {
			d.applyFilter()
		}
		return d, nil
	}

	// List view
	switch key {
	case "up", "k":
		if d.section == dnsSectionNameservers {
			if d.cursor > 0 {
				d.cursor--
			}
		} else {
			if d.zoneCursor > 0 {
				d.zoneCursor--
			}
		}
	case "down", "j":
		if d.section == dnsSectionNameservers {
			if d.cursor < len(d.filtered)-1 {
				d.cursor++
			}
		} else {
			if d.zoneCursor < len(d.zones)-1 {
				d.zoneCursor++
			}
		}
	case "enter":
		if d.section == dnsSectionNameservers {
			if len(d.filtered) > 0 {
				d.state = dnsViewDetail
			}
		} else {
			// Enter on zones opens edit directly
			if d.zoneCursor < len(d.zones) {
				zone := d.zones[d.zoneCursor]
				d.zoneEditID = zone.ID
				d.zoneData = zoneFormData{
					name:               zone.Name,
					domain:             zone.Domain,
					enabled:            zone.Enabled,
					enableSearchDomain: zone.EnableSearchDomain,
					selectedGroups:     zone.DistributionGroups,
				}
				d.zoneForm = newZoneEditForm(&d.zoneData, d.groupNames)
				d.state = dnsViewZoneEdit
				return d, d.zoneForm.Init()
			}
		}
	case "a":
		// Add record to selected zone
		if d.section == dnsSectionZones && d.zoneCursor < len(d.zones) {
			zone := d.zones[d.zoneCursor]
			d.recZoneID = zone.ID
			d.recData = recordFormData{recordType: "A", ttl: "300"}
			d.recForm = newRecordForm(&d.recData)
			d.state = dnsViewRecordForm
			return d, d.recForm.Init()
		}
	case "r":
		d.loading = true
		return d, fetchDNSData(c)
	case "c":
		if d.section == dnsSectionNameservers {
			d.formData = dnsFormData{}
			d.form = newDNSCreateForm(&d.formData, d.groupNames)
			d.state = dnsViewForm
			return d, d.form.Init()
		}
		d.zoneData = zoneFormData{enabled: true}
		d.zoneForm = newZoneCreateForm(&d.zoneData, d.groupNames)
		d.state = dnsViewZoneForm
		return d, d.zoneForm.Init()
	case "d":
		if d.section == dnsSectionNameservers {
			if len(d.filtered) > 0 {
				return d, deleteDNSGroup(c, d.filtered[d.cursor].ID)
			}
		} else {
			if len(d.zones) > 0 {
				return d, deleteZone(c, d.zones[d.zoneCursor].ID)
			}
		}
	case "t":
		if d.section == dnsSectionNameservers {
			if len(d.filtered) > 0 {
				group := d.filtered[d.cursor]
				return d, toggleDNSGroup(c, group, !group.Enabled)
			}
		} else {
			if len(d.zones) > 0 {
				zone := d.zones[d.zoneCursor]
				return d, toggleZone(c, zone, !zone.Enabled)
			}
		}
	case "/":
		d.searching = true
		d.search = ""
	case "esc":
		if d.search != "" {
			d.search = ""
			d.cursor = 0
			d.applyFilter()
		}
	}

	return d, nil
}

// ─── List View ───────────────────────────────────────────────────────

func (d *DNSPage) viewList(width, height int) string {
	var b strings.Builder

	if d.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(d.search) + "\u2588\n")
	} else if d.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", d.search)) + "\n")
	}

	// Split height between the two tables
	nsHeight := (height - 4) / 2
	zoneHeight := height - 4 - nsHeight

	// ── Nameservers table ──
	nsActive := d.section == dnsSectionNameservers
	nsTitle := fmt.Sprintf("Nameservers (%d)", len(d.filtered))
	if nsActive {
		b.WriteString(sectionHeaderStyle.Render("  "+nsTitle) + "\n")
	} else {
		b.WriteString(dimHintStyle.Render("  "+nsTitle) + "\n")
	}

	if len(d.filtered) == 0 {
		b.WriteString(dimHintStyle.Render("    No nameserver groups.") + "\n")
	} else {
		b.WriteString(d.renderNSTable(width, nsHeight, nsActive))
	}

	b.WriteString("\n")

	// ── Zones table ──
	znActive := d.section == dnsSectionZones
	znTitle := fmt.Sprintf("Zones (%d)", len(d.zones))
	if znActive {
		b.WriteString(sectionHeaderStyle.Render("  "+znTitle) + "\n")
	} else {
		b.WriteString(dimHintStyle.Render("  "+znTitle) + "\n")
	}

	if len(d.zones) == 0 {
		b.WriteString(dimHintStyle.Render("    No DNS zones.") + "\n")
	} else {
		b.WriteString(d.renderZoneTable(width, zoneHeight, znActive))
	}

	hints := "  tab: switch section  /: search  c: create  t: toggle  d: delete  r: refresh"
	if d.section == dnsSectionZones {
		hints = "  tab: switch section  enter: edit zone  a: add record  c: create  t: toggle  d: delete  r: refresh"
	}
	b.WriteString("\n" + dimHintStyle.Render(hints))
	return b.String()
}

func (d *DNSPage) renderNSTable(width, maxRows int, active bool) string {
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
		enabled := "No"
		if group.Enabled {
			enabled = "Yes"
		}
		nsIPs := make([]string, len(group.Nameservers))
		for j, ns := range group.Nameservers {
			nsIPs[j] = ns.IP
		}
		rows = append(rows, []string{
			group.Name,
			strings.Join(nsIPs, ", "),
			resolveGroupNames(group.Groups, d.groupNames),
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
		Headers("NAME", "NAMESERVERS", "GROUPS", "ENABLED").
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
			if dataIdx == d.cursor && active && d.focused {
				base = tableSelectedStyle
			}
			if col == 3 && row >= 0 && row < len(rows) {
				if rows[row][3] == "Yes" {
					return base.Foreground(colorSuccess)
				}
				return base.Foreground(colorDanger)
			}
			return base
		})

	return t.Render() + "\n"
}

// zoneRow maps a table row back to its zone index for cursor highlighting
type zoneRow struct {
	zoneIdx int
	cols    []string
}

func (d *DNSPage) renderZoneTable(width, maxRows int, active bool) string {
	if maxRows < 1 {
		maxRows = 1
	}

	// Build flat rows: one per record, zone domain shown on first record row only
	allRows := make([]zoneRow, 0)
	for i, zone := range d.zones {
		enabled := "No"
		if zone.Enabled {
			enabled = "Yes"
		}
		if len(zone.Records) == 0 {
			allRows = append(allRows, zoneRow{i, []string{zone.Domain, "-", "-", "-", "-", enabled}})
		} else {
			for j, rec := range zone.Records {
				domain := ""
				en := ""
				if j == 0 {
					domain = zone.Domain
					en = enabled
				}
				allRows = append(allRows, zoneRow{i, []string{domain, rec.Name, rec.Type, rec.Content, fmt.Sprintf("%d", rec.TTL), en}})
			}
		}
	}

	// Find the first row for the cursor's zone
	cursorRowStart := 0
	for i, r := range allRows {
		if r.zoneIdx == d.zoneCursor {
			cursorRowStart = i
			break
		}
	}

	offset := 0
	if cursorRowStart >= maxRows {
		offset = cursorRowStart - maxRows + 1
	}
	end := offset + maxRows
	if end > len(allRows) {
		end = len(allRows)
	}

	rows := make([][]string, 0, end-offset)
	rowMeta := make([]zoneRow, 0, end-offset)
	for i := offset; i < end; i++ {
		rows = append(rows, allRows[i].cols)
		rowMeta = append(rowMeta, allRows[i])
	}

	tw := width
	if tw > 120 {
		tw = 120
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("DOMAIN", "NAME", "TYPE", "CONTENT", "TTL", "ENABLED").
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
			if row >= 0 && row < len(rowMeta) && rowMeta[row].zoneIdx == d.zoneCursor && active && d.focused {
				base = tableSelectedStyle
			}
			if col == 5 && row >= 0 && row < len(rows) && rows[row][5] != "" {
				if rows[row][5] == "Yes" {
					return base.Foreground(colorSuccess)
				}
				if rows[row][5] == "No" {
					return base.Foreground(colorDanger)
				}
			}
			return base
		})

	return t.Render() + "\n"
}

// ─── Detail Views ────────────────────────────────────────────────────

func (d *DNSPage) viewNSDetail(width int) string {
	if d.cursor >= len(d.filtered) {
		return "No DNS group selected"
	}

	group := d.filtered[d.cursor]
	var b strings.Builder

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

	fields := []struct{ label, value string }{
		{"ID", group.ID},
		{"Name", group.Name},
		{"Description", group.Description},
		{"Primary", fmt.Sprintf("%v", group.Primary)},
		{"Enabled", fmt.Sprintf("%v", group.Enabled)},
		{"Search Domains", fmt.Sprintf("%v", group.SearchDomainsEnabled)},
		{"Nameservers", nsStr},
		{"Target Groups", resolveGroupNames(group.Groups, d.groupNames)},
		{"Match Domains", domainsStr},
	}
	for _, f := range fields {
		b.WriteString(fmt.Sprintf("%s  %s\n", detailLabelStyle.Render(f.label), detailValueStyle.Render(f.value)))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  e: edit  d: delete  r: refresh"))
	return b.String()
}


// ─── Nameserver API helpers ──────────────────────────────────────────

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
