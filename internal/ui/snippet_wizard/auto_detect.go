package snippet_wizard

import (
	"os"
	"regexp"
	"strings"
)

// DetectPatterns analyzes a Caddyfile and detects unique patterns
func DetectPatterns(caddyfileContent string) []DetectedPattern {
	var patterns []DetectedPattern

	// Detect request_body max_size patterns
	if p := detectRequestBodyLimit(caddyfileContent); p != nil {
		patterns = append(patterns, *p)
	}

	// Detect extended timeout patterns
	if p := detectExtendedTimeouts(caddyfileContent); p != nil {
		patterns = append(patterns, *p)
	}

	// Detect OAuth header forwarding patterns
	if p := detectOAuthForwarding(caddyfileContent); p != nil {
		patterns = append(patterns, *p)
	}

	// Detect flush_interval patterns
	if p := detectFlushInterval(caddyfileContent); p != nil {
		patterns = append(patterns, *p)
	}

	// Detect custom keepalive patterns
	if p := detectKeepaliveCustom(caddyfileContent); p != nil {
		patterns = append(patterns, *p)
	}

	return patterns
}

// detectRequestBodyLimit finds request_body max_size directives
func detectRequestBodyLimit(content string) *DetectedPattern {
	re := regexp.MustCompile(`(?m)^\s*request_body\s*\{[^}]*max_size\s+(\d+(?:MB|GB|KB))[^}]*\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return nil
	}

	// Extract example domain
	exampleDomain := extractDomainNearPattern(content, matches[0][0])

	return &DetectedPattern{
		Type:          PatternRequestBodyLimit,
		Count:         len(matches),
		ExampleDomain: exampleDomain,
		RawDirective:  strings.TrimSpace(matches[0][0]),
		SuggestedName: "large_uploads",
		Parameters: map[string]string{
			"max_size": matches[0][1],
		},
	}
}

// detectExtendedTimeouts finds custom timeout configurations
func detectExtendedTimeouts(content string) *DetectedPattern {
	// Look for transport blocks with read_timeout or write_timeout
	re := regexp.MustCompile(`(?s)transport\s+http\s*\{[^}]*(?:read_timeout|write_timeout)[^}]*\}`)
	matches := re.FindAllString(content, -1)

	if len(matches) == 0 {
		return nil
	}

	// Extract timeout values from first match
	readTimeoutRe := regexp.MustCompile(`read_timeout\s+(\d+[smh])`)
	writeTimeoutRe := regexp.MustCompile(`write_timeout\s+(\d+[smh])`)

	params := make(map[string]string)
	if rtMatch := readTimeoutRe.FindStringSubmatch(matches[0]); len(rtMatch) > 1 {
		params["read_timeout"] = rtMatch[1]
	}
	if wtMatch := writeTimeoutRe.FindStringSubmatch(matches[0]); len(wtMatch) > 1 {
		params["write_timeout"] = wtMatch[1]
	}

	// Only return if we found custom timeouts (not default)
	if len(params) == 0 {
		return nil
	}

	exampleDomain := extractDomainNearPattern(content, matches[0])

	return &DetectedPattern{
		Type:          PatternExtendedTimeouts,
		Count:         len(matches),
		ExampleDomain: exampleDomain,
		RawDirective:  strings.TrimSpace(matches[0]),
		SuggestedName: "extended_timeouts",
		Parameters:    params,
	}
}

// detectOAuthForwarding finds header_up X-Real-IP patterns
func detectOAuthForwarding(content string) *DetectedPattern {
	re := regexp.MustCompile(`(?m)^\s*header_up\s+X-Real-IP\s+\{remote_host\}`)
	matches := re.FindAllString(content, -1)

	if len(matches) == 0 {
		return nil
	}

	exampleDomain := extractDomainNearPattern(content, matches[0])

	return &DetectedPattern{
		Type:          PatternOAuthForwarding,
		Count:         len(matches),
		ExampleDomain: exampleDomain,
		RawDirective:  strings.TrimSpace(matches[0]),
		SuggestedName: "oauth_forwarding",
		Parameters: map[string]string{
			"forward_real_ip": "true",
		},
	}
}

// detectFlushInterval finds flush_interval -1 patterns
func detectFlushInterval(content string) *DetectedPattern {
	re := regexp.MustCompile(`(?m)^\s*flush_interval\s+(-?\d+[smh]?)`)
	matches := re.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return nil
	}

	exampleDomain := extractDomainNearPattern(content, matches[0][0])

	return &DetectedPattern{
		Type:          PatternCustomTransport,
		Count:         len(matches),
		ExampleDomain: exampleDomain,
		RawDirective:  strings.TrimSpace(matches[0][0]),
		SuggestedName: "streaming_support",
		Parameters: map[string]string{
			"flush_interval": matches[0][1],
		},
	}
}

// detectKeepaliveCustom finds custom keepalive_idle_conns_per_host values
func detectKeepaliveCustom(content string) *DetectedPattern {
	re := regexp.MustCompile(`(?m)^\s*keepalive_idle_conns_per_host\s+(\d+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	if len(matches) == 0 {
		return nil
	}

	exampleDomain := extractDomainNearPattern(content, matches[0][0])

	return &DetectedPattern{
		Type:          PatternKeepaliveCustom,
		Count:         len(matches),
		ExampleDomain: exampleDomain,
		RawDirective:  strings.TrimSpace(matches[0][0]),
		SuggestedName: "custom_keepalive",
		Parameters: map[string]string{
			"keepalive_idle_conns_per_host": matches[0][1],
		},
	}
}

