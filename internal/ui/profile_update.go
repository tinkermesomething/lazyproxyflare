package ui

import (
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"lazyproxyflare/internal/config"
)

// handleProfileSelectorKeyPress handles key presses in profile selector view
func (m Model) handleProfileSelectorKeyPress(key string) (Model, tea.Cmd) {
	switch key {
	case "esc":
		// Return to main view if profile loaded, otherwise quit
		if m.config == nil {
			m.quitting = true
			return m, tea.Quit
		}
		m.currentView = ViewList
		return m, nil

	case "up", "k":
		// Navigate up
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		// Navigate down
		maxCursor := len(m.profile.Available) // +1 for "add new" option
		if m.cursor < maxCursor {
			m.cursor++
		}
		return m, nil

	case "enter":
		// Select profile or launch wizard
		if m.cursor < len(m.profile.Available) {
			// Load selected profile
			profileName := m.profile.Available[m.cursor]
			return m.loadProfile(profileName)
		} else {
			// Launch wizard to create new profile
			return m.startWizard(), nil
		}

	case "+", "n":
		// Launch wizard to create new profile
		return m.startWizard(), nil

	case "e":
		// Edit selected profile (not "add new" option)
		if m.cursor < len(m.profile.Available) {
			profileName := m.profile.Available[m.cursor]
			return m.startProfileEdit(profileName)
		}
		return m, nil

	case "x":
		// Export selected profile
		if m.cursor < len(m.profile.Available) {
			profileName := m.profile.Available[m.cursor]
			exportDir, err := config.GetDefaultExportDir()
			if err != nil {
				m.err = err
				return m, nil
			}
			timestamp := time.Now().Format("20060102_150405")
			exportPath := fmt.Sprintf("%s/%s_%s.tar.gz", exportDir, profileName, timestamp)
			return m, exportProfileCmd(profileName, exportPath)
		}
		return m, nil

	case "i":
		// Import profile — switch to import view with path input
		m.profile.ImportPath = ""
		m.currentView = ViewConfirmImport
		m.err = nil
		return m, nil

	case "d":
		// Delete selected profile (not "add new" option)
		if m.cursor < len(m.profile.Available) {
			profileName := m.profile.Available[m.cursor]
			// Cannot delete the currently active profile
			if profileName == m.profile.CurrentName {
				m.err = fmt.Errorf("cannot delete the active profile")
				return m, nil
			}
			m.profile.DeleteProfileName = profileName
			m.currentView = ViewConfirmDeleteProfile
			m.err = nil
		}
		return m, nil

	case "p", "ctrl+p":
		// Close profile selector (same as ESC)
		if m.config == nil {
			m.quitting = true
			return m, tea.Quit
		}
		m.currentView = ViewList
		return m, nil
	}

	return m, nil
}

// loadProfile loads a profile and switches to it
func (m Model) loadProfile(profileName string) (Model, tea.Cmd) {
	// Load profile config
	profileConfig, err := config.LoadProfile(profileName)
	if err != nil {
		m.err = err
		m.currentView = ViewProfileSelector
		return m, nil
	}

	// Set as current profile
	m.profile.CurrentName = profileName

	// Save as last used
	config.SetLastUsedProfile(profileName)

	// Convert to legacy config format
	m.config = config.ProfileToLegacyConfig(profileConfig)

	// Clear current data (will be reloaded)
	m.entries = nil
	m.snippets = nil
	m.cursor = 0
	m.scrollOffset = 0

	// Reset filters/search
	m.searchQuery = ""
	m.searching = false
	m.statusFilter = FilterAll
	m.dnsTypeFilter = DNSTypeAll
	m.sortMode = SortAlphabetical

	// Clear selections
	m.selectedEntries = make(map[string]bool)

	// Return to main view
	m.currentView = ViewList

	// Set loading and return command to reload data
	m.loading = true
	return m, refreshDataCmd(m.config)
}





// startWizard initializes wizard state and switches to wizard view
func (m Model) startWizard() Model {
	// Reset wizard state
	m.wizardStep = WizardStepWelcome
	m.wizardData = WizardData{}
	m.wizardTextInput.SetValue("")
	m.wizardTextInput.Focus() // Focus text input so it can receive keypresses
	m.wizardCursor = 0

	// Switch to wizard view
	m.currentView = ViewWizard

	return m
}

