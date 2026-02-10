package ui

import (
	"fmt"
	"strings"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/diff"

	"github.com/charmbracelet/lipgloss"
)

// renderPanelLayout renders the main two-panel layout
func (m Model) renderPanelLayout() string {
	// Calculate panel dimensions
	layout := NewPanelLayout(m.width, m.height)

	// Render title bar with tab indicators
	titleBar := RenderTitleBarWithTabs(m.config.Domain, int(m.activeTab), m.width)

	// Render left panel (entry list)
	leftContent := m.renderLeftPanel(layout.LeftWidth, layout.LeftHeight)
	leftPanel := RenderPanel(
		m.getLeftPanelTitle(),
		leftContent,
		layout.LeftWidth,
		layout.LeftHeight,
		m.panelFocus == PanelFocusLeft,
	)

	// Render primary panel (top-right) - shows current tab's main info
	primaryContent := m.renderPrimaryPanel(layout.DetailsWidth, layout.DetailsHeight)
	primaryPanel := RenderPanel(
		m.getPrimaryPanelTitle(),
		primaryContent,
		layout.DetailsWidth,
		layout.DetailsHeight,
		m.panelFocus == PanelFocusDetails,
	)

	// Render complementary panel (bottom-right) - shows the OTHER tab's info
	complementaryContent := m.renderComplementaryPanel(layout.SnippetsHeight)
	complementaryPanel := RenderPanel(
		m.getComplementaryPanelTitle(),
		complementaryContent,
		layout.SnippetsWidth,
		layout.SnippetsHeight,
		false, // Never focused
	)

	// Combine all three panels
	panels := RenderThreePanelLayout(leftPanel, primaryPanel, complementaryPanel)

	// Render status bar
	statusBar := RenderStatusBar(m.getStatusBarContent(), m.width, m.err != nil)

	// Combine everything
	return RenderFullLayout(titleBar, panels, statusBar)
}

// getLeftPanelTitle returns the title for the left panel
func (m Model) getLeftPanelTitle() string {
	filteredEntries := m.getFilteredEntries()
	return fmt.Sprintf("Entries (%d)", len(filteredEntries))
}

// getDetailsPanelTitle returns the title for the details panel (top-right)
func (m Model) getDetailsPanelTitle() string {
	filteredEntries := m.getFilteredEntries()
	if m.cursor >= len(filteredEntries) {
		return "Welcome"
	}
	// Show tab-specific title
	if m.activeTab == TabCaddy {
		return "Caddy Config"
	}
	return "DNS Details"
}

// getPrimaryPanelTitle returns title for top-right panel (current tab's main view)
func (m Model) getPrimaryPanelTitle() string {
	filteredEntries := m.getFilteredEntries()
	if m.cursor >= len(filteredEntries) {
		return "Details"
	}
	if m.activeTab == TabCaddy {
		return "Caddy Config"
	}
	return "DNS Record"
}

// getComplementaryPanelTitle returns title for bottom-right panel (other tab's info)
func (m Model) getComplementaryPanelTitle() string {
	filteredEntries := m.getFilteredEntries()
	if m.cursor >= len(filteredEntries) {
		return "Related"
	}
	// Show the OPPOSITE of current tab
	if m.activeTab == TabCaddy {
		return "DNS Record"
	}
	return "Caddy Config"
}

// renderPrimaryPanel renders the main info for current tab (top-right)
func (m Model) renderPrimaryPanel(width, height int) string {
	filteredEntries := m.getFilteredEntries()
	if len(filteredEntries) == 0 {
		return StyleDim.Render("No entries to display.\n\nPress 'a' to add a new entry.")
	}
	if m.cursor >= len(filteredEntries) {
		return StyleDim.Render("Select an entry to view details.")
	}

	entry := filteredEntries[m.cursor]

	// Show current tab's details
	if m.activeTab == TabCaddy {
		return m.renderCaddyDetails(entry)
	}
	return m.renderCloudflareDetails(entry)
}

