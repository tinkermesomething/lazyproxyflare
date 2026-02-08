package snippet_wizard

// WizardStep represents the current step in the snippet creation wizard
type WizardStep int

const (
	// Phase 1: Existing steps (backward compatibility)
	StepWelcome WizardStep = iota
	StepAutoDetect          // Phase 2: Show detected patterns from Caddyfile
	StepTemplateSelection   // Phase 5: Select which templates to configure
	StepTemplateParams      // Phase 5: Configure parameters for selected templates
	StepIPRestriction
	StepSecurityHeaders
	StepPerformance
	StepCustomSnippet       // Phase 5: Free-form paste or templated creation
	StepSummary

	// Phase 2+: New steps for expanded wizard (future)
	StepModeSelection     // Choose Templated/Custom/Guided
	StepCategoryHub       // Select which categories to configure
	StepSecurityConfig    // Configure security snippets
	StepPerformanceConfig // Configure performance snippets
	StepBackendConfig     // Configure backend integration snippets
	StepContentConfig     // Configure content control snippets
	StepGuidedFlow        // Question-based guided creation
	StepFinalSummary      // Review all configured snippets with preview
	StepConfirmation      // Final confirmation before creation
)

// String returns human-readable step name
func (s WizardStep) String() string {
	switch s {
	case StepWelcome:
		return "Welcome"
	case StepIPRestriction:
		return "IP Restriction"
	case StepSecurityHeaders:
		return "Security Headers"
	case StepPerformance:
		return "Performance"
	case StepSummary:
		return "Summary"
	case StepModeSelection:
		return "Mode Selection"
	case StepAutoDetect:
		return "Auto-Detection"
	case StepCategoryHub:
		return "Category Hub"
	case StepSecurityConfig:
		return "Security Configuration"
	case StepPerformanceConfig:
		return "Performance Configuration"
	case StepBackendConfig:
		return "Backend Configuration"
	case StepContentConfig:
		return "Content Configuration"
	case StepCustomSnippet:
		return "Custom Snippet"
	case StepGuidedFlow:
		return "Guided Flow"
	case StepFinalSummary:
		return "Final Summary"
	case StepConfirmation:
		return "Confirmation"
	default:
		return "Unknown"
	}
}

// WizardState manages wizard navigation and state
type WizardState struct {
	CurrentStep          WizardStep
	CompletedSteps       map[WizardStep]bool
	VisitedSteps         []WizardStep
	Data                 *SnippetWizardData
	Cursor               int      // Current cursor position within a step
	CurrentTemplateIndex int      // Index of template being configured (for params step)
	TemplatesToConfigure []string // Ordered list of templates to configure
}

// NewWizardState creates a new wizard state
func NewWizardState() *WizardState {
	return &WizardState{
		CurrentStep:    StepWelcome,
		CompletedSteps: make(map[WizardStep]bool),
		VisitedSteps:   []WizardStep{StepWelcome},
		Data: &SnippetWizardData{
			EnabledCategories: make(map[string]bool),
			SelectedPatterns:  make(map[string]bool),
			SelectedTemplates: make(map[string]bool),
			SnippetConfigs:    make(map[string]SnippetConfig),
		},
		Cursor: 0,
	}
}

// NextStep determines the next step based on current state and user selections
// This implements smart navigation that skips unchecked categories
func (ws *WizardState) NextStep() WizardStep {
	// Mark current step as completed
	ws.CompletedSteps[ws.CurrentStep] = true

	// Mode-based navigation from Welcome screen
	if ws.CurrentStep == StepWelcome {
		switch ws.Data.Mode {
		case ModeTemplated:
			// Templated mode: Show auto-detect, then legacy flow
			return StepAutoDetect
		case ModeCustom:
			// Custom mode: Jump directly to custom snippet paste
			return StepCustomSnippet
		case ModeGuided:
			// Guided mode: Jump to guided flow (TBD)
			return StepGuidedFlow
		default:
			// Default: Legacy flow with auto-detect
			return StepAutoDetect
		}
	}

	// Templated mode flow (Phase 5)
	if ws.Data.Mode == ModeTemplated || ws.Data.Mode == 0 {
		switch ws.CurrentStep {
		case StepAutoDetect:
			return StepTemplateSelection
		case StepTemplateSelection:
			// Build list of templates that need configuration
			ws.buildTemplatesToConfigure()
			if len(ws.TemplatesToConfigure) > 0 {
				ws.CurrentTemplateIndex = 0
				return StepTemplateParams
			}
			return StepSummary
		case StepTemplateParams:
			// Move to next template or summary
			ws.CurrentTemplateIndex++
			if ws.CurrentTemplateIndex < len(ws.TemplatesToConfigure) {
				return StepTemplateParams
			}
			return StepSummary
		case StepIPRestriction:
			return StepSecurityHeaders
		case StepSecurityHeaders:
			return StepPerformance
		case StepPerformance:
			return StepSummary
		case StepCustomSnippet:
			return StepSummary
		default:
			return StepSummary
		}
	}

	// Phase 2+: New flow with mode selection
	switch ws.CurrentStep {
	case StepModeSelection:
		return StepAutoDetect
	case StepAutoDetect:
		return StepCategoryHub
	case StepCategoryHub:
		// Navigate to first enabled category
		if ws.Data.EnabledCategories["security"] {
			return StepSecurityConfig
		}
		if ws.Data.EnabledCategories["performance"] {
			return StepPerformanceConfig
		}
		if ws.Data.EnabledCategories["backend"] {
			return StepBackendConfig
		}
		if ws.Data.EnabledCategories["content"] {
			return StepContentConfig
		}
		// No categories selected, go to custom or summary
		return StepCustomSnippet
	case StepSecurityConfig:
		// Navigate to next enabled category
		if ws.Data.EnabledCategories["performance"] {
			return StepPerformanceConfig
		}
		if ws.Data.EnabledCategories["backend"] {
			return StepBackendConfig
		}
		if ws.Data.EnabledCategories["content"] {
			return StepContentConfig
		}
		return StepCustomSnippet
	case StepPerformanceConfig:
		if ws.Data.EnabledCategories["backend"] {
			return StepBackendConfig
		}
		if ws.Data.EnabledCategories["content"] {
			return StepContentConfig
		}
		return StepCustomSnippet
	case StepBackendConfig:
		if ws.Data.EnabledCategories["content"] {
			return StepContentConfig
		}
		return StepCustomSnippet
	case StepContentConfig:
		return StepCustomSnippet
	case StepCustomSnippet:
		return StepFinalSummary
	case StepFinalSummary:
		return StepConfirmation
	default:
		return StepSummary
	}
}

