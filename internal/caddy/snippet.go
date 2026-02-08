package caddy

import (
	"fmt"
	"strings"
)

// SnippetCategory represents the categorization of a Caddy snippet
type SnippetCategory int

const (
	SnippetUnknown SnippetCategory = iota
	SnippetIPRestriction
	SnippetSecurityHeaders
	SnippetPerformance
	SnippetHTTPSBackend
	SnippetOAuthHeaders
	SnippetWebSocketHeaders
	SnippetFrameEmbedding
	SnippetCORS
	SnippetCompression
	SnippetRateLimit
	SnippetCustom
)

// String returns human-readable category name
func (c SnippetCategory) String() string {
	switch c {
	case SnippetIPRestriction:
		return "IP Restriction"
	case SnippetSecurityHeaders:
		return "Security Headers"
	case SnippetPerformance:
		return "Performance"
	case SnippetHTTPSBackend:
		return "HTTPS Backend"
	case SnippetOAuthHeaders:
		return "OAuth Headers"
	case SnippetWebSocketHeaders:
		return "WebSocket Headers"
	case SnippetFrameEmbedding:
		return "Frame Embedding"
	case SnippetCORS:
		return "CORS"
	case SnippetCompression:
		return "Compression"
	case SnippetRateLimit:
		return "Rate Limiting"
	case SnippetCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// ColorCode returns the color for UI rendering
func (c SnippetCategory) ColorCode() string {
	switch c {
	case SnippetIPRestriction:
		return "#FF6B6B" // Red - security
	case SnippetSecurityHeaders:
		return "#FF6B6B" // Red - security
	case SnippetPerformance:
		return "#4ECDC4" // Cyan - performance
	case SnippetHTTPSBackend:
		return "#95E1D3" // Light green - connection
	case SnippetOAuthHeaders:
		return "#FFA07A" // Light orange - auth
	case SnippetWebSocketHeaders:
		return "#DDA0DD" // Plum - websockets
	case SnippetFrameEmbedding:
		return "#F9CA24" // Yellow - iframe
	case SnippetCORS:
		return "#6C5CE7" // Purple - CORS
	case SnippetCompression:
		return "#00B894" // Green - optimization
	case SnippetRateLimit:
		return "#FDCB6E" // Gold - rate limit
	case SnippetCustom:
		return "#B2BEC3" // Gray - custom
	default:
		return "#7F8C8D" // Dark gray - unknown
	}
}

// Snippet represents a Caddy snippet with metadata
type Snippet struct {
	Name         string          // e.g., "ip_restricted", "security_headers"
	Category     SnippetCategory // Auto-detected or manually set category
	Content      string          // Raw Caddy config block (without the (name) { } wrapper)
	Description  string          // User-friendly description
	LineStart    int             // Location in Caddyfile (1-indexed)
	LineEnd      int             // End location in Caddyfile (1-indexed)
	AutoDetected bool            // Was category auto-detected?
	Confidence   float64         // Confidence score for auto-detection (0.0-1.0)
}

// FullBlock returns the complete snippet definition including (name) { }
func (s *Snippet) FullBlock() string {
	return fmt.Sprintf("(%s) {\n%s\n}", s.Name, s.Content)
}

// Lines returns the number of lines in the snippet
func (s *Snippet) Lines() int {
	return len(strings.Split(s.Content, "\n"))
}

// CategorizationHint represents a pattern for auto-detecting snippet categories
type CategorizationHint struct {
	Category   SnippetCategory
	Keywords   []string  // Keywords to look for in content
	MinMatches int       // Minimum keyword matches required
	Confidence float64   // Confidence if all keywords match
}

// Auto-categorization hints for snippet detection
var categorizationHints = []CategorizationHint{
	{
		Category:   SnippetIPRestriction,
		Keywords:   []string{"remote_ip", "not remote_ip", "@external", "respond 403"},
		MinMatches: 2,
		Confidence: 0.95,
	},
	{
		Category:   SnippetSecurityHeaders,
		Keywords:   []string{"header", "X-Frame-Options", "X-Content-Type-Options", "Strict-Transport-Security"},
		MinMatches: 2,
		Confidence: 0.9,
	},
	{
		Category:   SnippetPerformance,
		Keywords:   []string{"encode", "gzip", "cache", "expires"},
		MinMatches: 1,
		Confidence: 0.8,
	},
	{
		Category:   SnippetHTTPSBackend,
		Keywords:   []string{"transport http", "tls_insecure_skip_verify"},
		MinMatches: 1,
		Confidence: 0.95,
	},
	{
		Category:   SnippetOAuthHeaders,
		Keywords:   []string{"header_up X-Real-IP", "header_up X-Forwarded-For", "header_up X-Forwarded-Proto"},
		MinMatches: 2,
		Confidence: 0.9,
	},
	{
		Category:   SnippetWebSocketHeaders,
		Keywords:   []string{"header_up Upgrade", "header_up Connection", "websocket"},
		MinMatches: 2,
		Confidence: 0.95,
	},
	{
		Category:   SnippetFrameEmbedding,
		Keywords:   []string{"X-Frame-Options SAMEORIGIN", "frame-ancestors"},
		MinMatches: 1,
		Confidence: 0.9,
	},
	{
		Category:   SnippetCORS,
		Keywords:   []string{"Access-Control-Allow-Origin", "Access-Control-Allow-Methods", "cors"},
		MinMatches: 1,
		Confidence: 0.9,
	},
	{
		Category:   SnippetCompression,
		Keywords:   []string{"encode gzip", "encode zstd", "encode br"},
		MinMatches: 1,
		Confidence: 0.85,
	},
	{
		Category:   SnippetRateLimit,
		Keywords:   []string{"rate_limit", "limit_req"},
		MinMatches: 1,
		Confidence: 0.9,
	},
}

// CategorizeSnippet auto-detects the category of a snippet based on its content
// Returns the detected category and a confidence score (0.0-1.0)
func CategorizeSnippet(content string) (SnippetCategory, float64) {
	contentLower := strings.ToLower(content)

	var bestCategory SnippetCategory = SnippetUnknown
	var bestConfidence float64 = 0.0

	for _, hint := range categorizationHints {
		matches := 0
		for _, keyword := range hint.Keywords {
			if strings.Contains(contentLower, strings.ToLower(keyword)) {
				matches++
			}
		}

		if matches >= hint.MinMatches {
			// Calculate confidence based on match percentage
			matchRatio := float64(matches) / float64(len(hint.Keywords))
			confidence := hint.Confidence * matchRatio

			if confidence > bestConfidence {
				bestCategory = hint.Category
				bestConfidence = confidence
			}
		}
	}

	return bestCategory, bestConfidence
}

// GenerateDescription creates a user-friendly description based on category and content
func GenerateDescription(category SnippetCategory, content string) string {
	switch category {
	case SnippetIPRestriction:
		if strings.Contains(content, "remote_ip") {
			return "Restricts access based on IP address ranges"
		}
		return "IP-based access control"

	case SnippetSecurityHeaders:
		headers := []string{}
		if strings.Contains(content, "X-Frame-Options") {
			headers = append(headers, "clickjacking protection")
		}
		if strings.Contains(content, "X-Content-Type-Options") {
			headers = append(headers, "MIME sniffing protection")
		}
		if strings.Contains(content, "Strict-Transport-Security") {
			headers = append(headers, "HSTS")
		}
		if len(headers) > 0 {
			return fmt.Sprintf("Security headers: %s", strings.Join(headers, ", "))
		}
		return "Security headers for enhanced protection"

	case SnippetPerformance:
		if strings.Contains(content, "encode") {
			return "Response compression for faster loading"
		}
		if strings.Contains(content, "cache") {
			return "Caching configuration for better performance"
		}
		return "Performance optimization"

	case SnippetHTTPSBackend:
		return "HTTPS connection to upstream backend"

	case SnippetOAuthHeaders:
		return "OAuth/SSO authentication headers"

	case SnippetWebSocketHeaders:
		return "WebSocket protocol support"

	case SnippetFrameEmbedding:
		return "Controls iframe embedding behavior"

	case SnippetCORS:
		return "Cross-Origin Resource Sharing (CORS) configuration"

	case SnippetCompression:
		return "Response compression (gzip/brotli/zstd)"

	case SnippetRateLimit:
		return "Rate limiting to prevent abuse"

	case SnippetCustom:
		return "Custom configuration snippet"

	default:
		return "Caddy configuration snippet"
	}
}