// renderComplementaryPanel renders the OTHER tab's info (bottom-right)
func (m Model) renderComplementaryPanel(height int) string {
	filteredEntries := m.getFilteredEntries()
	if len(filteredEntries) == 0 || m.cursor >= len(filteredEntries) {
		return StyleDim.Render("Select an entry to see related info.")
	}

	entry := filteredEntries[m.cursor]

	// Show the OPPOSITE of current tab
	if m.activeTab == TabCaddy {
		// On Caddy tab, show full DNS info in bottom panel
		if entry.DNS == nil {
			return StyleDim.Render("No DNS record for this entry.\n\nPress 's' to sync and create DNS.")
		}
		return m.renderCloudflareDetails(entry)
	}
	// On Cloudflare tab, show full Caddy info in bottom panel
	if entry.Caddy == nil {
		return StyleDim.Render("No Caddy config for this entry.\n\nEdit to add reverse proxy settings.")
	}
	return m.renderCaddyDetails(entry)
}

// renderLeftPanel renders the entry list for the left panel
func (m Model) renderLeftPanel(width, height int) string {
	var b strings.Builder

	// Get filtered entries
	displayEntries := m.getFilteredEntries()

	// Show filter/sort info if active
	filterParts := []string{}
	if m.statusFilter != FilterAll {
		filterParts = append(filterParts, fmt.Sprintf("Status: %s", m.statusFilter.String()))
	}
	if m.dnsTypeFilter != DNSTypeAll {
		filterParts = append(filterParts, fmt.Sprintf("Type: %s", m.dnsTypeFilter.String()))
	}
	if m.searchQuery != "" {
		filterParts = append(filterParts, fmt.Sprintf("Search: \"%s\"", m.searchQuery))
	}

	if len(filterParts) > 0 {
		b.WriteString(StyleDim.Render(fmt.Sprintf("Filter: %s", strings.Join(filterParts, ", "))))
		b.WriteString("\n")
	}

	if m.sortMode != SortAlphabetical {
		b.WriteString(StyleDim.Render(fmt.Sprintf("Sort: %s", m.sortMode.String())))
		b.WriteString("\n")
	}

	if len(m.selectedEntries) > 0 {
		b.WriteString(StyleInfo.Render(fmt.Sprintf("Selected: %d", len(m.selectedEntries))))
		b.WriteString("\n")
	}

	// Calculate visible range
	visibleHeight := height - 6 // Leave space for borders, title, filters
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	start := m.scrollOffset
	end := start + visibleHeight
	if end > len(displayEntries) {
		end = len(displayEntries)
	}

	// Entry list
	for i := start; i < end; i++ {
		entry := displayEntries[i]

		// Selection checkbox
		checkbox := "[ ]"
		if m.selectedEntries[entry.Domain] {
			checkbox = lipgloss.NewStyle().Foreground(ColorBlue).Bold(true).Render("[✓]")
		} else {
			checkbox = StyleDim.Render(checkbox)
		}

		// Icon based on status
		var icon string
		if entry.Status == diff.StatusSynced {
			icon = StyleIconSynced.Render("✓")
		} else {
			icon = StyleIconOrphan.Render("⚠")
		}

		// Domain name - show multi-domain format if applicable
		var domain string
		if entry.Caddy != nil && len(entry.Caddy.Domains) > 0 {
			domain = caddy.FormatDomainDisplay(entry.Caddy)
		} else {
			domain = entry.Domain
		}

		// Truncate if needed
		maxDomainLen := width - 10 // Account for checkbox, icon, padding
		if maxDomainLen < 10 {
			maxDomainLen = 10 // Minimum width
		}
		if len(domain) > maxDomainLen && maxDomainLen > 3 {
			domain = domain[:maxDomainLen-3] + "..."
		}

		// Build line with cursor indicator
		var line string
		if i == m.cursor {
			// Add visual cursor indicator and highlight
			line = StyleHighlight.Render(fmt.Sprintf("→ %s %s %s", checkbox, icon, domain))
		} else {
			// Normal line with spacing to align with cursor indicator
			line = fmt.Sprintf("  %s %s %s", checkbox, icon, domain)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(displayEntries) > visibleHeight {
		b.WriteString("\n")
		b.WriteString(StyleDim.Render(fmt.Sprintf("[%d-%d of %d]", start+1, end, len(displayEntries))))
	}

	return b.String()
}

// renderRightPanel renders the details/context panel (tab-aware)
func (m Model) renderRightPanel(width, height int) string {
	filteredEntries := m.getFilteredEntries()

	// If no entry selected or list is empty, show welcome message
	if len(filteredEntries) == 0 {
		return StyleDim.Render("No entries to display.\n\nPress 'a' to add a new entry.")
	}

	if m.cursor >= len(filteredEntries) {
		return StyleDim.Render("Select an entry to view details.\n\nUse ↑/↓ to navigate.")
	}

	// Get selected entry
	entry := filteredEntries[m.cursor]

	// Render based on active tab
	if m.activeTab == TabCaddy {
		return m.renderCaddyDetails(entry)
	}
	return m.renderCloudflareDetails(entry)
}

// renderCloudflareDetails renders DNS/Cloudflare specific information
func (m Model) renderCloudflareDetails(entry diff.SyncedEntry) string {
	var b strings.Builder

	// Domain header
	b.WriteString(StyleTitleFocused.Render(entry.Domain))
	b.WriteString("\n\n")

	// Sync status
	var statusLine string
	if entry.Status == diff.StatusSynced {
		statusLine = StyleIconSynced.Render("✓ Synced with Caddy")
	} else if entry.Status == diff.StatusOrphanedDNS {
		statusLine = StyleWarning.Render("⚠ DNS only (no Caddy config)")
	} else if entry.Status == diff.StatusOrphanedCaddy {
		statusLine = StyleDim.Render("○ No DNS record")
	}
	b.WriteString(statusLine)
	b.WriteString("\n\n")

	// DNS Information
	if entry.DNS != nil {
		b.WriteString(StyleInfo.Render("DNS Record"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Type:    %s\n", StyleKeybinding.Render(entry.DNS.Type)))
		b.WriteString(fmt.Sprintf("  Target:  %s\n", entry.DNS.Content))

		// Proxied status with color
		proxiedStr := "No"
		if entry.DNS.Proxied {
			proxiedStr = StyleSuccess.Render("Yes (CF Proxy)")
		}
		b.WriteString(fmt.Sprintf("  Proxied: %s\n", proxiedStr))

		// TTL info
		b.WriteString(fmt.Sprintf("  TTL:     %s\n", "Auto"))
		b.WriteString("\n")
	} else {
		b.WriteString(StyleDim.Render("No DNS record configured"))
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("Press 's' to sync and create DNS record"))
		b.WriteString("\n\n")
	}

	// Quick summary of Caddy if it exists (for context)
	if entry.Caddy != nil {
		b.WriteString(StyleDim.Render("─────────────────────────────"))
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("Caddy: "))
		scheme := "http://"
		if entry.Caddy.SSL {
			scheme = "https://"
		}
		b.WriteString(StyleDim.Render(fmt.Sprintf("%s%s:%d", scheme, entry.Caddy.Target, entry.Caddy.Port)))
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("(press Tab for Caddy details)"))
	}

	return b.String()
}

// renderCaddyDetails renders Caddy/reverse proxy specific information
func (m Model) renderCaddyDetails(entry diff.SyncedEntry) string {
	var b strings.Builder

	// Domain header (show multi-domain if applicable)
	if entry.Caddy != nil && caddy.GetDomainCount(entry.Caddy) > 1 {
		b.WriteString(StyleTitleFocused.Render("Domains:"))
		b.WriteString("\n")
		for i, domain := range entry.Caddy.Domains {
			if i == 0 {
				b.WriteString(StyleInfo.Render("  " + domain + " (primary)"))
			} else {
				b.WriteString(StyleDim.Render("  " + domain))
			}
			b.WriteString("\n")
		}
	} else {
		b.WriteString(StyleTitleFocused.Render(entry.Domain))
	}
	b.WriteString("\n")

	// Caddy config status
	if entry.Caddy == nil {
		b.WriteString(StyleWarning.Render("⚠ No Caddy configuration"))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("This entry has a DNS record but no reverse proxy config."))
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("Press 'e' to edit and add Caddy configuration."))
		return b.String()
	}

	// Reverse proxy target
	b.WriteString(StyleInfo.Render("Reverse Proxy"))
	b.WriteString("\n")
	scheme := "http://"
	if entry.Caddy.SSL {
		scheme = "https://"
	}
	targetURL := fmt.Sprintf("%s%s:%d", scheme, entry.Caddy.Target, entry.Caddy.Port)
	b.WriteString(fmt.Sprintf("  Target: %s\n", StyleKeybinding.Render(targetURL)))

	// Target validity check (basic)
	b.WriteString("  Status: ")
	// TODO: Add actual health check
	b.WriteString(StyleDim.Render("configured"))
	b.WriteString("\n\n")

	// Restrictions
	if entry.Caddy.IPRestricted {
		b.WriteString(StyleWarning.Render("⚠ IP Restricted"))
		b.WriteString("\n\n")
	}

	// Applied snippets with details
	if len(entry.Caddy.Imports) > 0 {
		b.WriteString(StyleInfo.Render("Applied Snippets"))
		b.WriteString("\n")
		for _, importName := range entry.Caddy.Imports {
			// Find snippet for category color
			var snippetCat caddy.SnippetCategory
			for _, s := range m.snippets {
				if s.Name == importName {
					snippetCat = s.Category
					break
				}
			}
			categoryColor := snippetCat.ColorCode()
			badge := lipgloss.NewStyle().
				Foreground(lipgloss.Color(categoryColor)).
				Bold(true).
				Render("● " + importName)
			b.WriteString("  " + badge + "\n")
		}
		b.WriteString("\n")
	}

	// Caddy Block Preview (truncated)
	if entry.Caddy.RawBlock != "" {
		b.WriteString(StyleInfo.Render("Caddy Block Preview"))
		b.WriteString("\n")

		// Truncate if too long
		block := entry.Caddy.RawBlock
		lines := strings.Split(block, "\n")
		if len(lines) > 8 {
			lines = lines[:8]
			lines = append(lines, "  ...")
		}
		block = strings.Join(lines, "\n")

		blockStyle := lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

		b.WriteString(blockStyle.Render(block))
		b.WriteString("\n")
	}

	return b.String()
}

