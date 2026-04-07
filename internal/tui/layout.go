package tui

import "charm.land/lipgloss/v2"

// navBarHeight is the height consumed by the top nav bar (tab row + bottom border)
const navBarHeight = 2

// renderLayout composes the header, nav bar, and content into a vertical layout
func renderLayout(header, navbar, content string, width, height int) string {
	// height budget: header(2) + navBar(2) + content + padding(2) + statusBar(1)
	contentHeight := height - headerHeight - navBarHeight - 1 - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	contentBox := contentStyle.
		Width(width).
		Height(contentHeight).
		Render(content)

	return header + "\n" + navbar + "\n" + contentBox
}

// renderStatusBar renders the segmented status bar at the bottom
func renderStatusBar(section, hints string, width int) string {
	sectionBadge := statusKeyStyle.Render(section)

	remaining := width - lipgloss.Width(sectionBadge)
	if remaining < 0 {
		remaining = 0
	}

	helpText := statusHelpStyle.Width(remaining).Render(hints)

	return lipgloss.JoinHorizontal(lipgloss.Top, sectionBadge, helpText)
}
