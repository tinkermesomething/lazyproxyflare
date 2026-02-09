package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/audit"
	"lazyproxyflare/internal/caddy"
	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// createSnippetsFromWizard creates snippets based on wizard selections
func (m Model) createSnippetsFromWizard() (Model, tea.Cmd) {
	// Use wizard's centralized generation logic
	wizard := snippet_wizard.Wizard{
		State: &snippet_wizard.WizardState{
			Data: &m.snippetWizardData,
		},
	}
	generatedSnippets := wizard.GenerateSnippets()

	if len(generatedSnippets) == 0 {
		m.currentView = ViewList
		m.err = fmt.Errorf("no snippets selected")
		return m, nil
	}

	// Convert GeneratedSnippets to snippet strings
	var snippetsToAdd []string
	var snippetNames []string
	for _, gen := range generatedSnippets {
		// Wrap content in snippet syntax
		snippetStr := fmt.Sprintf("(%s) {\n%s\n}", gen.Name, gen.Content)
		snippetsToAdd = append(snippetsToAdd, snippetStr)
		snippetNames = append(snippetNames, gen.Name)
	}

	// Read current Caddyfile
	content, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.currentView = ViewList
		m.err = fmt.Errorf("failed to read Caddyfile: %w", err)
		return m, nil
	}

	// Check for duplicate snippets and filter them out
	parsed := caddy.ParseCaddyfileWithSnippets(string(content))
	existingNames := make(map[string]bool)
	for _, s := range parsed.Snippets {
		existingNames[s.Name] = true
	}

	// Filter out duplicates - only create new snippets
	var snippetsToCreate []string
	var newSnippetNames []string
	var skippedDuplicates []string

	for i, name := range snippetNames {
		if existingNames[name] {
			skippedDuplicates = append(skippedDuplicates, name)
		} else {
			snippetsToCreate = append(snippetsToCreate, snippetsToAdd[i])
			newSnippetNames = append(newSnippetNames, name)
		}
	}

	// If all snippets are duplicates, inform user
	if len(snippetsToCreate) == 0 {
		m.currentView = ViewList
		m.err = fmt.Errorf("all selected snippets already exist: %s", strings.Join(skippedDuplicates, ", "))
		return m, nil
	}

	// Prepend snippets to Caddyfile
	var newContent strings.Builder
	newContent.WriteString("# === Snippets created by LazyProxyFlare ===\n")
	newContent.WriteString("# Generated: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")

	for _, snippet := range snippetsToCreate {
		newContent.WriteString(snippet)
		newContent.WriteString("\n\n")
	}

	newContent.WriteString("# === End of snippets ===\n\n")
	newContent.WriteString(string(content))

	// Backup current Caddyfile
	backupPath, err := caddy.BackupCaddyfile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.currentView = ViewList
		m.err = fmt.Errorf("failed to backup Caddyfile: %w", err)
		return m, nil
	}

	// Get permissions from backup to preserve them
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		m.currentView = ViewList
		m.err = fmt.Errorf("failed to stat backup: %w", err)
		return m, nil
	}
	backupPerms := backupInfo.Mode().Perm()

	// Write new content with original permissions
	if err := os.WriteFile(m.config.Caddy.CaddyfilePath, []byte(newContent.String()), backupPerms); err != nil {
		m.currentView = ViewList
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: write failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("failed to write Caddyfile (backup restored): %w", err)
		}
		return m, nil
	}

	// Validate Caddyfile
	if err := caddy.FormatAndValidateCaddyfile(
		m.config.Caddy.CaddyfilePath,
		m.config.Caddy.CaddyfileContainerPath,
		m.config.Caddy.ContainerName,
		m.config.Caddy.DockerMethod,
		m.config.Caddy.ComposeFilePath,
		m.config.Caddy.ValidationCommand,
	); err != nil {
		m.currentView = ViewList
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: validation failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("Caddyfile validation failed (backup restored): %w", err)
		}
		return m, nil
	}

	// Reload Caddy
	if err := caddy.RestartCaddy(m.config.Caddy.ContainerName); err != nil {
		m.currentView = ViewList
		m.err = fmt.Errorf("failed to restart Caddy: %w", err)
		return m, nil
	}

	// Reload snippets
	caddyContent, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err == nil {
		parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))
		m.snippets = parsed.Snippets
	}

	// Success! Return to list view with message
	m.currentView = ViewList
	m.snippetWizardData = SnippetWizardData{} // Reset wizard data

	// Build success message
	if len(skippedDuplicates) > 0 {
		// Some duplicates were skipped
		m.err = fmt.Errorf("✓ Created %d snippet(s): %s | Skipped %d duplicate(s): %s",
			len(newSnippetNames), strings.Join(newSnippetNames, ", "),
			len(skippedDuplicates), strings.Join(skippedDuplicates, ", "))
	} else {
		// All created successfully
		m.err = nil
	}

	// Log the operation (only log newly created snippets)
	if m.audit.Logger != nil && len(newSnippetNames) > 0 {
		logEntry := audit.LogEntry{
			Timestamp:  time.Now(),
			Operation:  audit.OperationCreate,
			EntityType: audit.EntityCaddy,
			Domain:     strings.Join(newSnippetNames, ", "),
			Details: map[string]interface{}{
				"snippets": newSnippetNames,
				"method":   "wizard",
				"skipped":  skippedDuplicates,
			},
			Result: audit.ResultSuccess,
		}
		_ = m.audit.Logger.Log(logEntry)
	}

	return m, nil
}

