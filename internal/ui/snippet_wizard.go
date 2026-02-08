package ui

import (
	"fmt"

	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
	snippet_wizard_views "lazyproxyflare/internal/ui/snippet_wizard/views"
)

// Type aliases for backward compatibility
type SnippetWizardStep = snippet_wizard.WizardStep
type SnippetWizardData = snippet_wizard.SnippetWizardData
type GeneratedSnippet = snippet_wizard.GeneratedSnippet

// Constants for backward compatibility
const (
	SnippetWizardWelcome           = snippet_wizard.StepWelcome
	SnippetWizardAutoDetect        = snippet_wizard.StepAutoDetect
	SnippetWizardIPRestriction     = snippet_wizard.StepIPRestriction
	SnippetWizardSecurityHeaders   = snippet_wizard.StepSecurityHeaders
	SnippetWizardPerformance       = snippet_wizard.StepPerformance
	SnippetWizardSummary           = snippet_wizard.StepSummary
)

// renderSnippetWizardWelcome renders the welcome screen for snippet wizard
func (m Model) renderSnippetWizardWelcome() string {
	return snippet_wizard.RenderWelcome(m.wizardCursor)
}

// renderSnippetWizardAutoDetect renders the auto-detection results screen
func (m Model) renderSnippetWizardAutoDetect() string {
	return snippet_wizard.RenderAutoDetect(
		m.snippetWizardData.DetectedPatterns,
		m.snippetWizardData.SelectedPatterns,
		m.wizardCursor,
	)
}

// renderSnippetWizardIPRestriction renders the IP restriction setup screen
func (m Model) renderSnippetWizardIPRestriction() string {
	return snippet_wizard.RenderIPRestriction(m.snippetWizardData, m.wizardCursor)
}

// renderSnippetWizardSecurityHeaders renders the security headers setup screen
func (m Model) renderSnippetWizardSecurityHeaders() string {
	return snippet_wizard.RenderSecurityHeaders(m.snippetWizardData, m.wizardCursor)
}

// renderSnippetWizardPerformance renders the performance optimization setup screen
func (m Model) renderSnippetWizardPerformance() string {
	return snippet_wizard.RenderPerformance(m.snippetWizardData)
}

// renderSnippetWizardSummary renders the summary screen showing all snippets to be created
func (m Model) renderSnippetWizardSummary() string {
	return snippet_wizard.RenderSummary(m.snippetWizardData)
}

// Backward compatibility - delegate to new package
func generateIPRestrictionSnippet(lanSubnet, externalIP string) string {
	return snippet_wizard.GenerateIPRestrictionSnippet(lanSubnet, externalIP)
}

func generateSecurityHeadersSnippet(preset string) string {
	return snippet_wizard.GenerateSecurityHeadersSnippet(preset)
}

func generatePerformanceSnippet() string {
	return snippet_wizard.GeneratePerformanceSnippet()
}

// renderSnippetWizardTemplateSelection renders the template selection screen
func (m Model) renderSnippetWizardTemplateSelection() string {
	// Build set of existing snippet names for quick lookup
	existingSnippets := make(map[string]bool)
	for _, snippet := range m.snippets {
		existingSnippets[snippet.Name] = true
	}

	// Build template info map
	availableTemplates := snippet_wizard.GetAvailableTemplates()
	templateInfoMap := make(map[string]snippet_wizard_views.TemplateInfo)
	for key, tmpl := range availableTemplates {
		templateInfoMap[key] = snippet_wizard_views.TemplateInfo{
			Name:        tmpl.Name,
			Description:tmpl.Description,
			Category:    tmpl.Category,
			Selected:    m.snippetWizardData.SelectedTemplates[key],
			InUse:       existingSnippets[key],
		}
	}
	// Calculate max content height: terminal height - title bar - borders - footer
	maxHeight := m.height - 8
	if maxHeight < 10 {
		maxHeight = 10
	}
	return snippet_wizard_views.RenderTemplateSelection(templateInfoMap, m.wizardCursor, maxHeight)
}

// renderSnippetWizardCustom renders the custom snippet paste screen
func (m Model) renderSnippetWizardCustom() string {
	return snippet_wizard_views.RenderCustomSnippet(
		m.snippetWizardData.CustomNameInput,
		m.snippetWizardData.CustomContentInput,
		m.wizardCursor,
	)
}

// renderSnippetWizardTemplateParams renders the template parameter configuration screen
func (m Model) renderSnippetWizardTemplateParams() string {
	// Find the first selected template to configure
	// Template order matches the order used in Enter key handling (app.go:2706-2708)
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
		return "No template selected"
	}

	// Get parameters for this template
	params := make(map[string]string)
	if config, exists := m.snippetWizardData.SnippetConfigs[currentTemplate]; exists {
		for k, v := range config.Parameters {
			params[k] = fmt.Sprint(v)
		}
	}

	// Render using the views package
	return snippet_wizard_views.RenderTemplateParams(currentTemplate, params, m.wizardCursor)
}
