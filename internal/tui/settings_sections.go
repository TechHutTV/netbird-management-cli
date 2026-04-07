package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"netbird-manage/internal/models"
)

// settingsSection identifies a sidebar section
type settingsSection int

const (
	sectionAuthentication settingsSection = iota
	sectionGroups
	sectionPermissions
	sectionNetworks
	sectionClients
)

var settingsSectionLabels = []string{
	"Authentication",
	"Groups",
	"Permissions",
	"Networks",
	"Clients",
}

// fieldKind distinguishes how a field is rendered and edited
type fieldKind int

const (
	fieldToggle   fieldKind = iota
	fieldText               // free-form text
	fieldDuration           // seconds, displayed/edited as "30d"/"24h" etc.
)

// fieldID is a unique identifier for each settings field
type fieldID int

const (
	// Authentication
	fieldPeerLoginExpEnabled fieldID = iota
	fieldPeerLoginExp
	fieldPeerInactivityExpEnabled
	fieldPeerInactivityExp
	fieldPeerApproval

	// Groups
	fieldGroupsPropagation
	fieldJWTGroupsEnabled
	fieldJWTGroupsClaim
	fieldJWTAllowGroups

	// Permissions
	fieldRegularUsersViewBlocked

	// Networks
	fieldDNSDomain
	fieldNetworkRange

	// Clients
	fieldTrafficLogging
)

// fieldDef describes a single settings field
type fieldDef struct {
	ID            fieldID
	Label         string
	Description   string
	Placeholder   string // shown greyed out when value is empty
	Kind          fieldKind
	CloudOnly     bool
	HasCondition  bool
	ConditionalOn fieldID
}

// fieldsForSection returns all fields (including conditional ones) for a section
func fieldsForSection(sec settingsSection) []fieldDef {
	switch sec {
	case sectionAuthentication:
		return []fieldDef{
			{
				ID: fieldPeerLoginExpEnabled, Label: "Peer Session Expiration", Kind: fieldToggle,
				Description: "Request periodic re-authentication of peers registered with SSO.",
			},
			{
				ID: fieldPeerLoginExp, Label: "Expiration Period", Kind: fieldDuration,
				Description:  "Time after which every peer added with SSO login will require re-authentication.",
				Placeholder: "e.g. 7d", HasCondition: true, ConditionalOn: fieldPeerLoginExpEnabled,
			},
			{
				ID: fieldPeerInactivityExpEnabled, Label: "Peer Inactivity Expiration", Kind: fieldToggle,
				Description: "Require authentication after users disconnect from management.",
			},
			{
				ID: fieldPeerInactivityExp, Label: "Inactivity Period", Kind: fieldDuration,
				Description:  "Time of inactivity after which peers require re-authentication.",
				Placeholder: "e.g. 10m", HasCondition: true, ConditionalOn: fieldPeerInactivityExpEnabled,
			},
			{
				ID: fieldPeerApproval, Label: "Peer Approval Required", Kind: fieldToggle, CloudOnly: true,
				Description: "Require manual approval for new users joining via domain matching.",
			},
		}
	case sectionGroups:
		return []fieldDef{
			{
				ID: fieldGroupsPropagation, Label: "User Group Propagation", Kind: fieldToggle,
				Description: "Allow group propagation from user's auto-groups to peers, sharing membership information.",
			},
			{
				ID: fieldJWTGroupsEnabled, Label: "JWT Group Sync", Kind: fieldToggle,
				Description: "Extract and sync groups from JWT claims with user's auto-groups, auto-creating groups from tokens.",
			},
			{
				ID: fieldJWTGroupsClaim, Label: "JWT Claim", Kind: fieldText,
				Description:  "Specify the JWT claim name for extracting group names.",
				Placeholder: "e.g. roles", HasCondition: true, ConditionalOn: fieldJWTGroupsEnabled,
			},
			{
				ID: fieldJWTAllowGroups, Label: "JWT Allow Groups", Kind: fieldText,
				Description:  "Limit NetBird access to specific group names from your identity provider.",
				HasCondition: true, ConditionalOn: fieldJWTGroupsEnabled,
			},
		}
	case sectionPermissions:
		return []fieldDef{
			{
				ID: fieldRegularUsersViewBlocked, Label: "Restrict Dashboard for Regular Users", Kind: fieldToggle,
				Description: "Access to the dashboard will be limited and regular users will not be able to view any peers.",
			},
		}
	case sectionNetworks:
		return []fieldDef{
			{
				ID: fieldDNSDomain, Label: "DNS Domain", Kind: fieldText,
				Description: "Specify a custom peer DNS domain for your network. This should not point to a valid domain to avoid overriding DNS results.",
				Placeholder: "netbird.cloud",
			},
			{
				ID: fieldNetworkRange, Label: "Network Range", Kind: fieldText,
				Description: "Specify a custom IPv4 range for your network in CIDR format. All peer IPs will be re-allocated when changed.",
				Placeholder: "e.g. 100.64.0.0/16",
			},
		}
	case sectionClients:
		return []fieldDef{
			{
				ID: fieldTrafficLogging, Label: "Traffic Logging", Kind: fieldToggle, CloudOnly: true,
				Description: "Enable network traffic logging for monitoring and analytics.",
			},
		}
	}
	return nil
}

