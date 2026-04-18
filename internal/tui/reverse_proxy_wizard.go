package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"netbird-manage/internal/client"
	"netbird-manage/internal/models"
)

// rpWizardStep enumerates every screen inside the create/edit wizard.
type rpWizardStep int

const (
	wizStepService rpWizardStep = iota

	wizStepTargets
	wizStepTargetEdit
	wizStepTargetHeaders
	wizStepTargetHeaderEdit

	wizStepAuth
	wizStepAuthPwd
	wizStepAuthPin
	wizStepAuthBearer
	wizStepAuthLink
	wizStepAuthHeaders
	wizStepAuthHeaderEdit

	wizStepAccess
	wizStepAccessRule

	wizStepAdvanced
	wizStepConfirm
	wizStepUnprotectedWarning
)

// ─── Wizard entry points ────────────────────────────────────────────

func (p *ReverseProxyPage) enterWizardForCreate() {
	p.wiz = rpWizard{
		editing:      false,
		svc:          models.ReverseProxyService{Mode: models.ReverseProxyModeHTTP, Enabled: true},
		step:         wizStepService,
		editingIndex: -1,
	}
	p.wiz.serviceData = rpServiceFormData{
		mode:    models.ReverseProxyModeHTTP,
		enabled: true,
	}
	p.wiz.form = newRPServiceForm(&p.wiz.serviceData, p.clusters, false)
	p.state = rpOuterWizard
}

func (p *ReverseProxyPage) enterWizardForEdit() {
	svc, ok := p.currentService()
	if !ok {
		return
	}
	p.wiz = rpWizard{
		editing:      true,
		editingID:    svc.ID,
		svc:          cloneService(svc),
		step:         wizStepService,
		editingIndex: -1,
	}
	p.wiz.serviceData = serviceFormFromSvc(p.wiz.svc)
	p.wiz.form = newRPServiceForm(&p.wiz.serviceData, p.clusters, true)
	p.state = rpOuterWizard
}

// ─── Update dispatch ────────────────────────────────────────────────

// updateWizard owns all wizard state transitions.
func (p *ReverseProxyPage) updateWizard(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	switch m := msg.(type) {
	case ReverseProxyUpdatedMsg:
		if m.Err != nil {
			// Submit failed — stay on the confirm screen so the user can retry
			// or abort. Surface the error in the wizard banner.
			p.wiz.err = m.Err
			p.wiz.step = wizStepConfirm
			return p, nil
		}
		// Success — bail to the list and refresh.
		p.state = rpOuterList
		p.wiz = rpWizard{}
		p.loading = true
		return p, FetchReverseProxyData(c)
	}

	// Confirm / warning screens are y/n only.
	if p.wiz.step == wizStepUnprotectedWarning {
		return p.handleUnprotectedWarningKey(msg, c)
	}
	if p.wiz.step == wizStepConfirm {
		return p.handleConfirmKey(msg, c)
	}

	// List-editor screens (no huh form).
	if isListEditorStep(p.wiz.step) {
		return p.updateListEditor(msg, c)
	}

	// Huh form screens.
	if p.wiz.form == nil {
		return p, nil
	}
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
		p.popWizardStep()
		return p, nil
	}

	m, cmd := p.wiz.form.Update(msg)
	if f, ok := m.(*huh.Form); ok {
		p.wiz.form = f
	}

	switch p.wiz.form.State {
	case huh.StateCompleted:
		return p.handleFormComplete(c), cmd
	case huh.StateAborted:
		p.popWizardStep()
		return p, cmd
	}
	return p, cmd
}

// isListEditorStep reports whether the given step renders a list-editor view (no huh form).
func isListEditorStep(s rpWizardStep) bool {
	switch s {
	case wizStepTargets, wizStepTargetHeaders, wizStepAuth, wizStepAuthHeaders, wizStepAccess:
		return true
	}
	return false
}

// ─── Form completion (per-step) ─────────────────────────────────────

func (p *ReverseProxyPage) handleFormComplete(c *client.Client) Page {
	p.wiz.err = nil
	switch p.wiz.step {
	case wizStepService:
		applyServiceForm(&p.wiz.svc, p.wiz.serviceData)
		p.wiz.form = nil
		p.gotoTargets()
	case wizStepTargetEdit:
		applyTargetForm(&p.wiz.svc, p.wiz.editingIndex, p.wiz.targetData)
		p.wiz.form = nil
		p.wiz.step = wizStepTargets
	case wizStepTargetHeaderEdit:
		applyCustomHeaderForm(&p.wiz.svc, p.wiz.customHeadersForTarget, p.wiz.editingIndex, p.wiz.customHeaderData)
		p.wiz.form = nil
		p.wiz.step = wizStepTargetHeaders
	case wizStepAuthPwd:
		applyPasswordForm(&p.wiz.svc, p.wiz.passwordData)
		p.wiz.form = nil
		p.wiz.step = wizStepAuth
	case wizStepAuthPin:
		applyPinForm(&p.wiz.svc, p.wiz.pinData)
		p.wiz.form = nil
		p.wiz.step = wizStepAuth
	case wizStepAuthBearer:
		applyBearerForm(&p.wiz.svc, p.wiz.bearerData)
		p.wiz.form = nil
		p.wiz.step = wizStepAuth
	case wizStepAuthLink:
		applyLinkForm(&p.wiz.svc, p.wiz.linkData)
		p.wiz.form = nil
		p.wiz.step = wizStepAuth
	case wizStepAuthHeaderEdit:
		applyHeaderAuthForm(&p.wiz.svc, p.wiz.editingIndex, p.wiz.headerAuthData)
		p.wiz.form = nil
		p.wiz.step = wizStepAuthHeaders
	case wizStepAccessRule:
		applyAccessRuleForm(&p.wiz.svc, p.wiz.editingIndex, p.wiz.accessData)
		p.wiz.form = nil
		p.wiz.step = wizStepAccess
	case wizStepAdvanced:
		applyAdvancedForm(&p.wiz.svc, p.wiz.advancedData)
		p.wiz.form = nil
		p.gotoConfirm()
	}
	return p
}

