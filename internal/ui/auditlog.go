package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"lazyproxyflare/internal/audit"
)

// getFilteredAuditLogs returns audit logs filtered by current filters
func (m Model) getFilteredAuditLogs() []audit.LogEntry {
	if m.audit.OpFilter == "" && m.audit.ResultFilter == "" && m.audit.SearchQuery == "" {
		return m.audit.Logs
	}

	var filtered []audit.LogEntry
	for _, entry := range m.audit.Logs {
		// Filter by operation type
		if m.audit.OpFilter != "" && string(entry.Operation) != m.audit.OpFilter {
			continue
		}
		// Filter by result
		if m.audit.ResultFilter != "" && string(entry.Result) != m.audit.ResultFilter {
			continue
		}
		// Filter by search query (domain)
		if m.audit.SearchQuery != "" && !strings.Contains(
			strings.ToLower(entry.Domain),
			strings.ToLower(m.audit.SearchQuery),
		) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

// cycleOpFilter cycles through operation type filters
func (m *Model) cycleOpFilter() {
	ops := []string{"", "create", "update", "delete", "sync", "restore"}
	for i, op := range ops {
		if op == m.audit.OpFilter {
			m.audit.OpFilter = ops[(i+1)%len(ops)]
			return
		}
	}
	m.audit.OpFilter = ""
}

// cycleResultFilter cycles through result filters
func (m *Model) cycleResultFilter() {
	results := []string{"", "success", "failure"}
	for i, r := range results {
		if r == m.audit.ResultFilter {
			m.audit.ResultFilter = results[(i+1)%len(results)]
			return
		}
	}
	m.audit.ResultFilter = ""
}

// renderAuditLogView renders the audit log viewer modal
func (m Model) renderAuditLogView() string {
	width := m.width
	height := m.height

	// Calculate modal dimensions (2/3 of screen, same as RenderModalOverlay)
	modalWidth := width * 2 / 3
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalWidth > 80 {
		modalWidth = 80
	}

	modalHeight := height * 2 / 3
	if modalHeight < 15 {
		modalHeight = 15
	}

	// Modal styling (matching RenderModalOverlay)
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorBlue).
		Background(lipgloss.Color("#0a0a0a")).
		Width(modalWidth - 2).
		Height(modalHeight - 2).
		Padding(1, 2)

	// Render content with calculated modal height
	content := m.renderAuditLogContent(modalHeight)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content),
	)
}

// renderAuditLogContent renders the audit log list content
func (m Model) renderAuditLogContent(modalHeight int) string {
	var b strings.Builder

	// Title
	b.WriteString(StyleInfo.Render("Audit Log"))
	b.WriteString("\n")

	// Filter bar
	opLabel := "ALL"
	if m.audit.OpFilter != "" {
		opLabel = strings.ToUpper(m.audit.OpFilter)
	}
	resultLabel := "ALL"
	if m.audit.ResultFilter != "" {
		resultLabel = strings.ToUpper(m.audit.ResultFilter)
	}
	filterBar := fmt.Sprintf("[Op: %s]  [Result: %s]", opLabel, resultLabel)
	if m.audit.SearchActive {
		filterBar += fmt.Sprintf("  Search: %s_", m.audit.SearchQuery)
	} else if m.audit.SearchQuery != "" {
		filterBar += fmt.Sprintf("  Search: %s", m.audit.SearchQuery)
	}
	b.WriteString(StyleDim.Render(filterBar))
	b.WriteString("\n\n")

	// Get filtered logs
	filteredLogs := m.getFilteredAuditLogs()

	// Show message if no logs
	if len(m.audit.Logs) == 0 {
		b.WriteString(StyleDim.Render("No audit log entries yet."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Operations will be logged here as you create, update, delete, or sync entries."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press ESC to close"))
		return b.String()
	}

	if len(filteredLogs) == 0 {
		b.WriteString(StyleDim.Render("No entries match the current filters."))
		b.WriteString("\n\n")
	} else {
		// Calculate visible area
		reservedLines := 13
		availableHeight := modalHeight - reservedLines
		if availableHeight < 3 {
			availableHeight = 3
		}
		visibleEntries := availableHeight / 3

		// Ensure scroll offset is valid
		if m.audit.Scroll > len(filteredLogs)-1 {
			m.audit.Scroll = len(filteredLogs) - 1
		}
		if m.audit.Scroll < 0 {
			m.audit.Scroll = 0
		}

		// Display entries in reverse chronological order (newest first)
		start := m.audit.Scroll
		end := start + visibleEntries
		if end > len(filteredLogs) {
			end = len(filteredLogs)
		}

		// Reverse the logs for display (newest first)
		reversedLogs := make([]audit.LogEntry, len(filteredLogs))
		for i, log := range filteredLogs {
			reversedLogs[len(filteredLogs)-1-i] = log
		}

		// Render visible entries
		for i := start; i < end; i++ {
			entry := reversedLogs[i]
			b.WriteString(m.formatLogEntry(entry))
			b.WriteString("\n")
		}

		// Show count
		b.WriteString("\n")
		if len(filteredLogs) != len(m.audit.Logs) {
			b.WriteString(StyleDim.Render(fmt.Sprintf("Showing %d of %d entries", len(filteredLogs), len(m.audit.Logs))))
		} else {
			b.WriteString(StyleDim.Render(fmt.Sprintf("%d entries", len(m.audit.Logs))))
		}
	}

	// Instructions
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("↑/↓: scroll  /: search  f: filter op  r: filter result  ESC: close"))

	return b.String()
}