// visibleFields filters conditional fields based on current draft state
func (s *SettingsPage) visibleFields(sec settingsSection) []fieldDef {
	all := fieldsForSection(sec)
	result := make([]fieldDef, 0, len(all))
	for _, f := range all {
		if f.HasCondition {
			if !s.getBoolField(f.ConditionalOn) {
				continue
			}
		}
		result = append(result, f)
	}
	return result
}

// ── Field value accessors ────────────────────────────────────────────

// fieldValueFrom returns the display string for a field from the given settings
func fieldValueFrom(st models.AccountSettings, id fieldID) string {
	switch id {
	case fieldPeerLoginExpEnabled:
		return boolStr(st.PeerLoginExpirationEnabled)
	case fieldPeerLoginExp:
		return formatSettingsDuration(st.PeerLoginExpiration)
	case fieldPeerInactivityExpEnabled:
		return boolStr(st.PeerInactivityExpirationEnabled)
	case fieldPeerInactivityExp:
		return formatSettingsDuration(st.PeerInactivityExpiration)
	case fieldPeerApproval:
		return boolStr(st.PeerApprovalEnabled)
	case fieldGroupsPropagation:
		return boolStr(st.GroupsPropagationEnabled)
	case fieldJWTGroupsEnabled:
		return boolStr(st.JWTGroupsEnabled)
	case fieldJWTGroupsClaim:
		return st.JWTGroupsClaim
	case fieldJWTAllowGroups:
		return strings.Join(st.JWTAllowGroups, ", ")
	case fieldRegularUsersViewBlocked:
		return boolStr(st.RegularUsersViewBlocked)
	case fieldDNSDomain:
		return st.DNSDomain
	case fieldNetworkRange:
		return st.NetworkRange
	case fieldTrafficLogging:
		return boolStr(st.TrafficLogging)
	}
	return ""
}

// getFieldValue returns the current draft value as a display string
func (s *SettingsPage) getFieldValue(id fieldID) string {
	return fieldValueFrom(s.draft, id)
}

// getOriginalFieldValue returns the original (pre-edit) value as a display string
func (s *SettingsPage) getOriginalFieldValue(id fieldID) string {
	return fieldValueFrom(s.original, id)
}

