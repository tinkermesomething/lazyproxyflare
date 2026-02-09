package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"lazyproxyflare/internal/caddy"
)

// migrationCompleteMsg is sent when migration completes
type migrationCompleteMsg struct {
	backupPath string
	err        error
}

// checkForCaddyfileMigration checks if Caddyfile exists and needs migration
// Returns true if migration wizard should be shown
func (m *Model) checkForCaddyfileMigration() bool {
	if m.config == nil {
		return false
	}

	caddyfilePath := m.config.Caddy.CaddyfilePath
	if caddyfilePath == "" {
		return false
	}

	// Parse existing Caddyfile
	parsedContent, err := caddy.ParseCaddyfileForMigration(caddyfilePath)
	if err != nil {
		// File doesn't exist or can't be parsed - no migration needed
		return false
	}

	// Only show migration if file has meaningful content
	return parsedContent.HasContent
}

// startMigrationWizard initializes and starts the migration wizard
func (m *Model) startMigrationWizard() error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	caddyfilePath := m.config.Caddy.CaddyfilePath
	if err := m.initMigrationWizard(caddyfilePath); err != nil {
		return err
	}

	m.currentView = ViewMigrationWizard
	m.migration.Active = true

	return nil
}

// handleMigrationWizardKeyPress handles keyboard input for migration wizard
func (m Model) handleMigrationWizardKeyPress(key string) (Model, tea.Cmd) {
	if m.migration.Data == nil {
		return m, nil
	}

	switch m.migration.Data.Step {
	case MigrationStepOptions:
		return m.handleMigrationOptionsKeyPress(key)
	case MigrationStepConfirm:
		return m.handleMigrationConfirmKeyPress(key)
	case MigrationStepProgress:
		// No input during progress
		return m, nil
	case MigrationStepComplete:
		return m.handleMigrationCompleteKeyPress(key)
	}

	return m, nil
}

// handleMigrationOptionsKeyPress handles keys in option selection screen
func (m Model) handleMigrationOptionsKeyPress(key string) (Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.migration.Data.OptionCursor > 0 {
			m.migration.Data.OptionCursor--
		}
		return m, nil

	case "down", "j":
		if m.migration.Data.OptionCursor < 4 { // 5 options (0-4)
			m.migration.Data.OptionCursor++
		}
		return m, nil

	case "enter":
		// Update options based on selection and move to confirm
		m.updateMigrationOptions()
		m.migration.Data.Step = MigrationStepConfirm
		return m, nil

	case "esc":
		// Cancel migration wizard
		m.migration.Active = false
		m.migration.Data = nil
		m.currentView = ViewList
		return m, nil
	}

	return m, nil
}

// handleMigrationConfirmKeyPress handles keys in confirmation screen
func (m Model) handleMigrationConfirmKeyPress(key string) (Model, tea.Cmd) {
	switch key {
	case "enter":
		// Check if "keep existing" option is selected (OptionCursor == 0)
		if m.migration.Data.OptionCursor == 0 {
			// Skip migration - just go to complete step
			m.migration.Data.Step = MigrationStepComplete
			m.migration.Data.InProgress = false
			m.migration.Data.BackupPath = "" // No backup created
			m.migration.Data.Error = nil     // No error
			return m, nil
		}

		// Start migration for other options
		m.migration.Data.Step = MigrationStepProgress
		m.migration.Data.InProgress = true
		return m, performMigrationCmd(
			m.migration.Data.CaddyfilePath,
			m.migration.Data.Options,
		)

	case "b":
		// Go back to options
		m.migration.Data.Step = MigrationStepOptions
		return m, nil

	case "esc":
		// Cancel migration wizard
		m.migration.Active = false
		m.migration.Data = nil
		m.currentView = ViewList
		return m, nil
	}

	return m, nil
}

// handleMigrationCompleteKeyPress handles keys in completion screen
func (m Model) handleMigrationCompleteKeyPress(key string) (Model, tea.Cmd) {
	// Any key returns to main view
	m.migration.Active = false
	m.migration.Data = nil
	m.currentView = ViewList

	// Reload data to reflect migrated Caddyfile
	m.loading = true
	return m, refreshDataCmd(m.config)
}

// performMigrationCmd executes the migration asynchronously
func performMigrationCmd(caddyfilePath string, options caddy.MigrationOptions) tea.Cmd {
	return func() tea.Msg {
		backupPath, err := caddy.MigrateCaddyfile(caddyfilePath, options)
		return migrationCompleteMsg{
			backupPath: backupPath,
			err:        err,
		}
	}
}

// handleMigrationComplete processes the migration completion message
func (m Model) handleMigrationComplete(msg migrationCompleteMsg) (Model, tea.Cmd) {
	if m.migration.Data == nil {
		return m, nil
	}

	m.migration.Data.InProgress = false
	m.migration.Data.Step = MigrationStepComplete
	m.migration.Data.BackupPath = msg.backupPath
	m.migration.Data.Error = msg.err

	return m, nil
}