// popWizardStep handles "esc" on huh-form screens — usually going back to the parent list.
func (p *ReverseProxyPage) popWizardStep() {
	p.wiz.err = nil
	switch p.wiz.step {
	case wizStepService:
		p.state = rpOuterList
		p.wiz = rpWizard{}
	case wizStepTargetEdit:
		p.wiz.form = nil
		p.wiz.step = wizStepTargets
	case wizStepTargetHeaderEdit:
		p.wiz.form = nil
		p.wiz.step = wizStepTargetHeaders
	case wizStepAuthPwd, wizStepAuthPin, wizStepAuthBearer, wizStepAuthLink:
		p.wiz.form = nil
		p.wiz.step = wizStepAuth
	case wizStepAuthHeaderEdit:
		p.wiz.form = nil
		p.wiz.step = wizStepAuthHeaders
	case wizStepAccessRule:
		p.wiz.form = nil
		p.wiz.step = wizStepAccess
	case wizStepAdvanced:
		p.wiz.form = nil
		// Back to Access (or Auth/Targets if L4, but Access still exists for all)
		p.wiz.step = wizStepAccess
	}
}

// ─── Forward navigation ─────────────────────────────────────────────

func (p *ReverseProxyPage) gotoTargets() {
	p.wiz.err = nil
	p.wiz.step = wizStepTargets
	p.wiz.targetsCursor = 0
}

func (p *ReverseProxyPage) gotoAuth() {
	p.wiz.err = nil
	p.wiz.step = wizStepAuth
	p.wiz.authCursor = 0
}

func (p *ReverseProxyPage) gotoAccess() {
	p.wiz.err = nil
	p.wiz.step = wizStepAccess
	p.wiz.accessCursor = 0
}

func (p *ReverseProxyPage) gotoAdvanced() {
	p.wiz.err = nil
	p.wiz.step = wizStepAdvanced
	p.wiz.advancedData = advancedFormFromSvc(p.wiz.svc)
	p.wiz.form = newRPAdvancedForm(&p.wiz.advancedData, p.wiz.svc.Mode)
}

func (p *ReverseProxyPage) gotoConfirm() {
	p.wiz.err = nil
	// Protected-warning gate: HTTP + no auth + no access rules.
	if p.wiz.svc.Mode == models.ReverseProxyModeHTTP && !svcHasAnyAuth(p.wiz.svc.Auth) && !svcHasAnyAccess(p.wiz.svc.AccessRestrictions) {
		p.wiz.step = wizStepUnprotectedWarning
		return
	}
	p.wiz.step = wizStepConfirm
}

// gotoNextAfterTargets picks Auth if HTTP, or Access directly for L4.
func (p *ReverseProxyPage) gotoNextAfterTargets() {
	if p.wiz.svc.Mode == models.ReverseProxyModeHTTP {
		p.gotoAuth()
		return
	}
	p.gotoAccess()
}

// ─── List-editor updates ────────────────────────────────────────────

func (p *ReverseProxyPage) updateListEditor(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	key := keyMsg.String()

	switch p.wiz.step {
	case wizStepTargets:
		return p.updateTargetsList(key)
	case wizStepTargetHeaders:
		return p.updateTargetHeadersList(key)
	case wizStepAuth:
		return p.updateAuthList(key)
	case wizStepAuthHeaders:
		return p.updateAuthHeadersList(key)
	case wizStepAccess:
		return p.updateAccessList(key)
	}
	return p, nil
}

// Targets list: each row is one target. Synthetic last row = "Add target".
func (p *ReverseProxyPage) updateTargetsList(key string) (Page, tea.Cmd) {
	n := len(p.wiz.svc.Targets)
	switch key {
	case "esc":
		p.wiz.step = wizStepService
		p.wiz.serviceData = serviceFormFromSvc(p.wiz.svc)
		p.wiz.form = newRPServiceForm(&p.wiz.serviceData, p.clusters, p.wiz.editing)
		return p, p.wiz.form.Init()
	case "up", "k":
		if p.wiz.targetsCursor > 0 {
			p.wiz.targetsCursor--
		}
	case "down", "j":
		if p.wiz.targetsCursor < n {
			p.wiz.targetsCursor++
		}
	case "enter":
		if p.wiz.targetsCursor == n {
			p.startTargetEdit(-1)
			return p, p.wiz.form.Init()
		}
		p.startTargetEdit(p.wiz.targetsCursor)
		return p, p.wiz.form.Init()
	case "h":
		// Drill into custom headers for the selected target (HTTP only).
		if p.wiz.svc.Mode != models.ReverseProxyModeHTTP {
			return p, nil
		}
		if p.wiz.targetsCursor >= n {
			return p, nil
		}
		p.wiz.customHeadersForTarget = p.wiz.targetsCursor
		p.wiz.customHeadersCursor = 0
		p.wiz.step = wizStepTargetHeaders
	case "d":
		if p.wiz.targetsCursor < n && n > 0 {
			p.wiz.svc.Targets = deleteTargetAt(p.wiz.svc.Targets, p.wiz.targetsCursor)
			if p.wiz.targetsCursor >= len(p.wiz.svc.Targets) {
				p.wiz.targetsCursor = len(p.wiz.svc.Targets)
			}
		}
	case "c":
		// Continue to the next step.
		if !p.validateTargets() {
			return p, nil
		}
		p.gotoNextAfterTargets()
	}
	return p, nil
}

// validateTargets gates forward navigation.
func (p *ReverseProxyPage) validateTargets() bool {
	if len(p.wiz.svc.Targets) == 0 {
		p.wiz.err = fmt.Errorf("at least one target is required")
		return false
	}
	// L4 modes: exactly one target allowed.
	if p.wiz.svc.Mode != models.ReverseProxyModeHTTP && len(p.wiz.svc.Targets) > 1 {
		p.wiz.err = fmt.Errorf("mode %s supports exactly one target", p.wiz.svc.Mode)
		return false
	}
	p.wiz.err = nil
	return true
}

func (p *ReverseProxyPage) startTargetEdit(idx int) {
	p.wiz.editingIndex = idx
	if idx == -1 {
		p.wiz.targetData = rpTargetFormData{
			targetType: "peer",
			protocol:   defaultProtocolForMode(p.wiz.svc.Mode),
			enabled:    true,
		}
	} else {
		p.wiz.targetData = targetFormFromTarget(p.wiz.svc.Targets[idx])
	}
	p.wiz.step = wizStepTargetEdit
	p.wiz.form = newRPTargetForm(&p.wiz.targetData, p.peers, p.wiz.svc.Mode)
}