// getBoolField reads a boolean field from the draft
func (s *SettingsPage) getBoolField(id fieldID) bool {
	st := s.draft
	switch id {
	case fieldPeerLoginExpEnabled:
		return st.PeerLoginExpirationEnabled
	case fieldPeerInactivityExpEnabled:
		return st.PeerInactivityExpirationEnabled
	case fieldPeerApproval:
		return st.PeerApprovalEnabled
	case fieldGroupsPropagation:
		return st.GroupsPropagationEnabled
	case fieldJWTGroupsEnabled:
		return st.JWTGroupsEnabled
	case fieldRegularUsersViewBlocked:
		return st.RegularUsersViewBlocked
	case fieldTrafficLogging:
		return st.TrafficLogging
	}
	return false
}

// toggleField flips a boolean field in the draft, returning a new draft
func (s *SettingsPage) toggleField(id fieldID) {
	st := s.draft
	switch id {
	case fieldPeerLoginExpEnabled:
		st.PeerLoginExpirationEnabled = !st.PeerLoginExpirationEnabled
	case fieldPeerInactivityExpEnabled:
		st.PeerInactivityExpirationEnabled = !st.PeerInactivityExpirationEnabled
	case fieldPeerApproval:
		st.PeerApprovalEnabled = !st.PeerApprovalEnabled
	case fieldGroupsPropagation:
		st.GroupsPropagationEnabled = !st.GroupsPropagationEnabled
	case fieldJWTGroupsEnabled:
		st.JWTGroupsEnabled = !st.JWTGroupsEnabled
	case fieldRegularUsersViewBlocked:
		st.RegularUsersViewBlocked = !st.RegularUsersViewBlocked
	case fieldTrafficLogging:
		st.TrafficLogging = !st.TrafficLogging
	}
	s.draft = st
}

// setFieldValue applies a parsed text/duration value to the draft
func (s *SettingsPage) setFieldValue(id fieldID, val string) {
	st := s.draft
	switch id {
	case fieldPeerLoginExp:
		st.PeerLoginExpiration = parseSettingsDuration(val)
	case fieldPeerInactivityExp:
		st.PeerInactivityExpiration = parseSettingsDuration(val)
	case fieldJWTGroupsClaim:
		st.JWTGroupsClaim = val
	case fieldJWTAllowGroups:
		st.JWTAllowGroups = splitCSV(val)
	case fieldDNSDomain:
		st.DNSDomain = val
	case fieldNetworkRange:
		st.NetworkRange = val
	}
	s.draft = st
}

// fieldChanged returns true if a field differs between draft and original
func (s *SettingsPage) fieldChanged(id fieldID) bool {
	return s.getFieldValue(id) != s.getOriginalFieldValue(id)
}

// settingsChanged returns true if any field differs between draft and original
func (s *SettingsPage) settingsChanged() bool {
	allFields := []fieldID{
		fieldPeerLoginExpEnabled, fieldPeerLoginExp,
		fieldPeerInactivityExpEnabled, fieldPeerInactivityExp,
		fieldPeerApproval,
		fieldGroupsPropagation, fieldJWTGroupsEnabled,
		fieldJWTGroupsClaim, fieldJWTAllowGroups,
		fieldRegularUsersViewBlocked,
		fieldDNSDomain, fieldNetworkRange,
		fieldTrafficLogging,
	}
	for _, id := range allFields {
		if s.fieldChanged(id) {
			return true
		}
	}
	return false
}

// ── Rendering ────────────────────────────────────────────────────────

const sidebarWidth = 22

// renderSidebar renders the left sidebar with section list
func (s *SettingsPage) renderSidebar(height int) string {
	var rows []string
	for i, label := range settingsSectionLabels {
		sec := settingsSection(i)

		prefix := "  "
		var style lipgloss.Style

		switch {
		case s.focused && s.innerFocus == focusSidebar && s.sidebarIdx == i:
			prefix = "▸ "
			style = lipgloss.NewStyle().
				Foreground(colorOrange).
				Background(colorSelected).
				Bold(true).
				Width(sidebarWidth)
		case sec == s.section:
			prefix = "  "
			style = lipgloss.NewStyle().
				Foreground(colorText).
				Bold(true).
				Width(sidebarWidth)
		default:
			style = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Width(sidebarWidth)
		}

		rows = append(rows, style.Render(prefix+label))
	}

	// Dirty indicator at the bottom
	if s.dirty {
		rows = append(rows, "")
		marker := lipgloss.NewStyle().Foreground(colorYellow).Bold(true).Render("  ● Unsaved changes")
		rows = append(rows, marker)
	}

	content := strings.Join(rows, "\n")

	return lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(height).
		Render(content)
}

