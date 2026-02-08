package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"lazyproxyflare/internal/audit"
)

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
	b.WriteString("\n\n")

	// Show message if no logs
	if len(m.audit.Logs) == 0 {
		b.WriteString(StyleDim.Render("No audit log entries yet."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Operations will be logged here as you create, update, delete, or sync entries."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press ESC to close"))
		return b.String()
	}

	// Calculate visible area based on modal height
	// Account for: title (3 lines), scroll indicator (2 lines), instructions (2 lines), padding/borders (4 lines)
	// Each log entry typically takes 2-4 lines, so divide available height by 3 to be safe
	reservedLines := 11
	availableHeight := modalHeight - reservedLines
	if availableHeight < 3 {
		availableHeight = 3
	}
	// Estimate lines per entry (each entry is ~3 lines on average)
	visibleEntries := availableHeight / 3

	// Ensure scroll offset is valid
	if m.audit.Scroll > len(m.audit.Logs)-1 {
		m.audit.Scroll = len(m.audit.Logs) - 1
	}
	if m.audit.Scroll < 0 {
		m.audit.Scroll = 0
	}

	// Display entries in reverse chronological order (newest first)
	start := m.audit.Scroll
	end := start + visibleEntries
	if end > len(m.audit.Logs) {
		end = len(m.audit.Logs)
	}

	// Reverse the logs for display (newest first)
	reversedLogs := make([]audit.LogEntry, len(m.audit.Logs))
	for i, log := range m.audit.Logs {
		reversedLogs[len(m.audit.Logs)-1-i] = log
	}

	// Render visible entries
	for i := start; i < end; i++ {
		entry := reversedLogs[i]
		b.WriteString(m.formatLogEntry(entry))
		b.WriteString("\n")
	}

	// Show scroll indicator
	if len(m.audit.Logs) > visibleEntries {
		totalEntries := len(m.audit.Logs)
		b.WriteString("\n")
		b.WriteString(StyleDim.Render(fmt.Sprintf("Showing %d-%d of %d entries", start+1, end, totalEntries)))
	}

	// Instructions
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("↑/↓: scroll  ESC: close"))

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
