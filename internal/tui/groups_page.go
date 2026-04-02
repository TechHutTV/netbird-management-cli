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

type groupsViewState int

const (
	groupsViewList groupsViewState = iota
	groupsViewDetail
	groupsViewForm
	groupsViewConfirm
)

type GroupsPage struct {
	groups    []models.PolicyGroup
	filtered  []models.PolicyGroup
	cursor    int
	loading   bool
	err       error
	state     groupsViewState
	detail    *models.GroupDetail
	form      *huh.Form
	formData  groupFormData
	focused   bool
	search    string
	searching bool
}

func NewGroupsPage() *GroupsPage {
	return &GroupsPage{loading: true}
}

func (g *GroupsPage) Title() string { return "Groups" }
func (g *GroupsPage) CursorPosition() int { return g.cursor }
func (g *GroupsPage) SetFocused(focused bool) { g.focused = focused }

func (g *GroupsPage) Init(c *client.Client) tea.Cmd {
	g.loading = true
	g.err = nil
	return FetchGroups(c)
}

func (g *GroupsPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Handle confirm screen
	if g.state == groupsViewConfirm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "y":
				g.state = groupsViewList
				return g, submitGroupCreate(c, g.formData)
			case "n", "esc":
				g.state = groupsViewList
			}
		}
		return g, nil
	}

	// Delegate to form when active
	if g.state == groupsViewForm && g.form != nil {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			g.state = groupsViewList
			g.form = nil
			return g, nil
		}
		m, cmd := g.form.Update(msg)
		if f, ok := m.(*huh.Form); ok {
			g.form = f
		}
		if g.form.State == huh.StateCompleted {
			g.state = groupsViewConfirm
			g.form = nil
			return g, nil
		}
		if g.form.State == huh.StateAborted {
			g.state = groupsViewList
			g.form = nil
		}
		return g, cmd
	}

	switch msg := msg.(type) {
	case GroupsLoadedMsg:
		g.loading = false
		if msg.Err != nil {
			g.err = msg.Err
			return g, nil
		}
		g.groups = msg.Groups
		g.applyFilter()
		return g, nil

	case groupDetailLoadedMsg:
		if msg.err != nil {
			g.err = msg.err
			return g, nil
		}
		g.detail = msg.group
		g.state = groupsViewDetail
		return g, nil

	case formCompleteMsg:
		return g, g.Init(c)

	case ToastMsg:
		return g, g.Init(c)

	case APIErrorMsg:
		g.err = msg.Err
		return g, nil

	case tea.KeyPressMsg:
		return g.handleKey(msg, c)
	}

	return g, nil
}

func (g *GroupsPage) View(width, height int) string {
	if g.loading {
		return loadingStyle.Render("  Loading groups...")
	}
	if g.err != nil {
		return errorStyle.Render("  Error: " + g.err.Error())
	}

	if g.state == groupsViewConfirm {
		return RenderConfirm("Confirm: Create Group", []ConfirmField{
			{Label: "Name", Value: g.formData.name},
		})
	}
	if g.state == groupsViewForm && g.form != nil {
		return pageTitleStyle.Render("Create Group") + "\n\n" + g.form.View()
	}
	if g.state == groupsViewDetail && g.detail != nil {
		return g.viewDetail(width)
	}
	return g.viewList(width, height)
}

func (g *GroupsPage) applyFilter() {
	if g.search == "" {
		g.filtered = g.groups
	} else {
		filtered := make([]models.PolicyGroup, 0)
		for _, group := range g.groups {
			if matchesQuery(g.search, group.Name) {
				filtered = append(filtered, group)
			}
		}
		g.filtered = filtered
	}
	g.cursor = clampCursor(g.cursor, len(g.filtered))
}