// renderContent renders the right content area for the active section
func (s *SettingsPage) renderContent(width, height int) string {
	if s.loading {
		return loadingStyle.Render("  Loading settings...")
	}
	if s.err != nil {
		return errorStyle.Render("  Error: " + s.err.Error())
	}
	if !s.loaded {
		return dimHintStyle.Render("  Press enter to load settings.")
	}

	visible := s.visibleFields(s.section)
	if len(visible) == 0 {
		return dimHintStyle.Render("  No fields for this section.")
	}

	sectionTitle := settingsSectionLabels[int(s.section)]
	header := lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true).
		MarginBottom(1).
		Render("  " + sectionTitle)

	var rows []string
	rows = append(rows, header)
	rows = append(rows, "")

	for i, f := range visible {
		isSelected := s.focused && s.innerFocus == focusContentPanel && i == s.contentIdx
		isEditing := s.editingIdx == i && s.innerFocus == focusContentPanel

		var card string
		switch f.Kind {
		case fieldToggle:
			card = s.renderToggleCard(f, isSelected, width)
		case fieldText, fieldDuration:
			if isEditing {
				card = s.renderTextEditCard(f, width)
			} else {
				card = s.renderTextCard(f, isSelected, width)
			}
		}
		rows = append(rows, card)
	}

	// Key hints at bottom
	rows = append(rows, "")
	rows = append(rows, dimHintStyle.Render("  "+s.contentHints()))

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Render(strings.Join(rows, "\n"))
}

// wrapText breaks long text into lines that fit within maxWidth
func wrapText(text string, maxWidth int) string {
	if maxWidth <= 0 || len(text) <= maxWidth {
		return text
	}
	words := strings.Fields(text)
	var lines []string
	current := ""
	for _, w := range words {
		if current == "" {
			current = w
		} else if len(current)+1+len(w) <= maxWidth {
			current += " " + w
		} else {
			lines = append(lines, current)
			current = w
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, "\n")
}

// cardSeparator renders a dim horizontal rule
func cardSeparator(width int) string {
	line := strings.Repeat("─", width-4)
	return lipgloss.NewStyle().Foreground(colorBorderFaint).Render("  " + line)
}

// renderToggleCard renders a toggle setting as a stacked card
func (s *SettingsPage) renderToggleCard(f fieldDef, selected bool, width int) string {
	val := s.getBoolField(f.ID)
	changed := s.fieldChanged(f.ID)
	cardWidth := width - 4
	if cardWidth < 20 {
		cardWidth = 20
	}

	// Toggle line
	var toggle string
	if f.CloudOnly {
		toggle = lipgloss.NewStyle().Foreground(colorFaint).Render("○ OFF  (cloud-only)")
	} else if val {
		toggle = enabledStyle.Render("● ON")
	} else {
		toggle = disabledStyle.Render("○ OFF")
	}
	if changed {
		toggle += lipgloss.NewStyle().Foreground(colorYellow).Render("  (modified)")
	}

	// Styles
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(colorText)
	descStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	if f.CloudOnly {
		labelStyle = lipgloss.NewStyle().Foreground(colorFaint)
		descStyle = lipgloss.NewStyle().Foreground(colorFaint)
	} else if selected {
		labelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorOrange)
	}

	// Build stacked card
	var lines []string
	lines = append(lines, labelStyle.Render(f.Label))
	lines = append(lines, descStyle.Render(wrapText(f.Description, cardWidth)))
	lines = append(lines, toggle)

	card := strings.Join(lines, "\n")

	if selected && !f.CloudOnly {
		return lipgloss.NewStyle().
			Background(colorSelected).
			Width(cardWidth).
			Padding(0, 2).
			Render(card) + "\n" + cardSeparator(width)
	}

	return lipgloss.NewStyle().
		Padding(0, 2).
		Render(card) + "\n" + cardSeparator(width)
}

