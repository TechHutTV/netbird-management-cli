package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
	"github.com/netbirdio/netbird/client/proto"
)

// DashboardPage shows combined daemon status and management API overview
type DashboardPage struct {
	daemon *DaemonClient

	// Daemon data
	daemonStatus *proto.StatusResponse
	daemonConfig *proto.GetConfigResponse
	daemonErr    error

	// Management API counts
	counts   *DashboardCountsMsg
	countsErr error

	loading bool
	focused bool
}

func NewDashboardPage(daemon *DaemonClient) *DashboardPage {
	return &DashboardPage{
		daemon:  daemon,
		loading: true,
	}
}

func (d *DashboardPage) Title() string { return "Status" }
func (d *DashboardPage) CursorPosition() int { return 0 }
func (d *DashboardPage) SetFocused(focused bool) { d.focused = focused }

func (d *DashboardPage) Init(c *client.Client) tea.Cmd {
	d.loading = true
	return tea.Batch(
		d.fetchDaemonStatus(),
		fetchDashboardCounts(c),
		d.scheduleTick(),
	)
}

func (d *DashboardPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case PageRefreshTickMsg:
		// Dashboard always safe to refresh
		return d, tea.Batch(
			d.fetchDaemonStatus(),
			fetchDashboardCounts(c),
		)

	case DaemonStatusMsg:
		d.loading = false
		d.daemonErr = msg.Err
		if msg.Err == nil {
			d.daemonStatus = msg.Status
			d.daemonConfig = msg.Config
		}
		return d, nil

	case DashboardCountsMsg:
		d.loading = false
		d.countsErr = msg.Err
		if msg.Err == nil {
			d.counts = &msg
		}
		return d, nil

	case DaemonTickMsg:
		return d, tea.Batch(
			d.fetchDaemonStatus(),
			d.scheduleTick(),
		)

	case tea.KeyPressMsg:
		if msg.String() == "r" {
			return d, tea.Batch(
				d.fetchDaemonStatus(),
				fetchDashboardCounts(c),
			)
		}
	}

	return d, nil
}

func (d *DashboardPage) View(width, height int) string {
	colWidth := (width - 4) / 2
	if colWidth < 30 {
		colWidth = 30
	}

	left := d.renderDaemonStatus(colWidth)
	right := d.renderManagementOverview(colWidth)

	leftBox := lipgloss.NewStyle().
		Width(colWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2).
		Render(left)

	rightBox := lipgloss.NewStyle().
		Width(colWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2).
		Render(right)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox)

	logo := renderNetbirdLogo()

	return columns + "\n\n" + logo
}

