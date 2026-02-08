package views

import (
	"github.com/charmbracelet/lipgloss"
)

// Shared color constants
var (
	ColorBlue  = lipgloss.Color("#61AFEF")
	ColorGray  = lipgloss.Color("#ABB2BF")
	ColorGreen = lipgloss.Color("#98C379")
	ColorRed   = lipgloss.Color("#E06C75")
)

// Shared styles
var (
	StyleInfo = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true)

	StyleDim = lipgloss.NewStyle().
			Foreground(ColorGray)

	StyleKeybinding = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleHighlight = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true).
			Underline(true)

	PreviewStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(1, 2)
)

// RenderCheckbox renders a checkbox with proper styling
func RenderCheckbox(checked bool) string {
	if checked {
		return lipgloss.NewStyle().Foreground(ColorBlue).Bold(true).Render("[✓]")
	}
	return StyleDim.Render("[ ]")
}

// RenderRadio renders a radio button with proper styling
func RenderRadio(selected bool) string {
	if selected {
		return lipgloss.NewStyle().Foreground(ColorBlue).Bold(true).Render("(•)")
	}
	return StyleDim.Render("( )")
}

// RenderProgressBar renders a simple progress bar
func RenderProgressBar(percentage int, width int) string {
	if width <= 0 {
		width = 40
	}
	filled := (percentage * width) / 100
	if filled > width {
		filled = width
	}

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	return lipgloss.NewStyle().Foreground(ColorBlue).Render(bar)
}

// RenderNavigationHint renders navigation hints at the bottom of a view
func RenderNavigationHint(hints ...string) string {
	if len(hints) == 0 {
		return ""
	}

	result := ""
	for i, hint := range hints {
		if i > 0 {
			result += "  "
		}
		result += hint
	}
	return StyleDim.Render(result)
}
