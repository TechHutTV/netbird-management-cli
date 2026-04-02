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

type networksViewState int

const (
	networksViewList networksViewState = iota
	networksViewDetail
	networksViewForm
	networksViewAddResource
	networksViewAddRouter
	networksViewConfirm
	networksViewConfirmResource
	networksViewConfirmRouter
)

// NetworksPage manages networks list and detail views with resources and routers
type NetworksPage struct {
	networks  []models.Network
	filtered  []models.Network
	cursor    int
	loading   bool
	err       error
	state     networksViewState
	resources []models.NetworkResource
	routers   []models.NetworkRouter
	form           *huh.Form
	formData       networkFormData
	resFormData    networkResourceFormData
	routerFormData networkRouterFormData
	groupNames     map[string]string
	focused        bool
	search         string
	searching      bool
}

func NewNetworksPage() *NetworksPage {
	return &NetworksPage{loading: true}
}

func (n *NetworksPage) Title() string { return "Networks" }
func (n *NetworksPage) CursorPosition() int { return n.cursor }
func (n *NetworksPage) SetFocused(focused bool) { n.focused = focused }

type networksDataLoadedMsg struct {
	networks   []models.Network
	groupNames map[string]string
	err        error
}

func fetchNetworksData(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		groupResp, err := c.MakeRequest("GET", "/groups", nil)
		if err != nil {
			return networksDataLoadedMsg{err: err}
		}
		defer groupResp.Body.Close()
		var allGroups []models.PolicyGroup
		if err := jsonDecode(groupResp.Body, &allGroups); err != nil {
			return networksDataLoadedMsg{err: err}
		}
		nameMap := make(map[string]string, len(allGroups))
		for _, g := range allGroups {
			nameMap[g.ID] = g.Name
		}

		netResp, err := c.MakeRequest("GET", "/networks", nil)
		if err != nil {
			return networksDataLoadedMsg{err: err}
		}
		defer netResp.Body.Close()
		var networks []models.Network
		if err := jsonDecode(netResp.Body, &networks); err != nil {
			return networksDataLoadedMsg{err: err}
		}

		return networksDataLoadedMsg{networks: networks, groupNames: nameMap}
	}
}

func (n *NetworksPage) Init(c *client.Client) tea.Cmd {
	n.loading = true
	n.err = nil
	return fetchNetworksData(c)
}

func (n *NetworksPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle confirm screens
	if n.state == networksViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				n.state = networksViewList
				return n, submitNetworkCreate(c, n.formData)
			case "n", "esc":
				n.state = networksViewList
			}
		}
		return n, nil
	}
	if n.state == networksViewConfirmResource {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				netID := n.filtered[n.cursor].ID
				n.state = networksViewDetail
				return n, submitNetworkResource(c, netID, n.resFormData)
			case "n", "esc":
				n.state = networksViewDetail
			}
		}
		return n, nil
	}
	if n.state == networksViewConfirmRouter {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				netID := n.filtered[n.cursor].ID
				n.state = networksViewDetail
				return n, submitNetworkRouter(c, netID, n.routerFormData)
			case "n", "esc":
				n.state = networksViewDetail
			}
		}
		return n, nil
	}

	// Delegate to form when active
	// Form delegation for create network, add resource, add router
	if (n.state == networksViewForm || n.state == networksViewAddResource || n.state == networksViewAddRouter) && n.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			if n.state == networksViewForm {
				n.state = networksViewList
			} else {
				n.state = networksViewDetail
			}
			n.form = nil
			return n, nil
		}
		m, cmd := n.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			n.form = f
		}
		if n.form.State == huh.StateCompleted {
			switch n.state {
			case networksViewForm:
				n.state = networksViewConfirm
				n.form = nil
				return n, nil
			case networksViewAddResource:
				n.state = networksViewConfirmResource
				n.form = nil
				return n, nil
			case networksViewAddRouter:
				n.state = networksViewConfirmRouter
				n.form = nil
				return n, nil
			}
		}
		if n.form.State == huh.StateAborted {
			if n.state == networksViewForm {
				n.state = networksViewList
			} else {
				n.state = networksViewDetail
			}
			n.form = nil
		}
		return n, cmd
	}

	switch msg := msg.(type) {
	case networksDataLoadedMsg:
		n.loading = false
		if msg.err != nil {
			n.err = msg.err
			return n, nil
		}
		n.networks = msg.networks
		n.groupNames = msg.groupNames
		n.applyFilter()
		return n, nil

	case networkDetailLoadedMsg:
		if msg.err != nil {
			n.err = msg.err
			return n, nil
		}
		n.resources = msg.resources
		n.routers = msg.routers
		n.state = networksViewDetail
		return n, nil

	case formCompleteMsg:
		return n, n.Init(c)

	case ToastMsg:
		return n, n.Init(c)

	case APIErrorMsg:
		n.err = msg.Err
		return n, nil

	case tea.KeyPressMsg:
		return n.handleKey(msg, c)
	}

	return n, nil
}

