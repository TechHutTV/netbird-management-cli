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

type usersViewState int

const (
	usersViewList usersViewState = iota
	usersViewDetail
	usersViewForm
	usersViewConfirm
	usersViewEdit
	usersViewEditConfirm
)

// UsersPage manages the users list and detail views
type UsersPage struct {
	users      []models.User
	filtered   []models.User
	cursor     int
	loading    bool
	err        error
	state      usersViewState
	form       *huh.Form
	formData   userInviteFormData
	focused    bool
	search     string
	searching  bool
	editForm   *huh.Form
	editData   userEditFormData
	editUserID string
	groupNames map[string]string
}

func NewUsersPage() *UsersPage {
	return &UsersPage{loading: true}
}

func (u *UsersPage) Title() string { return "Users" }
func (u *UsersPage) CursorPosition() int { return u.cursor }
func (u *UsersPage) SetFocused(focused bool) { u.focused = focused }

type usersDataLoadedMsg struct {
	users      []models.User
	groupNames map[string]string
	err        error
}

func fetchUsersData(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		groupResp, err := c.MakeRequest("GET", "/groups", nil)
		if err != nil {
			return usersDataLoadedMsg{err: err}
		}
		defer groupResp.Body.Close()
		var allGroups []models.PolicyGroup
		if err := jsonDecode(groupResp.Body, &allGroups); err != nil {
			return usersDataLoadedMsg{err: err}
		}
		nameMap := make(map[string]string, len(allGroups))
		for _, g := range allGroups {
			nameMap[g.ID] = g.Name
		}

		userResp, err := c.MakeRequest("GET", "/users", nil)
		if err != nil {
			return usersDataLoadedMsg{err: err}
		}
		defer userResp.Body.Close()
		var users []models.User
		if err := jsonDecode(userResp.Body, &users); err != nil {
			return usersDataLoadedMsg{err: err}
		}

		return usersDataLoadedMsg{users: users, groupNames: nameMap}
	}
}

func (u *UsersPage) Init(c *client.Client) tea.Cmd {
	if len(u.users) > 0 {
		return nil
	}
	u.loading = true
	u.err = nil
	return fetchUsersData(c)
}

func (u *UsersPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle edit confirm screen
	if u.state == usersViewEditConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				u.state = usersViewList
				return u, submitUserEdit(c, u.editUserID, u.editData)
			case "n", "esc":
				u.state = usersViewDetail
			}
		}
		return u, nil
	}

	// Delegate to edit form when active
	if u.state == usersViewEdit && u.editForm != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			u.state = usersViewDetail
			u.editForm = nil
			return u, nil
		}
		m, cmd := u.editForm.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			u.editForm = f
		}
		if u.editForm.State == huh.StateCompleted {
			u.state = usersViewEditConfirm
			u.editForm = nil
			return u, nil
		}
		if u.editForm.State == huh.StateAborted {
			u.state = usersViewDetail
			u.editForm = nil
		}
		return u, cmd
	}

	// Handle confirm screen
	if u.state == usersViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				u.state = usersViewList
				return u, submitUserInvite(c, u.formData)
			case "n", "esc":
				u.state = usersViewList
			}
		}
		return u, nil
	}

	// Delegate to form when active
	if u.state == usersViewForm && u.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			u.state = usersViewList
			u.form = nil
			return u, nil
		}
		m, cmd := u.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			u.form = f
		}
		if u.form.State == huh.StateCompleted {
			u.state = usersViewConfirm
			u.form = nil
			return u, nil
		}
		if u.form.State == huh.StateAborted {
			u.state = usersViewList
			u.form = nil
		}
		return u, cmd
	}

	switch msg := msg.(type) {
	case PageRefreshTickMsg:
		// Only refresh if in list view and not editing or in form
		if u.state == usersViewList && u.form == nil && u.editForm == nil {
			return u, fetchUsersData(c)
		}
		return u, nil

	case usersDataLoadedMsg:
		u.loading = false
		if msg.err != nil {
			u.err = msg.err
			return u, nil
		}
		u.groupNames = msg.groupNames
		regular := make([]models.User, 0)
		for _, user := range msg.users {
			if !user.IsServiceUser {
				regular = append(regular, user)
			}
		}
		u.users = regular
		u.applyFilter()
		return u, nil

	case UsersLoadedMsg:
		u.loading = false
		if msg.Err != nil {
			u.err = msg.Err
			return u, nil
		}
		regular := make([]models.User, 0)
		for _, user := range msg.Users {
			if !user.IsServiceUser {
				regular = append(regular, user)
			}
		}
		u.users = regular
		u.applyFilter()
		return u, nil

	case UserUpdatedMsg:
		if msg.Err != nil {
			u.err = msg.Err
			return u, nil
		}
		u.loading = true
		return u, FetchUsers(c)

	case formCompleteMsg:
		u.loading = true
		return u, FetchUsers(c)

	case ToastMsg:
		u.loading = true
		return u, FetchUsers(c)

	case APIErrorMsg:
		u.err = msg.Err
		return u, nil

	case tea.KeyPressMsg:
		return u.handleKey(msg, c)
	}

	return u, nil
}

