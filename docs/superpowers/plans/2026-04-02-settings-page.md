# Settings Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the read-only Account page with an interactive Settings page featuring a vertical sidebar for section navigation and inline editing with save/discard, mirroring the real NetBird dashboard.

**Architecture:** The settings page uses a two-panel layout (sidebar + content). State is managed via `original` and `draft` copies of `AccountSettings` for immutable change tracking. Each section is a pure render function that takes the draft settings and returns styled output. Navigation uses a focus model: sidebar → content → inline edit.

**Tech Stack:** Go, bubbletea v2, lipgloss v2, NetBird REST API (`PUT /accounts/{id}`)

---

### Task 1: Update AccountSettings Model

**Files:**
- Modify: `internal/models/models.go:555-567` (AccountSettings struct)

- [ ] **Step 1: Add missing expiration-enabled fields to AccountSettings**

The API returns `peer_login_expiration_enabled` and `peer_inactivity_expiration_enabled` booleans, but the model doesn't capture them. Add them:

```go
// AccountSettings contains account-wide configuration
type AccountSettings struct {
	PeerLoginExpirationEnabled      bool     `json:"peer_login_expiration_enabled"`
	PeerLoginExpiration             int      `json:"peer_login_expiration"`      // Seconds
	PeerInactivityExpirationEnabled bool     `json:"peer_inactivity_expiration_enabled"`
	PeerInactivityExpiration        int      `json:"peer_inactivity_expiration"` // Seconds
	DNSDomain                       string   `json:"dns_domain"`
	NetworkRange                    string   `json:"network_range"`
	JWTGroupsEnabled                bool     `json:"jwt_groups_enabled"`
	JWTGroupsClaim                  string   `json:"jwt_groups_claim"`
	JWTAllowGroups                  []string `json:"jwt_allow_groups"`
	GroupsPropagationEnabled        bool     `json:"groups_propagation_enabled"`
	RegularUsersViewBlocked         bool     `json:"regular_users_view_blocked"`
	PeerApprovalEnabled             bool     `json:"peer_approval_enabled,omitempty"` // Cloud-only
	TrafficLogging                  bool     `json:"traffic_logging,omitempty"`       // Cloud-only
}
```

- [ ] **Step 2: Verify the build compiles**

Run: `go build ./...`
Expected: PASS (the new fields are additive and zero-valued by default)

- [ ] **Step 3: Commit**

```bash
git add internal/models/models.go
git commit -m "feat(models): add peer expiration enabled flags to AccountSettings"
```

---

### Task 2: Add Settings Messages and API Commands

**Files:**
- Modify: `internal/tui/messages.go` (add SettingsLoadedMsg, SettingsSavedMsg)
- Modify: `internal/tui/api.go` (add FetchSettings, SaveSettings)

- [ ] **Step 1: Add message types to messages.go**

Add after the `AccountsLoadedMsg` block (line ~83):

```go
// SettingsLoadedMsg carries the result of fetching account settings
type SettingsLoadedMsg struct {
	Account models.Account
	Err     error
}

// SettingsSavedMsg carries the result of saving account settings
type SettingsSavedMsg struct {
	Account models.Account
	Err     error
}
```

- [ ] **Step 2: Add FetchSettings and SaveSettings to api.go**

Add after the `FetchAccounts` function (line ~211):

```go
// FetchSettings returns a tea.Cmd that loads the first account (settings)
func FetchSettings(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := c.MakeRequest("GET", "/accounts", nil)
		if err != nil {
			return SettingsLoadedMsg{Err: err}
		}
		defer resp.Body.Close()

		var accounts []models.Account
		if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
			return SettingsLoadedMsg{Err: fmt.Errorf("decode accounts: %w", err)}
		}
		if len(accounts) == 0 {
			return SettingsLoadedMsg{Err: fmt.Errorf("no accounts found")}
		}
		return SettingsLoadedMsg{Account: accounts[0]}
	}
}

// SaveSettings returns a tea.Cmd that PUTs updated account settings
func SaveSettings(c *client.Client, accountID string, settings models.AccountSettings) tea.Cmd {
	return func() tea.Msg {
		req := models.AccountUpdateRequest{
			Settings: settings,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return SettingsSavedMsg{Err: fmt.Errorf("marshal settings: %w", err)}
		}
		resp, err := c.MakeRequest("PUT", "/accounts/"+url.PathEscape(accountID), bytes.NewReader(body))
		if err != nil {
			return SettingsSavedMsg{Err: err}
		}
		defer resp.Body.Close()

		var account models.Account
		if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
			return SettingsSavedMsg{Err: fmt.Errorf("decode response: %w", err)}
		}
		return SettingsSavedMsg{Account: account}
	}
}
```

