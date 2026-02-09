package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"lazyproxyflare/internal/caddy"

	"github.com/charmbracelet/lipgloss"
)

// DiffLine represents a single line in a diff view
type DiffLine struct {
	Type    string // "same", "removed", "added", "context"
	Content string
	LineNum int // Line number in original/backup file
}

// GenerateDiff creates a simple line-by-line diff between two files
func GenerateDiff(currentPath, backupPath string) ([]DiffLine, error) {
	// Read both files
	currentContent, err := os.ReadFile(currentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read current file: %w", err)
	}

	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	// Split into lines
	currentLines := strings.Split(string(currentContent), "\n")
	backupLines := strings.Split(string(backupContent), "\n")

	var diff []DiffLine

	// Build maps for quick lookup
	currentMap := make(map[string]bool)
	for _, line := range currentLines {
		currentMap[line] = true
	}

	backupMap := make(map[string]bool)
	for _, line := range backupLines {
		backupMap[line] = true
	}

	// Process lines - simple algorithm showing all unique lines
	// First, show lines only in current (will be removed)
	for i, line := range currentLines {
		if !backupMap[line] {
			diff = append(diff, DiffLine{
				Type:    "removed",
				Content: line,
				LineNum: i + 1,
			})
		}
	}

	// Then show lines only in backup (will be added)
	for i, line := range backupLines {
		if !currentMap[line] {
			diff = append(diff, DiffLine{
				Type:    "added",
				Content: line,
				LineNum: i + 1,
			})
		}
	}

	// If diff is empty, files are identical
	if len(diff) == 0 {
		diff = append(diff, DiffLine{
			Type:    "same",
			Content: "No changes - backup is identical to current Caddyfile",
			LineNum: 0,
		})
	}

	return diff, nil
}