// Target headers list: edit the custom_headers map of a single target.
func (p *ReverseProxyPage) updateTargetHeadersList(key string) (Page, tea.Cmd) {
	target := &p.wiz.svc.Targets[p.wiz.customHeadersForTarget]
	if target.Options == nil {
		target.Options = &models.ReverseProxyTargetOpts{}
	}
	keys := sortedKeys(target.Options.CustomHeaders)
	n := len(keys)
	switch key {
	case "esc":
		p.wiz.step = wizStepTargets
	case "up", "k":
		if p.wiz.customHeadersCursor > 0 {
			p.wiz.customHeadersCursor--
		}
	case "down", "j":
		if p.wiz.customHeadersCursor < n {
			p.wiz.customHeadersCursor++
		}
	case "enter":
		if p.wiz.customHeadersCursor == n {
			p.wiz.customHeaderData = rpCustomHeaderFormData{}
			p.wiz.editingIndex = -1
		} else {
			k := keys[p.wiz.customHeadersCursor]
			p.wiz.customHeaderData = rpCustomHeaderFormData{header: k, value: target.Options.CustomHeaders[k]}
			p.wiz.editingIndex = p.wiz.customHeadersCursor
		}
		p.wiz.step = wizStepTargetHeaderEdit
		p.wiz.form = newRPCustomHeaderForm(&p.wiz.customHeaderData)
		return p, p.wiz.form.Init()
	case "d":
		if p.wiz.customHeadersCursor < n {
			k := keys[p.wiz.customHeadersCursor]
			delete(target.Options.CustomHeaders, k)
			if p.wiz.customHeadersCursor > 0 && p.wiz.customHeadersCursor >= len(target.Options.CustomHeaders) {
				p.wiz.customHeadersCursor--
			}
		}
	}
	return p, nil
}

// Auth list: 4 auth methods + the "header auths" sub-screen.
var rpAuthRows = []string{"Password", "PIN", "SSO (Bearer)", "Magic Link", "Header Rules"}

func (p *ReverseProxyPage) updateAuthList(key string) (Page, tea.Cmd) {
	if p.wiz.svc.Mode != models.ReverseProxyModeHTTP {
		// L4 — auth not supported. Skip straight to Access.
		p.gotoAccess()
		return p, nil
	}
	n := len(rpAuthRows)
	switch key {
	case "esc":
		p.wiz.step = wizStepTargets
	case "up", "k":
		if p.wiz.authCursor > 0 {
			p.wiz.authCursor--
		}
	case "down", "j":
		if p.wiz.authCursor < n-1 {
			p.wiz.authCursor++
		}
	case "enter":
		p.openAuthSub(p.wiz.authCursor)
		if p.wiz.form != nil {
			return p, p.wiz.form.Init()
		}
	case "c":
		p.gotoAccess()
	}
	return p, nil
}

func (p *ReverseProxyPage) openAuthSub(row int) {
	ensureAuth(&p.wiz.svc)
	switch row {
	case 0: // Password
		p.wiz.passwordData = passwordFormFromAuth(p.wiz.svc.Auth)
		p.wiz.step = wizStepAuthPwd
		p.wiz.form = newRPPasswordForm(&p.wiz.passwordData)
	case 1: // PIN
		p.wiz.pinData = pinFormFromAuth(p.wiz.svc.Auth)
		p.wiz.step = wizStepAuthPin
		p.wiz.form = newRPPinForm(&p.wiz.pinData)
	case 2: // Bearer
		p.wiz.bearerData = bearerFormFromAuth(p.wiz.svc.Auth)
		p.wiz.step = wizStepAuthBearer
		p.wiz.form = newRPBearerForm(&p.wiz.bearerData, p.groups)
	case 3: // Link
		p.wiz.linkData = linkFormFromAuth(p.wiz.svc.Auth)
		p.wiz.step = wizStepAuthLink
		p.wiz.form = newRPLinkForm(&p.wiz.linkData)
	case 4: // Header auths
		p.wiz.step = wizStepAuthHeaders
		p.wiz.headerAuthsCursor = 0
	}
}

// Auth-headers list — a list editor for header_auth rules.
func (p *ReverseProxyPage) updateAuthHeadersList(key string) (Page, tea.Cmd) {
	ensureAuth(&p.wiz.svc)
	headers := p.wiz.svc.Auth.HeaderAuths
	n := len(headers)
	switch key {
	case "esc":
		p.wiz.step = wizStepAuth
	case "up", "k":
		if p.wiz.headerAuthsCursor > 0 {
			p.wiz.headerAuthsCursor--
		}
	case "down", "j":
		if p.wiz.headerAuthsCursor < n {
			p.wiz.headerAuthsCursor++
		}
	case "enter":
		if p.wiz.headerAuthsCursor == n {
			p.wiz.headerAuthData = rpHeaderAuthFormData{enabled: true}
			p.wiz.editingIndex = -1
		} else {
			h := headers[p.wiz.headerAuthsCursor]
			p.wiz.headerAuthData = rpHeaderAuthFormData{enabled: h.Enabled, header: h.Header, value: h.Value}
			p.wiz.editingIndex = p.wiz.headerAuthsCursor
		}
		p.wiz.step = wizStepAuthHeaderEdit
		p.wiz.form = newRPHeaderAuthForm(&p.wiz.headerAuthData)
		return p, p.wiz.form.Init()
	case "d":
		if p.wiz.headerAuthsCursor < n {
			p.wiz.svc.Auth.HeaderAuths = append(headers[:p.wiz.headerAuthsCursor], headers[p.wiz.headerAuthsCursor+1:]...)
			if p.wiz.headerAuthsCursor > 0 && p.wiz.headerAuthsCursor >= len(p.wiz.svc.Auth.HeaderAuths) {
				p.wiz.headerAuthsCursor--
			}
		}
	}
	return p, nil
}

