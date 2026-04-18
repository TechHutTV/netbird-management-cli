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

// rpOuterState is the outermost mode switch of the ReverseProxy page.
type rpOuterState int

const (
	rpOuterList rpOuterState = iota
	rpOuterDetail
	rpOuterDomains
	rpOuterEvents
	rpOuterWizard
)

// ReverseProxyPage implements the Page interface for NetBird's reverse-proxy services.
// It mirrors the dashboard's /reverse-proxy/services experience: a list view, a tabbed
// detail view, a multi-screen wizard for create/edit, plus sub-views for custom domains
// and proxy events.
type ReverseProxyPage struct {
	// Data cache
	services []models.ReverseProxyService
	filtered []models.ReverseProxyService
	clusters []models.ReverseProxyCluster
	groups   map[string]string // group id → name
	peers    map[string]string // peer id → name

	// View lifecycle
	loading bool
	err     error
	focused bool

	// List-view state
	cursor    int
	search    string
	searching bool

	// Outer state
	state rpOuterState

	// Detail-view state
	detailTab rpDetailTab

	// Sub-views
	domains rpDomainsView
	events  rpEventsView

	// Wizard state
	wiz rpWizard
}

// rpWizard carries the in-progress service being created or edited.
type rpWizard struct {
	editing   bool   // true = editing existing; false = creating new
	editingID string // service ID when editing

	// err is a wizard-local validation/submission error. Shown above the current
	// wizard screen and cleared on any transition. Kept separate from page.err
	// so a wizard error never blanks the whole page.
	err error

	// The full service being built. Written into incrementally as wizard steps complete.
	svc models.ReverseProxyService

	// Current wizard step.
	step rpWizardStep

	// Linear return step for sub-edits (target-edit, access-rule, auth-sub...).
	// For example, when editing a single target, parentStep = wizStepTargets.
	parentStep rpWizardStep

	// Which row is being edited in list-editor sub-screens (−1 = adding new).
	editingIndex int

	// Cursor positions within list-editor screens.
	targetsCursor       int
	accessCursor        int
	authCursor          int // which of the 4 auth methods (+ headers list)
	headerAuthsCursor   int
	customHeadersCursor int

	// Which target index we're editing the custom-headers map of.
	customHeadersForTarget int

	// Current huh form + its backing data struct, only one at a time is non-nil.
	form             *huh.Form
	serviceData      rpServiceFormData
	targetData       rpTargetFormData
	accessData       rpAccessRuleFormData
	advancedData     rpAdvancedFormData
	passwordData     rpPasswordFormData
	pinData          rpPinFormData
	bearerData       rpBearerFormData
	linkData         rpLinkFormData
	headerAuthData   rpHeaderAuthFormData
	customHeaderData rpCustomHeaderFormData
}

// NewReverseProxyPage constructs a zero-value ReverseProxyPage.
func NewReverseProxyPage() *ReverseProxyPage {
	return &ReverseProxyPage{loading: true}
}

// Title returns the human name shown in the status bar.
func (p *ReverseProxyPage) Title() string { return "Reverse Proxy" }

// CursorPosition returns the logical cursor for the nav focus-model.
// In any non-list state, return non-zero to block "up" from returning to nav.
func (p *ReverseProxyPage) CursorPosition() int {
	if p.state != rpOuterList {
		return 1
	}
	return p.cursor
}

// SetFocused marks whether the page owns keyboard focus.
func (p *ReverseProxyPage) SetFocused(focused bool) {
	p.focused = focused
	p.domains.focused = focused
	p.events.focused = focused
}

// Init loads data if not already cached.
func (p *ReverseProxyPage) Init(c *client.Client) tea.Cmd {
	if len(p.services) > 0 {
		return nil
	}
	p.loading = true
	p.err = nil
	return FetchReverseProxyData(c)
}

// Update is the tea.Update entry point for the page.
func (p *ReverseProxyPage) Update(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	// Sub-view messages first so late-arriving async responses land correctly.
	switch p.state {
	case rpOuterDomains:
		cmd, close := p.domains.update(msg, c)
		if close {
			p.state = rpOuterDetail
			return p, nil
		}
		return p, cmd
	case rpOuterEvents:
		cmd, close := p.events.update(msg, c)
		if close {
			p.state = rpOuterDetail
			return p, nil
		}
		return p, cmd
	case rpOuterWizard:
		return p.updateWizard(msg, c)
	}

	// Top-level messages.
	switch msg := msg.(type) {
	case PageRefreshTickMsg:
		if p.state == rpOuterList && p.wiz.form == nil {
			p.loading = true
			return p, FetchReverseProxyData(c)
		}
		return p, nil

	case ReverseProxiesLoadedMsg:
		p.loading = false
		if msg.Err != nil {
			p.err = msg.Err
			return p, nil
		}
		p.services = msg.Services
		p.clusters = msg.Clusters
		p.groups = msg.Groups
		p.peers = msg.Peers
		p.applyFilter()
		return p, nil

	case ReverseProxyUpdatedMsg:
		if msg.Err != nil {
			p.err = msg.Err
			return p, nil
		}
		p.loading = true
		return p, FetchReverseProxyData(c)

	case ToastMsg:
		p.loading = true
		return p, FetchReverseProxyData(c)

	case APIErrorMsg:
		p.err = msg.Err
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg, c)
	}

	return p, nil
}

