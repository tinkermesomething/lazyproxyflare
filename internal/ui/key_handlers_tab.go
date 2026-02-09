package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// handleTabKey handles the Tab key across all views.
func (m Model) handleTabKey() (Model, tea.Cmd) {
	// In profile edit: navigate fields
	if m.currentView == ViewProfileEdit {
		return m.handleProfileEditKeyPress("tab")
	}
	// In setup wizard: navigate down through selections
	if m.currentView == ViewWizard {
		return m.handleWizardKeyPress("tab")
	}
	// In snippet wizard: navigate forward through fields/options
	if m.currentView == ViewSnippetWizard {
		switch m.snippetWizardStep {
		case SnippetWizardWelcome:
			// Navigate wizard mode selection (3 modes)
			if m.wizardCursor < 2 {
				m.wizardCursor++
			}
			return m, nil
		case snippet_wizard.StepTemplateSelection:
			// Navigate template list (15 templates)
			if m.wizardCursor < 14 {
				m.wizardCursor++
			}
			return m, nil
		case snippet_wizard.StepCustomSnippet:
			// Custom mode: cycle between name (0) and content (1)
			m.wizardCursor = (m.wizardCursor + 1) % 2
			// Update focus on textinput components
			if m.wizardCursor == 0 {
				m.snippetWizardData.CustomNameInput.Focus()
				m.snippetWizardData.CustomContentInput.Blur()
			} else {
				m.snippetWizardData.CustomNameInput.Blur()
				m.snippetWizardData.CustomContentInput.Focus()
			}
			return m, nil
		case snippet_wizard.StepIPRestriction:
			// IP Restriction: cycle between LAN subnet (0) and External IP (1)
			m.wizardCursor = (m.wizardCursor + 1) % 2
			return m, nil
		case snippet_wizard.StepTemplateParams:
			// Template params: navigate forward to next field
			maxCursor := m.getTemplateParamsCursorMax()
			if m.wizardCursor < maxCursor {
				m.wizardCursor++
			}
			return m, nil
		}
	}
	// In list view: switch between Cloudflare/Caddy tabs
	if m.currentView == ViewList {
		if m.activeTab == TabCloudflare {
			m.activeTab = TabCaddy
		} else {
			m.activeTab = TabCloudflare
		}
		// Reset cursor and scroll for new filtered list
		m.cursor = 0
		m.scrollOffset = 0
		return m, nil
	}
	// Navigate forward through form fields (in add/edit form)
	if m.currentView == ViewAdd || m.currentView == ViewEdit {
		// Smart navigation: skip Caddy fields if DNS-only mode
		if m.addForm.DNSOnly {
			// DNS-only mode: cycle through 0->1->2->3->6
			switch m.addForm.FocusedField {
			case 0:
				m.addForm.FocusedField = 1
			case 1:
				m.addForm.FocusedField = 2
			case 2:
				m.addForm.FocusedField = 3
			case 3:
				m.addForm.FocusedField = 6
			case 6:
				m.addForm.FocusedField = 0
			default:
				m.addForm.FocusedField = 0
			}
		} else {
			// Full mode: cycle through all fields (0-7 + snippets + custom config)
			maxFields := 8 + len(m.snippets) + 1
			m.addForm.FocusedField = (m.addForm.FocusedField + 1) % maxFields
		}
		return m, nil
	}
	return m, nil
}

// handleShiftTabKey handles the Shift+Tab key across all views.
func (m Model) handleShiftTabKey() (Model, tea.Cmd) {
	// In setup wizard: navigate up through selections
	if m.currentView == ViewWizard {
		return m.handleWizardKeyPress("shift+tab")
	}
	// In snippet wizard: navigate backward through fields/options
	if m.currentView == ViewSnippetWizard {
		switch m.snippetWizardStep {
		case SnippetWizardWelcome:
			// Navigate wizard mode selection backward
			if m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		case snippet_wizard.StepTemplateSelection:
			// Navigate template list backward
			if m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		case snippet_wizard.StepCustomSnippet:
			// Custom mode: cycle backward between content (1) and name (0)
			m.wizardCursor = (m.wizardCursor - 1 + 2) % 2
			// Update focus on textinput components
			if m.wizardCursor == 0 {
				m.snippetWizardData.CustomNameInput.Focus()
				m.snippetWizardData.CustomContentInput.Blur()
			} else {
				m.snippetWizardData.CustomNameInput.Blur()
				m.snippetWizardData.CustomContentInput.Focus()
			}
			return m, nil
		case snippet_wizard.StepIPRestriction:
			// IP Restriction: cycle backward between External IP (1) and LAN subnet (0)
			m.wizardCursor = (m.wizardCursor - 1 + 2) % 2
			return m, nil
		case snippet_wizard.StepTemplateParams:
			// Template params: navigate backward to previous field
			if m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		}
	}
	// In list view: cycle backward through panels
	if m.currentView == ViewList {
		switch m.panelFocus {
		case PanelFocusLeft:
			m.panelFocus = PanelFocusSnippets
		case PanelFocusSnippets:
			m.panelFocus = PanelFocusDetails
		case PanelFocusDetails:
			m.panelFocus = PanelFocusLeft
		}
		return m, nil
	}
	// Navigate backward through form fields (in add/edit form)
	if m.currentView == ViewAdd || m.currentView == ViewEdit {
		// Smart navigation: skip Caddy fields if DNS-only mode
		if m.addForm.DNSOnly {
			// DNS-only mode: cycle backward through 6->3->2->1->0
			switch m.addForm.FocusedField {
			case 0:
				m.addForm.FocusedField = 6
			case 1:
				m.addForm.FocusedField = 0
			case 2:
				m.addForm.FocusedField = 1
			case 3:
				m.addForm.FocusedField = 2
			case 6:
				m.addForm.FocusedField = 3
			default:
				m.addForm.FocusedField = 6
			}
		} else {
			// Full mode: cycle backward through all fields (0-7 + snippets + custom config)
			maxFields := 8 + len(m.snippets) + 1
			m.addForm.FocusedField = (m.addForm.FocusedField - 1 + maxFields) % maxFields
		}
		return m, nil
	}
	return m, nil
}