// formatKeybinding formats a keybinding with color
func formatKeybinding(key, description string) string {
	return fmt.Sprintf("%s:%s", StyleKeybinding.Render(key), description)
}

// getStatusBarContent returns the content for the status bar
func (m Model) getStatusBarContent() string {
	if m.loading {
		return StyleInfo.Render("⟳ Loading...")
	}

	// Errors are now shown in a modal, not in status bar
	if m.err != nil {
		return StyleDim.Render("Error occurred - see modal above")
	}

	if m.searching {
		return fmt.Sprintf("Search: %s_ (%s to accept, %s to cancel)",
			m.searchQuery,
			StyleKeybinding.Render("enter"),
			StyleKeybinding.Render("esc"))
	}

	// Context-sensitive keybindings
	if len(m.selectedEntries) > 0 {
		return fmt.Sprintf("Navigate: %s  Select: %s  Batch: %s %s  Clear: %s",
			StyleKeybinding.Render("↑↓"),
			StyleKeybinding.Render("space"),
			formatKeybinding("X", "delete"),
			formatKeybinding("S", "sync"),
			StyleKeybinding.Render("esc"))
	}

	// Tab indicator
	tabHint := ""
	if m.activeTab == TabCloudflare {
		tabHint = StyleDim.Render("[DNS]")
	} else {
		tabHint = StyleDim.Render("[Caddy]")
	}

	editorHint := ""
	if m.activeTab == TabCaddy {
		editorHint = formatKeybinding("E", "editor") + " "
	}

	return fmt.Sprintf("%s %s %s %s %s %s %s%s %s %s %s %s %s %s",
		tabHint,
		StyleKeybinding.Render("↑↓")+" nav",
		StyleKeybinding.Render("tab")+" view",
		formatKeybinding("a", "add"),
		formatKeybinding("⏎", "edit"),
		formatKeybinding("d", "del"),
		editorHint,
		formatKeybinding("w", "snippets"),
		formatKeybinding("p", "profile"),
		formatKeybinding("/", "search"),
		formatKeybinding("r", "refresh"),
		formatKeybinding("b", "backups"),
		formatKeybinding("?", "help"),
		formatKeybinding("q", "quit"))
}

