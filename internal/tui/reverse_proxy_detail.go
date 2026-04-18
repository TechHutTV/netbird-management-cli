package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"netbird-manage/internal/models"
)

// rpDetailTab enumerates the read-only detail tabs on a single reverse-proxy service.
type rpDetailTab int

const (
	rpDetailService rpDetailTab = iota
	rpDetailAuth
	rpDetailAccess
	rpDetailAdvanced
	rpDetailMeta
	rpDetailTabsCount
)

var rpDetailTabLabels = []string{"Service", "Auth", "Access", "Advanced", "Meta"}

func (t rpDetailTab) Next() rpDetailTab {
	n := t + 1
	if n >= rpDetailTabsCount {
		return 0
	}
	return n
}

func (t rpDetailTab) Prev() rpDetailTab {
	if t == 0 {
		return rpDetailTabsCount - 1
	}
	return t - 1
}

// renderRPDetail renders the tabbed detail view for a service.
func renderRPDetail(svc models.ReverseProxyService, tab rpDetailTab, peers, groups map[string]string) string {
	var b strings.Builder

	b.WriteString(renderRPDetailHeader(svc))
	b.WriteString("\n\n")
	b.WriteString(renderRPTabBar(tab))
	b.WriteString("\n\n")

	switch tab {
	case rpDetailService:
		b.WriteString(renderRPServiceTab(svc, peers))
	case rpDetailAuth:
		b.WriteString(renderRPAuthTab(svc.Auth, groups))
	case rpDetailAccess:
		b.WriteString(renderRPAccessTab(svc.AccessRestrictions))
	case rpDetailAdvanced:
		b.WriteString(renderRPAdvancedTab(svc))
	case rpDetailMeta:
		b.WriteString(renderRPMetaTab(svc))
	}

	b.WriteString("\n\n")
	b.WriteString(dimHintStyle.Render("  esc: back  ←/→: switch tab  e: edit  t: toggle  d: delete  D: domains  E: events"))
	return b.String()
}

func renderRPDetailHeader(svc models.ReverseProxyService) string {
	badge := disabledStyle.Render(" DISABLED ")
	if svc.Enabled {
		badge = enabledStyle.Render(" ENABLED ")
	}

	status := ""
	if svc.Meta != nil && svc.Meta.Status != "" {
		status = " " + rpStatusBadge(svc.Meta.Status)
	}

	modeBadge := ""
	if svc.Mode != "" {
		modeBadge = lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true).
			Render(" " + strings.ToUpper(svc.Mode) + " ")
	}

	title := fmt.Sprintf("  %s  %s%s%s", svc.Domain, modeBadge, badge, status)
	return detailTitleStyle.Render(title)
}

// renderRPTabBar renders a horizontal list of detail tabs with the active one highlighted.
func renderRPTabBar(active rpDetailTab) string {
	labels := make([]string, 0, len(rpDetailTabLabels))
	activeStyle := lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Padding(0, 2).Border(lipgloss.RoundedBorder()).BorderForeground(colorOrange)
	idleStyle := lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2)

	for i, l := range rpDetailTabLabels {
		if rpDetailTab(i) == active {
			labels = append(labels, activeStyle.Render(l))
		} else {
			labels = append(labels, idleStyle.Render(l))
		}
	}
	return strings.Join(labels, "  ")
}

func renderRPServiceTab(svc models.ReverseProxyService, peers map[string]string) string {
	var b strings.Builder

	b.WriteString(renderField("Name", svc.Name))
	b.WriteString(renderField("ID", svc.ID))
	b.WriteString(renderField("Domain", svc.Domain))
	b.WriteString(renderField("Mode", svc.Mode))
	listenPort := "(auto)"
	if svc.ListenPort > 0 {
		listenPort = fmt.Sprintf("%d", svc.ListenPort)
	}
	b.WriteString(renderField("Listen Port", listenPort))
	if svc.ProxyCluster != "" {
		b.WriteString(renderField("Proxy Cluster", svc.ProxyCluster))
	}
	b.WriteString(renderField("Enabled", fmt.Sprintf("%v", svc.Enabled)))

	b.WriteString("\n" + sectionHeaderStyle.Render(fmt.Sprintf("  Targets (%d)", len(svc.Targets))) + "\n")
	if len(svc.Targets) == 0 {
		b.WriteString(dimHintStyle.Render("  (no targets)"))
	} else {
		for i, t := range svc.Targets {
			b.WriteString(renderTargetLine(i+1, t, peers))
		}
	}
	return b.String()
}

