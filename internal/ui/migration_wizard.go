package ui

import (
	"fmt"
	"lazyproxyflare/internal/caddy"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MigrationWizardStep represents the current step in migration wizard
type MigrationWizardStep int

const (
	MigrationStepOptions MigrationWizardStep = iota // Option selection
	MigrationStepConfirm                            // Confirm migration
	MigrationStepProgress                           // Migration in progress
	MigrationStepComplete                           // Migration complete
)

// MigrationWizardData holds the state for migration wizard
type MigrationWizardData struct {
	Step            MigrationWizardStep
	ParsedContent   *caddy.MigrationContent
	Options         caddy.MigrationOptions
	OptionCursor    int    // Which option is selected (0-4)
	BackupPath      string // Path to created backup
	Error           error  // Any error during migration
	InProgress      bool   // Migration is running
	CaddyfilePath   string // Path to Caddyfile being migrated
}

// renderMigrationWizard renders the appropriate migration wizard screen
func (m Model) renderMigrationWizard() string {
	if m.migrationWizardData == nil {
		return "Error: Migration wizard data not initialized"
	}

	switch m.migrationWizardData.Step {
	case MigrationStepOptions:
		return m.renderMigrationOptions()
	case MigrationStepConfirm:
		return m.renderMigrationConfirm()
	case MigrationStepProgress:
		return m.renderMigrationProgress()
	case MigrationStepComplete:
		return m.renderMigrationComplete()
	default:
		return "Unknown migration step"
	}
}

// renderMigrationOptions renders the option selection screen
func (m Model) renderMigrationOptions() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	b.WriteString(titleStyle.Render("Existing Caddyfile Detected"))
	b.WriteString("\n\n")

	// Content summary
	if m.migrationWizardData.ParsedContent != nil {
		entriesCount, snippetsCount := m.migrationWizardData.ParsedContent.CountContent()

		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF87"))
		b.WriteString(infoStyle.Render("Found:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  • %d domain entries\n", entriesCount))
		b.WriteString(fmt.Sprintf("  • %d snippets\n", snippetsCount))
		if m.migrationWizardData.ParsedContent.HasGlobalOptions() {
			b.WriteString("  • Global options block\n")
		}
		b.WriteString("\n")
	}

	// Option selection
	b.WriteString("Choose migration strategy:\n\n")

	options := []struct {
		label       string
		description string
	}{
		{
			label:       "Keep existing Caddyfile unchanged",
			description: "Use current Caddyfile as-is - no changes, no import, no archive",
		},
		{
			label:       "Import all & regenerate managed file",
			description: "Keep all entries, snippets, and settings - generate clean managed Caddyfile",
		},
		{
			label:       "Import entries only (fresh snippets)",
			description: "Keep domain entries, discard old snippets - start with clean snippet library",
		},
		{
			label:       "Import snippets only (fresh entries)",
			description: "Keep snippets, discard entries - useful for preserving custom snippets",
		},
		{
			label:       "Archive & start completely fresh",
			description: "Archive everything and start with blank managed Caddyfile",
		},
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	for i, opt := range options {
		if i == m.migrationWizardData.OptionCursor {
			b.WriteString(selectedStyle.Render("→ [•] " + opt.label))
			b.WriteString("\n")
			b.WriteString("    " + opt.description)
		} else {
			b.WriteString(dimStyle.Render("  [ ] " + opt.label))
		}
		b.WriteString("\n")
		if i < len(options)-1 {
			b.WriteString("\n")
		}
	}

	// Archive notice (only show if not "keep existing")
	if m.migrationWizardData.OptionCursor != 0 {
		b.WriteString("\n")
		warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
		b.WriteString(warningStyle.Render("ℹ Original file will be archived to:"))
		b.WriteString("\n")
		backupName := fmt.Sprintf("  Caddyfile.backup.<timestamp>")
		b.WriteString(dimStyle.Render(backupName))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n")
		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF87"))
		b.WriteString(infoStyle.Render("ℹ No changes will be made to your Caddyfile"))
		b.WriteString("\n\n")
	}

	// Navigation
	b.WriteString(dimStyle.Render("↑/↓: navigate  Enter: proceed  ESC: cancel"))

	return b.String()
}

// renderMigrationConfirm renders the confirmation screen
func (m Model) renderMigrationConfirm() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	// Check if "keep existing" option is selected (OptionCursor == 0)
	if m.migrationWizardData.OptionCursor == 0 {
		// Keep existing - show different confirmation message
		b.WriteString(titleStyle.Render("Confirm: Keep Existing Caddyfile"))
		b.WriteString("\n\n")

		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF87"))
		b.WriteString(infoStyle.Render("✓ No changes will be made to your Caddyfile"))
		b.WriteString("\n\n")

		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
		entriesCount, snippetsCount := m.migrationWizardData.ParsedContent.CountContent()
		b.WriteString(dimStyle.Render("Your existing Caddyfile contains:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  • %d domain entries\n", entriesCount))
		b.WriteString(fmt.Sprintf("  • %d snippets\n", snippetsCount))
		if m.migrationWizardData.ParsedContent.HasGlobalOptions() {
			b.WriteString("  • Global options block\n")
		}
		b.WriteString("\n")

		// Navigation
		b.WriteString(dimStyle.Render("Enter: confirm  b: back  ESC: cancel"))
	} else {
		// Actual migration - show migration summary
		b.WriteString(titleStyle.Render("Confirm Migration"))
		b.WriteString("\n\n")

		// Show summary
		if m.migrationWizardData.ParsedContent != nil {
			summary := caddy.GetMigrationSummary(
				m.migrationWizardData.ParsedContent,
				m.migrationWizardData.Options,
			)
			b.WriteString(summary)
			b.WriteString("\n")
		}

		// Warning
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true)

		b.WriteString(warningStyle.Render("⚠ This will replace your current Caddyfile"))
		b.WriteString("\n")
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
		b.WriteString(dimStyle.Render("(Original will be backed up first)"))
		b.WriteString("\n\n")

		// Navigation
		b.WriteString(dimStyle.Render("Enter: confirm  b: back  ESC: cancel"))
	}

	return b.String()
}

