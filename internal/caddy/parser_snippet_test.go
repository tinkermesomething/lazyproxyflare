package caddy

import (
	"strings"
	"testing"
)

func TestParseSnippetBlock(t *testing.T) {
	caddyfile := `(ip_restricted) {
	@external {
		not remote_ip 10.0.0.0/8
	}
	respond @external 403
}`

	lines := strings.Split(caddyfile, "\n")
	snippet, endLine := parseSnippetBlock(lines, 0)

	if snippet == nil {
		t.Fatal("Expected snippet to be parsed, got nil")
	}

	// Verify snippet name
	if snippet.Name != "ip_restricted" {
		t.Errorf("Expected name 'ip_restricted', got '%s'", snippet.Name)
	}

	// Verify line numbers (1-indexed)
	if snippet.LineStart != 1 {
		t.Errorf("Expected LineStart=1, got %d", snippet.LineStart)
	}
	// LineEnd includes the closing brace line
	if snippet.LineEnd != 6 {
		t.Errorf("Expected LineEnd=6, got %d", snippet.LineEnd)
	}
	if endLine != 5 {
		t.Errorf("Expected endLine=5 (0-indexed), got %d", endLine)
	}

	// Verify content (should not include wrapper)
	expectedContent := `	@external {
		not remote_ip 10.0.0.0/8
	}
	respond @external 403`

	if snippet.Content != expectedContent {
		t.Errorf("Content mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedContent, snippet.Content)
	}

	// Verify auto-categorization
	if snippet.Category != SnippetIPRestriction {
		t.Errorf("Expected category IPRestriction, got %v", snippet.Category)
	}

	if !snippet.AutoDetected {
		t.Error("Expected AutoDetected=true")
	}

	if snippet.Confidence < 0.6 {
		t.Errorf("Expected reasonable confidence (>0.6), got %.2f", snippet.Confidence)
	}

	// Verify description was generated
	if snippet.Description == "" {
		t.Error("Expected description to be generated")
	}
}

func TestParseCaddyfileWithSnippets_MultipleSnippets(t *testing.T) {
	caddyfile := `{
	email admin@example.com
}

(ip_restricted) {
	@external {
		not remote_ip 10.0.0.0/8
	}
	respond @external 403
}

(security_headers) {
	header {
		X-Frame-Options "SAMEORIGIN"
		X-Content-Type-Options "nosniff"
		Strict-Transport-Security "max-age=31536000"
	}
}

example.com {
	reverse_proxy http://10.0.28.5:8080
	import ip_restricted
	import security_headers
}
`

	parsed := ParseCaddyfileWithSnippets(caddyfile)

	// Verify we found 2 snippets
	if len(parsed.Snippets) != 2 {
		t.Errorf("Expected 2 snippets, got %d", len(parsed.Snippets))
	}

	// Verify first snippet
	if parsed.Snippets[0].Name != "ip_restricted" {
		t.Errorf("Expected first snippet 'ip_restricted', got '%s'", parsed.Snippets[0].Name)
	}
	if parsed.Snippets[0].Category != SnippetIPRestriction {
		t.Errorf("Expected first snippet category IPRestriction, got %v", parsed.Snippets[0].Category)
	}

	// Verify second snippet
	if parsed.Snippets[1].Name != "security_headers" {
		t.Errorf("Expected second snippet 'security_headers', got '%s'", parsed.Snippets[1].Name)
	}
	if parsed.Snippets[1].Category != SnippetSecurityHeaders {
		t.Errorf("Expected second snippet category SecurityHeaders, got %v", parsed.Snippets[1].Category)
	}

	// Verify we still found the domain entry
	if len(parsed.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(parsed.Entries))
	}
	if parsed.Entries[0].Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%s'", parsed.Entries[0].Domain)
	}

	// Verify the entry has imports tracked
	if len(parsed.Entries[0].Imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(parsed.Entries[0].Imports))
	}
}

func TestParseCaddyfileWithSnippets_NoSnippets(t *testing.T) {
	caddyfile := `example.com {
	reverse_proxy http://10.0.28.5:8080
}
`

	parsed := ParseCaddyfileWithSnippets(caddyfile)

	if len(parsed.Snippets) != 0 {
		t.Errorf("Expected 0 snippets, got %d", len(parsed.Snippets))
	}

	if len(parsed.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(parsed.Entries))
	}
}

func TestParseCaddyfileWithSnippets_NestedBraces(t *testing.T) {
	caddyfile := `(complex_snippet) {
	@matcher {
		path /api/*
		header Content-Type application/json
	}
	reverse_proxy @matcher {
		to backend1:8080
		to backend2:8080
		health_uri /health
	}
}
`

	parsed := ParseCaddyfileWithSnippets(caddyfile)

	if len(parsed.Snippets) != 1 {
		t.Fatalf("Expected 1 snippet, got %d", len(parsed.Snippets))
	}

	snippet := parsed.Snippets[0]

	// Verify all content was captured (should handle nested braces)
	if !strings.Contains(snippet.Content, "path /api/*") {
		t.Error("Expected content to contain 'path /api/*'")
	}
	if !strings.Contains(snippet.Content, "health_uri /health") {
		t.Error("Expected content to contain 'health_uri /health'")
	}

	// Verify the closing brace count is correct
	openBraces := strings.Count(snippet.Content, "{")
	closeBraces := strings.Count(snippet.Content, "}")
	if openBraces != closeBraces {
		t.Errorf("Brace mismatch: %d open, %d close", openBraces, closeBraces)
	}
}