// netbirdLogoLines is the full "netbird" wordmark with bird icon, rendered in ANSI art.
// Generated from the official NetBird logo-white.png using ansizalizer.
var netbirdLogoLines = []string{
	"         \x1b[38;2;250;132;47m▄\x1b[38;2;253;135;49m▄\x1b[38;2;251;133;48m\x1b[48;2;251;133;48m█\x1b[38;2;251;133;48m\x1b[48;2;251;133;48m██\x1b[38;2;251;133;48m\x1b[48;2;251;133;48m█\x1b[38;2;250;134;48m\x1b[48;2;250;134;48m█\x1b[0m\x1b[38;2;251;133;47m▘\x1b[0m                           \x1b[38;2;255;255;255m▗▄\x1b[0m         \x1b[38;2;255;255;255m▄▖\x1b[0m              \x1b[38;2;255;255;255m▗▄\x1b[0m  ",
	"\x1b[38;2;249;132;47m▝\x1b[38;2;252;135;49m▀\x1b[38;2;253;135;49m\x1b[48;2;253;135;49m█\x1b[38;2;252;134;49m\x1b[48;2;252;134;49m█\x1b[38;2;250;133;48m\x1b[48;2;250;133;48m█\x1b[38;2;250;133;49m\x1b[48;2;250;133;49m█\x1b[38;2;247;132;47m\x1b[48;2;247;132;47m█\x1b[38;2;247;122;48m\x1b[48;2;247;122;48m█\x1b[38;2;245;100;49m\x1b[48;2;245;100;49m█\x1b[38;2;246;112;49m\x1b[48;2;246;112;49m█\x1b[38;2;245;125;47m\x1b[48;2;245;125;47m█\x1b[38;2;246;131;48m\x1b[48;2;246;131;48m█\x1b[38;2;246;131;48m\x1b[48;2;246;131;48m█\x1b[38;2;247;131;48m\x1b[48;2;247;131;48m█\x1b[0m\x1b[38;2;250;133;48m▛\x1b[38;2;247;132;46m▘\x1b[0m   \x1b[38;2;255;255;255m▄\x1b[0m \x1b[38;2;255;255;255m▄▄▄▖\x1b[0m     \x1b[38;2;255;255;255m▗▄▄▄▖\x1b[0m   \x1b[38;2;255;255;255m▄\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▙▄▖\x1b[0m \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m \x1b[38;2;255;255;255m▗▄▄▖\x1b[0m    \x1b[38;2;255;255;255m▄▖\x1b[0m  \x1b[38;2;255;255;255m▄\x1b[0m \x1b[38;2;255;255;255m▗▄\x1b[0m   \x1b[38;2;255;255;255m▗▄▄▄\x1b[0m \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m  ",
	"   \x1b[38;2;249;131;47m▀\x1b[38;2;249;132;48m\x1b[48;2;249;132;48m█\x1b[38;2;247;130;47m\x1b[48;2;247;130;47m█\x1b[38;2;243;111;48m\x1b[48;2;243;111;48m█\x1b[38;2;242;93;50m\x1b[48;2;242;93;50m█\x1b[38;2;243;94;50m\x1b[48;2;243;94;50m█\x1b[38;2;243;94;50m\x1b[48;2;243;94;50m█\x1b[38;2;242;94;50m\x1b[48;2;242;94;50m█\x1b[38;2;244;115;48m\x1b[48;2;244;115;48m█\x1b[38;2;248;132;47m\x1b[48;2;248;132;47m█\x1b[0m\x1b[38;2;251;132;48m▛\x1b[0m    \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▛▘\x1b[0m \x1b[38;2;255;255;255m▝▀\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▖\x1b[0m \x1b[38;2;255;255;255m▗\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▀▘\x1b[0m \x1b[38;2;255;255;255m▝▀\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▖\x1b[0m \x1b[38;2;255;255;255m▝\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m   \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▛▘\x1b[0m  \x1b[38;2;255;255;255m▀\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▄\x1b[0m  \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m  \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▛▀▀\x1b[0m \x1b[38;2;255;255;255m▗\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▀▘\x1b[0m  \x1b[38;2;255;255;255m▀\x1b[48;2;255;255;255m██\x1b[0m  ",
	"     \x1b[38;2;248;122;48m\x1b[48;2;248;122;48m█\x1b[38;2;242;98;49m\x1b[48;2;242;98;49m█\x1b[38;2;243;93;50m\x1b[48;2;243;93;50m█\x1b[38;2;243;94;50m\x1b[48;2;243;94;50m█\x1b[38;2;243;94;50m\x1b[48;2;243;94;50m█\x1b[38;2;243;94;50m\x1b[48;2;243;94;50m█\x1b[38;2;248;96;50m\x1b[48;2;248;96;50m█\x1b[0m\x1b[38;2;250;129;47m▘\x1b[0m     \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m     \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▀▀▀▀▀▀▘\x1b[0m  \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m   \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m     \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m  \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m  \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m   \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m     \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m  ",
	"   \x1b[38;2;250;132;47m▗\x1b[38;2;253;135;49m\x1b[48;2;253;135;49m█\x1b[38;2;250;133;48m\x1b[48;2;250;133;48m█\x1b[38;2;250;133;48m\x1b[48;2;250;133;48m█\x1b[38;2;249;122;49m\x1b[48;2;249;122;49m█\x1b[38;2;247;104;50m\x1b[48;2;247;104;50m█\x1b[38;2;247;95;51m\x1b[48;2;247;95;51m█\x1b[0m\x1b[38;2;247;95;50m▛\x1b[0m       \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m     \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m \x1b[38;2;255;255;255m▝\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▄▖\x1b[0m  \x1b[38;2;255;255;255m▄\x1b[48;2;255;255;255m█\x1b[0m   \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▙\x1b[0m   \x1b[38;2;255;255;255m▐\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▙▖\x1b[0m  \x1b[38;2;255;255;255m▄\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▀\x1b[0m  \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m  \x1b[38;2;255;255;255m\x1b[48;2;255;255;255m█\x1b[0m\x1b[38;2;255;255;255m▌\x1b[0m   \x1b[38;2;255;255;255m▝▜▄▖\x1b[0m  \x1b[38;2;255;255;255m▄\x1b[48;2;255;255;255m██\x1b[0m  ",
	"   \x1b[38;2;249;132;47m▀\x1b[38;2;246;130;47m▀▀▀\x1b[38;2;246;130;47m▀\x1b[38;2;246;132;47m▀\x1b[38;2;247;109;48m▀\x1b[0m         \x1b[38;2;255;255;255m▀\x1b[0m     \x1b[38;2;255;255;255m▀▘\x1b[0m   \x1b[38;2;255;255;255m▝▀▀▀▘\x1b[0m     \x1b[38;2;255;255;255m▀▀▘\x1b[0m \x1b[38;2;255;255;255m▝▘\x1b[0m \x1b[38;2;255;255;255m▝▀▀▘\x1b[0m    \x1b[38;2;255;255;255m▀▘\x1b[0m  \x1b[38;2;255;255;255m▀\x1b[0m      \x1b[38;2;255;255;255m▝▀▀▀\x1b[0m  \x1b[38;2;255;255;255m▀\x1b[0m  ",
}

