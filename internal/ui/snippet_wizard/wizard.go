package snippet_wizard

import (
	"fmt"
	"strings"

	"lazyproxyflare/internal/ui/snippet_wizard/views"
)

// toViewData converts internal SnippetWizardData to views.SnippetWizardData
func toViewData(data SnippetWizardData) views.SnippetWizardData {
	return views.SnippetWizardData{
		CreateIPRestriction:   data.CreateIPRestriction,
		LANSubnet:             data.LANSubnet,
		AllowedExternalIP:     data.AllowedExternalIP,
		CreateSecurityHeaders: data.CreateSecurityHeaders,
		SecurityPreset:        data.SecurityPreset,
		CreatePerformance:     data.CreatePerformance,
	}
}

// Package-level render functions for backward compatibility
func RenderWelcome(selectedMode int) string {
	return views.RenderWelcome(selectedMode)
}

func RenderIPRestriction(data SnippetWizardData, cursor int) string {
	// Validate CIDR fields
	var lanValidationMsg, externalValidationMsg string
	if data.CreateIPRestriction && data.LANSubnet != "" {
		lanResult := ValidateCIDR(data.LANSubnet)
		if !lanResult.Valid {
			lanValidationMsg = lanResult.Message
		}
	}
	if data.CreateIPRestriction && data.AllowedExternalIP != "" {
		externalResult := ValidateCIDR(data.AllowedExternalIP)
		if !externalResult.Valid {
			externalValidationMsg = externalResult.Message
		}
	}

	// Convert our data to views.SnippetWizardData
	viewData := toViewData(data)
	preview := GenerateIPRestrictionSnippet(data.LANSubnet, data.AllowedExternalIP)
	return views.RenderIPRestriction(viewData, cursor, preview, lanValidationMsg, externalValidationMsg)
}

func RenderSecurityHeaders(data SnippetWizardData, cursor int) string {
	viewData := toViewData(data)
	preview := GenerateSecurityHeadersSnippet(data.SecurityPreset)
	return views.RenderSecurityHeaders(viewData, cursor, preview)
}

func RenderPerformance(data SnippetWizardData) string {
	viewData := toViewData(data)
	preview := GeneratePerformanceSnippet()
	return views.RenderPerformance(viewData, preview)
}

func RenderSummary(data SnippetWizardData) string {
	// Build list of selected auto-detected snippets for summary
	var generatedSnippets []views.GeneratedSnippet
	for _, pattern := range data.DetectedPatterns {
		if data.SelectedPatterns[pattern.SuggestedName] {
			generatedSnippets = append(generatedSnippets, views.GeneratedSnippet{
				Name:        pattern.SuggestedName,
				Category:    pattern.Type.String(),
				Description: formatPatternDescription(pattern),
			})
		}
	}

	viewData := toViewData(data)
	viewData.GeneratedSnippets = generatedSnippets
	return views.RenderSummary(viewData)
}

// formatPatternDescription creates a human-readable description from detected pattern
func formatPatternDescription(pattern DetectedPattern) string {
	// Build parameter summary
	var params []string
	for key, value := range pattern.Parameters {
		params = append(params, key+": "+value)
	}
	if len(params) > 0 {
		return "Auto-detected (" + strings.Join(params, ", ") + ")"
	}
	return "Auto-detected from " + pattern.ExampleDomain
}

// RenderAutoDetect renders the auto-detection results view
func RenderAutoDetect(detectedPatterns []DetectedPattern, selectedPatterns map[string]bool, cursor int) string {
	// Convert to views.DetectedPattern
	viewPatterns := make([]views.DetectedPattern, len(detectedPatterns))
	for i, p := range detectedPatterns {
		viewPatterns[i] = views.DetectedPattern{
			Type:          int(p.Type),
			Count:         p.Count,
			ExampleDomain: p.ExampleDomain,
			RawDirective:  p.RawDirective,
			SuggestedName: p.SuggestedName,
			Parameters:    p.Parameters,
		}
	}
	return views.RenderAutoDetect(viewPatterns, selectedPatterns, cursor)
}

// Wizard is the main controller for the snippet wizard
type Wizard struct {
	State *WizardState
}

// NewWizard creates a new wizard instance
func NewWizard() *Wizard {
	return &Wizard{
		State: NewWizardState(),
	}
}