// startProfileEdit loads a profile for editing
func (m Model) startProfileEdit(profileName string) (Model, tea.Cmd) {
	// Load the profile
	profileConfig, err := config.LoadProfile(profileName)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Populate edit form with current values
	m.profile.EditData = ProfileEditData{
		OriginalName:  profileName,
		Name:          profileConfig.Profile.Name,
		APIToken:      profileConfig.Cloudflare.APIToken,
		ZoneID:        profileConfig.Cloudflare.ZoneID,
		Domain:        profileConfig.Domain,
		CaddyfilePath: profileConfig.Proxy.Caddy.CaddyfilePath,
		ContainerPath: profileConfig.Proxy.Caddy.CaddyfileContainerPath,
		ContainerName: profileConfig.Proxy.Caddy.ContainerName,
		CNAMETarget:   profileConfig.Defaults.CNAMETarget,
		Port:          fmt.Sprintf("%d", profileConfig.Defaults.Port),
		SSL:           profileConfig.Defaults.SSL,
		Proxied:       profileConfig.Defaults.Proxied,
		Editor:        profileConfig.UI.Editor,
	}
	m.profile.EditCursor = 0
	m.currentView = ViewProfileEdit
	m.err = nil

	return m, nil
}

// handleProfileEditKeyPress handles key presses in profile edit view
func (m Model) handleProfileEditKeyPress(key string) (Model, tea.Cmd) {
	const numFields = 13 // Total editable fields

	switch key {
	case "esc", "ctrl+w":
		// Cancel editing, return to profile selector
		m.currentView = ViewProfileSelector
		m.err = nil
		return m, nil

	case "tab", "down", "j":
		// Move to next field
		m.profile.EditCursor = (m.profile.EditCursor + 1) % numFields
		return m, nil

	case "shift+tab", "up", "k":
		// Move to previous field
		m.profile.EditCursor = (m.profile.EditCursor - 1 + numFields) % numFields
		return m, nil

	case " ":
		// Toggle boolean fields
		switch m.profile.EditCursor {
		case 11: // SSL
			m.profile.EditData.SSL = !m.profile.EditData.SSL
		case 12: // Proxied
			m.profile.EditData.Proxied = !m.profile.EditData.Proxied
		}
		return m, nil

	case "enter", "ctrl+s":
		// Save profile
		return m.saveProfileEdit()

	case "backspace":
		// Delete last character from current text field
		m.profileEditDeleteChar()
		return m, nil
	}

	// Handle text input for text fields
	if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
		m.profileEditAppendChar(key)
	}

	return m, nil
}

// profileEditAppendChar appends a character to the current text field
func (m *Model) profileEditAppendChar(char string) {
	switch m.profile.EditCursor {
	case 0:
		m.profile.EditData.Name += char
	case 1:
		m.profile.EditData.APIToken += char
	case 2:
		m.profile.EditData.ZoneID += char
	case 3:
		m.profile.EditData.Domain += char
	case 4:
		m.profile.EditData.CaddyfilePath += char
	case 5:
		m.profile.EditData.ContainerPath += char
	case 6:
		m.profile.EditData.ContainerName += char
	case 7:
		m.profile.EditData.CNAMETarget += char
	case 8:
		m.profile.EditData.Port += char
	case 9:
		m.profile.EditData.Editor += char
	}
}

// profileEditDeleteChar deletes the last character from the current text field
func (m *Model) profileEditDeleteChar() {
	deleteLastChar := func(s string) string {
		if len(s) > 0 {
			return s[:len(s)-1]
		}
		return s
	}

	switch m.profile.EditCursor {
	case 0:
		m.profile.EditData.Name = deleteLastChar(m.profile.EditData.Name)
	case 1:
		m.profile.EditData.APIToken = deleteLastChar(m.profile.EditData.APIToken)
	case 2:
		m.profile.EditData.ZoneID = deleteLastChar(m.profile.EditData.ZoneID)
	case 3:
		m.profile.EditData.Domain = deleteLastChar(m.profile.EditData.Domain)
	case 4:
		m.profile.EditData.CaddyfilePath = deleteLastChar(m.profile.EditData.CaddyfilePath)
	case 5:
		m.profile.EditData.ContainerPath = deleteLastChar(m.profile.EditData.ContainerPath)
	case 6:
		m.profile.EditData.ContainerName = deleteLastChar(m.profile.EditData.ContainerName)
	case 7:
		m.profile.EditData.CNAMETarget = deleteLastChar(m.profile.EditData.CNAMETarget)
	case 8:
		m.profile.EditData.Port = deleteLastChar(m.profile.EditData.Port)
	case 9:
		m.profile.EditData.Editor = deleteLastChar(m.profile.EditData.Editor)
	}
}

