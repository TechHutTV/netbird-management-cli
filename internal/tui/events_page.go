package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// EventsPage displays audit log events
type EventsPage struct {
	events    []models.AuditEvent
	filtered  []models.AuditEvent
	cursor      int
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

func (e *EventsPage) Title() string { return "Events" }
func (e *EventsPage) CursorPosition() int { return e.cursor }
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
			return e, eventsTickCmd() // keep ticking, refresh when focused
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

	return e.viewList(width, height)
}

func (e *EventsPage) applyFilter() {
	if e.search == "" {
		e.filtered = e.events
	} else {
		filtered := make([]models.AuditEvent, 0)
		for _, event := range e.events {
			if matchesQuery(e.search, event.Activity, event.InitiatorName, event.TargetID) {
				filtered = append(filtered, event)
			}
		}
		e.filtered = filtered
	}
	e.cursor = clampCursor(e.cursor, len(e.filtered))
}

func (e *EventsPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	// Search input handling
	if e.searching {
		newSearch, still, changed := handleSearchKey(key, e.search)
		e.search = newSearch
		e.searching = still
		if changed || !still {
			e.applyFilter()
		}
		return e, nil
	}

	switch key {
	case "up", "k":
		if e.cursor > 0 {
			e.cursor--
		}
	case "down", "j":
		if e.cursor < len(e.filtered)-1 {
			e.cursor++
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

func (e *EventsPage) viewList(width, height int) string {
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

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if e.cursor >= maxRows {
		offset = e.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(e.filtered) {
		end = len(e.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		event := e.filtered[i]

		timestamp := event.Timestamp
		if len(timestamp) > 19 {
			timestamp = strings.Replace(timestamp[:19], "T", " ", 1)
		}

		initiator := event.InitiatorEmail
		if initiator == "" {
			initiator = event.InitiatorName
		}
		if initiator == "" {
			initiator = event.InitiatorID
		}

		rows = append(rows, []string{
			timestamp,
			event.Activity,
			initiator,
			event.TargetID,
		})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("TIMESTAMP", "ACTIVITY", "INITIATOR", "TARGET ID").
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

			if dataIdx == e.cursor && e.focused {
				base = tableSelectedStyle
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  p: pause  r: refresh", e.cursor+1, len(e.filtered))) + "\n")

	if e.autoRefresh && !e.lastRefresh.IsZero() {
		ago := time.Since(e.lastRefresh).Truncate(time.Second)
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  auto-refresh: %s ago  (p to pause)", ago)))
	} else if !e.autoRefresh {
		b.WriteString(dimHintStyle.Render("  auto-refresh paused  (p to resume)"))
	}

	return b.String()
}
