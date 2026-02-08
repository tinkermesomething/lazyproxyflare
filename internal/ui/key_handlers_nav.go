package ui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// handleNavigateDown dispatches 'down' arrow key per-view.
func (m Model) handleNavigateDown() (Model, tea.Cmd) {
	// In wizard: navigate down
	if m.currentView == ViewWizard {
		return m.handleWizardKeyPress("down")
	}
	// In migration wizard: navigate down
	if m.currentView == ViewMigrationWizard {
		return m.handleMigrationWizardKeyPress("down")
	}
	// In snippet wizard: navigate down
	if m.currentView == ViewSnippetWizard {
		switch m.snippetWizardStep {
		case SnippetWizardWelcome:
			// Navigate wizard mode selection (3 modes: Templated, Custom, Guided)
			if m.wizardCursor < 2 {
				m.wizardCursor++
			}
			return m, nil
		case snippet_wizard.StepTemplateSelection:
			// Navigate template selection (15 templates)
			if m.wizardCursor < 14 {
				m.wizardCursor++
			}
			return m, nil
		case snippet_wizard.StepCustomSnippet:
			// Navigate between name (0) and content (1) fields
			if m.wizardCursor < 1 {
				m.wizardCursor++
				// Update focus on components
				m.snippetWizardData.CustomNameInput.Blur()
				m.snippetWizardData.CustomContentInput.Focus()
			}
			return m, nil
		case SnippetWizardAutoDetect:
			// Navigate detected patterns
			if m.wizardCursor < len(m.snippetWizardData.DetectedPatterns)-1 {
				m.wizardCursor++
			}
			return m, nil
		case SnippetWizardIPRestriction:
			// Navigate between LAN subnet (0) and external IP (1) fields
			if m.snippetWizardData.CreateIPRestriction && m.wizardCursor < 1 {
				m.wizardCursor++
			}
			return m, nil
		case SnippetWizardSecurityHeaders:
			// Navigate security presets (3 options: basic, strict, paranoid)
			if m.wizardCursor < 2 {
				m.wizardCursor++
				// Update selected preset based on cursor
				presets := []string{"basic", "strict", "paranoid"}
				m.snippetWizardData.SecurityPreset = presets[m.wizardCursor]
			}
			return m, nil
		case snippet_wizard.StepTemplateParams:
			// Navigate template parameter fields
			maxCursor := m.getTemplateParamsCursorMax()
			if m.wizardCursor < maxCursor {
				m.wizardCursor++
			}
			return m, nil
		}
	}
	// In profile selector: navigate down
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("down")
	}
	// In profile edit: navigate down
	if m.currentView == ViewProfileEdit {
		return m.handleProfileEditKeyPress("down")
	}
	// In backup manager: navigate down
	if m.currentView == ViewBackupManager && !m.loading {
		backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
		if err == nil && m.backup.Cursor < len(backups)-1 {
			m.backup.Cursor++
			// Adjust scroll if needed
			visibleHeight := m.height - 8
			if visibleHeight < 1 {
				visibleHeight = 10
			}
			if m.backup.Cursor >= m.backup.ScrollOffset+visibleHeight {
				m.backup.ScrollOffset++
			}
		}
		return m, nil
	}
	// In delete scope selection: navigate down
	if m.currentView == ViewDeleteScope && !m.loading && m.delete.ScopeCursor < 2 {
		m.delete.ScopeCursor++
		// Update selected scope based on cursor
		m.delete.Scope = DeleteScope(m.delete.ScopeCursor)
		return m, nil
	}
	// In restore scope selection: navigate down
	if m.currentView == ViewRestoreScope && !m.loading && m.backup.RestoreScopeCursor < 2 {
		m.backup.RestoreScopeCursor++
		// Update selected scope based on cursor
		m.backup.RestoreScope = RestoreScope(m.backup.RestoreScopeCursor)
		return m, nil
	}
	// In backup preview: scroll down
	if m.currentView == ViewBackupPreview && !m.loading {
		// Read backup to get line count
		content, err := os.ReadFile(m.backup.PreviewPath)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			visibleHeight := m.height - 12
			if visibleHeight < 5 {
				visibleHeight = 5
			}
			// Only scroll if there's more content below
			if m.backup.PreviewScroll+visibleHeight < len(lines) {
				m.backup.PreviewScroll++
			}
		}
		return m, nil
	}
	// In audit log: scroll down
	if m.currentView == ViewAuditLog && !m.loading {
		// Calculate visible entries based on modal height
		modalHeight := m.height * 2 / 3
		if modalHeight < 15 {
			modalHeight = 15
		}
		reservedLines := 11
		availableHeight := modalHeight - reservedLines
		if availableHeight < 3 {
			availableHeight = 3
		}
		visibleEntries := availableHeight / 3
		// Only scroll if there's more content below
		if m.audit.Scroll+visibleEntries < len(m.audit.Logs) {
			m.audit.Scroll++
		}
		return m, nil
	}
	// In bulk delete menu: navigate down
	if m.currentView == ViewBulkDeleteMenu && m.bulkDelete.MenuCursor < 1 {
		m.bulkDelete.MenuCursor++
		return m, nil
	}
	// In add/edit form: navigate to next field
	if m.currentView == ViewAdd || m.currentView == ViewEdit {
		// Smart navigation: skip Caddy fields if DNS-only mode
		if m.addForm.DNSOnly {
			// DNS-only mode: cycle through 0, 1, 2, 3, 6
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
			// Full mode: cycle through all fields following visual order
			// Visual order: 0(Subdomain), 1(DNSType), 2(DNSTarget), 3(DNSOnly), 6(Proxied), 4(ReverseProxy), 5(Port), 7(SSL), snippets(8+), custom(8+len)
			customConfigFieldIndex := 8 + len(m.snippets)
			switch m.addForm.FocusedField {
			case 0:
				m.addForm.FocusedField = 1
			case 1:
				m.addForm.FocusedField = 2
			case 2:
				m.addForm.FocusedField = 3
			case 3:
				m.addForm.FocusedField = 6 // Jump to Proxied
			case 6:
				m.addForm.FocusedField = 4 // Back to Reverse Proxy Target
			case 4:
				m.addForm.FocusedField = 5
			case 5:
				m.addForm.FocusedField = 7 // Jump to SSL (skip 6)
			case 7:
				if len(m.snippets) > 0 {
					m.addForm.FocusedField = 8 // First snippet
				} else {
					m.addForm.FocusedField = customConfigFieldIndex // Custom config
				}
			default:
				// Snippets and custom config (8+)
				if m.addForm.FocusedField >= 8 && m.addForm.FocusedField < customConfigFieldIndex {
					// In snippets, go to next snippet or custom config
					m.addForm.FocusedField++
				} else if m.addForm.FocusedField == customConfigFieldIndex {
					// At custom config, wrap to start
					m.addForm.FocusedField = 0
				} else {
					m.addForm.FocusedField = 0
				}
			}
		}
		return m, nil
	}
	// In list view: navigate down (entries or snippets depending on focus)
	if m.currentView == ViewList && !m.searching {
		if m.panelFocus == PanelFocusSnippets {
			// Navigate snippets
			if m.snippetPanel.Cursor < len(m.snippets)-1 {
				m.snippetPanel.Cursor++
				// Auto-scroll snippet panel
				layout := NewPanelLayout(m.width, m.height)
				availableHeight := layout.SnippetsHeight - 5
				if availableHeight < 4 {
					availableHeight = 4
				}
				visibleSnippets := availableHeight / 4 // 4 lines per snippet
				if visibleSnippets < 1 {
					visibleSnippets = 1
				}
				if m.snippetPanel.Cursor >= m.snippetPanel.ScrollOffset+visibleSnippets {
					m.snippetPanel.ScrollOffset++
				}
			}
		} else if m.cursor < len(m.getFilteredEntries())-1 {
			// Navigate entries
			m.cursor++
			// Auto-scroll
			if m.cursor >= m.scrollOffset+m.height-5 {
				m.scrollOffset++
			}
		}
	}
	return m, nil
}

