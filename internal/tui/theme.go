package tui

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

// Detect dark/light terminal background
var hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)

// Adaptive color helper — picks the right shade for the terminal background
func adaptive(light, dark string) color.Color {
	if hasDarkBG {
		return lipgloss.Color(dark)
	}
	return lipgloss.Color(light)
}

// Brand colors
var (
	colorOrange color.Color = lipgloss.Color("#F58220") // NetBird brand
	colorBlue   color.Color = lipgloss.Color("#4A9FD9")
	colorPurple color.Color = lipgloss.Color("#874BFD")
	colorGreen  color.Color = lipgloss.Color("#2ECC71")
	colorRed    color.Color = lipgloss.Color("#E74C3C")
	colorYellow color.Color = lipgloss.Color("#F1C40F")
)

// Adaptive text colors
var (
	colorPrimary   = adaptive("#1A1A2E", "#F58220")
	colorSecondary = adaptive("#2C5282", "#4A9FD9")
	colorText      = adaptive("#1A1A2E", "#E0E0E0")
	colorTextDim   = adaptive("#718096", "#6C757D")
	colorFaint     = adaptive("#A0AEC0", "#444444")
	colorSuccess   = adaptive("#27AE60", "#2ECC71")
	colorDanger    = adaptive("#C0392B", "#E74C3C")
)

// Border colors
var (
	colorBorder      = adaptive("#CBD5E0", "#3B3B3B")
	colorBorderFocus = adaptive("#F58220", "#F58220")
	colorBorderFaint = adaptive("#E2E8F0", "#2A2A2A")
)

// Background colors
var (
	colorSidebarBg  = adaptive("#F7FAFC", "#0F1923")
	colorSelected   = adaptive("#EBF8FF", "#1A3A5C")
	colorHeaderBg   = adaptive("#F58220", "#F58220")
)

// ─── Nav bar (top) ──────────────────────────────────────────────────
// Nav bar styles are defined in nav.go alongside the NavModel

// ─── Content ────────────────────────────────────────────────────────

var (
	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Page title
	pageTitleStyle = lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true).
			MarginBottom(1)

	// Section sub-header
	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(colorBlue).
				Bold(true).
				MarginTop(1)
)

// ─── Tables ─────────────────────────────────────────────────────────

var (
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Bold(true).
				Padding(0, 1)

	tableCellStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 1)

	tableSelectedStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorSelected).
				Bold(true).
				Padding(0, 1)

	tableDimCellStyle = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Padding(0, 1)

	tableBorderStyle = lipgloss.NewStyle().
				Foreground(colorBorder)
)

// ─── Status badges ──────────────────────────────────────────────────

var (
	onlineStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	offlineStyle = lipgloss.NewStyle().
			Foreground(colorDanger)

	validStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	expiredStyle = lipgloss.NewStyle().
			Foreground(colorDanger)

	revokedStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Strikethrough(true)

	enabledStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	disabledStyle = lipgloss.NewStyle().
			Foreground(colorDanger)
)

// ─── Status bar (bottom) ────────────────────────────────────────────

var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Background(adaptive("#EDF2F7", "#111927"))

	statusKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF")).
			Background(colorOrange).
			Bold(true).
			Padding(0, 1)

	statusValueStyle = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Background(adaptive("#E2E8F0", "#1A2332")).
				Padding(0, 1)

	statusHelpStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Background(adaptive("#EDF2F7", "#111927")).
			Padding(0, 1)
)

// ─── Detail views ───────────────────────────────────────────────────

var (
	detailLabelStyle = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Width(20).
				Align(lipgloss.Right).
				PaddingRight(1)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(colorText)

	detailTitleStyle = lipgloss.NewStyle().
				Foreground(colorOrange).
				Bold(true).
				BorderBottom(true).
				BorderStyle(lipgloss.Border{Bottom: "─"}).
				BorderForeground(colorBorder).
				MarginBottom(1)

	dimHintStyle = lipgloss.NewStyle().
			Foreground(colorFaint).
			MarginTop(1)
)

// ─── Misc ───────────────────────────────────────────────────────────

var (
	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	loadingStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Italic(true)
)
