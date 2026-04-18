package tui

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"netbird-manage/internal/models"
)

// ─── List-editor rendering helpers ──────────────────────────────────
//
// Wizard list screens (Targets, Access Rules, Header Auths, Custom Headers)
// all share the same shape: a scrollable list of rows plus a synthetic
// "[+ Add] ..." row at the end for appending. These helpers render that
// shape consistently.

// listEditorRow is a single row for a list editor.
type listEditorRow struct {
	cells []string
}

// renderListEditor produces a table + trailing hint line.
// headers: column titles. rows: data rows (stable order). addLabel: label for the synthetic trailing add row.
// cursor: currently-highlighted row index in [0, len(rows)] where len(rows) = the add row.
func renderListEditor(title string, headers []string, rows []listEditorRow, addLabel string, cursor int, width int, focused bool, hints string) string {
	var b strings.Builder

	b.WriteString(pageTitleStyle.Render(title) + "\n")

	// Build table rows, appending a synthetic "add" row when addLabel is non-empty.
	tableRows := make([][]string, 0, len(rows)+1)
	for _, r := range rows {
		tableRows = append(tableRows, r.cells)
	}
	hasAddRow := addLabel != ""
	if hasAddRow {
		addRow := make([]string, len(headers))
		addRow[0] = addLabel
		tableRows = append(tableRows, addRow)
	}

	tw := width
	if tw > 110 {
		tw = 110
	}
	if tw < 40 {
		tw = 40
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers(headers...).
		Rows(tableRows...).
		Width(tw).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return tableHeaderStyle
			}

			base := tableCellStyle
			if row%2 == 0 {
				base = tableDimCellStyle
			}
			if hasAddRow && row == len(tableRows)-1 {
				base = base.Foreground(colorOrange).Italic(true)
			}
			if row == cursor && focused {
				return tableSelectedStyle
			}
			return base
		})

	b.WriteString(t.Render() + "\n")
	b.WriteString(dimHintStyle.Render("  " + hints))

	return b.String()
}

// targetRows builds listEditorRow entries for a service's targets.
func targetRows(targets []models.ReverseProxyTarget, peers map[string]string) []listEditorRow {
	rows := make([]listEditorRow, 0, len(targets))
	for _, t := range targets {
		rows = append(rows, listEditorRow{cells: []string{
			targetDisplay(t, peers),
			t.Protocol,
			fmt.Sprintf("%d", t.Port),
			defaultStr(t.Path, "-"),
			boolBadge(t.Enabled),
		}})
	}
	return rows
}

// targetDisplay returns a compact description of a target's destination.
func targetDisplay(t models.ReverseProxyTarget, peers map[string]string) string {
	switch t.TargetType {
	case "peer":
		if name, ok := peers[t.TargetID]; ok {
			return fmt.Sprintf("peer:%s", name)
		}
		return fmt.Sprintf("peer:%s", shortID(t.TargetID))
	default:
		host := t.Host
		if host == "" {
			host = "?"
		}
		return fmt.Sprintf("%s:%s", t.TargetType, host)
	}
}

// accessRuleRows flattens AccessRestrictions into a list of (action, type, value) rows.
func accessRuleRows(r *models.ReverseProxyAccessRestrictions) []listEditorRow {
	if r == nil {
		return nil
	}
	rows := make([]listEditorRow, 0)
	for _, cidr := range r.AllowedCIDRs {
		action, kind, val := classifyCIDREntry("allow", cidr)
		rows = append(rows, listEditorRow{cells: []string{action, kind, val}})
	}
	for _, cidr := range r.BlockedCIDRs {
		action, kind, val := classifyCIDREntry("block", cidr)
		rows = append(rows, listEditorRow{cells: []string{action, kind, val}})
	}
	for _, c := range r.AllowedCountries {
		rows = append(rows, listEditorRow{cells: []string{"allow", "country", c}})
	}
	for _, c := range r.BlockedCountries {
		rows = append(rows, listEditorRow{cells: []string{"block", "country", c}})
	}
	return rows
}

// classifyCIDREntry reports whether a CIDR-looking string is actually a /32 single IP.
func classifyCIDREntry(action, value string) (string, string, string) {
	if strings.HasSuffix(value, "/32") {
		return action, "ip", strings.TrimSuffix(value, "/32")
	}
	return action, "cidr", value
}

// headerAuthRows flattens HeaderAuths into display rows.
func headerAuthRows(headers []models.ReverseProxyHeaderAuth) []listEditorRow {
	rows := make([]listEditorRow, 0, len(headers))
	for _, h := range headers {
		rows = append(rows, listEditorRow{cells: []string{
			h.Header,
			truncate(h.Value, 30),
			boolBadge(h.Enabled),
		}})
	}
	return rows
}

// customHeaderRows flattens a map[string]string into stable-ordered display rows.
func customHeaderRows(m map[string]string) []listEditorRow {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// sort for stable display
	sortStrings(keys)
	rows := make([]listEditorRow, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, listEditorRow{cells: []string{k, truncate(m[k], 40)}})
	}
	return rows
}

// ─── Small utilities ────────────────────────────────────────────────

func defaultStr(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func boolBadge(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8] + "…"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func sortStrings(ss []string) {
	sort.Strings(ss)
}
