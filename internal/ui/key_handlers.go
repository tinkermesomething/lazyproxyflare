package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyMsg handles all keyboard input for the application.
// This method was extracted from Update() to improve maintainability.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Handle text input first — if consumed, skip the switch
	if m, cmd, handled := m.handleTextInput(msg); handled {
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "ctrl+q":
		// In snippet wizard: Ctrl+Q closes the modal
		if m.currentView == ViewSnippetWizard {
			m.currentView = ViewList
			m.snippetWizardData = SnippetWizardData{}
			m.err = nil
			return m, nil
		}
		// Otherwise quit app
		m.quitting = true
		return m, tea.Quit

	case "esc", "ctrl+w":
		return m.handleDismiss()

	case "tab":
		return m.handleTabKey()

	case "shift+tab":
		return m.handleShiftTabKey()

	case "/":
		return m.handleSearchStart()

	case "f":
		return m.handleStatusFilterCycle()

	case "t":
		return m.handleDNSTypeFilterCycle()

	case "o":
		return m.handleSortModeCycle()

	case "a":
		return m.handleAddEntry()

	case "e":
		return m.handleEditProfile()

	case "w", "ctrl+s":
		return m.handleOpenSnippetWizard()

	case "d":
		return m.handleDeleteAction()

	case "E":
		return m.handleOpenEditor()

	case "D":
		return m.handleOpenBulkDeleteMenu()

	case "b", "ctrl+b":
		return m.handleOpenBackupManager()

	case "l", "L":
		return m.handleOpenAuditLog()

	case "p", "ctrl+p":
		return m.handleProfileOrPreview()

	case "R":
		return m.handleRestoreBackup()

	case "right":
		return m.handleNavigateRight()

	case "left":
		return m.handleNavigateLeft()

	case "i":
		return m.handleImportProfile()

	case "x":
		return m.handleDeleteBackup()

	case "c":
		return m.handleCleanupBackups()

	case "X":
		return m.handleBatchDeleteSelected()

	case "S":
		return m.handleBatchSyncSelected()

	case "enter":
		return m.handleEnterKey()

	case "s":
		return m.handleSyncEntry()

	case "r":
		return m.handleRefreshData()

	case "m":
		return m.handleOpenMigrationWizard()

	case "?", "h", "ctrl+h":
		return m.handleOpenHelp()

	case "1", "2", "3", "4", "5":
		return m.handleHelpPageJump(msg.String())

	case "q":
		return m.handleQuit()

	case "y":
		return m.handleConfirmAction()

	case "+":
		return m.handleAddProfile()

	case "n":
		return m.handleCancelConfirmation()

	case " ":
		return m.handleSpaceKey()

	case "backspace":
		return m.handleBackspaceKey()

	case "j":
		return m.handleListDown()

	case "k":
		return m.handleListUp()

	case "down":
		return m.handleNavigateDown()

	case "up":
		return m.handleNavigateUp()

	case "g":
		return m.handleGoToTop()

	case "G":
		return m.handleGoToBottom()

	case "pgup":
		return m.handleBackupPageUp()

	case "pgdown":
		return m.handleBackupPageDown()

	case "home":
		return m.handleListHome()

	case "end":
		return m.handleListEnd()

	default:
		// No specific keybinding matched
	}

	return m, nil
}
