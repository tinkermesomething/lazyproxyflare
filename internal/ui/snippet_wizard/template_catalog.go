package snippet_wizard

// GetAvailableTemplates returns all available templates with their metadata
func GetAvailableTemplates() map[string]struct {
	Name        string
	Description string
	Category    string
} {
	return map[string]struct {
		Name        string
		Description string
		Category    string
	}{
		// Security
		"cors_headers": {
			Name:        "CORS Headers",
			Description: "Cross-Origin Resource Sharing configuration with preflight",
			Category:    "Security",
		},
		"rate_limiting": {
			Name:        "Rate Limiting",
			Description: "Zone-based request rate limiting (100 req/s default)",
			Category:    "Security",
		},
		"auth_headers": {
			Name:        "Auth Headers",
			Description: "Forward authentication headers (X-Real-IP, X-Forwarded-*)",
			Category:    "Security",
		},
		"ip_restricted": {
			Name:        "IP Restriction",
			Description: "Limit access to LAN or specific IPs",
			Category:    "Security",
		},
		"security_headers": {
			Name:        "Security Headers",
			Description: "Security response headers (basic/strict/paranoid)",
			Category:    "Security",
		},

		// Performance
		"static_caching": {
			Name:        "Static Caching",
			Description: "Cache-Control headers for static assets (1 day default)",
			Category:    "Performance",
		},
		"compression_advanced": {
			Name:        "Advanced Compression",
			Description: "Multi-encoder compression (gzip, zstd, brotli)",
			Category:    "Performance",
		},
		"performance": {
			Name:        "Basic Performance",
			Description: "Standard gzip + zstd compression",
			Category:    "Performance",
		},

		// Backend Integration
		"websocket_advanced": {
			Name:        "WebSocket Support",
			Description: "Full WebSocket upgrade with infinite timeouts",
			Category:    "Backend Integration",
		},
		"extended_timeouts": {
			Name:        "Extended Timeouts",
			Description: "Custom read/write/dial timeout configuration",
			Category:    "Backend Integration",
		},
		"https_backend": {
			Name:        "HTTPS Backend",
			Description: "HTTPS upstream with TLS skip verify option",
			Category:    "Backend Integration",
		},

		// Content Control
		"large_uploads": {
			Name:        "Large Uploads",
			Description: "Request body size limits (512MB default)",
			Category:    "Content Control",
		},
		"custom_headers_inject": {
			Name:        "Custom Headers",
			Description: "Inject custom headers (upstream/response/both)",
			Category:    "Content Control",
		},
		"frame_embedding": {
			Name:        "Frame Embedding",
			Description: "CSP frame-ancestors for iframe control",
			Category:    "Content Control",
		},
		"rewrite_rules": {
			Name:        "URL Rewriting",
			Description: "Matcher-based URL path rewriting",
			Category:    "Content Control",
		},
	}
}
