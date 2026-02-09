package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/diff"
	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// handleEnterKey handles the Enter key across all views.
func (m Model) handleEnterKey() (Model, tea.Cmd) {
	// Edit selected entry (from list view)
	if m.currentView == ViewList && !m.searching && !m.loading {
		// Skip if snippets panel is focused (no action needed)
		if m.panelFocus == PanelFocusSnippets {
			return m, nil
		}

		// Edit DNS entry
		filteredEntries := m.getFilteredEntries()
		if m.cursor < len(filteredEntries) {
			entry := filteredEntries[m.cursor]

			// Pre-populate form with existing entry values
			// For multi-domain entries, populate all domains as newline-separated
			subdomain := ""
			if entry.Caddy != nil && len(entry.Caddy.Domains) > 0 {
				// Multi-domain entry - format all domains as newline-separated subdomains
				subdomain = GetSubdomainsTextareaValue(entry.Caddy.Domains, m.config.Domain)
			} else if len(entry.Domain) > len(m.config.Domain)+1 {
				// Single domain (backwards compatibility)
				subdomain = entry.Domain[:len(entry.Domain)-len(m.config.Domain)-1]
			}

			dnsType := "CNAME"
			dnsTarget := m.config.Defaults.CNAMETarget
			proxied := m.config.Defaults.Proxied
			if entry.DNS != nil {
				dnsType = entry.DNS.Type
				dnsTarget = entry.DNS.Content
				proxied = entry.DNS.Proxied
			}

			dnsOnly := entry.Caddy == nil // If no Caddy config, it's DNS-only

			reverseProxyTarget := "localhost"
			servicePort := fmt.Sprintf("%d", m.config.Defaults.Port)
			ssl := m.config.Defaults.SSL
			lanOnly := false
			oauth := false
			webSocket := false
			if entry.Caddy != nil {
				reverseProxyTarget = entry.Caddy.Target
				servicePort = fmt.Sprintf("%d", entry.Caddy.Port)
				ssl = entry.Caddy.SSL
				lanOnly = entry.Caddy.IPRestricted
				oauth = entry.Caddy.OAuthHeaders
				webSocket = entry.Caddy.WebSocket
			}

			// Pre-populate selected snippets from entry's imports
			selectedSnippets := make(map[string]bool)
			if entry.Caddy != nil {
				for _, importName := range entry.Caddy.Imports {
					selectedSnippets[importName] = true
				}
			}

			m.addForm = AddFormData{
				Subdomain:          subdomain,
				DNSType:            dnsType,
				DNSTarget:          dnsTarget,
				DNSOnly:            dnsOnly,
				ReverseProxyTarget: reverseProxyTarget,
				ServicePort:        servicePort,
				Proxied:            proxied,
				LANOnly:            lanOnly,
				SSL:                ssl,
				OAuth:              oauth,
				WebSocket:          webSocket,
				SelectedSnippets:   selectedSnippets,
				FocusedField:       0,
			}
			m.editingEntry = &entry
			m.currentView = ViewEdit
			return m, nil
		}
	}

	// If in snippet detail view mode, enter edit mode
	if m.currentView == ViewSnippetDetail && !m.snippetPanel.Editing {
		if m.snippetPanel.Cursor < len(m.snippets) {
			// Initialize textarea with snippet content
			ta := textarea.New()
			ta.SetWidth(70)
			ta.SetHeight(15)
			ta.SetValue(m.snippets[m.snippetPanel.Cursor].Content)
			ta.Focus()
			m.snippetPanel.EditTextarea = ta
			m.snippetPanel.Editing = true
			m.snippetPanel.EditingIndex = m.snippetPanel.Cursor
			return m, nil
		}
	}

	// Dismiss error modal
	if m.currentView == ViewError {
		m.currentView = m.previousView
		m.err = nil
		return m, nil
	}
	// Save profile edit
	if m.currentView == ViewProfileEdit {
		return m.handleProfileEditKeyPress("enter")
	}
	// Handle Enter in wizard
	if m.currentView == ViewWizard {
		return m.handleWizardKeyPress("enter")
	}
	// Handle Enter in migration wizard
	if m.currentView == ViewMigrationWizard {
		return m.handleMigrationWizardKeyPress("enter")
	}
	// Handle Enter in profile selector
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("enter")
	}
	// Handle Enter in snippet wizard
	if m.currentView == ViewSnippetWizard {
		switch m.snippetWizardStep {
		case snippet_wizard.StepWelcome:
			// Set mode based on cursor selection and advance
			switch m.wizardCursor {
			case 0: // Templated - Advanced
				m.snippetWizardData.Mode = snippet_wizard.ModeTemplated
				m.snippetWizardStep = snippet_wizard.StepTemplateSelection
			case 1: // Custom - Paste Your Own
				m.snippetWizardData.Mode = snippet_wizard.ModeCustom
				m.snippetWizardStep = snippet_wizard.StepCustomSnippet
				// Initialize text input components for custom mode
				m.snippetWizardData.CustomNameInput = textinput.New()
				m.snippetWizardData.CustomNameInput.Placeholder = "snippet_name"
				m.snippetWizardData.CustomNameInput.Focus()
				m.snippetWizardData.CustomContentInput = textarea.New()
				m.snippetWizardData.CustomContentInput.Placeholder = "Paste your Caddy snippet here..."
			case 2: // Guided - Step by Step
				m.snippetWizardData.Mode = snippet_wizard.ModeGuided
				// Guided mode goes through legacy flow
				m.snippetWizardStep = snippet_wizard.StepIPRestriction
			}
			m.wizardCursor = 0
			return m, nil
		case snippet_wizard.StepCustomSnippet:
			// Validate custom snippet has name and content
			if m.snippetWizardData.CustomSnippetName == "" {
				m.err = fmt.Errorf("snippet name is required")
				return m, nil
			}
			if m.snippetWizardData.CustomSnippetContent == "" {
				m.err = fmt.Errorf("snippet content is required")
				return m, nil
			}
			// Set mode to custom and advance to summary
			m.snippetWizardData.Mode = snippet_wizard.ModeCustom
			m.snippetWizardStep = snippet_wizard.StepSummary
			m.wizardCursor = 0
			return m, nil
		case snippet_wizard.StepTemplateSelection:
			// Check if any templates selected
			hasSelection := false
			for _, selected := range m.snippetWizardData.SelectedTemplates {
				if selected {
					hasSelection = true
					break
				}
			}
			if !hasSelection {
				m.err = fmt.Errorf("select at least one template")
				return m, nil
			}
			// Advance to template params or summary
			m.snippetWizardData.Mode = snippet_wizard.ModeTemplated
			m.snippetWizardStep = snippet_wizard.StepTemplateParams
			m.wizardCursor = 0
			return m, nil
		case snippet_wizard.StepTemplateParams:
			// Advance to summary
			m.snippetWizardStep = snippet_wizard.StepSummary
			m.wizardCursor = 0
			return m, nil
		case snippet_wizard.StepIPRestriction:
			// Guided mode: Advance to security headers
			m.snippetWizardStep = snippet_wizard.StepSecurityHeaders
			m.wizardCursor = 0
			// Set default security preset if not set
			if m.snippetWizardData.SecurityPreset == "" {
				m.snippetWizardData.SecurityPreset = "basic"
			}
			return m, nil
		case snippet_wizard.StepSecurityHeaders:
			// Guided mode: Advance to performance
			m.snippetWizardStep = snippet_wizard.StepPerformance
			m.wizardCursor = 0
			return m, nil
		case snippet_wizard.StepPerformance:
			// Guided mode: Advance to summary
			m.snippetWizardStep = snippet_wizard.StepSummary
			m.wizardCursor = 0
			return m, nil
		case snippet_wizard.StepSummary:
			// Create the snippets
			return m.createSnippetsFromWizard()
		}
		return m, nil
	}
	// Handle Enter in delete scope selection: proceed to confirm
	if m.currentView == ViewDeleteScope && !m.loading {
		m.currentView = ViewConfirmDelete
		return m, nil
	}
	// Handle Enter in restore scope selection: proceed to confirm
	if m.currentView == ViewRestoreScope && !m.loading {
		m.currentView = ViewConfirmRestore
		return m, nil
	}
	// In bulk delete menu: select option
	if m.currentView == ViewBulkDeleteMenu {
		// Collect entries based on selection
		if m.bulkDelete.MenuCursor == 0 {
			// Delete all orphaned DNS
			m.bulkDelete.Type = "dns"
			m.bulkDelete.Entries = []diff.SyncedEntry{}
			for _, entry := range m.entries {
				if entry.Status == diff.StatusOrphanedDNS {
					m.bulkDelete.Entries = append(m.bulkDelete.Entries, entry)
				}
			}
		} else if m.bulkDelete.MenuCursor == 1 {
			// Delete all orphaned Caddy
			m.bulkDelete.Type = "caddy"
			m.bulkDelete.Entries = []diff.SyncedEntry{}
			for _, entry := range m.entries {
				if entry.Status == diff.StatusOrphanedCaddy {
					m.bulkDelete.Entries = append(m.bulkDelete.Entries, entry)
				}
			}
		}

		// Only proceed if there are entries to delete
		if len(m.bulkDelete.Entries) > 0 {
			m.currentView = ViewConfirmBulkDelete
		} else {
			// No entries to delete, show error and return to list
			m.err = fmt.Errorf("No %s entries to delete", m.bulkDelete.Type)
			m.currentView = ViewList
		}
		return m, nil
	}

	// In add/edit form: validate and go to preview
	if m.currentView == ViewAdd || m.currentView == ViewEdit {
		// Validate required fields
		if m.addForm.Subdomain == "" {
			m.err = fmt.Errorf("Subdomain is required")
			return m, nil
		}
		if m.addForm.DNSTarget == "" {
			m.err = fmt.Errorf("DNS Target is required")
			return m, nil
		}

		// Validate A record IP address format
		if m.addForm.DNSType == "A" {
			if !isValidIPAddress(m.addForm.DNSTarget) {
				m.err = fmt.Errorf("Invalid IP address format for A record")
				return m, nil
			}
		}

		// Validate Caddy fields if not DNS-only
		if !m.addForm.DNSOnly {
			if m.addForm.ReverseProxyTarget == "" {
				m.err = fmt.Errorf("Reverse Proxy Target is required (or enable DNS Only)")
				return m, nil
			}
		}

		// Clear any previous errors and go to preview
		m.err = nil
		m.currentView = ViewPreview
		return m, nil
	}

	// In search mode, enter accepts the search
	if m.searching {
		m.searching = false
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
	// Preview backup with Enter key (from backup manager)
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