- [ ] **Step 3: Verify the build compiles**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/tui/messages.go internal/tui/api.go
git commit -m "feat(tui): add settings fetch/save messages and API commands"
```

---

### Task 3: Create Settings Page — Sidebar and Section Navigation

**Files:**
- Create: `internal/tui/settings_page.go`

This task builds the page skeleton: sidebar rendering, section switching, and focus management. Content area shows a placeholder per section. Editing comes in Task 4.

- [ ] **Step 1: Create settings_page.go with the SettingsPage struct and sidebar**

```go
package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// settingsSection identifies a settings sidebar section
type settingsSection int

const (
	sectionAuth settingsSection = iota
	sectionGroups
	sectionPermissions
	sectionNetworks
	sectionClients
)

var settingsSections = []struct {
	Section settingsSection
	Label   string
}{
	{sectionAuth, "Authentication"},
	{sectionGroups, "Groups"},
	{sectionPermissions, "Permissions"},
	{sectionNetworks, "Networks"},
	{sectionClients, "Clients"},
}

// settingsFocus tracks which panel has focus
type settingsFocus int

const (
	settingsFocusSidebar settingsFocus = iota
	settingsFocusContent
)

// SettingsPage displays and edits account settings
type SettingsPage struct {
	accountID string
	original  models.AccountSettings // last-fetched state
	draft     models.AccountSettings // working copy being edited

	section    settingsSection // active sidebar section
	sidebarIdx int             // cursor position in sidebar
	contentIdx int             // cursor position in content area
	editingIdx int             // which field is in edit mode (-1 = none)
	editBuf    string          // text buffer for inline editing

	innerFocus settingsFocus // sidebar vs content panel
	focused    bool          // whether the page itself has focus from the app
	loading    bool
	saving     bool
	err        error
	toast      string
	dirty      bool // whether draft differs from original
}

func NewSettingsPage() *SettingsPage {
	return &SettingsPage{
		loading:    true,
		editingIdx: -1,
	}
}

func (s *SettingsPage) Title() string { return "Settings" }

func (s *SettingsPage) CursorPosition() int {
	if s.innerFocus == settingsFocusSidebar {
		return s.sidebarIdx
	}
	return s.contentIdx
}

func (s *SettingsPage) SetFocused(focused bool) {
	s.focused = focused
	if focused && s.innerFocus == settingsFocusContent {
		// Keep content focus if we had it
	} else if focused {
		s.innerFocus = settingsFocusSidebar
	}
}

func (s *SettingsPage) Init(c *client.Client) tea.Cmd {
	if s.accountID != "" {
		return nil // already loaded (cached)
	}
	s.loading = true
	s.err = nil
	return FetchSettings(c)
}

func (s *SettingsPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case SettingsLoadedMsg:
		s.loading = false
		if msg.Err != nil {
			s.err = msg.Err
			return s, nil
		}
		s.accountID = msg.Account.ID
		s.original = msg.Account.Settings
		s.draft = msg.Account.Settings
		s.dirty = false
		return s, nil

	case SettingsSavedMsg:
		s.saving = false
		if msg.Err != nil {
			s.err = msg.Err
			return s, nil
		}
		s.original = msg.Account.Settings
		s.draft = msg.Account.Settings
		s.dirty = false
		s.toast = "Settings saved"
		return s, nil

	case tea.KeyPressMsg:
		return s.handleKeyPress(msg, c)
	}

	return s, nil
}

func (s *SettingsPage) handleKeyPress(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	// Global settings keys
	if key == keyRefresh {
		s.loading = true
		s.accountID = "" // force re-fetch
		return s, FetchSettings(c)
	}

	if s.innerFocus == settingsFocusSidebar {
		return s.handleSidebarKey(key, c)
	}
	return s.handleContentKey(key, c)
}

func (s *SettingsPage) handleSidebarKey(key string, c *client.Client) (Page, tea.Cmd) {
	switch key {
	case "up", "k":
		if s.sidebarIdx > 0 {
			s.sidebarIdx--
			s.section = settingsSections[s.sidebarIdx].Section
			s.contentIdx = 0
		}
	case "down", "j":
		if s.sidebarIdx < len(settingsSections)-1 {
			s.sidebarIdx++
			s.section = settingsSections[s.sidebarIdx].Section
			s.contentIdx = 0
		}
	case "right", "l", keyEnter:
		s.innerFocus = settingsFocusContent
		s.contentIdx = 0
	}
	return s, nil
}

