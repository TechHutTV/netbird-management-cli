package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

type postureViewState int

const (
	postureViewList postureViewState = iota
	postureViewDetail
	postureViewForm
	postureViewConfirm
	postureViewEdit
	postureViewEditConfirm
)

// PostureChecksPage manages posture checks list and detail views
type PostureChecksPage struct {
	checks      []models.PostureCheck
	filtered    []models.PostureCheck
	cursor      int
	loading     bool
	err         error
	state       postureViewState
	focused     bool
	search      string
	searching   bool
	form        *huh.Form
	formData    postureCheckFormData
	editForm    *huh.Form
	editData    postureCheckFormData
	editCheckID string
}

func NewPostureChecksPage() *PostureChecksPage {
	return &PostureChecksPage{loading: true}
}

func (pc *PostureChecksPage) Title() string        { return "Posture Checks" }
func (pc *PostureChecksPage) CursorPosition() int  { return pc.cursor }
func (pc *PostureChecksPage) SetFocused(focused bool) { pc.focused = focused }

func (pc *PostureChecksPage) Init(c *client.Client) tea.Cmd {
	if len(pc.checks) > 0 {
		return nil
	}
	pc.loading = true
	pc.err = nil
	return FetchPostureChecks(c)
}

func (pc *PostureChecksPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle edit confirm screen
	if pc.state == postureViewEditConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				pc.state = postureViewList
				return pc, submitPostureCheckEdit(c, pc.editCheckID, pc.editData)
			case "n", "esc":
				pc.state = postureViewDetail
			}
		}
		return pc, nil
	}

	// Delegate to edit form when active
	if pc.state == postureViewEdit && pc.editForm != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			pc.state = postureViewDetail
			pc.editForm = nil
			return pc, nil
		}
		m, cmd := pc.editForm.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			pc.editForm = f
		}
		if pc.editForm.State == huh.StateCompleted {
			pc.state = postureViewEditConfirm
			pc.editForm = nil
			return pc, nil
		}
		if pc.editForm.State == huh.StateAborted {
			pc.state = postureViewDetail
			pc.editForm = nil
		}
		return pc, cmd
	}

	// Handle create confirm screen
	if pc.state == postureViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				pc.state = postureViewList
				return pc, submitPostureCheckCreate(c, pc.formData)
			case "n", "esc":
				pc.state = postureViewList
			}
		}
		return pc, nil
	}

	// Delegate to create form when active
	if pc.state == postureViewForm && pc.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			pc.state = postureViewList
			pc.form = nil
			return pc, nil
		}
		m, cmd := pc.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			pc.form = f
		}
		if pc.form.State == huh.StateCompleted {
			pc.state = postureViewConfirm
			pc.form = nil
			return pc, nil
		}
		if pc.form.State == huh.StateAborted {
			pc.state = postureViewList
			pc.form = nil
		}
		return pc, cmd
	}

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

	case PostureCheckUpdatedMsg:
		if msg.Err != nil {
			pc.err = msg.Err
			return pc, nil
		}
		pc.loading = true
		return pc, FetchPostureChecks(c)

	case PostureCheckDeletedMsg:
		if msg.Err != nil {
			pc.err = msg.Err
			return pc, nil
		}
		pc.state = postureViewList
		pc.loading = true
		return pc, FetchPostureChecks(c)

	case formCompleteMsg:
		pc.loading = true
		return pc, FetchPostureChecks(c)

	case ToastMsg:
		pc.loading = true
		return pc, FetchPostureChecks(c)

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

	if pc.state == postureViewEditConfirm {
		return RenderConfirm("Confirm: Edit Posture Check", postureCheckConfirmFields(pc.editData))
	}
	if pc.state == postureViewEdit && pc.editForm != nil {
		return pageTitleStyle.Render("Edit Posture Check") + "\n\n" + pc.editForm.View()
	}
	if pc.state == postureViewConfirm {
		return RenderConfirm("Confirm: Create Posture Check", postureCheckConfirmFields(pc.formData))
	}
	if pc.state == postureViewForm && pc.form != nil {
		return pageTitleStyle.Render("Create Posture Check") + "\n\n" + pc.form.View()
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
		switch key {
		case "esc", "backspace", "q":
			pc.state = postureViewList
		case "e":
			if pc.cursor < len(pc.filtered) {
				check := pc.filtered[pc.cursor]
				pc.editData = postureCheckToFormData(check)
				pc.editForm = newPostureCheckForm(&pc.editData)
				pc.editCheckID = check.ID
				pc.state = postureViewEdit
				return pc, pc.editForm.Init()
			}
		case "d":
			if pc.cursor < len(pc.filtered) {
				check := pc.filtered[pc.cursor]
				return pc, DeletePostureCheck(c, check.ID)
			}
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
		pc.loading = true
		return pc, FetchPostureChecks(c)
	case "c":
		pc.formData = postureCheckFormData{checkType: "nb_version", geoAction: "allow", networkRangeAction: "allow"}
		pc.form = newPostureCheckForm(&pc.formData)
		pc.state = postureViewForm
		return pc, pc.form.Init()
	case "d":
		if len(pc.filtered) > 0 {
			check := pc.filtered[pc.cursor]
			return pc, DeletePostureCheck(c, check.ID)
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
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  c: create  d: delete  r: refresh", pc.cursor+1, len(pc.filtered))))

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
	if check.Checks.OSVersionCheck != nil {
		osCheck := check.Checks.OSVersionCheck
		if osCheck.Android != nil {
			fields = append(fields, struct{ label, value string }{"Android Min", osCheck.Android.MinVersion})
		}
		if osCheck.Darwin != nil {
			fields = append(fields, struct{ label, value string }{"macOS Min", osCheck.Darwin.MinVersion})
		}
		if osCheck.IOS != nil {
			fields = append(fields, struct{ label, value string }{"iOS Min", osCheck.IOS.MinVersion})
		}
		if osCheck.Linux != nil {
			fields = append(fields, struct{ label, value string }{"Linux Min Kernel", osCheck.Linux.MinKernelVersion})
		}
		if osCheck.Windows != nil {
			fields = append(fields, struct{ label, value string }{"Windows Min Kernel", osCheck.Windows.MinKernelVersion})
		}
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
	if check.Checks.ProcessCheck != nil {
		procs := check.Checks.ProcessCheck
		paths := make([]string, 0, len(procs.Processes))
		for _, p := range procs.Processes {
			if p.LinuxPath != "" {
				paths = append(paths, p.LinuxPath)
			} else if p.MacPath != "" {
				paths = append(paths, p.MacPath)
			} else if p.WindowsPath != "" {
				paths = append(paths, p.WindowsPath)
			}
		}
		fields = append(fields, struct{ label, value string }{"Processes", strings.Join(paths, ", ")})
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  e: edit  d: delete  r: refresh"))

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