func renderTargetLine(idx int, t models.ReverseProxyTarget, peers map[string]string) string {
	var sb strings.Builder
	header := fmt.Sprintf("  %d. %s %s:%d", idx, targetDisplay(t, peers), t.Protocol, t.Port)
	if t.Path != "" {
		header += " " + t.Path
	}
	if !t.Enabled {
		header += " " + disabledStyle.Render("(disabled)")
	}
	sb.WriteString(detailValueStyle.Render(header) + "\n")

	if t.Options != nil {
		opts := []string{}
		if t.Options.SkipTLSVerify {
			opts = append(opts, "skip-tls")
		}
		if t.Options.ProxyProtocol {
			opts = append(opts, "proxy-protocol")
		}
		if t.Options.RequestTimeout != "" {
			opts = append(opts, "timeout="+t.Options.RequestTimeout)
		}
		if t.Options.SessionIdleTimeout != "" {
			opts = append(opts, "idle="+t.Options.SessionIdleTimeout)
		}
		if t.Options.PathRewrite != "" {
			opts = append(opts, "rewrite="+t.Options.PathRewrite)
		}
		if len(t.Options.CustomHeaders) > 0 {
			opts = append(opts, fmt.Sprintf("headers=%d", len(t.Options.CustomHeaders)))
		}
		if len(opts) > 0 {
			sb.WriteString(dimHintStyle.Render("     "+strings.Join(opts, ", ")) + "\n")
		}
	}
	return sb.String()
}

func renderRPAuthTab(auth *models.ReverseProxyAuth, groups map[string]string) string {
	var b strings.Builder

	if auth == nil {
		return dimHintStyle.Render("  No authentication configured.")
	}

	b.WriteString(renderAuthMethodLine("Password", auth.PasswordAuth != nil && auth.PasswordAuth.Enabled, "shared password"))
	b.WriteString(renderAuthMethodLine("PIN", auth.PinAuth != nil && auth.PinAuth.Enabled, "numeric code"))

	bearerDesc := "bearer / SSO"
	if auth.BearerAuth != nil && auth.BearerAuth.Enabled && len(auth.BearerAuth.DistributionGroups) > 0 {
		names := resolveGroupNames(auth.BearerAuth.DistributionGroups, groups)
		bearerDesc = "groups: " + names
	}
	b.WriteString(renderAuthMethodLine("Bearer (SSO)", auth.BearerAuth != nil && auth.BearerAuth.Enabled, bearerDesc))

	b.WriteString(renderAuthMethodLine("Magic Link", auth.LinkAuth != nil && auth.LinkAuth.Enabled, "one-time email"))

	if len(auth.HeaderAuths) > 0 {
		b.WriteString("\n" + sectionHeaderStyle.Render(fmt.Sprintf("  Header Auths (%d)", len(auth.HeaderAuths))) + "\n")
		for i, h := range auth.HeaderAuths {
			status := "disabled"
			if h.Enabled {
				status = "enabled"
			}
			b.WriteString(detailValueStyle.Render(fmt.Sprintf("  %d. %s: %s (%s)", i+1, h.Header, truncate(h.Value, 40), status)) + "\n")
		}
	}
	return b.String()
}

func renderAuthMethodLine(name string, enabled bool, desc string) string {
	label := detailLabelStyle.Render(name)
	state := disabledStyle.Render("off")
	if enabled {
		state = enabledStyle.Render("on")
	}
	return fmt.Sprintf("%s  %s  %s\n", label, state, detailValueStyle.Render(desc))
}

