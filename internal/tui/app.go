package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"netbird-manage/internal/client"
)

// Page is the interface that all TUI pages must implement
type Page interface {
	Init(c *client.Client) tea.Cmd
	Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd)
	View(width, height int) string
	Title() string
	CursorPosition() int     // returns current cursor index in list (0 = top)
	SetFocused(focused bool) // called when page gains/loses focus
}

// TabCycler is an optional interface for pages that control tab behavior.
// AcceptTab returns true if the page has focusable content (called when entering from nav).
// CycleTab advances to the next section, returning true to stay on the page, false to return to nav.
type TabCycler interface {
	AcceptTab() bool
	CycleTab() bool
}

// focusArea tracks which part of the UI has keyboard focus
type focusArea int

const (
	focusNav focusArea = iota
	focusContent
)

// App is the root model for the TUI application
type App struct {
	client *client.Client
	daemon *DaemonClient
	nav    NavModel
	pages  map[Section]Page
	focus  focusArea
	width  int
	height int
	toast  string
	ready  bool
}

// NewApp creates a new App model
func NewApp(c *client.Client, daemon *DaemonClient) App {
	p := make(map[Section]Page)
	for _, item := range navItems {
		p[item.Section] = newPlaceholderPage(item.Label)
	}
	// Dashboard (first tab)
	p[SectionDashboard] = NewDashboardPage(daemon)
	// Management pages
	peersPage := NewPeersPage()
	peersPage.SetDaemon(daemon)
	p[SectionPeers] = peersPage
	p[SectionGroups] = NewGroupsPage()
	p[SectionNetworks] = NewNetworksPage()
	p[SectionPolicies] = NewPoliciesPage()
	p[SectionRoutes] = NewRoutesPage()
	p[SectionSetupKeys] = NewSetupKeysPage()
	p[SectionUsers] = NewUsersPage()
	p[SectionServiceUsers] = NewServiceUsersPage()
	p[SectionDNS] = NewDNSPage()
	p[SectionPostureChecks] = NewPostureChecksPage()
	p[SectionEvents] = NewEventsPage()
	p[SectionSettings] = NewSettingsPage()
	p[SectionExportImport] = NewExportImportPage()
	p[SectionReverseProxy] = NewReverseProxyPage()

	return App{
		client: c,
		daemon: daemon,
		nav:    NewNavModel().SetFocused(true),
		pages:  p,
		focus:  focusNav,
	}
}

// Init implements tea.Model
func (a App) Init() tea.Cmd {
	cmds := []tea.Cmd{pageRefreshTickCmd()}
	active := a.nav.Active()
	if page, ok := a.pages[active]; ok {
		cmds = append(cmds, page.Init(a.client))
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		return a, nil

	case tea.KeyPressMsg:
		return a.handleKeyPress(msg)

	case NavSelectMsg:
		return a.switchPage(msg.Section)

	case ToastMsg:
		a.toast = msg.Message
		return a, nil

	case APIErrorMsg:
		a.toast = "Error: " + msg.Err.Error()
		return a, nil

	case PageRefreshTickMsg:
		// Refresh active page in background every 60s
		cmds := []tea.Cmd{pageRefreshTickCmd()}
		active := a.nav.Active()
		if page, ok := a.pages[active]; ok {
			updated, cmd := page.Update(msg, a.client)
			if cmd != nil {
				newPages := make(map[Section]Page, len(a.pages))
				for k, v := range a.pages {
					newPages[k] = v
				}
				newPages[active] = updated
				a.pages = newPages
				cmds = append(cmds, cmd)
			}
		}
		return a, tea.Batch(cmds...)
	}

	return a.updateActivePage(msg)
}