// formatLogEntry formats a single log entry for display
func (m Model) formatLogEntry(entry audit.LogEntry) string {
	var b strings.Builder

	// Timestamp
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	b.WriteString(StyleDim.Render(timestamp))
	b.WriteString("  ")

	// Result icon
	if entry.Result == audit.ResultSuccess {
		b.WriteString(StyleSuccess.Render("✓"))
	} else {
		b.WriteString(StyleError.Render("✗"))
	}
	b.WriteString(" ")

	// Operation (uppercase, colored)
	opText := strings.ToUpper(string(entry.Operation))
	b.WriteString(StyleInfo.Render(opText))
	b.WriteString("  ")

	// Domain
	b.WriteString(entry.Domain)

	// Entity type indicator if not "both"
	if entry.EntityType == audit.EntityDNS {
		b.WriteString(StyleDim.Render(" (DNS only)"))
	} else if entry.EntityType == audit.EntityCaddy {
		b.WriteString(StyleDim.Render(" (Caddy only)"))
	}

	// Show details on next line if present
	if len(entry.Details) > 0 {
		b.WriteString("\n  ")
		b.WriteString(StyleDim.Render(m.formatDetails(entry.Details)))
	}

	// Show error if present
	if entry.Error != "" {
		b.WriteString("\n  ")
		b.WriteString(StyleError.Render(fmt.Sprintf("Error: %s", entry.Error)))
	}

	// Show batch count if present
	if entry.BatchCount > 0 {
		b.WriteString("\n  ")
		b.WriteString(StyleDim.Render(fmt.Sprintf("(%d entries)", entry.BatchCount)))
	}

	return b.String()
}

// formatDetails formats the details map into a readable string
func (m Model) formatDetails(details map[string]interface{}) string {
	var parts []string

	// Format specific known details
	if val, ok := details["dns_type"]; ok {
		parts = append(parts, fmt.Sprintf("Type: %v", val))
	}
	if val, ok := details["target"]; ok {
		parts = append(parts, fmt.Sprintf("Target: %v", val))
	}
	if val, ok := details["proxied"]; ok {
		if proxied, ok := val.(bool); ok && proxied {
			parts = append(parts, "Proxied")
		}
	}
	if val, ok := details["dns_only"]; ok {
		if dnsOnly, ok := val.(bool); ok && dnsOnly {
			parts = append(parts, "DNS-only mode")
		}
	}
	if val, ok := details["sync_direction"]; ok {
		parts = append(parts, fmt.Sprintf("Direction: %v", val))
	}

	// If no known details, show generic key-value pairs
	if len(parts) == 0 {
		for k, v := range details {
			parts = append(parts, fmt.Sprintf("%s: %v", k, v))
		}
	}

	return strings.Join(parts, ", ")
}
