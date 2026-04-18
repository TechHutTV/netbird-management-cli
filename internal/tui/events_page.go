package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

const eventsPageSize = 50

// EventsPage displays audit log events in a timeline feed
type EventsPage struct {
	events      []models.AuditEvent
	filtered    []models.AuditEvent
	cursor      int
	page        int // current page (0-indexed)
	loading     bool
	err         error
	focused     bool
	search      string
	searching   bool
	autoRefresh bool
	lastRefresh time.Time
}

func NewEventsPage() *EventsPage {
	return &EventsPage{loading: true, autoRefresh: true}
}

func (e *EventsPage) totalPages() int {
	if len(e.filtered) == 0 {
		return 1
	}
	return (len(e.filtered) + eventsPageSize - 1) / eventsPageSize
}

func (e *EventsPage) pageStart() int {
	return e.page * eventsPageSize
}

func (e *EventsPage) pageEnd() int {
	end := e.pageStart() + eventsPageSize
	if end > len(e.filtered) {
		end = len(e.filtered)
	}
	return end
}

func (e *EventsPage) pageEvents() []models.AuditEvent {
	start := e.pageStart()
	end := e.pageEnd()
	if start >= len(e.filtered) {
		return nil
	}
	return e.filtered[start:end]
}

func (e *EventsPage) Title() string        { return "Events" }
func (e *EventsPage) CursorPosition() int   { return e.cursor }
func (e *EventsPage) SetFocused(focused bool) { e.focused = focused }

func (e *EventsPage) Init(c *client.Client) tea.Cmd {
	if len(e.events) > 0 {
		if e.autoRefresh {
			return eventsTickCmd()
		}
		return nil
	}
	e.loading = true
	e.err = nil
	cmds := []tea.Cmd{FetchEvents(c)}
	if e.autoRefresh {
		cmds = append(cmds, eventsTickCmd())
	}
	return tea.Batch(cmds...)
}

func (e *EventsPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case EventsTickMsg:
		if e.focused && e.autoRefresh {
			e.lastRefresh = time.Now()
			return e, tea.Batch(FetchEvents(c), eventsTickCmd())
		}
		if e.autoRefresh {
			return e, eventsTickCmd()
		}
		return e, nil

	case EventsLoadedMsg:
		e.loading = false
		if msg.Err != nil {
			e.err = msg.Err
			return e, nil
		}
		e.events = msg.Events
		e.lastRefresh = time.Now()
		e.applyFilter()
		return e, nil

	case APIErrorMsg:
		e.err = msg.Err
		return e, nil

	case tea.KeyPressMsg:
		return e.handleKey(msg, c)
	}

	return e, nil
}

func (e *EventsPage) View(width, height int) string {
	if e.loading {
		return loadingStyle.Render("  Loading events...")
	}
	if e.err != nil {
		return errorStyle.Render("  Error: " + e.err.Error())
	}

	return e.viewFeed(width, height)
}

func (e *EventsPage) applyFilter() {
	if e.search == "" {
		e.filtered = e.events
	} else {
		filtered := make([]models.AuditEvent, 0)
		for _, event := range e.events {
			if matchesQuery(e.search, event.Activity, event.InitiatorName, event.InitiatorEmail, eventTargetName(event)) {
				filtered = append(filtered, event)
			}
		}
		e.filtered = filtered
	}
	e.cursor = 0
	e.page = 0
}

func (e *EventsPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if e.searching {
		newSearch, still, changed := handleSearchKey(key, e.search)
		e.search = newSearch
		e.searching = still
		if changed || !still {
			e.applyFilter()
		}
		return e, nil
	}

	pageEvts := e.pageEvents()

	switch key {
	case "up", "k":
		if e.cursor > 0 {
			e.cursor--
		}
	case "down", "j":
		if e.cursor < len(pageEvts)-1 {
			e.cursor++
		}
	case "n", "]":
		if e.page < e.totalPages()-1 {
			e.page++
			e.cursor = 0
		}
	case "N", "[":
		if e.page > 0 {
			e.page--
			e.cursor = 0
		}
	case "r":
		e.loading = true
		return e, FetchEvents(c)
	case "/":
		e.searching = true
		e.search = ""
	case "esc":
		if e.search != "" {
			e.search = ""
			e.cursor = 0
			e.page = 0
			e.applyFilter()
		}
	case "p":
		e.autoRefresh = !e.autoRefresh
		if e.autoRefresh {
			return e, eventsTickCmd()
		}
	}

	return e, nil
}

