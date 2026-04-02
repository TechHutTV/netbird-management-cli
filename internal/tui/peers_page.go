package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

type peersViewState int

const (
	peersViewList peersViewState = iota
	peersViewDetail
	peersViewAccessible
	peersViewEdit
	peersViewEditConfirm
)

// PeersPage manages the peers list and detail views
type PeersPage struct {
	peers           []models.Peer
	filtered        []models.Peer
	cursor          int
	loading         bool
	err             error
	state           peersViewState
	search          string
	searching       bool
	accessiblePeers []models.Peer
	focused         bool
	editForm        *huh.Form
	editData        peerEditFormData
	editPeerID      string
	daemon          *DaemonClient
	connections     map[string]PeerConnectionInfo
	autoRefresh     bool
	lastRefresh     time.Time
}

func NewPeersPage() *PeersPage {
	return &PeersPage{loading: true, autoRefresh: true}
}

// SetDaemon sets the daemon client for peer connection info
func (p *PeersPage) SetDaemon(d *DaemonClient) {
	p.daemon = d
}

func (p *PeersPage) Title() string { return "Peers" }
func (p *PeersPage) CursorPosition() int { return p.cursor }
func (p *PeersPage) SetFocused(focused bool) { p.focused = focused }

func (p *PeersPage) Init(c *client.Client) tea.Cmd {
	p.loading = true
	p.err = nil
	cmds := []tea.Cmd{FetchPeers(c)}
	if p.autoRefresh {
		cmds = append(cmds, peersTickCmd())
	}
	return tea.Batch(cmds...)
}

func (p *PeersPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle edit confirm screen
	if p.state == peersViewEditConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				p.state = peersViewList
				return p, submitPeerEdit(c, p.editPeerID, p.editData)
			case "n", "esc":
				p.state = peersViewDetail
			}
		}
		return p, nil
	}

	// Delegate to edit form when active
	if p.state == peersViewEdit && p.editForm != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			p.state = peersViewDetail
			p.editForm = nil
			return p, nil
		}
		m, cmd := p.editForm.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			p.editForm = f
		}
		if p.editForm.State == huh.StateCompleted {
			p.state = peersViewEditConfirm
			p.editForm = nil
			return p, nil
		}
		if p.editForm.State == huh.StateAborted {
			p.state = peersViewDetail
			p.editForm = nil
		}
		return p, cmd
	}

	switch msg := msg.(type) {
	case PeerUpdatedMsg:
		if msg.Err != nil {
			p.err = msg.Err
			return p, nil
		}
		return p, p.Init(c)

	case PeersTickMsg:
		if p.focused && p.autoRefresh && p.state == peersViewList {
			p.lastRefresh = time.Now()
			return p, tea.Batch(FetchPeers(c), peersTickCmd())
		}
		if p.autoRefresh {
			return p, peersTickCmd() // keep ticking, refresh when focused on list
		}
		return p, nil

	case ConnectionTickMsg:
		if p.focused && p.state == peersViewDetail && p.daemon != nil && p.autoRefresh {
			return p, tea.Batch(FetchPeerConnections(p.daemon), connectionTickCmd())
		}
		return p, nil

	case PeersLoadedMsg:
		p.loading = false
		if msg.Err != nil {
			p.err = msg.Err
			return p, nil
		}
		p.peers = msg.Peers
		p.lastRefresh = time.Now()
		p.applyFilter()
		return p, nil

	case PeerConnectionMsg:
		if msg.Err == nil {
			p.connections = msg.Connections
		}
		return p, nil

	case accessiblePeersLoadedMsg:
		if msg.err != nil {
			p.err = msg.err
			return p, nil
		}
		p.accessiblePeers = msg.peers
		p.state = peersViewAccessible
		return p, nil

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

func (p *PeersPage) View(width, height int) string {
	if p.loading {
		return loadingStyle.Render("  Loading peers...")
	}
	if p.err != nil {
		return errorStyle.Render("  Error: " + p.err.Error())
	}

	switch p.state {
	case peersViewEditConfirm:
		return RenderConfirm("Confirm: Edit Peer", []ConfirmField{
			{Label: "Name", Value: p.editData.name},
			{Label: "SSH Enabled", Value: fmt.Sprintf("%v", p.editData.sshEnabled)},
			{Label: "Login Expiration", Value: fmt.Sprintf("%v", p.editData.loginExpirationEnabled)},
			{Label: "Inactivity Expiration", Value: fmt.Sprintf("%v", p.editData.inactivityExpirationEnabled)},
		})
	case peersViewEdit:
		if p.editForm != nil {
			return pageTitleStyle.Render("Edit Peer") + "\n\n" + p.editForm.View()
		}
		return p.viewDetail(width)
	case peersViewDetail:
		return p.viewDetail(width)
	case peersViewAccessible:
		return p.viewAccessible(width, height)
	default:
		return p.viewList(width, height)
	}
}

func (p *PeersPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	// Search input mode
	if p.searching {
		newSearch, still, changed := handleSearchKey(key, p.search)
		p.search = newSearch
		p.searching = still
		if changed || !still {
			p.applyFilter()
		}
		return p, nil
	}

	// Detail view — e to edit, esc to go back
	if p.state == peersViewDetail {
		switch key {
		case "esc", "backspace", "q":
			p.state = peersViewList
		case "e":
			if len(p.filtered) > 0 {
				peer := p.filtered[p.cursor]
				p.editData = peerEditFormData{
					name:                        peer.Name,
					sshEnabled:                  peer.SSHEnabled,
					loginExpirationEnabled:      peer.LoginExpirationEnabled,
					inactivityExpirationEnabled: peer.InactivityExpirationEnabled,
				}
				p.editForm = newPeerEditForm(&p.editData)
				p.editPeerID = peer.ID
				p.state = peersViewEdit
				return p, p.editForm.Init()
			}
		}
		return p, nil
	}

	// Accessible view — esc to go back
	if p.state == peersViewAccessible {
		if key == "esc" || key == "backspace" || key == "q" {
			p.state = peersViewList
			p.accessiblePeers = nil
		}
		return p, nil
	}

	// List view keys
	switch key {
	case "esc":
		// Clear active filter if any
		if p.search != "" {
			p.search = ""
			p.cursor = 0
			p.applyFilter()
		}
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
			p.state = peersViewDetail
			if p.daemon != nil {
				cmds := []tea.Cmd{FetchPeerConnections(p.daemon)}
				if p.autoRefresh {
					cmds = append(cmds, connectionTickCmd())
				}
				return p, tea.Batch(cmds...)
			}
		}
	case "r":
		return p, p.Init(c)
	case "d":
		if len(p.filtered) > 0 {
			peer := p.filtered[p.cursor]
			return p, DeletePeer(c, peer.ID)
		}
	case "a":
		if len(p.filtered) > 0 {
			peer := p.filtered[p.cursor]
			return p, FetchAccessiblePeers(c, peer.ID)
		}
	case "/":
		p.searching = true
		p.search = ""
	case "p":
		p.autoRefresh = !p.autoRefresh
		if p.autoRefresh {
			return p, peersTickCmd()
		}
	}

	return p, nil
}

