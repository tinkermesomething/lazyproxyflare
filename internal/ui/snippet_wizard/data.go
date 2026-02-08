package snippet_wizard

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

// SnippetCategory represents a snippet category (temporary until caddy package is updated)
type SnippetCategory int

const (
	CategoryUnknown SnippetCategory = iota
	CategorySecurity
	CategoryPerformance
	CategoryBackend
	CategoryContent
	CategoryAdvanced
)

// WizardMode represents the creation mode selected by user
type WizardMode int

const (
	ModeTemplated WizardMode = iota // Templated - Advanced
	ModeCustom                       // Custom - Paste Your Own
	ModeGuided                       // Guided - Step by Step
)

// String returns human-readable mode name
func (m WizardMode) String() string {
	switch m {
	case ModeTemplated:
		return "Templated - Advanced"
	case ModeCustom:
		return "Custom - Paste Your Own"
	case ModeGuided:
		return "Guided - Step by Step"
	default:
		return "Unknown"
	}
}

// SnippetType represents the type of snippet being created
type SnippetType int

const (
	// Existing snippet types
	SnippetIPRestricted SnippetType = iota
	SnippetSecurityHeaders
	SnippetPerformance

	// New snippet types (Phase 3-4)
	SnippetCORSHeaders
	SnippetRateLimiting
	SnippetAuthHeaders
	SnippetStaticCaching
	SnippetCompressionAdvanced
	SnippetWebSocketAdvanced
	SnippetExtendedTimeouts
	SnippetLargeUploads
	SnippetCustomHeadersInject
	SnippetRewriteRules
	SnippetFrameEmbedding
	SnippetHTTPSBackend
)

// String returns human-readable snippet type name
func (s SnippetType) String() string {
	switch s {
	case SnippetIPRestricted:
		return "IP Restriction"
	case SnippetSecurityHeaders:
		return "Security Headers"
	case SnippetPerformance:
		return "Performance"
	case SnippetCORSHeaders:
		return "CORS Headers"
	case SnippetRateLimiting:
		return "Rate Limiting"
	case SnippetAuthHeaders:
		return "Auth Headers"
	case SnippetStaticCaching:
		return "Static Caching"
	case SnippetCompressionAdvanced:
		return "Compression Advanced"
	case SnippetWebSocketAdvanced:
		return "WebSocket Advanced"
	case SnippetExtendedTimeouts:
		return "Extended Timeouts"
	case SnippetLargeUploads:
		return "Large Uploads"
	case SnippetCustomHeadersInject:
		return "Custom Headers Inject"
	case SnippetRewriteRules:
		return "Rewrite Rules"
	case SnippetFrameEmbedding:
		return "Frame Embedding"
	case SnippetHTTPSBackend:
		return "HTTPS Backend"
	default:
		return "Unknown"
	}
}

// Category returns the category this snippet belongs to
func (s SnippetType) Category() SnippetCategory {
	switch s {
	case SnippetIPRestricted, SnippetSecurityHeaders, SnippetCORSHeaders, SnippetRateLimiting, SnippetAuthHeaders:
		return CategorySecurity
	case SnippetPerformance, SnippetStaticCaching, SnippetCompressionAdvanced:
		return CategoryPerformance
	case SnippetWebSocketAdvanced, SnippetExtendedTimeouts, SnippetHTTPSBackend:
		return CategoryBackend
	case SnippetLargeUploads, SnippetCustomHeadersInject, SnippetFrameEmbedding:
		return CategoryContent
	case SnippetRewriteRules:
		return CategoryAdvanced
	default:
		return CategoryUnknown
	}
}

// PatternType represents a type of detected pattern in Caddyfile
type PatternType int

const (
	PatternRequestBodyLimit PatternType = iota
	PatternExtendedTimeouts
	PatternOAuthForwarding
	PatternKeepaliveCustom
	PatternCORSHeaders
	PatternWebSocketExtended
	PatternCustomTransport
)

// String returns human-readable pattern type name
func (p PatternType) String() string {
	switch p {
	case PatternRequestBodyLimit:
		return "Request Body Limit"
	case PatternExtendedTimeouts:
		return "Extended Timeouts"
	case PatternOAuthForwarding:
		return "OAuth Forwarding"
	case PatternKeepaliveCustom:
		return "Keepalive Custom"
	case PatternCORSHeaders:
		return "CORS Headers"
	case PatternWebSocketExtended:
		return "WebSocket Extended"
	case PatternCustomTransport:
		return "Custom Transport"
	default:
		return "Unknown"
	}
}

// DetectedPattern represents a pattern found in the user's Caddyfile
type DetectedPattern struct {
	Type          PatternType       // Type of pattern detected
	Count         int               // Number of occurrences
	ExampleDomain string            // Example domain using this pattern
	RawDirective  string            // Actual Caddy config snippet
	SuggestedName string            // Proposed snippet name
	Parameters    map[string]string // Extracted parameters
}

// SnippetConfig represents configuration for a single snippet
type SnippetConfig struct {
	Type             SnippetType            // Type of snippet
	Name             string                 // Snippet name
	Category         SnippetCategory        // Category
	Enabled          bool                   // Whether this snippet is enabled
	Parameters       map[string]interface{} // Dynamic parameters
	RawContent       string                 // For custom mode - raw pasted content
	ValidationErrors map[string]string      // Validation errors for parameters (key: field name, value: error message)
}

// SnippetWizardData holds all data collected during snippet wizard
type SnippetWizardData struct {
	// Mode selection
	Mode WizardMode // Which creation mode is active

	// Category selection
	EnabledCategories map[string]bool // Which categories to configure

	// Auto-detection
	DetectedPatterns []DetectedPattern   // Patterns found in Caddyfile
	SelectedPatterns map[string]bool     // Which patterns user selected to extract

	// Dynamic snippet configurations
	SnippetConfigs map[string]SnippetConfig // Key: snippet name

	// Templated mode fields
	SelectedTemplates map[string]bool // Which templates user selected (key: template name)

	// Custom mode fields
	CustomSnippetName      string             // User-provided snippet name
	CustomSnippetContent   string             // User-pasted snippet content
	CustomNameInput        textinput.Model    // Text input for snippet name (supports paste)
	CustomContentInput     textarea.Model     // Textarea for snippet content (multi-line, supports paste)

	// Legacy fields (backward compatibility for existing 3 snippets)
	CreateIPRestriction   bool
	LANSubnet             string // e.g., "10.0.28.0/24"
	AllowedExternalIP     string // optional, e.g., "166.1.123.74/32"
	CreateSecurityHeaders bool
	SecurityPreset        string // "basic", "strict", "paranoid"
	CreatePerformance     bool

	// Generated output
	GeneratedSnippets []GeneratedSnippet
}

// GeneratedSnippet represents a snippet to be created
type GeneratedSnippet struct {
	Name        string
	Category    string
	Content     string
	Description string
}