// Access list.
func (p *ReverseProxyPage) updateAccessList(key string) (Page, tea.Cmd) {
	rows := accessRuleRows(p.wiz.svc.AccessRestrictions)
	n := len(rows)
	switch key {
	case "esc":
		// Go back to the upstream step — Auth if HTTP, Targets for L4.
		if p.wiz.svc.Mode == models.ReverseProxyModeHTTP {
			p.wiz.step = wizStepAuth
		} else {
			p.wiz.step = wizStepTargets
		}
	case "up", "k":
		if p.wiz.accessCursor > 0 {
			p.wiz.accessCursor--
		}
	case "down", "j":
		if p.wiz.accessCursor < n {
			p.wiz.accessCursor++
		}
	case "enter":
		if p.wiz.accessCursor == n {
			p.wiz.accessData = rpAccessRuleFormData{action: "allow", kind: "cidr"}
			p.wiz.editingIndex = -1
		} else {
			row := rows[p.wiz.accessCursor]
			p.wiz.accessData = rpAccessRuleFormData{action: row.cells[0], kind: row.cells[1], value: row.cells[2]}
			p.wiz.editingIndex = p.wiz.accessCursor
		}
		p.wiz.step = wizStepAccessRule
		p.wiz.form = newRPAccessRuleForm(&p.wiz.accessData)
		return p, p.wiz.form.Init()
	case "d":
		if p.wiz.accessCursor < n {
			p.wiz.svc.AccessRestrictions = removeAccessRuleAt(p.wiz.svc.AccessRestrictions, p.wiz.accessCursor)
			if p.wiz.accessCursor > 0 && p.wiz.accessCursor >= len(accessRuleRows(p.wiz.svc.AccessRestrictions)) {
				p.wiz.accessCursor--
			}
		}
	case "c":
		p.gotoAdvanced()
		return p, p.wiz.form.Init()
	}
	return p, nil
}

// ─── Confirm / warning screens ──────────────────────────────────────

func (p *ReverseProxyPage) handleUnprotectedWarningKey(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	switch keyMsg.String() {
	case "y":
		p.wiz.step = wizStepConfirm
	case "n", "esc":
		// Back to Advanced to let the user add protections.
		p.wiz.step = wizStepAccess
	}
	return p, nil
}

func (p *ReverseProxyPage) handleConfirmKey(msg tea.Msg, c *client.Client) (Page, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return p, nil
	}
	switch keyMsg.String() {
	case "y":
		return p, p.submit(c)
	case "n", "esc":
		p.state = rpOuterList
		p.wiz = rpWizard{}
	}
	return p, nil
}

func (p *ReverseProxyPage) submit(c *client.Client) tea.Cmd {
	if p.wiz.editing {
		return UpdateReverseProxy(c, p.wiz.editingID, ReverseProxyUpdateFromService(p.wiz.svc))
	}
	req := models.ReverseProxyCreateRequest{
		Name:               p.wiz.svc.Name,
		Domain:             p.wiz.svc.Domain,
		Mode:               p.wiz.svc.Mode,
		ListenPort:         p.wiz.svc.ListenPort,
		ProxyCluster:       p.wiz.svc.ProxyCluster,
		Targets:            p.wiz.svc.Targets,
		Enabled:            p.wiz.svc.Enabled,
		PassHostHeader:     p.wiz.svc.PassHostHeader,
		RewriteRedirects:   p.wiz.svc.RewriteRedirects,
		Auth:               p.wiz.svc.Auth,
		AccessRestrictions: p.wiz.svc.AccessRestrictions,
	}
	return CreateReverseProxy(c, req)
}

// ─── View dispatch ──────────────────────────────────────────────────

func (p *ReverseProxyPage) viewWizard(width int) string {
	title := p.wizardTitle()
	body := p.viewWizardBody(width, title)

	if p.wiz.err != nil {
		banner := errorStyle.Render("  " + p.wiz.err.Error() + "\n\n")
		return banner + body
	}
	return body
}

func (p *ReverseProxyPage) viewWizardBody(width int, title string) string {
	switch p.wiz.step {
	case wizStepService:
		return pageTitleStyle.Render(title) + "\n\n" + p.wiz.form.View()

	case wizStepTargets:
		rows := targetRows(p.wiz.svc.Targets, p.peers)
		p.clampCursor(&p.wiz.targetsCursor, len(rows))
		hints := "esc: back  enter: edit/add  h: headers  d: delete  c: continue"
		return renderListEditor(title, []string{"TARGET", "PROTO", "PORT", "PATH", "ON"}, rows, "[+ Add target]", p.wiz.targetsCursor, width, p.focused, hints)

	case wizStepTargetEdit:
		return pageTitleStyle.Render(title) + "\n\n" + p.wiz.form.View()

	case wizStepTargetHeaders:
		target := p.wiz.svc.Targets[p.wiz.customHeadersForTarget]
		var m map[string]string
		if target.Options != nil {
			m = target.Options.CustomHeaders
		}
		rows := customHeaderRows(m)
		p.clampCursor(&p.wiz.customHeadersCursor, len(rows))
		hints := "esc: back  enter: edit/add  d: delete"
		return renderListEditor(title, []string{"HEADER", "VALUE"}, rows, "[+ Add header]", p.wiz.customHeadersCursor, width, p.focused, hints)

	case wizStepTargetHeaderEdit:
		return pageTitleStyle.Render(title) + "\n\n" + p.wiz.form.View()

	case wizStepAuth:
		rows := p.authListRows()
		p.clampCursor(&p.wiz.authCursor, len(rows)-1) // no synthetic add row
		hints := "esc: back  enter: configure  c: continue"
		return renderListEditor(title, []string{"METHOD", "STATE", "DETAIL"}, rows, "", p.wiz.authCursor, width, p.focused, hints)

	case wizStepAuthPwd, wizStepAuthPin, wizStepAuthBearer, wizStepAuthLink:
		return pageTitleStyle.Render(title) + "\n\n" + p.wiz.form.View()

	case wizStepAuthHeaders:
		ensureAuth(&p.wiz.svc)
		rows := headerAuthRows(p.wiz.svc.Auth.HeaderAuths)
		p.clampCursor(&p.wiz.headerAuthsCursor, len(rows))
		hints := "esc: back  enter: edit/add  d: delete"
		return renderListEditor(title, []string{"HEADER", "VALUE", "ON"}, rows, "[+ Add header rule]", p.wiz.headerAuthsCursor, width, p.focused, hints)

	case wizStepAuthHeaderEdit:
		return pageTitleStyle.Render(title) + "\n\n" + p.wiz.form.View()

	case wizStepAccess:
		rows := accessRuleRows(p.wiz.svc.AccessRestrictions)
		p.clampCursor(&p.wiz.accessCursor, len(rows))
		hints := "esc: back  enter: edit/add  d: delete  c: continue"
		return renderListEditor(title, []string{"ACTION", "TYPE", "VALUE"}, rows, "[+ Add rule]", p.wiz.accessCursor, width, p.focused, hints)

	case wizStepAccessRule:
		return pageTitleStyle.Render(title) + "\n\n" + p.wiz.form.View()

	case wizStepAdvanced:
		return pageTitleStyle.Render(title) + "\n\n" + p.wiz.form.View()

	case wizStepConfirm:
		return RenderConfirm(title, confirmFieldsForSvc(p.wiz.svc, p.peers, p.groups))

	case wizStepUnprotectedWarning:
		return renderUnprotectedWarning()
	}
	return ""
}

