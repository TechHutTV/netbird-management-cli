package tui

import (
	"bytes"
	"encoding/json"
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

type svcUsersViewState int

const (
	svcUsersViewList svcUsersViewState = iota
	svcUsersViewDetail
	svcUsersViewForm
	svcUsersViewConfirm
)

// ServiceUsersPage manages service users list and detail views
type ServiceUsersPage struct {
	users     []models.User
	filtered  []models.User
	cursor    int
	loading   bool
	err       error
	state     svcUsersViewState
	form      *huh.Form
	formData  svcUserFormData
	focused   bool
	search    string
	searching bool
}

type svcUserFormData struct {
	name       string
	role       string
	autoGroups string
}

func NewServiceUsersPage() *ServiceUsersPage {
	return &ServiceUsersPage{loading: true}
}

func (s *ServiceUsersPage) Title() string { return "Service Users" }
func (s *ServiceUsersPage) CursorPosition() int { return s.cursor }
func (s *ServiceUsersPage) SetFocused(focused bool) { s.focused = focused }

func (s *ServiceUsersPage) Init(c *client.Client) tea.Cmd {
	s.loading = true
	s.err = nil
	return FetchUsers(c)
}

func (s *ServiceUsersPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle confirm screen
	if s.state == svcUsersViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				s.state = svcUsersViewList
				return s, submitServiceUserCreate(c, s.formData)
			case "n", "esc":
				s.state = svcUsersViewList
			}
		}
		return s, nil
	}

	// Form delegation
	if s.state == svcUsersViewForm && s.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			s.state = svcUsersViewList
			s.form = nil
			return s, nil
		}
		m, cmd := s.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			s.form = f
		}
		if s.form.State == huh.StateCompleted {
			s.state = svcUsersViewConfirm
			s.form = nil
			return s, nil
		}
		if s.form.State == huh.StateAborted {
			s.state = svcUsersViewList
			s.form = nil
		}
		return s, cmd
	}

	switch msg := msg.(type) {
	case UsersLoadedMsg:
		s.loading = false
		if msg.Err != nil {
			s.err = msg.Err
			return s, nil
		}
		// Filter to service users only
		svcUsers := make([]models.User, 0)
		for _, user := range msg.Users {
			if user.IsServiceUser {
				svcUsers = append(svcUsers, user)
			}
		}
		s.users = svcUsers
		s.applyFilter()
		return s, nil

	case formCompleteMsg:
		return s, s.Init(c)

	case ToastMsg:
		return s, s.Init(c)

	case APIErrorMsg:
		s.err = msg.Err
		return s, nil

	case tea.KeyPressMsg:
		return s.handleKey(msg, c)
	}

	return s, nil
}

func (s *ServiceUsersPage) View(width, height int) string {
	if s.loading {
		return loadingStyle.Render("  Loading service users...")
	}
	if s.err != nil {
		return errorStyle.Render("  Error: " + s.err.Error())
	}

	if s.state == svcUsersViewConfirm {
		return RenderConfirm("Confirm: Create Service User", []ConfirmField{
			{Label: "Name", Value: s.formData.name},
			{Label: "Role", Value: s.formData.role},
			{Label: "Auto Groups", Value: s.formData.autoGroups},
		})
	}
	if s.state == svcUsersViewForm && s.form != nil {
		return pageTitleStyle.Render("Create Service User") + "\n\n" + s.form.View()
	}
	if s.state == svcUsersViewDetail {
		return s.viewDetail(width)
	}
	return s.viewList(width, height)
}

func (s *ServiceUsersPage) applyFilter() {
	if s.search == "" {
		s.filtered = s.users
	} else {
		filtered := make([]models.User, 0)
		for _, user := range s.users {
			if matchesQuery(s.search, user.Name) {
				filtered = append(filtered, user)
			}
		}
		s.filtered = filtered
	}
	s.cursor = clampCursor(s.cursor, len(s.filtered))
}

