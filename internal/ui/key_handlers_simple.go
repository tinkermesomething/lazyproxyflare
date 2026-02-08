package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/diff"
)

// handleSearchStart enters search mode from list view.
func (m Model) handleSearchStart() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching && !m.loading {
		m.searching = true
		m.searchQuery = ""
		return m, nil
	}
	return m, nil
}

// handleStatusFilterCycle cycles through status filters.
func (m Model) handleStatusFilterCycle() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching && !m.loading {
		m.statusFilter = (m.statusFilter + 1) % 4
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil
	}
	return m, nil
}

// handleDNSTypeFilterCycle cycles through DNS type filters.
func (m Model) handleDNSTypeFilterCycle() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching && !m.loading {
		m.dnsTypeFilter = (m.dnsTypeFilter + 1) % 3
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil
	}
	return m, nil
}

// handleSortModeCycle cycles through sort modes.
func (m Model) handleSortModeCycle() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching && !m.loading {
		m.sortMode = (m.sortMode + 1) % 2
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil
	}
	return m, nil
}

// handleAddEntry opens the add entry form with defaults.
func (m Model) handleAddEntry() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching && !m.loading {
		m.addForm = AddFormData{
			Subdomain:          "",
			DNSType:            "CNAME",
			DNSTarget:          m.config.Defaults.CNAMETarget,
			DNSOnly:            false,
			ReverseProxyTarget: "localhost",
			ServicePort:        fmt.Sprintf("%d", m.config.Defaults.Port),
			Proxied:            m.config.Defaults.Proxied,
			LANOnly:            false,
			SSL:                m.config.Defaults.SSL,
			OAuth:              false,
			WebSocket:          false,
			SelectedSnippets:   make(map[string]bool),
			FocusedField:       0,
		}
		m.currentView = ViewAdd
		return m, nil
	}
	return m, nil
}

// handleOpenBulkDeleteMenu opens the bulk delete menu.
func (m Model) handleOpenBulkDeleteMenu() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.loading {
		m.bulkDelete.MenuCursor = 0
		m.currentView = ViewBulkDeleteMenu
		return m, nil
	}
	return m, nil
}

// handleOpenBackupManager opens the backup manager.
func (m Model) handleOpenBackupManager() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.loading {
		m.backup.Cursor = 0
		m.backup.ScrollOffset = 0
		m.currentView = ViewBackupManager
		return m, nil
	}
	return m, nil
}

// handleOpenAuditLog opens the audit log viewer.
func (m Model) handleOpenAuditLog() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.loading {
		if m.audit.Logger != nil {
			logs, err := m.audit.Logger.LoadLogs()
			if err == nil {
				m.audit.Logs = logs
				m.audit.Cursor = 0
				m.audit.Scroll = 0
			}
		}
		m.currentView = ViewAuditLog
		return m, nil
	}
	return m, nil
}

// handleBatchDeleteSelected opens batch delete confirmation for selected entries.
func (m Model) handleBatchDeleteSelected() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.loading && len(m.selectedEntries) > 0 {
		m.currentView = ViewConfirmBatchDelete
		return m, nil
	}
	return m, nil
}

// handleBatchSyncSelected opens batch sync confirmation for selected entries.
func (m Model) handleBatchSyncSelected() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.loading && len(m.selectedEntries) > 0 {
		m.currentView = ViewConfirmBatchSync
		return m, nil
	}
	return m, nil
}

// handleSyncEntry opens sync confirmation for an orphaned entry.
func (m Model) handleSyncEntry() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching && !m.loading {
		filteredEntries := m.getFilteredEntries()
		if m.cursor < len(filteredEntries) {
			entry := filteredEntries[m.cursor]
			if entry.Status == diff.StatusOrphanedDNS || entry.Status == diff.StatusOrphanedCaddy {
				m.sync.Entry = &entry
				m.currentView = ViewConfirmSync
				return m, nil
			}
		}
	}
	return m, nil
}