// renderHelpModalContent returns help content for modal display (paginated)
func (m Model) renderHelpModalContent() string {
	var content strings.Builder
	const totalPages = 5

	// Page title
	pageTitle := fmt.Sprintf("Help — Page %d/%d", m.helpPage+1, totalPages)
	content.WriteString(StyleInfo.Render(pageTitle))
	content.WriteString("\n\n")

	// Render content based on current page
	switch m.helpPage {
	case 0: // Page 1: Quick Reference
		content.WriteString(m.renderHelpPage1())
	case 1: // Page 2: Navigation & Entry Operations
		content.WriteString(m.renderHelpPage2())
	case 2: // Page 3: Forms & Batch Operations
		content.WriteString(m.renderHelpPage3())
	case 3: // Page 4: Tools
		content.WriteString(m.renderHelpPage4())
	case 4: // Page 5: Wizard & Profile
		content.WriteString(m.renderHelpPage5())
	}

	// Footer with navigation info
	content.WriteString("\n")
	content.WriteString(StyleDim.Render("────────────────────────────────────────────"))
	content.WriteString("\n")
	footer := fmt.Sprintf("Page %d/%d  %s/%s navigate  %s close  %s-%s jump",
		m.helpPage+1, totalPages,
		StyleKeybinding.Render("←"),
		StyleKeybinding.Render("→"),
		StyleKeybinding.Render("Esc"),
		StyleKeybinding.Render("1"),
		StyleKeybinding.Render("5"))
	content.WriteString(footer)

	return content.String()
}