func TestParseCaddyfileWithSnippets_EmptySnippet(t *testing.T) {
	caddyfile := `(empty) {
}
`

	parsed := ParseCaddyfileWithSnippets(caddyfile)

	if len(parsed.Snippets) != 1 {
		t.Fatalf("Expected 1 snippet, got %d", len(parsed.Snippets))
	}

	snippet := parsed.Snippets[0]

	if snippet.Name != "empty" {
		t.Errorf("Expected name 'empty', got '%s'", snippet.Name)
	}

	// Content should be empty or just whitespace
	if strings.TrimSpace(snippet.Content) != "" {
		t.Errorf("Expected empty content, got: '%s'", snippet.Content)
	}
}

func TestParseCaddyfileWithSnippets_RealWorldExample(t *testing.T) {
	caddyfile := `{
	email admin@example.com
}

(ip_restricted) {
	@external {
		not remote_ip 10.0.28.0/24
		not remote_ip 166.1.123.74/32
	}
	respond @external "Access Denied" 403
}

(oauth_headers) {
	header_up X-Real-IP {remote_host}
	header_up X-Forwarded-For {remote_host}
	header_up X-Forwarded-Proto {scheme}
}

(websocket_headers) {
	header_up Upgrade {http.request.header.Upgrade}
	header_up Connection {http.request.header.Connection}
}

# === sonobarr.angelsomething.com ===
sonobarr.angelsomething.com {
	reverse_proxy http://10.0.28.9:8989 {
		header_up X-Real-IP {remote_host}
		header_up X-Forwarded-For {remote_host}
		header_up X-Forwarded-Proto {scheme}
	}
	import ip_restricted
}

# === plex.angelsomething.com ===
plex.angelsomething.com {
	reverse_proxy http://10.0.28.9:32400
	import ip_restricted
}
`

	parsed := ParseCaddyfileWithSnippets(caddyfile)

	// Verify snippets
	if len(parsed.Snippets) != 3 {
		t.Errorf("Expected 3 snippets, got %d", len(parsed.Snippets))
	}

	snippetNames := make(map[string]bool)
	for _, s := range parsed.Snippets {
		snippetNames[s.Name] = true
	}

	expectedSnippets := []string{"ip_restricted", "oauth_headers", "websocket_headers"}
	for _, name := range expectedSnippets {
		if !snippetNames[name] {
			t.Errorf("Expected to find snippet '%s'", name)
		}
	}

	// Verify entries
	if len(parsed.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(parsed.Entries))
	}

	// Verify first entry
	if parsed.Entries[0].Domain != "sonobarr.angelsomething.com" {
		t.Errorf("Expected first domain 'sonobarr.angelsomething.com', got '%s'", parsed.Entries[0].Domain)
	}
	if parsed.Entries[0].Target != "10.0.28.9" {
		t.Errorf("Expected target '10.0.28.9', got '%s'", parsed.Entries[0].Target)
	}
	if parsed.Entries[0].Port != 8989 {
		t.Errorf("Expected port 8989, got %d", parsed.Entries[0].Port)
	}
	if !parsed.Entries[0].HasMarker {
		t.Error("Expected first entry to have marker comment")
	}

	// Verify second entry
	if parsed.Entries[1].Domain != "plex.angelsomething.com" {
		t.Errorf("Expected second domain 'plex.angelsomething.com', got '%s'", parsed.Entries[1].Domain)
	}
	if !parsed.Entries[1].HasMarker {
		t.Error("Expected second entry to have marker comment")
	}

	// Verify categorization accuracy
	for _, snippet := range parsed.Snippets {
		switch snippet.Name {
		case "ip_restricted":
			if snippet.Category != SnippetIPRestriction {
				t.Errorf("ip_restricted: expected IPRestriction, got %v", snippet.Category)
			}
		case "oauth_headers":
			if snippet.Category != SnippetOAuthHeaders {
				t.Errorf("oauth_headers: expected OAuthHeaders, got %v", snippet.Category)
			}
		case "websocket_headers":
			if snippet.Category != SnippetWebSocketHeaders {
				t.Errorf("websocket_headers: expected WebSocketHeaders, got %v", snippet.Category)
			}
		}
	}
}

func TestCountBracesOutsideQuotes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOpen  int
		wantClose int
	}{
		{"simple open", "foo {", 1, 0},
		{"simple close", "}", 0, 1},
		{"both", "{ foo }", 1, 1},
		{"quoted braces ignored", `header X-Custom "{json}"`, 0, 0},
		{"mixed", `@test { header "{value}" }`, 1, 1},
		{"escaped quote", `header "test \"nested\" value"`, 0, 0},
		{"single quotes", `header 'value {with} braces'`, 0, 0},
		{"no braces", "reverse_proxy localhost:8080", 0, 0},
		{"nested real braces", "@matcher { not { remote_ip } }", 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpen, gotClose := countBracesOutsideQuotes(tt.input)
			if gotOpen != tt.wantOpen || gotClose != tt.wantClose {
				t.Errorf("countBracesOutsideQuotes(%q) = (%d, %d), want (%d, %d)",
					tt.input, gotOpen, gotClose, tt.wantOpen, tt.wantClose)
			}
		})
	}
}

func TestParseCaddyfileWithQuotedBraces(t *testing.T) {
	// This Caddyfile has braces inside quoted strings that should NOT affect parsing
	caddyfile := `example.com {
	header X-Custom "{json-value}"
	header X-Another "test { with } braces"
	reverse_proxy localhost:8080
}`

	parsed := ParseCaddyfileWithSnippets(caddyfile)

	if len(parsed.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(parsed.Entries))
	}

	entry := parsed.Entries[0]
	if entry.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%s'", entry.Domain)
	}
	if entry.LineEnd != 5 {
		t.Errorf("Expected LineEnd=5, got %d (parsing may have been confused by quoted braces)", entry.LineEnd)
	}
}
