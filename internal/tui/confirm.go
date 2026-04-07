package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// ConfirmField is a label-value pair shown in the confirmation summary
type ConfirmField struct {
	Label string
	Value string
}

// RenderConfirm renders a confirmation screen with a title, fields, and y/n prompt
func RenderConfirm(title string, fields []ConfirmField) string {
	var sb strings.Builder

	labelStyle := lipgloss.NewStyle().Foreground(colorTextDim)
	lbl := func(s string) string {
		return labelStyle.Render(padLabel(s, 20))
	}

	sb.WriteString(pageTitleStyle.Render(title) + "\n\n")

	for _, f := range fields {
		if f.Value == "" {
			continue
		}
		sb.WriteString(lbl(f.Label+":") + detailValueStyle.Render(f.Value) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorYellow).
		Foreground(colorYellow).
		Padding(0, 2).
		Render("Submit? Press y to confirm, n or esc to cancel"))
	sb.WriteString("\n")

	return sb.String()
}

func padLabel(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
