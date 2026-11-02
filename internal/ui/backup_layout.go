package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Backup views render in a large modal. The renderers, key handlers and mouse
// handler must agree on how many rows fit, or the cursor can land on a row
// that is not drawn. Keep every such calculation here.

// backupModalHeight is the height of the backup modals including borders.
func (m Model) backupModalHeight() int {
	_, h := modalDimensions(m.width, m.height, true)
	return h
}

// Rows of the modal not available to list/diff lines: borders, headers, the
// scroll indicator, the hint line and one spare line for an error message.
const backupModalChrome = 10

// backupListRows is how many backups the manager list shows at once.
func (m Model) backupListRows() int {
	return max(m.backupModalHeight()-backupModalChrome, 3)
}

// backupPreviewRows is how many diff lines the preview shows at once.
func (m Model) backupPreviewRows() int {
	return max(m.backupModalHeight()-backupModalChrome, 5)
}

// backupListTopY is the screen row of the first backup in the manager list.
// Layout: top border (1), "Total backups" line (1), blank line (1).
func (m Model) backupListTopY() int {
	return (m.height-m.backupModalHeight())/2 + 3
}

// backupMaxPreviewScroll is the largest valid PreviewScroll for the current diff.
func (m Model) backupMaxPreviewScroll() int {
	diff, err := GenerateDiff(m.config.Caddy.CaddyfilePath, m.backup.PreviewPath)
	if err != nil {
		return 0
	}
	return max(len(diff)-m.backupPreviewRows(), 0)
}

// scrollIntoView returns the scroll offset that keeps cursor inside a window
// of rows lines, moving the window as little as possible.
func scrollIntoView(cursor, offset, rows int) int {
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+rows {
		return cursor - rows + 1
	}
	return offset
}

// fitLine expands tabs and shortens s to at most width display columns,
// ending with an ellipsis when it had to cut.
func fitLine(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = strings.ReplaceAll(s, "\t", "    ")
	if lipgloss.Width(s) <= width {
		return s
	}
	var b strings.Builder
	used := 0
	for _, r := range s {
		w := lipgloss.Width(string(r))
		if used+w > width-1 {
			break
		}
		b.WriteRune(r)
		used += w
	}
	return b.String() + "…"
}

// backupDiffTextWidth is the columns available for diff text inside the
// preview box: modal border and padding (4), box border and padding (4).
func (m Model) backupDiffTextWidth() int {
	w, _ := modalDimensions(m.width, m.height, true)
	return max(w-8, 10)
}