func (n *NetworksPage) View(width, height int) string {
	if n.loading {
		return loadingStyle.Render("  Loading networks...")
	}
	if n.err != nil {
		return errorStyle.Render("  Error: " + n.err.Error())
	}

	if n.state == networksViewConfirm {
		return RenderConfirm("Confirm: Create Network", []ConfirmField{
			{Label: "Name", Value: n.formData.name},
			{Label: "Description", Value: n.formData.description},
		})
	}
	if n.state == networksViewConfirmResource {
		return RenderConfirm("Confirm: Add Resource", []ConfirmField{
			{Label: "Name", Value: n.resFormData.name},
			{Label: "Address", Value: n.resFormData.address},
			{Label: "Description", Value: n.resFormData.description},
			{Label: "Groups", Value: n.resolveGroupNames(n.resFormData.selectedGroups)},
		})
	}
	if n.state == networksViewConfirmRouter {
		return RenderConfirm("Confirm: Add Routing Peer", []ConfirmField{
			{Label: "Peer Groups", Value: n.resolveGroupNames(n.routerFormData.selectedPeerGroups)},
			{Label: "Metric", Value: n.routerFormData.metric},
			{Label: "Masquerade", Value: fmt.Sprintf("%v", n.routerFormData.masquerade)},
		})
	}
	if n.state == networksViewForm && n.form != nil {
		return pageTitleStyle.Render("Create Network") + "\n\n" + n.form.View()
	}
	if n.state == networksViewAddResource && n.form != nil {
		return pageTitleStyle.Render("Add Resource") + "\n\n" + n.form.View()
	}
	if n.state == networksViewAddRouter && n.form != nil {
		return pageTitleStyle.Render("Add Routing Peer") + "\n\n" + n.form.View()
	}
	if n.state == networksViewDetail {
		return n.viewDetail(width)
	}
	return n.viewList(width, height)
}

func (n *NetworksPage) applyFilter() {
	if n.search == "" {
		n.filtered = n.networks
	} else {
		filtered := make([]models.Network, 0)
		for _, net := range n.networks {
			if matchesQuery(n.search, net.Name, net.Description) {
				filtered = append(filtered, net)
			}
		}
		n.filtered = filtered
	}
	n.cursor = clampCursor(n.cursor, len(n.filtered))
}

func (n *NetworksPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if n.state == networksViewDetail {
		switch key {
		case "esc", "backspace", "q":
			n.state = networksViewList
			n.resources = nil
			n.routers = nil
		case "s":
			// Add resource
			if len(n.filtered) > 0 {
				n.resFormData = networkResourceFormData{}
				n.form = newNetworkResourceForm(&n.resFormData, n.groupNames)
				n.state = networksViewAddResource
				return n, n.form.Init()
			}
		case "p":
			// Add routing peer
			if len(n.filtered) > 0 {
				n.routerFormData = networkRouterFormData{}
				n.form = newNetworkRouterForm(&n.routerFormData, n.groupNames)
				n.state = networksViewAddRouter
				return n, n.form.Init()
			}
		case "r":
			// Refresh detail
			if len(n.filtered) > 0 {
				return n, fetchNetworkDetail(c, n.filtered[n.cursor].ID)
			}
		}
		return n, nil
	}

	// Search input handling (list view only)
	if n.searching {
		newSearch, still, changed := handleSearchKey(key, n.search)
		n.search = newSearch
		n.searching = still
		if changed || !still {
			n.applyFilter()
		}
		return n, nil
	}

	switch key {
	case "up", "k":
		if n.cursor > 0 {
			n.cursor--
		}
	case "down", "j":
		if n.cursor < len(n.filtered)-1 {
			n.cursor++
		}
	case "enter":
		if len(n.filtered) > 0 {
			net := n.filtered[n.cursor]
			return n, fetchNetworkDetail(c, net.ID)
		}
	case "r":
		return n, n.Init(c)
	case "c":
		n.formData = networkFormData{}
		n.form = newNetworkCreateForm(&n.formData)
		n.state = networksViewForm
		return n, n.form.Init()
	case "d":
		if len(n.filtered) > 0 {
			net := n.filtered[n.cursor]
			return n, deleteNetwork(c, net.ID)
		}
	case "/":
		n.searching = true
		n.search = ""
	case "esc":
		if n.search != "" {
			n.search = ""
			n.cursor = 0
			n.applyFilter()
		}
	}

	return n, nil
}

