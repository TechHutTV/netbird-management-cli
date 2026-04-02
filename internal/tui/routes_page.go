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

type routesViewState int

const (
	routesViewList routesViewState = iota
	routesViewDetail
	routesViewForm
	routesViewConfirm
	routesViewEdit
	routesViewEditConfirm
)

// RoutesPage manages routes list and detail views
type RoutesPage struct {
	routes     []models.Route
	filtered   []models.Route
	cursor     int
	loading    bool
	err        error
	state      routesViewState
	form       *huh.Form
	formData   routeFormData
	groupNames map[string]string
	focused    bool
	search     string
	searching  bool
	editForm   *huh.Form
	editData   routeFormData
	editRoute  models.Route
}

type routesDataLoadedMsg struct {
	routes     []models.Route
	groupNames map[string]string
	err        error
}

func fetchRoutesData(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		groupResp, err := c.MakeRequest("GET", "/groups", nil)
		if err != nil {
			return routesDataLoadedMsg{err: err}
		}
		defer groupResp.Body.Close()
		var allGroups []models.PolicyGroup
		if err := jsonDecode(groupResp.Body, &allGroups); err != nil {
			return routesDataLoadedMsg{err: err}
		}
		nameMap := make(map[string]string, len(allGroups))
		for _, g := range allGroups {
			nameMap[g.ID] = g.Name
		}

		routeResp, err := c.MakeRequest("GET", "/routes", nil)
		if err != nil {
			return routesDataLoadedMsg{err: err}
		}
		defer routeResp.Body.Close()
		var routes []models.Route
		if err := jsonDecode(routeResp.Body, &routes); err != nil {
			return routesDataLoadedMsg{err: err}
		}

		return routesDataLoadedMsg{routes: routes, groupNames: nameMap}
	}
}

func NewRoutesPage() *RoutesPage {
	return &RoutesPage{loading: true}
}

func (r *RoutesPage) Title() string { return "Routes" }
func (r *RoutesPage) CursorPosition() int { return r.cursor }
func (r *RoutesPage) SetFocused(focused bool) { r.focused = focused }

func (r *RoutesPage) Init(c *client.Client) tea.Cmd {
	if len(r.routes) > 0 {
		return nil
	}
	r.loading = true
	r.err = nil
	return fetchRoutesData(c)
}

func (r *RoutesPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle edit confirm screen
	if r.state == routesViewEditConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				r.state = routesViewList
				return r, submitRouteEdit(c, r.editRoute, r.editData)
			case "n", "esc":
				r.state = routesViewDetail
			}
		}
		return r, nil
	}

	// Delegate to edit form when active
	if r.state == routesViewEdit && r.editForm != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			r.state = routesViewDetail
			r.editForm = nil
			return r, nil
		}
		m, cmd := r.editForm.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			r.editForm = f
		}
		if r.editForm.State == huh.StateCompleted {
			r.state = routesViewEditConfirm
			r.editForm = nil
			return r, nil
		}
		if r.editForm.State == huh.StateAborted {
			r.state = routesViewDetail
			r.editForm = nil
		}
		return r, cmd
	}

	// Handle create confirm screen
	if r.state == routesViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				r.state = routesViewList
				return r, submitRouteCreate(c, r.formData)
			case "n", "esc":
				r.state = routesViewList
			}
		}
		return r, nil
	}

	// Delegate to create form when active
	if r.state == routesViewForm && r.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			r.state = routesViewList
			r.form = nil
			return r, nil
		}
		m, cmd := r.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			r.form = f
		}
		if r.form.State == huh.StateCompleted {
			r.state = routesViewConfirm
			r.form = nil
			return r, nil
		}
		if r.form.State == huh.StateAborted {
			r.state = routesViewList
			r.form = nil
		}
		return r, cmd
	}

	switch msg := msg.(type) {
	case RouteUpdatedMsg:
		if msg.Err != nil {
			r.err = msg.Err
			return r, nil
		}
		r.loading = true
		return r, fetchRoutesData(c)

	case routesDataLoadedMsg:
		r.loading = false
		if msg.err != nil {
			r.err = msg.err
			return r, nil
		}
		r.routes = msg.routes
		r.groupNames = msg.groupNames
		r.applyFilter()
		return r, nil

	case formCompleteMsg:
		r.loading = true
		return r, fetchRoutesData(c)

	case ToastMsg:
		r.loading = true
		return r, fetchRoutesData(c)

	case APIErrorMsg:
		r.err = msg.Err
		return r, nil

	case tea.KeyPressMsg:
		return r.handleKey(msg, c)
	}

	return r, nil
}

