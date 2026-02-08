package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
)

// configureTextInputForField configures the textinput component for the current field
func (m *Model) configureTextInputForField() {
	m.wizardTextInput.Focus()

	switch m.wizardStep {
	case WizardStepBasicInfo:
		switch m.wizardData.CurrentField {
		case FieldProfileName:
			m.wizardTextInput.Placeholder = "my-profile"
			m.wizardTextInput.CharLimit = 50
			m.wizardTextInput.SetValue(m.wizardData.ProfileName)
			m.wizardTextInput.Validate = validateProfileName
		case FieldDomain:
			m.wizardTextInput.Placeholder = "example.com"
			m.wizardTextInput.CharLimit = 100
			m.wizardTextInput.SetValue(m.wizardData.Domain)
			m.wizardTextInput.Validate = validateDomain
		}

	case WizardStepCloudflare:
		switch m.wizardData.CurrentField {
		case FieldAPIToken:
			m.wizardTextInput.Placeholder = "your_api_token_here"
			m.wizardTextInput.CharLimit = 256
			m.wizardTextInput.EchoMode = textinput.EchoNormal
			m.wizardTextInput.SetValue(m.wizardData.APIToken)
			m.wizardTextInput.Validate = validateAPIToken
		case FieldZoneID:
			m.wizardTextInput.Placeholder = "32_hex_characters"
			m.wizardTextInput.CharLimit = 32
			m.wizardTextInput.EchoMode = textinput.EchoNormal
			m.wizardTextInput.SetValue(m.wizardData.ZoneID)
			m.wizardTextInput.Validate = validateZoneID
		}

	case WizardStepDockerConfig:
		switch m.wizardData.CurrentField {
		case FieldDeploymentMethod:
			// Radio selection - no text input config needed
			m.wizardTextInput.Blur()
		case FieldComposeFilePath:
			m.wizardTextInput.Placeholder = "/path/to/docker-compose.yml"
			m.wizardTextInput.CharLimit = 256
			m.wizardTextInput.SetValue(m.wizardData.ComposeFilePath)
			m.wizardTextInput.Validate = validatePath
		case FieldContainerName:
			m.wizardTextInput.Placeholder = "caddy"
			m.wizardTextInput.CharLimit = 100
			m.wizardTextInput.SetValue(m.wizardData.ContainerName)
			m.wizardTextInput.Validate = validateContainerName
		case FieldCaddyfilePath:
			if m.wizardData.DeploymentMethod == config.DeploymentSystem {
				m.wizardTextInput.Placeholder = "/etc/caddy/Caddyfile"
			} else {
				m.wizardTextInput.Placeholder = "/path/to/Caddyfile"
			}
			m.wizardTextInput.CharLimit = 256
			m.wizardTextInput.SetValue(m.wizardData.CaddyfilePath)
			m.wizardTextInput.Validate = validatePath
		case FieldCaddyfileContainerPath:
			m.wizardTextInput.Placeholder = "/etc/caddy/Caddyfile"
			m.wizardTextInput.CharLimit = 256
			m.wizardTextInput.SetValue(m.wizardData.CaddyfileContainerPath)
			m.wizardTextInput.Validate = validatePath
		case FieldCaddyBinaryPath:
			m.wizardTextInput.Placeholder = "/usr/bin/caddy"
			m.wizardTextInput.CharLimit = 256
			m.wizardTextInput.SetValue(m.wizardData.CaddyBinaryPath)
			m.wizardTextInput.Validate = validatePath
		}
	}
}

// Validation functions
func validateProfileName(s string) error {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return fmt.Errorf("invalid character")
		}
	}
	return nil
}

func validateDomain(s string) error {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '.') {
			return fmt.Errorf("invalid character")
		}
	}
	return nil
}

func validateAPIToken(s string) error {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return fmt.Errorf("invalid character")
		}
	}
	return nil
}

func validateZoneID(s string) error {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9')) {
			return fmt.Errorf("invalid character")
		}
	}
	return nil
}

func validatePath(s string) error {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '/') {
			return fmt.Errorf("invalid character")
		}
	}
	return nil
}

func validateContainerName(s string) error {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("invalid character")
		}
	}
	return nil
}

