package components

import (
	"charm.land/lipgloss/v2"
)

var (
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Background(lipgloss.Color("#16213E")).
			Padding(0, 1)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F58220")).
			Bold(true)
)

// StatusBar renders a status bar with key hints
func StatusBar(hints string, width int) string {
	return statusStyle.Width(width).Render(hints)
}

// KeyHint renders a styled key hint like "q: quit"
func KeyHint(key, desc string) string {
	return keyStyle.Render(key) + " " + desc
}