func (n *NetworksPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Networks (%d)", len(n.filtered))) + "\n")

	if n.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(n.search) + "\u2588\n")
	} else if n.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", n.search)) + "\n")
	}

	if len(n.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No networks found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if n.cursor >= maxRows {
		offset = n.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(n.filtered) {
		end = len(n.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		net := n.filtered[i]
		rows = append(rows, []string{
			net.Name,
			fmt.Sprintf("%d", len(net.Routers)),
			fmt.Sprintf("%d", len(net.Resources)),
			fmt.Sprintf("%d", len(net.Policies)),
			net.Description,
		})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "ROUTING PEERS", "RESOURCES", "POLICIES", "DESCRIPTION").
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

			if dataIdx == n.cursor && n.focused {
				base = tableSelectedStyle
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  c: create", n.cursor+1, len(n.filtered))))

	return b.String()
}

func (n *NetworksPage) viewDetail(width int) string {
	if n.cursor >= len(n.filtered) {
		return "No network selected"
	}

	net := n.filtered[n.cursor]
	var b strings.Builder

	b.WriteString(detailTitleStyle.Render("  "+net.Name) + "\n\n")

	fields := []struct{ label, value string }{
		{"ID", net.ID},
		{"Name", net.Name},
		{"Description", net.Description},
		{"Routing Peers", fmt.Sprintf("%d", len(net.Routers))},
		{"Resources", fmt.Sprintf("%d", len(net.Resources))},
		{"Policies", fmt.Sprintf("%d", len(net.Policies))},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	// Resources sub-section
	if len(n.resources) > 0 {
		b.WriteString("\n" + sectionHeaderStyle.Render("  Resources") + "\n")
		for _, res := range n.resources {
			enabled := disabledStyle.Render("disabled")
			if res.Enabled {
				enabled = enabledStyle.Render("enabled")
			}
			b.WriteString(fmt.Sprintf("    %s  %s  %s  %s\n",
				detailValueStyle.Render(res.Name),
				dimHintStyle.Render(res.Address),
				dimHintStyle.Render(res.Type),
				enabled))
		}
	}

	// Routing peers sub-section
	if len(n.routers) > 0 {
		b.WriteString("\n" + sectionHeaderStyle.Render("  Routing Peers") + "\n")
		for _, router := range n.routers {
			peerInfo := router.Peer
			if len(router.PeerGroups) > 0 {
				peerInfo = strings.Join(router.PeerGroups, ", ")
			}
			masq := "no-masq"
			if router.Masquerade {
				masq = "masq"
			}
			b.WriteString(fmt.Sprintf("    %s  metric:%s  %s\n",
				detailValueStyle.Render(peerInfo),
				dimHintStyle.Render(fmt.Sprintf("%d", router.Metric)),
				dimHintStyle.Render(masq)))
		}
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  s: add resource  p: add routing peer  d: delete  r: refresh"))

	return b.String()
}

// resolveGroupNames converts a slice of group IDs to a comma-separated string of names
func (n *NetworksPage) resolveGroupNames(ids []string) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := n.groupNames[id]; ok {
			names = append(names, name)
		} else {
			names = append(names, id)
		}
	}
	return strings.Join(names, ", ")
}

type networkDetailLoadedMsg struct {
	resources []models.NetworkResource
	routers   []models.NetworkRouter
	err       error
}

func fetchNetworkDetail(c *client.Client, networkID string) tea.Cmd {
	return func() tea.Msg {
		// Fetch resources
		resp, err := c.MakeRequest("GET", "/networks/"+url.PathEscape(networkID)+"/resources", nil)
		if err != nil {
			return networkDetailLoadedMsg{err: err}
		}
		defer resp.Body.Close()

		var resources []models.NetworkResource
		if err := jsonDecode(resp.Body, &resources); err != nil {
			return networkDetailLoadedMsg{err: err}
		}

		// Fetch routers
		resp2, err := c.MakeRequest("GET", "/networks/"+url.PathEscape(networkID)+"/routers", nil)
		if err != nil {
			return networkDetailLoadedMsg{resources: resources, err: nil}
		}
		defer resp2.Body.Close()

		var routers []models.NetworkRouter
		if err := jsonDecode(resp2.Body, &routers); err != nil {
			return networkDetailLoadedMsg{resources: resources, err: nil}
		}

		return networkDetailLoadedMsg{resources: resources, routers: routers}
	}
}

func deleteNetwork(c *client.Client, networkID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/networks/"+url.PathEscape(networkID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete network"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Network deleted"}
	}
}
