package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/bubbles/textinput"

	"lazyproxyflare/internal/audit"
	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
)

// NewModel creates a new Bubbletea model
func NewModel(entries []diff.SyncedEntry, snippets []caddy.Snippet, cfg *config.Config) Model {
	// Sort entries alphabetically by domain
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Domain < entries[j].Domain
	})

	// Initialize audit logger
	// Use ~/.config/lazyproxyflare as default config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")

	auditLogger, err := audit.NewLogger(configDir)
	if err != nil {
		// If we can't create logger, log error but continue (non-fatal)
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
	}

	return Model{
		entries:             entries,
		snippets:            snippets,
		config:              cfg,
		cursor:              0,
		currentView:         ViewList,
		selectedEntries:     make(map[string]bool),
		backup:              BackupState{RetentionDays: 30}, // Default: keep backups for 30 days
		panelFocus:          PanelFocusLeft,
		audit:               AuditState{Logger: auditLogger},
	}
}

// NewModelWithWizard creates a new model starting in wizard view (no profile, no data)
func NewModelWithWizard() Model {
	// Initialize audit logger
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")

	auditLogger, err := audit.NewLogger(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
	}

	// Initialize text input for wizard
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return Model{
		entries:             []diff.SyncedEntry{},
		snippets:            []caddy.Snippet{},
		config:              nil,
		cursor:              0,
		currentView:         ViewWizard,
		wizardStep:          WizardStepWelcome,
		wizardData:          WizardData{},
		wizardCursor:        0,
		wizardTextInput:     ti,
		profile:             ProfileState{Available: []string{}},
		selectedEntries:     make(map[string]bool),
		backup:              BackupState{RetentionDays: 30},
		panelFocus:          PanelFocusLeft,
		audit:               AuditState{Logger: auditLogger},
	}
}

// NewModelWithProfile creates a new model with a loaded profile and data
func NewModelWithProfile(entries []diff.SyncedEntry, snippets []caddy.Snippet, cfg *config.Config, profileName string, password string) Model {
	// Sort entries alphabetically by domain
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Domain < entries[j].Domain
	})

	// Initialize audit logger
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")

	auditLogger, err := audit.NewLogger(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
	}

	// Initialize text input for wizard
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return Model{
		entries:             entries,
		snippets:            snippets,
		config:              cfg,
		profile:             ProfileState{CurrentName: profileName},
		cursor:              0,
		currentView:         ViewList,
		selectedEntries:     make(map[string]bool),
		backup:              BackupState{RetentionDays: 30},
		panelFocus:          PanelFocusLeft,
		audit:               AuditState{Logger: auditLogger},
		wizardTextInput:     ti,
	}
}

// NewModelWithProfileSelector creates a new model starting in profile selector view
func NewModelWithProfileSelector(profiles []string, lastUsedProfile string) Model {
	// Initialize audit logger
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")

	auditLogger, err := audit.NewLogger(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
	}

	// Set cursor to last used profile if found
	cursor := 0
	for i, profile := range profiles {
		if profile == lastUsedProfile {
			cursor = i
			break
		}
	}

	// Initialize text input for wizard
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return Model{
		entries:             []diff.SyncedEntry{},
		snippets:            []caddy.Snippet{},
		config:              nil,
		profile:             ProfileState{CurrentName: lastUsedProfile, Available: profiles},
		cursor:              cursor,
		currentView:         ViewProfileSelector,
		selectedEntries:     make(map[string]bool),
		backup:              BackupState{RetentionDays: 30},
		panelFocus:          PanelFocusLeft,
		audit:               AuditState{Logger: auditLogger},
		wizardTextInput:     ti,
	}
}