// renderBackupManagerView renders the backup manager screen
func (m Model) renderBackupManagerView() string {
	var b strings.Builder

	// Get backups
	backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
	if err != nil {
		b.WriteString(StyleError.Render(fmt.Sprintf("Error loading backups: %v", err)))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press esc to go back"))
		return b.String()
	}

	if len(backups) == 0 {
		b.WriteString("No backups found.\n\n")
		b.WriteString(StyleDim.Render("Backups are created automatically before destructive operations."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press esc to go back"))
		return b.String()
	}

	// Summary
	b.WriteString(fmt.Sprintf("Total backups: %d\n\n", len(backups)))

	// Calculate visible range
	visibleHeight := m.height - 10
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	start := m.backup.ScrollOffset
	end := start + visibleHeight
	if end > len(backups) {
		end = len(backups)
	}

	// Backup list
	for i := start; i < end; i++ {
		backup := backups[i]

		// Format timestamp
		timestamp := backup.Timestamp.Format("2006-01-02 15:04:05")

		// Format size
		sizeKB := float64(backup.Size) / 1024.0
		sizeStr := fmt.Sprintf("%.1f KB", sizeKB)

		// Build line
		line := fmt.Sprintf("%-19s  %8s", timestamp, sizeStr)

		// Apply cursor style
		if i == m.backup.Cursor {
			line = StyleHighlight.Render("→ " + line)
		} else {
			line = normalStyle.Render("  " + line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(backups) > visibleHeight {
		scrollInfo := fmt.Sprintf("\n[Showing %d-%d of %d]", start+1, end, len(backups))
		b.WriteString(StyleDim.Render(scrollInfo))
		b.WriteString("\n")
	}

	// Error display
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(StyleError.Render(fmt.Sprintf("✗ Error: %v", m.err)))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	if m.loading {
		b.WriteString(StyleInfo.Render("⟳ Please wait..."))
	} else {
		b.WriteString(StyleDim.Render("Navigate: ↑/↓  Preview: Enter  Restore: R  Delete: x  Cleanup: c  Back: esc"))
	}

	return b.String()
}

// renderBackupPreviewView renders the backup preview screen with diff view
func (m Model) renderBackupPreviewView() string {
	var b strings.Builder

	// Show backup info
	info, err := os.Stat(m.backup.PreviewPath)
	if err != nil {
		b.WriteString(StyleError.Render(fmt.Sprintf("Error getting backup info: %v", err)))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press esc to go back"))
		return b.String()
	}

	timestamp := info.ModTime().Format("2006-01-02 15:04:05")
	sizeKB := float64(info.Size()) / 1024.0
	b.WriteString(fmt.Sprintf("Backup from %s (%.1f KB)\n", timestamp, sizeKB))
	b.WriteString(StyleDim.Render("Showing changes if restored (- current, + backup)"))
	b.WriteString("\n\n")

	// Generate diff between current and backup
	diffLines, err := GenerateDiff(m.config.Caddy.CaddyfilePath, m.backup.PreviewPath)
	if err != nil {
		b.WriteString(StyleError.Render(fmt.Sprintf("Error generating diff: %v", err)))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press esc to go back"))
		return b.String()
	}

	// Calculate visible area for modal content
	// Modal is 2/3 of screen height (from RenderModalOverlay)
	// Account for: header (3 lines), scroll indicator (1 line), instructions (1 line), modal padding/borders (6 lines)
	modalHeight := m.height * 2 / 3
	if modalHeight < 15 {
		modalHeight = 15
	}
	reservedLines := 11
	visibleHeight := modalHeight - reservedLines
	if visibleHeight < 5 {
		visibleHeight = 5
	}

	// Apply scroll offset
	start := m.backup.PreviewScroll
	end := start + visibleHeight
	if end > len(diffLines) {
		end = len(diffLines)
	}
	if start >= len(diffLines) {
		start = len(diffLines) - 1
		if start < 0 {
			start = 0
		}
	}

	// Build visible content with color coding
	var visibleContent strings.Builder
	for i := start; i < end; i++ {
		if i < len(diffLines) {
			line := diffLines[i]

			// Style based on diff type
			var styledLine string
			switch line.Type {
			case "removed":
				// Red for lines that will be removed
				styledLine = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FF6B6B")).
					Render(fmt.Sprintf("- %s", line.Content))
			case "added":
				// Green for lines that will be added
				styledLine = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#51CF66")).
					Render(fmt.Sprintf("+ %s", line.Content))
			case "same":
				// Gray for no changes message
				styledLine = StyleDim.Render(line.Content)
			default:
				styledLine = line.Content
			}

			visibleContent.WriteString(styledLine)
			if i < end-1 {
				visibleContent.WriteString("\n")
			}
		}
	}

	// Show content in a box (let width adapt to modal, don't set explicit dimensions)
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBlue).
		Padding(0, 1)

	b.WriteString(contentBox.Render(visibleContent.String()))
	b.WriteString("\n")

	// Scroll indicator
	if len(diffLines) > visibleHeight {
		scrollInfo := fmt.Sprintf("[Showing %d-%d of %d changes]", start+1, end, len(diffLines))
		b.WriteString(StyleDim.Render(scrollInfo))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString(StyleDim.Render("Scroll: j/k/↑/↓  Page: PgUp/PgDn  Top: g  Bottom: G  Next backup: →  Prev backup: ←  Delete: d  Restore: R  Back: esc"))

	return b.String()
}

// renderRestoreScopeView renders the restore scope selection screen
func (m Model) renderRestoreScopeView() string {
	var b strings.Builder

	// Show backup info
	info, err := os.Stat(m.backup.PreviewPath)
	if err != nil {
		b.WriteString(StyleError.Render(fmt.Sprintf("Error getting backup info: %v", err)))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press esc to go back"))
		return b.String()
	}

	timestamp := info.ModTime().Format("2006-01-02 15:04:05")
	sizeKB := float64(info.Size()) / 1024.0

	b.WriteString(fmt.Sprintf("Backup from %s (%.1f KB)\n", timestamp, sizeKB))
	b.WriteString("\n")
	b.WriteString(StyleInfo.Render("Select what to restore:"))
	b.WriteString("\n\n")

	// Restore scope options
	scopes := []RestoreScope{RestoreAll, RestoreDNSOnly, RestoreCaddyOnly}
	for i, scope := range scopes {
		// Build option line
		var line string
		if i == m.backup.RestoreScopeCursor {
			// Highlight selected option
			line = StyleHighlight.Render(fmt.Sprintf("→ %s", scope.String()))
		} else {
			line = fmt.Sprintf("  %s", scope.String())
		}

		b.WriteString(line)
		b.WriteString("\n")

		// Show description
		desc := scope.Description()
		if i == m.backup.RestoreScopeCursor {
			b.WriteString(StyleDim.Render(fmt.Sprintf("  %s", desc)))
		} else {
			b.WriteString(StyleDim.Render(fmt.Sprintf("  %s", desc)))
		}
		b.WriteString("\n\n")
	}

	// Instructions
	b.WriteString(StyleDim.Render("Navigate: j/k/↑/↓  Select: enter  Cancel: esc"))

	return b.String()
}

// renderConfirmRestoreView renders the restore confirmation screen
func (m Model) renderConfirmRestoreView() string {
	var b strings.Builder

	// Warning message based on restore scope
	var warningMsg string
	switch m.backup.RestoreScope {
	case RestoreAll:
		warningMsg = "⚠ WARNING: This will restore both Caddyfile and DNS records from backup"
	case RestoreDNSOnly:
		warningMsg = "⚠ WARNING: This will restore DNS records from backup"
	case RestoreCaddyOnly:
		warningMsg = "⚠ WARNING: This will restore Caddyfile from backup"
	}
	b.WriteString(StyleWarning.Render(warningMsg))
	b.WriteString("\n\n")

	// Show backup info
	info, err := os.Stat(m.backup.PreviewPath)
	if err != nil {
		b.WriteString(StyleError.Render(fmt.Sprintf("Error getting backup info: %v", err)))
		return b.String()
	}

	timestamp := info.ModTime().Format("2006-01-02 15:04:05")
	sizeKB := float64(info.Size()) / 1024.0

	b.WriteString(fmt.Sprintf("Backup: %s (%.1f KB)\n", timestamp, sizeKB))
	b.WriteString(fmt.Sprintf("Restore scope: %s\n", StyleInfo.Render(m.backup.RestoreScope.String())))
	b.WriteString("\n")

	// Show what will happen based on restore scope
	b.WriteString("This will:\n")
	switch m.backup.RestoreScope {
	case RestoreAll:
		b.WriteString("  1. Create a new backup of your current Caddyfile\n")
		b.WriteString("  2. Replace your current Caddyfile with the backup\n")
		b.WriteString("  3. Validate the restored configuration\n")
		b.WriteString("  4. Restart Caddy\n")
		b.WriteString("  5. Update DNS records to match the backup\n")
	case RestoreDNSOnly:
		b.WriteString("  1. Parse domains from backup Caddyfile\n")
		b.WriteString("  2. Create/update DNS records for each domain\n")
		b.WriteString("  3. Leave current Caddyfile unchanged\n")
	case RestoreCaddyOnly:
		b.WriteString("  1. Create a new backup of your current Caddyfile\n")
		b.WriteString("  2. Replace your current Caddyfile with the backup\n")
		b.WriteString("  3. Validate the restored configuration\n")
		b.WriteString("  4. Restart Caddy\n")
		b.WriteString("  5. Leave DNS records unchanged\n")
	}
	b.WriteString("\n")

	// Status/Error display
	if m.loading {
		b.WriteString(StyleInfo.Render("⟳ Restoring backup..."))
	} else if m.err != nil {
		b.WriteString(StyleError.Render(fmt.Sprintf("✗ Error: %v", m.err)))
	} else {
		b.WriteString("Press 'y' to confirm or 'n' to cancel")
	}
	b.WriteString("\n\n")

	// Instructions
	if !m.loading {
		b.WriteString(StyleDim.Render("Confirm: y  Cancel: n/esc"))
	}

	return b.String()
}

// renderConfirmCleanupView renders the cleanup old backups confirmation screen
func (m Model) renderConfirmCleanupView() string {
	var b strings.Builder

	// Get list of old backups
	maxAge := time.Duration(m.backup.RetentionDays) * 24 * time.Hour
	oldBackups, err := caddy.GetOldBackups(m.config.Caddy.CaddyfilePath, maxAge)

	if err != nil {
		b.WriteString(StyleError.Render(fmt.Sprintf("Error loading old backups: %v", err)))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press esc to go back"))
		return b.String()
	}

	if len(oldBackups) == 0 {
		b.WriteString(fmt.Sprintf("No backups older than %d days found.\n\n", m.backup.RetentionDays))
		b.WriteString(StyleDim.Render("All backups are within the retention period."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press esc to go back"))
		return b.String()
	}

	// Warning
	b.WriteString(StyleWarning.Render(fmt.Sprintf("⚠ Found %d backups older than %d days", len(oldBackups), m.backup.RetentionDays)))
	b.WriteString("\n\n")

	// List backups to be deleted
	b.WriteString("The following backups will be deleted:\n\n")

	// Create a box for the list
	listBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorOrange).
		Padding(0, 1).
		Width(70).
		Height(15)

	listContent := strings.Builder{}
	maxDisplay := 10
	for i, backup := range oldBackups {
		if i >= maxDisplay {
			listContent.WriteString(StyleDim.Render(fmt.Sprintf("\n... and %d more backups", len(oldBackups)-maxDisplay)))
			break
		}
		timestamp := backup.Timestamp.Format("2006-01-02 15:04:05")
		sizeKB := float64(backup.Size) / 1024.0
		age := int(time.Since(backup.Timestamp).Hours() / 24)
		listContent.WriteString(fmt.Sprintf("%s  %.1f KB  (%d days old)\n", timestamp, sizeKB, age))
	}

	b.WriteString(listBox.Render(listContent.String()))
	b.WriteString("\n\n")

	// Calculate total size to free
	totalBytes := int64(0)
	for _, backup := range oldBackups {
		totalBytes += backup.Size
	}
	totalMB := float64(totalBytes) / (1024.0 * 1024.0)
	b.WriteString(StyleInfo.Render(fmt.Sprintf("Total space to free: %.2f MB", totalMB)))
	b.WriteString("\n\n")

	// Status/Error display
	if m.loading {
		b.WriteString(StyleInfo.Render("⟳ Cleaning up old backups..."))
	} else if m.err != nil {
		b.WriteString(StyleError.Render(fmt.Sprintf("✗ Error: %v", m.err)))
	} else {
		b.WriteString("Press 'y' to confirm or 'n' to cancel")
	}
	b.WriteString("\n\n")

	// Instructions
	if !m.loading {
		b.WriteString(StyleDim.Render("Confirm: y  Cancel: n/esc"))
	}

	return b.String()
}

// Modal content wrappers
func (m Model) renderBackupManagerContent() string {
	return m.renderBackupManagerView()
}

func (m Model) renderBackupPreviewContent() string {
	return m.renderBackupPreviewView()
}

func (m Model) renderRestoreScopeContent() string {
	return m.renderRestoreScopeView()
}

func (m Model) renderConfirmRestoreContent() string {
	return m.renderConfirmRestoreView()
}

func (m Model) renderConfirmCleanupContent() string {
	return m.renderConfirmCleanupView()
}