func (s *SettingsPage) handleContentKey(key string, c *client.Client) (Page, tea.Cmd) {
	fields := s.fieldsForSection(s.section)

	// If currently editing a text field
	if s.editingIdx >= 0 {
		return s.handleEditKey(key, c)
	}

	switch key {
	case "left", "h", keyEsc:
		s.innerFocus = settingsFocusSidebar
	case "up", "k":
		if s.contentIdx > 0 {
			s.contentIdx--
		}
	case "down", "j":
		if s.contentIdx < len(fields)-1 {
			s.contentIdx++
		}
	case keyEnter, " ":
		if s.contentIdx < len(fields) {
			f := fields[s.contentIdx]
			if f.CloudOnly {
				return s, nil // no-op for cloud-only
			}
			if f.Kind == fieldToggle {
				s.toggleField(f.ID)
				s.dirty = s.settingsChanged()
			} else {
				s.editingIdx = s.contentIdx
				s.editBuf = s.getFieldValue(f.ID)
			}
		}
	case "s":
		if s.dirty {
			s.saving = true
			return s, SaveSettings(c, s.accountID, s.draft)
		}
	}
	return s, nil
}

func (s *SettingsPage) handleEditKey(key string, c *client.Client) (Page, tea.Cmd) {
	switch key {
	case keyEnter:
		fields := s.fieldsForSection(s.section)
		if s.editingIdx < len(fields) {
			s.setFieldValue(fields[s.editingIdx].ID, s.editBuf)
			s.dirty = s.settingsChanged()
		}
		s.editingIdx = -1
		s.editBuf = ""
	case keyEsc:
		s.editingIdx = -1
		s.editBuf = ""
	case "backspace":
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

// settingsChanged compares draft to original
func (s *SettingsPage) settingsChanged() bool {
	o, d := s.original, s.draft
	if o.PeerLoginExpirationEnabled != d.PeerLoginExpirationEnabled ||
		o.PeerLoginExpiration != d.PeerLoginExpiration ||
		o.PeerInactivityExpirationEnabled != d.PeerInactivityExpirationEnabled ||
		o.PeerInactivityExpiration != d.PeerInactivityExpiration ||
		o.DNSDomain != d.DNSDomain ||
		o.NetworkRange != d.NetworkRange ||
		o.JWTGroupsEnabled != d.JWTGroupsEnabled ||
		o.JWTGroupsClaim != d.JWTGroupsClaim ||
		o.GroupsPropagationEnabled != d.GroupsPropagationEnabled ||
		o.RegularUsersViewBlocked != d.RegularUsersViewBlocked ||
		o.PeerApprovalEnabled != d.PeerApprovalEnabled ||
		o.TrafficLogging != d.TrafficLogging {
		return true
	}
	if len(o.JWTAllowGroups) != len(d.JWTAllowGroups) {
		return true
	}
	for i := range o.JWTAllowGroups {
		if o.JWTAllowGroups[i] != d.JWTAllowGroups[i] {
			return true
		}
	}
	return false
}

// View renders the two-panel settings layout
func (s *SettingsPage) View(width, height int) string {
	if s.loading {
		return loadingStyle.Render("  Loading settings...")
	}
	if s.err != nil {
		return errorStyle.Render("  Error: " + s.err.Error())
	}

	sidebarWidth := 22
	contentWidth := width - sidebarWidth - 3 // 3 for border + padding
	if contentWidth < 20 {
		contentWidth = 20
	}

	sidebar := s.renderSidebar(sidebarWidth, height)
	content := s.renderContent(contentWidth, height)

	divider := lipgloss.NewStyle().
		Foreground(colorBorder).
		Render(strings.Repeat("│\n", height))

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, divider, content)
}

func (s *SettingsPage) renderSidebar(width, height int) string {
	var b strings.Builder

	b.WriteString(sectionHeaderStyle.Render("  Settings") + "\n\n")

	for i, sec := range settingsSections {
		label := "  " + sec.Label
		var style lipgloss.Style

		switch {
		case s.focused && s.innerFocus == settingsFocusSidebar && i == s.sidebarIdx:
			style = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorSelected).
				Bold(true).
				Width(width)
		case sec.Section == s.section:
			style = lipgloss.NewStyle().
				Foreground(colorOrange).
				Bold(true).
				Width(width)
		default:
			style = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Width(width)
		}

		b.WriteString(style.Render(label) + "\n")
	}

	// Save hint at the bottom
	if s.dirty {
		b.WriteString("\n")
		modifiedStyle := lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
		b.WriteString(modifiedStyle.Render("  * Unsaved changes"))
		b.WriteString("\n")
		b.WriteString(dimHintStyle.Render("  s: save"))
	}

	if s.toast != "" {
		b.WriteString("\n")
		b.WriteString(onlineStyle.Render("  " + s.toast))
	}

	return b.String()
}
```

- [ ] **Step 2: Verify the build compiles**

Run: `go build ./...`
Expected: Likely fails because `fieldsForSection`, `renderContent`, field types, `toggleField`, `getFieldValue`, `setFieldValue` are not yet defined. That's expected — Task 4 adds them.

- [ ] **Step 3: Commit (WIP)**

```bash
git add internal/tui/settings_page.go
git commit -m "feat(tui): add settings page skeleton with sidebar navigation (WIP)"
```

---

### Task 4: Create Settings Sections — Field Definitions and Content Rendering

**Files:**
- Create: `internal/tui/settings_sections.go`

This file defines the field metadata per section and the content panel renderer. It also implements `toggleField`, `getFieldValue`, `setFieldValue` called from the page.

- [ ] **Step 1: Create settings_sections.go**

```go
package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// fieldKind identifies the type of a settings field
type fieldKind int