// renderNetbirdLogo renders the ANSI art logo centered
func renderNetbirdLogo() string {
	logo := strings.Join(netbirdLogoLines, "\n")
	return lipgloss.NewStyle().
		Foreground(colorTextDim).
		Align(lipgloss.Center).
		Render(logo)
}

// ─── Daemon status (left column) ────────────────────────────────────

func (d *DashboardPage) renderDaemonStatus(width int) string {
	var sb strings.Builder

	sb.WriteString(sectionHeaderStyle.Render("NetBird Client") + "\n\n")

	if d.daemon == nil {
		noMarginDim := lipgloss.NewStyle().Foreground(colorFaint)
		sb.WriteString(noMarginDim.Render("Daemon client not initialized.\nSet NETBIRD_SOCKET or run with sudo."))
		return sb.String()
	}

	if d.daemonErr != nil {
		noMarginDim := lipgloss.NewStyle().Foreground(colorFaint)
		sb.WriteString(offlineStyle.Render("○ Daemon not connected") + "\n")
		sb.WriteString(noMarginDim.Render("  Is NetBird running? (sudo netbird up)") + "\n")
		sb.WriteString(noMarginDim.Render("  " + d.daemonErr.Error()))
		return sb.String()
	}

	if d.daemonStatus == nil {
		sb.WriteString(loadingStyle.Render("  Loading..."))
		return sb.String()
	}

	fs := d.daemonStatus.FullStatus
	if fs == nil {
		sb.WriteString(lipgloss.NewStyle().Foreground(colorFaint).Render("  No status available"))
		return sb.String()
	}

	// Use plain fmt for fixed-width labels — lipgloss Width doesn't work
	// reliably with ANSI codes from styled values on the same line.
	labelStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	lbl := func(s string) string {
		return labelStyle.Render(fmt.Sprintf("%-18s", s))
	}
	indent := strings.Repeat(" ", 18)

	// Management connection
	if fs.ManagementState != nil && fs.ManagementState.Connected {
		url := ""
		if fs.ManagementState.URL != "" {
			url = "  " + fs.ManagementState.URL
		}
		sb.WriteString(lbl("Management:") + onlineStyle.Render("● Connected") + labelStyle.Render(url) + "\n")
	} else {
		sb.WriteString(lbl("Management:") + offlineStyle.Render("○ Disconnected") + "\n")
	}

	// Signal connection
	if fs.SignalState != nil && fs.SignalState.Connected {
		url := ""
		if fs.SignalState.URL != "" {
			url = "  " + fs.SignalState.URL
		}
		sb.WriteString(lbl("Signal:") + onlineStyle.Render("● Connected") + labelStyle.Render(url) + "\n")
	} else {
		sb.WriteString(lbl("Signal:") + offlineStyle.Render("○ Disconnected") + "\n")
	}
	sb.WriteString("\n")

	// Local peer info
	if fs.LocalPeerState != nil {
		lp := fs.LocalPeerState
		sb.WriteString(lbl("IP:") + detailValueStyle.Render(lp.IP) + "\n")
		sb.WriteString(lbl("FQDN:") + detailValueStyle.Render(lp.Fqdn) + "\n")
		if lp.KernelInterface {
			sb.WriteString(lbl("Kernel:") + onlineStyle.Render("Yes") + "\n")
		} else {
			sb.WriteString(lbl("Kernel:") + offlineStyle.Render("No") + "\n")
		}
		if lp.RosenpassEnabled {
			sb.WriteString(lbl("Rosenpass:") + onlineStyle.Render("● Enabled") + "\n")
		} else {
			sb.WriteString(lbl("Rosenpass:") + labelStyle.Render("○ Disabled") + "\n")
		}
		sb.WriteString("\n")
	}

	// Peers summary
	total := len(fs.Peers)
	online := 0
	p2p := 0
	relayed := 0
	var relayedPeers []string
	for _, p := range fs.Peers {
		if p.ConnStatus == "Connected" {
			online++
			if p.Relayed {
				relayed++
				name := p.Fqdn
				if name == "" {
					name = p.IP
				}
				relayedPeers = append(relayedPeers, name)
			} else {
				p2p++
			}
		}
	}
	sb.WriteString(lbl("Peers:") + detailValueStyle.Render(fmt.Sprintf("%d online / %d total", online, total)) + "\n")

	connVal := onlineStyle.Render(fmt.Sprintf("%d P2P", p2p))
	if relayed > 0 {
		connVal += lipgloss.NewStyle().Foreground(colorYellow).Render(fmt.Sprintf("  %d Relayed", relayed))
	} else {
		connVal += labelStyle.Render("  0 Relayed")
	}
	sb.WriteString(lbl("Connections:") + connVal + "\n")

	if len(relayedPeers) > 0 {
		sb.WriteString(lbl("Relayed:") + "\n")
		for _, name := range relayedPeers {
			sb.WriteString(indent + lipgloss.NewStyle().Foreground(colorYellow).Render("◆ ") + detailValueStyle.Render(name) + "\n")
		}
	}

	// Relays
	if len(fs.Relays) > 0 {
		sb.WriteString("\n" + lbl("Relays:") + "\n")
		for _, r := range fs.Relays {
			if r.Available {
				sb.WriteString(indent + onlineStyle.Render("● ") + detailValueStyle.Render(r.URI) + "\n")
			} else {
				sb.WriteString(indent + offlineStyle.Render("○ ") + detailValueStyle.Render(r.URI) + "\n")
			}
		}
	}

	// DNS servers
	if len(fs.DnsServers) > 0 {
		sb.WriteString("\n" + lbl("DNS Servers:") + "\n")
		for _, ns := range fs.DnsServers {
			for _, srv := range ns.Servers {
				domains := ""
				if len(ns.Domains) > 0 {
					domains = " (" + strings.Join(ns.Domains, ", ") + ")"
				}
				if ns.Error != "" {
					sb.WriteString(indent + offlineStyle.Render("○ ") + detailValueStyle.Render(srv) + errorStyle.Render(" ["+ns.Error+"]") + "\n")
				} else if ns.Enabled {
					sb.WriteString(indent + onlineStyle.Render("● ") + detailValueStyle.Render(srv) + labelStyle.Render(domains) + "\n")
				} else {
					sb.WriteString(indent + labelStyle.Render("○ ") + detailValueStyle.Render(srv) + labelStyle.Render(domains) + "\n")
				}
			}
		}
	}

	// SSH server
	if fs.SshServerState != nil {
		sb.WriteString("\n")
		if fs.SshServerState.Enabled {
			sessions := len(fs.SshServerState.Sessions)
			sessionStr := ""
			if sessions > 0 {
				sessionStr = fmt.Sprintf("  (%d sessions)", sessions)
			}
			sb.WriteString(lbl("SSH Server:") + onlineStyle.Render("● Enabled") + labelStyle.Render(sessionStr) + "\n")
		} else {
			sb.WriteString(lbl("SSH Server:") + labelStyle.Render("○ Disabled") + "\n")
		}
	}

	// System events (last 3)
	if len(fs.Events) > 0 {
		sb.WriteString("\n" + lbl("Events:") + "\n")
		events := fs.Events
		start := len(events) - 3
		if start < 0 {
			start = 0
		}
		for _, ev := range events[start:] {
			icon := "ℹ"
			evStyle := labelStyle
			switch ev.Severity {
			case proto.SystemEvent_WARNING:
				icon = "⚠"
				evStyle = lipgloss.NewStyle().Foreground(colorYellow)
			case proto.SystemEvent_ERROR, proto.SystemEvent_CRITICAL:
				icon = "✗"
				evStyle = errorStyle
			}
			ts := ""
			if ev.Timestamp != nil {
				ts = "[" + ev.Timestamp.AsTime().Local().Format("15:04") + "] "
			}
			msg := ev.UserMessage
			if msg == "" {
				msg = ev.Message
			}
			sb.WriteString("  " + evStyle.Render(icon+" "+ts+msg) + "\n")
		}
	}

	return sb.String()
}

