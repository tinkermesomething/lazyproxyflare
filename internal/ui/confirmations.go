package ui

import (
	"fmt"
	"strings"

	"lazyproxyflare/internal/diff"

	"github.com/charmbracelet/lipgloss"
)

// renderConfirmDeleteView renders the delete confirmation screen
func (m Model) renderConfirmDeleteView() string {
	var b strings.Builder

	// Get the entry to delete
	filteredEntries := m.getFilteredEntries()
	if m.deleteEntryIndex >= len(filteredEntries) {
		return "Error: Invalid entry index"
	}
	entry := filteredEntries[m.deleteEntryIndex]

	// Show what will be deleted
	b.WriteString("You are about to delete:\n\n")

	// Entry box
	entryBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorRed).
		Padding(0, 1).
		Width(70)

	entryContent := strings.Builder{}
	entryContent.WriteString(StyleInfo.Render(fmt.Sprintf("Domain: %s", entry.Domain)))
	entryContent.WriteString("\n")

	// Show DNS info if exists
	if entry.DNS != nil {
		entryContent.WriteString("\n")
		entryContent.WriteString("DNS Record (Cloudflare):\n")
		entryContent.WriteString(fmt.Sprintf("  Type:     %s\n", entry.DNS.Type))
		entryContent.WriteString(fmt.Sprintf("  Name:     %s\n", entry.DNS.Name))
		entryContent.WriteString(fmt.Sprintf("  Content:  %s\n", entry.DNS.Content))
		entryContent.WriteString(fmt.Sprintf("  Proxied:  %v\n", entry.DNS.Proxied))
		entryContent.WriteString(fmt.Sprintf("  ID:       %s\n", entry.DNS.ID))
	}

	// Show Caddy info if exists
	if entry.Caddy != nil {
		entryContent.WriteString("\n")
		entryContent.WriteString("Caddy Configuration:\n")
		entryContent.WriteString(fmt.Sprintf("  Target:   %s:%d\n", entry.Caddy.Target, entry.Caddy.Port))
		entryContent.WriteString(fmt.Sprintf("  SSL:      %v\n", entry.Caddy.SSL))
		if entry.Caddy.IPRestricted {
			entryContent.WriteString("  Features: IP Restricted\n")
		}
		if entry.Caddy.OAuthHeaders {
			entryContent.WriteString("  Features: OAuth Headers\n")
		}
		if entry.Caddy.WebSocket {
			entryContent.WriteString("  Features: WebSocket\n")
		}
	}

	b.WriteString(entryBox.Render(entryContent.String()))
	b.WriteString("\n\n")

	// Show what will be deleted
	var deleteMessage string
	if entry.DNS != nil && entry.Caddy != nil {
		deleteMessage = StyleWarning.Render("⚠ This will delete the entry from BOTH Cloudflare DNS and Caddyfile")
	} else if entry.DNS != nil {
		deleteMessage = StyleWarning.Render("⚠ This will delete the DNS record from Cloudflare only")
	} else if entry.Caddy != nil {
		deleteMessage = StyleWarning.Render("⚠ This will delete the Caddy entry from Caddyfile only")
	}
	b.WriteString(deleteMessage)
	b.WriteString("\n\n")

	// Status/Error display
	if m.loading {
		b.WriteString(StyleInfo.Render("⟳ Deleting entry..."))
	} else if m.err != nil {
		b.WriteString(StyleError.Render(formatErrorForDisplay(m.err, 66)))
	} else {
		b.WriteString("Are you sure? This action cannot be undone.")
	}
	b.WriteString("\n\n")

	// Instructions
	if !m.loading {
		b.WriteString(StyleDim.Render("Confirm: y  Cancel: n/esc"))
	}

	return b.String()
}