func (u *UsersPage) View(width, height int) string {
	if u.loading {
		return loadingStyle.Render("  Loading users...")
	}
	if u.err != nil {
		return errorStyle.Render("  Error: " + u.err.Error())
	}

	if u.state == usersViewEditConfirm {
		return RenderConfirm("Confirm: Edit User", []ConfirmField{
			{Label: "Role", Value: u.editData.role},
			{Label: "Auto Groups", Value: u.editData.autoGroups},
		})
	}
	if u.state == usersViewEdit && u.editForm != nil {
		return pageTitleStyle.Render("Edit User") + "\n\n" + u.editForm.View()
	}
	if u.state == usersViewConfirm {
		groupDisplay := make([]string, 0, len(u.formData.selectedGroups))
		for _, id := range u.formData.selectedGroups {
			if name, ok := u.groupNames[id]; ok {
				groupDisplay = append(groupDisplay, name)
			} else {
				groupDisplay = append(groupDisplay, id)
			}
		}
		return RenderConfirm("Confirm: Invite User", []ConfirmField{
			{Label: "Name", Value: u.formData.name},
			{Label: "Email", Value: u.formData.email},
			{Label: "Role", Value: u.formData.role},
			{Label: "Auto Groups", Value: strings.Join(groupDisplay, ", ")},
			{Label: "Service User", Value: fmt.Sprintf("%v", u.formData.isService)},
		})
	}
	if u.state == usersViewForm && u.form != nil {
		return pageTitleStyle.Render("Invite User") + "\n\n" + u.form.View()
	}
	if u.state == usersViewDetail {
		return u.viewDetail(width)
	}
	return u.viewList(width, height)
}

func (u *UsersPage) applyFilter() {
	if u.search == "" {
		u.filtered = u.users
	} else {
		filtered := make([]models.User, 0)
		for _, user := range u.users {
			if matchesQuery(u.search, user.Name, user.Email, user.Role) {
				filtered = append(filtered, user)
			}
		}
		u.filtered = filtered
	}
	u.cursor = clampCursor(u.cursor, len(u.filtered))
}

func (u *UsersPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if u.state == usersViewDetail {
		switch key {
		case "esc", "backspace", "q":
			u.state = usersViewList
		case "e":
			if len(u.filtered) > 0 {
				user := u.filtered[u.cursor]
				u.editData = userEditFormData{
					role:       user.Role,
					autoGroups: strings.Join(user.AutoGroups, ", "),
				}
				u.editForm = newUserEditForm(&u.editData)
				u.editUserID = user.ID
				u.state = usersViewEdit
				return u, u.editForm.Init()
			}
		}
		return u, nil
	}

	// Search input handling (list view only)
	if u.searching {
		newSearch, still, changed := handleSearchKey(key, u.search)
		u.search = newSearch
		u.searching = still
		if changed || !still {
			u.applyFilter()
		}
		return u, nil
	}

	switch key {
	case "up", "k":
		if u.cursor > 0 {
			u.cursor--
		}
	case "down", "j":
		if u.cursor < len(u.filtered)-1 {
			u.cursor++
		}
	case "enter":
		if len(u.filtered) > 0 {
			u.state = usersViewDetail
		}
	case "r":
		u.loading = true
		return u, FetchUsers(c)
	case "c":
		u.formData = userInviteFormData{}
		u.form = newUserInviteForm(&u.formData, u.groupNames)
		u.state = usersViewForm
		return u, u.form.Init()
	case "b":
		if len(u.filtered) > 0 {
			user := u.filtered[u.cursor]
			return u, ToggleUserBlock(c, user, !user.IsBlocked)
		}
	case "/":
		u.searching = true
		u.search = ""
	case "esc":
		if u.search != "" {
			u.search = ""
			u.cursor = 0
			u.applyFilter()
		}
	}

	return u, nil
}

func (u *UsersPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Users (%d)", len(u.filtered))) + "\n")

	if u.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(u.search) + "\u2588\n")
	} else if u.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", u.search)) + "\n")
	}

	if len(u.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No users found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}

	offset := 0
	if u.cursor >= maxRows {
		offset = u.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(u.filtered) {
		end = len(u.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		user := u.filtered[i]

		blocked := "No"
		if user.IsBlocked {
			blocked = "Yes"
		}

		rows = append(rows, []string{
			user.Email,
			user.Name,
			user.Role,
			user.Status,
			blocked,
			formatLastSeen(user.LastLogin),
		})
	}

	tw := width
	if tw > 100 {
		tw = 100
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("EMAIL", "NAME", "ROLE", "STATUS", "BLOCKED", "LAST SEEN").
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

			if dataIdx == u.cursor && u.focused {
				base = tableSelectedStyle
			}

			// Blocked column coloring
			if col == 4 && row >= 0 && row < len(rows) {
				if rows[row][4] == "Yes" {
					return base.Foreground(colorDanger)
				}
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  c: invite  b: block/unblock  r: refresh", u.cursor+1, len(u.filtered))))

	return b.String()
}

func (u *UsersPage) viewDetail(width int) string {
	if u.cursor >= len(u.filtered) {
		return "No user selected"
	}

	user := u.filtered[u.cursor]
	var b strings.Builder

	b.WriteString(detailTitleStyle.Render("  "+user.Name) + "\n\n")

	groupsStr := "None"
	if len(user.AutoGroups) > 0 {
		groupsStr = strings.Join(user.AutoGroups, ", ")
	}

	fields := []struct{ label, value string }{
		{"ID", user.ID},
		{"Email", user.Email},
		{"Name", user.Name},
		{"Role", user.Role},
		{"Status", user.Status},
		{"Service User", fmt.Sprintf("%v", user.IsServiceUser)},
		{"Blocked", fmt.Sprintf("%v", user.IsBlocked)},
		{"Last Login", user.LastLogin},
		{"Auto Groups", groupsStr},
	}

	for _, f := range fields {
		label := detailLabelStyle.Render(f.label)
		value := detailValueStyle.Render(f.value)
		b.WriteString(fmt.Sprintf("%s  %s\n", label, value))
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  e: edit  r: refresh"))

	return b.String()
}
