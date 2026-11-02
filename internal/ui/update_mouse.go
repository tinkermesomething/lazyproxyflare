package ui

import (
	tea "charm.land/bubbletea/v2"

	"lazyproxyflare/internal/caddy"
)

// mouseAction is the part of a mouse event the app reacts to. Bubble Tea v2
// splits mouse input into click, wheel and motion messages; Update maps them
// to this so the handlers (and their tests) do not depend on library types.
type mouseAction int

const (
	mouseWheelUp mouseAction = iota
	mouseWheelDown
	mouseClick // left button press
)

// mouseEvent is a mouse action at a screen position.
type mouseEvent struct {
	Action mouseAction
	X, Y   int
}

// handleMouse handles all mouse input events
func (m Model) handleMouse(msg mouseEvent) (Model, tea.Cmd) {
	switch m.currentView {
	case ViewList:
		layout := NewPanelLayout(m.width, m.height)

		switch msg.Action {
		case mouseWheelUp:
			// Scroll up
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case mouseWheelDown:
			// Scroll down
			filtered := m.getFilteredEntries()
			if m.cursor < len(filtered)-1 {
				m.cursor++
				maxVisible := layout.LeftHeight - 5 // Account for header and padding
				if m.cursor >= m.scrollOffset+maxVisible {
					m.scrollOffset++
				}
			}

		case mouseClick:
			// Determine which panel was clicked
			if msg.X < layout.LeftWidth {
				// Left panel clicked
				m.panelFocus = PanelFocusLeft

				// Calculate which entry was clicked
				// Count actual header lines (dynamic based on filters/sorts)
				headerLines := 0
				if m.statusFilter != FilterAll || m.dnsTypeFilter != DNSTypeAll || m.searchQuery != "" {
					headerLines++ // Filter line
				}
				if m.sortMode != SortAlphabetical {
					headerLines++ // Sort line
				}
				if len(m.selectedEntries) > 0 {
					headerLines++ // Selection line
				}

				// Account for: title bar (1) + border (1) + panel title (1) + header lines (N)
				clickY := msg.Y - 3 - headerLines
				if clickY >= 0 {
					filtered := m.getFilteredEntries()
					clickedIndex := m.scrollOffset + clickY
					if clickedIndex >= 0 && clickedIndex < len(filtered) {
						// Check if checkbox area was clicked
						// Account for cursor indicator (2 chars: "→ " or "  ")
						// Checkbox starts at position 2, spans 4 chars: "[ ] " or "[✓] "
						if msg.X >= 2 && msg.X <= 7 {
							// Toggle selection
							domain := filtered[clickedIndex].Domain
							if m.selectedEntries[domain] {
								delete(m.selectedEntries, domain)
							} else {
								m.selectedEntries[domain] = true
							}
						} else {
							// Move cursor to clicked entry
							m.cursor = clickedIndex
						}
					}
				}
			} else {
				// Right panels clicked - determine if details or snippets based on Y position
				// Title bar (1) + Details panel height + border (2) = threshold for snippets panel
				detailsThresholdY := 1 + layout.DetailsHeight + 2
				if msg.Y < detailsThresholdY {
					m.panelFocus = PanelFocusDetails
				} else {
					m.panelFocus = PanelFocusSnippets
				}
			}
		}

	case ViewBackupManager:
		// Mouse support in backup manager
		switch msg.Action {
		case mouseWheelUp:
			// Scroll up
			if m.backup.Cursor > 0 {
				m.backup.Cursor--
				m.backup.ScrollOffset = scrollIntoView(m.backup.Cursor, m.backup.ScrollOffset, m.backupListRows())
			}

		case mouseWheelDown:
			// Scroll down
			backups, _ := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
			if m.backup.Cursor < len(backups)-1 {
				m.backup.Cursor++
				m.backup.ScrollOffset = scrollIntoView(m.backup.Cursor, m.backup.ScrollOffset, m.backupListRows())
			}

		case mouseClick:
			// Click to select a backup: map the screen row to a list row
			relativeY := msg.Y - m.backupListTopY()

			if relativeY >= 0 && relativeY < m.backupListRows() {
				clickedIndex := m.backup.ScrollOffset + relativeY
				backups, _ := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
				if clickedIndex >= 0 && clickedIndex < len(backups) {
					m.backup.Cursor = clickedIndex
				}
			}
		}
	}

	return m, nil
}
