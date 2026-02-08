package views

import (
	"fmt"
	"strings"
)

// RenderTemplateParams renders parameter configuration for a specific template
func RenderTemplateParams(templateName string, params map[string]string, cursor int) string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Configure: %s", templateName)))
	b.WriteString("\n\n")

	// Render fields based on template type
	switch templateName {
	case "cors_headers":
		return renderCORSParams(params, cursor)
	case "rate_limiting":
		return renderRateLimitParams(params, cursor)
	case "large_uploads":
		return renderLargeUploadsParams(params, cursor)
	case "extended_timeouts":
		return renderExtendedTimeoutsParams(params, cursor)
	case "static_caching":
		return renderStaticCachingParams(params, cursor)
	case "ip_restricted":
		return renderIPRestrictionParams(params, cursor)
	case "security_headers":
		return renderSecurityHeadersParams(params, cursor)
	case "compression_advanced":
		return renderCompressionAdvancedParams(params, cursor)
	case "https_backend":
		return renderHTTPSBackendParams(params, cursor)
	case "performance":
		return renderPerformanceParams(params, cursor)
	case "auth_headers":
		return renderAuthHeadersParams(params, cursor)
	case "websocket_advanced":
		return renderWebSocketAdvancedParams(params, cursor)
	case "custom_headers_inject":
		return renderCustomHeadersInjectParams(params, cursor)
	case "rewrite_rules":
		return renderRewriteRulesParams(params, cursor)
	case "frame_embedding":
		return renderFrameEmbeddingParams(params, cursor)
	default:
		// Template doesn't need configuration
		b.WriteString(StyleInfo.Render("This template uses default configuration."))
		b.WriteString("\n\n")
		b.WriteString(RenderNavigationHint("Enter: continue"))
		return b.String()
	}
}

func renderCORSParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: CORS Headers"))
	b.WriteString("\n\n")

	// Field 0: Allowed Origins
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Allowed Origins:"))
	} else {
		b.WriteString(StyleDim.Render("  Allowed Origins:"))
	}
	b.WriteString(" ")
	origin := params["allowed_origins"]
	if origin == "" {
		b.WriteString(StyleDim.Render("(default: *)"))
	} else {
		b.WriteString(origin)
	}
	b.WriteString("\n")
	if cursor == 0 {
		b.WriteString(StyleDim.Render("  Examples: *, https://example.com, https://app1.com https://app2.com"))
	}
	b.WriteString("\n\n")

	// Field 1: Allowed Methods
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ Allowed Methods:"))
	} else {
		b.WriteString(StyleDim.Render("  Allowed Methods:"))
	}
	b.WriteString(" ")
	methods := params["allowed_methods"]
	if methods == "" {
		b.WriteString(StyleDim.Render("(default: GET, POST, PUT, DELETE, OPTIONS)"))
	} else {
		b.WriteString(methods)
	}
	b.WriteString("\n")
	if cursor == 1 {
		b.WriteString(StyleDim.Render("  Comma-separated list of HTTP methods"))
	}
	b.WriteString("\n\n")

	// Field 2: Allow Credentials (checkbox)
	if cursor == 2 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	credentials := params["allow_credentials"]
	if credentials == "true" {
		b.WriteString(StyleKeybinding.Render("[✓] Allow Credentials"))
	} else {
		b.WriteString(StyleDim.Render("[ ] Allow Credentials"))
	}
	b.WriteString("\n")
	if cursor == 2 {
		b.WriteString(StyleDim.Render("  Enable Access-Control-Allow-Credentials header"))
	}
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Space: toggle checkbox", "Enter: continue"))

	return b.String()
}

func renderRateLimitParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Rate Limiting"))
	b.WriteString("\n\n")

	// Field 0: Requests per second
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Requests per Second:"))
	} else {
		b.WriteString(StyleDim.Render("  Requests per Second:"))
	}
	b.WriteString(" ")
	rps := params["requests_per_second"]
	if rps == "" {
		b.WriteString(StyleDim.Render("(default: 100)"))
	} else {
		b.WriteString(rps)
	}
	b.WriteString("\n\n")

	// Field 1: Burst size
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ Burst Size:"))
	} else {
		b.WriteString(StyleDim.Render("  Burst Size:"))
	}
	b.WriteString(" ")
	burst := params["burst_size"]
	if burst == "" {
		b.WriteString(StyleDim.Render("(default: 50)"))
	} else {
		b.WriteString(burst)
	}
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("  Number of requests allowed to burst above the rate"))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Enter: continue"))

	return b.String()
}

func renderLargeUploadsParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Large Uploads"))
	b.WriteString("\n\n")

	// Field 0: Max size
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Maximum Upload Size:"))
	} else {
		b.WriteString(StyleDim.Render("  Maximum Upload Size:"))
	}
	b.WriteString(" ")
	maxSize := params["max_size"]
	if maxSize == "" {
		b.WriteString(StyleDim.Render("(default: 512MB)"))
	} else {
		b.WriteString(maxSize)
	}
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("  Examples: 100MB, 1GB, 5GB"))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Enter: continue"))

	return b.String()
}

func renderExtendedTimeoutsParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Extended Timeouts"))
	b.WriteString("\n\n")

	// Field 0: Read timeout
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Read Timeout:"))
	} else {
		b.WriteString(StyleDim.Render("  Read Timeout:"))
	}
	b.WriteString(" ")
	readTimeout := params["read_timeout"]
	if readTimeout == "" {
		b.WriteString(StyleDim.Render("(default: 120s)"))
	} else {
		b.WriteString(readTimeout)
	}
	b.WriteString("\n\n")

	// Field 1: Write timeout
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ Write Timeout:"))
	} else {
		b.WriteString(StyleDim.Render("  Write Timeout:"))
	}
	b.WriteString(" ")
	writeTimeout := params["write_timeout"]
	if writeTimeout == "" {
		b.WriteString(StyleDim.Render("(default: 120s)"))
	} else {
		b.WriteString(writeTimeout)
	}
	b.WriteString("\n\n")

	// Field 2: Dial timeout
	if cursor == 2 {
		b.WriteString(StyleKeybinding.Render("→ Dial Timeout:"))
	} else {
		b.WriteString(StyleDim.Render("  Dial Timeout:"))
	}
	b.WriteString(" ")
	dialTimeout := params["dial_timeout"]
	if dialTimeout == "" {
		b.WriteString(StyleDim.Render("(default: 30s)"))
	} else {
		b.WriteString(dialTimeout)
	}
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("  Examples: 30s, 1m, 5m"))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Enter: continue"))

	return b.String()
}

func renderStaticCachingParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Static Caching"))
	b.WriteString("\n\n")

	// Field 0: Max age
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Cache Max-Age:"))
	} else {
		b.WriteString(StyleDim.Render("  Cache Max-Age:"))
	}
	b.WriteString(" ")
	maxAge := params["max_age"]
	if maxAge == "" {
		b.WriteString(StyleDim.Render("(default: 86400 seconds / 1 day)"))
	} else {
		b.WriteString(maxAge + " seconds")
	}
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("  Common values: 3600 (1 hour), 86400 (1 day), 604800 (1 week)"))
	b.WriteString("\n\n")

	// Field 1: Enable ETag (checkbox)
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	etag := params["enable_etag"]
	if etag == "true" {
		b.WriteString(StyleKeybinding.Render("[✓] Enable ETag"))
	} else {
		b.WriteString(StyleDim.Render("[ ] Enable ETag"))
	}
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("  Adds ETag header for cache validation"))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Space: toggle checkbox", "Enter: continue"))

	return b.String()
}

func renderIPRestrictionParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: IP Restriction"))
	b.WriteString("\n\n")

	// Field 0: LAN Subnet
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ LAN Subnet (CIDR):"))
	} else {
		b.WriteString(StyleDim.Render("  LAN Subnet (CIDR):"))
	}
	b.WriteString(" ")
	lanSubnet := params["lan_subnet"]
	if lanSubnet == "" {
		b.WriteString(StyleDim.Render("(default: 10.0.0.0/8)"))
	} else {
		b.WriteString(lanSubnet)
	}
	b.WriteString("\n")
	if cursor == 0 {
		b.WriteString(StyleDim.Render("  Examples: 10.0.28.0/24, 192.168.1.0/24"))
	}
	b.WriteString("\n\n")

	// Field 1: Allowed External IP
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ Allowed External IP (optional):"))
	} else {
		b.WriteString(StyleDim.Render("  Allowed External IP (optional):"))
	}
	b.WriteString(" ")
	externalIP := params["external_ip"]
	if externalIP == "" {
		b.WriteString(StyleDim.Render("(none)"))
	} else {
		b.WriteString(externalIP)
	}
	b.WriteString("\n")
	if cursor == 1 {
		b.WriteString(StyleDim.Render("  Examples: 203.0.113.5/32, 198.51.100.0/24"))
	}
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Enter: continue"))

	return b.String()
}

func renderSecurityHeadersParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Security Headers"))
	b.WriteString("\n\n")

	b.WriteString(StyleInfo.Render("Select security preset:"))
	b.WriteString("\n\n")

	presets := []struct {
		name        string
		description string
	}{
		{"basic", "Basic security headers (X-Content-Type-Options, X-Frame-Options)"},
		{"strict", "Strict security (+ Referrer-Policy, Permissions-Policy, CSP)"},
		{"paranoid", "Paranoid security (+ HSTS, X-XSS-Protection)"},
	}

	selectedPreset := params["preset"]
	if selectedPreset == "" {
		selectedPreset = "strict"
	}

	for i, preset := range presets {
		if i == cursor {
			b.WriteString(StyleKeybinding.Render("→ "))
		} else {
			b.WriteString("  ")
		}

		if preset.name == selectedPreset {
			b.WriteString(StyleKeybinding.Render("[•] " + preset.name))
		} else {
			b.WriteString(StyleDim.Render("[ ] " + preset.name))
		}
		b.WriteString("\n")
		if i == cursor {
			b.WriteString("  " + StyleInfo.Render(preset.description))
		} else {
			b.WriteString("  " + StyleDim.Render(preset.description))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(RenderNavigationHint("↑/↓: navigate", "Enter: select & continue"))

	return b.String()
}

func renderCompressionAdvancedParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Advanced Compression"))
	b.WriteString("\n\n")

	// Field 0: Compression level
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Compression Level:"))
	} else {
		b.WriteString(StyleDim.Render("  Compression Level:"))
	}
	b.WriteString(" ")
	level := params["compression_level"]
	if level == "" {
		b.WriteString(StyleDim.Render("(default: 5)"))
	} else {
		b.WriteString(level)
	}
	b.WriteString("\n")
	if cursor == 0 {
		b.WriteString(StyleDim.Render("  Range: 1 (fastest) to 9 (best compression)"))
	}
	b.WriteString("\n\n")

	// Field 1-3: Encoders (checkboxes)
	b.WriteString(StyleInfo.Render("Encoders:"))
	b.WriteString("\n")

	gzipEnabled := params["enable_gzip"] != "false"
	zstdEnabled := params["enable_zstd"] != "false"
	brotliEnabled := params["enable_brotli"] == "true"

	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	if gzipEnabled {
		b.WriteString(StyleKeybinding.Render("[✓] gzip"))
	} else {
		b.WriteString(StyleDim.Render("[ ] gzip"))
	}
	b.WriteString("\n")

	if cursor == 2 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	if zstdEnabled {
		b.WriteString(StyleKeybinding.Render("[✓] zstd"))
	} else {
		b.WriteString(StyleDim.Render("[ ] zstd"))
	}
	b.WriteString("\n")

	if cursor == 3 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	if brotliEnabled {
		b.WriteString(StyleKeybinding.Render("[✓] brotli"))
	} else {
		b.WriteString(StyleDim.Render("[ ] brotli"))
	}
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Space: toggle", "Enter: continue"))

	return b.String()
}

func renderHTTPSBackendParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: HTTPS Backend"))
	b.WriteString("\n\n")

	// Field 0: Skip TLS verify (checkbox)
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	skipVerify := params["skip_verify"] == "true"
	if skipVerify {
		b.WriteString(StyleKeybinding.Render("[✓] Skip TLS Verification"))
	} else {
		b.WriteString(StyleDim.Render("[ ] Skip TLS Verification"))
	}
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("  Use for self-signed certificates (insecure)"))
	b.WriteString("\n\n")

	// Field 1: Keepalive connections
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ Keepalive Connections per Host:"))
	} else {
		b.WriteString(StyleDim.Render("  Keepalive Connections per Host:"))
	}
	b.WriteString(" ")
	keepalive := params["keepalive"]
	if keepalive == "" {
		b.WriteString(StyleDim.Render("(default: 100)"))
	} else {
		b.WriteString(keepalive)
	}
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("  Number of idle connections to maintain"))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Space: toggle checkbox", "Enter: continue"))

	return b.String()
}

func renderPerformanceParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Performance"))
	b.WriteString("\n\n")

	b.WriteString(StyleInfo.Render("Select compression encoders:"))
	b.WriteString("\n\n")

	// Checkboxes for encoders
	gzipEnabled := params["enable_gzip"] != "false"
	zstdEnabled := params["enable_zstd"] != "false"

	// Field 0: gzip
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	if gzipEnabled {
		b.WriteString(StyleKeybinding.Render("[✓] gzip"))
	} else {
		b.WriteString(StyleDim.Render("[ ] gzip"))
	}
	b.WriteString("\n")

	// Field 1: zstd
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	if zstdEnabled {
		b.WriteString(StyleKeybinding.Render("[✓] zstd"))
	} else {
		b.WriteString(StyleDim.Render("[ ] zstd"))
	}
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Space: toggle", "Enter: continue"))
	return b.String()
}

func renderAuthHeadersParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Authentication Headers"))
	b.WriteString("\n\n")

	// Field 0: Forward IP (checkbox)
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	forwardIP := params["forward_ip"] == "true"
	if forwardIP {
		b.WriteString(StyleKeybinding.Render("[✓] Forward Client IP (X-Real-IP, X-Forwarded-For)"))
	} else {
		b.WriteString(StyleDim.Render("[ ] Forward Client IP (X-Real-IP, X-Forwarded-For)"))
	}
	b.WriteString("\n\n")

	// Field 1: Forward Proto (checkbox)
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ "))
	} else {
		b.WriteString("  ")
	}
	forwardProto := params["forward_proto"] == "true"
	if forwardProto {
		b.WriteString(StyleKeybinding.Render("[✓] Forward Protocol (X-Forwarded-Proto)"))
	} else {
		b.WriteString(StyleDim.Render("[ ] Forward Protocol (X-Forwarded-Proto)"))
	}
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("  Useful for OAuth/JWT authentication behind a proxy"))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Space: toggle", "Enter: continue"))
	return b.String()
}

func renderWebSocketAdvancedParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: WebSocket Advanced"))
	b.WriteString("\n\n")

	// Field 0: Upgrade timeout
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Upgrade Timeout:"))
	} else {
		b.WriteString(StyleDim.Render("  Upgrade Timeout:"))
	}
	b.WriteString(" ")
	upgradeTimeout := params["upgrade_timeout"]
	if upgradeTimeout == "" {
		b.WriteString(StyleDim.Render("(default: 10s)"))
	} else {
		b.WriteString(upgradeTimeout)
	}
	b.WriteString("\n")
	if cursor == 0 {
		b.WriteString(StyleDim.Render("  Examples: 5s, 10s, 30s"))
	}
	b.WriteString("\n\n")

	// Field 1: Ping interval
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ Ping Interval:"))
	} else {
		b.WriteString(StyleDim.Render("  Ping Interval:"))
	}
	b.WriteString(" ")
	pingInterval := params["ping_interval"]
	if pingInterval == "" {
		b.WriteString(StyleDim.Render("(default: 30s)"))
	} else {
		b.WriteString(pingInterval)
	}
	b.WriteString("\n")
	if cursor == 1 {
		b.WriteString(StyleDim.Render("  Keepalive ping frequency. Examples: 15s, 30s, 60s"))
	}
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Enter: continue"))
	return b.String()
}

func renderCustomHeadersInjectParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Custom Headers Inject"))
	b.WriteString("\n\n")

	b.WriteString(StyleInfo.Render("Header injection direction:"))
	b.WriteString("\n\n")

	directions := []struct {
		name        string
		description string
	}{
		{"upstream", "Add headers to upstream requests (to backend)"},
		{"response", "Add headers to client responses"},
		{"both", "Add headers to both upstream and downstream"},
	}

	selectedDirection := params["direction"]
	if selectedDirection == "" {
		selectedDirection = "upstream"
	}

	for i, dir := range directions {
		if i == cursor {
			b.WriteString(StyleKeybinding.Render("→ "))
		} else {
			b.WriteString("  ")
		}

		if dir.name == selectedDirection {
			b.WriteString(StyleKeybinding.Render("[•] " + dir.name))
		} else {
			b.WriteString(StyleDim.Render("[ ] " + dir.name))
		}
		b.WriteString("\n")
		if i == cursor {
			b.WriteString("  " + StyleInfo.Render(dir.description))
		} else {
			b.WriteString("  " + StyleDim.Render(dir.description))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(StyleDim.Render("Note: Configure custom headers in the next step"))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("↑/↓: navigate", "Enter: select & continue"))
	return b.String()
}

func renderRewriteRulesParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Rewrite Rules"))
	b.WriteString("\n\n")

	// Field 0: Path pattern
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Path Pattern:"))
	} else {
		b.WriteString(StyleDim.Render("  Path Pattern:"))
	}
	b.WriteString(" ")
	pathPattern := params["path_pattern"]
	if pathPattern == "" {
		b.WriteString(StyleDim.Render("(default: /api/*)"))
	} else {
		b.WriteString(pathPattern)
	}
	b.WriteString("\n")
	if cursor == 0 {
		b.WriteString(StyleDim.Render("  Examples: /api/*, /old-path/*, /v1/*"))
	}
	b.WriteString("\n\n")

	// Field 1: Rewrite to
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ Rewrite To:"))
	} else {
		b.WriteString(StyleDim.Render("  Rewrite To:"))
	}
	b.WriteString(" ")
	rewriteTo := params["rewrite_to"]
	if rewriteTo == "" {
		b.WriteString(StyleDim.Render("(default: /new-api{uri})"))
	} else {
		b.WriteString(rewriteTo)
	}
	b.WriteString("\n")
	if cursor == 1 {
		b.WriteString(StyleDim.Render("  Use {uri} for original path. Examples: /new-api{uri}, /v2{uri}"))
	}
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Tab: next field", "Enter: continue"))
	return b.String()
}

func renderFrameEmbeddingParams(params map[string]string, cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Configure: Frame Embedding"))
	b.WriteString("\n\n")

	// Field 0: Allowed origins
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Allowed Origins:"))
	} else {
		b.WriteString(StyleDim.Render("  Allowed Origins:"))
	}
	b.WriteString(" ")
	allowedOrigins := params["allowed_origins"]
	if allowedOrigins == "" {
		b.WriteString(StyleDim.Render("(default: 'self')"))
	} else {
		b.WriteString(allowedOrigins)
	}
	b.WriteString("\n")
	if cursor == 0 {
		b.WriteString(StyleDim.Render("  Examples: 'self', https://trusted.com, 'none'"))
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("  Multiple origins: space-separated (e.g., 'self' https://app1.com)"))
	}
	b.WriteString("\n\n")

	b.WriteString(StyleInfo.Render("Sets Content-Security-Policy frame-ancestors directive"))
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("Controls which sites can embed this page in iframes"))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Enter: continue"))
	return b.String()
}