func renderRPAccessTab(r *models.ReverseProxyAccessRestrictions) string {
	if r == nil {
		return dimHintStyle.Render("  No access restrictions.")
	}
	var b strings.Builder
	emitRules := func(label string, rules []string) {
		if len(rules) == 0 {
			return
		}
		b.WriteString(sectionHeaderStyle.Render("  "+label) + "\n")
		for _, v := range rules {
			b.WriteString(detailValueStyle.Render("    - "+v) + "\n")
		}
	}
	emitRules(fmt.Sprintf("Allowed CIDRs (%d)", len(r.AllowedCIDRs)), r.AllowedCIDRs)
	emitRules(fmt.Sprintf("Blocked CIDRs (%d)", len(r.BlockedCIDRs)), r.BlockedCIDRs)
	emitRules(fmt.Sprintf("Allowed Countries (%d)", len(r.AllowedCountries)), r.AllowedCountries)
	emitRules(fmt.Sprintf("Blocked Countries (%d)", len(r.BlockedCountries)), r.BlockedCountries)
	if b.Len() == 0 {
		return dimHintStyle.Render("  No access restrictions.")
	}
	return b.String()
}

func renderRPAdvancedTab(svc models.ReverseProxyService) string {
	var b strings.Builder
	switch svc.Mode {
	case models.ReverseProxyModeHTTP:
		b.WriteString(renderField("Pass Host Header", fmt.Sprintf("%v", svc.PassHostHeader)))
		b.WriteString(renderField("Rewrite Redirects", fmt.Sprintf("%v", svc.RewriteRedirects)))
	case models.ReverseProxyModeTCP, models.ReverseProxyModeTLS:
		pp := false
		to := ""
		if len(svc.Targets) > 0 && svc.Targets[0].Options != nil {
			pp = svc.Targets[0].Options.ProxyProtocol
			to = svc.Targets[0].Options.RequestTimeout
		}
		b.WriteString(renderField("Proxy Protocol", fmt.Sprintf("%v", pp)))
		b.WriteString(renderField("Request Timeout", defaultStr(to, "(none)")))
	case models.ReverseProxyModeUDP:
		to := ""
		if len(svc.Targets) > 0 && svc.Targets[0].Options != nil {
			to = svc.Targets[0].Options.SessionIdleTimeout
		}
		b.WriteString(renderField("Session Idle Timeout", defaultStr(to, "(none)")))
	}
	if b.Len() == 0 {
		return dimHintStyle.Render("  No advanced settings for this mode.")
	}
	return b.String()
}

func renderRPMetaTab(svc models.ReverseProxyService) string {
	if svc.Meta == nil {
		return dimHintStyle.Render("  No metadata available.")
	}
	var b strings.Builder
	b.WriteString(renderField("Status", rpStatusBadge(svc.Meta.Status)))
	b.WriteString(renderField("Created", svc.Meta.CreatedAt))
	b.WriteString(renderField("Updated", svc.Meta.UpdatedAt))
	return b.String()
}

// rpStatusBadge renders a colored status string.
func rpStatusBadge(status string) string {
	switch status {
	case models.ReverseProxyStatusActive:
		return enabledStyle.Render(" active ")
	case models.ReverseProxyStatusPending, models.ReverseProxyStatusCertificatePending:
		return lipgloss.NewStyle().Foreground(colorYellow).Render(" " + status + " ")
	case models.ReverseProxyStatusCertificateFailed, models.ReverseProxyStatusError, models.ReverseProxyStatusTunnelNotCreated:
		return disabledStyle.Render(" " + status + " ")
	default:
		if status == "" {
			return detailValueStyle.Render("(unknown)")
		}
		return detailValueStyle.Render(" " + status + " ")
	}
}

// renderField is a detail-view helper producing a label+value row.
func renderField(label, value string) string {
	if value == "" {
		value = "(none)"
	}
	return fmt.Sprintf("%s  %s\n", detailLabelStyle.Render(label), detailValueStyle.Render(value))
}
