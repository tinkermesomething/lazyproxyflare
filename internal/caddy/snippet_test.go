package caddy

import (
	"strings"
	"testing"
)

func TestCategorizeSnippet_IPRestriction(t *testing.T) {
	content := `@external {
		not remote_ip 10.0.0.0/8
	}
	respond @external 403`

	category, confidence := CategorizeSnippet(content)

	if category != SnippetIPRestriction {
		t.Errorf("Expected category %v, got %v", SnippetIPRestriction, category)
	}

	if confidence < 0.6 {
		t.Errorf("Expected reasonable confidence (>0.6), got %.2f", confidence)
	}
}

func TestCategorizeSnippet_SecurityHeaders(t *testing.T) {
	content := `header {
		X-Frame-Options "SAMEORIGIN"
		X-Content-Type-Options "nosniff"
		Strict-Transport-Security "max-age=31536000"
	}`

	category, confidence := CategorizeSnippet(content)

	if category != SnippetSecurityHeaders {
		t.Errorf("Expected category %v, got %v", SnippetSecurityHeaders, category)
	}

	if confidence < 0.8 {
		t.Errorf("Expected high confidence (>0.8), got %.2f", confidence)
	}
}

func TestCategorizeSnippet_WebSocketHeaders(t *testing.T) {
	content := `reverse_proxy {
		header_up Upgrade {http.request.header.Upgrade}
		header_up Connection {http.request.header.Connection}
	}`

	category, confidence := CategorizeSnippet(content)

	if category != SnippetWebSocketHeaders {
		t.Errorf("Expected category %v, got %v", SnippetWebSocketHeaders, category)
	}

	if confidence < 0.6 {
		t.Errorf("Expected reasonable confidence (>0.6), got %.2f", confidence)
	}
}

func TestCategorizeSnippet_OAuthHeaders(t *testing.T) {
	content := `reverse_proxy {
		header_up X-Real-IP {remote_host}
		header_up X-Forwarded-For {remote_host}
		header_up X-Forwarded-Proto {scheme}
	}`

	category, confidence := CategorizeSnippet(content)

	if category != SnippetOAuthHeaders {
		t.Errorf("Expected category %v, got %v", SnippetOAuthHeaders, category)
	}

	if confidence < 0.8 {
		t.Errorf("Expected high confidence (>0.8), got %.2f", confidence)
	}
}

func TestCategorizeSnippet_Unknown(t *testing.T) {
	content := `some_directive {
		custom_option value
	}`

	category, confidence := CategorizeSnippet(content)

	if category != SnippetUnknown {
		t.Errorf("Expected category %v, got %v", SnippetUnknown, category)
	}

	if confidence != 0.0 {
		t.Errorf("Expected zero confidence for unknown category, got %.2f", confidence)
	}
}

func TestGenerateDescription_IPRestriction(t *testing.T) {
	content := `@external {
		not remote_ip 10.0.0.0/8
	}`

	desc := GenerateDescription(SnippetIPRestriction, content)

	if !strings.Contains(desc, "IP address") {
		t.Errorf("Expected description to mention IP address, got: %s", desc)
	}
}

func TestGenerateDescription_SecurityHeaders(t *testing.T) {
	content := `header {
		X-Frame-Options "SAMEORIGIN"
		X-Content-Type-Options "nosniff"
	}`

	desc := GenerateDescription(SnippetSecurityHeaders, content)

	if !strings.Contains(desc, "clickjacking") || !strings.Contains(desc, "MIME") {
		t.Errorf("Expected description to mention specific headers, got: %s", desc)
	}
}

func TestSnippet_FullBlock(t *testing.T) {
	snippet := Snippet{
		Name:    "test_snippet",
		Content: "\theader X-Test \"value\"\n\trespond \"OK\"",
	}

	expected := "(test_snippet) {\n\theader X-Test \"value\"\n\trespond \"OK\"\n}"
	result := snippet.FullBlock()

	if result != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
	}
}

func TestSnippet_Lines(t *testing.T) {
	snippet := Snippet{
		Content: "line1\nline2\nline3",
	}

	if snippet.Lines() != 3 {
		t.Errorf("Expected 3 lines, got %d", snippet.Lines())
	}
}

func TestSnippetCategory_String(t *testing.T) {
	tests := []struct {
		category SnippetCategory
		expected string
	}{
		{SnippetIPRestriction, "IP Restriction"},
		{SnippetSecurityHeaders, "Security Headers"},
		{SnippetWebSocketHeaders, "WebSocket Headers"},
		{SnippetOAuthHeaders, "OAuth Headers"},
		{SnippetUnknown, "Unknown"},
	}

	for _, test := range tests {
		result := test.category.String()
		if result != test.expected {
			t.Errorf("Category %d: expected %s, got %s", test.category, test.expected, result)
		}
	}
}

func TestSnippetCategory_ColorCode(t *testing.T) {
	// Just verify each category returns a valid hex color
	categories := []SnippetCategory{
		SnippetIPRestriction,
		SnippetSecurityHeaders,
		SnippetPerformance,
		SnippetHTTPSBackend,
		SnippetOAuthHeaders,
		SnippetWebSocketHeaders,
		SnippetFrameEmbedding,
		SnippetCORS,
		SnippetCompression,
		SnippetRateLimit,
		SnippetCustom,
		SnippetUnknown,
	}

	for _, cat := range categories {
		color := cat.ColorCode()
		if !strings.HasPrefix(color, "#") || len(color) != 7 {
			t.Errorf("Category %s: expected hex color, got %s", cat, color)
		}
	}
}