// renderMigrationProgress renders the progress screen
func (m Model) renderMigrationProgress() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	b.WriteString(titleStyle.Render("Migration in Progress"))
	b.WriteString("\n\n")

	spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
	b.WriteString(spinnerStyle.Render("⏳ Migrating Caddyfile..."))
	b.WriteString("\n\n")

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	b.WriteString(dimStyle.Render("• Parsing existing Caddyfile"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("• Creating backup"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("• Generating managed Caddyfile"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("• Writing new file"))

	return b.String()
}

// renderMigrationComplete renders the completion screen
func (m Model) renderMigrationComplete() string {
	var b strings.Builder

	if m.migrationWizardData.Error != nil {
		// Error state
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true)

		b.WriteString(errorStyle.Render("✗ Migration Failed"))
		b.WriteString("\n\n")

		b.WriteString(fmt.Sprintf("Error: %v\n", m.migrationWizardData.Error))
		b.WriteString("\n")

		if m.migrationWizardData.BackupPath != "" {
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
			b.WriteString(dimStyle.Render("Original Caddyfile has been restored from backup"))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString("Press any key to return")
	} else {
		// Success state - check if this was "keep existing" or actual migration
		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF87")).
			Bold(true)

		// Check if "keep existing" option was selected (OptionCursor == 0)
		if m.migrationWizardData.OptionCursor == 0 {
			// Keep existing - no migration performed
			b.WriteString(successStyle.Render("✓ Caddyfile Unchanged"))
			b.WriteString("\n\n")

			infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF"))
			b.WriteString(infoStyle.Render("Using existing Caddyfile without changes"))
			b.WriteString("\n\n")

			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
			entriesCount, snippetsCount := m.migrationWizardData.ParsedContent.CountContent()
			b.WriteString(dimStyle.Render("Your Caddyfile contains:"))
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("  • %d domain entries\n", entriesCount))
			b.WriteString(fmt.Sprintf("  • %d snippets\n", snippetsCount))

			b.WriteString("\n")
			b.WriteString(dimStyle.Render("Press any key to continue"))
		} else {
			// Actual migration was performed
			b.WriteString(successStyle.Render("✓ Migration Complete"))
			b.WriteString("\n\n")

			infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D9FF"))
			b.WriteString(infoStyle.Render("Your Caddyfile has been migrated successfully!"))
			b.WriteString("\n\n")

			if m.migrationWizardData.BackupPath != "" {
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
				b.WriteString("Original backed up to:\n")
				b.WriteString(dimStyle.Render("  " + m.migrationWizardData.BackupPath))
				b.WriteString("\n\n")
			}

			entriesCount, snippetsCount := m.migrationWizardData.ParsedContent.CountContent()
			b.WriteString("Imported:\n")

			if m.migrationWizardData.Options.ImportEntries {
				b.WriteString(fmt.Sprintf("  • %d domain entries\n", entriesCount))
			}
			if m.migrationWizardData.Options.ImportSnippets {
				b.WriteString(fmt.Sprintf("  • %d snippets\n", snippetsCount))
			}

			b.WriteString("\n")
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
			b.WriteString(dimStyle.Render("Press any key to continue"))
		}
	}

	return b.String()
}

// initMigrationWizard initializes the migration wizard with parsed content
func (m *Model) initMigrationWizard(caddyfilePath string) error {
	// Parse existing Caddyfile
	parsedContent, err := caddy.ParseCaddyfileForMigration(caddyfilePath)
	if err != nil {
		return fmt.Errorf("failed to parse Caddyfile: %w", err)
	}

	// Initialize wizard data
	m.migrationWizardData = &MigrationWizardData{
		Step:          MigrationStepOptions,
		ParsedContent: parsedContent,
		Options: caddy.MigrationOptions{
			ImportEntries:  true, // Default: import all
			ImportSnippets: true,
			ArchiveOld:     true,
			FreshTemplate:  false,
		},
		OptionCursor:  0,
		CaddyfilePath: caddyfilePath,
	}

	return nil
}

// updateMigrationOptions updates migration options based on selected option
func (m *Model) updateMigrationOptions() {
	if m.migrationWizardData == nil {
		return
	}

	switch m.migrationWizardData.OptionCursor {
	case 0: // Keep existing - no migration
		m.migrationWizardData.Options = caddy.MigrationOptions{
			ImportEntries:  false,
			ImportSnippets: false,
			ArchiveOld:     false,
			FreshTemplate:  false,
		}
	case 1: // Import all
		m.migrationWizardData.Options = caddy.MigrationOptions{
			ImportEntries:  true,
			ImportSnippets: true,
			ArchiveOld:     true,
			FreshTemplate:  false,
		}
	case 2: // Import entries only
		m.migrationWizardData.Options = caddy.MigrationOptions{
			ImportEntries:  true,
			ImportSnippets: false,
			ArchiveOld:     true,
			FreshTemplate:  false,
		}
	case 3: // Import snippets only
		m.migrationWizardData.Options = caddy.MigrationOptions{
			ImportEntries:  false,
			ImportSnippets: true,
			ArchiveOld:     true,
			FreshTemplate:  false,
		}
	case 4: // Fresh start
		m.migrationWizardData.Options = caddy.MigrationOptions{
			ImportEntries:  false,
			ImportSnippets: false,
			ArchiveOld:     true,
			FreshTemplate:  true,
		}
	}
}