// renderTextCard renders a text/duration setting as a stacked card
func (s *SettingsPage) renderTextCard(f fieldDef, selected bool, width int) string {
	val := s.getFieldValue(f.ID)
	changed := s.fieldChanged(f.ID)
	cardWidth := width - 4
	if cardWidth < 20 {
		cardWidth = 20
	}

	// Value display
	valBoxWidth := cardWidth
	if valBoxWidth > 40 {
		valBoxWidth = 40
	}

	var valBox string
	if val == "" && f.Placeholder != "" {
		// Show placeholder greyed out
		valBox = lipgloss.NewStyle().
			Foreground(colorFaint).
			Background(adaptive("#E2E8F0", "#1A2332")).
			Width(valBoxWidth).
			Padding(0, 1).
			Render(f.Placeholder)
	} else {
		displayVal := val
		if displayVal == "" {
			displayVal = "(empty)"
		}
		valBox = lipgloss.NewStyle().
			Foreground(colorText).
			Background(adaptive("#E2E8F0", "#1A2332")).
			Width(valBoxWidth).
			Padding(0, 1).
			Render(displayVal)
	}

	changeStr := ""
	if changed {
		changeStr = lipgloss.NewStyle().Foreground(colorYellow).Render("  (modified)")
	}

	// Styles
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(colorText)
	if selected {
		labelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorOrange)
	}
	descStyle := lipgloss.NewStyle().Foreground(colorTextDim)

	// Build stacked card
	var lines []string
	lines = append(lines, labelStyle.Render(f.Label)+changeStr)
	lines = append(lines, descStyle.Render(wrapText(f.Description, cardWidth)))
	lines = append(lines, valBox)

	card := strings.Join(lines, "\n")

	if selected {
		return lipgloss.NewStyle().
			Background(colorSelected).
			Width(cardWidth).
			Padding(0, 2).
			Render(card) + "\n" + cardSeparator(width)
	}

	return lipgloss.NewStyle().
		Padding(0, 2).
		Render(card) + "\n" + cardSeparator(width)
}

// renderTextEditCard renders a field in edit mode as a stacked card
func (s *SettingsPage) renderTextEditCard(f fieldDef, width int) string {
	cardWidth := width - 4
	if cardWidth < 20 {
		cardWidth = 20
	}
	valBoxWidth := cardWidth
	if valBoxWidth > 40 {
		valBoxWidth = 40
	}

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(colorOrange)
	descStyle := lipgloss.NewStyle().Foreground(colorTextDim)

	cursor := lipgloss.NewStyle().Foreground(colorOrange).Render("▌")
	editBox := lipgloss.NewStyle().
		Foreground(colorText).
		Background(adaptive("#CBD5E0", "#1A2332")).
		Width(valBoxWidth).
		Padding(0, 1).
		Render(s.editBuf + cursor)

	hint := dimHintStyle.Render("enter: confirm  esc: cancel  backspace: delete")

	var lines []string
	lines = append(lines, labelStyle.Render(f.Label))
	lines = append(lines, descStyle.Render(wrapText(f.Description, cardWidth)))
	lines = append(lines, editBox)
	lines = append(lines, hint)

	card := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Background(colorSelected).
		Width(cardWidth).
		Padding(0, 2).
		Render(card) + "\n" + cardSeparator(width)
}