// Render renders the current step of the wizard
func (w *Wizard) Render() string {
	switch w.State.CurrentStep {
	case StepWelcome:
		return views.RenderWelcome(w.State.Cursor)
	case StepTemplateSelection:
		// Build template info map from catalog
		availableTemplates := GetAvailableTemplates()
		templateInfoMap := make(map[string]views.TemplateInfo)
		for key, tmpl := range availableTemplates {
			templateInfoMap[key] = views.TemplateInfo{
				Name:        tmpl.Name,
				Description: tmpl.Description,
				Category:    tmpl.Category,
				Selected:    w.State.Data.SelectedTemplates[key],
			}
		}
		// Use a reasonable default height for standalone wizard (not integrated with TUI)
		maxHeight := 30
		return views.RenderTemplateSelection(templateInfoMap, w.State.Cursor, maxHeight)
	case StepIPRestriction:
		// Validate CIDR fields
		var lanValidationMsg, externalValidationMsg string
		if w.State.Data.CreateIPRestriction && w.State.Data.LANSubnet != "" {
			lanResult := ValidateCIDR(w.State.Data.LANSubnet)
			if !lanResult.Valid {
				lanValidationMsg = lanResult.Message
			}
		}
		if w.State.Data.CreateIPRestriction && w.State.Data.AllowedExternalIP != "" {
			externalResult := ValidateCIDR(w.State.Data.AllowedExternalIP)
			if !externalResult.Valid {
				externalValidationMsg = externalResult.Message
			}
		}

		// Convert data and generate preview
		viewData := toViewData(*w.State.Data)
		preview := GenerateIPRestrictionSnippet(w.State.Data.LANSubnet, w.State.Data.AllowedExternalIP)
		return views.RenderIPRestriction(viewData, w.State.Cursor, preview, lanValidationMsg, externalValidationMsg)
	case StepSecurityHeaders:
		viewData := toViewData(*w.State.Data)
		preview := GenerateSecurityHeadersSnippet(w.State.Data.SecurityPreset)
		return views.RenderSecurityHeaders(viewData, w.State.Cursor, preview)
	case StepPerformance:
		viewData := toViewData(*w.State.Data)
		preview := GeneratePerformanceSnippet()
		return views.RenderPerformance(viewData, preview)
	case StepTemplateParams:
		// Get current template being configured
		if w.State.CurrentTemplateIndex < len(w.State.TemplatesToConfigure) {
			templateKey := w.State.TemplatesToConfigure[w.State.CurrentTemplateIndex]
			// Get parameters for this template
			params := make(map[string]string)
			if config, exists := w.State.Data.SnippetConfigs[templateKey]; exists {
				for k, v := range config.Parameters {
					params[k] = fmt.Sprint(v)
				}
			}
			return views.RenderTemplateParams(templateKey, params, w.State.Cursor)
		}
		return "Error: invalid template index"
	case StepCustomSnippet:
		return views.RenderCustomSnippet(
			w.State.Data.CustomNameInput,
			w.State.Data.CustomContentInput,
			w.State.Cursor,
		)
	case StepSummary:
		// Generate snippets before showing summary
		generatedSnippets := w.GenerateSnippets()

		// Convert to views.GeneratedSnippet
		viewSnippets := make([]views.GeneratedSnippet, len(generatedSnippets))
		for i, gen := range generatedSnippets {
			viewSnippets[i] = views.GeneratedSnippet{
				Name:        gen.Name,
				Category:    gen.Category,
				Description: gen.Description,
			}
		}

		viewData := views.SnippetWizardData{
			GeneratedSnippets: viewSnippets,
		}
		return views.RenderSummary(viewData)
	default:
		return "Unknown step"
	}
}

// Next advances to the next step
func (w *Wizard) Next() {
	// If we're on the welcome screen, save the selected mode
	if w.State.CurrentStep == StepWelcome {
		w.State.Data.Mode = WizardMode(w.State.Cursor)
	}

	// Validate before progressing from IP restriction step
	if w.State.CurrentStep == StepIPRestriction && w.State.Data.CreateIPRestriction {
		// Validate LAN Subnet (required)
		if w.State.Data.LANSubnet == "" {
			// Block progression - LAN subnet is required
			return
		}
		lanResult := ValidateCIDR(w.State.Data.LANSubnet)
		if !lanResult.Valid {
			// Block progression - invalid LAN subnet
			return
		}

		// Validate External IP (optional, but must be valid if provided)
		if w.State.Data.AllowedExternalIP != "" {
			externalResult := ValidateCIDR(w.State.Data.AllowedExternalIP)
			if !externalResult.Valid {
				// Block progression - invalid external IP
				return
			}
		}
	}

	nextStep := w.State.NextStep()
	w.State.GoToStep(nextStep)
}

