package caddy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseCaddyfileForMigration(t *testing.T) {
	tmpDir := t.TempDir()
	caddyfilePath := filepath.Join(tmpDir, "Caddyfile")

	testContent := `{
	admin off
	auto_https off
}

(ip_restricted) {
	@external {
		not remote_ip 10.0.28.0/24
	}
	respond @external 404
}

(security_headers) {
	header X-Frame-Options "DENY"
	header X-Content-Type-Options "nosniff"
}

plex.example.com {
	import ip_restricted
	reverse_proxy http://localhost:32400
}

grafana.example.com {
	import security_headers
	reverse_proxy https://localhost:3000
}
`
	if err := os.WriteFile(caddyfilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create temp Caddyfile: %v", err)
	}

	content, err := ParseCaddyfileForMigration(caddyfilePath)
	if err != nil {
		t.Fatalf("Failed to parse Caddyfile: %v", err)
	}

	entriesCount, snippetsCount := content.CountContent()

	if !content.HasContent {
		t.Error("Expected Caddyfile to have content")
	}

	if entriesCount != 2 {
		t.Errorf("Expected 2 entries, got %d", entriesCount)
	}

	if snippetsCount != 2 {
		t.Errorf("Expected 2 snippets, got %d", snippetsCount)
	}

	if !content.HasGlobalOptions() {
		t.Error("Expected global options to be present")
	}

	fmt.Printf("\n=== Migration Parser Test ===\n")
	fmt.Printf("Entries: %d, Snippets: %d, Global: %v\n",
		entriesCount, snippetsCount, content.HasGlobalOptions())
}

func TestExtractGlobalOptions(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Simple global block",
			input: `{
	admin off
	auto_https off
}

(snippet) {
	test
}`,
			expected: "\tadmin off\n\tauto_https off",
		},
		{
			name: "Global block with comments",
			input: `# Top comment
{
	# Global options
	admin off
	auto_https off
}

domain.com {
}`,
			expected: "\t# Global options\n\tadmin off\n\tauto_https off",
		},
		{
			name:     "No global block",
			input:    `domain.com {\n}\n`,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractGlobalOptions(tc.input)
			if result != tc.expected {
				t.Errorf("Expected:\n%q\n\nGot:\n%q", tc.expected, result)
			}
		})
	}
}

func TestGetCaddyfileTemplate(t *testing.T) {
	template := GetCaddyfileTemplate()

	if template == "" {
		t.Error("Template should not be empty")
	}

	// Should contain key sections
	if !containsAll(template, []string{
		"SNIPPETS",
		"TUNNEL ENTRIES",
		"Managed by LazyProxyFlare",
	}) {
		t.Error("Template missing expected sections")
	}

	fmt.Printf("\n=== Fresh Caddyfile Template ===\n%s\n", template)
}

func containsAll(s string, substrs []string) bool {
	for _, substr := range substrs {
		if !contains(s, substr) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