const (
	fieldToggle fieldKind = iota
	fieldText
	fieldDuration
)

// fieldID uniquely identifies a setting for get/set
type fieldID int

const (
	fieldPeerLoginExpEnabled fieldID = iota
	fieldPeerLoginExp
	fieldPeerInactivityExpEnabled
	fieldPeerInactivityExp
	fieldPeerApproval
	fieldGroupsPropagation
	fieldJWTGroupsEnabled
	fieldJWTGroupsClaim
	fieldJWTAllowGroups
	fieldRegularUsersViewBlocked
	fieldDNSDomain
	fieldNetworkRange
	fieldTrafficLogging
)

// settingsField describes one editable setting
type settingsField struct {
	ID        fieldID
	Label     string
	Kind      fieldKind
	CloudOnly bool
	// ConditionalOn, if set, means this field is only visible when the given field is true
	ConditionalOn fieldID
	HasCondition  bool
}

// fieldsForSection returns the field definitions for a section
func (s *SettingsPage) fieldsForSection(section settingsSection) []settingsField {
	switch section {
	case sectionAuth:
		return []settingsField{
			{fieldPeerLoginExpEnabled, "Peer Session Expiration", fieldToggle, false, 0, false},
			{fieldPeerLoginExp, "  Expiration Period", fieldDuration, false, fieldPeerLoginExpEnabled, true},
			{fieldPeerInactivityExpEnabled, "Peer Inactivity Expiration", fieldToggle, false, 0, false},
			{fieldPeerInactivityExp, "  Inactivity Period", fieldDuration, false, fieldPeerInactivityExpEnabled, true},
			{fieldPeerApproval, "Peer Approval Required", fieldToggle, true, 0, false},
		}
	case sectionGroups:
		return []settingsField{
			{fieldGroupsPropagation, "User Group Propagation", fieldToggle, false, 0, false},
			{fieldJWTGroupsEnabled, "JWT Group Sync", fieldToggle, false, 0, false},
			{fieldJWTGroupsClaim, "  JWT Claim", fieldText, false, fieldJWTGroupsEnabled, true},
			{fieldJWTAllowGroups, "  JWT Allow Groups", fieldText, false, fieldJWTGroupsEnabled, true},
		}
	case sectionPermissions:
		return []settingsField{
			{fieldRegularUsersViewBlocked, "Restrict Dashboard for Regular Users", fieldToggle, false, 0, false},
		}
	case sectionNetworks:
		return []settingsField{
			{fieldDNSDomain, "DNS Domain", fieldText, false, 0, false},
			{fieldNetworkRange, "Network Range", fieldText, false, 0, false},
		}
	case sectionClients:
		return []settingsField{
			{fieldTrafficLogging, "Traffic Logging", fieldToggle, true, 0, false},
		}
	}
	return nil
}

// visibleFields filters out conditionally hidden fields
func (s *SettingsPage) visibleFields(section settingsSection) []settingsField {
	all := s.fieldsForSection(section)
	var visible []settingsField
	for _, f := range all {
		if f.HasCondition {
			if !s.getBoolField(f.ConditionalOn) {
				continue
			}
		}
		visible = append(visible, f)
	}
	return visible
}

// getFieldValue returns the string representation of a field
func (s *SettingsPage) getFieldValue(id fieldID) string {
	switch id {
	case fieldPeerLoginExpEnabled:
		return fmt.Sprintf("%v", s.draft.PeerLoginExpirationEnabled)
	case fieldPeerLoginExp:
		return formatSettingsDuration(s.draft.PeerLoginExpiration)
	case fieldPeerInactivityExpEnabled:
		return fmt.Sprintf("%v", s.draft.PeerInactivityExpirationEnabled)
	case fieldPeerInactivityExp:
		return formatSettingsDuration(s.draft.PeerInactivityExpiration)
	case fieldPeerApproval:
		return fmt.Sprintf("%v", s.draft.PeerApprovalEnabled)
	case fieldGroupsPropagation:
		return fmt.Sprintf("%v", s.draft.GroupsPropagationEnabled)
	case fieldJWTGroupsEnabled:
		return fmt.Sprintf("%v", s.draft.JWTGroupsEnabled)
	case fieldJWTGroupsClaim:
		return s.draft.JWTGroupsClaim
	case fieldJWTAllowGroups:
		return strings.Join(s.draft.JWTAllowGroups, ", ")
	case fieldRegularUsersViewBlocked:
		return fmt.Sprintf("%v", s.draft.RegularUsersViewBlocked)
	case fieldDNSDomain:
		return s.draft.DNSDomain
	case fieldNetworkRange:
		return s.draft.NetworkRange
	case fieldTrafficLogging:
		return fmt.Sprintf("%v", s.draft.TrafficLogging)
	}
	return ""
}

