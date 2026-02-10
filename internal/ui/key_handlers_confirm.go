package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
)

// handleConfirmAction dispatches 'y' key confirmation per-view.
func (m Model) handleConfirmAction() (Model, tea.Cmd) {
	// Confirm import profile
	if m.currentView == ViewConfirmImport {
		if m.profile.ImportPath == "" {
			m.err = fmt.Errorf("please enter a file path")
			return m, nil
		}
		return m, importProfileCmd(m.profile.ImportPath)
	}
	// Confirm profile deletion
	if m.currentView == ViewConfirmDeleteProfile {
		if err := config.DeleteProfile(m.profile.DeleteProfileName); err != nil {
			m.err = fmt.Errorf("failed to delete profile: %v", err)
			m.currentView = ViewProfileSelector
			return m, nil
		}
		// Refresh profile list
		profiles, err := config.ListProfiles()
		if err == nil {
			m.profile.Available = profiles
		}
		m.profile.DeleteProfileName = ""
		m.err = nil
		// If no profiles left, launch wizard
		if len(m.profile.Available) == 0 {
			return m.startWizard(), nil
		}
		// Reset cursor if needed
		if m.cursor >= len(m.profile.Available) {
			m.cursor = len(m.profile.Available) - 1
		}
		m.currentView = ViewProfileSelector
		return m, nil
	}
	// Save snippet changes when in edit mode
	if m.currentView == ViewSnippetDetail && m.snippetPanel.Editing {
		return m.saveSnippetEdit()
	}
	// Handle 'y' (confirm) in wizard summary
	if m.currentView == ViewWizard && m.wizardStep == WizardStepSummary {
		return m.handleWizardSummaryConfirm()
	}
	// Confirm and create/update entry (only in preview screen)
	if m.currentView == ViewPreview && !m.loading {
		m.loading = true
		if m.editingEntry != nil {
			apiToken, err := m.config.GetAPIToken()
			if err != nil {
				m.err = fmt.Errorf("failed to get API token: %w", err)
				return m, nil
			}
			return m, updateEntryCmd(m.config, m.addForm, *m.editingEntry, apiToken)
		} else {
			apiToken, err := m.config.GetAPIToken()
			if err != nil {
				m.err = fmt.Errorf("failed to get API token: %w", err)
				return m, nil
			}
			return m, createEntryCmd(m.config, m.addForm, apiToken)
		}
	}
	// Confirm and delete entry (only in confirm delete screen)
	if m.currentView == ViewConfirmDelete && !m.loading {
		m.loading = true
		filteredEntries := m.getFilteredEntries()
		if m.delete.EntryIndex < len(filteredEntries) {
			apiToken, err := m.config.GetAPIToken()
			if err != nil {
				m.err = fmt.Errorf("failed to get API token: %w", err)
				return m, nil
			}
			return m, deleteEntryCmd(m.config, filteredEntries[m.delete.EntryIndex], m.delete.Scope, apiToken)
		}
	}
	// Confirm and sync entry (only in confirm sync screen)
	if m.currentView == ViewConfirmSync && !m.loading {
		m.loading = true
		if m.sync.Entry != nil {
			apiToken, err := m.config.GetAPIToken()
			if err != nil {
				m.err = fmt.Errorf("failed to get API token: %w", err)
				return m, nil
			}
			return m, syncEntryCmd(m.config, *m.sync.Entry, apiToken)
		}
	}
	// Confirm bulk delete (only in confirm bulk delete screen)
	if m.currentView == ViewConfirmBulkDelete && !m.loading {
		m.loading = true
		if m.bulkDelete.Type == "dns" {
			apiToken, err := m.config.GetAPIToken()
			if err != nil {
				m.err = fmt.Errorf("failed to get API token: %w", err)
				return m, nil
			}
			return m, bulkDeleteDNSCmd(m.config, m.bulkDelete.Entries, apiToken)
		} else if m.bulkDelete.Type == "caddy" {
			return m, bulkDeleteCaddyCmd(m.config, m.bulkDelete.Entries)
		}
	}
	// Confirm batch delete selected (only in confirm batch delete screen)
	if m.currentView == ViewConfirmBatchDelete && !m.loading {
		m.loading = true
		apiToken, err := m.config.GetAPIToken()
		if err != nil {
			m.err = fmt.Errorf("failed to get API token: %w", err)
			return m, nil
		}
		return m, batchDeleteSelectedCmd(m.config, m.entries, m.selectedEntries, apiToken)
	}
	// Confirm batch sync selected (only in confirm batch sync screen)
	if m.currentView == ViewConfirmBatchSync && !m.loading {
		m.loading = true
		apiToken, err := m.config.GetAPIToken()
		if err != nil {
			m.err = fmt.Errorf("failed to get API token: %w", err)
			return m, nil
		}
		return m, batchSyncSelectedCmd(m.config, m.entries, m.selectedEntries, apiToken)
	}
	// Preview backup with y key (from backup manager)
	if m.currentView == ViewBackupManager && !m.loading {
		backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
		if err == nil && m.backup.Cursor < len(backups) {
			m.backup.PreviewPath = backups[m.backup.Cursor].Path
			m.backup.PreviewScroll = 0
			m.currentView = ViewBackupPreview
		}
		return m, nil
	}
	// Confirm restore backup (only in confirm restore screen)
	if m.currentView == ViewConfirmRestore && !m.loading {
		m.loading = true
		apiToken, err := m.config.GetAPIToken()
		if err != nil {
			m.err = fmt.Errorf("failed to get API token: %w", err)
			return m, nil
		}
		return m, restoreBackupCmd(m.config, m.backup.PreviewPath, m.backup.RestoreScope, apiToken)
	}
	// Confirm cleanup old backups (only in confirm cleanup screen)
	if m.currentView == ViewConfirmCleanup && !m.loading {
		m.loading = true
		return m, cleanupBackupsCmd(m.config.Caddy.CaddyfilePath, m.backup.RetentionDays)
	}
	return m, nil
}