// saveSnippetEdit saves the edited snippet content back to Caddyfile
func (m Model) saveSnippetEdit() (Model, tea.Cmd) {
	if m.snippetPanel.EditingIndex < 0 || m.snippetPanel.EditingIndex >= len(m.snippets) {
		m.err = fmt.Errorf("invalid snippet index")
		return m, nil
	}

	snippet := m.snippets[m.snippetPanel.EditingIndex]
	newContent := m.snippetPanel.EditTextarea.Value()

	// Read current Caddyfile
	content, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.err = fmt.Errorf("failed to read Caddyfile: %w", err)
		return m, nil
	}

	// Replace the snippet content in the Caddyfile
	// We need to find the snippet block and replace it
	lines := strings.Split(string(content), "\n")

	// Build new Caddyfile content with updated snippet
	var newCaddyfile strings.Builder
	for i := 0; i < len(lines); i++ {
		// Check if this is the start of our snippet (LineStart is 1-indexed)
		if i+1 == snippet.LineStart {
			// Write the snippet name line
			newCaddyfile.WriteString(lines[i])
			newCaddyfile.WriteString("\n")

			// Write the new content (already includes proper indentation from textarea)
			newCaddyfile.WriteString(newContent)
			newCaddyfile.WriteString("\n")

			// Skip the old content lines (until LineEnd-1, as LineEnd is the closing brace)
			// Then write the closing brace
			i = snippet.LineEnd - 1 // -1 because LineEnd is 1-indexed and we're 0-indexed
			if i < len(lines) {
				newCaddyfile.WriteString(lines[i])
				newCaddyfile.WriteString("\n")
			}
		} else {
			newCaddyfile.WriteString(lines[i])
			newCaddyfile.WriteString("\n")
		}
	}

	// Backup current Caddyfile
	backupPath, err := caddy.BackupCaddyfile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.err = fmt.Errorf("failed to backup Caddyfile: %w", err)
		return m, nil
	}

	// Get permissions from backup to preserve them
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		m.err = fmt.Errorf("failed to stat backup: %w", err)
		return m, nil
	}
	backupPerms := backupInfo.Mode().Perm()

	// Write new content with original permissions
	if err := os.WriteFile(m.config.Caddy.CaddyfilePath, []byte(newCaddyfile.String()), backupPerms); err != nil {
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: write failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("failed to write Caddyfile (backup restored): %w", err)
		}
		return m, nil
	}

	// Validate Caddyfile
	if err := caddy.FormatAndValidateCaddyfile(
		m.config.Caddy.CaddyfilePath,
		m.config.Caddy.CaddyfileContainerPath,
		m.config.Caddy.ContainerName,
		m.config.Caddy.DockerMethod,
		m.config.Caddy.ComposeFilePath,
		m.config.Caddy.ValidationCommand,
	); err != nil {
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: validation failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("Caddyfile validation failed (backup restored): %w", err)
		}
		return m, nil
	}

	// Reload Caddy
	if err := caddy.RestartCaddy(m.config.Caddy.ContainerName); err != nil {
		m.err = fmt.Errorf("failed to restart Caddy: %w", err)
		return m, nil
	}

	// Reload snippets
	caddyContent, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err == nil {
		parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))
		m.snippets = parsed.Snippets
	}

	// Exit edit mode, stay in detail view
	m.snippetPanel.Editing = false
	m.err = nil

	// Log the operation
	if m.audit.Logger != nil {
		logEntry := audit.LogEntry{
			Timestamp:  time.Now(),
			Operation:  audit.OperationUpdate,
			EntityType: audit.EntityCaddy,
			Domain:     snippet.Name,
			Details: map[string]interface{}{
				"snippet": snippet.Name,
				"method":  "edit",
			},
			Result: audit.ResultSuccess,
		}
		_ = m.audit.Logger.Log(logEntry)
	}

	return m, nil
}