func (s *ServiceUsersPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if s.state == svcUsersViewDetail {
		if key == "esc" || key == "backspace" || key == "q" {
			s.state = svcUsersViewList
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
			s.state = svcUsersViewDetail
		}
	case "r":
		return s, s.Init(c)
	case "c":
		s.formData = svcUserFormData{role: "user"}
		s.form = newServiceUserCreateForm(&s.formData)
		s.state = svcUsersViewForm
		return s, s.form.Init()
	case "b":
		if len(s.filtered) > 0 {
			user := s.filtered[s.cursor]
			return s, ToggleUserBlock(c, user, !user.IsBlocked)
		}
	case "d":
		if len(s.filtered) > 0 {
			user := s.filtered[s.cursor]
			return s, deleteUser(c, user.ID)
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

func (s *ServiceUsersPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Service Users (%d)", len(s.filtered))) + "\n")

	if s.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(s.search) + "\u2588\n")
	} else if s.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", s.search)) + "\n")
	}

	if len(s.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No service users found."))
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
		user := s.filtered[i]

		blocked := "No"
		if user.IsBlocked {
			blocked = "Yes"
		}

		rows = append(rows, []string{
			user.Name,
			user.Role,
			user.Status,
			blocked,
		})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "ROLE", "STATUS", "BLOCKED").
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

			if col == 3 && row >= 0 && row < len(rows) {
				if rows[row][3] == "Yes" {
					return base.Foreground(colorDanger)
				}
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  c: create  b: block/unblock  d: delete  r: refresh", s.cursor+1, len(s.filtered))))

	return b.String()
}

func (s *ServiceUsersPage) viewDetail(width int) string {
	if s.cursor >= len(s.filtered) {
		return "No user selected"
	}

	user := s.filtered[s.cursor]
	var b strings.Builder

	b.WriteString(detailTitleStyle.Render("  "+user.Name) + "\n\n")

	groupsStr := "None"
	if len(user.AutoGroups) > 0 {
		groupsStr = strings.Join(user.AutoGroups, ", ")
	}

	fields := []struct{ label, value string }{
		{"ID", user.ID},
		{"Name", user.Name},
		{"Role", user.Role},
		{"Status", user.Status},
		{"Blocked", fmt.Sprintf("%v", user.IsBlocked)},
		{"Auto Groups", groupsStr},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  b: block/unblock  d: delete"))

	return b.String()
}

func newServiceUserCreateForm(data *svcUserFormData) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("e.g. ci-bot").
				Value(&data.name),
			huh.NewSelect[string]().
				Title("Role").
				Options(
					huh.NewOption("User", "user"),
					huh.NewOption("Admin", "admin"),
				).
				Value(&data.role),
			huh.NewInput().
				Title("Auto Groups").
				Description("Comma-separated group IDs (optional)").
				Placeholder("group-id").
				Value(&data.autoGroups),
		),
	)
}

func submitServiceUserCreate(c *client.Client, data svcUserFormData) tea.Cmd {
	return func() tea.Msg {
		var autoGroups []string
		if data.autoGroups != "" {
			autoGroups = splitTrim(data.autoGroups)
		}

		req := models.UserCreateRequest{
			Name:          data.name,
			Role:          data.role,
			AutoGroups:    autoGroups,
			IsServiceUser: true,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return APIErrorMsg{Err: fmt.Errorf("marshal request: %w", err), Context: "create service user"}
		}
		resp, err := c.MakeRequest("POST", "/users", bytes.NewReader(body))
		if err != nil {
			return APIErrorMsg{Err: err, Context: "create service user"}
		}
		defer resp.Body.Close()
		return formCompleteMsg{message: fmt.Sprintf("Service user '%s' created", data.name)}
	}
}

func deleteUser(c *client.Client, userID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("DELETE", "/users/"+url.PathEscape(userID), nil)
		if err != nil {
			return APIErrorMsg{Err: err, Context: "delete user"}
		}
		defer resp.Body.Close()
		return ToastMsg{Message: "User deleted"}
	}
}