// renderConfirmSyncView renders the sync confirmation screen
func (m Model) renderConfirmSyncView() string {
	var b strings.Builder

	// Get the entry to sync (stored when 's' was pressed)
	if m.syncEntry == nil {
		return "Error: No entry selected for sync"
	}
	entry := *m.syncEntry

	// Show what will be synced
	b.WriteString("You are about to sync:\n\n")

	// Entry box
	entryBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBlue).
		Padding(0, 1).
		Width(70)

	entryContent := strings.Builder{}
	entryContent.WriteString(StyleInfo.Render(fmt.Sprintf("Domain: %s", entry.Domain)))
	entryContent.WriteString("\n")

	// Show current state
	entryContent.WriteString("\n")
	entryContent.WriteString("Current State:\n")

	if entry.DNS != nil {
		entryContent.WriteString(fmt.Sprintf("  DNS:   %s → %s\n", entry.DNS.Type, entry.DNS.Content))
	} else {
		entryContent.WriteString("  DNS:   (missing)\n")
	}

	if entry.Caddy != nil {
		entryContent.WriteString(fmt.Sprintf("  Caddy: %s:%d\n", entry.Caddy.Target, entry.Caddy.Port))
	} else {
		entryContent.WriteString("  Caddy: (missing)\n")
	}

	// Show what will be created
	entryContent.WriteString("\n")
	if entry.Status == diff.StatusOrphanedDNS {
		entryContent.WriteString(StyleSuccess.Render("Will create Caddy entry:\n"))
		entryContent.WriteString(fmt.Sprintf("  Target: localhost:%d\n", m.config.Defaults.Port))
		entryContent.WriteString(fmt.Sprintf("  SSL:    %v\n", m.config.Defaults.SSL))
		entryContent.WriteString("  (using defaults from config)\n")
	} else if entry.Status == diff.StatusOrphanedCaddy {
		entryContent.WriteString(StyleSuccess.Render("Will create DNS record:\n"))
		entryContent.WriteString(fmt.Sprintf("  Type:    CNAME\n"))
		entryContent.WriteString(fmt.Sprintf("  Target:  %s\n", m.config.Defaults.CNAMETarget))
		entryContent.WriteString(fmt.Sprintf("  Proxied: %v\n", m.config.Defaults.Proxied))
		entryContent.WriteString("  (using defaults from config)\n")
	}

	b.WriteString(entryBox.Render(entryContent.String()))
	b.WriteString("\n\n")

	// Status/Error display
	if m.loading {
		b.WriteString(StyleInfo.Render("⟳ Syncing entry..."))
	} else if m.err != nil {
		b.WriteString(StyleError.Render(formatErrorForDisplay(m.err, 66)))
	}
	b.WriteString("\n\n")

	// Instructions
	if !m.loading {
		b.WriteString(StyleDim.Render("Confirm: y  Cancel: n/esc"))
	}

	return b.String()
}

// renderBulkDeleteMenu renders the bulk delete menu
func (m Model) renderBulkDeleteMenu() string {
	var b strings.Builder

	// Count orphaned entries
	orphanedDNSCount := 0
	orphanedCaddyCount := 0
	for _, entry := range m.entries {
		if entry.Status == diff.StatusOrphanedDNS {
			orphanedDNSCount++
		} else if entry.Status == diff.StatusOrphanedCaddy {
			orphanedCaddyCount++
		}
	}

	// Menu options
	b.WriteString("Select bulk delete operation:\n\n")

	// Option 1: Delete all orphaned DNS
	option1Style := normalStyle
	if m.bulkDeleteMenuCursor == 0 {
		option1Style = StyleSelected
	}
	b.WriteString(option1Style.Render(fmt.Sprintf("  Delete all orphaned DNS records (%d entries)", orphanedDNSCount)))
	b.WriteString("\n")

	// Option 2: Delete all orphaned Caddy
	option2Style := normalStyle
	if m.bulkDeleteMenuCursor == 1 {
		option2Style = StyleSelected
	}
	b.WriteString(option2Style.Render(fmt.Sprintf("  Delete all orphaned Caddy entries (%d entries)", orphanedCaddyCount)))
	b.WriteString("\n\n")

	// Help text
	b.WriteString(StyleDim.Render("Orphaned DNS: DNS records without corresponding Caddy entries"))
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("Orphaned Caddy: Caddy entries without corresponding DNS records"))
	b.WriteString("\n\n")

	// Instructions
	b.WriteString(StyleDim.Render("Navigate: ↑/↓  Select: enter  Cancel: esc"))

	return b.String()
}

