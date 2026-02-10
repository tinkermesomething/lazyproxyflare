package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// handleSearchStart enters search mode from list view or audit log.
func (m Model) handleSearchStart() (Model, tea.Cmd) {
	if m.currentView == ViewAuditLog {
		m.audit.SearchActive = true
		m.audit.SearchQuery = ""
		m.audit.Scroll = 0
		return m, nil
	}
	if m.currentView == ViewList && !m.searching && !m.loading {
		m.searching = true
		m.searchQuery = ""
		return m, nil
	}
	return m, nil
}

// handleStatusFilterCycle cycles through status filters or audit op filter.
func (m Model) handleStatusFilterCycle() (Model, tea.Cmd) {
	if m.currentView == ViewAuditLog && !m.audit.SearchActive {
		m.cycleOpFilter()
		m.audit.Scroll = 0
		return m, nil
	}
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
	if m.currentView == ViewConfirmDeleteProfile {
		m.profile.DeleteProfileName = ""
		m.currentView = ViewProfileSelector
		m.err = nil
		return m, nil
	}
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

// handleBackspaceKey handles the Backspace key across all views.
func (m Model) handleBackspaceKey() (Model, tea.Cmd) {
	// Handle backspace in profile edit (only when editing a field)
	if m.currentView == ViewProfileEdit && m.profile.EditingField {
		return m.handleProfileEditKeyPress("backspace")
	}
	// Handle backspace in add/edit form
	if m.currentView == ViewAdd || m.currentView == ViewEdit {
		// Check if we're in the custom config field
		customConfigFieldIndex := 8 + len(m.snippets)
		if m.addForm.FocusedField == customConfigFieldIndex {
			// Custom Caddy Config field
			if len(m.addForm.CustomCaddyConfig) > 0 {
				m.addForm.CustomCaddyConfig = m.addForm.CustomCaddyConfig[:len(m.addForm.CustomCaddyConfig)-1]
			}
			return m, nil
		}

		// Handle other fields
		switch m.addForm.FocusedField {
		case 0: // Subdomain
			if len(m.addForm.Subdomain) > 0 {
				m.addForm.Subdomain = m.addForm.Subdomain[:len(m.addForm.Subdomain)-1]
			}
		case 2: // DNS Target
			if len(m.addForm.DNSTarget) > 0 {
				m.addForm.DNSTarget = m.addForm.DNSTarget[:len(m.addForm.DNSTarget)-1]
			}
		case 4: // Reverse Proxy Target
			if len(m.addForm.ReverseProxyTarget) > 0 {
				m.addForm.ReverseProxyTarget = m.addForm.ReverseProxyTarget[:len(m.addForm.ReverseProxyTarget)-1]
			}
		case 5: // Service Port
			if len(m.addForm.ServicePort) > 0 {
				m.addForm.ServicePort = m.addForm.ServicePort[:len(m.addForm.ServicePort)-1]
			}
		}
		return m, nil
	}

	// Handle backspace in search mode
	if m.searching && len(m.searchQuery) > 0 {
		m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil
	}

	// Handle backspace in snippet wizard IP restriction step
	if m.currentView == ViewSnippetWizard && m.snippetWizardStep == SnippetWizardIPRestriction && m.snippetWizardData.CreateIPRestriction {
		if m.wizardCursor == 0 {
			// LAN Subnet field
			if len(m.snippetWizardData.LANSubnet) > 0 {
				m.snippetWizardData.LANSubnet = m.snippetWizardData.LANSubnet[:len(m.snippetWizardData.LANSubnet)-1]
			}
			return m, nil
		} else if m.wizardCursor == 1 {
			// External IP field
			if len(m.snippetWizardData.AllowedExternalIP) > 0 {
				m.snippetWizardData.AllowedExternalIP = m.snippetWizardData.AllowedExternalIP[:len(m.snippetWizardData.AllowedExternalIP)-1]
			}
			return m, nil
		}
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

// handleOpenSnippetWizard opens the snippet wizard from list view.
func (m Model) handleOpenSnippetWizard() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching && !m.loading {
		// Run auto-detection on Caddyfile
		detectedPatterns, err := snippet_wizard.DetectPatternsFromFile(m.config.Caddy.CaddyfilePath)
		if err != nil {
			// If detection fails, continue with empty patterns
			detectedPatterns = []snippet_wizard.DetectedPattern{}
		}

		// Initialize snippet wizard
		m.snippetWizardStep = SnippetWizardWelcome
		m.snippetWizardData = SnippetWizardData{
			// Set defaults from config
			LANSubnet:         m.config.Defaults.LANSubnet,
			AllowedExternalIP: m.config.Defaults.AllowedExternalIP,
			SecurityPreset:    "basic", // default to basic

			// Auto-detection data
			DetectedPatterns: detectedPatterns,
			SelectedPatterns: make(map[string]bool),

			// Template mode data
			SelectedTemplates: make(map[string]bool),
			SnippetConfigs:    make(map[string]snippet_wizard.SnippetConfig),

			// Custom mode data
			CustomSnippetName:    "",
			CustomSnippetContent: "",
		}
		m.wizardCursor = 0
		m.currentView = ViewSnippetWizard
		return m, nil
	}
	return m, nil
}

// handleOpenEditor opens the Caddyfile in an external editor.
func (m Model) handleOpenEditor() (Model, tea.Cmd) {
	if m.currentView != ViewList || m.activeTab != TabCaddy || m.config == nil || m.loading {
		return m, nil
	}

	// Resolve editor: profile setting → $EDITOR env → prompt user
	editor := m.config.UI.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		// No editor configured — prompt user to set one
		m.profile.EditorInput = ""
		m.currentView = ViewSetEditor
		m.err = nil
		return m, nil
	}

	caddyfilePath := m.config.Caddy.CaddyfilePath
	if caddyfilePath == "" {
		m.err = fmt.Errorf("no Caddyfile path configured")
		return m, nil
	}

	c := exec.Command(editor, caddyfilePath)
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

// handleDeleteAction handles the 'd' key for deleting snippets, backups, or entries.
func (m Model) handleDeleteAction() (Model, tea.Cmd) {
	// Delete profile from profile selector
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("d")
	}
	// Delete snippet when in edit mode
	if m.currentView == ViewSnippetDetail && m.snippetPanel.Editing {
		return m.deleteSnippet()
	}
	// Delete backup from preview mode
	if m.currentView == ViewBackupPreview && !m.loading {
		backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
		if err == nil && m.backup.Cursor < len(backups) {
			m.loading = true
			m.currentView = ViewBackupManager // Return to manager after deletion
			return m, deleteBackupCmd(backups[m.backup.Cursor].Path)
		}
		return m, nil
	}
	// Delete selected entry (from list view)
	if m.currentView == ViewList && !m.searching && !m.loading && len(m.getFilteredEntries()) > 0 {
		m.delete.EntryIndex = m.cursor
		m.delete.ScopeCursor = 0         // Reset cursor
		m.delete.Scope = DeleteAll       // Default to delete all
		m.currentView = ViewDeleteScope // Go to scope selection first
		return m, nil
	}
	return m, nil
}

// handleProfileOrPreview handles 'p/ctrl+p' for profile selector or backup preview.
func (m Model) handleProfileOrPreview() (Model, tea.Cmd) {
	// Handle in profile selector
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("p")
	}
	// Open profile selector from main view
	if m.currentView == ViewList {
		m.cursor = 0
		// Load available profiles
		profiles, err := config.ListProfiles()
		if err == nil {
			m.profile.Available = profiles
		}
		m.currentView = ViewProfileSelector
		return m, nil
	}
	// Preview backup (from backup manager)
	if m.currentView == ViewBackupManager && !m.loading {
		backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
		if err == nil && m.backup.Cursor < len(backups) {
			m.backup.PreviewPath = backups[m.backup.Cursor].Path
			m.backup.PreviewScroll = 0 // Reset scroll position
			m.currentView = ViewBackupPreview
		}
		return m, nil
	}
	return m, nil
}