// contentHints returns the key-hint line for the content area
func (s *SettingsPage) contentHints() string {
	if s.editingIdx >= 0 {
		return "enter: confirm  esc: cancel  backspace: delete"
	}
	hints := "↑↓: navigate  ←/esc: sidebar"
	visible := s.visibleFields(s.section)
	if s.contentIdx >= 0 && s.contentIdx < len(visible) {
		f := visible[s.contentIdx]
		if !f.CloudOnly {
			switch f.Kind {
			case fieldToggle:
				hints += "  enter/space: toggle"
			case fieldText, fieldDuration:
				hints += "  enter: edit"
			}
		}
	}
	if s.dirty {
		hints += "  s: save"
	}
	return hints
}

// ── Duration helpers ─────────────────────────────────────────────────

// formatSettingsDuration converts seconds to a human-readable string like "30d", "24h", "60m"
func formatSettingsDuration(seconds int) string {
	if seconds <= 0 {
		return "0"
	}
	days := seconds / 86400
	if days > 0 && seconds%86400 == 0 {
		return fmt.Sprintf("%dd", days)
	}
	hours := seconds / 3600
	if hours > 0 && seconds%3600 == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	minutes := seconds / 60
	if minutes > 0 && seconds%60 == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%ds", seconds)
}

// parseSettingsDuration parses "30d", "24h", "60m", "90s", or plain integer seconds
func parseSettingsDuration(s string) int {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0
	}
	// Try suffixed formats (e.g. "30d", "24h", "60m", "90s")
	if len(s) >= 2 {
		suffix := s[len(s)-1]
		numStr := s[:len(s)-1]
		var n int
		if _, err := fmt.Sscanf(numStr, "%d", &n); err == nil {
			switch suffix {
			case 'd':
				return n * 86400
			case 'h':
				return n * 3600
			case 'm':
				return n * 60
			case 's':
				return n
			}
		}
	}
	// Try parsing the whole string as plain integer seconds
	var total int
	if _, err := fmt.Sscanf(s, "%d", &total); err != nil {
		return 0
	}
	return total
}

// splitCSV splits a comma-separated string into trimmed parts
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// boolStr converts a bool to "enabled"/"disabled"
func boolStr(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

// changedConfirmFields builds a ConfirmField list of all modified settings for the confirm screen
func (s *SettingsPage) changedConfirmFields() []ConfirmField {
	allSections := []settingsSection{
		sectionAuthentication, sectionGroups, sectionPermissions, sectionNetworks, sectionClients,
	}
	var fields []ConfirmField
	for _, sec := range allSections {
		for _, f := range fieldsForSection(sec) {
			if !s.fieldChanged(f.ID) {
				continue
			}
			old := s.getOriginalFieldValue(f.ID)
			new := s.getFieldValue(f.ID)
			fields = append(fields, ConfirmField{
				Label: f.Label,
				Value: fmt.Sprintf("%s → %s", old, new),
			})
		}
	}
	return fields
}

// draftFromAccount copies relevant fields from an Account into an AccountSettings draft
func draftFromAccount(a models.Account) models.AccountSettings {
	return models.AccountSettings{
		PeerLoginExpirationEnabled:      a.Settings.PeerLoginExpirationEnabled,
		PeerLoginExpiration:             a.Settings.PeerLoginExpiration,
		PeerInactivityExpirationEnabled: a.Settings.PeerInactivityExpirationEnabled,
		PeerInactivityExpiration:        a.Settings.PeerInactivityExpiration,
		DNSDomain:                       a.Settings.DNSDomain,
		NetworkRange:                    a.Settings.NetworkRange,
		JWTGroupsEnabled:                a.Settings.JWTGroupsEnabled,
		JWTGroupsClaim:                  a.Settings.JWTGroupsClaim,
		JWTAllowGroups:                  append([]string(nil), a.Settings.JWTAllowGroups...),
		GroupsPropagationEnabled:        a.Settings.GroupsPropagationEnabled,
		RegularUsersViewBlocked:         a.Settings.RegularUsersViewBlocked,
		PeerApprovalEnabled:             a.Settings.PeerApprovalEnabled,
		TrafficLogging:                  a.Settings.TrafficLogging,
	}
}
