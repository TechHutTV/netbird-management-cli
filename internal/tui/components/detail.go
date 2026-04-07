package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	detailLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Width(20)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E0E0E0"))

	detailHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F58220")).
				Bold(true).
				MarginBottom(1)
)

// DetailField represents a label-value pair for detail views
type DetailField struct {
	Label string
	Value string
}

// RenderDetail renders a list of label-value fields
func RenderDetail(title string, fields []DetailField) string {
	var b strings.Builder

	b.WriteString(detailHeaderStyle.Render(title))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", 50))
	b.WriteString("\n")

	for _, f := range fields {
		label := detailLabelStyle.Render(f.Label + ":")
		value := detailValueStyle.Render(f.Value)
		b.WriteString(fmt.Sprintf("%s %s\n", label, value))
	}

	return b.String()
}