// renderUnprotectedWarning is the "no auth + no access" confirmation gate.
func renderUnprotectedWarning() string {
	var b strings.Builder
	b.WriteString(pageTitleStyle.Render("Unprotected Service") + "\n\n")
	b.WriteString(detailValueStyle.Render("This HTTP service has no authentication and no access restrictions.") + "\n")
	b.WriteString(detailValueStyle.Render("Anyone who can reach the domain will be able to access it.") + "\n\n")
	b.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorDanger).
		Foreground(colorDanger).
		Padding(0, 2).
		Render("Continue anyway? Press y to accept, n or esc to go back and add protection"))
	return b.String()
}

// wizardTitle produces a breadcrumb like "Edit service ▸ Targets".
func (p *ReverseProxyPage) wizardTitle() string {
	prefix := "Create reverse proxy"
	if p.wiz.editing {
		prefix = "Edit reverse proxy"
	}
	step := wizardStepLabel(p.wiz.step)
	return prefix + " ▸ " + step
}

func wizardStepLabel(s rpWizardStep) string {
	switch s {
	case wizStepService:
		return "Service"
	case wizStepTargets:
		return "Targets"
	case wizStepTargetEdit:
		return "Target"
	case wizStepTargetHeaders:
		return "Target ▸ Custom Headers"
	case wizStepTargetHeaderEdit:
		return "Target ▸ Header"
	case wizStepAuth:
		return "Authentication"
	case wizStepAuthPwd:
		return "Auth ▸ Password"
	case wizStepAuthPin:
		return "Auth ▸ PIN"
	case wizStepAuthBearer:
		return "Auth ▸ SSO"
	case wizStepAuthLink:
		return "Auth ▸ Magic Link"
	case wizStepAuthHeaders:
		return "Auth ▸ Header Rules"
	case wizStepAuthHeaderEdit:
		return "Auth ▸ Header Rule"
	case wizStepAccess:
		return "Access Control"
	case wizStepAccessRule:
		return "Access Rule"
	case wizStepAdvanced:
		return "Advanced"
	case wizStepConfirm:
		return "Review & Submit"
	case wizStepUnprotectedWarning:
		return "Warning"
	}
	return ""
}

// authListRows produces the rows for the Auth screen.
func (p *ReverseProxyPage) authListRows() []listEditorRow {
	rows := make([]listEditorRow, 0, len(rpAuthRows))
	for i, label := range rpAuthRows {
		state := "off"
		detail := ""
		if p.wiz.svc.Auth != nil {
			switch i {
			case 0:
				if p.wiz.svc.Auth.PasswordAuth != nil && p.wiz.svc.Auth.PasswordAuth.Enabled {
					state = "on"
					detail = "configured"
				}
			case 1:
				if p.wiz.svc.Auth.PinAuth != nil && p.wiz.svc.Auth.PinAuth.Enabled {
					state = "on"
					detail = "configured"
				}
			case 2:
				if p.wiz.svc.Auth.BearerAuth != nil && p.wiz.svc.Auth.BearerAuth.Enabled {
					state = "on"
					detail = fmt.Sprintf("%d group(s)", len(p.wiz.svc.Auth.BearerAuth.DistributionGroups))
				}
			case 3:
				if p.wiz.svc.Auth.LinkAuth != nil && p.wiz.svc.Auth.LinkAuth.Enabled {
					state = "on"
				}
			case 4:
				if n := len(p.wiz.svc.Auth.HeaderAuths); n > 0 {
					state = "on"
					detail = fmt.Sprintf("%d rule(s)", n)
				}
			}
		}
		rows = append(rows, listEditorRow{cells: []string{label, state, detail}})
	}
	return rows
}

func (p *ReverseProxyPage) clampCursor(cursor *int, total int) {
	if total < 0 {
		total = 0
	}
	if *cursor < 0 {
		*cursor = 0
	}
	if *cursor > total {
		*cursor = total
	}
}

// ─── Form ↔ Model serialization ─────────────────────────────────────

func serviceFormFromSvc(svc models.ReverseProxyService) rpServiceFormData {
	port := ""
	if svc.ListenPort > 0 {
		port = strconv.Itoa(svc.ListenPort)
	}
	return rpServiceFormData{
		name:         svc.Name,
		domain:       svc.Domain,
		listenPort:   port,
		mode:         svc.Mode,
		proxyCluster: svc.ProxyCluster,
		enabled:      svc.Enabled,
	}
}

func applyServiceForm(svc *models.ReverseProxyService, data rpServiceFormData) {
	svc.Name = strings.TrimSpace(data.name)
	svc.Domain = strings.TrimSpace(data.domain)
	svc.Mode = data.mode
	svc.ProxyCluster = data.proxyCluster
	svc.Enabled = data.enabled
	if n, err := strconv.Atoi(strings.TrimSpace(data.listenPort)); err == nil {
		svc.ListenPort = n
	} else {
		svc.ListenPort = 0
	}
}