func (g *GroupsPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if g.state == groupsViewDetail {
		if key == "esc" || key == "backspace" || key == "q" {
			g.state = groupsViewList
			g.detail = nil
		}
		return g, nil
	}

	// Search input handling (list view only)
	if g.searching {
		newSearch, still, changed := handleSearchKey(key, g.search)
		g.search = newSearch
		g.searching = still
		if changed || !still {
			g.applyFilter()
		}
		return g, nil
	}

	switch key {
	case "up", "k":
		if g.cursor > 0 {
			g.cursor--
		}
	case "down", "j":
		if g.cursor < len(g.filtered)-1 {
			g.cursor++
		}
	case "enter":
		if len(g.filtered) > 0 {
			group := g.filtered[g.cursor]
			return g, fetchGroupDetail(c, group.ID)
		}
	case "r":
		return g, g.Init(c)
	case "c":
		g.formData = groupFormData{}
		g.form = newGroupCreateForm(&g.formData)
		g.state = groupsViewForm
		return g, g.form.Init()
	case "d":
		if len(g.filtered) > 0 {
			group := g.filtered[g.cursor]
			return g, DeleteGroup(c, group.ID)
		}
	case "/":
		g.searching = true
		g.search = ""
	case "esc":
		if g.search != "" {
			g.search = ""
			g.cursor = 0
			g.applyFilter()
		}
	}

	return g, nil
}

func (g *GroupsPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Groups (%d)", len(g.filtered))) + "\n")

	if g.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(g.search) + "\u2588\n")
	} else if g.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", g.search)) + "\n")
	}

	if len(g.filtered) == 0 {
		b.WriteString(dimHintStyle.Render("  No groups found."))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}
	offset := 0
	if g.cursor >= maxRows {
		offset = g.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(g.filtered) {
		end = len(g.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		group := g.filtered[i]
		rows = append(rows, []string{
			group.Name,
			fmt.Sprintf("%d", group.PeersCount),
			fmt.Sprintf("%d", group.ResourcesCount),
		})
	}

	tw := width
	if tw > 80 {
		tw = 80
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("NAME", "PEERS", "RESOURCES").
		Rows(rows...).
		Width(tw).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return tableHeaderStyle
			}
			if row+offset == g.cursor && g.focused {
				return tableSelectedStyle
			}
			if row%2 == 0 {
				return tableDimCellStyle
			}
			return tableCellStyle
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  c: create", g.cursor+1, len(g.filtered))))
	return b.String()
}

func (g *GroupsPage) viewDetail(width int) string {
	d := g.detail
	var b strings.Builder

	b.WriteString(detailTitleStyle.Render("  "+d.Name) + "\n\n")

	fields := []struct{ label, value string }{
		{"ID", d.ID},
		{"Name", d.Name},
		{"Peers", fmt.Sprintf("%d", d.PeersCount)},
		{"Resources", fmt.Sprintf("%d", d.ResourcesCount)},
		{"Issued By", d.Issued},
	}

	for _, f := range fields {
		b.WriteString(fmt.Sprintf("%s  %s\n",
			detailLabelStyle.Render(f.label), detailValueStyle.Render(f.value)))
	}

	if len(d.Peers) > 0 {
		b.WriteString("\n" + sectionHeaderStyle.Render("  Members") + "\n")
		for _, peer := range d.Peers {
			status := offlineStyle.Render("offline")
			if peer.Connected {
				status = onlineStyle.Render("online")
			}
			b.WriteString(fmt.Sprintf("    %-20s %s  %s\n", peer.Name, peer.IP, status))
		}
	}

	b.WriteString("\n" + dimHintStyle.Render("  esc: back  d: delete"))
	return b.String()
}

type groupDetailLoadedMsg struct {
	group *models.GroupDetail
	err   error
}

func fetchGroupDetail(c *client.Client, groupID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/groups/"+url.PathEscape(groupID), nil)
		if err != nil {
			return groupDetailLoadedMsg{err: err}
		}
		defer resp.Body.Close()

		var group models.GroupDetail
		if err := jsonDecode(resp.Body, &group); err != nil {
			return groupDetailLoadedMsg{err: err}
		}
		return groupDetailLoadedMsg{group: &group}
	}
}