// Styles for the event feed
var (
	eventCardStyle = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "│"}).
		BorderForeground(colorBorder).
		PaddingLeft(1).
		MarginLeft(1)

	eventCardSelectedStyle = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.Border{Left: "┃"}).
		BorderForeground(colorOrange).
		PaddingLeft(1).
		MarginLeft(1)

	eventInitiatorStyle = lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true)

	eventEmailStyle = lipgloss.NewStyle().
		Foreground(colorTextDim)

	eventTimestampStyle = lipgloss.NewStyle().
		Foreground(colorFaint)

	eventActivityStyle = lipgloss.NewStyle().
		Foreground(colorText)

	eventBadgeStyle = lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true)
)

func (e *EventsPage) viewFeed(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Audit Events (%d)", len(e.filtered))) + "\n")

	if e.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(e.search) + "\u2588\n")
	} else if e.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", e.search)) + "\n")
	}

	if len(e.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No events found."))
		return b.String()
	}

	pageEvts := e.pageEvents()
	if len(pageEvts) == 0 {
		e.page = 0
		pageEvts = e.pageEvents()
	}

	// Viewport scroll within the current page
	linesPerEvent := 3
	maxVisible := (height - 7) / linesPerEvent
	if maxVisible < 1 {
		maxVisible = 1
	}

	offset := 0
	if e.cursor >= maxVisible {
		offset = e.cursor - maxVisible + 1
	}
	end := offset + maxVisible
	if end > len(pageEvts) {
		end = len(pageEvts)
	}

	contentWidth := width - 6
	if contentWidth < 40 {
		contentWidth = 40
	}

	for i := offset; i < end; i++ {
		event := pageEvts[i]
		selected := i == e.cursor && e.focused

		// Header line: Name  email  ·  timestamp
		initiatorName := event.InitiatorName
		if initiatorName == "" {
			initiatorName = "System"
		}

		timestamp := formatEventTimestamp(event.Timestamp)

		header := eventInitiatorStyle.Render(initiatorName)
		if event.InitiatorEmail != "" {
			header += "  " + eventEmailStyle.Render(event.InitiatorEmail)
		}

		// Right-align timestamp
		headerPlain := initiatorName
		if event.InitiatorEmail != "" {
			headerPlain += "  " + event.InitiatorEmail
		}
		gap := contentWidth - len(headerPlain) - len(timestamp)
		if gap < 2 {
			gap = 2
		}
		header += strings.Repeat(" ", gap) + eventTimestampStyle.Render(timestamp)

		// Activity line with entity badges
		activity := formatEventActivity(event)

		// Render card
		cardContent := header + "\n" + activity
		if selected {
			b.WriteString(eventCardSelectedStyle.Render(cardContent) + "\n")
		} else {
			b.WriteString(eventCardStyle.Render(cardContent) + "\n")
		}
	}

	globalIdx := e.pageStart() + e.cursor + 1
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  page %d/%d  n/N: page  /: search  p: pause  r: refresh", globalIdx, len(e.filtered), e.page+1, e.totalPages())) + "\n")

	if e.autoRefresh && !e.lastRefresh.IsZero() {
		ago := time.Since(e.lastRefresh).Truncate(time.Second)
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  auto-refresh: %s ago  (p to pause)", ago)))
	} else if !e.autoRefresh {
		b.WriteString(dimHintStyle.Render("  auto-refresh paused  (p to resume)"))
	}

	return b.String()
}

// formatEventTimestamp formats an ISO timestamp into a readable form like "Apr 2, 2026 at 10:55 AM"
func formatEventTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Try without timezone
		t, err = time.Parse("2006-01-02T15:04:05", ts[:min(len(ts), 19)])
		if err != nil {
			if len(ts) > 19 {
				return strings.Replace(ts[:19], "T", " ", 1)
			}
			return ts
		}
	}
	return t.Local().Format("Jan 2, 2006 at 3:04 PM")
}

// badge wraps a string in the highlighted badge style.
func badge(s string) string {
	if s == "" {
		return ""
	}
	return eventBadgeStyle.Render(s)
}

// metaStr safely extracts a string from the meta map.
func metaStr(meta map[string]interface{}, key string) string {
	if meta == nil {
		return ""
	}
	if v, ok := meta[key].(string); ok {
		return v
	}
	return ""
}