// handleNavigateUp dispatches 'up' arrow key per-view.
func (m Model) handleNavigateUp() (Model, tea.Cmd) {
	// In wizard: navigate up
	if m.currentView == ViewWizard {
		return m.handleWizardKeyPress("up")
	}
	// In migration wizard: navigate up
	if m.currentView == ViewMigrationWizard {
		return m.handleMigrationWizardKeyPress("up")
	}
	// In snippet wizard: navigate up
	if m.currentView == ViewSnippetWizard {
		switch m.snippetWizardStep {
		case SnippetWizardWelcome:
			// Navigate wizard mode selection (3 modes: Templated, Custom, Guided)
			if m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		case snippet_wizard.StepTemplateSelection:
			// Navigate template selection (15 templates)
			if m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		case snippet_wizard.StepCustomSnippet:
			// Navigate between name (0) and content (1) fields
			if m.wizardCursor > 0 {
				m.wizardCursor--
				// Update focus on components
				m.snippetWizardData.CustomNameInput.Focus()
				m.snippetWizardData.CustomContentInput.Blur()
			}
			return m, nil
		case SnippetWizardAutoDetect:
			// Navigate detected patterns
			if m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		case SnippetWizardIPRestriction:
			// Navigate between LAN subnet (0) and external IP (1) fields
			if m.snippetWizardData.CreateIPRestriction && m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		case SnippetWizardSecurityHeaders:
			// Navigate security presets (3 options: basic, strict, paranoid)
			if m.wizardCursor > 0 {
				m.wizardCursor--
				// Update selected preset based on cursor
				presets := []string{"basic", "strict", "paranoid"}
				m.snippetWizardData.SecurityPreset = presets[m.wizardCursor]
			}
			return m, nil
		case snippet_wizard.StepTemplateParams:
			// Navigate template parameter fields
			if m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		}
	}
	// In profile selector: navigate up
	if m.currentView == ViewProfileSelector {
		return m.handleProfileSelectorKeyPress("up")
	}
	// In profile edit: navigate up
	if m.currentView == ViewProfileEdit {
		return m.handleProfileEditKeyPress("up")
	}
	// In backup manager: navigate up
	if m.currentView == ViewBackupManager && !m.loading && m.backup.Cursor > 0 {
		m.backup.Cursor--
		// Adjust scroll if needed
		if m.backup.Cursor < m.backup.ScrollOffset {
			m.backup.ScrollOffset--
		}
		return m, nil
	}
	// In delete scope selection: navigate up
	if m.currentView == ViewDeleteScope && !m.loading && m.delete.ScopeCursor > 0 {
		m.delete.ScopeCursor--
		// Update selected scope based on cursor
		m.delete.Scope = DeleteScope(m.delete.ScopeCursor)
		return m, nil
	}
	// In restore scope selection: navigate up
	if m.currentView == ViewRestoreScope && !m.loading && m.backup.RestoreScopeCursor > 0 {
		m.backup.RestoreScopeCursor--
		// Update selected scope based on cursor
		m.backup.RestoreScope = RestoreScope(m.backup.RestoreScopeCursor)
		return m, nil
	}
	// In backup preview: scroll up
	if m.currentView == ViewBackupPreview && !m.loading && m.backup.PreviewScroll > 0 {
		m.backup.PreviewScroll--
		return m, nil
	}
	// In audit log: scroll up
	if m.currentView == ViewAuditLog && !m.loading && m.audit.Scroll > 0 {
		m.audit.Scroll--
		return m, nil
	}
	// In bulk delete menu: navigate up
	if m.currentView == ViewBulkDeleteMenu && m.bulkDelete.MenuCursor > 0 {
		m.bulkDelete.MenuCursor--
		return m, nil
	}
	// In add/edit form: navigate to previous field
	if m.currentView == ViewAdd || m.currentView == ViewEdit {
		// Smart navigation: skip Caddy fields if DNS-only mode
		if m.addForm.DNSOnly {
			// DNS-only mode: cycle through 6, 3, 2, 1, 0
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
				m.addForm.FocusedField = 0
			}
		} else {
			// Full mode: cycle through all fields following visual order (reverse)
			// Visual order reverse: custom(8+len), snippets(8+), 7(SSL), 5(Port), 4(ReverseProxy), 6(Proxied), 3(DNSOnly), 2(DNSTarget), 1(DNSType), 0(Subdomain)
			customConfigFieldIndex := 8 + len(m.snippets)
			switch m.addForm.FocusedField {
			case 0:
				m.addForm.FocusedField = customConfigFieldIndex // Wrap to custom config
			case 1:
				m.addForm.FocusedField = 0
			case 2:
				m.addForm.FocusedField = 1
			case 3:
				m.addForm.FocusedField = 2
			case 6:
				m.addForm.FocusedField = 3 // Back to DNS Only
			case 4:
				m.addForm.FocusedField = 6 // Back to Proxied
			case 5:
				m.addForm.FocusedField = 4
			case 7:
				m.addForm.FocusedField = 5 // Back to Port
			case 8:
				m.addForm.FocusedField = 7 // Back to SSL
			default:
				// Snippets and custom config (8+)
				if m.addForm.FocusedField > 8 && m.addForm.FocusedField <= customConfigFieldIndex {
					// Go to previous snippet or SSL
					m.addForm.FocusedField--
				} else {
					m.addForm.FocusedField = 0
				}
			}
		}
		return m, nil
	}
	// In list view: navigate up (entries or snippets depending on focus)
	if m.currentView == ViewList && !m.searching {
		if m.panelFocus == PanelFocusSnippets {
			// Navigate snippets
			if m.snippetPanel.Cursor > 0 {
				m.snippetPanel.Cursor--
				// Auto-scroll snippet panel
				if m.snippetPanel.Cursor < m.snippetPanel.ScrollOffset {
					m.snippetPanel.ScrollOffset--
				}
			}
		} else if m.cursor > 0 {
			// Navigate entries
			m.cursor--
			// Auto-scroll
			if m.cursor < m.scrollOffset {
				m.scrollOffset--
			}
		}
	}
	return m, nil
}