func (p *PeersPage) applyFilter() {
	if p.search == "" {
		p.filtered = p.peers
	} else {
		filtered := make([]models.Peer, 0)
		for _, peer := range p.peers {
			if matchesQuery(p.search, peer.Name, peer.IP, peer.Hostname) {
				filtered = append(filtered, peer)
			}
		}
		p.filtered = filtered
	}
	p.cursor = clampCursor(p.cursor, len(p.filtered))
}

func (p *PeersPage) viewList(width, height int) string {
	var b strings.Builder

	// Header with count
	online := 0
	for _, peer := range p.filtered {
		if peer.Connected {
			online++
		}
	}
	header := pageTitleStyle.Render(fmt.Sprintf("Peers  %s  %s",
		onlineStyle.Render(fmt.Sprintf("%d online", online)),
		offlineStyle.Render(fmt.Sprintf("%d offline", len(p.filtered)-online))))
	b.WriteString(header + "\n")

	// Search bar
	if p.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(p.search) + "█\n")
	} else if p.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", p.search)) + "\n")
	}

	if len(p.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No peers found."))
		return b.String()
	}

	// Build table rows
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
		peer := p.filtered[i]

		status := "Offline"
		if peer.Connected {
			status = "Online"
		}

		rows = append(rows, []string{
			peer.Name,
			peer.IP,
			status,
			peer.OS,
			peer.Version,
			formatLastSeen(peer.LastSeen),
		})
	}

	// Build the lipgloss table
	tableWidth := width
	if tableWidth > 100 {
		tableWidth = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "IP", "STATUS", "OS", "VERSION", "LAST SEEN").
		Rows(rows...).
		Width(tableWidth).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return tableHeaderStyle
			}

			// Determine actual data index for highlighting
			dataIdx := row + offset
			base := tableCellStyle
			if row%2 == 0 {
				base = tableDimCellStyle
			}

			// Cursor row
			if dataIdx == p.cursor && p.focused {
				base = tableSelectedStyle
			}

			// Status column coloring
			if col == 2 && row >= 0 && row < len(rows) {
				if rows[row][2] == "Online" {
					return base.Foreground(colorSuccess)
				}
				return base.Foreground(colorDanger)
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	hints := fmt.Sprintf("  %d/%d  /: search", p.cursor+1, len(p.filtered))
	if p.search != "" {
		hints += "  esc: clear filter"
	}
	hints += "  a: accessible  d: delete  r: refresh"
	b.WriteString(dimHintStyle.Render(hints) + "\n")

	if p.autoRefresh && !p.lastRefresh.IsZero() {
		ago := time.Since(p.lastRefresh).Truncate(time.Second)
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  auto-refresh: %s ago  (p to pause)", ago)))
	} else if !p.autoRefresh {
		b.WriteString(dimHintStyle.Render("  auto-refresh paused  (p to resume)"))
	}

	return b.String()
}