// renderHelpPage1 returns Page 1: Quick Reference (Keyboard Shortcuts)
func (m Model) renderHelpPage1() string {
	var b strings.Builder

	// Left column
	var left strings.Builder
	left.WriteString(StyleInfo.Render("Navigation"))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("  %s  Move down/up\n", StyleKeybinding.Render("↑↓ j/k")))
	left.WriteString(fmt.Sprintf("  %s  Top / Bottom\n", StyleKeybinding.Render("g / G")))
	left.WriteString(fmt.Sprintf("  %s  Switch tab\n", StyleKeybinding.Render("Tab")))
	left.WriteString(fmt.Sprintf("  %s  Switch panel\n", StyleKeybinding.Render("Shift+Tab")))
	left.WriteString("\n")

	left.WriteString(StyleInfo.Render("Quick Actions"))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("  %s  Add entry\n", StyleKeybinding.Render("a")))
	left.WriteString(fmt.Sprintf("  %s  Edit entry\n", StyleKeybinding.Render("Enter")))
	left.WriteString(fmt.Sprintf("  %s  Delete entry\n", StyleKeybinding.Render("d")))
	left.WriteString(fmt.Sprintf("  %s  Sync entry\n", StyleKeybinding.Render("s")))
	left.WriteString(fmt.Sprintf("  %s  Search\n", StyleKeybinding.Render("/")))

	// Right column
	var right strings.Builder
	right.WriteString(StyleInfo.Render("Tools"))
	right.WriteString("\n")
	right.WriteString(fmt.Sprintf("  %s  Snippet wizard\n", StyleKeybinding.Render("w")))
	right.WriteString(fmt.Sprintf("  %s  Backup manager\n", StyleKeybinding.Render("b")))
	right.WriteString(fmt.Sprintf("  %s  Audit log\n", StyleKeybinding.Render("l")))
	right.WriteString(fmt.Sprintf("  %s  Profile selector\n", StyleKeybinding.Render("p")))
	right.WriteString(fmt.Sprintf("  %s  Refresh data\n", StyleKeybinding.Render("r")))
	right.WriteString("\n")

	right.WriteString(StyleInfo.Render("General"))
	right.WriteString("\n")
	right.WriteString(fmt.Sprintf("  %s  Help (this screen)\n", StyleKeybinding.Render("?")))
	right.WriteString(fmt.Sprintf("  %s  Quit\n", StyleKeybinding.Render("q / Ctrl+Q")))
	right.WriteString(fmt.Sprintf("  %s  Force quit\n", StyleKeybinding.Render("Ctrl+C")))

	// Combine columns
	leftBox := lipgloss.NewStyle().Width(28).Render(left.String())
	rightBox := lipgloss.NewStyle().Width(28).Render(right.String())
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	b.WriteString(columns)
	return b.String()
}

