package caddy

import (
	"fmt"
	"os"
	"strings"
)

// MigrationContent represents a complete parsed Caddyfile for migration
type MigrationContent struct {
	GlobalOptions string         // Global block content (everything inside the { } at top)
	Snippets      []Snippet      // All parsed snippets with metadata
	Entries       []CaddyEntry   // All parsed domain entries
	RawContent    string         // Original file content
	HasContent    bool           // True if file has any entries or snippets
}

// ParseCaddyfileForMigration reads and parses a Caddyfile from disk for migration purposes
func ParseCaddyfileForMigration(filepath string) (*MigrationContent, error) {
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Caddyfile: %w", err)
	}

	contentStr := string(content)

	// Parse using existing parser
	parsed := ParseCaddyfileWithSnippets(contentStr)

	// Extract global options
	globalOptions := extractGlobalOptions(contentStr)

	// Determine if file has meaningful content
	hasContent := len(parsed.Entries) > 0 || len(parsed.Snippets) > 0

	return &MigrationContent{
		GlobalOptions: globalOptions,
		Snippets:      parsed.Snippets,
		Entries:       parsed.Entries,
		RawContent:    contentStr,
		HasContent:    hasContent,
	}, nil
}

// extractGlobalOptions extracts the global configuration block from Caddyfile
// Returns the content inside the top-level { } block
func extractGlobalOptions(content string) string {
	lines := strings.Split(content, "\n")

	// Look for opening brace at start (after optional whitespace/comments)
	inGlobalBlock := false
	globalLines := []string{}
	braceCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is the start of global block (first non-comment, non-empty line should be {)
		if !inGlobalBlock {
			// Skip empty lines and comments
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			// First real line should be the global block opener
			if trimmed == "{" {
				inGlobalBlock = true
				braceCount = 1
				continue
			} else {
				// No global block found
				return ""
			}
		}

		// We're inside the global block
		if inGlobalBlock {
			// Count braces
			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")

			// If we hit zero, we've closed the global block
			if braceCount == 0 {
				// Don't include the closing brace
				break
			}

			// Add this line to global options (preserve indentation)
			globalLines = append(globalLines, lines[i])
		}
	}

	if len(globalLines) == 0 {
		return ""
	}

	return strings.Join(globalLines, "\n")
}

// GetDefaultGlobalOptions returns the default global options for a fresh Caddyfile
func GetDefaultGlobalOptions() string {
	return `	# Global options
	admin off
	auto_https off`
}

// GetCaddyfileTemplate returns a clean Caddyfile template for fresh start
func GetCaddyfileTemplate() string {
	return `{
` + GetDefaultGlobalOptions() + `
}

#──────────────────────────────────────────────────────────────
# SNIPPETS (Managed by LazyProxyFlare)
#──────────────────────────────────────────────────────────────

# Snippets will be added here by the application

#──────────────────────────────────────────────────────────────
# TUNNEL ENTRIES (Managed by LazyProxyFlare)
#──────────────────────────────────────────────────────────────

# Domain entries will be added here by the application
`
}

// CountContent returns human-readable summary of Caddyfile content
func (mc *MigrationContent) CountContent() (entriesCount, snippetsCount int) {
	return len(mc.Entries), len(mc.Snippets)
}

// HasGlobalOptions returns true if the Caddyfile has a global options block
func (mc *MigrationContent) HasGlobalOptions() bool {
	return strings.TrimSpace(mc.GlobalOptions) != ""
}