// renderConfirmBulkDeleteView renders the bulk delete confirmation screen
func (m Model) renderConfirmBulkDeleteView() string {
	var b strings.Builder

	// Warning
	b.WriteString(StyleError.Render(fmt.Sprintf("⚠ WARNING: You are about to delete %d entries!", len(m.bulkDeleteEntries))))
	b.WriteString("\n\n")

	// List entries to be deleted
	b.WriteString("The following entries will be deleted:\n\n")

	// Create a box for the list
	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorRed).
		Padding(0, 1).
		Width(70).
		Height(15)

	listContent := strings.Builder{}
	maxDisplay := 10
	for i, entry := range m.bulkDeleteEntries {
		if i >= maxDisplay {
			listContent.WriteString(StyleDim.Render(fmt.Sprintf("\n... and %d more entries", len(m.bulkDeleteEntries)-maxDisplay)))
			break
		}
		if m.bulkDeleteType == "dns" {
			listContent.WriteString(fmt.Sprintf("%s (DNS: %s → %s)\n", entry.Domain, entry.DNS.Type, entry.DNS.Content))
		} else {
			listContent.WriteString(fmt.Sprintf("%s (Caddy: %s:%d)\n", entry.Domain, entry.Caddy.Target, entry.Caddy.Port))
		}
	}

	b.WriteString(listBox.Render(listContent.String()))
	b.WriteString("\n\n")

	// Status/Error display
	if m.loading {
		b.WriteString(StyleInfo.Render("⟳ Deleting entries..."))
	} else if m.err != nil {
		b.WriteString(StyleError.Render(formatErrorForDisplay(m.err, 66)))
	}
	b.WriteString("\n\n")

	// Instructions
	if !m.loading {
		b.WriteString(StyleDim.Render("Confirm: y  Cancel: n/esc"))
	}

	return b.String()
}

// renderConfirmBatchDeleteView renders the batch delete confirmation view
func (m Model) renderConfirmBatchDeleteView() string {
	var b strings.Builder

	// Warning
	b.WriteString(StyleError.Render(fmt.Sprintf("⚠ WARNING: You are about to delete %d selected entries!", len(m.selectedEntries))))
	b.WriteString("\n\n")

	// Collect selected entries for display
	var selectedList []diff.SyncedEntry
	for _, entry := range m.entries {
		if m.selectedEntries[entry.Domain] {
			selectedList = append(selectedList, entry)
		}
	}

	// List entries to be deleted
	b.WriteString("The following entries will be deleted:\n\n")

	// Create a box for the list
	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorRed).
		Padding(0, 1).
		Width(70).
		Height(15)

	listContent := strings.Builder{}
	maxDisplay := 10
	for i, entry := range selectedList {
		if i >= maxDisplay {
			listContent.WriteString(StyleDim.Render(fmt.Sprintf("\n... and %d more entries", len(selectedList)-maxDisplay)))
			break
		}
		details := ""
		if entry.DNS != nil && entry.Caddy != nil {
			details = fmt.Sprintf("(DNS: %s → %s, Caddy: %s:%d)", entry.DNS.Type, entry.DNS.Content, entry.Caddy.Target, entry.Caddy.Port)
		} else if entry.DNS != nil {
			details = fmt.Sprintf("(DNS: %s → %s)", entry.DNS.Type, entry.DNS.Content)
		} else if entry.Caddy != nil {
			details = fmt.Sprintf("(Caddy: %s:%d)", entry.Caddy.Target, entry.Caddy.Port)
		}
		listContent.WriteString(fmt.Sprintf("%s %s\n", entry.Domain, details))
	}

	b.WriteString(listBox.Render(listContent.String()))
	b.WriteString("\n\n")

	// Status/Error display
	if m.loading {
		b.WriteString(StyleInfo.Render("⟳ Deleting entries..."))
	} else if m.err != nil {
		b.WriteString(StyleError.Render(formatErrorForDisplay(m.err, 66)))
	}
	b.WriteString("\n\n")

	// Instructions
	if !m.loading {
		b.WriteString(StyleDim.Render("Confirm: y  Cancel: n/esc"))
	}

	return b.String()
}

// renderConfirmBatchSyncView renders the batch sync confirmation view
func (m Model) renderConfirmBatchSyncView() string {
	var b strings.Builder

	// Warning
	b.WriteString(StyleWarning.Render(fmt.Sprintf("⚠ You are about to sync %d selected entries", len(m.selectedEntries))))
	b.WriteString("\n\n")

	// Collect selected entries for display
	var selectedList []diff.SyncedEntry
	for _, entry := range m.entries {
		if m.selectedEntries[entry.Domain] {
			selectedList = append(selectedList, entry)
		}
	}

	// List entries to be synced
	b.WriteString("The following entries will be synced:\n\n")

	// Create a box for the list
	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGreen).
		Padding(0, 1).
		Width(70).
		Height(15)

	listContent := strings.Builder{}
	maxDisplay := 10
	for i, entry := range selectedList {
		if i >= maxDisplay {
			listContent.WriteString(StyleDim.Render(fmt.Sprintf("\n... and %d more entries", len(selectedList)-maxDisplay)))
			break
		}
		action := ""
		if entry.Status == diff.StatusOrphanedDNS {
			action = "→ Create Caddy entry"
		} else if entry.Status == diff.StatusOrphanedCaddy {
			action = "→ Create DNS record"
		} else {
			action = "→ Already synced"
		}
		listContent.WriteString(fmt.Sprintf("%s %s\n", entry.Domain, action))
	}

	b.WriteString(listBox.Render(listContent.String()))
	b.WriteString("\n\n")

	// Info about defaults
	b.WriteString(StyleDim.Render("Note: New entries will use config defaults (CNAME, proxy, SSL, etc.)"))
	b.WriteString("\n\n")

	// Status/Error display
	if m.loading {
		b.WriteString(StyleInfo.Render("⟳ Syncing entries..."))
	} else if m.err != nil {
		b.WriteString(StyleError.Render(formatErrorForDisplay(m.err, 66)))
	}
	b.WriteString("\n\n")

	// Instructions
	if !m.loading {
		b.WriteString(StyleDim.Render("Confirm: y  Cancel: n/esc"))
	}

	return b.String()
}