// View implements tea.Model
func (a App) View() tea.View {
	if !a.ready {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	header := a.renderHeader()
	navbar := a.nav.View(a.width)

	contentWidth := a.width - 4
	contentHeight := a.height - headerHeight - navBarHeight - 1 - 2
	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentHeight < 5 {
		contentHeight = 5
	}

	content := ""
	activePage, ok := a.pages[a.nav.Active()]
	if ok {
		content = activePage.View(contentWidth, contentHeight)
	}

	sectionName := ""
	if ok {
		sectionName = activePage.Title()
	}

	hints := statusHints(a.focus)
	if a.toast != "" {
		hints = a.toast
	}

	layout := renderLayout(header, navbar, content, a.width, a.height)
	statusBar := renderStatusBar(sectionName, hints, a.width)

	v := tea.NewView(layout + "\n" + statusBar)
	v.AltScreen = true
	return v
}

// handleKeyPress processes global key bindings
func (a App) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == keyForceQuit {
		return a, tea.Quit
	}

	// Number hotkeys jump to tab but keep focus in nav bar
	if a.focus == focusNav {
		if nav, section, ok := a.nav.SelectByHotkey(key); ok {
			a.nav = nav
			if page, ok := a.pages[section]; ok {
				page.SetFocused(false)
			}
			return a.switchPage(section)
		}
	}

	if key == keyTab {
		if page, ok := a.pages[a.nav.Active()]; ok {
			if tc, ok := page.(TabCycler); ok {
				if a.focus == focusNav {
					if tc.AcceptTab() {
						a.focus = focusContent
						a.nav = a.nav.SetFocused(false)
						page.SetFocused(true)
						return a, nil
					}
					return a, nil
				}
				// Content focused: cycle sections, return to nav when done
				if tc.CycleTab() {
					return a, nil
				}
				return a.toggleFocus(), nil
			}
		}
		return a.toggleFocus(), nil
	}

	// Nav bar focused
	if a.focus == focusNav {
		switch key {
		case keyQuit:
			return a, tea.Quit
		case "left", "h":
			a.nav = a.nav.MoveLeft()
			nav, section := a.nav.Select()
			a.nav = nav
			if page, ok := a.pages[section]; ok {
				page.SetFocused(false)
			}
			return a.switchPage(section)
		case "right", "l":
			a.nav = a.nav.MoveRight()
			nav, section := a.nav.Select()
			a.nav = nav
			if page, ok := a.pages[section]; ok {
				page.SetFocused(false)
			}
			return a.switchPage(section)
		case keyEnter:
			// Select tab and load page, but stay in nav bar (page unfocused)
			nav, section := a.nav.Select()
			a.nav = nav
			if page, ok := a.pages[section]; ok {
				page.SetFocused(false)
			}
			return a.switchPage(section)
		case "down", "j":
			// Drop into page content (page focused)
			nav, section := a.nav.Select()
			a.nav = nav
			if section == SectionDashboard {
				return a.switchPage(section)
			}
			a.focus = focusContent
			a.nav = a.nav.SetFocused(false)
			if page, ok := a.pages[section]; ok {
				page.SetFocused(true)
			}
			return a.switchPage(section)
		}
		return a, nil
	}

	// Content focused
	if key == keyQuit {
		return a.toggleFocus(), nil
	}

	// Up arrow at cursor 0 goes back to nav bar; otherwise page handles it
	if key == "up" {
		if page, ok := a.pages[a.nav.Active()]; ok && page.CursorPosition() == 0 {
			return a.toggleFocus(), nil
		}
	}

	return a.updateActivePage(msg)
}

// toggleFocus switches focus between nav bar and content
func (a App) toggleFocus() App {
	active := a.nav.Active()
	if a.focus == focusNav {
		a.focus = focusContent
		a.nav = a.nav.SetFocused(false)
		if page, ok := a.pages[active]; ok {
			page.SetFocused(true)
		}
	} else {
		a.focus = focusNav
		a.nav = a.nav.SetFocused(true)
		if page, ok := a.pages[active]; ok {
			page.SetFocused(false)
		}
	}
	return a
}

// switchPage activates a new section and initializes its page
func (a App) switchPage(section Section) (tea.Model, tea.Cmd) {
	a.toast = ""
	if page, ok := a.pages[section]; ok {
		cmd := page.Init(a.client)
		return a, cmd
	}
	return a, nil
}

// updateActivePage delegates a message to the currently active page
func (a App) updateActivePage(msg tea.Msg) (tea.Model, tea.Cmd) {
	active := a.nav.Active()
	if page, ok := a.pages[active]; ok {
		updatedPage, cmd := page.Update(msg, a.client)
		newPages := make(map[Section]Page, len(a.pages))
		for k, v := range a.pages {
			newPages[k] = v
		}
		newPages[active] = updatedPage
		a.pages = newPages
		return a, cmd
	}
	return a, nil
}

// headerHeight is the height consumed by the top header bar
const headerHeight = 2

// renderHeader renders the top header bar with management server info
func (a App) renderHeader() string {
	title := lipgloss.NewStyle().
		Foreground(colorOrange).
		Bold(true).
		Padding(0, 1).
		Render("NetBird Management")

	// Extract display URL (strip /api suffix)
	url := a.client.ManagementURL
	url = strings.TrimSuffix(url, "/api")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	sep := lipgloss.NewStyle().Foreground(colorBorder).Render("  │  ")
	status := onlineStyle.Render("● Connected") + sep +
		detailValueStyle.Render(url)

	headerContent := title + sep + status

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false).
		Width(a.width).
		Padding(0, 0).
		Render(headerContent)
}

// statusHints returns context-appropriate key binding hints
func statusHints(focus focusArea) string {
	if focus == focusNav {
		return "1-9,0: jump  ←/→: navigate  enter: select  tab: content  q: quit"
	}
	return "1-9,0: jump  tab/q: nav  esc: back  /: search  c: create  d: delete  r: refresh"
}