// getBoolField returns the boolean value for toggle fields
func (s *SettingsPage) getBoolField(id fieldID) bool {
	switch id {
	case fieldPeerLoginExpEnabled:
		return s.draft.PeerLoginExpirationEnabled
	case fieldPeerInactivityExpEnabled:
		return s.draft.PeerInactivityExpirationEnabled
	case fieldPeerApproval:
		return s.draft.PeerApprovalEnabled
	case fieldGroupsPropagation:
		return s.draft.GroupsPropagationEnabled
	case fieldJWTGroupsEnabled:
		return s.draft.JWTGroupsEnabled
	case fieldRegularUsersViewBlocked:
		return s.draft.RegularUsersViewBlocked
	case fieldTrafficLogging:
		return s.draft.TrafficLogging
	}
	return false
}

// toggleField flips a boolean setting in the draft
func (s *SettingsPage) toggleField(id fieldID) {
	switch id {
	case fieldPeerLoginExpEnabled:
		s.draft.PeerLoginExpirationEnabled = !s.draft.PeerLoginExpirationEnabled
	case fieldPeerInactivityExpEnabled:
		s.draft.PeerInactivityExpirationEnabled = !s.draft.PeerInactivityExpirationEnabled
	case fieldPeerApproval:
		s.draft.PeerApprovalEnabled = !s.draft.PeerApprovalEnabled
	case fieldGroupsPropagation:
		s.draft.GroupsPropagationEnabled = !s.draft.GroupsPropagationEnabled
	case fieldJWTGroupsEnabled:
		s.draft.JWTGroupsEnabled = !s.draft.JWTGroupsEnabled
	case fieldRegularUsersViewBlocked:
		s.draft.RegularUsersViewBlocked = !s.draft.RegularUsersViewBlocked
	case fieldTrafficLogging:
		s.draft.TrafficLogging = !s.draft.TrafficLogging
	}
}

// setFieldValue sets a text/duration field from a string
func (s *SettingsPage) setFieldValue(id fieldID, val string) {
	switch id {
	case fieldPeerLoginExp:
		s.draft.PeerLoginExpiration = parseSettingsDuration(val)
	case fieldPeerInactivityExp:
		s.draft.PeerInactivityExpiration = parseSettingsDuration(val)
	case fieldJWTGroupsClaim:
		s.draft.JWTGroupsClaim = val
	case fieldJWTAllowGroups:
		parts := strings.Split(val, ",")
		groups := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				groups = append(groups, p)
			}
		}
		s.draft.JWTAllowGroups = groups
	case fieldDNSDomain:
		s.draft.DNSDomain = val
	case fieldNetworkRange:
		s.draft.NetworkRange = val
	}
}

// renderContent renders the active section's fields
func (s *SettingsPage) renderContent(width, height int) string {
	var b strings.Builder

	sectionLabel := settingsSections[s.sidebarIdx].Label
	b.WriteString(sectionHeaderStyle.Render("  "+sectionLabel) + "\n\n")

	fields := s.visibleFields(s.section)

	for i, f := range fields {
		isSelected := s.focused && s.innerFocus == settingsFocusContent && i == s.contentIdx
		isEditing := s.editingIdx >= 0 && i == s.editingIdx

		label := f.Label
		value := s.getFieldValue(f.ID)
		changed := s.fieldChanged(f.ID)

		var line string
		switch f.Kind {
		case fieldToggle:
			line = s.renderToggle(label, value == "true", isSelected, f.CloudOnly, changed)
		case fieldText:
			if isEditing {
				line = s.renderTextEditing(label, s.editBuf, f.CloudOnly)
			} else {
				line = s.renderTextDisplay(label, value, isSelected, f.CloudOnly, changed)
			}
		case fieldDuration:
			if isEditing {
				line = s.renderTextEditing(label, s.editBuf, f.CloudOnly)
			} else {
				line = s.renderTextDisplay(label, value, isSelected, f.CloudOnly, changed)
			}
		}

		b.WriteString(line + "\n")
	}

	// Key hints
	b.WriteString("\n")
	hints := "  ↑↓: navigate  enter/space: toggle  ←/esc: sidebar"
	if s.dirty {
		hints += "  s: save"
	}
	b.WriteString(dimHintStyle.Render(hints))

	return b.String()
}