// Back goes to the previous step
func (w *Wizard) Back() {
	w.State.GoBack()
}

// ToggleCheckbox toggles a checkbox at the current step
func (w *Wizard) ToggleCheckbox() {
	switch w.State.CurrentStep {
	case StepTemplateSelection:
		w.toggleTemplateSelection()
	case StepTemplateParams:
		w.toggleTemplateParamCheckbox()
	case StepIPRestriction:
		w.State.Data.CreateIPRestriction = !w.State.Data.CreateIPRestriction
	case StepSecurityHeaders:
		w.State.Data.CreateSecurityHeaders = !w.State.Data.CreateSecurityHeaders
	case StepPerformance:
		w.State.Data.CreatePerformance = !w.State.Data.CreatePerformance
	}
}

// toggleTemplateParamCheckbox toggles checkbox parameters in template config
func (w *Wizard) toggleTemplateParamCheckbox() {
	if w.State.CurrentTemplateIndex >= len(w.State.TemplatesToConfigure) {
		return
	}

	templateKey := w.State.TemplatesToConfigure[w.State.CurrentTemplateIndex]
	config := w.State.Data.SnippetConfigs[templateKey]
	if config.Parameters == nil {
		config.Parameters = make(map[string]interface{})
	}

	switch templateKey {
	case "cors_headers":
		if w.State.Cursor == 2 {
			current := config.Parameters["allow_credentials"]
			if current == "true" {
				config.Parameters["allow_credentials"] = "false"
			} else {
				config.Parameters["allow_credentials"] = "true"
			}
		}
	case "static_caching":
		if w.State.Cursor == 1 {
			current := config.Parameters["enable_etag"]
			if current == "true" {
				config.Parameters["enable_etag"] = "false"
			} else {
				config.Parameters["enable_etag"] = "true"
			}
		}
	case "compression_advanced":
		switch w.State.Cursor {
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
		if w.State.Cursor == 0 {
			current := config.Parameters["skip_verify"]
			if current == "true" {
				config.Parameters["skip_verify"] = "false"
			} else {
				config.Parameters["skip_verify"] = "true"
			}
		}
	case "performance":
		switch w.State.Cursor {
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
		switch w.State.Cursor {
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

	w.State.Data.SnippetConfigs[templateKey] = config
}

// toggleTemplateSelection toggles the template at the current cursor position
func (w *Wizard) toggleTemplateSelection() {
	// Build ordered list of template keys (matches view order)
	templateKeys := []string{
		// Security
		"cors_headers", "rate_limiting", "auth_headers", "ip_restricted", "security_headers",
		// Performance
		"static_caching", "compression_advanced", "performance",
		// Backend
		"websocket_advanced", "extended_timeouts", "https_backend",
		// Content
		"large_uploads", "custom_headers_inject", "frame_embedding", "rewrite_rules",
	}

	if w.State.Cursor >= 0 && w.State.Cursor < len(templateKeys) {
		key := templateKeys[w.State.Cursor]
		w.State.Data.SelectedTemplates[key] = !w.State.Data.SelectedTemplates[key]
	}
}

// SetText sets text for the current focused field
func (w *Wizard) SetText(text string) {
	switch w.State.CurrentStep {
	case StepIPRestriction:
		if w.State.Cursor == 0 {
			w.State.Data.LANSubnet = text
		} else if w.State.Cursor == 1 {
			w.State.Data.AllowedExternalIP = text
		}
	case StepCustomSnippet:
		if w.State.Cursor == 0 {
			w.State.Data.CustomSnippetName = text
		} else if w.State.Cursor == 1 {
			w.State.Data.CustomSnippetContent = text
		}
	case StepTemplateParams:
		w.setTemplateParamText(text)
	case StepSecurityHeaders:
		// Security headers uses radio buttons, not text input
		// Selection is handled by cursor movement
	}
}

// setTemplateParamText sets parameter text for the current template
func (w *Wizard) setTemplateParamText(text string) {
	if w.State.CurrentTemplateIndex >= len(w.State.TemplatesToConfigure) {
		return
	}

	templateKey := w.State.TemplatesToConfigure[w.State.CurrentTemplateIndex]
	config := w.State.Data.SnippetConfigs[templateKey]
	if config.Parameters == nil {
		config.Parameters = make(map[string]interface{})
	}
	if config.ValidationErrors == nil {
		config.ValidationErrors = make(map[string]string)
	}

	var fieldName string

	switch templateKey {
	case "cors_headers":
		switch w.State.Cursor {
		case 0:
			fieldName = "origins"
			config.Parameters["allowed_origins"] = text
		case 1:
			fieldName = "methods"
			config.Parameters["allowed_methods"] = text
		}
	case "rate_limiting":
		switch w.State.Cursor {
		case 0:
			fieldName = "requests_per_second"
			config.Parameters["requests_per_second"] = text
		case 1:
			fieldName = "burst_size"
			config.Parameters["burst_size"] = text
		}
	case "large_uploads":
		fieldName = "max_size"
		config.Parameters["max_size"] = text
	case "extended_timeouts":
		switch w.State.Cursor {
		case 0:
			fieldName = "read_timeout"
			config.Parameters["read_timeout"] = text
		case 1:
			fieldName = "write_timeout"
			config.Parameters["write_timeout"] = text
		case 2:
			fieldName = "dial_timeout"
			config.Parameters["dial_timeout"] = text
		}
	case "static_caching":
		if w.State.Cursor == 0 {
			fieldName = "max_age"
			config.Parameters["max_age"] = text
		}
	case "ip_restricted":
		switch w.State.Cursor {
		case 0:
			fieldName = "lan_subnet"
			config.Parameters["lan_subnet"] = text
		case 1:
			fieldName = "external_ip"
			config.Parameters["external_ip"] = text
		}
	case "compression_advanced":
		if w.State.Cursor == 0 {
			fieldName = "compression_level"
			config.Parameters["compression_level"] = text
		}
	case "https_backend":
		if w.State.Cursor == 1 {
			fieldName = "keepalive"
			config.Parameters["keepalive"] = text
		}
	case "websocket_advanced":
		switch w.State.Cursor {
		case 0:
			fieldName = "upgrade_timeout"
			config.Parameters["upgrade_timeout"] = text
		case 1:
			fieldName = "ping_interval"
			config.Parameters["ping_interval"] = text
		}
	case "rewrite_rules":
		switch w.State.Cursor {
		case 0:
			fieldName = "path_pattern"
			config.Parameters["path_pattern"] = text
		case 1:
			fieldName = "rewrite_to"
			config.Parameters["rewrite_to"] = text
		}
	case "frame_embedding":
		if w.State.Cursor == 0 {
			fieldName = "allowed_origins"
			config.Parameters["allowed_origins"] = text
		}
	}

	// Validate the field
	if fieldName != "" {
		validation := GetValidationForField(templateKey, fieldName, text)
		if !validation.Valid && text != "" {
			config.ValidationErrors[fieldName] = validation.Message
		} else {
			delete(config.ValidationErrors, fieldName)
		}
	}

	w.State.Data.SnippetConfigs[templateKey] = config
}

// SelectOption selects an option at the current cursor position
func (w *Wizard) SelectOption() {
	switch w.State.CurrentStep {
	case StepSecurityHeaders:
		presets := []string{"basic", "strict", "paranoid"}
		if w.State.Cursor >= 0 && w.State.Cursor < len(presets) {
			w.State.Data.SecurityPreset = presets[w.State.Cursor]
		}
	case StepTemplateParams:
		w.selectTemplateParamOption()
	}
}

// selectTemplateParamOption selects an option for templates with radio button choices
func (w *Wizard) selectTemplateParamOption() {
	if w.State.CurrentTemplateIndex >= len(w.State.TemplatesToConfigure) {
		return
	}

	templateKey := w.State.TemplatesToConfigure[w.State.CurrentTemplateIndex]
	config := w.State.Data.SnippetConfigs[templateKey]
	if config.Parameters == nil {
		config.Parameters = make(map[string]interface{})
	}

	switch templateKey {
	case "security_headers":
		presets := []string{"basic", "strict", "paranoid"}
		if w.State.Cursor >= 0 && w.State.Cursor < len(presets) {
			config.Parameters["preset"] = presets[w.State.Cursor]
		}
	case "custom_headers_inject":
		directions := []string{"upstream", "response", "both"}
		if w.State.Cursor >= 0 && w.State.Cursor < len(directions) {
			config.Parameters["direction"] = directions[w.State.Cursor]
		}
	}

	w.State.Data.SnippetConfigs[templateKey] = config
}

// MoveCursorUp moves cursor up in the current step
func (w *Wizard) MoveCursorUp() {
	if w.State.Cursor > 0 {
		w.State.Cursor--
	}
}

// MoveCursorDown moves cursor down in the current step
func (w *Wizard) MoveCursorDown() {
	maxCursor := 0
	switch w.State.CurrentStep {
	case StepWelcome:
		maxCursor = 2 // 3 modes (0, 1, 2)
	case StepTemplateSelection:
		maxCursor = 14 // 15 templates (0-14)
	case StepTemplateParams:
		maxCursor = w.getTemplateParamsCursorMax()
	case StepIPRestriction:
		maxCursor = 1 // 2 fields (LAN subnet, external IP)
	case StepSecurityHeaders:
		maxCursor = 2 // 3 presets
	case StepCustomSnippet:
		maxCursor = 1 // 2 fields (name, content)
	}

	if w.State.Cursor < maxCursor {
		w.State.Cursor++
	}
}

// getTemplateParamsCursorMax returns max cursor for current template params
func (w *Wizard) getTemplateParamsCursorMax() int {
	if w.State.CurrentTemplateIndex >= len(w.State.TemplatesToConfigure) {
		return 0
	}

	templateKey := w.State.TemplatesToConfigure[w.State.CurrentTemplateIndex]
	switch templateKey {
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

// GenerateSnippets generates all selected snippets
func (w *Wizard) GenerateSnippets() []GeneratedSnippet {
	var snippets []GeneratedSnippet

	if w.State.Data.CreateIPRestriction {
		snippets = append(snippets, GeneratedSnippet{
			Name:        "ip_restricted",
			Category:    "Security",
			Content:     GenerateIPRestrictionSnippet(w.State.Data.LANSubnet, w.State.Data.AllowedExternalIP),
			Description: "IP restriction for LAN access",
		})
	}

	if w.State.Data.CreateSecurityHeaders {
		snippets = append(snippets, GeneratedSnippet{
			Name:        "security_headers",
			Category:    "Security",
			Content:     GenerateSecurityHeadersSnippet(w.State.Data.SecurityPreset),
			Description: "Security headers (" + w.State.Data.SecurityPreset + ")",
		})
	}

	if w.State.Data.CreatePerformance {
		snippets = append(snippets, GeneratedSnippet{
			Name:        "performance",
			Category:    "Performance",
			Content:     GeneratePerformanceSnippet(),
			Description: "Performance optimization (compression)",
		})
	}

	// Custom mode snippet
	if w.State.Data.Mode == ModeCustom && w.State.Data.CustomSnippetName != "" && w.State.Data.CustomSnippetContent != "" {
		snippets = append(snippets, GeneratedSnippet{
			Name:        w.State.Data.CustomSnippetName,
			Category:    "Custom",
			Content:     w.State.Data.CustomSnippetContent,
			Description: "Custom pasted snippet",
		})
	}

	// Templated mode: Generate snippets for selected templates with default parameters
	if w.State.Data.Mode == ModeTemplated {
		snippets = append(snippets, w.generateTemplatedSnippets()...)
	}

	w.State.Data.GeneratedSnippets = snippets
	return snippets
}

// generateTemplatedSnippets generates snippets for selected templates with default parameters
func (w *Wizard) generateTemplatedSnippets() []GeneratedSnippet {
	var snippets []GeneratedSnippet

	for templateKey, selected := range w.State.Data.SelectedTemplates {
		if !selected {
			continue
		}

		var content string
		var description string
		category := "Unknown"

		// Get template metadata
		availableTemplates := GetAvailableTemplates()
		if tmpl, exists := availableTemplates[templateKey]; exists {
			category = tmpl.Category
			description = tmpl.Description
		}

		// Get configured parameters or use defaults
		config := w.State.Data.SnippetConfigs[templateKey]
		params := config.Parameters

		// Generate content with configured parameters
		switch templateKey {
		case "cors_headers":
			origins := getStringParam(params, "allowed_origins", "*")
			methods := getStringParam(params, "allowed_methods", "GET, POST, PUT, DELETE, OPTIONS")
			credentials := getBoolParam(params, "allow_credentials", false)
			content = GenerateCORSHeadersSnippet(origins, methods, credentials)
		case "rate_limiting":
			rps := getIntParam(params, "requests_per_second", 100)
			burst := getIntParam(params, "burst_size", 50)
			content = GenerateRateLimitingSnippet(rps, burst, "default")
		case "ip_restricted":
			lanSubnet := getStringParam(params, "lan_subnet", "10.0.0.0/8")
			externalIP := getStringParam(params, "external_ip", "")
			content = GenerateIPRestrictionSnippet(lanSubnet, externalIP)
		case "security_headers":
			preset := getStringParam(params, "preset", "strict")
			content = GenerateSecurityHeadersSnippet(preset)
		case "static_caching":
			maxAge := getIntParam(params, "max_age", 86400)
			etag := getBoolParam(params, "enable_etag", true)
			content = GenerateStaticCachingSnippet(maxAge, nil, etag)
		case "compression_advanced":
			level := getIntParam(params, "compression_level", 5)
			var encoders []string
			if getBoolParam(params, "enable_gzip", true) {
				encoders = append(encoders, "gzip")
			}
			if getBoolParam(params, "enable_zstd", true) {
				encoders = append(encoders, "zstd")
			}
			if getBoolParam(params, "enable_brotli", false) {
				encoders = append(encoders, "brotli")
			}
			if len(encoders) == 0 {
				encoders = []string{"gzip", "zstd"} // Fallback to defaults
			}
			content = GenerateCompressionAdvancedSnippet(encoders, level, nil)
		case "performance":
			// Use configured encoders or defaults
			var encoders []string
			if getBoolParam(params, "enable_gzip", true) {
				encoders = append(encoders, "gzip")
			}
			if getBoolParam(params, "enable_zstd", true) {
				encoders = append(encoders, "zstd")
			}
			// Generate snippet with selected encoders
			content = GeneratePerformanceSnippet()
		case "auth_headers":
			forwardIP := getBoolParam(params, "forward_ip", true)
			forwardProto := getBoolParam(params, "forward_proto", true)
			content = GenerateAuthHeadersSnippet(forwardIP, forwardProto, nil)
		case "websocket_advanced":
			upgradeTimeout := getIntParam(params, "upgrade_timeout", 10)
			pingInterval := getIntParam(params, "ping_interval", 30)
			content = GenerateWebSocketAdvancedSnippet(upgradeTimeout, pingInterval)
		case "extended_timeouts":
			readTimeout := getStringParam(params, "read_timeout", "120s")
			writeTimeout := getStringParam(params, "write_timeout", "120s")
			dialTimeout := getStringParam(params, "dial_timeout", "30s")
			content = GenerateExtendedTimeoutsSnippet(readTimeout, writeTimeout, dialTimeout)
		case "https_backend":
			skipVerify := getBoolParam(params, "skip_verify", false)
			keepalive := getIntParam(params, "keepalive", 100)
			content = GenerateHTTPSBackendSnippet(skipVerify, keepalive)
		case "large_uploads":
			maxSize := getStringParam(params, "max_size", "512MB")
			content = GenerateLargeUploadsSnippet(maxSize)
		case "custom_headers_inject":
			direction := getStringParam(params, "direction", "upstream")
			content = GenerateCustomHeadersInjectSnippet(map[string]string{"X-Custom": "value"}, direction)
		case "frame_embedding":
			allowedOrigins := getStringParam(params, "allowed_origins", "'self'")
			content = GenerateFrameEmbeddingSnippet(allowedOrigins)
		case "rewrite_rules":
			pathPattern := getStringParam(params, "path_pattern", "/api/*")
			rewriteTo := getStringParam(params, "rewrite_to", "/new-api{uri}")
			content = GenerateRewriteRulesSnippet(pathPattern, rewriteTo)
		default:
			continue
		}

		snippets = append(snippets, GeneratedSnippet{
			Name:        templateKey,
			Category:    category,
			Content:     content,
			Description: description,
		})
	}

	return snippets
}

// GetProgress returns the current progress percentage
func (w *Wizard) GetProgress() int {
	return w.State.ProgressPercentage()
}

// Parameter helper functions

func getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if params == nil {
		return defaultValue
	}
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}
	return defaultValue
}

func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if params == nil {
		return defaultValue
	}
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case string:
			// Try parsing string to int
			if v != "" {
				var parsed int
				if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil {
					return parsed
				}
			}
		}
	}
	return defaultValue
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if params == nil {
		return defaultValue
	}
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true"
		}
	}
	return defaultValue
}