func (p *PeersPage) viewAccessible(width, height int) string {
	if p.cursor >= len(p.filtered) {
		return "No peer selected"
	}

	peer := p.filtered[p.cursor]
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Peers accessible from: %s", peer.Name)) + "\n")

	if len(p.accessiblePeers) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No accessible peers found."))
		b.WriteString("\n\n" + dimHintStyle.Render("  esc: back"))
		return b.String()
	}

	rows := make([][]string, 0, len(p.accessiblePeers))
	for _, ap := range p.accessiblePeers {
		status := "Offline"
		if ap.Connected {
			status = "Online"
		}
		rows = append(rows, []string{ap.Name, ap.IP, status, ap.OS})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "IP", "STATUS", "OS").
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
			if col == 2 && row >= 0 && row < len(rows) {
				if rows[row][2] == "Online" {
					return base.Foreground(colorSuccess)
				}
				return base.Foreground(colorDanger)
			}
			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d accessible peers  esc: back", len(p.accessiblePeers))))
	return b.String()
}

func (p *PeersPage) viewDetail(width int) string {
	if p.cursor >= len(p.filtered) {
		return "No peer selected"
	}

	peer := p.filtered[p.cursor]
	var b strings.Builder

	// Title with status badge
	status := offlineStyle.Render(" OFFLINE ")
	if peer.Connected {
		status = onlineStyle.Render(" ONLINE ")
	}

	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("  %s  %s", peer.Name, status)))
	b.WriteString("\n\n")

	// Detail fields
	fields := []struct{ label, value string }{
		{"ID", peer.ID},
		{"Name", peer.Name},
		{"IP Address", peer.IP},
		{"Hostname", peer.Hostname},
		{"OS", peer.OS},
		{"Version", peer.Version},
		{"Last Seen", peer.LastSeen},
		{"SSH Enabled", fmt.Sprintf("%v", peer.SSHEnabled)},
		{"Login Exp.", fmt.Sprintf("%v", peer.LoginExpirationEnabled)},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	// Groups sub-section
	if len(peer.Groups) > 0 {
		b.WriteString("\n" + sectionHeaderStyle.Render("  Groups") + "\n")
		for _, g := range peer.Groups {
			b.WriteString(fmt.Sprintf("    %s  %s\n",
				detailValueStyle.Render(g.Name),
				dimHintStyle.Render(g.ID)))
		}
	}

	// Connection detail from local daemon
	if conn, ok := p.connections[peer.IP]; ok {
		b.WriteString("\n" + sectionHeaderStyle.Render("  Connection Detail") + "\n")
		connFields := []struct{ label, value string }{
			{"Type", conn.ConnType},
			{"Remote Endpoint", conn.RemoteEndpoint},
			{"Local ICE", conn.LocalICEType},
			{"Remote ICE", conn.RemoteICEType},
			{"Latency", conn.Latency},
			{"Sent", formatBytes(conn.BytesSent)},
			{"Received", formatBytes(conn.BytesReceived)},
			{"Last Handshake", conn.LastHandshake},
		}
		for _, f := range connFields {
			if f.value == "" {
				continue
			}
			label := detailLabelStyle.Render(f.label)
			value := detailValueStyle.Render(f.value)
			b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
		}
		if conn.RelayAddress != "" {
			label := detailLabelStyle.Render("Relay")
			value := detailValueStyle.Render(conn.RelayAddress)
			b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
		}
	} else if p.daemon != nil {
		b.WriteString("\n" + dimHintStyle.Render("  Peer not connected locally") + "\n")
	} else {
		b.WriteString("\n" + dimHintStyle.Render("  Connect to NetBird daemon for connection details") + "\n")
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  e: edit  d: delete  r: refresh"))

	return b.String()
}
