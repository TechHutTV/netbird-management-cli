package tui

import (
	tea "charm.land/bubbletea/v2"

	"netbird-manage/internal/client"
)

// placeholderPage is a temporary page shown for sections not yet implemented
type placeholderPage struct {
	title   string
	focused bool
}

func newPlaceholderPage(title string) *placeholderPage {
	return &placeholderPage{title: title}
}

func (p *placeholderPage) Init(_ *client.Client) tea.Cmd {
	return nil
}

func (p *placeholderPage) Update(msg tea.Msg, _ *client.Client) (Page, tea.Cmd) {
	return p, nil
}

func (p *placeholderPage) View(width, height int) string {
	t := pageTitleStyle.Render(p.title)
	desc := dimHintStyle.Render("  Press tab to return to the sidebar, then select a section.")
	return t + "\n\n" + desc
}

func (p *placeholderPage) Title() string {
	return p.title
}
func (p *placeholderPage) CursorPosition() int { return 0 }
func (p *placeholderPage) SetFocused(focused bool) { p.focused = focused }