// saveProfileEdit saves the edited profile
func (m Model) saveProfileEdit() (Model, tea.Cmd) {
	data := m.profile.EditData

	// Validate required fields
	if data.Name == "" {
		m.err = fmt.Errorf("profile name is required")
		return m, nil
	}
	if data.APIToken == "" {
		m.err = fmt.Errorf("API token is required")
		return m, nil
	}
	if data.ZoneID == "" {
		m.err = fmt.Errorf("Zone ID is required")
		return m, nil
	}
	if data.Domain == "" {
		m.err = fmt.Errorf("domain is required")
		return m, nil
	}

	// Parse port
	port := 80
	if data.Port != "" {
		var err error
		port, err = strconv.Atoi(data.Port)
		if err != nil {
			m.err = fmt.Errorf("invalid port: %v", err)
			return m, nil
		}
	}

	// Load existing profile to preserve fields we don't edit
	existingProfile, err := config.LoadProfile(data.OriginalName)
	if err != nil {
		m.err = fmt.Errorf("failed to load profile: %v", err)
		return m, nil
	}

	// Update fields
	existingProfile.Profile.Name = data.Name
	existingProfile.Cloudflare.APIToken = data.APIToken
	existingProfile.Cloudflare.ZoneID = data.ZoneID
	existingProfile.Domain = data.Domain
	existingProfile.Proxy.Caddy.CaddyfilePath = data.CaddyfilePath
	existingProfile.Proxy.Caddy.CaddyfileContainerPath = data.ContainerPath
	existingProfile.Proxy.Caddy.ContainerName = data.ContainerName
	existingProfile.Defaults.CNAMETarget = data.CNAMETarget
	existingProfile.Defaults.Port = port
	existingProfile.Defaults.SSL = data.SSL
	existingProfile.Defaults.Proxied = data.Proxied
	existingProfile.UI.Editor = data.Editor

	// Handle rename
	if data.Name != data.OriginalName {
		// Delete old profile file
		if err := config.DeleteProfile(data.OriginalName); err != nil {
			// Log but don't fail - old file might not exist
		}
	}

	// Save profile
	if err := config.SaveProfile(data.Name, existingProfile); err != nil {
		m.err = fmt.Errorf("failed to save profile: %v", err)
		return m, nil
	}

	// If editing current profile, reload it
	if data.OriginalName == m.profile.CurrentName {
		m.profile.CurrentName = data.Name
		m.config = config.ProfileToLegacyConfig(existingProfile)
		config.SetLastUsedProfile(data.Name)
	}

	// Refresh available profiles list
	profiles, err := config.ListProfiles()
	if err == nil {
		m.profile.Available = profiles
	}

	// Return to profile selector
	m.currentView = ViewProfileSelector
	m.err = nil

	return m, nil
}

// exportProfileCmd creates an async command to export a profile
func exportProfileCmd(profileName, outputPath string) tea.Cmd {
	return func() tea.Msg {
		err := config.ExportProfile(profileName, outputPath)
		if err != nil {
			return exportProfileMsg{success: false, err: err}
		}
		return exportProfileMsg{success: true, path: outputPath}
	}
}

// importProfileCmd creates an async command to import a profile
func importProfileCmd(archivePath string) tea.Cmd {
	return func() tea.Msg {
		name, err := config.ImportProfile(archivePath, false)
		if err != nil {
			return importProfileMsg{success: false, err: err}
		}
		return importProfileMsg{success: true, profileName: name}
	}
}