// renderHelpPage2 returns Page 2: Navigation & Entry Operations
func (m Model) renderHelpPage2() string {
	var b strings.Builder

	var left strings.Builder
	left.WriteString(StyleInfo.Render("List Navigation"))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("  %s  Down/Up\n", StyleKeybinding.Render("j/k ↓↑")))
	left.WriteString(fmt.Sprintf("  %s  Top/Bottom\n", StyleKeybinding.Render("g/G")))
	left.WriteString(fmt.Sprintf("  %s  Page Up/Down\n", StyleKeybinding.Render("PgUp/PgDn")))
	left.WriteString(fmt.Sprintf("  %s  Home/End\n", StyleKeybinding.Render("Home/End")))
	left.WriteString("\n")

	left.WriteString(StyleInfo.Render("Entry Operations"))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("  %s  Add entry\n", StyleKeybinding.Render("a")))
	left.WriteString(fmt.Sprintf("  %s  Edit entry\n", StyleKeybinding.Render("Enter")))
	left.WriteString(fmt.Sprintf("  %s  Delete entry\n", StyleKeybinding.Render("d")))
	left.WriteString(fmt.Sprintf("  %s  Sync orphaned\n", StyleKeybinding.Render("s")))

	var right strings.Builder
	right.WriteString(StyleInfo.Render("View Control"))
	right.WriteString("\n")
	right.WriteString(fmt.Sprintf("  %s  Cloudflare/Caddy\n", StyleKeybinding.Render("Tab")))
	right.WriteString(fmt.Sprintf("  %s  Panel focus\n", StyleKeybinding.Render("Shift+Tab")))
	right.WriteString("\n")

	right.WriteString(StyleInfo.Render("Search & Filter"))
	right.WriteString("\n")
	right.WriteString(fmt.Sprintf("  %s  Search\n", StyleKeybinding.Render("/")))
	right.WriteString(fmt.Sprintf("  %s  Status\n", StyleKeybinding.Render("f")))
	right.WriteString(fmt.Sprintf("  %s  DNS type\n", StyleKeybinding.Render("t")))
	right.WriteString(fmt.Sprintf("  %s  Sort\n", StyleKeybinding.Render("o")))

	leftBox := lipgloss.NewStyle().Width(28).Render(left.String())
	rightBox := lipgloss.NewStyle().Width(28).Render(right.String())
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	b.WriteString(columns)
	return b.String()
}