// View dispatches to the appropriate renderer based on state.
func (p *ReverseProxyPage) View(width, height int) string {
	if p.loading && len(p.services) == 0 {
		return loadingStyle.Render("  Loading reverse proxies...")
	}
	if p.err != nil {
		return errorStyle.Render("  Error: " + p.err.Error())
	}

	switch p.state {
	case rpOuterDetail:
		svc, ok := p.currentService()
		if !ok {
			p.state = rpOuterList
			return p.viewList(width, height)
		}
		return renderRPDetail(svc, p.detailTab, p.peers, p.groups)
	case rpOuterDomains:
		return p.domains.view(width)
	case rpOuterEvents:
		return p.events.view(width)
	case rpOuterWizard:
		return p.viewWizard(width)
	}
	return p.viewList(width, height)
}

// applyFilter filters services by p.search and clamps cursor.
func (p *ReverseProxyPage) applyFilter() {
	if p.search == "" {
		p.filtered = p.services
	} else {
		filtered := make([]models.ReverseProxyService, 0)
		for _, s := range p.services {
			if matchesQuery(p.search, s.Name, s.Domain, s.Mode, s.ProxyCluster) {
				filtered = append(filtered, s)
			}
		}
		p.filtered = filtered
	}
	p.cursor = clampCursor(p.cursor, len(p.filtered))
}

// currentService returns the service at the list cursor, if any.
func (p *ReverseProxyPage) currentService() (models.ReverseProxyService, bool) {
	if p.cursor < 0 || p.cursor >= len(p.filtered) {
		return models.ReverseProxyService{}, false
	}
	return p.filtered[p.cursor], true
}

// handleKey handles keys at the list + detail levels. Wizard keys are handled in updateWizard.
func (p *ReverseProxyPage) handleKey(msg tea.KeyPressMsg, c *client.Client) (Page, tea.Cmd) {
	key := msg.String()

	if p.state == rpOuterDetail {
		switch key {
		case "esc", "backspace", "q":
			p.state = rpOuterList
		case "left", "h":
			p.detailTab = p.detailTab.Prev()
		case "right", "l":
			p.detailTab = p.detailTab.Next()
		case "e":
			p.enterWizardForEdit()
			if p.wiz.form == nil {
				return p, nil
			}
			return p, p.wiz.form.Init()
		case "t":
			if svc, ok := p.currentService(); ok {
				return p, ToggleReverseProxy(c, svc, !svc.Enabled)
			}
		case "d":
			if svc, ok := p.currentService(); ok {
				return p, DeleteReverseProxy(c, svc.ID)
			}
		case "r":
			p.loading = true
			return p, FetchReverseProxyData(c)
		case "D":
			svc, ok := p.currentService()
			if !ok {
				return p, nil
			}
			p.domains.setServiceID(svc.ID)
			p.state = rpOuterDomains
			return p, p.domains.load(c)
		case "E":
			svc, ok := p.currentService()
			if !ok {
				return p, nil
			}
			p.events.setServiceID(svc.ID)
			p.state = rpOuterEvents
			return p, p.events.load(c)
		}
		return p, nil
	}

	// List view
	if p.searching {
		newSearch, still, changed := handleSearchKey(key, p.search)
		p.search = newSearch
		p.searching = still
		if changed || !still {
			p.applyFilter()
		}
		return p, nil
	}

	switch key {
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
		}
	case "down", "j":
		if p.cursor < len(p.filtered)-1 {
			p.cursor++
		}
	case "enter":
		if len(p.filtered) > 0 {
			p.state = rpOuterDetail
			p.detailTab = rpDetailService
		}
	case "c":
		p.enterWizardForCreate()
		return p, p.wiz.form.Init()
	case "t":
		if svc, ok := p.currentService(); ok {
			return p, ToggleReverseProxy(c, svc, !svc.Enabled)
		}
	case "d":
		if svc, ok := p.currentService(); ok {
			return p, DeleteReverseProxy(c, svc.ID)
		}
	case "r":
		p.loading = true
		return p, FetchReverseProxyData(c)
	case "/":
		p.searching = true
		p.search = ""
	case "esc":
		if p.search != "" {
			p.search = ""
			p.cursor = 0
			p.applyFilter()
		}
	}
	return p, nil
}