func targetFormFromTarget(t models.ReverseProxyTarget) rpTargetFormData {
	d := rpTargetFormData{
		targetType:  t.TargetType,
		targetID:    t.TargetID,
		host:        t.Host,
		protocol:    t.Protocol,
		port:        strconv.Itoa(t.Port),
		path:        t.Path,
		enabled:     t.Enabled,
		accessLocal: t.AccessLocal,
	}
	if t.Options != nil {
		d.skipTLSVerify = t.Options.SkipTLSVerify
		d.requestTimeout = t.Options.RequestTimeout
		d.sessionIdleTimeout = t.Options.SessionIdleTimeout
		d.proxyProtocol = t.Options.ProxyProtocol
	}
	return d
}

func applyTargetForm(svc *models.ReverseProxyService, idx int, data rpTargetFormData) {
	port, _ := strconv.Atoi(strings.TrimSpace(data.port))
	t := models.ReverseProxyTarget{
		TargetType:  data.targetType,
		Protocol:    data.protocol,
		Port:        port,
		Path:        strings.TrimSpace(data.path),
		Enabled:     data.enabled,
		AccessLocal: data.accessLocal,
	}
	if data.targetType == "peer" {
		t.TargetID = data.targetID
	} else {
		t.Host = strings.TrimSpace(data.host)
	}
	opts := &models.ReverseProxyTargetOpts{}
	if data.skipTLSVerify {
		opts.SkipTLSVerify = true
	}
	if data.requestTimeout != "" {
		opts.RequestTimeout = strings.TrimSpace(data.requestTimeout)
	}
	if data.sessionIdleTimeout != "" {
		opts.SessionIdleTimeout = strings.TrimSpace(data.sessionIdleTimeout)
	}
	if data.proxyProtocol {
		opts.ProxyProtocol = true
	}
	// Preserve existing custom_headers map if the target already had one.
	if idx >= 0 && idx < len(svc.Targets) && svc.Targets[idx].Options != nil {
		opts.CustomHeaders = svc.Targets[idx].Options.CustomHeaders
		opts.PathRewrite = svc.Targets[idx].Options.PathRewrite
	}
	if hasAnyOpts(opts) {
		t.Options = opts
	}
	if idx == -1 {
		svc.Targets = append(svc.Targets, t)
	} else {
		svc.Targets[idx] = t
	}
}

func applyCustomHeaderForm(svc *models.ReverseProxyService, targetIdx, entryIdx int, data rpCustomHeaderFormData) {
	if targetIdx < 0 || targetIdx >= len(svc.Targets) {
		return
	}
	if svc.Targets[targetIdx].Options == nil {
		svc.Targets[targetIdx].Options = &models.ReverseProxyTargetOpts{}
	}
	opts := svc.Targets[targetIdx].Options
	if opts.CustomHeaders == nil {
		opts.CustomHeaders = map[string]string{}
	}
	// If the user edited an existing row, the header name may have changed —
	// remove the old entry first.
	if entryIdx >= 0 {
		oldKeys := sortedKeys(opts.CustomHeaders)
		if entryIdx < len(oldKeys) {
			oldK := oldKeys[entryIdx]
			if oldK != strings.TrimSpace(data.header) {
				delete(opts.CustomHeaders, oldK)
			}
		}
	}
	opts.CustomHeaders[strings.TrimSpace(data.header)] = data.value
}

func passwordFormFromAuth(a *models.ReverseProxyAuth) rpPasswordFormData {
	if a == nil || a.PasswordAuth == nil {
		return rpPasswordFormData{}
	}
	return rpPasswordFormData{enabled: a.PasswordAuth.Enabled, password: a.PasswordAuth.Password}
}

func applyPasswordForm(svc *models.ReverseProxyService, d rpPasswordFormData) {
	ensureAuth(svc)
	svc.Auth.PasswordAuth = &models.ReverseProxyPasswordAuth{Enabled: d.enabled, Password: d.password}
}

func pinFormFromAuth(a *models.ReverseProxyAuth) rpPinFormData {
	if a == nil || a.PinAuth == nil {
		return rpPinFormData{}
	}
	return rpPinFormData{enabled: a.PinAuth.Enabled, pin: a.PinAuth.Pin}
}

func applyPinForm(svc *models.ReverseProxyService, d rpPinFormData) {
	ensureAuth(svc)
	svc.Auth.PinAuth = &models.ReverseProxyPinAuth{Enabled: d.enabled, Pin: d.pin}
}

func bearerFormFromAuth(a *models.ReverseProxyAuth) rpBearerFormData {
	if a == nil || a.BearerAuth == nil {
		return rpBearerFormData{}
	}
	return rpBearerFormData{enabled: a.BearerAuth.Enabled, distributionGroups: a.BearerAuth.DistributionGroups}
}

func applyBearerForm(svc *models.ReverseProxyService, d rpBearerFormData) {
	ensureAuth(svc)
	svc.Auth.BearerAuth = &models.ReverseProxyBearerAuth{Enabled: d.enabled, DistributionGroups: d.distributionGroups}
}

func linkFormFromAuth(a *models.ReverseProxyAuth) rpLinkFormData {
	if a == nil || a.LinkAuth == nil {
		return rpLinkFormData{}
	}
	return rpLinkFormData{enabled: a.LinkAuth.Enabled}
}

func applyLinkForm(svc *models.ReverseProxyService, d rpLinkFormData) {
	ensureAuth(svc)
	svc.Auth.LinkAuth = &models.ReverseProxyLinkAuth{Enabled: d.enabled}
}

func applyHeaderAuthForm(svc *models.ReverseProxyService, idx int, d rpHeaderAuthFormData) {
	ensureAuth(svc)
	entry := models.ReverseProxyHeaderAuth{Enabled: d.enabled, Header: strings.TrimSpace(d.header), Value: d.value}
	if idx == -1 {
		svc.Auth.HeaderAuths = append(svc.Auth.HeaderAuths, entry)
	} else if idx < len(svc.Auth.HeaderAuths) {
		svc.Auth.HeaderAuths[idx] = entry
	}
}