// renderHelpPage3 returns Page 3: Forms & Batch Operations
func (m Model) renderHelpPage3() string {
	var b strings.Builder

	var left strings.Builder
	left.WriteString(StyleInfo.Render("Form Navigation"))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("  %s  Next/Prev field\n", StyleKeybinding.Render("Tab/Shift+Tab")))
	left.WriteString(fmt.Sprintf("  %s  Next/Prev field\n", StyleKeybinding.Render("↓/↑")))
	left.WriteString(fmt.Sprintf("  %s  Newline (multi)\n", StyleKeybinding.Render("Ctrl+M")))
	left.WriteString(fmt.Sprintf("  %s  Next/Preview\n", StyleKeybinding.Render("Enter")))
	left.WriteString(fmt.Sprintf("  %s  Toggle checkbox\n", StyleKeybinding.Render("Space")))
	left.WriteString(fmt.Sprintf("  %s  Cancel form\n", StyleKeybinding.Render("Esc")))

	var right strings.Builder
	right.WriteString(StyleInfo.Render("Batch Operations"))
	right.WriteString("\n")
	right.WriteString(fmt.Sprintf("  %s  Toggle selection\n", StyleKeybinding.Render("Space")))
	right.WriteString(fmt.Sprintf("  %s  Delete selected\n", StyleKeybinding.Render("X")))
	right.WriteString(fmt.Sprintf("  %s  Sync selected\n", StyleKeybinding.Render("S")))
	right.WriteString(fmt.Sprintf("  %s  Bulk delete menu\n", StyleKeybinding.Render("D")))
	right.WriteString(fmt.Sprintf("  %s  Clear selection\n", StyleKeybinding.Render("Esc")))

	leftBox := lipgloss.NewStyle().Width(28).Render(left.String())
	rightBox := lipgloss.NewStyle().Width(28).Render(right.String())
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	b.WriteString(columns)
	return b.String()
}

// renderHelpPage4 returns Page 4: Tools
func (m Model) renderHelpPage4() string {
	var b strings.Builder

	var left strings.Builder
	left.WriteString(StyleInfo.Render("Tools Menu"))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("  %s  Snippet wizard\n", StyleKeybinding.Render("w / Ctrl+S")))
	left.WriteString(fmt.Sprintf("  %s  Backup manager\n", StyleKeybinding.Render("b / Ctrl+B")))
	left.WriteString(fmt.Sprintf("  %s  Audit log\n", StyleKeybinding.Render("l")))
	left.WriteString(fmt.Sprintf("  %s  Profile selector\n", StyleKeybinding.Render("p / Ctrl+P")))
	left.WriteString(fmt.Sprintf("  %s  Refresh data\n", StyleKeybinding.Render("r")))
	left.WriteString(fmt.Sprintf("  %s  Migrate Caddyfile\n", StyleKeybinding.Render("m")))

	var right strings.Builder
	right.WriteString(StyleInfo.Render("Tool Operations"))
	right.WriteString("\n")
	right.WriteString(fmt.Sprintf("  %s  Navigate menu\n", StyleKeybinding.Render("↓/↑")))
	right.WriteString(fmt.Sprintf("  %s  Select/preview\n", StyleKeybinding.Render("Enter")))
	right.WriteString(fmt.Sprintf("  %s  Delete/Restore\n", StyleKeybinding.Render("x/R")))
	right.WriteString(fmt.Sprintf("  %s  Cleanup old\n", StyleKeybinding.Render("c")))
	right.WriteString(fmt.Sprintf("  %s  Close tool\n", StyleKeybinding.Render("Esc")))

	leftBox := lipgloss.NewStyle().Width(28).Render(left.String())
	rightBox := lipgloss.NewStyle().Width(28).Render(right.String())
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	b.WriteString(columns)
	return b.String()
}

