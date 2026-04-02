package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// settingsFocus tracks which panel has keyboard focus within the settings page
type settingsFocus int

const (
	focusSidebar settingsFocus = iota
	focusContentPanel
)

// SettingsPage is the TUI page for editing account-wide settings
type SettingsPage struct {
	accountID  string
	original   models.AccountSettings
	draft      models.AccountSettings
	section    settingsSection
	sidebarIdx int
	contentIdx int
	editingIdx int // -1 means not editing
	editBuf    string
	innerFocus settingsFocus
	focused    bool
	loading    bool
	loaded     bool
	saving     bool
	confirming bool
	err        error
	toast      string
	dirty      bool
}

// NewSettingsPage creates a fresh SettingsPage
func NewSettingsPage() *SettingsPage {
	return &SettingsPage{
		section:    sectionAuthentication,
		sidebarIdx: 0,
		contentIdx: 0,
		editingIdx: -1,
		innerFocus: focusSidebar,
	}
}

func (s *SettingsPage) Title() string        { return "Settings" }
func (s *SettingsPage) CursorPosition() int  { return s.contentIdx }
func (s *SettingsPage) SetFocused(f bool)    { s.focused = f }

// Init fetches settings if not yet loaded
func (s *SettingsPage) Init(c *client.Client) tea.Cmd {
	if s.loaded {
		return nil
	}
	s.loading = true
	s.err = nil
	return FetchSettings(c)
}

// Update handles all messages for the settings page
func (s *SettingsPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	switch m := msg.(type) {

	case SettingsLoadedMsg:
		s.loading = false
		if m.Err != nil {
			s.err = m.Err
			return s, nil
		}
		s.accountID = m.Account.ID
		s.original = draftFromAccount(m.Account)
		s.draft = draftFromAccount(m.Account)
		s.loaded = true
		s.dirty = false
		return s, nil

	case SettingsSavedMsg:
		s.saving = false
		if m.Err != nil {
			s.err = m.Err
			return s, nil
		}
		s.accountID = m.Account.ID
		s.original = draftFromAccount(m.Account)
		s.draft = draftFromAccount(m.Account)
		s.dirty = false
		s.toast = "Settings saved"
		return s, nil

	case tea.KeyPressMsg:
		return s.handleKey(m, c)
	}

	return s, nil
}

// handleKey dispatches key events based on current page state
func (s *SettingsPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	// ── Confirm screen ───────────────────────────────────────────────
	if s.confirming {
		switch key {
		case "y":
			s.confirming = false
			s.saving = true
			return s, SaveSettings(c, s.accountID, s.draft)
		case "n", keyEsc:
			s.confirming = false
		}
		return s, nil
	}

	// ── Sidebar focus ────────────────────────────────────────────────
	if s.innerFocus == focusSidebar {
		return s.handleSidebarKey(key)
	}

	// ── Content focus ────────────────────────────────────────────────
	return s.handleContentKey(key, c)
}

func (s *SettingsPage) handleSidebarKey(key string) (Page, tea.Cmd) {
	switch key {
	case keyUp:
		if s.sidebarIdx > 0 {
			s.sidebarIdx--
		}
	case keyDown:
		if s.sidebarIdx < len(settingsSectionLabels)-1 {
			s.sidebarIdx++
		}
	case keyEnter, "right", "l":
		s.section = settingsSection(s.sidebarIdx)
		s.contentIdx = 0
		s.editingIdx = -1
		s.innerFocus = focusContentPanel
	}
	return s, nil
}

func (s *SettingsPage) handleContentKey(key string, c *client.Client) (Page, tea.Cmd) {
	visible := s.visibleFields(s.section)

	// ── Edit mode ────────────────────────────────────────────────────
	if s.editingIdx >= 0 {
		switch key {
		case keyEnter:
			if s.editingIdx < len(visible) {
				s.setFieldValue(visible[s.editingIdx].ID, s.editBuf)
				s.dirty = s.settingsChanged()
			}
			s.editingIdx = -1
			s.editBuf = ""
		case keyEsc:
			s.editingIdx = -1
			s.editBuf = ""
		case keyBack:
			if len(s.editBuf) > 0 {
				s.editBuf = s.editBuf[:len(s.editBuf)-1]
			}
		default:
			if len(key) == 1 {
				s.editBuf += key
			}
		}
		return s, nil
	}

	// ── Navigation mode ──────────────────────────────────────────────
	switch key {
	case keyUp:
		if s.contentIdx > 0 {
			s.contentIdx--
		} else {
			// Return to sidebar when at the top
			s.innerFocus = focusSidebar
		}

	case keyDown:
		if s.contentIdx < len(visible)-1 {
			s.contentIdx++
		}

	case "left", keyEsc:
		s.innerFocus = focusSidebar
		s.editingIdx = -1
		s.editBuf = ""

	case keyEnter, " ":
		if s.contentIdx >= 0 && s.contentIdx < len(visible) {
			f := visible[s.contentIdx]
			if f.CloudOnly {
				// Cloud-only fields are read-only
				break
			}
			switch f.Kind {
			case fieldToggle:
				s.toggleField(f.ID)
				s.dirty = s.settingsChanged()
				// Clamp cursor after toggle (conditional fields may appear/disappear)
				newVisible := s.visibleFields(s.section)
				if s.contentIdx >= len(newVisible) {
					s.contentIdx = len(newVisible) - 1
				}
				if s.contentIdx < 0 {
					s.contentIdx = 0
				}
			case fieldText, fieldDuration:
				s.editingIdx = s.contentIdx
				s.editBuf = s.getFieldValue(f.ID)
			}
		}

	case "s":
		if s.loaded && s.dirty {
			s.confirming = true
		}

	case keyRefresh:
		s.loading = true
		s.loaded = false
		s.err = nil
		return s, FetchSettings(c)
	}

	return s, nil
}

// View renders the full settings page
func (s *SettingsPage) View(width, height int) string {
	// ── Confirm screen ───────────────────────────────────────────────
	if s.confirming {
		fields := s.changedConfirmFields()
		return RenderConfirm("Save Settings", fields)
	}

	// ── Loading / error states ───────────────────────────────────────
	if s.saving {
		return loadingStyle.Render("  Saving settings...")
	}

	// Compute panel dimensions
	dividerWidth := 1
	contentWidth := width - sidebarWidth - dividerWidth
	if contentWidth < 10 {
		contentWidth = 10
	}

	sidebarStr := s.renderSidebar(height)

	divider := strings.Repeat("│\n", height)
	dividerStr := lipgloss.NewStyle().
		Foreground(colorBorder).
		Width(dividerWidth).
		Height(height).
		Render(divider)

	contentStr := s.renderContent(contentWidth, height)

	// Toast overlay at top of content if present
	if s.toast != "" {
		toastLine := lipgloss.NewStyle().Foreground(colorSuccess).Render("  ✓ " + s.toast)
		contentStr = toastLine + "\n" + contentStr
	}

	result := lipgloss.JoinHorizontal(lipgloss.Top, sidebarStr, dividerStr, contentStr)

	return result
}