func (r *RoutesPage) View(width, height int) string {
	if r.loading {
		return loadingStyle.Render("  Loading routes...")
	}
	if r.err != nil {
		return errorStyle.Render("  Error: " + r.err.Error())
	}

	if r.state == routesViewEditConfirm {
		return RenderConfirm("Confirm: Edit Route", []ConfirmField{
			{Label: "Network ID", Value: r.editData.networkID},
			{Label: "Network CIDR", Value: r.editData.network},
			{Label: "Description", Value: r.editData.description},
			{Label: "Peer Groups", Value: r.resolveGroupNames(r.editData.selectedPeerGrps)},
			{Label: "Dist Groups", Value: r.resolveGroupNames(r.editData.selectedDistGrps)},
			{Label: "Metric", Value: r.editData.metric},
			{Label: "Masquerade", Value: fmt.Sprintf("%v", r.editData.masquerade)},
			{Label: "Enabled", Value: fmt.Sprintf("%v", r.editData.enabled)},
		})
	}
	if r.state == routesViewEdit && r.editForm != nil {
		return pageTitleStyle.Render("Edit Route") + "\n\n" + r.editForm.View()
	}
	if r.state == routesViewConfirm {
		return RenderConfirm("Confirm: Create Route", []ConfirmField{
			{Label: "Network ID", Value: r.formData.networkID},
			{Label: "Network CIDR", Value: r.formData.network},
			{Label: "Description", Value: r.formData.description},
			{Label: "Peer Groups", Value: r.resolveGroupNames(r.formData.selectedPeerGrps)},
			{Label: "Dist Groups", Value: r.resolveGroupNames(r.formData.selectedDistGrps)},
			{Label: "Metric", Value: r.formData.metric},
			{Label: "Masquerade", Value: fmt.Sprintf("%v", r.formData.masquerade)},
			{Label: "Enabled", Value: fmt.Sprintf("%v", r.formData.enabled)},
		})
	}
	if r.state == routesViewForm && r.form != nil {
		return pageTitleStyle.Render("Create Route") + "\n\n" + r.form.View()
	}
	if r.state == routesViewDetail {
		return r.viewDetail(width)
	}
	return r.viewList(width, height)
}

func (r *RoutesPage) applyFilter() {
	if r.search == "" {
		r.filtered = r.routes
	} else {
		filtered := make([]models.Route, 0)
		for _, route := range r.routes {
			if matchesQuery(r.search, route.NetworkID, route.Description) {
				filtered = append(filtered, route)
			}
		}
		r.filtered = filtered
	}
	r.cursor = clampCursor(r.cursor, len(r.filtered))
}

func (r *RoutesPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if r.state == routesViewDetail {
		switch key {
		case "esc", "backspace", "q":
			r.state = routesViewList
		case "e":
			if r.cursor < len(r.filtered) {
				route := r.filtered[r.cursor]
				r.editRoute = route
				r.editData = routeFormData{
					networkID:        route.NetworkID,
					network:          route.Network,
					description:      route.Description,
					metric:           fmt.Sprintf("%d", route.Metric),
					masquerade:       route.Masquerade,
					enabled:          route.Enabled,
					selectedPeerGrps: route.PeerGroups,
					selectedDistGrps: route.Groups,
				}
				r.editForm = newRouteEditForm(&r.editData, r.groupNames)
				r.state = routesViewEdit
				return r, r.editForm.Init()
			}
		}
		return r, nil
	}

	// Search input handling (list view only)
	if r.searching {
		newSearch, still, changed := handleSearchKey(key, r.search)
		r.search = newSearch
		r.searching = still
		if changed || !still {
			r.applyFilter()
		}
		return r, nil
	}

	switch key {
	case "up", "k":
		if r.cursor > 0 {
			r.cursor--
		}
	case "down", "j":
		if r.cursor < len(r.filtered)-1 {
			r.cursor++
		}
	case "enter":
		if len(r.filtered) > 0 {
			r.state = routesViewDetail
		}
	case "r":
		r.loading = true
		return r, fetchRoutesData(c)
	case "c":
		r.formData = routeFormData{}
		r.form = newRouteCreateForm(&r.formData, r.groupNames)
		r.state = routesViewForm
		return r, r.form.Init()
	case "d":
		if len(r.filtered) > 0 {
			route := r.filtered[r.cursor]
			return r, deleteRoute(c, route.ID)
		}
	case "t":
		if len(r.filtered) > 0 {
			route := r.filtered[r.cursor]
			return r, ToggleRoute(c, route, !route.Enabled)
		}
	case "/":
		r.searching = true
		r.search = ""
	case "esc":
		if r.search != "" {
			r.search = ""
			r.cursor = 0
			r.applyFilter()
		}
	}

	return r, nil
}

