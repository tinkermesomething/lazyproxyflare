package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PanelFocus represents which panel is currently focused
type PanelFocus int

const (
	PanelFocusLeft PanelFocus = iota
	PanelFocusDetails
	PanelFocusSnippets
)

// PanelLayout calculates panel dimensions based on terminal size
type PanelLayout struct {
	Width  int
	Height int

	// Left panel (entry list) - 40% width, full height
	LeftWidth  int
	LeftHeight int

	// Top-right panel (details) - 60% width, 60% height
	DetailsWidth  int
	DetailsHeight int

	// Bottom-right panel (snippets) - 60% width, 40% height
	SnippetsWidth  int
	SnippetsHeight int

	// Legacy right panel dimensions (for backward compatibility)
	RightWidth  int
	RightHeight int

	// Status bar
	StatusHeight int
}

// NewPanelLayout creates a new panel layout for the given terminal size
// Layout: Left panel (40%) | Details panel (60% top) + Snippets panel (60% bottom)
func NewPanelLayout(width, height int) PanelLayout {
	// Enforce minimum dimensions
	if width < 40 {
		width = 40
	}
	if height < 10 {
		height = 10
	}

	// Reserve space for title bar (1 line) and status bar (1 line)
	contentHeight := height - 2
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Left panel: 40% width, full content height
	leftWidth := int(float64(width) * 0.4)
	if leftWidth < 20 {
		leftWidth = 20
	}

	// Right panels share remaining width (60%)
	rightWidth := width - leftWidth
	if rightWidth < 20 {
		rightWidth = 20
	}

	// Details panel: 50% of content height (equal split)
	detailsHeight := contentHeight / 2
	if detailsHeight < 3 {
		detailsHeight = 3
	}

	// Complementary panel: remaining height (50%)
	snippetsHeight := contentHeight - detailsHeight
	if snippetsHeight < 3 {
		snippetsHeight = 3
	}

	return PanelLayout{
		Width:          width,
		Height:         height,
		LeftWidth:      leftWidth,
		LeftHeight:     contentHeight,
		DetailsWidth:   rightWidth,
		DetailsHeight:  detailsHeight,
		SnippetsWidth:  rightWidth,
		SnippetsHeight: snippetsHeight,
		// Legacy compatibility
		RightWidth:     rightWidth,
		RightHeight:    contentHeight,
		StatusHeight:   1,
	}
}