// handleWizardKeyPress handles key presses in wizard view
func (m Model) handleWizardKeyPress(key string) (Model, tea.Cmd) {
	switch key {
	case "esc":
		// Go back to previous step, or cancel if at start
		if m.wizardStep == WizardStepWelcome {
			if len(m.profile.Available) > 0 {
				m.currentView = ViewProfileSelector
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		} else {
			m.wizardStep = GetPreviousStep(m.wizardStep, &m.wizardData)
			m.wizardData.CurrentField = 0
			m.wizardCursor = 0
			m.configureTextInputForField()
		}
		return m, nil

	case "enter":
		return m.handleWizardEnter()

	case "tab", "down":
		return m.handleWizardNextField()

	case "shift+tab", "up":
		return m.handleWizardPrevField()

	case "b":
		// 'b' for back from summary
		if m.wizardStep == WizardStepSummary {
			m.wizardStep = GetPreviousStep(m.wizardStep, &m.wizardData)
			m.wizardData.CurrentField = 0
			m.wizardCursor = 0
			m.configureTextInputForField()
		}
		return m, nil

	case "y":
		// 'y' to confirm on summary
		if m.wizardStep == WizardStepSummary {
			return m.handleWizardSummaryConfirm()
		}
		return m, nil

	case "n":
		// 'n' to cancel on summary
		if m.wizardStep == WizardStepSummary {
			if len(m.profile.Available) > 0 {
				m.currentView = ViewProfileSelector
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		}
		return m, nil
	}

	return m, nil
}

// handleWizardNextField moves to the next field within the current step
func (m Model) handleWizardNextField() (Model, tea.Cmd) {
	// Save current field value before moving
	m.saveCurrentFieldValue()

	switch m.wizardStep {
	case WizardStepWelcome:
		// Welcome screen has no fields - Tab acts like Enter to proceed
		return m.handleWizardEnter()

	case WizardStepBasicInfo:
		if m.wizardData.CurrentField < FieldDomain {
			m.wizardData.CurrentField++
			m.configureTextInputForField()
		}

	case WizardStepCloudflare:
		if m.wizardData.CurrentField < FieldZoneID {
			m.wizardData.CurrentField++
			m.configureTextInputForField()
		}

	case WizardStepDockerConfig:
		field := m.wizardData.CurrentField

		// Handle radio selection navigation within DeploymentMethod
		if field == FieldDeploymentMethod {
			options := DeploymentOptions()
			if m.wizardCursor < len(options)-1 {
				m.wizardCursor++
				return m, nil
			}
			// Move to next field based on deployment type
			m.moveToNextFieldAfterDeployment()
			return m, nil
		}

		// Handle container name selection navigation (Docker only)
		if field == FieldContainerName && len(m.wizardDockerContainers) > 0 {
			maxCursor := len(m.wizardDockerContainers) // +1 for manual entry, but index is len-1
			if m.wizardCursor < maxCursor {
				m.wizardCursor++
				return m, nil
			}
			// Move to next field
			m.wizardData.CurrentField = FieldCaddyfilePath
			m.wizardCursor = 0
			m.configureTextInputForField()
			return m, nil
		}

		// Regular field navigation
		nextField := m.getNextCaddyField(field)
		if nextField != field {
			m.wizardData.CurrentField = nextField
			if nextField == FieldContainerName && m.wizardData.DeploymentMethod == config.DeploymentDocker {
				// Detect containers when entering container field
				containers, _ := caddy.ListDockerContainers(true)
				m.wizardDockerContainers = containers
				m.wizardCursor = 0
			}
			m.configureTextInputForField()
		}
	}

	return m, nil
}

// handleWizardPrevField moves to the previous field within the current step
func (m Model) handleWizardPrevField() (Model, tea.Cmd) {
	// Save current field value before moving
	m.saveCurrentFieldValue()

	switch m.wizardStep {
	case WizardStepBasicInfo:
		if m.wizardData.CurrentField > FieldProfileName {
			m.wizardData.CurrentField--
			m.configureTextInputForField()
		}

	case WizardStepCloudflare:
		if m.wizardData.CurrentField > FieldAPIToken {
			m.wizardData.CurrentField--
			m.configureTextInputForField()
		}

	case WizardStepDockerConfig:
		field := m.wizardData.CurrentField

		// Handle radio selection navigation within DeploymentMethod
		if field == FieldDeploymentMethod {
			if m.wizardCursor > 0 {
				m.wizardCursor--
			}
			return m, nil
		}

		// Handle container name selection navigation (Docker only)
		if field == FieldContainerName && len(m.wizardDockerContainers) > 0 {
			if m.wizardCursor > 0 {
				m.wizardCursor--
				return m, nil
			}
			// Move to prev field
			if m.wizardData.DockerMethod == "compose" {
				m.wizardData.CurrentField = FieldComposeFilePath
			} else {
				m.wizardData.CurrentField = FieldDeploymentMethod
				m.wizardCursor = len(DeploymentOptions()) - 1
			}
			m.configureTextInputForField()
			return m, nil
		}

		// Regular field navigation
		prevField := m.getPrevCaddyField(field)
		if prevField != field {
			m.wizardData.CurrentField = prevField
			if prevField == FieldDeploymentMethod {
				m.wizardCursor = len(DeploymentOptions()) - 1
			} else if prevField == FieldContainerName && len(m.wizardDockerContainers) > 0 {
				m.wizardCursor = len(m.wizardDockerContainers)
			}
			m.configureTextInputForField()
		}
	}

	return m, nil
}

// getNextCaddyField returns the next field in caddy config step
func (m Model) getNextCaddyField(current WizardField) WizardField {
	if m.wizardData.DeploymentMethod == config.DeploymentSystem {
		// System deployment: DeploymentMethod -> CaddyfilePath -> CaddyBinaryPath
		switch current {
		case FieldDeploymentMethod:
			return FieldCaddyfilePath
		case FieldCaddyfilePath:
			return FieldCaddyBinaryPath
		case FieldCaddyBinaryPath:
			return FieldCaddyBinaryPath // Stay at last field
		}
	} else {
		// Docker deployment
		switch current {
		case FieldDeploymentMethod:
			if m.wizardData.DockerMethod == "compose" {
				return FieldComposeFilePath
			}
			return FieldContainerName
		case FieldComposeFilePath:
			return FieldContainerName
		case FieldContainerName:
			return FieldCaddyfilePath
		case FieldCaddyfilePath:
			return FieldCaddyfileContainerPath
		case FieldCaddyfileContainerPath:
			return FieldCaddyfileContainerPath // Stay at last field
		}
	}
	return current
}

// getPrevCaddyField returns the previous field in caddy config step
func (m Model) getPrevCaddyField(current WizardField) WizardField {
	if m.wizardData.DeploymentMethod == config.DeploymentSystem {
		// System deployment: CaddyBinaryPath -> CaddyfilePath -> DeploymentMethod
		switch current {
		case FieldDeploymentMethod:
			return FieldDeploymentMethod // Stay at first field
		case FieldCaddyfilePath:
			return FieldDeploymentMethod
		case FieldCaddyBinaryPath:
			return FieldCaddyfilePath
		}
	} else {
		// Docker deployment
		switch current {
		case FieldDeploymentMethod:
			return FieldDeploymentMethod // Stay at first field
		case FieldComposeFilePath:
			return FieldDeploymentMethod
		case FieldContainerName:
			if m.wizardData.DockerMethod == "compose" {
				return FieldComposeFilePath
			}
			return FieldDeploymentMethod
		case FieldCaddyfilePath:
			return FieldContainerName
		case FieldCaddyfileContainerPath:
			return FieldCaddyfilePath
		}
	}
	return current
}

// moveToNextFieldAfterDeployment moves to the appropriate field after deployment method selection
func (m *Model) moveToNextFieldAfterDeployment() {
	if m.wizardData.DeploymentMethod == config.DeploymentSystem {
		// System deployment: go to Caddyfile path, auto-detect caddy binary
		m.wizardData.CurrentField = FieldCaddyfilePath
		// Set default Caddyfile path for system
		if m.wizardData.CaddyfilePath == "" {
			m.wizardData.CaddyfilePath = "/etc/caddy/Caddyfile"
		}
		// Auto-detect caddy binary
		if m.wizardData.CaddyBinaryPath == "" {
			m.wizardData.CaddyBinaryPath = detectCaddyBinary()
		}
	} else if m.wizardData.DockerMethod == "compose" {
		m.wizardData.CurrentField = FieldComposeFilePath
	} else {
		m.wizardData.CurrentField = FieldContainerName
		// Detect containers when entering container field
		containers, _ := caddy.ListDockerContainers(true)
		m.wizardDockerContainers = containers
	}
	m.wizardCursor = 0
	m.configureTextInputForField()
}

// detectCaddyBinary attempts to find the caddy binary using which/where
func detectCaddyBinary() string {
	// Try common locations first
	commonPaths := []string{
		"/usr/bin/caddy",
		"/usr/local/bin/caddy",
		"/snap/bin/caddy",
		"/home/linuxbrew/.linuxbrew/bin/caddy",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Try using exec.LookPath equivalent
	pathEnv := os.Getenv("PATH")
	for _, dir := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		path := filepath.Join(dir, "caddy")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "" // Not found, user will need to enter manually
}

// saveCurrentFieldValue saves the current text input value to the appropriate field
func (m *Model) saveCurrentFieldValue() {
	value := m.wizardTextInput.Value()

	switch m.wizardStep {
	case WizardStepBasicInfo:
		switch m.wizardData.CurrentField {
		case FieldProfileName:
			m.wizardData.ProfileName = value
		case FieldDomain:
			m.wizardData.Domain = value
		}

	case WizardStepCloudflare:
		switch m.wizardData.CurrentField {
		case FieldAPIToken:
			m.wizardData.APIToken = value
		case FieldZoneID:
			m.wizardData.ZoneID = value
		}

	case WizardStepDockerConfig:
		switch m.wizardData.CurrentField {
		case FieldComposeFilePath:
			m.wizardData.ComposeFilePath = value
		case FieldContainerName:
			// Only save if in manual entry mode or no containers detected
			if len(m.wizardDockerContainers) == 0 || m.wizardCursor == len(m.wizardDockerContainers) {
				m.wizardData.ContainerName = value
			}
		case FieldCaddyfilePath:
			m.wizardData.CaddyfilePath = value
		case FieldCaddyfileContainerPath:
			m.wizardData.CaddyfileContainerPath = value
		case FieldCaddyBinaryPath:
			m.wizardData.CaddyBinaryPath = value
		}
	}
}

// handleWizardEnter handles Enter key in wizard
func (m Model) handleWizardEnter() (Model, tea.Cmd) {
	m.err = nil

	switch m.wizardStep {
	case WizardStepWelcome:
		m.wizardStep = GetNextStep(m.wizardStep, &m.wizardData)
		m.wizardData.CurrentField = 0
		// Set defaults for Caddy + Docker
		m.wizardData.ProxyType = config.ProxyTypeCaddy
		m.wizardData.DeploymentMethod = config.DeploymentDocker
		m.wizardData.DockerMethod = "compose" // Default to compose
		m.configureTextInputForField()
		return m, nil

	case WizardStepBasicInfo:
		m.saveCurrentFieldValue()
		// If not on last field, move to next field
		if m.wizardData.CurrentField < FieldDomain {
			m.wizardData.CurrentField++
			m.configureTextInputForField()
			return m, nil
		}
		// On last field - validate and proceed to next step
		if err := ValidateCurrentStep(m.wizardStep, &m.wizardData); err != nil {
			m.err = err
			return m, nil
		}
		m.wizardStep = GetNextStep(m.wizardStep, &m.wizardData)
		m.wizardData.CurrentField = 0
		m.configureTextInputForField()
		return m, nil

	case WizardStepCloudflare:
		m.saveCurrentFieldValue()
		// If not on last field, move to next field
		if m.wizardData.CurrentField < FieldZoneID {
			m.wizardData.CurrentField++
			m.configureTextInputForField()
			return m, nil
		}
		// On last field - validate and proceed to next step
		if err := ValidateCurrentStep(m.wizardStep, &m.wizardData); err != nil {
			m.err = err
			return m, nil
		}
		m.wizardStep = GetNextStep(m.wizardStep, &m.wizardData)
		m.wizardData.CurrentField = 0
		m.wizardCursor = 0
		// Detect containers for container selection
		containers, _ := caddy.ListDockerContainers(true)
		m.wizardDockerContainers = containers
		m.configureTextInputForField()
		return m, nil

	case WizardStepDockerConfig:
		// Handle selection confirmation within fields
		field := m.wizardData.CurrentField

		// Deployment method selection
		if field == FieldDeploymentMethod {
			options := DeploymentOptions()
			if m.wizardCursor < len(options) {
				selectedValue := options[m.wizardCursor].Value
				if selectedValue == "system" {
					m.wizardData.DeploymentMethod = config.DeploymentSystem
				} else {
					m.wizardData.DeploymentMethod = config.DeploymentDocker
					m.wizardData.DockerMethod = selectedValue // "compose" or "plain"
				}
			}
			// Move to next field based on deployment type
			m.moveToNextFieldAfterDeployment()
			return m, nil
		}

		// Container name selection (Docker only)
		if field == FieldContainerName && len(m.wizardDockerContainers) > 0 {
			if m.wizardCursor < len(m.wizardDockerContainers) {
				// Selected a detected container
				m.wizardData.ContainerName = m.wizardDockerContainers[m.wizardCursor].Name
			} else {
				// Manual entry
				m.wizardData.ContainerName = m.wizardTextInput.Value()
			}
			// Move to next field
			m.wizardData.CurrentField = FieldCaddyfilePath
			m.wizardCursor = 0
			m.configureTextInputForField()
			return m, nil
		}

		// Save current field and validate
		m.saveCurrentFieldValue()
		if err := ValidateCurrentStep(m.wizardStep, &m.wizardData); err != nil {
			m.err = err
			return m, nil
		}

		// Set defaults for Docker deployment
		if m.wizardData.DeploymentMethod == config.DeploymentDocker {
			if m.wizardData.CaddyfileContainerPath == "" {
				m.wizardData.CaddyfileContainerPath = "/etc/caddy/Caddyfile"
			}
		}

		// Initialize defaults and go to summary
		m.initializeWizardDefaults()
		m.wizardStep = GetNextStep(m.wizardStep, &m.wizardData)
		return m, nil

	case WizardStepSummary:
		// Enter doesn't do anything on summary - use y/n
		return m, nil
	}

	return m, nil
}

// handleWizardSummaryConfirm handles 'y' in summary screen (save profile)
func (m Model) handleWizardSummaryConfirm() (Model, tea.Cmd) {
	// Validate profile before proceeding
	if err := ValidateProfile(&m.wizardData); err != nil {
		m.err = fmt.Errorf("validation failed: %w", err)
		return m, nil
	}

	// Convert wizard data to profile config
	profileConfig, err := m.wizardData.ToProfileConfig()
	if err != nil {
		m.err = fmt.Errorf("failed to create profile config: %w", err)
		return m, nil
	}

	// Validate Caddyfile path accessibility
	caddyfilePath := m.wizardData.CaddyfilePath
	if caddyfilePath != "" {
		parentDir := filepath.Dir(caddyfilePath)
		if info, err := os.Stat(parentDir); err != nil || !info.IsDir() {
			m.err = fmt.Errorf("Caddyfile parent directory does not exist: %s", parentDir)
			return m, nil
		}
	}

	// Save profile
	err = config.SaveProfile(m.wizardData.ProfileName, profileConfig)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Set as last used profile
	config.SetLastUsedProfile(m.wizardData.ProfileName)

	// Load the newly created profile
	m.profile.CurrentName = m.wizardData.ProfileName
	m.config = config.ProfileToLegacyConfig(profileConfig)

	// Switch to list view
	m.currentView = ViewList

	// Reset wizard state
	m.wizardStep = WizardStepWelcome
	m.wizardData = WizardData{}
	m.wizardTextInput.SetValue("")
	m.wizardCursor = 0

	// Check if Caddyfile exists and needs migration
	if m.checkForCaddyfileMigration() {
		if err := m.startMigrationWizard(); err != nil {
			m.err = fmt.Errorf("failed to start migration wizard: %w", err)
			m.loading = true
			return m, refreshDataCmd(m.config)
		}
		return m, nil
	}

	// Proceed to main view and reload data
	m.loading = true
	return m, refreshDataCmd(m.config)
}

// isWizardTextInputStep returns true if current step/field accepts text input
func (m Model) isWizardTextInputStep() bool {
	switch m.wizardStep {
	case WizardStepBasicInfo, WizardStepCloudflare:
		return true
	case WizardStepDockerConfig:
		field := m.wizardData.CurrentField
		// DeploymentMethod is radio selection, ContainerName may be selection
		if field == FieldDeploymentMethod {
			return false
		}
		if field == FieldContainerName && len(m.wizardDockerContainers) > 0 {
			// Only text input if "Enter manually" is selected
			return m.wizardCursor == len(m.wizardDockerContainers)
		}
		return true
	}
	return false
}

// initializeWizardDefaults sets default values for wizard defaults
func (m *Model) initializeWizardDefaults() {
	if m.wizardData.DefaultCNAMETarget == "" {
		if m.wizardData.Domain != "" {
			parts := strings.Split(m.wizardData.Domain, ".")
			if len(parts) >= 2 {
				m.wizardData.DefaultCNAMETarget = "mail." + m.wizardData.Domain
			} else {
				m.wizardData.DefaultCNAMETarget = m.wizardData.Domain
			}
		}
	}

	if m.wizardData.DefaultPort == 0 {
		m.wizardData.DefaultPort = 80
	}

	m.wizardData.DefaultProxied = true
	m.wizardData.DefaultSSL = false

	if m.wizardData.DefaultLANSubnet == "" {
		m.wizardData.DefaultLANSubnet = "10.0.0.0/8"
	}

	if m.wizardData.DefaultAllowedExternal == "" {
		m.wizardData.DefaultAllowedExternal = "127.0.0.1/32"
	}
}