// formatEventActivity builds a rich activity description from the activity code and meta,
// mimicking the NetBird dashboard's event display.
func formatEventActivity(event models.AuditEvent) string {
	m := event.Meta
	code := event.ActivityCode

	// Extract common meta fields
	name := metaStr(m, "name")
	hostname := metaStr(m, "hostname")
	ip := metaStr(m, "ip")
	fqdn := metaStr(m, "fqdn")
	peerIP := metaStr(m, "peer_ip")
	sourceIP := metaStr(m, "source_ip")
	city := metaStr(m, "city")
	country := metaStr(m, "country")
	email := metaStr(m, "email")
	userName := metaStr(m, "username")
	networkRange := metaStr(m, "network_range")
	networkName := metaStr(m, "network_name")

	// Build location string
	location := ""
	if city != "" && country != "" {
		location = country + ", " + city
	} else if country != "" {
		location = country
	}

	// Peer identifier: prefer hostname, then name, then fqdn
	peerLabel := hostname
	if peerLabel == "" {
		peerLabel = name
	}
	if peerLabel == "" {
		peerLabel = fqdn
	}

	// User label
	userLabel := userName
	if userLabel == "" {
		userLabel = email
	}
	if userLabel == "" {
		userLabel = event.InitiatorName
	}

	// Build rich description based on activity code patterns
	parts := strings.SplitN(code, ".", 2)
	resource := ""
	action := ""
	if len(parts) == 2 {
		resource = parts[0]
		action = parts[1]
	}

	var desc string

	switch {
	// Peer events
	case resource == "peer" && action == "login":
		desc = "Peer " + badge(peerLabel)
		if peerIP != "" {
			desc += " " + badge(peerIP)
		} else if ip != "" {
			desc += " " + badge(ip)
		}
		desc += " logged in"
		if sourceIP != "" {
			desc += " from " + badge(sourceIP)
		}
		if location != "" {
			desc += " - " + badge(location)
		}

	case resource == "peer" && (action == "delete" || action == "remove"):
		desc = "Peer " + badge(peerLabel)
		if peerIP != "" {
			desc += " " + badge(peerIP)
		} else if ip != "" {
			desc += " " + badge(ip)
		}
		desc += " was deleted"
		if sourceIP != "" {
			desc += " from " + badge(sourceIP)
		}
		if location != "" {
			desc += " - " + badge(location)
		}

	case resource == "peer" && (action == "add" || action == "register"):
		desc = "Peer " + badge(peerLabel)
		if peerIP != "" {
			desc += " " + badge(peerIP)
		} else if ip != "" {
			desc += " " + badge(ip)
		}
		desc += " was registered"
		if location != "" {
			desc += " from " + badge(location)
		}

	case resource == "peer":
		desc = "Peer " + badge(peerLabel)
		if peerIP != "" || ip != "" {
			desc += " " + badge(peerIP+ip)
		}
		desc += " " + event.Activity

	// User events
	case resource == "user" && action == "login":
		desc = badge(userLabel) + " logged in to the dashboard"

	case resource == "user" && action == "join":
		desc = "User " + badge(userLabel) + " joined NetBird"

	case resource == "user" && action == "invite":
		desc = "User " + badge(userLabel) + " was invited"

	case resource == "user" && (action == "block" || action == "blocked"):
		desc = "User " + badge(userLabel) + " was blocked"

	case resource == "user" && (action == "unblock" || action == "unblocked"):
		desc = "User " + badge(userLabel) + " was unblocked"

	case resource == "user" && action == "role":
		desc = "User " + badge(userLabel) + " role updated"

	case resource == "user":
		desc = "User " + badge(userLabel) + " " + event.Activity

	// Route events
	case resource == "route" && action == "add":
		desc = "Route " + badge(name) + " created"
		if networkRange != "" {
			desc += " for " + badge(networkRange)
		}

	case resource == "route" && (action == "delete" || action == "remove"):
		desc = "Route " + badge(name) + " deleted"

	case resource == "route" && action == "update":
		desc = "Route " + badge(name) + " updated"

	case resource == "route":
		desc = "Route " + badge(name) + " " + event.Activity

	// Group membership events (group.peer.add, group.resource.remove, etc.)
	case resource == "group" && strings.Contains(action, "."):
		subParts := strings.SplitN(action, ".", 2)
		memberType := subParts[0]  // "peer", "resource", etc.
		memberAction := subParts[1] // "add", "remove", etc.

		// The group name is typically in "name", the member is in meta fields
		groupName := name
		memberName := metaStr(m, "peer_name")
		if memberName == "" {
			memberName = metaStr(m, "resource_name")
		}
		if memberName == "" {
			memberName = hostname
		}
		if memberName == "" {
			memberName = metaStr(m, "target_name")
		}

		memberLabel := strings.Title(memberType)
		if memberName != "" {
			desc = memberLabel + " " + badge(memberName)
		} else {
			desc = memberLabel
		}

		switch memberAction {
		case "add":
			desc += " added to group " + badge(groupName)
		case "remove", "delete":
			desc += " removed from group " + badge(groupName)
		default:
			desc += " " + memberAction + " in group " + badge(groupName)
		}

	// Group CRUD events
	case resource == "group" && action == "add":
		desc = "Group " + badge(name) + " created"

	case resource == "group" && (action == "delete" || action == "remove"):
		desc = "Group " + badge(name) + " deleted"

	case resource == "group" && action == "update":
		desc = "Group " + badge(name) + " updated"

	case resource == "group":
		desc = "Group " + badge(name) + " " + event.Activity

	// Policy events
	case resource == "policy" && action == "add":
		desc = "Policy " + badge(name) + " created"

	case resource == "policy" && (action == "delete" || action == "remove"):
		desc = "Policy " + badge(name) + " deleted"

	case resource == "policy" && action == "update":
		desc = "Policy " + badge(name) + " updated"

	case resource == "policy":
		desc = "Policy " + badge(name) + " " + event.Activity

	// Network events
	case resource == "network" && action == "add":
		desc = "Network " + badge(name) + " created"

	case resource == "network" && (action == "delete" || action == "remove"):
		desc = "Network " + badge(name) + " deleted"

	case resource == "network" && action == "update":
		desc = "Network " + badge(name) + " updated"

	case strings.HasPrefix(code, "network.resource"):
		desc = "Resource " + badge(name)
		if action == "resource.add" {
			desc += " created"
		} else if action == "resource.delete" || action == "resource.remove" {
			desc += " deleted"
		} else {
			desc += " updated"
		}
		if networkName != "" {
			desc += " for network " + badge(networkName)
		}

	case strings.HasPrefix(code, "network.router"):
		desc = "Router"
		if name != "" {
			desc += " " + badge(name)
		}
		if strings.HasSuffix(code, "add") {
			desc += " added"
		} else if strings.HasSuffix(code, "delete") || strings.HasSuffix(code, "remove") {
			desc += " removed"
		} else {
			desc += " updated"
		}
		if networkName != "" {
			desc += " for network " + badge(networkName)
		}

	case resource == "network":
		desc = "Network " + badge(name) + " " + event.Activity

	// Setup key events
	case resource == "setupkey" || resource == "setup-key" || resource == "setup_key":
		desc = "Setup key " + badge(name) + " " + event.Activity

	// DNS events
	case resource == "dns" || resource == "nameserver":
		desc = "DNS " + badge(name) + " " + event.Activity

	// Posture check events
	case resource == "posture" || strings.HasPrefix(code, "posture"):
		desc = "Posture check " + badge(name) + " " + event.Activity

	// Account events
	case resource == "account":
		desc = event.Activity

	// Service user events
	case strings.HasPrefix(code, "service"):
		desc = "Service user " + badge(userLabel) + " " + event.Activity

	// Billing/system events
	case resource == "billing" || resource == "integration":
		desc = event.Activity

	// Fallback: use the Activity field as-is with any available entity highlighted
	default:
		desc = event.Activity
		if name != "" && strings.Contains(desc, name) {
			desc = strings.ReplaceAll(desc, name, badge(name))
		} else if name != "" {
			desc += " " + badge(name)
		}
	}

	return desc
}

// eventTargetName extracts a human-readable name from an event's meta or target ID.
func eventTargetName(event models.AuditEvent) string {
	if event.Meta != nil {
		for _, key := range []string{"name", "hostname", "fqdn", "email", "username"} {
			if v := metaStr(event.Meta, key); v != "" {
				return v
			}
		}
	}
	id := event.TargetID
	if len(id) > 12 {
		return id[:12] + "..."
	}
	return id
}
