package ui

import (
	"fmt"
	"strings"

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
	ops := []string{""}
	for _, op := range audit.Operations {
		ops = append(ops, string(op))
	}
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
	return RenderModalOverlayLarge("", "Audit Log", m.renderAuditLogContent(), m.width, m.height)
}

// Rows of the modal not available to entries: borders (2), filter bar (1),
// blank (1), blank before the count (1), count (1), hint (1) and one spare.
const auditModalChrome = 8

// auditEntryBudget is how many screen lines the entries may use.
func (m Model) auditEntryBudget() int {
	_, h := modalDimensions(m.width, m.height, true)
	return max(h-auditModalChrome, 3)
}

// auditDisplayEntries returns the filtered entries, newest first.
func (m Model) auditDisplayEntries() []audit.LogEntry {
	filtered := m.getFilteredAuditLogs()
	out := make([]audit.LogEntry, len(filtered))
	for i, e := range filtered {
		out[len(filtered)-1-i] = e
	}
	return out
}

// auditEntryHeight is the number of screen lines one entry takes (1 to 4).
func (m Model) auditEntryHeight(e audit.LogEntry) int {
	return strings.Count(m.formatLogEntry(e), "\n") + 1
}

// auditPageEnd returns the index after the last entry that fits when drawing
// from start. At least one entry is always shown.
func (m Model) auditPageEnd(entries []audit.LogEntry, start int) int {
	budget, used, end := m.auditEntryBudget(), 0, start
	for end < len(entries) {
		h := m.auditEntryHeight(entries[end])
		if used+h > budget && end > start {
			break
		}
		used += h
		end++
	}
	return end
}

// auditMaxScroll is the largest valid Scroll: the one whose page ends on the
// oldest entry. Entries vary in height, so count back from the end.
func (m Model) auditMaxScroll() int {
	entries := m.auditDisplayEntries()
	budget, used, shown := m.auditEntryBudget(), 0, 0
	for i := len(entries) - 1; i >= 0; i-- {
		h := m.auditEntryHeight(entries[i])
		if used+h > budget && shown > 0 {
			break
		}
		used += h
		shown++
	}
	return len(entries) - shown
}

// renderAuditLogContent renders the audit log list content
func (m Model) renderAuditLogContent() string {
	var b strings.Builder

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
		entries := m.auditDisplayEntries()
		start := min(max(m.audit.Scroll, 0), m.auditMaxScroll())
		end := m.auditPageEnd(entries, start)

		// Render visible entries
		for i := start; i < end; i++ {
			b.WriteString(m.formatLogEntry(entries[i]))
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