// RenderPanel renders a panel with title cutting through the top border (lazygit style)
func RenderPanel(title, content string, width, height int, focused bool) string {
	// Border characters (rounded)
	topLeft := "╭"
	topRight := "╮"
	bottomLeft := "╰"
	bottomRight := "╯"
	horizontal := "─"
	vertical := "│"

	// Colors based on focus
	var borderColor, titleColor lipgloss.Color
	if focused {
		borderColor = ColorBlue
		titleColor = ColorBlue
	} else {
		borderColor = ColorGray
		titleColor = ColorGray
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(focused)

	// Inner dimensions
	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := height - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	var result strings.Builder

	// Top border with title cutting through
	titleText := " " + title + " "
	titleLen := len(titleText)
	leftBorderLen := 2
	rightBorderLen := innerWidth - leftBorderLen - titleLen
	if rightBorderLen < 0 {
		// Truncate title if too long
		maxLen := innerWidth - leftBorderLen - 1
		if maxLen > 4 {
			titleText = titleText[:maxLen-2] + "… "
			titleLen = len(titleText)
			rightBorderLen = innerWidth - leftBorderLen - titleLen
		} else {
			titleText = ""
			titleLen = 0
			rightBorderLen = innerWidth - leftBorderLen
		}
	}
	if rightBorderLen < 0 {
		rightBorderLen = 0
	}

	result.WriteString(borderStyle.Render(topLeft))
	result.WriteString(borderStyle.Render(strings.Repeat(horizontal, leftBorderLen)))
	result.WriteString(titleStyle.Render(titleText))
	result.WriteString(borderStyle.Render(strings.Repeat(horizontal, rightBorderLen)))
	result.WriteString(borderStyle.Render(topRight))
	result.WriteString("\n")

	// Content lines
	contentLines := strings.Split(content, "\n")
	for i := 0; i < innerHeight; i++ {
		var lineContent string
		if i < len(contentLines) {
			lineContent = contentLines[i]
		}

		// Truncate if needed
		lineLen := lipgloss.Width(lineContent)
		if lineLen > innerWidth {
			lineContent = lineContent[:innerWidth-3] + "..."
			lineLen = innerWidth
		}
		padding := innerWidth - lineLen
		if padding < 0 {
			padding = 0
		}

		result.WriteString(borderStyle.Render(vertical))
		result.WriteString(lineContent)
		result.WriteString(strings.Repeat(" ", padding))
		result.WriteString(borderStyle.Render(vertical))
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString(borderStyle.Render(bottomLeft))
	result.WriteString(borderStyle.Render(strings.Repeat(horizontal, innerWidth)))
	result.WriteString(borderStyle.Render(bottomRight))

	return result.String()
}

// RenderTitleBarWithTabs renders the main title bar with Cloudflare/Caddy tab indicators
func RenderTitleBarWithTabs(domain string, activeTab int, width int) string {
	// Tab styles
	activeTabStyle := lipgloss.NewStyle().
		Foreground(ColorBlue).
		Bold(true).
		Underline(true)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(ColorGray)

	// Build tab indicators
	var cloudflareTab, caddyTab string
	if activeTab == 0 { // TabCloudflare
		cloudflareTab = activeTabStyle.Render("Cloudflare")
		caddyTab = inactiveTabStyle.Render("Caddy")
	} else { // TabCaddy
		cloudflareTab = inactiveTabStyle.Render("Cloudflare")
		caddyTab = activeTabStyle.Render("Caddy")
	}

	tabs := fmt.Sprintf("[ %s | %s ]", cloudflareTab, caddyTab)

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorBlue).
		Bold(true)
	title := titleStyle.Render(fmt.Sprintf("LazyProxyFlare - %s", domain))

	// Calculate spacing
	tabHint := lipgloss.NewStyle().Foreground(ColorDim).Render("(tab to switch)")

	// Layout: title on left, tabs in center, hint on right
	leftPart := title
	centerPart := tabs
	rightPart := tabHint

	leftWidth := lipgloss.Width(leftPart)
	centerWidth := lipgloss.Width(centerPart)
	rightWidth := lipgloss.Width(rightPart)

	// Calculate padding
	totalUsed := leftWidth + centerWidth + rightWidth
	remainingSpace := width - totalUsed
	if remainingSpace < 0 {
		remainingSpace = 0
	}

	leftPad := remainingSpace / 2
	rightPad := remainingSpace - leftPad

	return leftPart + strings.Repeat(" ", leftPad) + centerPart + strings.Repeat(" ", rightPad) + rightPart
}

// RenderStatusBar renders the bottom status/keybinding bar
func RenderStatusBar(content string, width int, hasError bool) string {
	style := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1)

	if hasError {
		style = style.Foreground(ColorRed)
	} else {
		style = style.Foreground(ColorWhite)
	}

	return style.Render(content)
}

// RenderTwoColumnLayout renders left and right panels side by side
func RenderTwoColumnLayout(leftPanel, rightPanel string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

// RenderThreePanelLayout renders three panels: left full height, details top-right, snippets bottom-right
func RenderThreePanelLayout(leftPanel, detailsPanel, snippetsPanel string) string {
	// Stack details and snippets vertically
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, detailsPanel, snippetsPanel)

	// Join left panel with right column horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightColumn)
}

// RenderFullLayout combines title bar, panels, and status bar
func RenderFullLayout(titleBar, panels, statusBar string) string {
	return lipgloss.JoinVertical(lipgloss.Left, titleBar, panels, statusBar)
}

