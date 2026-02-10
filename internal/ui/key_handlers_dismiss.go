package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// handleDismiss dispatches 'esc'/'ctrl+w' key per-view.
func (m Model) handleDismiss() (Model, tea.Cmd) {
	// If showing error modal, dismiss it and return to previous view
	if m.currentView == ViewError {
		m.currentView = m.previousView
		m.err = nil
		return m, nil
	}
	// If in confirm delete profile, return to profile selector
	if m.currentView == ViewConfirmDeleteProfile {
		m.profile.DeleteProfileName = ""
		m.currentView = ViewProfileSelector
		m.err = nil
		return m, nil
	}
	// If in bulk delete menu, return to list
	if m.currentView == ViewBulkDeleteMenu {
		m.currentView = ViewList
		m.bulkDelete.MenuCursor = 0
		return m, nil
	}
	// If in confirm bulk delete, return to list
	if m.currentView == ViewConfirmBulkDelete {
		m.currentView = ViewList
		m.bulkDelete.Entries = nil
		m.err = nil // Clear any error
		return m, nil
	}
	// If in confirm batch delete, return to list
	if m.currentView == ViewConfirmBatchDelete {
		m.currentView = ViewList
		m.err = nil // Clear any error
		return m, nil
	}
	// If in confirm batch sync, return to list
	if m.currentView == ViewConfirmBatchSync {
		m.currentView = ViewList
		m.err = nil // Clear any error
		return m, nil
	}
	// If in confirm delete, return to list
	if m.currentView == ViewConfirmDelete {
		m.currentView = ViewList
		m.err = nil // Clear any error
		return m, nil
	}
	// If in confirm sync, return to list
	if m.currentView == ViewConfirmSync {
		m.currentView = ViewList
		m.err = nil // Clear any error
		return m, nil
	}
	// If in backup manager, return to list
	if m.currentView == ViewBackupManager {
		m.currentView = ViewList
		m.err = nil
		return m, nil
	}
	// If in backup preview, return to backup manager
	if m.currentView == ViewBackupPreview {
		m.currentView = ViewBackupManager
		m.err = nil
		return m, nil
	}
	// If in delete scope selection, return to list
	if m.currentView == ViewDeleteScope {
		m.currentView = ViewList
		m.err = nil
		return m, nil
	}
	// If in confirm delete, return to delete scope selection
	if m.currentView == ViewConfirmDelete {
		m.currentView = ViewDeleteScope
		m.err = nil
		return m, nil
	}
	// If in restore scope selection, return to backup manager
	if m.currentView == ViewRestoreScope {
		m.currentView = ViewBackupManager
		m.err = nil
		return m, nil
	}
	// If in confirm restore, return to restore scope selection
	if m.currentView == ViewConfirmRestore {
		m.currentView = ViewRestoreScope
		m.err = nil
		return m, nil
	}
	// If in confirm cleanup, return to backup manager
	if m.currentView == ViewConfirmCleanup {
		m.currentView = ViewBackupManager
		m.err = nil
		return m, nil
	}
	// If in snippet detail view, handle edit mode or return to list
	if m.currentView == ViewSnippetDetail {
		if m.snippetPanel.Editing {
			// Cancel edit mode, return to view mode
			m.snippetPanel.Editing = false
			m.err = nil
			return m, nil
		}
		// Return to list from view mode
		m.currentView = ViewList
		m.err = nil
		return m, nil
	}
	// If in audit log view, return to list
	if m.currentView == ViewAuditLog {
		m.currentView = ViewList
		m.err = nil
		return m, nil
	}
	// If in snippet wizard, ESC goes back one step (Ctrl+Q to close)
	if m.currentView == ViewSnippetWizard {
		m.err = nil
		switch m.snippetWizardStep {
		case snippet_wizard.StepWelcome:
			// On welcome screen, ESC closes the wizard
			m.currentView = ViewList
			m.snippetWizardData = SnippetWizardData{}
		case snippet_wizard.StepTemplateSelection:
			// Back to welcome
			m.snippetWizardStep = snippet_wizard.StepWelcome
			m.wizardCursor = 0
		case snippet_wizard.StepTemplateParams:
			// Back to template selection
			m.snippetWizardStep = snippet_wizard.StepTemplateSelection
			m.wizardCursor = 0
		case snippet_wizard.StepCustomSnippet:
			// Back to welcome
			m.snippetWizardStep = snippet_wizard.StepWelcome
			m.wizardCursor = 0
		case snippet_wizard.StepIPRestriction:
			// Guided mode: back to welcome
			m.snippetWizardStep = snippet_wizard.StepWelcome
			m.wizardCursor = 0
		case snippet_wizard.StepSecurityHeaders:
			// Guided mode: back to IP restriction
			m.snippetWizardStep = snippet_wizard.StepIPRestriction
			m.wizardCursor = 0
		case snippet_wizard.StepPerformance:
			// Guided mode: back to security headers
			m.snippetWizardStep = snippet_wizard.StepSecurityHeaders
			m.wizardCursor = 0
		case snippet_wizard.StepSummary:
			// Back to previous step based on mode
			switch m.snippetWizardData.Mode {
			case snippet_wizard.ModeTemplated:
				m.snippetWizardStep = snippet_wizard.StepTemplateParams
			case snippet_wizard.ModeCustom:
				m.snippetWizardStep = snippet_wizard.StepCustomSnippet
			case snippet_wizard.ModeGuided:
				m.snippetWizardStep = snippet_wizard.StepPerformance
			default:
				m.snippetWizardStep = snippet_wizard.StepWelcome
			}
			m.wizardCursor = 0
		default:
			// Unknown step, go to welcome
			m.snippetWizardStep = snippet_wizard.StepWelcome
			m.wizardCursor = 0
		}
		return m, nil
	}
	// If in preview, go back to add or edit form
	if m.currentView == ViewPreview {
		if m.editingEntry != nil {
			m.currentView = ViewEdit
		} else {
			m.currentView = ViewAdd
		}
		m.err = nil // Clear any error
		return m, nil
	}
	// If in add form, return to list and clear editing state
	if m.currentView == ViewAdd {
		m.currentView = ViewList
		m.editingEntry = nil
		m.err = nil // Clear any error
		return m, nil
	}
	// If in edit form, return to list and clear editing state
	if m.currentView == ViewEdit {
		m.currentView = ViewList
		m.editingEntry = nil
		m.err = nil // Clear any error
		return m, nil
	}
	// If in search mode, exit search mode
	if m.searching {
		m.searching = false
		m.searchQuery = ""
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil
	}
	// If in list view with any active filters/sort/selections, reset everything
	if m.currentView == ViewList && (m.searchQuery != "" || m.statusFilter != FilterAll || m.dnsTypeFilter != DNSTypeAll || m.sortMode != SortAlphabetical || len(m.selectedEntries) > 0) {
		m.searchQuery = ""
		m.statusFilter = FilterAll
		m.dnsTypeFilter = DNSTypeAll
		m.sortMode = SortAlphabetical
		m.selectedEntries = make(map[string]bool)
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil
	}
	// If in wizard view, handle wizard ESC
	if m.currentView == ViewWizard {
		return m.handleWizardKeyPress("esc")
	}
	// If in migration wizard, handle migration ESC
	if m.currentView == ViewMigrationWizard {
		return m.handleMigrationWizardKeyPress("esc")
	}
	// If in profile selector, return to list
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("esc")
	}
	// If in profile edit, return to selector
	if m.currentView == ViewProfileEdit {
		return m.handleProfileEditKeyPress("esc")
	}
	// Otherwise go back to list view
	m.currentView = ViewList
	return m, nil
}