// PreviousStep returns the previous step
func (ws *WizardState) PreviousStep() WizardStep {
	if len(ws.VisitedSteps) > 1 {
		return ws.VisitedSteps[len(ws.VisitedSteps)-2]
	}
	return ws.CurrentStep
}

// GoToStep transitions to a new step
func (ws *WizardState) GoToStep(step WizardStep) {
	ws.CurrentStep = step
	ws.VisitedSteps = append(ws.VisitedSteps, step)
	ws.Cursor = 0 // Reset cursor when changing steps
}

// GoBack goes to the previous step
func (ws *WizardState) GoBack() {
	if len(ws.VisitedSteps) > 1 {
		// Remove current step
		ws.VisitedSteps = ws.VisitedSteps[:len(ws.VisitedSteps)-1]
		// Go to previous
		ws.CurrentStep = ws.VisitedSteps[len(ws.VisitedSteps)-1]
		ws.Cursor = 0

		// If going back to template params, decrement the index
		if ws.CurrentStep == StepTemplateParams && ws.CurrentTemplateIndex > 0 {
			ws.CurrentTemplateIndex--
		}
	}
}

// buildTemplatesToConfigure builds the list of templates that need parameter configuration
func (ws *WizardState) buildTemplatesToConfigure() {
	// Templates that support parameter configuration
	configurableTemplates := map[string]bool{
		"cors_headers":          true,
		"rate_limiting":         true,
		"large_uploads":         true,
		"extended_timeouts":     true,
		"static_caching":        true,
		"ip_restricted":         true,
		"security_headers":      true,
		"compression_advanced":  true,
		"https_backend":         true,
		"performance":           true,
		"auth_headers":          true,
		"websocket_advanced":    true,
		"custom_headers_inject": true,
		"rewrite_rules":         true,
		"frame_embedding":       true,
	}

	ws.TemplatesToConfigure = []string{}
	for templateKey, selected := range ws.Data.SelectedTemplates {
		if selected && configurableTemplates[templateKey] {
			ws.TemplatesToConfigure = append(ws.TemplatesToConfigure, templateKey)
			// Initialize parameters map for this template if not exists
			if ws.Data.SnippetConfigs[templateKey].Parameters == nil {
				config := ws.Data.SnippetConfigs[templateKey]
				config.Parameters = make(map[string]interface{})
				ws.Data.SnippetConfigs[templateKey] = config
			}
		}
	}
}

// ProgressPercentage calculates wizard completion percentage
func (ws *WizardState) ProgressPercentage() int {
	totalSteps := 5 // Legacy flow has 5 steps
	if ws.Data.Mode != 0 {
		// New flow - calculate based on enabled categories
		totalSteps = 3 // ModeSelection, AutoDetect, CategoryHub
		if ws.Data.EnabledCategories["security"] {
			totalSteps++
		}
		if ws.Data.EnabledCategories["performance"] {
			totalSteps++
		}
		if ws.Data.EnabledCategories["backend"] {
			totalSteps++
		}
		if ws.Data.EnabledCategories["content"] {
			totalSteps++
		}
		totalSteps += 3 // CustomSnippet, FinalSummary, Confirmation
	}

	completed := len(ws.CompletedSteps)
	if totalSteps == 0 {
		return 0
	}
	return (completed * 100) / totalSteps
}