// handleRestoreBackup handles 'R' for restoring a backup.
func (m Model) handleRestoreBackup() (Model, tea.Cmd) {
	if (m.currentView == ViewBackupManager || m.currentView == ViewBackupPreview) && !m.loading {
		backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
		if err == nil && m.backup.Cursor < len(backups) {
			m.backup.PreviewPath = backups[m.backup.Cursor].Path
			m.backup.RestoreScopeCursor = 0    // Reset cursor
			m.backup.RestoreScope = RestoreAll // Default to restore all
			m.currentView = ViewRestoreScope
		}
		return m, nil
	}
	return m, nil
}

// handleNavigateRight handles the right arrow key for help pages and backup nav.
func (m Model) handleNavigateRight() (Model, tea.Cmd) {
	// Navigate to next help page
	if m.currentView == ViewHelp && m.helpPage < 4 {
		m.helpPage++
		return m, nil
	}
	// Navigate to next backup in preview
	if m.currentView == ViewBackupPreview && !m.loading {
		backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
		if err == nil && m.backup.Cursor < len(backups)-1 {
			m.backup.Cursor++
			m.backup.PreviewPath = backups[m.backup.Cursor].Path
			m.backup.PreviewScroll = 0 // Reset scroll position
		}
		return m, nil
	}
	return m, nil
}

