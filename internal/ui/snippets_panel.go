package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"lazyproxyflare/internal/caddy"
)

// renderSnippetsPanel renders the snippets panel content with scrolling support
func (m Model) renderSnippetsPanel(height int) string {
	// When on Caddy tab with a selected entry, show only that entry's snippets
	if m.activeTab == TabCaddy {
		return m.renderAppliedSnippetsPanel(height)
	}

	// Default: show all available snippets
	return m.renderAllSnippetsPanel(height)
}

// renderAppliedSnippetsPanel shows snippets used by the selected entry
func (m Model) renderAppliedSnippetsPanel(height int) string {
	var b strings.Builder

	filteredEntries := m.getFilteredEntries()
	if m.cursor >= len(filteredEntries) {
		b.WriteString(StyleDim.Render("Select an entry to see applied snippets"))
		return b.String()
	}

	entry := filteredEntries[m.cursor]
	if entry.Caddy == nil || len(entry.Caddy.Imports) == 0 {
		b.WriteString(StyleDim.Render("No snippets applied to this entry\n\n"))
		b.WriteString(StyleDim.Render("Use 'e' to edit and add snippets"))
		return b.String()
	}

	// Show applied snippets with details
	for i, importName := range entry.Caddy.Imports {
		// Find the snippet details
		var snippet *caddy.Snippet
		for j := range m.snippets {
			if m.snippets[j].Name == importName {
				snippet = &m.snippets[j]
				break
			}
		}

		if snippet == nil {
			// Snippet reference but not found in Caddyfile
			b.WriteString(StyleWarning.Render("⚠ " + importName + " (not found)"))
			b.WriteString("\n")
			continue
		}

		// Snippet name with category color
		nameStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(snippet.Category.ColorCode())).
			Bold(true)
		b.WriteString(nameStyle.Render("● " + snippet.Name))
		b.WriteString("\n")

		// Category badge
		categoryStyle := lipgloss.NewStyle().
			Foreground(ColorGray).
			Italic(true)
		b.WriteString(categoryStyle.Render("  " + snippet.Category.String()))
		b.WriteString("\n")

		// Description
		descStyle := lipgloss.NewStyle().
			Foreground(ColorDim).
			PaddingLeft(2)
		b.WriteString(descStyle.Render(snippet.Description))

		if i < len(entry.Caddy.Imports)-1 {
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

// renderAllSnippetsPanel shows all available snippets
func (m Model) renderAllSnippetsPanel(height int) string {
	var b strings.Builder

	if len(m.snippets) == 0 {
		// No snippets available
		b.WriteString(StyleDim.Render("No snippets found in Caddyfile\n"))
		b.WriteString(StyleDim.Render("\n"))
		b.WriteString(StyleInfo.Render("Press 'w' to create a new snippet"))
		return b.String()
	}

	// Header removed - title is in panel border now

	// Calculate snippet usage statistics
	snippetUsage := m.calculateSnippetUsage()

	// Calculate visible range for scrolling
	// Each snippet takes approximately 4 lines (cursor+name+badge+desc+location+spacing)
	// Account for header (2 lines) and scroll indicator (1 line) and borders/padding (2 lines)
	availableHeight := height - 5
	if availableHeight < 4 {
		availableHeight = 4
	}
	linesPerSnippet := 4
	visibleSnippets := availableHeight / linesPerSnippet
	if visibleSnippets < 1 {
		visibleSnippets = 1
	}

	start := m.snippetPanel.ScrollOffset
	end := start + visibleSnippets

	// Adjust bounds
	if end > len(m.snippets) {
		end = len(m.snippets)
	}
	if start >= len(m.snippets) {
		start = 0
		end = len(m.snippets)
	}

	// List visible snippets with category badges and confidence
	for i := start; i < end; i++ {
		snippet := m.snippets[i]
		// Cursor indicator (→ for selected, space otherwise)
		cursorStr := "  "
		if i == m.snippetPanel.Cursor && m.panelFocus == PanelFocusSnippets {
			cursorStr = StyleHighlight.Render("→ ")
		}
		b.WriteString(cursorStr)

		// Snippet name
		nameStyle := lipgloss.NewStyle().
			Foreground(ColorBlue).
			Bold(true)
		if i == m.snippetPanel.Cursor && m.panelFocus == PanelFocusSnippets {
			nameStyle = nameStyle.Foreground(lipgloss.Color("#00D9FF")) // Brighter blue when selected
		}
		b.WriteString(nameStyle.Render(snippet.Name))
		b.WriteString(" ")

		// Category badge with color
		categoryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(snippet.Category.ColorCode())).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1).
			MarginRight(1)
		b.WriteString(categoryStyle.Render(snippet.Category.String()))

		// Auto-detected badge if applicable
		if snippet.AutoDetected {
			confidenceStyle := lipgloss.NewStyle().
				Foreground(ColorGreen).
				Italic(true)
			b.WriteString(confidenceStyle.Render(fmt.Sprintf("%.0f%%", snippet.Confidence*100)))
			b.WriteString(" ")
		}

		// Usage indicator
		usageCount := snippetUsage[snippet.Name]
		if usageCount > 0 {
			usageStyle := lipgloss.NewStyle().
				Foreground(ColorGreen).
				Italic(true)
			b.WriteString(usageStyle.Render(fmt.Sprintf("(used by %d)", usageCount)))
		} else {
			// Unused snippet - show warning
			unusedStyle := lipgloss.NewStyle().
				Foreground(ColorOrange).
				Italic(true)
			b.WriteString(unusedStyle.Render("(unused)"))
		}

		b.WriteString("\n")

		// Description (dimmed)
		descStyle := lipgloss.NewStyle().
			Foreground(ColorGray).
			Italic(true).
			PaddingLeft(4) // Extra indent for alignment
		b.WriteString(descStyle.Render(snippet.Description))
		b.WriteString("\n")

		// Location info (very dim, compact)
		locationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")).
			PaddingLeft(4)
		b.WriteString(locationStyle.Render(fmt.Sprintf("Lines %d-%d (%d lines)",
			snippet.LineStart, snippet.LineEnd, snippet.Lines())))

		// Add spacing between snippets (except last visible)
		if i < end-1 {
			b.WriteString("\n\n")
		}
	}

	// Show scroll indicator if needed
	if len(m.snippets) > visibleSnippets {
		b.WriteString("\n")
		scrollInfo := fmt.Sprintf("[Showing %d-%d of %d]", start+1, end, len(m.snippets))
		b.WriteString(StyleDim.Render(scrollInfo))
	}

	return b.String()
}

