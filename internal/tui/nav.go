package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// Section represents a navigation section
type Section int

const (
	SectionDashboard Section = iota
	SectionPeers
	SectionGroups
	SectionNetworks
	SectionPolicies
	SectionRoutes
	SectionSetupKeys
	SectionUsers
	SectionServiceUsers
	SectionDNS
	SectionPostureChecks
	SectionEvents
	SectionSettings
	SectionIngress
	SectionExportImport
)

// NavItem holds display info for a navigation section
type NavItem struct {
	Section Section
	Label   string
	Hotkey  string // "1"-"9", "0", or "" for no hotkey
}

var navItems = []NavItem{
	{SectionDashboard, "Status", "1"},
	{SectionPeers, "Peers", "2"},
	{SectionPolicies, "Policies", "3"},
	{SectionNetworks, "Networks", "4"},
	{SectionDNS, "DNS", "5"},
	{SectionRoutes, "Routes", "6"},
	{SectionUsers, "Users", "7"},
	{SectionGroups, "Groups", "8"},
	{SectionPostureChecks, "Posture", "9"},
	{SectionSetupKeys, "Keys", "0"},
	{SectionEvents, "Events", ""},
	{SectionServiceUsers, "Service Users", ""},
	{SectionSettings, "Settings", ""},
	{SectionIngress, "Ingress", ""},
	{SectionExportImport, "Export", ""},
}

// hotkeyMap maps key strings to section indices for quick lookup
var hotkeyMap = buildHotkeyMap()

func buildHotkeyMap() map[string]int {
	m := make(map[string]int)
	for i, item := range navItems {
		if item.Hotkey != "" {
			m[item.Hotkey] = i
		}
	}
	return m
}

// NavModel manages top bar navigation state
type NavModel struct {
	items   []NavItem
	cursor  int
	active  Section
	focused bool
	counts  map[Section]int
}

func NewNavModel() NavModel {
	return NavModel{
		items:  navItems,
		cursor: 0,
		active: SectionDashboard,
		counts: make(map[Section]int),
	}
}

func (n NavModel) SetFocused(focused bool) NavModel {
	n.focused = focused
	return n
}

func (n NavModel) SetCount(section Section, count int) NavModel {
	newCounts := make(map[Section]int, len(n.counts))
	for k, v := range n.counts {
		newCounts[k] = v
	}
	newCounts[section] = count
	n.counts = newCounts
	return n
}

func (n NavModel) MoveLeft() NavModel {
	if n.cursor > 0 {
		n.cursor--
	}
	return n
}

func (n NavModel) MoveRight() NavModel {
	if n.cursor < len(n.items)-1 {
		n.cursor++
	}
	return n
}

func (n NavModel) Select() (NavModel, Section) {
	n.active = n.items[n.cursor].Section
	return n, n.active
}

// SelectByHotkey jumps to a section by its hotkey. Returns updated model, section, and whether the key matched.
func (n NavModel) SelectByHotkey(key string) (NavModel, Section, bool) {
	idx, ok := hotkeyMap[key]
	if !ok {
		return n, n.active, false
	}
	n.cursor = idx
	n.active = n.items[idx].Section
	return n, n.active, true
}

func (n NavModel) Active() Section { return n.active }
func (n NavModel) Focused() bool   { return n.focused }

// View renders the horizontal top navigation bar with numbered hotkeys
func (n NavModel) View(width int) string {
	brand := navBrandStyle.Render("NetBird")

	var tabs []string
	for i, item := range n.items {
		label := item.Label
		if item.Hotkey != "" {
			label = fmt.Sprintf("[%s]%s", item.Hotkey, item.Label)
		}

		if count, ok := n.counts[item.Section]; ok && count > 0 {
			label += fmt.Sprintf(" %d", count)
		}

		switch {
		case n.focused && i == n.cursor:
			tabs = append(tabs, navTabActiveStyle.Render(label))
		case item.Section == n.active:
			tabs = append(tabs, navTabSelectedStyle.Render(label))
		default:
			tabs = append(tabs, navTabStyle.Render(label))
		}
	}

	tabContent := brand + "  " + strings.Join(tabs, "  ")

	// Single container with bottom border — no per-tab borders
	return navBarContainerStyle.
		Width(width).
		Render(tabContent)
}

// ─── Nav bar styles ─────────────────────────────────────────────────

var (
	navBrandStyle = lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true).
			Padding(0, 1)

	navTabStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Padding(0, 1)

	navTabSelectedStyle = lipgloss.NewStyle().
				Foreground(colorOrange).
				Bold(true).
				Padding(0, 1)

	navTabActiveStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorSelected).
				Bold(true).
				Padding(0, 1)

	// Container wraps all tabs with a single bottom border
	navBarContainerStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colorBorder).
				BorderBottom(true).
				BorderTop(false).
				BorderLeft(false).
				BorderRight(false).
				Padding(0, 0)
)
