package ui

import "github.com/charmbracelet/lipgloss"

// Color palette for LazyProxyFlare
var (
	// Status Colors
	ColorGreen  = lipgloss.Color("#80C080") // Synced, success
	ColorYellow = lipgloss.Color("#E0E080") // Orphaned DNS
	ColorOrange = lipgloss.Color("#FFAA55") // Orphaned Caddy, warnings
	ColorRed    = lipgloss.Color("#FF0000") // Errors, destructive
	ColorBlue   = lipgloss.Color("#6A9CEE") // Selected, focused, info
	ColorGray   = lipgloss.Color("#A0A0A0") // Disabled, dimmed, borders
	ColorWhite  = lipgloss.Color("#FFFFFF") // Normal text
	ColorDim    = lipgloss.Color("#707070") // Very dimmed text

	// Themed Styles
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorGreen)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorOrange)
	StyleError   = lipgloss.NewStyle().Foreground(ColorRed)
	StyleInfo    = lipgloss.NewStyle().Foreground(ColorBlue)
	StyleDim     = lipgloss.NewStyle().Foreground(ColorDim)

	// Panel Border Styles
	StyleBorderFocused = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorBlue)

	StyleBorderUnfocused = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorGray)

	// Panel Title Styles
	StyleTitleFocused = lipgloss.NewStyle().
				Foreground(ColorBlue).
				Bold(true)

	StyleTitleUnfocused = lipgloss.NewStyle().
				Foreground(ColorGray)

	// Status Icon Styles
	StyleIconSynced = lipgloss.NewStyle().Foreground(ColorGreen)
	StyleIconOrphan = lipgloss.NewStyle().Foreground(ColorOrange)

	// Selection/Highlight Style
	StyleSelected = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true)

	// Subtle cursor highlight (less harsh than full background)
	StyleHighlight = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true)

	// Keybinding hint styles
	StyleKeybinding = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true)

	StyleKeybindingDesc = lipgloss.NewStyle().
			Foreground(ColorWhite)
)