// renderSnippetsPanelTitle returns the title for the snippets panel (tab-aware)
func (m Model) renderSnippetsPanelTitle() string {
	// When on Caddy tab with a selected entry, show applied snippets count
	if m.activeTab == TabCaddy {
		filteredEntries := m.getFilteredEntries()
		if m.cursor < len(filteredEntries) {
			entry := filteredEntries[m.cursor]
			if entry.Caddy != nil && len(entry.Caddy.Imports) > 0 {
				return fmt.Sprintf("Applied Snippets (%d)", len(entry.Caddy.Imports))
			}
			return "Applied Snippets (0)"
		}
	}

	// Default: show all snippets count
	count := len(m.snippets)
	if count == 0 {
		return "Snippets (0)"
	}

	// Count by category for summary
	categoryCount := make(map[caddy.SnippetCategory]int)
	for _, snippet := range m.snippets {
		categoryCount[snippet.Category]++
	}

	// Show count with category breakdown if only 1-2 categories
	if len(categoryCount) == 1 {
		// Single category
		for cat := range categoryCount {
			return fmt.Sprintf("Snippets (%d %s)", count, cat.String())
		}
	}

	return fmt.Sprintf("Available Snippets (%d)", count)
}

// calculateSnippetUsage returns a map of snippet name → usage count across all entries
func (m Model) calculateSnippetUsage() map[string]int {
	usage := make(map[string]int)

	for _, entry := range m.entries {
		// Only count entries that have Caddy configuration
		if entry.Caddy != nil {
			for _, importedSnippet := range entry.Caddy.Imports {
				usage[importedSnippet]++
			}
		}
	}

	return usage
}

