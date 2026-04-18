package tui

import (
	tea "charm.land/bubbletea/v2"

	"netbird-manage/internal/client"
)

// ExportImportPage provides export and import functionality
type ExportImportPage struct {
	focused bool
}

func NewExportImportPage() *ExportImportPage {
	return &ExportImportPage{}
}

func (e *ExportImportPage) Title() string { return "Export/Import" }
func (e *ExportImportPage) CursorPosition() int { return 0 }
func (e *ExportImportPage) SetFocused(focused bool) { e.focused = focused }

func (e *ExportImportPage) Init(_ *client.Client) tea.Cmd {
	return nil
}

func (e *ExportImportPage) Update(msg tea.Msg, _ *client.Client) (Page, tea.Cmd) {
	return e, nil
}

func (e *ExportImportPage) View(width, height int) string {
	return pageTitleStyle.Render("Export / Import") + "\n\n" +
		detailValueStyle.Render("Export and import are currently available via the CLI:") + "\n\n" +
		sectionHeaderStyle.Render("  Export") + "\n" +
		"  netbird-manage export -full              Export to YAML\n" +
		"  netbird-manage export -full -format json  Export to JSON\n" +
		"  netbird-manage export -split              Export to multiple files\n\n" +
		sectionHeaderStyle.Render("  Import") + "\n" +
		"  netbird-manage import <file>              Dry-run preview\n" +
		"  netbird-manage import <file> --apply      Apply changes\n\n" +
		dimHintStyle.Render("  TUI export/import coming in a future update.")
}
