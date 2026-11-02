package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/cmdlog"
)

// Rows of the modal not available to entries: borders (2), summary line (1),
// blank (1), blank before the hint (1), hint (1) and one spare.
const commandLogChrome = 8

// commandLogRows is how many entries the modal shows at once.
func (m Model) commandLogRows() int {
	_, h := modalDimensions(m.width, m.height, true)
	return max(h-commandLogChrome, 3)
}

// commandLogMaxScroll is the largest valid cmdLogScroll.
func (m Model) commandLogMaxScroll() int {
	return max(len(cmdlog.Default.Entries())-m.commandLogRows(), 0)
}

// handleOpenCommandLog opens the command log from the list, and closes it again.
func (m Model) handleOpenCommandLog() (Model, tea.Cmd) {
	switch {
	case m.currentView == ViewCommandLog:
		m.currentView = ViewList
	case m.currentView == ViewList && !m.loading && !m.searching:
		m.cmdLogScroll = 0
		m.currentView = ViewCommandLog
	}
	return m, nil
}

// isScrollModal reports whether the current view is a modal that scrolls with
// the arrow keys. j/k must do the same there, as the hint lines promise.
func (m Model) isScrollModal() bool {
	switch m.currentView {
	case ViewBackupManager, ViewBackupPreview, ViewAuditLog, ViewCommandLog:
		return true
	}
	return false
}

// renderCommandLogView renders the command log modal.
func (m Model) renderCommandLogView() string {
	return RenderModalOverlayLarge("", "Command Log", m.renderCommandLogContent(), m.width, m.height)
}

func (m Model) renderCommandLogContent() string {
	entries := cmdlog.Default.Entries()
	var b strings.Builder

	if len(entries) == 0 {
		b.WriteString(StyleDim.Render("Nothing has run yet."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Cloudflare API calls and shell commands appear here as the app makes them."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("esc: close"))
		return b.String()
	}

	b.WriteString(StyleDim.Render(fmt.Sprintf("%d commands this session (request bodies and tokens are never logged)", len(entries))))
	b.WriteString("\n\n")

	// Newest first: display index k is entries[len-1-k].
	start := min(max(m.cmdLogScroll, 0), m.commandLogMaxScroll())
	end := min(start+m.commandLogRows(), len(entries))
	for k := start; k < end; k++ {
		b.WriteString(formatCommandEntry(entries[len(entries)-1-k]))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(StyleDim.Render("↑/↓ j/k: scroll  g/G: top/bottom  @/esc: close"))
	return b.String()
}

// formatCommandEntry renders one log entry on a single line.
func formatCommandEntry(e cmdlog.Entry) string {
	kind := "API "
	if e.Kind == cmdlog.Exec {
		kind = "EXEC"
	}

	status := StyleSuccess.Render(e.Status)
	if e.Status != "ok" && !strings.HasPrefix(e.Status, "2") {
		status = StyleError.Render(e.Status)
	}

	line := fmt.Sprintf("%s  %s  %s  %s  %s",
		StyleDim.Render(e.Time.Format("15:04:05")),
		StyleInfo.Render(kind),
		e.Summary,
		status,
		StyleDim.Render(e.Duration.Round(time.Millisecond).String()))
	if e.Err != "" {
		line += "  " + StyleError.Render(e.Err)
	}
	return line
}