// RenderModalOverlay renders a centered modal/popup window.
// Following lazygit's pattern: modals render on top without showing background.
// This avoids ANSI code corruption from string slicing and the "black box" issue.
func RenderModalOverlay(background, title, modalContent string, width, height int) string {
	// Modal dimensions - 4/7 width like lazygit, capped at 80
	modalWidth := width * 4 / 7
	if modalWidth < 40 {
		modalWidth = min(width-2, 40)
	}
	if modalWidth > 80 {
		modalWidth = 80
	}

	// Modal height - 3/4 of screen like lazygit
	modalHeight := height * 3 / 4
	if modalHeight < 15 {
		modalHeight = min(height-2, 15)
	}

	// Use lipgloss Border for clean rendering
	modal := renderModalBox(title, modalContent, modalWidth, modalHeight)

	// Center on screen - no background compositing (lazygit approach)
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

// renderModalBox creates a modal with manual border rendering
func renderModalBox(title, content string, width, height int) string {
	// Inner dimensions
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	innerHeight := height - 2
	if innerHeight < 1 {
		innerHeight = 1
	}

	// Border characters
	borderStyle := lipgloss.NewStyle().Foreground(ColorBlue)
	titleStyle := lipgloss.NewStyle().Foreground(ColorBlue).Bold(true)

	topLeft := borderStyle.Render("╭")
	topRight := borderStyle.Render("╮")
	bottomLeft := borderStyle.Render("╰")
	bottomRight := borderStyle.Render("╯")
	horizontal := "─"
	vertical := borderStyle.Render("│")

	var result strings.Builder

	// Build top border with title
	titleText := " " + title + " "
	titleLen := len(titleText)
	leftBars := 2
	rightBars := innerWidth - leftBars - titleLen
	if rightBars < 0 {
		// Title too long, truncate
		if innerWidth > 8 {
			titleText = " " + title[:innerWidth-7] + "… "
			titleLen = len(titleText)
			rightBars = innerWidth - leftBars - titleLen
		} else {
			titleText = ""
			titleLen = 0
			rightBars = innerWidth - leftBars
		}
	}

	result.WriteString(topLeft)
	result.WriteString(borderStyle.Render(strings.Repeat(horizontal, leftBars)))
	if titleText != "" {
		result.WriteString(titleStyle.Render(titleText))
	}
	result.WriteString(borderStyle.Render(strings.Repeat(horizontal, rightBars)))
	result.WriteString(topRight)
	result.WriteString("\n")

	// Process content lines
	contentLines := strings.Split(content, "\n")
	availableWidth := innerWidth - 2 // padding

	for i := 0; i < innerHeight; i++ {
		var line string
		if i < len(contentLines) {
			line = contentLines[i]
		}

		// Truncate if too wide
		lineWidth := lipgloss.Width(line)
		if lineWidth > availableWidth && availableWidth > 3 {
			line = truncateWithANSI(line, availableWidth-3) + "..."
			lineWidth = availableWidth
		} else if lineWidth > availableWidth {
			line = truncateWithANSI(line, availableWidth)
			lineWidth = lipgloss.Width(line)
		}

		// Pad to fill width
		padding := availableWidth - lineWidth
		if padding < 0 {
			padding = 0
		}

		result.WriteString(vertical)
		result.WriteString(" ")
		result.WriteString(line)
		result.WriteString(strings.Repeat(" ", padding+1))
		result.WriteString(vertical)
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString(bottomLeft)
	result.WriteString(borderStyle.Render(strings.Repeat(horizontal, innerWidth)))
	result.WriteString(bottomRight)

	return result.String()
}

// truncateWithANSI truncates a string with ANSI codes to a visual width
func truncateWithANSI(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Simple approach: iterate runes, skip ANSI sequences
	var result strings.Builder
	var visualWidth int
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}

		if visualWidth >= maxWidth {
			break
		}
		result.WriteRune(r)
		visualWidth++
	}

	return result.String()
}