func (r *RoutesPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Routes (%d)", len(r.filtered))) + "\n")

	if r.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(r.search) + "\u2588\n")
	} else if r.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", r.search)) + "\n")
	}

	if len(r.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No routes found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if r.cursor >= maxRows {
		offset = r.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(r.filtered) {
		end = len(r.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		route := r.filtered[i]

		masq := "No"
		if route.Masquerade {
			masq = "Yes"
		}
		enabled := "No"
		if route.Enabled {
			enabled = "Yes"
		}

		network := route.Network
		if len(route.Domains) > 0 {
			network = strings.Join(route.Domains, ", ")
		}

		rows = append(rows, []string{
			network,
			route.NetworkType,
			fmt.Sprintf("%d", route.Metric),
			masq,
			enabled,
		})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NETWORK", "TYPE", "METRIC", "MASQ", "ENABLED").
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

			if dataIdx == r.cursor && r.focused {
				base = tableSelectedStyle
			}

			// Enabled column coloring
			if col == 4 && row >= 0 && row < len(rows) {
				if rows[row][4] == "Yes" {
					return base.Foreground(colorSuccess)
				}
				return base.Foreground(colorDanger)
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  c: create  t: toggle  d: delete  r: refresh", r.cursor+1, len(r.filtered))))

	return b.String()
}

func (r *RoutesPage) viewDetail(width int) string {
	if r.cursor >= len(r.filtered) {
		return "No route selected"
	}

	route := r.filtered[r.cursor]
	var b strings.Builder

	// Title with enabled badge
	badge := disabledStyle.Render(" DISABLED ")
	if route.Enabled {
		badge = enabledStyle.Render(" ENABLED ")
	}

	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("  %s  %s", route.NetworkID, badge)))
	b.WriteString("\n\n")

	network := route.Network
	if len(route.Domains) > 0 {
		network = strings.Join(route.Domains, ", ")
	}

	peerInfo := route.Peer
	if len(route.PeerGroups) > 0 {
		peerInfo = strings.Join(route.PeerGroups, ", ") + " (groups)"
	}

	groupsStr := "None"
	if len(route.Groups) > 0 {
		groupsStr = strings.Join(route.Groups, ", ")
	}

	fields := []struct{ label, value string }{
		{"ID", route.ID},
		{"Network ID", route.NetworkID},
		{"Network", network},
		{"Type", route.NetworkType},
		{"Peer/Groups", peerInfo},
		{"Metric", fmt.Sprintf("%d", route.Metric)},
		{"Masquerade", fmt.Sprintf("%v", route.Masquerade)},
		{"Enabled", fmt.Sprintf("%v", route.Enabled)},
		{"Groups", groupsStr},
		{"Description", route.Description},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  e: edit  d: delete  r: refresh"))

	return b.String()
}

func (r *RoutesPage) resolveGroupNames(ids []string) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := r.groupNames[id]; ok {
			names = append(names, name)
		} else {
			names = append(names, id)
		}
	}
	return strings.Join(names, ", ")
}

func deleteRoute(c *client.Client, routeID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/routes/"+url.PathEscape(routeID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete route"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Route deleted"}
	}
}
