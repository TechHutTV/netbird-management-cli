package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

type ingressViewState int

const (
	ingressViewList ingressViewState = iota
	ingressViewDetail
)

// ingressPeersLoadedMsg carries the result of fetching ingress peers
type ingressPeersLoadedMsg struct {
	peers []models.IngressPeer
	err   error
}

// IngressPage manages ingress peers (cloud-only)
type IngressPage struct {
	peers     []models.IngressPeer
	filtered  []models.IngressPeer
	cursor    int
	loading   bool
	err       error
	state     ingressViewState
	focused   bool
	search    string
	searching bool
}

func NewIngressPage() *IngressPage {
	return &IngressPage{loading: true}
}

func (ip *IngressPage) Title() string { return "Ingress" }
func (ip *IngressPage) CursorPosition() int { return ip.cursor }
func (ip *IngressPage) SetFocused(focused bool) { ip.focused = focused }

func (ip *IngressPage) Init(c *client.Client) tea.Cmd {
	ip.loading = true
	ip.err = nil
	return fetchIngressPeers(c)
}

func (ip *IngressPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case ingressPeersLoadedMsg:
		ip.loading = false
		if msg.err != nil {
			ip.err = msg.err
			return ip, nil
		}
		ip.peers = msg.peers
		ip.applyFilter()
		return ip, nil

	case APIErrorMsg:
		ip.err = msg.Err
		return ip, nil

	case tea.KeyPressMsg:
		return ip.handleKey(msg, c)
	}

	return ip, nil
}

func (ip *IngressPage) View(width, height int) string {
	if ip.loading {
		return loadingStyle.Render("  Loading ingress peers...")
	}
	if ip.err != nil {
		errStr := ip.err.Error()
		if strings.Contains(errStr, "404") {
			return dimHintStyle.Render("  Ingress is a Cloud-only feature.\n  Not available on self-hosted instances.")
		}
		return errorStyle.Render("  Error: " + errStr)
	}

	if ip.state == ingressViewDetail {
		return ip.viewDetail(width)
	}
	return ip.viewList(width, height)
}

func (ip *IngressPage) applyFilter() {
	if ip.search == "" {
		ip.filtered = ip.peers
	} else {
		filtered := make([]models.IngressPeer, 0)
		for _, peer := range ip.peers {
			if matchesQuery(ip.search, peer.Name) {
				filtered = append(filtered, peer)
			}
		}
		ip.filtered = filtered
	}
	ip.cursor = clampCursor(ip.cursor, len(ip.filtered))
}

func (ip *IngressPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if ip.state == ingressViewDetail {
		if key == "esc" || key == "backspace" || key == "q" {
			ip.state = ingressViewList
		}
		return ip, nil
	}

	// Search input handling (list view only)
	if ip.searching {
		newSearch, still, changed := handleSearchKey(key, ip.search)
		ip.search = newSearch
		ip.searching = still
		if changed || !still {
			ip.applyFilter()
		}
		return ip, nil
	}

	switch key {
	case "up", "k":
		if ip.cursor > 0 {
			ip.cursor--
		}
	case "down", "j":
		if ip.cursor < len(ip.filtered)-1 {
			ip.cursor++
		}
	case "enter":
		if len(ip.filtered) > 0 {
			ip.state = ingressViewDetail
		}
	case "r":
		return ip, ip.Init(c)
	case "/":
		ip.searching = true
		ip.search = ""
	case "esc":
		if ip.search != "" {
			ip.search = ""
			ip.cursor = 0
			ip.applyFilter()
		}
	}

	return ip, nil
}

func (ip *IngressPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Ingress Peers (%d)", len(ip.filtered))) + "\n")

	if ip.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(ip.search) + "\u2588\n")
	} else if ip.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", ip.search)) + "\n")
	}

	if len(ip.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No ingress peers found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if ip.cursor >= maxRows {
		offset = ip.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(ip.filtered) {
		end = len(ip.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		peer := ip.filtered[i]

		enabled := "No"
		if peer.Enabled {
			enabled = "Yes"
		}

		rows = append(rows, []string{
			peer.Name,
			peer.Location,
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
		Headers("NAME", "LOCATION", "ENABLED").
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

			if dataIdx == ip.cursor && ip.focused {
				base = tableSelectedStyle
			}

			// Enabled column coloring
			if col == 2 && row >= 0 && row < len(rows) {
				if rows[row][2] == "Yes" {
					return base.Foreground(colorSuccess)
				}
				return base.Foreground(colorDanger)
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  r: refresh", ip.cursor+1, len(ip.filtered))))

	return b.String()
}

func (ip *IngressPage) viewDetail(width int) string {
	if ip.cursor >= len(ip.filtered) {
		return "No peer selected"
	}

	peer := ip.filtered[ip.cursor]
	var b strings.Builder

	// Title with enabled badge
	badge := disabledStyle.Render(" DISABLED ")
	if peer.Enabled {
		badge = enabledStyle.Render(" ENABLED ")
	}

	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("  %s  %s", peer.Name, badge)))
	b.WriteString("\n\n")

	fields := []struct{ label, value string }{
		{"ID", peer.ID},
		{"Name", peer.Name},
		{"Location", peer.Location},
		{"Hostname", peer.Hostname},
		{"Enabled", fmt.Sprintf("%v", peer.Enabled)},
		{"Created At", peer.CreatedAt},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  r: refresh"))

	return b.String()
}

func fetchIngressPeers(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/ingress/peers", nil)
		if err != nil {
			return ingressPeersLoadedMsg{err: err}
		}
		defer resp.Body.Close()

		var peers []models.IngressPeer
		if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
			return ingressPeersLoadedMsg{err: fmt.Errorf("decode ingress peers: %w", err)}
		}
		return ingressPeersLoadedMsg{peers: peers}
	}
}
