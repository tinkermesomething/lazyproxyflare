package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D7FF")).
			MarginBottom(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			Reverse(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	orphanedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))

	syncedIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	orphanedIconStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00"))
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.EnableMouseAllMotion
}

// isValidIPAddress validates IPv4 address format
func isValidIPAddress(ip string) bool {
	if ip == "" {
		return false
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}

		// Check for leading zeros (except for "0" itself)
		if len(part) > 1 && part[0] == '0' {
			return false
		}

		// Parse as number
		num := 0
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return false
			}
			num = num*10 + int(ch-'0')
		}

		// Check range 0-255
		if num > 255 {
			return false
		}
	}

	return true
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.MouseMsg:
		return m.handleMouseMsg(msg)

	default:
		if m2, cmd, handled := m.handleAsyncMsg(msg); handled {
			return m2, cmd
		}
	}
	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Handle full-screen views that don't need the base layout
	switch m.currentView {
	case ViewWizard:
		return m.renderWizardView()
	case ViewProfileSelector:
		return m.renderProfileSelectorView()
	case ViewProfileEdit:
		return m.renderProfileEditView()
	case ViewAuditLog:
		return m.renderAuditLogView()
	}

	// Always render the main panel layout as base (for views that need it)
	base := m.renderPanelLayout()

	// Overlay modals on top of the main view
	switch m.currentView {
	case ViewHelp:
		return RenderModalOverlay(base, "Keyboard Shortcuts", m.renderHelpModalContent(), m.width, m.height)

	case ViewAdd:
		return RenderModalOverlay(base, "Add New Entry", m.renderAddFormContent(), m.width, m.height)

	case ViewEdit:
		return RenderModalOverlay(base, "Edit Entry", m.renderEditFormContent(), m.width, m.height)

	case ViewPreview:
		return RenderModalOverlay(base, "Preview Changes", m.renderPreviewContent(), m.width, m.height)

	case ViewConfirmDelete:
		return RenderModalOverlay(base, "Confirm Delete", m.renderConfirmDeleteContent(), m.width, m.height)

	case ViewConfirmSync:
		return RenderModalOverlay(base, "Confirm Sync", m.renderConfirmSyncContent(), m.width, m.height)

	case ViewBulkDeleteMenu:
		return RenderModalOverlay(base, "Bulk Delete", m.renderBulkDeleteMenuContent(), m.width, m.height)

	case ViewConfirmBulkDelete:
		return RenderModalOverlay(base, "Confirm Bulk Delete", m.renderConfirmBulkDeleteContent(), m.width, m.height)

	case ViewConfirmBatchDelete:
		return RenderModalOverlay(base, "Confirm Batch Delete", m.renderConfirmBatchDeleteContent(), m.width, m.height)

	case ViewConfirmBatchSync:
		return RenderModalOverlay(base, "Confirm Batch Sync", m.renderConfirmBatchSyncContent(), m.width, m.height)

	case ViewBackupManager:
		return RenderModalOverlay(base, "Backup Manager", m.renderBackupManagerContent(), m.width, m.height)

	case ViewBackupPreview:
		return RenderModalOverlay(base, "Backup Preview", m.renderBackupPreviewContent(), m.width, m.height)

	case ViewDeleteScope:
		return RenderModalOverlay(base, "Delete Scope", m.renderDeleteScopeContent(), m.width, m.height)

	case ViewRestoreScope:
		return RenderModalOverlay(base, "Restore Scope", m.renderRestoreScopeContent(), m.width, m.height)

	case ViewConfirmRestore:
		return RenderModalOverlay(base, "Confirm Restore", m.renderConfirmRestoreContent(), m.width, m.height)

	case ViewConfirmCleanup:
		return RenderModalOverlay(base, "Confirm Cleanup", m.renderConfirmCleanupContent(), m.width, m.height)

	case ViewSnippetDetail:
		return RenderModalOverlay(base, "Snippet Details", m.renderSnippetDetailView(), m.width, m.height)

	case ViewSnippetWizard:
		title := "Snippet Creation Wizard"
		var content string
		switch m.snippetWizardStep {
		case SnippetWizardWelcome:
			content = m.renderSnippetWizardWelcome()
		case SnippetWizardAutoDetect:
			content = m.renderSnippetWizardAutoDetect()
		case SnippetWizardIPRestriction:
			content = m.renderSnippetWizardIPRestriction()
		case SnippetWizardSecurityHeaders:
			content = m.renderSnippetWizardSecurityHeaders()
		case SnippetWizardPerformance:
			content = m.renderSnippetWizardPerformance()
		case SnippetWizardSummary:
			content = m.renderSnippetWizardSummary()
		case snippet_wizard.StepTemplateSelection:
			content = m.renderSnippetWizardTemplateSelection()
		case snippet_wizard.StepTemplateParams:
			content = m.renderSnippetWizardTemplateParams()
		case snippet_wizard.StepCustomSnippet:
			content = m.renderSnippetWizardCustom()
		default:
			content = "Wizard step not implemented"
		}
		return RenderModalOverlay(base, title, content, m.width, m.height)

	case ViewConfirmDeleteProfile:
		return RenderModalOverlay(base, "Delete Profile", m.renderConfirmDeleteProfileContent(), m.width, m.height)

	case ViewMigrationWizard:
		title := "Caddyfile Migration"
		content := m.renderMigrationWizard()
		return RenderModalOverlay(base, title, content, m.width, m.height)

	case ViewError:
		return RenderModalOverlay(base, "Error", m.renderErrorModalContent(), m.width, m.height)

	default:
		// Check if there's an error - show error modal overlay on top of current view
		if m.err != nil {
			return RenderModalOverlay(base, "Error", m.renderErrorModalContent(), m.width, m.height)
		}
		// Main view without modal
		return base
	}
}