// handleOpenMigrationWizard opens the migration wizard.
func (m Model) handleOpenMigrationWizard() (Model, tea.Cmd) {
	if m.currentView == ViewList && m.config != nil && !m.loading {
		if err := m.startMigrationWizard(); err != nil {
			m.err = err
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

// handleOpenHelp opens the help screen.
func (m Model) handleOpenHelp() (Model, tea.Cmd) {
	if !m.searching && !m.loading {
		m.currentView = ViewHelp
		m.helpPage = 0
		return m, nil
	}
	return m, nil
}

// handleHelpPageJump jumps to a specific help page (1-5).
func (m Model) handleHelpPageJump(key string) (Model, tea.Cmd) {
	if m.currentView == ViewHelp {
		pageNum := int(key[0] - '0')
		if pageNum >= 1 && pageNum <= 5 {
			m.helpPage = pageNum - 1
		}
		return m, nil
	}
	return m, nil
}

// handleQuit quits the application from list view.
func (m Model) handleQuit() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching {
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// handleCancelConfirmation cancels confirmation dialogs with 'n'.
func (m Model) handleCancelConfirmation() (Model, tea.Cmd) {
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("n")
	}
	if m.currentView == ViewConfirmBulkDelete {
		m.currentView = ViewList
		m.bulkDelete.Entries = nil
		m.err = nil
		return m, nil
	}
	if m.currentView == ViewConfirmDelete {
		m.currentView = ViewList
		m.err = nil
		return m, nil
	}
	if m.currentView == ViewConfirmSync {
		m.currentView = ViewList
		m.err = nil
		return m, nil
	}
	if m.currentView == ViewConfirmBatchDelete {
		m.currentView = ViewList
		m.err = nil
		return m, nil
	}
	if m.currentView == ViewConfirmBatchSync {
		m.currentView = ViewList
		m.err = nil
		return m, nil
	}
	if m.currentView == ViewConfirmRestore {
		m.currentView = ViewBackupManager
		m.err = nil
		return m, nil
	}
	if m.currentView == ViewConfirmCleanup {
		m.currentView = ViewBackupManager
		m.err = nil
		return m, nil
	}
	return m, nil
}

// handleBackupPageUp pages up in backup preview.
func (m Model) handleBackupPageUp() (Model, tea.Cmd) {
	if m.currentView == ViewBackupPreview && !m.loading {
		visibleHeight := m.height - 12
		if visibleHeight < 5 {
			visibleHeight = 5
		}
		m.backup.PreviewScroll -= visibleHeight
		if m.backup.PreviewScroll < 0 {
			m.backup.PreviewScroll = 0
		}
	}
	return m, nil
}

// handleBackupPageDown pages down in backup preview.
func (m Model) handleBackupPageDown() (Model, tea.Cmd) {
	if m.currentView == ViewBackupPreview && !m.loading {
		content, err := os.ReadFile(m.backup.PreviewPath)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			visibleHeight := m.height - 12
			if visibleHeight < 5 {
				visibleHeight = 5
			}
			m.backup.PreviewScroll += visibleHeight
			maxScroll := len(lines) - visibleHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.backup.PreviewScroll > maxScroll {
				m.backup.PreviewScroll = maxScroll
			}
		}
	}
	return m, nil
}

// handleListHome jumps to the start of the list.
func (m Model) handleListHome() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching {
		m.cursor = 0
		m.scrollOffset = 0
	}
	return m, nil
}

// handleListEnd jumps to the end of the list.
func (m Model) handleListEnd() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching {
		filtered := m.getFilteredEntries()
		m.cursor = len(filtered) - 1
		if len(filtered) > m.height-5 {
			m.scrollOffset = len(filtered) - (m.height - 5)
		}
	}
	return m, nil
}

// handleListDown navigates down in list view (j key, entries only).
func (m Model) handleListDown() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching {
		if m.cursor < len(m.getFilteredEntries())-1 {
			m.cursor++
			if m.cursor >= m.scrollOffset+m.height-5 {
				m.scrollOffset++
			}
		}
		return m, nil
	}
	return m, nil
}

// handleListUp navigates up in list view (k key, entries only).
func (m Model) handleListUp() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching {
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scrollOffset {
				m.scrollOffset--
			}
		}
		return m, nil
	}
	return m, nil
}

