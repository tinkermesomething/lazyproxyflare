package snippet_wizard

import (
	"fmt"
	"strings"
)

// GenerateIPRestrictionSnippet generates the IP restriction snippet content
func GenerateIPRestrictionSnippet(lanSubnet, externalIP string) string {
	var b strings.Builder

	b.WriteString("(ip_restricted) {\n")
	b.WriteString("\t@external {\n")

	// Build the "not" condition
	if lanSubnet != "" && externalIP != "" {
		b.WriteString(fmt.Sprintf("\t\tnot remote_ip %s %s\n", lanSubnet, externalIP))
	} else if lanSubnet != "" {
		b.WriteString(fmt.Sprintf("\t\tnot remote_ip %s\n", lanSubnet))
	} else {
		b.WriteString("\t\tnot remote_ip 10.0.0.0/8\n") // fallback
	}

	b.WriteString("\t}\n")
	b.WriteString("\trespond @external 404\n")
	b.WriteString("}")

	return b.String()
}

// GenerateSecurityHeadersSnippet generates the security headers snippet content
func GenerateSecurityHeadersSnippet(preset string) string {
	var b strings.Builder

	b.WriteString("(security_headers) {\n")
	b.WriteString("\theader {\n")
	b.WriteString("\t\t-Server\n")
	b.WriteString("\t\tX-Content-Type-Options \"nosniff\"\n")
	b.WriteString("\t\tX-Frame-Options \"SAMEORIGIN\"\n")

	if preset == "strict" || preset == "paranoid" {
		b.WriteString("\t\tReferrer-Policy \"strict-origin-when-cross-origin\"\n")
		b.WriteString("\t\tPermissions-Policy \"geolocation=(), microphone=(), camera=()\"\n")
		b.WriteString("\t\tContent-Security-Policy \"default-src 'self'\"\n")
	}

	if preset == "paranoid" {
		b.WriteString("\t\tStrict-Transport-Security \"max-age=63072000; includeSubDomains; preload\"\n")
		b.WriteString("\t\tX-XSS-Protection \"1; mode=block\"\n")
	}

	b.WriteString("\t}\n")
	b.WriteString("}")

	return b.String()
}

// GeneratePerformanceSnippet generates the performance optimization snippet content
func GeneratePerformanceSnippet() string {
	var b strings.Builder

	b.WriteString("(performance) {\n")
	b.WriteString("\tencode gzip zstd\n")
	b.WriteString("}")

	return b.String()
}

// Phase 3+: New snippet templates (stubs for now, will implement in Phase 3-4)

// GenerateCORSHeadersSnippet generates CORS configuration snippet
func GenerateCORSHeadersSnippet(allowedOrigins, allowedMethods string, allowCredentials bool) string {
	var b strings.Builder
	b.WriteString("(cors_headers) {\n")
	b.WriteString("\t@cors_preflight {\n")
	b.WriteString("\t\tmethod OPTIONS\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\theader {\n")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}
	b.WriteString(fmt.Sprintf("\t\tAccess-Control-Allow-Origin %s\n", allowedOrigins))
	if allowedMethods == "" {
		allowedMethods = "GET, POST, PUT, DELETE, OPTIONS"
	}
	b.WriteString(fmt.Sprintf("\t\tAccess-Control-Allow-Methods \"%s\"\n", allowedMethods))
	b.WriteString("\t\tAccess-Control-Allow-Headers \"Content-Type, Authorization\"\n")
	if allowCredentials {
		b.WriteString("\t\tAccess-Control-Allow-Credentials \"true\"\n")
	}
	b.WriteString("\t\tAccess-Control-Max-Age \"3600\"\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\trespond @cors_preflight 204\n")
	b.WriteString("}")
	return b.String()
}

// GenerateRateLimitingSnippet generates rate limiting snippet using Caddy rate limit module
func GenerateRateLimitingSnippet(requestsPerSecond int, burstSize int, zone string) string {
	var b strings.Builder
	if requestsPerSecond == 0 {
		requestsPerSecond = 100
	}
	if burstSize == 0 {
		burstSize = 50
	}
	if zone == "" {
		zone = "rate_limit_zone"
	}

	b.WriteString("(rate_limiting) {\n")
	b.WriteString("\t@ratelimit {\n")
	b.WriteString(fmt.Sprintf("\t\tnot remote_ip forwarded %d/s %d burst\n", requestsPerSecond, burstSize))
	b.WriteString("\t}\n")
	b.WriteString("\trespond @ratelimit \"Too Many Requests\" 429\n")
	b.WriteString("}")
	return b.String()
}