// deleteSnippet removes the selected snippet from the Caddyfile
func (m Model) deleteSnippet() (Model, tea.Cmd) {
	if m.snippetPanel.EditingIndex < 0 || m.snippetPanel.EditingIndex >= len(m.snippets) {
		m.err = fmt.Errorf("invalid snippet index")
		return m, nil
	}

	snippet := m.snippets[m.snippetPanel.EditingIndex]

	// Check if snippet is currently in use
	usage := m.calculateSnippetUsage()
	if usage[snippet.Name] > 0 {
		m.err = fmt.Errorf("cannot delete snippet '%s': currently used by %d entries", snippet.Name, usage[snippet.Name])
		return m, nil
	}

	// Read current Caddyfile
	content, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.err = fmt.Errorf("failed to read Caddyfile: %w", err)
		return m, nil
	}

	// Remove the snippet from the Caddyfile
	lines := strings.Split(string(content), "\n")

	// Build new Caddyfile content without this snippet
	var newCaddyfile strings.Builder
	skipUntilLine := -1

	for i := 0; i < len(lines); i++ {
		lineNum := i + 1 // Convert to 1-indexed

		// If we're at the snippet start, mark to skip until end
		if lineNum == snippet.LineStart {
			skipUntilLine = snippet.LineEnd
			continue
		}

		// Skip lines that are part of the snippet
		if skipUntilLine >= 0 && lineNum <= skipUntilLine {
			continue
		}

		// Write non-snippet lines
		newCaddyfile.WriteString(lines[i])
		newCaddyfile.WriteString("\n")
	}

	// Backup current Caddyfile
	backupPath, err := caddy.BackupCaddyfile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.err = fmt.Errorf("failed to backup Caddyfile: %w", err)
		return m, nil
	}

	// Get permissions from backup to preserve them
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		m.err = fmt.Errorf("failed to stat backup: %w", err)
		return m, nil
	}
	backupPerms := backupInfo.Mode().Perm()

	// Write new content with original permissions
	if err := os.WriteFile(m.config.Caddy.CaddyfilePath, []byte(newCaddyfile.String()), backupPerms); err != nil {
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: write failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("failed to write Caddyfile (backup restored): %w", err)
		}
		return m, nil
	}

	// Validate Caddyfile
	if err := caddy.FormatAndValidateCaddyfile(
		m.config.Caddy.CaddyfilePath,
		m.config.Caddy.CaddyfileContainerPath,
		m.config.Caddy.ContainerName,
		m.config.Caddy.DockerMethod,
		m.config.Caddy.ComposeFilePath,
		m.config.Caddy.ValidationCommand,
	); err != nil {
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: validation failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("Caddyfile validation failed (backup restored): %w", err)
		}
		return m, nil
	}

	// Reload Caddy
	if err := caddy.RestartCaddy(m.config.Caddy.ContainerName); err != nil {
		m.err = fmt.Errorf("failed to restart Caddy: %w", err)
		return m, nil
	}

	// Reload snippets
	caddyContent, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err == nil {
		parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))
		m.snippets = parsed.Snippets
	}

	// Return to list view after deletion
	m.currentView = ViewList
	m.snippetPanel.Editing = false
	m.err = nil

	// Log the operation
	if m.audit.Logger != nil {
		logEntry := audit.LogEntry{
			Timestamp:  time.Now(),
			Operation:  audit.OperationDelete,
			EntityType: audit.EntityCaddy,
			Domain:     snippet.Name,
			Details: map[string]interface{}{
				"snippet": snippet.Name,
				"method":  "delete",
			},
			Result: audit.ResultSuccess,
		}
		_ = m.audit.Logger.Log(logEntry)
	}

	return m, nil
}