// renderHelpPage5 returns Page 5: Wizard & Profile
func (m Model) renderHelpPage5() string {
	var b strings.Builder

	var left strings.Builder
	left.WriteString(StyleInfo.Render("Snippet Wizard"))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("  %s  Navigate\n", StyleKeybinding.Render("↓/↑")))
	left.WriteString(fmt.Sprintf("  %s  Toggle option\n", StyleKeybinding.Render("Space")))
	left.WriteString(fmt.Sprintf("  %s  Next step\n", StyleKeybinding.Render("Enter")))
	left.WriteString(fmt.Sprintf("  %s  Back / Cancel\n", StyleKeybinding.Render("Esc")))
	left.WriteString("\n")
	left.WriteString(StyleInfo.Render("Snippet Editing"))
	left.WriteString("\n")
	left.WriteString(fmt.Sprintf("  %s  Edit snippet\n", StyleKeybinding.Render("e")))
	left.WriteString(fmt.Sprintf("  %s  Save changes\n", StyleKeybinding.Render("y")))
	left.WriteString(fmt.Sprintf("  %s  Delete snippet\n", StyleKeybinding.Render("d")))

	var right strings.Builder
	right.WriteString(StyleInfo.Render("Setup Wizard"))
	right.WriteString("\n")
	right.WriteString(fmt.Sprintf("  %s  Navigate fields\n", StyleKeybinding.Render("Tab / ↓↑")))
	right.WriteString(fmt.Sprintf("  %s  Next step\n", StyleKeybinding.Render("Enter")))
	right.WriteString(fmt.Sprintf("  %s  Back / Cancel\n", StyleKeybinding.Render("Esc")))
	right.WriteString("\n")
	right.WriteString(StyleInfo.Render("Anywhere"))
	right.WriteString("\n")
	right.WriteString(fmt.Sprintf("  %s  Close modal\n", StyleKeybinding.Render("Esc / Ctrl+W")))
	right.WriteString(fmt.Sprintf("  %s  Help\n", StyleKeybinding.Render("? / Ctrl+H")))

	leftBox := lipgloss.NewStyle().Width(28).Render(left.String())
	rightBox := lipgloss.NewStyle().Width(28).Render(right.String())
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	b.WriteString(columns)
	return b.String()
}

// renderHelpView renders the help screen (wrapper for modal content)
func (m Model) renderHelpView() string {
	return m.renderHelpModalContent()
}

// renderErrorModalContent renders full error text with word wrapping
func (m Model) renderErrorModalContent() string {
	if m.err == nil {
		return "No error to display."
	}

	var b strings.Builder

	// Error message with wrapping
	errorMsg := m.err.Error()
	b.WriteString(StyleError.Render("Error Details:"))
	b.WriteString("\n\n")

	// Wrap text to ~60 chars for readability
	const maxLineWidth = 60
	words := strings.Fields(errorMsg)
	var line strings.Builder

	for i, word := range words {
		// Check if adding word would exceed line width
		testLine := line.String()
		if len(testLine) > 0 {
			testLine += " "
		}
		testLine += word

		if len(testLine) > maxLineWidth && line.Len() > 0 {
			// Write current line and start new one
			b.WriteString(line.String())
			b.WriteString("\n")
			line.Reset()
			line.WriteString(word)
		} else {
			// Add word to current line
			if line.Len() > 0 {
				line.WriteString(" ")
			}
			line.WriteString(word)
		}

		// Add extra space between sentences for readability
		if i < len(words)-1 && (word[len(word)-1] == '.' || word[len(word)-1] == ':') {
			testLine = line.String()
			if len(testLine) > maxLineWidth {
				b.WriteString(line.String())
				b.WriteString("\n")
				line.Reset()
			}
		}
	}

	// Write remaining content
	if line.Len() > 0 {
		b.WriteString(line.String())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(StyleDim.Render("Press Enter or Esc to dismiss"))

	return b.String()
}