func applyAccessRuleForm(svc *models.ReverseProxyService, idx int, d rpAccessRuleFormData) {
	if svc.AccessRestrictions == nil {
		svc.AccessRestrictions = &models.ReverseProxyAccessRestrictions{}
	}
	value := strings.TrimSpace(d.value)
	if value == "" {
		if idx >= 0 {
			svc.AccessRestrictions = removeAccessRuleAt(svc.AccessRestrictions, idx)
		}
		return
	}
	if d.kind == "ip" {
		value = value + "/32"
	}
	newBucket := accessBucketFor(d.action, d.kind)

	// When editing, check whether the destination bucket matches the rule's
	// current bucket. If so, replace in place so the row doesn't jump.
	if idx >= 0 {
		curBucket, curLocal, ok := accessBucketOfFlattened(svc.AccessRestrictions, idx)
		if ok && curBucket == newBucket {
			replaceAccessRuleAt(svc.AccessRestrictions, curBucket, curLocal, value)
			return
		}
		// Bucket changed — remove old, fall through to append in new bucket.
		svc.AccessRestrictions = removeAccessRuleAt(svc.AccessRestrictions, idx)
	}
	appendAccessRule(svc.AccessRestrictions, newBucket, value)
}

// accessBucket enumerates which slice an access rule lives in.
type accessBucket int

const (
	bucketAllowCIDR accessBucket = iota
	bucketBlockCIDR
	bucketAllowCountry
	bucketBlockCountry
)

func accessBucketFor(action, kind string) accessBucket {
	switch {
	case kind == "country" && action == "allow":
		return bucketAllowCountry
	case kind == "country" && action == "block":
		return bucketBlockCountry
	case action == "block":
		return bucketBlockCIDR
	default:
		return bucketAllowCIDR
	}
}

// accessBucketOfFlattened returns which bucket (and bucket-local index) a
// flattened access-rule index points into.
func accessBucketOfFlattened(r *models.ReverseProxyAccessRestrictions, idx int) (accessBucket, int, bool) {
	if r == nil || idx < 0 {
		return 0, 0, false
	}
	i := idx
	if i < len(r.AllowedCIDRs) {
		return bucketAllowCIDR, i, true
	}
	i -= len(r.AllowedCIDRs)
	if i < len(r.BlockedCIDRs) {
		return bucketBlockCIDR, i, true
	}
	i -= len(r.BlockedCIDRs)
	if i < len(r.AllowedCountries) {
		return bucketAllowCountry, i, true
	}
	i -= len(r.AllowedCountries)
	if i < len(r.BlockedCountries) {
		return bucketBlockCountry, i, true
	}
	return 0, 0, false
}

func replaceAccessRuleAt(r *models.ReverseProxyAccessRestrictions, bucket accessBucket, idx int, value string) {
	switch bucket {
	case bucketAllowCIDR:
		r.AllowedCIDRs[idx] = value
	case bucketBlockCIDR:
		r.BlockedCIDRs[idx] = value
	case bucketAllowCountry:
		r.AllowedCountries[idx] = value
	case bucketBlockCountry:
		r.BlockedCountries[idx] = value
	}
}

func appendAccessRule(r *models.ReverseProxyAccessRestrictions, bucket accessBucket, value string) {
	switch bucket {
	case bucketAllowCIDR:
		r.AllowedCIDRs = append(r.AllowedCIDRs, value)
	case bucketBlockCIDR:
		r.BlockedCIDRs = append(r.BlockedCIDRs, value)
	case bucketAllowCountry:
		r.AllowedCountries = append(r.AllowedCountries, value)
	case bucketBlockCountry:
		r.BlockedCountries = append(r.BlockedCountries, value)
	}
}

func removeAccessRuleAt(r *models.ReverseProxyAccessRestrictions, idx int) *models.ReverseProxyAccessRestrictions {
	if r == nil {
		return nil
	}
	i := idx
	if i < len(r.AllowedCIDRs) {
		r.AllowedCIDRs = append(r.AllowedCIDRs[:i], r.AllowedCIDRs[i+1:]...)
		return r
	}
	i -= len(r.AllowedCIDRs)
	if i < len(r.BlockedCIDRs) {
		r.BlockedCIDRs = append(r.BlockedCIDRs[:i], r.BlockedCIDRs[i+1:]...)
		return r
	}
	i -= len(r.BlockedCIDRs)
	if i < len(r.AllowedCountries) {
		r.AllowedCountries = append(r.AllowedCountries[:i], r.AllowedCountries[i+1:]...)
		return r
	}
	i -= len(r.AllowedCountries)
	if i < len(r.BlockedCountries) {
		r.BlockedCountries = append(r.BlockedCountries[:i], r.BlockedCountries[i+1:]...)
		return r
	}
	return r
}

func advancedFormFromSvc(svc models.ReverseProxyService) rpAdvancedFormData {
	d := rpAdvancedFormData{
		passHostHeader:   svc.PassHostHeader,
		rewriteRedirects: svc.RewriteRedirects,
	}
	if len(svc.Targets) > 0 && svc.Targets[0].Options != nil {
		d.proxyProtocol = svc.Targets[0].Options.ProxyProtocol
		d.requestTimeout = svc.Targets[0].Options.RequestTimeout
		d.sessionIdleTimeout = svc.Targets[0].Options.SessionIdleTimeout
	}
	return d
}

func applyAdvancedForm(svc *models.ReverseProxyService, d rpAdvancedFormData) {
	switch svc.Mode {
	case models.ReverseProxyModeHTTP:
		svc.PassHostHeader = d.passHostHeader
		svc.RewriteRedirects = d.rewriteRedirects
	case models.ReverseProxyModeTCP, models.ReverseProxyModeTLS:
		if len(svc.Targets) > 0 {
			if svc.Targets[0].Options == nil {
				svc.Targets[0].Options = &models.ReverseProxyTargetOpts{}
			}
			svc.Targets[0].Options.ProxyProtocol = d.proxyProtocol
			svc.Targets[0].Options.RequestTimeout = strings.TrimSpace(d.requestTimeout)
		}
	case models.ReverseProxyModeUDP:
		if len(svc.Targets) > 0 {
			if svc.Targets[0].Options == nil {
				svc.Targets[0].Options = &models.ReverseProxyTargetOpts{}
			}
			svc.Targets[0].Options.SessionIdleTimeout = strings.TrimSpace(d.sessionIdleTimeout)
		}
	}
}