func (s *SettingsPage) renderToggle(label string, on, selected, cloudOnly, changed bool) string {
	indicator := "○"
	valStyle := disabledStyle
	if on {
		indicator = "●"
		valStyle = enabledStyle
	}

	if cloudOnly {
		indicator = "○"
		valStyle = lipgloss.NewStyle().Foreground(colorFaint)
	}

	changeMarker := ""
	if changed {
		changeMarker = lipgloss.NewStyle().Foreground(colorYellow).Render(" *")
	}

	cloudLabel := ""
	if cloudOnly {
		cloudLabel = lipgloss.NewStyle().Foreground(colorFaint).Render(" (cloud)")
	}

	labelStr := detailLabelStyle.Copy().Width(36).Render(label)
	valStr := valStyle.Render(indicator + " " + boolLabel(on))

	line := fmt.Sprintf("  %s  %s%s%s", labelStr, valStr, cloudLabel, changeMarker)

	if selected {
		return lipgloss.NewStyle().Background(colorSelected).Render(line)
	}
	return line
}

func (s *SettingsPage) renderTextDisplay(label, value string, selected, cloudOnly, changed bool) string {
	if cloudOnly {
		label = label + " (cloud)"
	}

	changeMarker := ""
	if changed {
		changeMarker = lipgloss.NewStyle().Foreground(colorYellow).Render(" *")
	}

	labelStr := detailLabelStyle.Copy().Width(36).Render(label)
	valStr := detailValueStyle.Render(value)
	if cloudOnly {
		valStr = lipgloss.NewStyle().Foreground(colorFaint).Render(value)
	}

	line := fmt.Sprintf("  %s  %s%s", labelStr, valStr, changeMarker)

	if selected {
		return lipgloss.NewStyle().Background(colorSelected).Render(line)
	}
	return line
}

func (s *SettingsPage) renderTextEditing(label, buf string, cloudOnly bool) string {
	labelStr := detailLabelStyle.Copy().Width(36).Render(label)
	cursor := lipgloss.NewStyle().Foreground(colorOrange).Render("█")
	editStyle := lipgloss.NewStyle().
		Foreground(colorText).
		Background(lipgloss.Color("#2A2A4A")).
		Padding(0, 1)
	valStr := editStyle.Render(buf + cursor)

	return fmt.Sprintf("  %s  %s", labelStr, valStr)
}

// fieldChanged checks if a specific field differs between draft and original
func (s *SettingsPage) fieldChanged(id fieldID) bool {
	switch id {
	case fieldPeerLoginExpEnabled:
		return s.draft.PeerLoginExpirationEnabled != s.original.PeerLoginExpirationEnabled
	case fieldPeerLoginExp:
		return s.draft.PeerLoginExpiration != s.original.PeerLoginExpiration
	case fieldPeerInactivityExpEnabled:
		return s.draft.PeerInactivityExpirationEnabled != s.original.PeerInactivityExpirationEnabled
	case fieldPeerInactivityExp:
		return s.draft.PeerInactivityExpiration != s.original.PeerInactivityExpiration
	case fieldPeerApproval:
		return s.draft.PeerApprovalEnabled != s.original.PeerApprovalEnabled
	case fieldGroupsPropagation:
		return s.draft.GroupsPropagationEnabled != s.original.GroupsPropagationEnabled
	case fieldJWTGroupsEnabled:
		return s.draft.JWTGroupsEnabled != s.original.JWTGroupsEnabled
	case fieldJWTGroupsClaim:
		return s.draft.JWTGroupsClaim != s.original.JWTGroupsClaim
	case fieldJWTAllowGroups:
		return strings.Join(s.draft.JWTAllowGroups, ",") != strings.Join(s.original.JWTAllowGroups, ",")
	case fieldRegularUsersViewBlocked:
		return s.draft.RegularUsersViewBlocked != s.original.RegularUsersViewBlocked
	case fieldDNSDomain:
		return s.draft.DNSDomain != s.original.DNSDomain
	case fieldNetworkRange:
		return s.draft.NetworkRange != s.original.NetworkRange
	case fieldTrafficLogging:
		return s.draft.TrafficLogging != s.original.TrafficLogging
	}
	return false
}

func boolLabel(b bool) string {
	if b {
		return "Enabled"
	}
	return "Disabled"
}

