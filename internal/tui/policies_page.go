package tui

import (
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

type policiesViewState int

const (
	policiesViewList policiesViewState = iota
	policiesViewDetail
	policiesViewForm
	policiesViewConfirm
	policiesViewEdit
	policiesViewEditConfirm
)

// PoliciesPage manages policies list and detail views with rules
type PoliciesPage struct {
	policies   []models.Policy
	filtered   []models.Policy
	cursor     int
	loading    bool
	err        error
	state      policiesViewState
	form       *huh.Form
	formData   policyFormData
	groupNames map[string]string
	focused    bool
	search     string
	searching  bool
	editForm   *huh.Form
	editData   policyFormData
	editID     string
}

func NewPoliciesPage() *PoliciesPage {
	return &PoliciesPage{loading: true}
}

func (p *PoliciesPage) Title() string { return "Policies" }
func (p *PoliciesPage) CursorPosition() int { return p.cursor }
func (p *PoliciesPage) SetFocused(focused bool) { p.focused = focused }

type policiesDataLoadedMsg struct {
	policies   []models.Policy
	groupNames map[string]string
	err        error
}

func fetchPoliciesData(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		// Fetch groups for name resolution
		groupResp, err := c.MakeRequest("GET", "/groups", nil)
		if err != nil {
			return policiesDataLoadedMsg{err: err}
		}
		defer groupResp.Body.Close()
		var allGroups []models.PolicyGroup
		if err := jsonDecode(groupResp.Body, &allGroups); err != nil {
			return policiesDataLoadedMsg{err: err}
		}
		nameMap := make(map[string]string, len(allGroups))
		for _, g := range allGroups {
			nameMap[g.ID] = g.Name
		}

		// Fetch policies
		polResp, err := c.MakeRequest("GET", "/policies", nil)
		if err != nil {
			return policiesDataLoadedMsg{err: err}
		}
		defer polResp.Body.Close()
		var policies []models.Policy
		if err := jsonDecode(polResp.Body, &policies); err != nil {
			return policiesDataLoadedMsg{err: err}
		}

		return policiesDataLoadedMsg{policies: policies, groupNames: nameMap}
	}
}

func (p *PoliciesPage) Init(c *client.Client) tea.Cmd {
	p.loading = true
	p.err = nil
	return fetchPoliciesData(c)
}

func (p *PoliciesPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle edit confirm screen
	if p.state == policiesViewEditConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				p.state = policiesViewList
				return p, submitPolicyEdit(c, p.editID, p.editData)
			case "n", "esc":
				p.state = policiesViewDetail
			}
		}
		return p, nil
	}

	// Delegate to edit form when active
	if p.state == policiesViewEdit && p.editForm != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			p.state = policiesViewDetail
			p.editForm = nil
			return p, nil
		}
		m, cmd := p.editForm.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			p.editForm = f
		}
		if p.editForm.State == huh.StateCompleted {
			p.state = policiesViewEditConfirm
			p.editForm = nil
			return p, nil
		}
		if p.editForm.State == huh.StateAborted {
			p.state = policiesViewDetail
			p.editForm = nil
		}
		return p, cmd
	}

	// Handle create confirm screen
	if p.state == policiesViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				p.state = policiesViewList
				return p, submitPolicyCreate(c, p.formData)
			case "n", "esc":
				p.state = policiesViewList
			}
		}
		return p, nil
	}

	// Delegate to create form when active
	if p.state == policiesViewForm && p.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			p.state = policiesViewList
			p.form = nil
			return p, nil
		}
		m, cmd := p.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			p.form = f
		}
		if p.form.State == huh.StateCompleted {
			p.state = policiesViewConfirm
			p.form = nil
			return p, nil
		}
		if p.form.State == huh.StateAborted {
			p.state = policiesViewList
			p.form = nil
		}
		return p, cmd
	}

	switch msg := msg.(type) {
	case PolicyUpdatedMsg:
		if msg.Err != nil {
			p.err = msg.Err
			return p, nil
		}
		return p, p.Init(c)

	case policiesDataLoadedMsg:
		p.loading = false
		if msg.err != nil {
			p.err = msg.err
			return p, nil
		}
		p.policies = msg.policies
		p.groupNames = msg.groupNames
		p.applyFilter()
		return p, nil

	case formCompleteMsg:
		return p, p.Init(c)

	case ToastMsg:
		return p, p.Init(c)

	case APIErrorMsg:
		p.err = msg.Err
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg, c)
	}

	return p, nil
}