// renderSnippetDetailView renders a detailed view of the currently selected snippet
func (m Model) renderSnippetDetailView() string {
	if m.snippetPanel.Cursor >= len(m.snippets) {
		return StyleError.Render("No snippet selected")
	}

	snippet := m.snippets[m.snippetPanel.Cursor]
	usage := m.calculateSnippetUsage()
	usageCount := usage[snippet.Name]

	var b strings.Builder

	// Title with snippet name
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorBlue).
		Bold(true).
		Underline(true)
	b.WriteString(titleStyle.Render(snippet.Name))
	b.WriteString("\n\n")

	// Category badge
	b.WriteString(StyleInfo.Render("Category: "))
	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(snippet.Category.ColorCode())).
		Bold(true)
	b.WriteString(categoryStyle.Render(snippet.Category.String()))
	b.WriteString("\n\n")

	// Description
	b.WriteString(StyleInfo.Render("Description:"))
	b.WriteString("\n")
	b.WriteString(StyleDim.Render(snippet.Description))
	b.WriteString("\n\n")

	// Auto-detection info
	if snippet.AutoDetected {
		b.WriteString(StyleInfo.Render("Auto-detected: "))
		confidenceStyle := lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)
		b.WriteString(confidenceStyle.Render(fmt.Sprintf("%.1f%% confidence", snippet.Confidence*100)))
		b.WriteString("\n\n")
	}

	// Usage statistics
	b.WriteString(StyleInfo.Render("Usage: "))
	if usageCount > 0 {
		usageStyle := lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)
		b.WriteString(usageStyle.Render(fmt.Sprintf("%d entries", usageCount)))

		// List which entries use this snippet
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("  Used by: "))
		usedBy := []string{}
		for _, entry := range m.entries {
			if entry.Caddy != nil {
				for _, importedSnippet := range entry.Caddy.Imports {
					if importedSnippet == snippet.Name {
						usedBy = append(usedBy, entry.Domain)
						break
					}
				}
			}
		}
		b.WriteString(StyleDim.Render(strings.Join(usedBy, ", ")))
	} else {
		unusedStyle := lipgloss.NewStyle().
			Foreground(ColorOrange).
			Italic(true)
		b.WriteString(unusedStyle.Render("Not currently used"))
	}
	b.WriteString("\n\n")

	// Location in Caddyfile
	b.WriteString(StyleInfo.Render("Location: "))
	locationText := fmt.Sprintf("Lines %d-%d (%d lines)",
		snippet.LineStart, snippet.LineEnd, snippet.Lines())
	b.WriteString(StyleDim.Render(locationText))
	b.WriteString("\n\n")

	// Full content - show editable textarea if in edit mode, otherwise read-only
	b.WriteString(StyleInfo.Render("Content:"))
	b.WriteString("\n")

	if m.snippetPanel.Editing {
		// Edit mode: Show editable textarea
		b.WriteString(m.snippetPanel.EditTextarea.View())
		b.WriteString("\n\n")

		// Navigation hint for edit mode
		b.WriteString(StyleInfo.Render("y/Enter: save  "))
		b.WriteString(StyleError.Render("d: delete  "))
		b.WriteString(StyleDim.Render("ESC: cancel"))
	} else {
		// View mode: Show read-only content
		contentStyle := lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(1, 2).
			MarginTop(1)
		b.WriteString(contentStyle.Render(snippet.Content))
		b.WriteString("\n\n")

		// Navigation hint for view mode
		b.WriteString(StyleInfo.Render("e: edit  "))
		b.WriteString(StyleDim.Render("ESC: return"))
	}

	return b.String()
}