// viewList renders the list of services.
func (p *ReverseProxyPage) viewList(width, height int) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(fmt.Sprintf("Reverse Proxy (%d)", len(p.filtered))) + "\n")

	if p.searching {
		b.WriteString(sectionHeaderStyle.Render("  / ") + detailValueStyle.Render(p.search) + "\u2588\n")
	} else if p.search != "" {
		b.WriteString(dimHintStyle.Render(fmt.Sprintf("  filter: %s", p.search)) + "\n")
	}

	if len(p.filtered) == 0 {
		b.WriteString("\n" + dimHintStyle.Render("  No reverse-proxy services. Press 'c' to create one.") + "\n")
		b.WriteString(dimHintStyle.Render("  /: search  c: create  r: refresh"))
		return b.String()
	}

	maxRows := height - 5
	if maxRows < 1 {
		maxRows = 1
	}
	offset := 0
	if p.cursor >= maxRows {
		offset = p.cursor - maxRows + 1
	}
	end := offset + maxRows
	if end > len(p.filtered) {
		end = len(p.filtered)
	}

	rows := make([][]string, 0, end-offset)
	for i := offset; i < end; i++ {
		s := p.filtered[i]
		status := "-"
		if s.Meta != nil {
			status = s.Meta.Status
		}
		enabled := boolBadge(s.Enabled)
		auth := rpAuthSummary(s.Auth)
		access := rpAccessSummary(s.AccessRestrictions)
		cluster := defaultStr(s.ProxyCluster, "-")

		rows = append(rows, []string{
			s.Domain,
			strings.ToUpper(s.Mode),
			status,
			enabled,
			fmt.Sprintf("%d", len(s.Targets)),
			auth,
			access,
			cluster,
		})
	}

	tw := width
	if tw > 140 {
		tw = 140
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers("DOMAIN", "MODE", "STATUS", "ON", "TGTS", "AUTH", "ACCESS", "CLUSTER").
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

			if dataIdx == p.cursor && p.focused {
				return tableSelectedStyle
			}

			if col == 3 && row >= 0 && row < len(rows) {
				if rows[row][3] == "Yes" {
					return base.Foreground(colorSuccess)
				}
				return base.Foreground(colorDanger)
			}

			if col == 2 && row >= 0 && row < len(rows) {
				switch rows[row][2] {
				case models.ReverseProxyStatusActive:
					return base.Foreground(colorSuccess)
				case models.ReverseProxyStatusCertificateFailed, models.ReverseProxyStatusError, models.ReverseProxyStatusTunnelNotCreated:
					return base.Foreground(colorDanger)
				case models.ReverseProxyStatusPending, models.ReverseProxyStatusCertificatePending:
					return base.Foreground(colorYellow)
				}
			}

			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render(fmt.Sprintf("  %d/%d  /: search  c: create  t: toggle  d: delete  enter: detail  r: refresh", p.cursor+1, len(p.filtered))))
	return b.String()
}

// ─── Summary helpers ────────────────────────────────────────────────

// rpAuthSummary describes the auth stack in a short list-cell string.
func rpAuthSummary(a *models.ReverseProxyAuth) string {
	if a == nil {
		return "-"
	}
	methods := []string{}
	if a.PasswordAuth != nil && a.PasswordAuth.Enabled {
		methods = append(methods, "pwd")
	}
	if a.PinAuth != nil && a.PinAuth.Enabled {
		methods = append(methods, "pin")
	}
	if a.BearerAuth != nil && a.BearerAuth.Enabled {
		methods = append(methods, "sso")
	}
	if a.LinkAuth != nil && a.LinkAuth.Enabled {
		methods = append(methods, "link")
	}
	if n := len(a.HeaderAuths); n > 0 {
		methods = append(methods, fmt.Sprintf("hdr×%d", n))
	}
	if len(methods) == 0 {
		return "-"
	}
	return strings.Join(methods, ",")
}

// rpAccessSummary summarises an access-restriction block into a short cell.
func rpAccessSummary(r *models.ReverseProxyAccessRestrictions) string {
	if r == nil {
		return "-"
	}
	total := len(r.AllowedCIDRs) + len(r.BlockedCIDRs) + len(r.AllowedCountries) + len(r.BlockedCountries)
	if total == 0 {
		return "-"
	}
	return fmt.Sprintf("%d rule%s", total, pluralS(total))
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