func (p *PoliciesPage) View(width, height int) string {
	if p.loading {
		return loadingStyle.Render("  Loading policies...")
	}
	if p.err != nil {
		return errorStyle.Render("  Error: " + p.err.Error())
	}

	if p.state == policiesViewEditConfirm {
		return RenderConfirm("Confirm: Edit Policy", []ConfirmField{
			{Label: "Name", Value: p.editData.name},
			{Label: "Description", Value: p.editData.description},
			{Label: "Protocol", Value: p.editData.protocol},
			{Label: "Action", Value: p.editData.action},
			{Label: "Ports", Value: p.editData.ports},
			{Label: "Source Groups", Value: p.resolveGroupNames(p.editData.selectedSrc)},
			{Label: "Dest Groups", Value: p.resolveGroupNames(p.editData.selectedDst)},
		})
	}
	if p.state == policiesViewEdit && p.editForm != nil {
		return pageTitleStyle.Render("Edit Policy") + "\n\n" + p.editForm.View()
	}
	if p.state == policiesViewConfirm {
		return RenderConfirm("Confirm: Create Policy", []ConfirmField{
			{Label: "Name", Value: p.formData.name},
			{Label: "Description", Value: p.formData.description},
			{Label: "Protocol", Value: p.formData.protocol},
			{Label: "Action", Value: p.formData.action},
			{Label: "Ports", Value: p.formData.ports},
			{Label: "Source Groups", Value: p.resolveGroupNames(p.formData.selectedSrc)},
			{Label: "Dest Groups", Value: p.resolveGroupNames(p.formData.selectedDst)},
		})
	}
	if p.state == policiesViewForm && p.form != nil {
		return pageTitleStyle.Render("Create Policy") + "\n\n" + p.form.View()
	}
	if p.state == policiesViewDetail {
		return p.viewDetail(width)
	}
	return p.viewList(width, height)
}

func (p *PoliciesPage) applyFilter() {
	if p.search == "" {
		p.filtered = p.policies
	} else {
		filtered := make([]models.Policy, 0)
		for _, policy := range p.policies {
			if matchesQuery(p.search, policy.Name) {
				filtered = append(filtered, policy)
			}
		}
		p.filtered = filtered
	}
	p.cursor = clampCursor(p.cursor, len(p.filtered))
}

func (p *PoliciesPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if p.state == policiesViewDetail {
		switch key {
		case "esc", "backspace", "q":
			p.state = policiesViewList
		case "e":
			if p.cursor < len(p.filtered) {
				policy := p.filtered[p.cursor]
				p.editID = policy.ID
				p.editData = policyFormData{
					name:        policy.Name,
					description: policy.Description,
				}
				// Extract from first rule if available
				if len(policy.Rules) > 0 {
					rule := policy.Rules[0]
					p.editData.protocol = rule.Protocol
					p.editData.action = rule.Action
					p.editData.ports = strings.Join(rule.Ports, ", ")
					p.editData.selectedSrc = extractGroupIDs(rule.Sources)
					p.editData.selectedDst = extractGroupIDs(rule.Destinations)
				}
				p.editForm = newPolicyEditForm(&p.editData, p.groupNames)
				p.state = policiesViewEdit
				return p, p.editForm.Init()
			}
		}
		return p, nil
	}

	// Search input handling (list view only)
	if p.searching {
		newSearch, still, changed := handleSearchKey(key, p.search)
		p.search = newSearch
		p.searching = still
		if changed || !still {
			p.applyFilter()
		}
		return p, nil
	}

	switch key {
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
		}
	case "down", "j":
		if p.cursor < len(p.filtered)-1 {
			p.cursor++
		}
	case "enter":
		if len(p.filtered) > 0 {
			p.state = policiesViewDetail
		}
	case "r":
		return p, p.Init(c)
	case "c":
		p.formData = policyFormData{}
		p.form = newPolicyCreateForm(&p.formData, p.groupNames)
		p.state = policiesViewForm
		return p, p.form.Init()
	case "d":
		if len(p.filtered) > 0 {
			policy := p.filtered[p.cursor]
			return p, deletePolicy(c, policy.ID)
		}
	case "t":
		if len(p.filtered) > 0 {
			policy := p.filtered[p.cursor]
			return p, TogglePolicy(c, policy, !policy.Enabled)
		}
	case "/":
		p.searching = true
		p.search = ""
	case "esc":
		if p.search != "" {
			p.search = ""
			p.cursor = 0
			p.applyFilter()
		}
	}

	return p, nil
}