// formatSettingsDuration formats seconds into a human-friendly string
func formatSettingsDuration(seconds int) string {
	if seconds <= 0 {
		return "disabled"
	}
	d := time.Duration(seconds) * time.Second
	if d >= 24*time.Hour {
		days := int(d.Hours()) / 24
		return fmt.Sprintf("%dd", days)
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// parseSettingsDuration parses a human-friendly duration string to seconds
// Supports: "30d", "24h", "60m", or raw seconds "3600"
func parseSettingsDuration(s string) int {
	s = strings.TrimSpace(s)
	if s == "" || s == "disabled" || s == "0" {
		return 0
	}

	// Try suffixed formats
	if len(s) > 1 {
		numStr := s[:len(s)-1]
		suffix := s[len(s)-1]
		var n int
		if _, err := fmt.Sscanf(numStr, "%d", &n); err == nil {
			switch suffix {
			case 'd':
				return n * 86400
			case 'h':
				return n * 3600
			case 'm':
				return n * 60
			}
		}
	}

	// Try raw seconds
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err == nil {
		return n
	}
	return 0
}
```

- [ ] **Step 2: Verify the build compiles**

Run: `go build ./...`
Expected: PASS — all methods referenced by settings_page.go are now defined.

- [ ] **Step 3: Commit**

```bash
git add internal/tui/settings_sections.go
git commit -m "feat(tui): add settings field definitions and content rendering"
```

---

### Task 5: Wire Settings Page into App Navigation

**Files:**
- Modify: `internal/tui/nav.go:13,38-54` (rename SectionAccount → SectionSettings, update navItems)
- Modify: `internal/tui/app.go:44-76` (replace AccountsPage with SettingsPage)
- Delete: `internal/tui/accounts_page.go`

- [ ] **Step 1: Rename SectionAccount to SectionSettings in nav.go**

In `nav.go`, change the Section constant:

```go
// Replace line 25:
SectionAccount
// With:
SectionSettings
```

In the `navItems` slice, change the Account entry:

```go
// Replace line 53:
{SectionAccount, "Account", ""},
// With:
{SectionSettings, "Settings", ""},
```

- [ ] **Step 2: Update app.go to use SettingsPage**

In `app.go`, replace line 65:

```go
// Replace:
p[SectionAccount] = NewAccountsPage()
// With:
p[SectionSettings] = NewSettingsPage()
```

- [ ] **Step 3: Delete accounts_page.go**

Run: `rm internal/tui/accounts_page.go`

- [ ] **Step 4: Update any remaining references to SectionAccount or AccountsPage**

Search the codebase for `SectionAccount` and `AccountsPage` — these should only appear in files we've already modified (nav.go, app.go) and messages.go. The `AccountsLoadedMsg` in messages.go and `FetchAccounts` in api.go can stay — they're used by the dashboard page for counts. But no page references them anymore, so they're inert but harmless.

Run: `grep -rn "SectionAccount\|AccountsPage\|NewAccountsPage" internal/tui/`
Expected: No matches (all replaced or deleted)

- [ ] **Step 5: Verify the build compiles**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/tui/nav.go internal/tui/app.go
git rm internal/tui/accounts_page.go
git commit -m "feat(tui): replace Account page with Settings page"
```

---

### Task 6: Add Save Confirmation Screen

**Files:**
- Modify: `internal/tui/settings_page.go` (add confirm state and rendering)

The save flow should show a confirmation screen listing changed fields before committing. This reuses the existing `RenderConfirm` pattern from `confirm.go`.

- [ ] **Step 1: Add confirm state to SettingsPage**

Add a `confirming` boolean field to the `SettingsPage` struct (after `saving`):

```go
confirming bool // showing save confirmation screen
```

- [ ] **Step 2: Update handleContentKey to show confirm instead of saving directly**

Replace the `"s"` case in `handleContentKey`:

```go
case "s":
	if s.dirty {
		s.confirming = true
	}
```

- [ ] **Step 3: Add confirm key handling in handleKeyPress**

Add at the top of `handleKeyPress`, before the refresh check:

```go
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
```

- [ ] **Step 4: Add confirm rendering in View**

Add after the loading/error checks at the top of `View`, before the sidebar/content rendering:

```go
if s.confirming {
	return s.renderConfirm(width)
}
if s.saving {
	return loadingStyle.Render("  Saving settings...")
}
```

- [ ] **Step 5: Implement renderConfirm**

Add this method to `settings_page.go`:

```go
func (s *SettingsPage) renderConfirm(width int) string {
	var fields []ConfirmField

	allFields := []fieldID{
		fieldPeerLoginExpEnabled, fieldPeerLoginExp,
		fieldPeerInactivityExpEnabled, fieldPeerInactivityExp,
		fieldPeerApproval, fieldGroupsPropagation,
		fieldJWTGroupsEnabled, fieldJWTGroupsClaim, fieldJWTAllowGroups,
		fieldRegularUsersViewBlocked, fieldDNSDomain, fieldNetworkRange,
		fieldTrafficLogging,
	}

	fieldLabels := map[fieldID]string{
		fieldPeerLoginExpEnabled:      "Peer Session Expiration",
		fieldPeerLoginExp:             "Expiration Period",
		fieldPeerInactivityExpEnabled: "Peer Inactivity Expiration",
		fieldPeerInactivityExp:        "Inactivity Period",
		fieldPeerApproval:             "Peer Approval",
		fieldGroupsPropagation:        "Group Propagation",
		fieldJWTGroupsEnabled:         "JWT Group Sync",
		fieldJWTGroupsClaim:           "JWT Claim",
		fieldJWTAllowGroups:           "JWT Allow Groups",
		fieldRegularUsersViewBlocked:  "Restrict Dashboard",
		fieldDNSDomain:                "DNS Domain",
		fieldNetworkRange:             "Network Range",
		fieldTrafficLogging:           "Traffic Logging",
	}

	for _, id := range allFields {
		if s.fieldChanged(id) {
			origVal := s.getOriginalFieldValue(id)
			draftVal := s.getFieldValue(id)
			fields = append(fields, ConfirmField{
				Label: fieldLabels[id],
				Value: origVal + " → " + draftVal,
			})
		}
	}

	return RenderConfirm("Save Settings", fields)
}
```

- [ ] **Step 6: Add getOriginalFieldValue helper**

Add to `settings_sections.go`, copying `getFieldValue` but reading from `s.original` instead of `s.draft`:

```go
func (s *SettingsPage) getOriginalFieldValue(id fieldID) string {
	switch id {
	case fieldPeerLoginExpEnabled:
		return fmt.Sprintf("%v", s.original.PeerLoginExpirationEnabled)
	case fieldPeerLoginExp:
		return formatSettingsDuration(s.original.PeerLoginExpiration)
	case fieldPeerInactivityExpEnabled:
		return fmt.Sprintf("%v", s.original.PeerInactivityExpirationEnabled)
	case fieldPeerInactivityExp:
		return formatSettingsDuration(s.original.PeerInactivityExpiration)
	case fieldPeerApproval:
		return fmt.Sprintf("%v", s.original.PeerApprovalEnabled)
	case fieldGroupsPropagation:
		return fmt.Sprintf("%v", s.original.GroupsPropagationEnabled)
	case fieldJWTGroupsEnabled:
		return fmt.Sprintf("%v", s.original.JWTGroupsEnabled)
	case fieldJWTGroupsClaim:
		return s.original.JWTGroupsClaim
	case fieldJWTAllowGroups:
		return strings.Join(s.original.JWTAllowGroups, ", ")
	case fieldRegularUsersViewBlocked:
		return fmt.Sprintf("%v", s.original.RegularUsersViewBlocked)
	case fieldDNSDomain:
		return s.original.DNSDomain
	case fieldNetworkRange:
		return s.original.NetworkRange
	case fieldTrafficLogging:
		return fmt.Sprintf("%v", s.original.TrafficLogging)
	}
	return ""
}
```

- [ ] **Step 7: Verify the build compiles**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/tui/settings_page.go internal/tui/settings_sections.go
git commit -m "feat(tui): add save confirmation screen to settings page"
```

---

### Task 7: Handle Content Cursor with Visible Fields

**Files:**
- Modify: `internal/tui/settings_page.go` (fix content navigation to use visible fields)

The content cursor (`contentIdx`) must track against `visibleFields`, not `fieldsForSection`, since conditional fields can appear/disappear when toggles change.

- [ ] **Step 1: Update handleContentKey to use visibleFields**

Replace the `fields := s.fieldsForSection(s.section)` line in `handleContentKey` with:

```go
fields := s.visibleFields(s.section)
```

- [ ] **Step 2: Clamp contentIdx when fields collapse**

Add after every `toggleField` call in `handleContentKey`:

```go
// Clamp cursor if visible fields changed
visible := s.visibleFields(s.section)
if s.contentIdx >= len(visible) {
	s.contentIdx = len(visible) - 1
}
if s.contentIdx < 0 {
	s.contentIdx = 0
}
```

- [ ] **Step 3: Verify the build compiles**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/tui/settings_page.go
git commit -m "fix(tui): use visible fields for settings content cursor navigation"
```

---

### Task 8: Smoke Test and Polish

**Files:**
- Modify: `internal/tui/settings_page.go` (status bar hints)
- Modify: `internal/tui/app.go` (status hints for settings)

- [ ] **Step 1: Build and run the TUI**

Run: `go build -o netbird-manage ./cmd/netbird-manage/ && ./netbird-manage tui`

Manually verify:
1. Navigate to the Settings tab (previously Account)
2. Sidebar shows 5 sections: Authentication, Groups, Permissions, Networks, Clients
3. Arrow keys navigate sidebar, right/enter moves to content
4. Toggle fields flip with enter/space, show `*` marker
5. Text fields enter edit mode, accept input, save on enter
6. `s` shows confirmation with changed fields
7. `y` saves, API call succeeds, `*` markers clear
8. Cloud-only fields are grayed out and non-interactive
9. JWT sub-fields appear/disappear when JWT toggle changes

- [ ] **Step 2: Fix any issues found during smoke test**

Address visual alignment, key binding conflicts, or rendering issues.

- [ ] **Step 3: Final commit**

```bash
git add -A
git commit -m "feat(tui): settings page with inline editing and save confirmation"
```