// ─── Management overview (right column) ─────────────────────────────

func (d *DashboardPage) renderManagementOverview(width int) string {
	var sb strings.Builder

	sb.WriteString(sectionHeaderStyle.Render("Management Overview") + "\n\n")

	if d.counts == nil && d.countsErr == nil {
		sb.WriteString(loadingStyle.Render("  Loading..."))
		return sb.String()
	}

	if d.countsErr != nil {
		sb.WriteString(errorStyle.Render("  Error: " + d.countsErr.Error()))
		return sb.String()
	}

	c := d.counts

	rows := []struct{ label, value string }{
		{"Account", c.AccountDomain},
		{"Peers", fmt.Sprintf("%d online / %d total", c.PeersOnline, c.PeersTotal)},
		{"Groups", fmt.Sprintf("%d", c.Groups)},
		{"Networks", fmt.Sprintf("%d", c.Networks)},
		{"Policies", fmt.Sprintf("%d", c.Policies)},
		{"Routes", fmt.Sprintf("%d", c.Routes)},
		{"Setup Keys", fmt.Sprintf("%d active", c.SetupKeys)},
		{"DNS Groups", fmt.Sprintf("%d", c.DNSGroups)},
		{"Posture Checks", fmt.Sprintf("%d", c.Posture)},
	}

	labelStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	lbl := func(s string) string {
		return labelStyle.Render(fmt.Sprintf("%-18s", s))
	}
	for _, r := range rows {
		sb.WriteString(lbl(r.label+":") + detailValueStyle.Render(r.value) + "\n")
	}

	sb.WriteString("\n" + dimHintStyle.Render("r: refresh"))

	return sb.String()
}