func (p *PoliciesPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Policies (%d)", len(p.filtered))) + "\n")

	if p.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(p.search) + "\u2588\n")
	} else if p.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", p.search)) + "\n")
	}

	if len(p.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No policies found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if p.cursor >= maxRows {
		offset = p.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(p.filtered) {
		end = len(p.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		policy := p.filtered[i]

		enabled := "No"
		if policy.Enabled {
			enabled = "Yes"
		}

		rows = append(rows, []string{
			policy.Name,
			enabled,
			fmt.Sprintf("%d", len(policy.Rules)),
			policy.Description,
		})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "ENABLED", "RULES", "DESCRIPTION").
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

			if dataIdx == p.cursor && p.focused {
				base = tableSelectedStyle
			}

			// Enabled column coloring
			if col == 1 && row >= 0 && row < len(rows) {
				if rows[row][1] == "Yes" {
					return base.Foreground(colorSuccess)
				}
				return base.Foreground(colorDanger)
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  c: create  t: toggle  d: delete  r: refresh", p.cursor+1, len(p.filtered))))

	return b.String()
}

func (p *PoliciesPage) viewDetail(width int) string {
	if p.cursor >= len(p.filtered) {
		return "No policy selected"
	}

	policy := p.filtered[p.cursor]
	var b strings.Builder

	// Title with enabled badge
	badge := disabledStyle.Render(" DISABLED ")
	if policy.Enabled {
		badge = enabledStyle.Render(" ENABLED ")
	}

	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("  %s  %s", policy.Name, badge)))
	b.WriteString("\n\n")

	fields := []struct{ label, value string }{
		{"ID", policy.ID},
		{"Name", policy.Name},
		{"Description", policy.Description},
		{"Enabled", fmt.Sprintf("%v", policy.Enabled)},
		{"Rules", fmt.Sprintf("%d", len(policy.Rules))},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	// Rules sub-sections
	for i, rule := range policy.Rules {
		b.WriteString("\n" + sectionHeaderStyle.Render(fmt.Sprintf("  Rule %d: %s", i+1, rule.Name)) + "\n")

		sources := make([]string, len(rule.Sources))
		for j, s := range rule.Sources {
			sources[j] = s.Name
		}
		destinations := make([]string, len(rule.Destinations))
		for j, d := range rule.Destinations {
			destinations[j] = d.Name
		}

		ports := "all"
		if len(rule.Ports) > 0 {
			ports = strings.Join(rule.Ports, ", ")
		}

		srcStr := "any"
		if len(sources) > 0 {
			srcStr = strings.Join(sources, ", ")
		}
		dstStr := "any"
		if len(destinations) > 0 {
			dstStr = strings.Join(destinations, ", ")
		}

		ruleFields := []struct{ label, value string }{
			{"Action", rule.Action},
			{"Protocol", rule.Protocol},
			{"Ports", ports},
			{"Sources", srcStr},
			{"Destinations", dstStr},
			{"Bidirectional", fmt.Sprintf("%v", rule.Bidirectional)},
		}

		for _, f := range ruleFields {
			label := detailLabelStyle.Render(f.label)
			value := detailValueStyle.Render(f.value)
			b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
		}
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  e: edit  d: delete  r: refresh"))

	return b.String()
}

func (p *PoliciesPage) resolveGroupNames(ids []string) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := p.groupNames[id]; ok {
			names = append(names, name)
		} else {
			names = append(names, id)
		}
	}
	return strings.Join(names, ", ")
}

func deletePolicy(c *client.Client, policyID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/policies/"+url.PathEscape(policyID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete policy"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Policy deleted"}
	}
}
