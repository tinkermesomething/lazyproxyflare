package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// handleSpaceKey handles the Space key across all views (toggles and selections).
func (m Model) handleSpaceKey() (Model, tea.Cmd) {
	// In profile edit: toggle boolean fields
	if m.currentView == ViewProfileEdit {
		return m.handleProfileEditKeyPress(" ")
	}
	// In snippet wizard: space toggles checkboxes
	if m.currentView == ViewSnippetWizard {
		switch m.snippetWizardStep {
		case snippet_wizard.StepTemplateSelection:
			// Toggle template selection at cursor
			// Template order matches template_selection.go view
			templateKeys := []string{
				// Security
				"cors_headers", "rate_limiting", "auth_headers", "ip_restricted", "security_headers",
				// Performance
				"static_caching", "compression_advanced", "performance",
				// Backend Integration
				"websocket_advanced", "extended_timeouts", "https_backend",
				// Content Control
				"large_uploads", "custom_headers_inject", "frame_embedding", "rewrite_rules",
			}
			if m.wizardCursor >= 0 && m.wizardCursor < len(templateKeys) {
				key := templateKeys[m.wizardCursor]
				m.snippetWizardData.SelectedTemplates[key] = !m.snippetWizardData.SelectedTemplates[key]
			}
			return m, nil
		case SnippetWizardAutoDetect:
			// Toggle pattern selection at cursor
			if m.wizardCursor >= 0 && m.wizardCursor < len(m.snippetWizardData.DetectedPatterns) {
				pattern := m.snippetWizardData.DetectedPatterns[m.wizardCursor]
				m.snippetWizardData.SelectedPatterns[pattern.SuggestedName] = !m.snippetWizardData.SelectedPatterns[pattern.SuggestedName]
			}
			return m, nil
		case SnippetWizardIPRestriction:
			m.snippetWizardData.CreateIPRestriction = !m.snippetWizardData.CreateIPRestriction
			return m, nil
		case SnippetWizardSecurityHeaders:
			m.snippetWizardData.CreateSecurityHeaders = !m.snippetWizardData.CreateSecurityHeaders
			return m, nil
		case SnippetWizardPerformance:
			m.snippetWizardData.CreatePerformance = !m.snippetWizardData.CreatePerformance
			return m, nil
		case snippet_wizard.StepTemplateParams:
			// Toggle checkbox parameters in template config
			m.toggleSnippetTemplateParamCheckbox()
			return m, nil
		}
	}
	// In list view: space toggles selection
	if m.currentView == ViewList && !m.searching && !m.loading {
		filteredEntries := m.getFilteredEntries()
		if m.cursor < len(filteredEntries) {
			domain := filteredEntries[m.cursor].Domain
			// Toggle selection
			if m.selectedEntries[domain] {
				delete(m.selectedEntries, domain)
			} else {
				m.selectedEntries[domain] = true
			}
		}
		return m, nil
	}
	// In add/edit form: space toggles DNS type and checkboxes
	if m.currentView == ViewAdd || m.currentView == ViewEdit {
		switch m.addForm.FocusedField {
		case 1: // DNS Type toggle
			if m.addForm.DNSType == "CNAME" {
				m.addForm.DNSType = "A"
			} else {
				m.addForm.DNSType = "CNAME"
			}
		case 3: // DNS Only checkbox
			m.addForm.DNSOnly = !m.addForm.DNSOnly
		case 6: // Proxied checkbox
			m.addForm.Proxied = !m.addForm.Proxied
		case 7: // SSL checkbox
			m.addForm.SSL = !m.addForm.SSL
		default:
			// Fields 8+ are snippet checkboxes
			if m.addForm.FocusedField >= 8 {
				snippetIndex := m.addForm.FocusedField - 8
				if snippetIndex < len(m.snippets) {
					snippetName := m.snippets[snippetIndex].Name
					// Initialize map if nil
					if m.addForm.SelectedSnippets == nil {
						m.addForm.SelectedSnippets = make(map[string]bool)
					}
					// Toggle snippet selection
					m.addForm.SelectedSnippets[snippetName] = !m.addForm.SelectedSnippets[snippetName]
				}
			}
		}
		return m, nil
	}
	return m, nil
}