// ─── Commands ───────────────────────────────────────────────────────

func (d *DashboardPage) fetchDaemonStatus() tea.Cmd {
	if d.daemon == nil {
		return func() tea.Msg {
			return DaemonStatusMsg{Err: fmt.Errorf("daemon not available")}
		}
	}
	return func() tea.Msg {
		status, err := d.daemon.Status()
		if err != nil {
			return DaemonStatusMsg{Err: err}
		}
		config, _ := d.daemon.GetConfig() // config is optional
		return DaemonStatusMsg{Status: status, Config: config}
	}
}

func (d *DashboardPage) scheduleTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return DaemonTickMsg{}
	})
}

func fetchDashboardCounts(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		// Intentionally ignoring decode errors — partial dashboard data is acceptable.
		// Each section is independent; if one fails, the count stays at zero.
		msg := DashboardCountsMsg{}

		// Peers
		if resp, err := c.MakeRequest("GET", "/peers", nil); err == nil {
			var peers []models.Peer
			json.NewDecoder(resp.Body).Decode(&peers)
			resp.Body.Close()
			msg.PeersTotal = len(peers)
			for _, p := range peers {
				if p.Connected {
					msg.PeersOnline++
				}
			}
		}

		// Groups
		if resp, err := c.MakeRequest("GET", "/groups", nil); err == nil {
			var items []json.RawMessage
			json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			msg.Groups = len(items)
		}

		// Networks
		if resp, err := c.MakeRequest("GET", "/networks", nil); err == nil {
			var items []json.RawMessage
			json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			msg.Networks = len(items)
		}

		// Policies
		if resp, err := c.MakeRequest("GET", "/policies", nil); err == nil {
			var items []json.RawMessage
			json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			msg.Policies = len(items)
		}

		// Routes
		if resp, err := c.MakeRequest("GET", "/routes", nil); err == nil {
			var items []json.RawMessage
			json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			msg.Routes = len(items)
		}

		// Setup keys (count active)
		if resp, err := c.MakeRequest("GET", "/setup-keys", nil); err == nil {
			var keys []models.SetupKey
			json.NewDecoder(resp.Body).Decode(&keys)
			resp.Body.Close()
			for _, k := range keys {
				if k.Valid && !k.Revoked {
					msg.SetupKeys++
				}
			}
		}

		// DNS
		if resp, err := c.MakeRequest("GET", "/dns/nameservers", nil); err == nil {
			var items []json.RawMessage
			json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			msg.DNSGroups = len(items)
		}

		// Posture checks
		if resp, err := c.MakeRequest("GET", "/posture-checks", nil); err == nil {
			var items []json.RawMessage
			json.NewDecoder(resp.Body).Decode(&items)
			resp.Body.Close()
			msg.Posture = len(items)
		}

		// Account domain
		if resp, err := c.MakeRequest("GET", "/accounts", nil); err == nil {
			var accounts []models.Account
			json.NewDecoder(resp.Body).Decode(&accounts)
			resp.Body.Close()
			if len(accounts) > 0 {
				msg.AccountDomain = accounts[0].Domain
			}
		}

		return msg
	}
}

// ─── Dashboard styles ───────────────────────────────────────────────

// dimHintStyle is used for labels — defined in theme.go
// Labels are padded with fmt.Sprintf("%-18s") before styling to avoid
// lipgloss Width issues with ANSI escape codes.