// ─── Small helpers ──────────────────────────────────────────────────

func ensureAuth(svc *models.ReverseProxyService) {
	if svc.Auth == nil {
		svc.Auth = &models.ReverseProxyAuth{}
	}
}

func svcHasAnyAuth(a *models.ReverseProxyAuth) bool {
	if a == nil {
		return false
	}
	if a.PasswordAuth != nil && a.PasswordAuth.Enabled {
		return true
	}
	if a.PinAuth != nil && a.PinAuth.Enabled {
		return true
	}
	if a.BearerAuth != nil && a.BearerAuth.Enabled {
		return true
	}
	if a.LinkAuth != nil && a.LinkAuth.Enabled {
		return true
	}
	for _, h := range a.HeaderAuths {
		if h.Enabled {
			return true
		}
	}
	return false
}

func svcHasAnyAccess(r *models.ReverseProxyAccessRestrictions) bool {
	if r == nil {
		return false
	}
	return len(r.AllowedCIDRs)+len(r.BlockedCIDRs)+len(r.AllowedCountries)+len(r.BlockedCountries) > 0
}

func hasAnyOpts(o *models.ReverseProxyTargetOpts) bool {
	return o != nil && (o.SkipTLSVerify || o.RequestTimeout != "" || o.SessionIdleTimeout != "" || o.PathRewrite != "" || len(o.CustomHeaders) > 0 || o.ProxyProtocol)
}

func deleteTargetAt(ts []models.ReverseProxyTarget, idx int) []models.ReverseProxyTarget {
	if idx < 0 || idx >= len(ts) {
		return ts
	}
	return append(ts[:idx], ts[idx+1:]...)
}

func defaultProtocolForMode(mode string) string {
	switch mode {
	case models.ReverseProxyModeHTTP:
		return "http"
	case models.ReverseProxyModeTCP, models.ReverseProxyModeTLS:
		return "tcp"
	case models.ReverseProxyModeUDP:
		return "udp"
	}
	return "http"
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

func cloneService(svc models.ReverseProxyService) models.ReverseProxyService {
	out := svc
	out.Targets = make([]models.ReverseProxyTarget, len(svc.Targets))
	for i, t := range svc.Targets {
		out.Targets[i] = t
		if t.Options != nil {
			optsCopy := *t.Options
			if t.Options.CustomHeaders != nil {
				optsCopy.CustomHeaders = make(map[string]string, len(t.Options.CustomHeaders))
				for k, v := range t.Options.CustomHeaders {
					optsCopy.CustomHeaders[k] = v
				}
			}
			out.Targets[i].Options = &optsCopy
		}
	}
	if svc.Auth != nil {
		authCopy := *svc.Auth
		// Deep-copy pointer-to-value auth method structs so edits don't leak.
		if svc.Auth.PasswordAuth != nil {
			p := *svc.Auth.PasswordAuth
			authCopy.PasswordAuth = &p
		}
		if svc.Auth.PinAuth != nil {
			p := *svc.Auth.PinAuth
			authCopy.PinAuth = &p
		}
		if svc.Auth.BearerAuth != nil {
			b := *svc.Auth.BearerAuth
			b.DistributionGroups = append([]string(nil), svc.Auth.BearerAuth.DistributionGroups...)
			authCopy.BearerAuth = &b
		}
		if svc.Auth.LinkAuth != nil {
			l := *svc.Auth.LinkAuth
			authCopy.LinkAuth = &l
		}
		authCopy.HeaderAuths = append([]models.ReverseProxyHeaderAuth(nil), svc.Auth.HeaderAuths...)
		out.Auth = &authCopy
	}
	if svc.AccessRestrictions != nil {
		arCopy := models.ReverseProxyAccessRestrictions{
			AllowedCIDRs:     append([]string(nil), svc.AccessRestrictions.AllowedCIDRs...),
			BlockedCIDRs:     append([]string(nil), svc.AccessRestrictions.BlockedCIDRs...),
			AllowedCountries: append([]string(nil), svc.AccessRestrictions.AllowedCountries...),
			BlockedCountries: append([]string(nil), svc.AccessRestrictions.BlockedCountries...),
		}
		out.AccessRestrictions = &arCopy
	}
	if svc.Meta != nil {
		mCopy := *svc.Meta
		out.Meta = &mCopy
	}
	return out
}

func confirmFieldsForSvc(svc models.ReverseProxyService, peers, groups map[string]string) []ConfirmField {
	fields := []ConfirmField{
		{Label: "Name", Value: svc.Name},
		{Label: "Domain", Value: svc.Domain},
		{Label: "Mode", Value: svc.Mode},
		{Label: "Enabled", Value: fmt.Sprintf("%v", svc.Enabled)},
	}
	if svc.ListenPort > 0 {
		fields = append(fields, ConfirmField{Label: "Listen Port", Value: strconv.Itoa(svc.ListenPort)})
	}
	if svc.ProxyCluster != "" {
		fields = append(fields, ConfirmField{Label: "Proxy Cluster", Value: svc.ProxyCluster})
	}
	tgtSummary := []string{}
	for _, t := range svc.Targets {
		tgtSummary = append(tgtSummary, fmt.Sprintf("%s %s:%d%s", targetDisplay(t, peers), t.Protocol, t.Port, defaultStr(t.Path, "")))
	}
	fields = append(fields, ConfirmField{Label: fmt.Sprintf("Targets (%d)", len(svc.Targets)), Value: strings.Join(tgtSummary, "; ")})
	fields = append(fields, ConfirmField{Label: "Auth", Value: rpAuthSummary(svc.Auth)})
	fields = append(fields, ConfirmField{Label: "Access Rules", Value: rpAccessSummary(svc.AccessRestrictions)})
	if svc.Mode == models.ReverseProxyModeHTTP {
		fields = append(fields, ConfirmField{Label: "Pass Host Header", Value: fmt.Sprintf("%v", svc.PassHostHeader)})
		fields = append(fields, ConfirmField{Label: "Rewrite Redirects", Value: fmt.Sprintf("%v", svc.RewriteRedirects)})
	}
	return fields
}
