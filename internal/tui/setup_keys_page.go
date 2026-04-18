package tui

import (
	"fmt"
	"net/url"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

type setupKeysViewState int

const (
	setupKeysViewList setupKeysViewState = iota
	setupKeysViewDetail
	setupKeysViewForm
	setupKeysViewConfirm
)

// SetupKeysPage manages setup keys list and detail views
type SetupKeysPage struct {
	keys       []models.SetupKey
	filtered   []models.SetupKey
	cursor     int
	loading    bool
	err        error
	state      setupKeysViewState
	form       *huh.Form
	formData   setupKeyFormData
	focused    bool
	search     string
	searching  bool
	groupNames map[string]string
}

func NewSetupKeysPage() *SetupKeysPage {
	return &SetupKeysPage{loading: true}
}

type setupKeysDataLoadedMsg struct {
	keys       []models.SetupKey
	groupNames map[string]string
	err        error
}

func fetchSetupKeysData(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		// Fetch groups for name resolution
		groupResp, err := c.MakeRequest("GET", "/groups", nil)
		if err != nil {
			return setupKeysDataLoadedMsg{err: err}
		}
		defer groupResp.Body.Close()
		var allGroups []models.PolicyGroup
		if err := jsonDecode(groupResp.Body, &allGroups); err != nil {
			return setupKeysDataLoadedMsg{err: err}
		}
		nameMap := make(map[string]string, len(allGroups))
		for _, g := range allGroups {
			nameMap[g.ID] = g.Name
		}

		// Fetch setup keys
		keysResp, err := c.MakeRequest("GET", "/setup-keys", nil)
		if err != nil {
			return setupKeysDataLoadedMsg{err: err}
		}
		defer keysResp.Body.Close()
		var keys []models.SetupKey
		if err := jsonDecode(keysResp.Body, &keys); err != nil {
			return setupKeysDataLoadedMsg{err: err}
		}

		return setupKeysDataLoadedMsg{keys: keys, groupNames: nameMap}
	}
}

func (s *SetupKeysPage) Title() string { return "Setup Keys" }
func (s *SetupKeysPage) CursorPosition() int {
	if s.state != setupKeysViewList {
		return 1
	}
	return s.cursor
}
func (s *SetupKeysPage) SetFocused(focused bool) { s.focused = focused }

func (s *SetupKeysPage) Init(c *client.Client) tea.Cmd {
	if len(s.keys) > 0 {
		return nil
	}
	s.loading = true
	s.err = nil
	return fetchSetupKeysData(c)
}

func (s *SetupKeysPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle confirm screen
	if s.state == setupKeysViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				s.state = setupKeysViewList
				return s, submitSetupKeyCreate(c, s.formData)
			case "n", "esc":
				s.state = setupKeysViewList
			}
		}
		return s, nil
	}

	// Delegate to form when active
	if s.state == setupKeysViewForm && s.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			s.state = setupKeysViewList
			s.form = nil
			return s, nil
		}
		m, cmd := s.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			s.form = f
		}
		if s.form.State == huh.StateCompleted {
			s.state = setupKeysViewConfirm
			s.form = nil
			return s, nil
		}
		if s.form.State == huh.StateAborted {
			s.state = setupKeysViewList
			s.form = nil
		}
		return s, cmd
	}

	switch msg := msg.(type) {
	case PageRefreshTickMsg:
		// Only refresh if in list view and not in form
		if s.state == setupKeysViewList && s.form == nil {
			s.loading = true
			return s, fetchSetupKeysData(c)
		}
		return s, nil

	case setupKeysDataLoadedMsg:
		s.loading = false
		if msg.err != nil {
			s.err = msg.err
			return s, nil
		}
		s.keys = msg.keys
		s.groupNames = msg.groupNames
		s.applyFilter()
		return s, nil

	case formCompleteMsg:
		s.loading = true
		return s, fetchSetupKeysData(c)

	case ToastMsg:
		s.loading = true
		return s, fetchSetupKeysData(c)

	case APIErrorMsg:
		s.err = msg.Err
		return s, nil

	case tea.KeyPressMsg:
		return s.handleKey(msg, c)
	}

	return s, nil
}

func (s *SetupKeysPage) View(width, height int) string {
	if s.loading {
		return loadingStyle.Render("  Loading setup keys...")
	}
	if s.err != nil {
		return errorStyle.Render("  Error: " + s.err.Error())
	}

	if s.state == setupKeysViewConfirm {
		return RenderConfirm("Confirm: Create Setup Key", []ConfirmField{
			{Label: "Name", Value: s.formData.name},
			{Label: "Type", Value: s.formData.keyType},
			{Label: "Expires In", Value: s.formData.expiresIn},
			{Label: "Usage Limit", Value: s.formData.usageLimit},
			{Label: "Auto Groups", Value: resolveGroupNames(s.formData.selectedGroups, s.groupNames)},
			{Label: "Ephemeral", Value: fmt.Sprintf("%v", s.formData.ephemeral)},
			{Label: "Extra DNS Labels", Value: fmt.Sprintf("%v", s.formData.allowExtraDNSLabels)},
		})
	}
	if s.state == setupKeysViewForm && s.form != nil {
		return pageTitleStyle.Render("Create Setup Key") + "\n\n" + s.form.View()
	}
	if s.state == setupKeysViewDetail {
		return s.viewDetail(width)
	}
	return s.viewList(width, height)
}