// GenerateAuthHeadersSnippet generates OAuth/JWT header forwarding snippet
func GenerateAuthHeadersSnippet(forwardIP, forwardProto bool, customHeaders map[string]string) string {
	var b strings.Builder
	b.WriteString("(auth_headers) {\n")
	if forwardIP {
		b.WriteString("\theader_up X-Real-IP {remote_host}\n")
		b.WriteString("\theader_up X-Forwarded-For {remote_host}\n")
	}
	if forwardProto {
		b.WriteString("\theader_up X-Forwarded-Proto {scheme}\n")
	}
	for key, value := range customHeaders {
		b.WriteString(fmt.Sprintf("\theader_up %s %s\n", key, value))
	}
	b.WriteString("}")
	return b.String()
}

// GenerateStaticCachingSnippet generates static caching snippet
func GenerateStaticCachingSnippet(maxAge int, fileTypes []string, enableETag bool) string {
	var b strings.Builder
	if maxAge == 0 {
		maxAge = 86400 // 1 day default
	}

	b.WriteString("(static_caching) {\n")
	b.WriteString("\t@static {\n")
	if len(fileTypes) == 0 {
		b.WriteString("\t\tpath *.css *.js *.jpg *.jpeg *.png *.gif *.ico *.svg *.woff *.woff2 *.ttf *.eot\n")
	} else {
		b.WriteString("\t\tpath")
		for _, ft := range fileTypes {
			b.WriteString(" *." + ft)
		}
		b.WriteString("\n")
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\theader @static {\n")
	b.WriteString(fmt.Sprintf("\t\tCache-Control \"public, max-age=%d\"\n", maxAge))
	if enableETag {
		b.WriteString("\t\t+ETag\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}")
	return b.String()
}

// GenerateCompressionAdvancedSnippet generates advanced compression snippet
func GenerateCompressionAdvancedSnippet(encoders []string, level int, mimeTypes []string) string {
	var b strings.Builder
	if len(encoders) == 0 {
		encoders = []string{"gzip", "zstd"}
	}
	if level == 0 {
		level = 5
	}

	b.WriteString("(compression_advanced) {\n")
	b.WriteString("\tencode ")
	b.WriteString(strings.Join(encoders, " "))
	b.WriteString(fmt.Sprintf(" %d", level))
	if len(mimeTypes) > 0 {
		b.WriteString(" {\n")
		b.WriteString("\t\tmatch {\n")
		for _, mt := range mimeTypes {
			b.WriteString(fmt.Sprintf("\t\t\theader Content-Type %s*\n", mt))
		}
		b.WriteString("\t\t}\n")
		b.WriteString("\t}")
	}
	b.WriteString("\n}")
	return b.String()
}

// GenerateWebSocketAdvancedSnippet generates WebSocket support snippet
func GenerateWebSocketAdvancedSnippet(upgradeTimeout int, pingInterval int) string {
	var b strings.Builder
	b.WriteString("(websocket_advanced) {\n")
	b.WriteString("\t@websockets {\n")
	b.WriteString("\t\theader Connection *Upgrade*\n")
	b.WriteString("\t\theader Upgrade websocket\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\theader_up @websockets Connection {http.request.header.Connection}\n")
	b.WriteString("\theader_up @websockets Upgrade {http.request.header.Upgrade}\n")
	b.WriteString("\theader_up @websockets X-Real-IP {remote_host}\n")
	b.WriteString("\n\ttransport http {\n")
	if upgradeTimeout > 0 {
		b.WriteString(fmt.Sprintf("\t\tread_timeout %ds\n", upgradeTimeout))
		b.WriteString(fmt.Sprintf("\t\twrite_timeout %ds\n", upgradeTimeout))
	} else {
		b.WriteString("\t\tread_timeout 0\n")
		b.WriteString("\t\twrite_timeout 0\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}")
	return b.String()
}

// GenerateExtendedTimeoutsSnippet generates extended timeouts snippet
func GenerateExtendedTimeoutsSnippet(readTimeout, writeTimeout, dialTimeout string) string {
	var b strings.Builder
	b.WriteString("(extended_timeouts) {\n")
	b.WriteString("\ttransport http {\n")
	if readTimeout != "" {
		b.WriteString(fmt.Sprintf("\t\tread_timeout %s\n", readTimeout))
	}
	if writeTimeout != "" {
		b.WriteString(fmt.Sprintf("\t\twrite_timeout %s\n", writeTimeout))
	}
	if dialTimeout != "" {
		b.WriteString(fmt.Sprintf("\t\tdial_timeout %s\n", dialTimeout))
	}
	b.WriteString("\t}\n")
	b.WriteString("}")
	return b.String()
}

// GenerateLargeUploadsSnippet generates large uploads snippet
func GenerateLargeUploadsSnippet(maxSize string) string {
	var b strings.Builder
	b.WriteString("(large_uploads) {\n")
	b.WriteString("\trequest_body {\n")
	b.WriteString(fmt.Sprintf("\t\tmax_size %s\n", maxSize))
	b.WriteString("\t}\n")
	b.WriteString("}")
	return b.String()
}

// GenerateCustomHeadersInjectSnippet generates custom headers injection snippet
func GenerateCustomHeadersInjectSnippet(headers map[string]string, direction string) string {
	var b strings.Builder
	b.WriteString("(custom_headers_inject) {\n")

	if direction == "" || direction == "both" {
		// Default: inject both upstream and response headers
		if len(headers) > 0 {
			b.WriteString("\theader {\n")
			for key, value := range headers {
				b.WriteString(fmt.Sprintf("\t\t%s \"%s\"\n", key, value))
			}
			b.WriteString("\t}\n")
		}
	} else if direction == "upstream" {
		// Inject headers to upstream backend
		for key, value := range headers {
			b.WriteString(fmt.Sprintf("\theader_up %s \"%s\"\n", key, value))
		}
	} else if direction == "response" {
		// Inject headers to client response
		b.WriteString("\theader {\n")
		for key, value := range headers {
			b.WriteString(fmt.Sprintf("\t\t%s \"%s\"\n", key, value))
		}
		b.WriteString("\t}\n")
	}

	b.WriteString("}")
	return b.String()
}

// GenerateRewriteRulesSnippet generates URL rewriting snippet
func GenerateRewriteRulesSnippet(pathPattern, rewriteTo string) string {
	var b strings.Builder
	b.WriteString("(rewrite_rules) {\n")

	if pathPattern == "" {
		pathPattern = "/old-path/*"
	}
	if rewriteTo == "" {
		rewriteTo = "/new-path/{http.request.uri.path}"
	}

	// Use Caddy matcher for pattern-based rewriting
	b.WriteString("\t@rewrite {\n")
	b.WriteString(fmt.Sprintf("\t\tpath %s\n", pathPattern))
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\trewrite @rewrite %s\n", rewriteTo))

	b.WriteString("}")
	return b.String()
}

// GenerateFrameEmbeddingSnippet generates frame embedding snippet
func GenerateFrameEmbeddingSnippet(allowedOrigins string) string {
	var b strings.Builder
	b.WriteString("(frame_embedding) {\n")
	b.WriteString("\theader {\n")
	if allowedOrigins != "" {
		b.WriteString(fmt.Sprintf("\t\tContent-Security-Policy \"frame-ancestors %s\"\n", allowedOrigins))
	}
	b.WriteString("\t}\n")
	b.WriteString("}")
	return b.String()
}

// GenerateHTTPSBackendSnippet generates HTTPS backend snippet
func GenerateHTTPSBackendSnippet(skipVerify bool, keepalive int) string {
	var b strings.Builder
	b.WriteString("(https_backend) {\n")
	b.WriteString("\ttransport http {\n")
	if skipVerify {
		b.WriteString("\t\ttls_insecure_skip_verify\n")
	}
	if keepalive > 0 {
		b.WriteString(fmt.Sprintf("\t\tkeepalive_idle_conns_per_host %d\n", keepalive))
	}
	b.WriteString("\t}\n")
	b.WriteString("}")
	return b.String()
}