// renderDeleteScopeView renders the delete scope selection screen
func (m Model) renderDeleteScopeView() string {
	var b strings.Builder

	// Get the entry to delete
	filteredEntries := m.getFilteredEntries()
	if m.deleteEntryIndex >= len(filteredEntries) {
		return StyleError.Render("Error: Invalid entry index")
	}
	entry := filteredEntries[m.deleteEntryIndex]

	// Show entry info
	b.WriteString(StyleInfo.Render(fmt.Sprintf("Delete: %s", entry.Domain)))
	b.WriteString("\n\n")

	// Only show available options based on what exists
	var availableScopes []DeleteScope
	if entry.DNS != nil && entry.Caddy != nil {
		// Both exist - show all options
		availableScopes = []DeleteScope{DeleteAll, DeleteDNSOnly, DeleteCaddyOnly}
	} else if entry.DNS != nil {
		// Only DNS exists
		availableScopes = []DeleteScope{DeleteDNSOnly}
		b.WriteString(StyleWarning.Render("⚠ This entry only has DNS record (no Caddyfile entry)"))
		b.WriteString("\n\n")
	} else if entry.Caddy != nil {
		// Only Caddy exists
		availableScopes = []DeleteScope{DeleteCaddyOnly}
		b.WriteString(StyleWarning.Render("⚠ This entry only has Caddyfile entry (no DNS record)"))
		b.WriteString("\n\n")
	}

	// If there's only one option, auto-select it and show message
	if len(availableScopes) == 1 {
		b.WriteString(StyleDim.Render("Only one deletion option available - press enter to confirm"))
		b.WriteString("\n\n")
	} else {
		b.WriteString("Select what to delete:")
		b.WriteString("\n\n")
	}

	// Delete scope options
	for i, scope := range availableScopes {
		// Build option line
		var line string
		// Map to cursor position
		cursorPos := m.deleteScopeCursor
		if cursorPos >= len(availableScopes) {
			cursorPos = len(availableScopes) - 1
		}

		if i == cursorPos {
			// Highlight selected option
			line = StyleHighlight.Render(fmt.Sprintf("→ %s", scope.String()))
		} else {
			line = fmt.Sprintf("  %s", scope.String())
		}

		b.WriteString(line)
		b.WriteString("\n")

		// Show description
		b.WriteString(StyleDim.Render(fmt.Sprintf("  %s", scope.Description())))
		b.WriteString("\n\n")
	}

	// Instructions
	if len(availableScopes) > 1 {
		b.WriteString(StyleDim.Render("Navigate: ↑/↓  Select: enter  Cancel: esc"))
	} else {
		b.WriteString(StyleDim.Render("Confirm: enter  Cancel: esc"))
	}

	return b.String()
}

// Modal content wrappers
func (m Model) renderDeleteScopeContent() string {
	return m.renderDeleteScopeView()
}

func (m Model) renderConfirmDeleteContent() string {
	return m.renderConfirmDeleteView()
}

func (m Model) renderConfirmSyncContent() string {
	return m.renderConfirmSyncView()
}

func (m Model) renderBulkDeleteMenuContent() string {
	return m.renderBulkDeleteMenu()
}

func (m Model) renderConfirmBulkDeleteContent() string {
	return m.renderConfirmBulkDeleteView()
}

func (m Model) renderConfirmBatchDeleteContent() string {
	return m.renderConfirmBatchDeleteView()
}

func (m Model) renderConfirmBatchSyncContent() string {
	return m.renderConfirmBatchSyncView()
}