// handleNavigateLeft handles the left arrow key for help pages and backup nav.
func (m Model) handleNavigateLeft() (Model, tea.Cmd) {
	// Navigate to previous help page
	if m.currentView == ViewHelp && m.helpPage > 0 {
		m.helpPage--
		return m, nil
	}
	// Navigate to previous backup in preview
	if m.currentView == ViewBackupPreview && !m.loading {
		backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
		if err == nil && m.backup.Cursor > 0 {
			m.backup.Cursor--
			m.backup.PreviewPath = backups[m.backup.Cursor].Path
			m.backup.PreviewScroll = 0 // Reset scroll position
		}
		return m, nil
	}
	return m, nil
}

// handleImportProfile handles 'i' for importing a profile.
func (m Model) handleImportProfile() (Model, tea.Cmd) {
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("i")
	}
	return m, nil
}

// handleDeleteBackup handles 'x' for deleting a backup or exporting a profile.
func (m Model) handleDeleteBackup() (Model, tea.Cmd) {
	// Export profile from profile selector
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("x")
	}
	if m.currentView == ViewBackupManager && !m.loading {
		backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
		if err == nil && m.backup.Cursor < len(backups) {
			m.loading = true
			return m, deleteBackupCmd(backups[m.backup.Cursor].Path)
		}
		return m, nil
	}
	return m, nil
}

// handleCleanupBackups handles 'c' for cleanup confirmation.
func (m Model) handleCleanupBackups() (Model, tea.Cmd) {
	if m.currentView == ViewBackupManager && !m.loading {
		m.currentView = ViewConfirmCleanup
		return m, nil
	}
	return m, nil
}

// handleRefreshData handles 'r' for refreshing data, cycling audit result filter, or dismissing errors.
func (m Model) handleRefreshData() (Model, tea.Cmd) {
	// Cycle result filter in audit log
	if m.currentView == ViewAuditLog && !m.audit.SearchActive {
		m.cycleResultFilter()
		m.audit.Scroll = 0
		return m, nil
	}
	// If showing error modal, retry by clearing error and returning to previous view
	if m.currentView == ViewError {
		m.currentView = m.previousView
		m.err = nil
		return m, nil
	}
	// Refresh data (only from list view, not while searching or loading)
	if m.currentView == ViewList && !m.searching && !m.loading {
		m.loading = true
		return m, refreshDataCmd(m.config)
	}
	return m, nil
}

// handleGoToTop handles 'g' for going to top of list or backup preview.
func (m Model) handleGoToTop() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching {
		if m.panelFocus == PanelFocusSnippets {
			m.snippetPanel.Cursor = 0
			m.snippetPanel.ScrollOffset = 0
		} else {
			m.cursor = 0
			m.scrollOffset = 0
		}
	}
	// In backup preview: scroll to top
	if m.currentView == ViewBackupPreview && !m.loading {
		m.backup.PreviewScroll = 0
	}
	return m, nil
}

// handleGoToBottom handles 'G' for going to bottom of list or backup preview.
func (m Model) handleGoToBottom() (Model, tea.Cmd) {
	if m.currentView == ViewList && !m.searching {
		if m.panelFocus == PanelFocusSnippets {
			if len(m.snippets) > 0 {
				m.snippetPanel.Cursor = len(m.snippets) - 1
				// Calculate scroll offset (if needed in future)
			}
		} else {
			filtered := m.getFilteredEntries()
			m.cursor = len(filtered) - 1
			if len(filtered) > m.height-5 {
				m.scrollOffset = len(filtered) - (m.height - 5)
			}
		}
	}
	// In backup preview: scroll to bottom
	if m.currentView == ViewBackupPreview && !m.loading {
		content, err := os.ReadFile(m.backup.PreviewPath)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			visibleHeight := m.height - 12
			if visibleHeight < 5 {
				visibleHeight = 5
			}
			// Scroll to show the last page
			maxScroll := len(lines) - visibleHeight
			if maxScroll > 0 {
				m.backup.PreviewScroll = maxScroll
			} else {
				m.backup.PreviewScroll = 0
			}
		}
	}
	return m, nil
}

// handleEditProfile handles 'e' in profile selector.
func (m Model) handleEditProfile() (Model, tea.Cmd) {
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("e")
	}
	return m, nil
}

// handleAddProfile handles '+' in profile selector.
func (m Model) handleAddProfile() (Model, tea.Cmd) {
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("+")
	}
	return m, nil
}