func (s *SetupKeysPage) applyFilter() {
	if s.search == "" {
		s.filtered = s.keys
	} else {
		filtered := make([]models.SetupKey, 0)
		for _, sk := range s.keys {
			state := "Expired"
			if sk.Valid && !sk.Revoked {
				state = "Valid"
			} else if sk.Revoked {
				state = "Revoked"
			}
			if matchesQuery(s.search, sk.Name, state) {
				filtered = append(filtered, sk)
			}
		}
		s.filtered = filtered
	}
	s.cursor = clampCursor(s.cursor, len(s.filtered))
}

func (s *SetupKeysPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if s.state == setupKeysViewDetail {
		if key == "esc" || key == "backspace" || key == "q" {
			s.state = setupKeysViewList
		}
		return s, nil
	}

	// Search input handling (list view only)
	if s.searching {
		newSearch, still, changed := handleSearchKey(key, s.search)
		s.search = newSearch
		s.searching = still
		if changed || !still {
			s.applyFilter()
		}
		return s, nil
	}

	switch key {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(s.filtered)-1 {
			s.cursor++
		}
	case "enter":
		if len(s.filtered) > 0 {
			s.state = setupKeysViewDetail
		}
	case "r":
		s.loading = true
		return s, FetchSetupKeys(c)
	case "c":
		s.formData = setupKeyFormData{}
		s.form = newSetupKeyCreateForm(&s.formData, s.groupNames)
		s.state = setupKeysViewForm
		return s, s.form.Init()
	case "d":
		if len(s.filtered) > 0 {
			sk := s.filtered[s.cursor]
			return s, deleteSetupKey(c, sk.ID)
		}
	case "v":
		if len(s.filtered) > 0 {
			sk := s.filtered[s.cursor]
			return s, RevokeSetupKey(c, sk, !sk.Revoked)
		}
	case "/":
		s.searching = true
		s.search = ""
	case "esc":
		if s.search != "" {
			s.search = ""
			s.cursor = 0
			s.applyFilter()
		}
	}

	return s, nil
}

func (s *SetupKeysPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Setup Keys (%d)", len(s.filtered))) + "\n")

	if s.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(s.search) + "\u2588\n")
	} else if s.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", s.search)) + "\n")
	}

	if len(s.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No setup keys found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if s.cursor >= maxRows {
		offset = s.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(s.filtered) {
		end = len(s.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		sk := s.filtered[i]

		state := "Expired"
		if sk.Valid && !sk.Revoked {
			state = "Valid"
		} else if sk.Revoked {
			state = "Revoked"
		}

		usage := fmt.Sprintf("%d/%d", sk.UsedTimes, sk.UsageLimit)
		if sk.UsageLimit == 0 {
			usage = fmt.Sprintf("%d/inf", sk.UsedTimes)
		}

		rows = append(rows, []string{
			sk.Name,
			sk.Type,
			state,
			usage,
			sk.Expires,
		})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "TYPE", "STATE", "USED/LIMIT", "EXPIRES").
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

			if dataIdx == s.cursor && s.focused {
				base = tableSelectedStyle
			}

			// State column coloring
			if col == 2 && row >= 0 && row < len(rows) {
				switch rows[row][2] {
				case "Valid":
					return base.Foreground(colorSuccess)
				case "Expired":
					return base.Foreground(colorDanger)
				case "Revoked":
					return base.Foreground(colorTextDim)
				}
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  c: create  v: revoke  d: delete  r: refresh", s.cursor+1, len(s.filtered))))

	return b.String()
}

func (s *SetupKeysPage) viewDetail(width int) string {
	if s.cursor >= len(s.filtered) {
		return "No key selected"
	}

	sk := s.filtered[s.cursor]
	var b strings.Builder

	state := "Expired"
	if sk.Valid && !sk.Revoked {
		state = "Valid"
	} else if sk.Revoked {
		state = "Revoked"
	}

	// Title with state badge
	stateBadge := expiredStyle.Render(" EXPIRED ")
	if state == "Valid" {
		stateBadge = validStyle.Render(" VALID ")
	} else if state == "Revoked" {
		stateBadge = revokedStyle.Render(" REVOKED ")
	}

	b.WriteString(detailTitleStyle.Render(fmt.Sprintf("  %s  %s", sk.Name, stateBadge)))
	b.WriteString("\n\n")

	groupsStr := resolveGroupNames(sk.AutoGroups, s.groupNames)
	if groupsStr == "" {
		groupsStr = "(none)"
	}

	usageLimit := "Unlimited"
	if sk.UsageLimit > 0 {
		usageLimit = fmt.Sprintf("%d", sk.UsageLimit)
	}

	fields := []struct{ label, value string }{
		{"ID", sk.ID},
		{"Name", sk.Name},
		{"Type", sk.Type},
		{"State", state},
		{"Used Times", fmt.Sprintf("%d", sk.UsedTimes)},
		{"Usage Limit", usageLimit},
		{"Expires", sk.Expires},
		{"Ephemeral", fmt.Sprintf("%v", sk.Ephemeral)},
		{"Extra DNS Labels", fmt.Sprintf("%v", sk.AllowExtraDNSLabels)},
		{"Auto Groups", groupsStr},
		{"Updated At", sk.UpdatedAt},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  d: delete  r: refresh"))

	return b.String()
}

func deleteSetupKey(c *client.Client, keyID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/setup-keys/"+url.PathEscape(keyID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete setup key"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "Setup key deleted"}
	}
}