// getTemplateParamsCursorMax returns the max cursor value for the current template being configured
func (m Model) getTemplateParamsCursorMax() int {
	// Find the first selected template to configure (matches render logic)
	templateKeys := []string{
		"cors_headers", "rate_limiting", "large_uploads", "extended_timeouts",
		"static_caching", "compression_advanced", "https_backend", "auth_headers",
		"websocket_advanced", "custom_headers_inject", "rewrite_rules", "frame_embedding",
		"ip_restricted", "security_headers", "performance",
	}

	var currentTemplate string
	for _, key := range templateKeys {
		if m.snippetWizardData.SelectedTemplates[key] {
			currentTemplate = key
			break
		}
	}

	if currentTemplate == "" {
		return 0
	}

	// Return max cursor for each template type
	switch currentTemplate {
	case "cors_headers":
		return 2 // 3 fields: origins, methods, credentials
	case "rate_limiting":
		return 1 // 2 fields: req/s, burst
	case "large_uploads":
		return 0 // 1 field: max size
	case "extended_timeouts":
		return 2 // 3 fields: read, write, dial
	case "static_caching":
		return 1 // 2 fields: max age, etag
	case "ip_restricted":
		return 1 // 2 fields: lan subnet, external IP
	case "security_headers":
		return 2 // 3 presets
	case "compression_advanced":
		return 3 // 4 fields: level, gzip, zstd, brotli
	case "https_backend":
		return 1 // 2 fields: skip verify, keepalive
	case "performance":
		return 1 // 2 fields: gzip, zstd
	case "auth_headers":
		return 1 // 2 fields: forward IP, forward proto
	case "websocket_advanced":
		return 1 // 2 fields: upgrade timeout, ping interval
	case "custom_headers_inject":
		return 2 // 3 options: upstream, response, both
	case "rewrite_rules":
		return 1 // 2 fields: path pattern, rewrite to
	case "frame_embedding":
		return 0 // 1 field: allowed origins
	default:
		return 0
	}
}

// toggleSnippetTemplateParamCheckbox toggles checkbox parameters in the current template
func (m *Model) toggleSnippetTemplateParamCheckbox() {
	// Find the first selected template to configure
	templateKeys := []string{
		"cors_headers", "rate_limiting", "large_uploads", "extended_timeouts",
		"static_caching", "compression_advanced", "https_backend", "auth_headers",
		"websocket_advanced", "custom_headers_inject", "rewrite_rules", "frame_embedding",
		"ip_restricted", "security_headers", "performance",
	}

	var currentTemplate string
	for _, key := range templateKeys {
		if m.snippetWizardData.SelectedTemplates[key] {
			currentTemplate = key
			break
		}
	}

	if currentTemplate == "" {
		return
	}

	config := m.snippetWizardData.SnippetConfigs[currentTemplate]
	if config.Parameters == nil {
		config.Parameters = make(map[string]interface{})
	}

	// Toggle checkbox parameters based on template and cursor position
	switch currentTemplate {
	case "cors_headers":
		if m.wizardCursor == 2 {
			current := config.Parameters["allow_credentials"]
			if current == "true" {
				config.Parameters["allow_credentials"] = "false"
			} else {
				config.Parameters["allow_credentials"] = "true"
			}
		}
	case "static_caching":
		if m.wizardCursor == 1 {
			current := config.Parameters["enable_etag"]
			if current == "true" {
				config.Parameters["enable_etag"] = "false"
			} else {
				config.Parameters["enable_etag"] = "true"
			}
		}
	case "compression_advanced":
		switch m.wizardCursor {
		case 1: // gzip
			current := config.Parameters["enable_gzip"]
			if current == "false" {
				config.Parameters["enable_gzip"] = "true"
			} else {
				config.Parameters["enable_gzip"] = "false"
			}
		case 2: // zstd
			current := config.Parameters["enable_zstd"]
			if current == "false" {
				config.Parameters["enable_zstd"] = "true"
			} else {
				config.Parameters["enable_zstd"] = "false"
			}
		case 3: // brotli
			current := config.Parameters["enable_brotli"]
			if current == "true" {
				config.Parameters["enable_brotli"] = "false"
			} else {
				config.Parameters["enable_brotli"] = "true"
			}
		}
	case "https_backend":
		if m.wizardCursor == 0 {
			current := config.Parameters["skip_verify"]
			if current == "true" {
				config.Parameters["skip_verify"] = "false"
			} else {
				config.Parameters["skip_verify"] = "true"
			}
		}
	case "performance":
		switch m.wizardCursor {
		case 0: // gzip
			current := config.Parameters["enable_gzip"]
			if current == "false" {
				config.Parameters["enable_gzip"] = "true"
			} else {
				config.Parameters["enable_gzip"] = "false"
			}
		case 1: // zstd
			current := config.Parameters["enable_zstd"]
			if current == "false" {
				config.Parameters["enable_zstd"] = "true"
			} else {
				config.Parameters["enable_zstd"] = "false"
			}
		}
	case "auth_headers":
		switch m.wizardCursor {
		case 0: // forward IP
			current := config.Parameters["forward_ip"]
			if current == "true" {
				config.Parameters["forward_ip"] = "false"
			} else {
				config.Parameters["forward_ip"] = "true"
			}
		case 1: // forward proto
			current := config.Parameters["forward_proto"]
			if current == "true" {
				config.Parameters["forward_proto"] = "false"
			} else {
				config.Parameters["forward_proto"] = "true"
			}
		}
	}

	m.snippetWizardData.SnippetConfigs[currentTemplate] = config
}
