package tui

import (
	"fmt"
	"net/url"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

type postureViewState int

const (
	postureViewList postureViewState = iota
	postureViewDetail
)

// PostureChecksPage manages posture checks list and detail views
type PostureChecksPage struct {
	checks    []models.PostureCheck
	filtered  []models.PostureCheck
	cursor    int
	loading   bool
	err       error
	state     postureViewState
	focused   bool
	search    string
	searching bool
}

func NewPostureChecksPage() *PostureChecksPage {
	return &PostureChecksPage{loading: true}
}

func (pc *PostureChecksPage) Title() string { return "Posture Checks" }
func (pc *PostureChecksPage) CursorPosition() int { return pc.cursor }
func (pc *PostureChecksPage) SetFocused(focused bool) { pc.focused = focused }

func (pc *PostureChecksPage) Init(c *client.Client) tea.Cmd {
	pc.loading = true
	pc.err = nil
	return FetchPostureChecks(c)
}

func (pc *PostureChecksPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case PostureChecksLoadedMsg:
		pc.loading = false
		if msg.Err != nil {
			pc.err = msg.Err
			return pc, nil
		}
		pc.checks = msg.Checks
		pc.applyFilter()
		return pc, nil

	case ToastMsg:
		return pc, pc.Init(c)

	case APIErrorMsg:
		pc.err = msg.Err
		return pc, nil

	case tea.KeyPressMsg:
		return pc.handleKey(msg, c)
	}

	return pc, nil
}

func (pc *PostureChecksPage) View(width, height int) string {
	if pc.loading {
		return loadingStyle.Render("  Loading posture checks...")
	}
	if pc.err != nil {
		return errorStyle.Render("  Error: " + pc.err.Error())
	}

	if pc.state == postureViewDetail {
		return pc.viewDetail(width)
	}
	return pc.viewList(width, height)
}

func (pc *PostureChecksPage) applyFilter() {
	if pc.search == "" {
		pc.filtered = pc.checks
	} else {
		filtered := make([]models.PostureCheck, 0)
		for _, check := range pc.checks {
			if matchesQuery(pc.search, check.Name) {
				filtered = append(filtered, check)
			}
		}
		pc.filtered = filtered
	}
	pc.cursor = clampCursor(pc.cursor, len(pc.filtered))
}

func (pc *PostureChecksPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if pc.state == postureViewDetail {
		if key == "esc" || key == "backspace" || key == "q" {
			pc.state = postureViewList
		}
		return pc, nil
	}

	// Search input handling (list view only)
	if pc.searching {
		newSearch, still, changed := handleSearchKey(key, pc.search)
		pc.search = newSearch
		pc.searching = still
		if changed || !still {
			pc.applyFilter()
		}
		return pc, nil
	}

	switch key {
	case "up", "k":
		if pc.cursor > 0 {
			pc.cursor--
		}
	case "down", "j":
		if pc.cursor < len(pc.filtered)-1 {
			pc.cursor++
		}
	case "enter":
		if len(pc.filtered) > 0 {
			pc.state = postureViewDetail
		}
	case "r":
		return pc, pc.Init(c)
	case "d":
		if len(pc.filtered) > 0 {
			check := pc.filtered[pc.cursor]
			return pc, deletePostureCheck(c, check.ID)
		}
	case "/":
		pc.searching = true
		pc.search = ""
	case "esc":
		if pc.search != "" {
			pc.search = ""
			pc.cursor = 0
			pc.applyFilter()
		}
	}

	return pc, nil
}

func (pc *PostureChecksPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Posture Checks (%d)", len(pc.filtered))) + "\n")

	if pc.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(pc.search) + "\u2588\n")
	} else if pc.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", pc.search)) + "\n")
	}

	if len(pc.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No posture checks found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if pc.cursor >= maxRows {
		offset = pc.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(pc.filtered) {
		end = len(pc.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		check := pc.filtered[i]
		rows = append(rows, []string{
			check.Name,
			postureCheckType(check),
			check.Description,
		})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "TYPE", "DESCRIPTION").
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

			if dataIdx == pc.cursor && pc.focused {
				base = tableSelectedStyle
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  ", pc.cursor+1, len(pc.filtered))))

	return b.String()
}

func (pc *PostureChecksPage) viewDetail(width int) string {
	if pc.cursor >= len(pc.filtered) {
		return "No check selected"
	}

	check := pc.filtered[pc.cursor]
	var b strings.Builder

	b.WriteString(detailTitleStyle.Render("  "+check.Name) + "\n\n")

	fields := []struct{ label, value string }{
		{"ID", check.ID},
		{"Name", check.Name},
		{"Type", postureCheckType(check)},
		{"Description", check.Description},
	}

	// Add type-specific details
	if check.Checks.NBVersionCheck != nil {
		fields = append(fields, struct{ label, value string }{
			"Min NB Version", check.Checks.NBVersionCheck.MinVersion})
	}
	if check.Checks.GeoLocationCheck != nil {
		geo := check.Checks.GeoLocationCheck
		locs := make([]string, len(geo.Locations))
		for i, loc := range geo.Locations {
			if loc.CityName != "" {
				locs[i] = loc.CountryCode + "/" + loc.CityName
			} else {
				locs[i] = loc.CountryCode
			}
		}
		fields = append(fields,
			struct{ label, value string }{"Geo Action", geo.Action},
			struct{ label, value string }{"Locations", strings.Join(locs, ", ")})
	}
	if check.Checks.PeerNetworkRangeCheck != nil {
		nr := check.Checks.PeerNetworkRangeCheck
		fields = append(fields,
			struct{ label, value string }{"Net Range Action", nr.Action},
			struct{ label, value string }{"Ranges", strings.Join(nr.Ranges, ", ")})
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  d: delete  r: refresh"))

	return b.String()
}

func postureCheckType(check models.PostureCheck) string {
	switch {
	case check.Checks.NBVersionCheck != nil:
		return "nb-version"
	case check.Checks.OSVersionCheck != nil:
		return "os-version"
	case check.Checks.GeoLocationCheck != nil:
		return "geo-location"
	case check.Checks.PeerNetworkRangeCheck != nil:
		return "network-range"
	case check.Checks.ProcessCheck != nil:
		return "process"
	default:
		return "unknown"
	}
}

func deletePostureCheck(c *client.Client, checkID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/posture-checks/"+url.PathEscape(checkID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete posture check"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Posture check deleted"}
	}
}