// extractDomainNearPattern finds the domain name closest to the pattern
func extractDomainNearPattern(content, pattern string) string {
	// Find position of pattern
	idx := strings.Index(content, pattern)
	if idx == -1 {
		return "unknown"
	}

	// Look backward for domain definition (*.angelsomething.com {)
	beforePattern := content[:idx]
	lines := strings.Split(beforePattern, "\n")

	// Search backward for domain line
	domainRe := regexp.MustCompile(`^([a-zA-Z0-9.-]+\.angelsomething\.com)(?:,\s*[a-zA-Z0-9.-]+)*\s*\{`)
	for i := len(lines) - 1; i >= 0; i-- {
		if match := domainRe.FindStringSubmatch(lines[i]); len(match) > 1 {
			return match[1]
		}
	}

	return "unknown"
}

// CalculateConfidence returns a confidence score (0-100) for a detected pattern
func CalculateConfidence(pattern DetectedPattern) int {
	score := 50 // Base score

	// More occurrences = higher confidence
	if pattern.Count >= 2 {
		score += 20
	}
	if pattern.Count >= 3 {
		score += 10
	}

	// Known domain = higher confidence
	if pattern.ExampleDomain != "unknown" {
		score += 10
	}

	// Has parameters = higher confidence
	if len(pattern.Parameters) > 0 {
		score += 10
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// GenerateSnippetFromPattern generates snippet content from a detected pattern
func GenerateSnippetFromPattern(pattern DetectedPattern) string {
	switch pattern.Type {
	case PatternRequestBodyLimit:
		maxSize := pattern.Parameters["max_size"]
		if maxSize == "" {
			maxSize = "512MB"
		}
		return GenerateLargeUploadsSnippet(maxSize)

	case PatternExtendedTimeouts:
		readTimeout := pattern.Parameters["read_timeout"]
		writeTimeout := pattern.Parameters["write_timeout"]
		dialTimeout := pattern.Parameters["dial_timeout"]
		if readTimeout == "" {
			readTimeout = "120s"
		}
		if writeTimeout == "" {
			writeTimeout = "120s"
		}
		return GenerateExtendedTimeoutsSnippet(readTimeout, writeTimeout, dialTimeout)

	case PatternOAuthForwarding:
		return GenerateAuthHeadersSnippet(true, true, nil)

	case PatternCustomTransport:
		// Streaming support with flush_interval
		flushInterval := pattern.Parameters["flush_interval"]
		return "(" + pattern.SuggestedName + ") {\n\tflush_interval " + flushInterval + "\n}"

	case PatternKeepaliveCustom:
		conns := pattern.Parameters["keepalive_idle_conns_per_host"]
		return "(" + pattern.SuggestedName + ") {\n\ttransport http {\n\t\tkeepalive_idle_conns_per_host " + conns + "\n\t}\n}"

	default:
		return ""
	}
}

// DetectPatternsFromFile reads a Caddyfile and detects patterns
func DetectPatternsFromFile(caddyfilePath string) ([]DetectedPattern, error) {
	// Read Caddyfile
	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		return nil, err
	}

	// Detect patterns
	patterns := DetectPatterns(string(content))

	return patterns, nil
}