// renderModalWithTitle creates a modal box with title cutting through the top border
func renderModalWithTitle(title, content string, width, height int) string {
	// Border characters (rounded)
	topLeft := "╭"
	topRight := "╮"
	bottomLeft := "╰"
	bottomRight := "╯"
	horizontal := "─"
	vertical := "│"

	// Colors - no explicit background to avoid "box" effect
	borderColor := lipgloss.NewStyle().Foreground(ColorBlue)
	titleStyle := lipgloss.NewStyle().Foreground(ColorBlue).Bold(true)

	// Inner dimensions (account for borders)
	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	innerHeight := height - 2
	if innerHeight < 0 {
		innerHeight = 0
	}

	var result strings.Builder

	// Build top border with title cutting through
	titleText := " " + title + " "
	titleLen := len(titleText)
	leftBorderLen := 2 // After corner
	rightBorderLen := innerWidth - leftBorderLen - titleLen

	// Handle edge cases for very small widths or long titles
	if innerWidth < 6 {
		// Too small for any title, just render border
		titleText = ""
		titleLen = 0
		leftBorderLen = 0
		rightBorderLen = innerWidth
	} else if rightBorderLen < 0 {
		// Title too long, truncate
		maxTitleLen := innerWidth - leftBorderLen - 1 // Leave at least 1 char for right border
		if maxTitleLen < 4 {
			// Not enough room, skip title
			titleText = ""
			titleLen = 0
			rightBorderLen = innerWidth - leftBorderLen
		} else {
			titleText = titleText[:maxTitleLen-2] + "… "
			titleLen = len(titleText)
			rightBorderLen = innerWidth - leftBorderLen - titleLen
		}
	}

	// Ensure non-negative values for Repeat
	if leftBorderLen < 0 {
		leftBorderLen = 0
	}
	if rightBorderLen < 0 {
		rightBorderLen = 0
	}

	topBorder := borderColor.Render(topLeft) +
		borderColor.Render(strings.Repeat(horizontal, leftBorderLen)) +
		titleStyle.Render(titleText) +
		borderColor.Render(strings.Repeat(horizontal, rightBorderLen)) +
		borderColor.Render(topRight)
	result.WriteString(topBorder)
	result.WriteString("\n")

	// Split content into lines
	contentLines := strings.Split(content, "\n")

	// Render content lines with side borders
	for i := 0; i < innerHeight; i++ {
		var lineContent string
		if i < len(contentLines) {
			lineContent = contentLines[i]
		} else {
			lineContent = ""
		}

		// Calculate available space for content (between borders and padding)
		availableWidth := innerWidth - 2 // Account for 1 space padding on each side
		if availableWidth < 0 {
			availableWidth = 0
		}

		// Truncate if needed
		lineLen := lipgloss.Width(lineContent)
		if lineLen > availableWidth && availableWidth > 3 {
			// Truncate with ellipsis
			lineContent = lineContent[:availableWidth-3] + "..."
			lineLen = availableWidth
		} else if lineLen > availableWidth {
			// Too small for ellipsis, just truncate
			if availableWidth > 0 {
				lineContent = lineContent[:availableWidth]
			} else {
				lineContent = ""
			}
			lineLen = len(lineContent)
		}

		// Calculate padding
		padding := availableWidth - lineLen
		if padding < 0 {
			padding = 0
		}

		// Build line with borders (no background styling to avoid box effect)
		line := borderColor.Render(vertical) +
			" " + lineContent + strings.Repeat(" ", padding+1) +
			borderColor.Render(vertical)
		result.WriteString(line)
		result.WriteString("\n")
	}

	// Bottom border
	bottomBorder := borderColor.Render(bottomLeft) +
		borderColor.Render(strings.Repeat(horizontal, innerWidth)) +
		borderColor.Render(bottomRight)
	result.WriteString(bottomBorder)

	return result.String()
}

// RenderModal renders a centered modal/popup window (legacy, now uses RenderModalOverlay)
func RenderModal(title, content string, width, height int) string {
	return RenderModalOverlay("", title, content, width, height)
}
