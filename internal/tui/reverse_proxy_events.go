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

// rpEventsPageSize is how many events load per request.
const rpEventsPageSize = 100

// rpEventsView holds state for the per-service proxy events sub-view.
type rpEventsView struct {
	serviceID string
	events    []models.ReverseProxyEvent
	cursor    int
	loading   bool
	err       error
	focused   bool

	// Date-range filter, expressed as day offsets from today (0 = today, -7 = 7d ago).
	startOffsetDays int
	endOffsetDays   int
}

func (e *rpEventsView) setServiceID(id string) {
	if e.serviceID != id {
		e.events = nil
		e.cursor = 0
		e.startOffsetDays = -7
		e.endOffsetDays = 0
	}
	e.serviceID = id
}

func (e *rpEventsView) load(c *client.Client) tea.Cmd {
	e.loading = true
	e.err = nil
	start, end := e.dateRange()
	return FetchReverseProxyEvents(c, e.serviceID, start, end, 0, rpEventsPageSize)
}

func (e *rpEventsView) dateRange() (string, string) {
	now := time.Now().UTC()
	start := now.AddDate(0, 0, e.startOffsetDays)
	end := now.AddDate(0, 0, e.endOffsetDays)
	return start.Format(time.RFC3339), end.Format(time.RFC3339)
}

// update processes a bubbletea message for the events sub-view.
// Returns: command, whether the sub-view wants to close.
func (e *rpEventsView) update(msg tea.Msg, c *client.Client) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case ReverseProxyEventsLoadedMsg:
		// Ignore late-arriving responses for a previously-viewed service.
		if msg.ServiceID != e.serviceID {
			return nil, false
		}
		e.loading = false
		if msg.Err != nil {
			e.err = msg.Err
			return nil, false
		}
		e.events = msg.Events
		if e.cursor >= len(e.events) {
			e.cursor = 0
		}
		return nil, false
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, false
	}
	switch keyMsg.String() {
	case "esc":
		return nil, true
	case "up", "k":
		if e.cursor > 0 {
			e.cursor--
		}
	case "down", "j":
		if e.cursor < len(e.events)-1 {
			e.cursor++
		}
	case "r":
		return e.load(c), false
	case "1":
		e.startOffsetDays = -1
		return e.load(c), false
	case "2":
		e.startOffsetDays = -7
		return e.load(c), false
	case "3":
		e.startOffsetDays = -30
		return e.load(c), false
	}
	return nil, false
}

func (e *rpEventsView) view(width int) string {
	var b strings.Builder
	start, end := e.dateRange()
	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Proxy Events (%d)", len(e.events))) + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %s  →  %s", start, end)) + "\n")

	if e.loading {
		b.WriteString(loadingStyle.Render("  Loading...") + "\n")
		return b.String()
	}
	if e.err != nil {
		b.WriteString(errorStyle.Render("  Error: "+e.err.Error()) + "\n")
	}

	if len(e.events) == 0 {
		b.WriteString(dimHintStyle.Render("  No events in this range.") + "\n")
	} else {
		rows := make([][]string, 0, len(e.events))
		for _, ev := range e.events {
			rows = append(rows, []string{
				formatEventTime(ev.Timestamp),
				defaultStr(ev.Method, "-"),
				truncate(ev.Host+ev.Path, 35),
				eventStatusCell(ev.StatusCode),
				defaultStr(ev.AuthMethodUsed, "-"),
				formatLocation(ev.CountryCode, ev.CityName),
				formatBytes(ev.BytesUpload + ev.BytesDownload),
				fmt.Sprintf("%dms", ev.DurationMs),
			})
		}

		tw := width
		if tw > 130 {
			tw = 130
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(tableBorderStyle).
			Headers("TIME", "METHOD", "HOST+PATH", "STATUS", "AUTH", "LOCATION", "BYTES", "DURATION").
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
				if row == e.cursor && e.focused {
					return tableSelectedStyle
				}
				if col == 3 && row < len(rows) {
					code := rows[row][3]
					if len(code) > 0 && code[0] == '2' {
						return base.Foreground(colorSuccess)
					}
					if len(code) > 0 && (code[0] == '4' || code[0] == '5') {
						return base.Foreground(colorDanger)
					}
				}
				return base
			})
		b.WriteString(t.Render() + "\n")
	}

	b.WriteString(dimHintStyle.Render("  esc: back  1/2/3: 1d/7d/30d range  r: refresh"))
	return b.String()
}

func formatEventTime(ts string) string {
	if ts == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02 15:04:05")
}

func eventStatusCell(code int) string {
	if code == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", code)
}
