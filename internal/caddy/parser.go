package caddy

import (
	"log"
	"strconv"
	"strings"
)

// countBracesOutsideQuotes counts { and } characters that are not inside quoted strings.
// Returns (openCount, closeCount).
func countBracesOutsideQuotes(line string) (int, int) {
	openCount := 0
	closeCount := 0
	inQuote := false
	quoteChar := rune(0)
	escaped := false

	for _, ch := range line {
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if !inQuote && (ch == '"' || ch == '\'') {
			inQuote = true
			quoteChar = ch
			continue
		}
		if inQuote && ch == quoteChar {
			inQuote = false
			quoteChar = 0
			continue
		}
		if !inQuote {
			if ch == '{' {
				openCount++
			} else if ch == '}' {
				closeCount++
			}
		}
	}
	return openCount, closeCount
}

// ParserLogger is an optional logger for parser debugging
var ParserLogger *log.Logger

// ParsedCaddyfile contains both domain entries and snippets
type ParsedCaddyfile struct {
	Entries  []CaddyEntry
	Snippets []Snippet // List of parsed snippets with full metadata
}

// ParseCaddyfile parses a Caddyfile and extracts domain entries
func ParseCaddyfile(content string) ([]CaddyEntry, error) {
	parsed := ParseCaddyfileWithSnippets(content)
	return parsed.Entries, nil
}

// ParseCaddyfileWithSnippets parses a Caddyfile and extracts both domain entries and snippets
func ParseCaddyfileWithSnippets(content string) ParsedCaddyfile {
	var entries []CaddyEntry
	var snippets []Snippet
	lines := strings.Split(content, "\n")

	// Skip global block and snippet definitions
	inGlobalBlock := false

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Detect global block
		if line == "{" && i == 0 {
			inGlobalBlock = true
			continue
		}

		// Parse snippet definitions: (name) {
		if strings.HasPrefix(line, "(") && strings.Contains(line, ")") {
			snippet, endLine := parseSnippetBlock(lines, i)
			if snippet != nil {
				snippets = append(snippets, *snippet)
				if ParserLogger != nil {
					ParserLogger.Printf("✓ Parsed snippet: %s (category: %s, confidence: %.2f) at lines %d-%d",
						snippet.Name, snippet.Category, snippet.Confidence, snippet.LineStart, snippet.LineEnd)
				}
			}
			i = endLine
			continue
		}

		// Look for domain blocks: domain.com { or domain.com, other.com {
		if strings.HasSuffix(line, "{") && !inGlobalBlock {
			// Check if this looks like a domain (has a dot or is a domain pattern)
			domainPart := strings.TrimSuffix(line, "{")
			domainPart = strings.TrimSpace(domainPart)

			// Skip if it's a matcher block like @external {
			if strings.HasPrefix(domainPart, "@") {
				if ParserLogger != nil {
					ParserLogger.Printf("Skipping matcher block: %s at line %d", domainPart, i+1)
				}
				i = skipBlock(lines, i)
				continue
			}

			// Parse domain block
			if ParserLogger != nil {
				ParserLogger.Printf("Parsing domain block: %s at line %d", domainPart, i+1)
			}
			entry, endLine := parseDomainBlock(lines, i)
			if entry != nil {
				entries = append(entries, *entry)
				if ParserLogger != nil {
					ParserLogger.Printf("✓ Parsed entry for %s (lines %d-%d)", entry.Domain, entry.LineStart, entry.LineEnd)
				}
			}
			i = endLine
		}

		// Check if we're exiting global block
		if inGlobalBlock && line == "}" {
			inGlobalBlock = false
		}
	}

	return ParsedCaddyfile{
		Entries:  entries,
		Snippets: snippets,
	}
}

// skipBlock skips an entire block by counting braces
func skipBlock(lines []string, start int) int {
	// Count braces on start line
	open, close := countBracesOutsideQuotes(lines[start])
	braceCount := open - close

	for i := start + 1; i < len(lines); i++ {
		line := lines[i]
		open, close := countBracesOutsideQuotes(line)
		braceCount += open
		braceCount -= close

		if braceCount == 0 {
			return i
		}
	}

	return len(lines) - 1
}

// parseSnippetBlock extracts a snippet definition and its content
func parseSnippetBlock(lines []string, start int) (*Snippet, int) {
	snippet := &Snippet{
		LineStart: start + 1, // 1-indexed
	}

	// Extract snippet name from first line: (snippet_name) {
	firstLine := strings.TrimSpace(lines[start])
	snippetName := strings.TrimPrefix(firstLine, "(")
	snippetName = strings.TrimSuffix(snippetName, " {")
	snippetName = strings.TrimSuffix(snippetName, "{")
	snippetName = strings.TrimSpace(snippetName)
	snippetName = strings.TrimSuffix(snippetName, ")")
	snippet.Name = snippetName

	// Extract complete block with brace counting
	braceCount := 1 // Opening brace on first line
	endLine := start
	var contentLines []string

	for i := start + 1; i < len(lines); i++ {
		line := lines[i]

		open, close := countBracesOutsideQuotes(line)
		braceCount += open
		braceCount -= close

		// Don't include the closing brace in content
		if braceCount == 0 {
			endLine = i
			break
		}

		contentLines = append(contentLines, line)
	}

	snippet.LineEnd = endLine + 1 // 1-indexed
	snippet.Content = strings.Join(contentLines, "\n")

	// Auto-categorize the snippet
	category, confidence := CategorizeSnippet(snippet.Content)
	snippet.Category = category
	snippet.Confidence = confidence
	snippet.AutoDetected = confidence > 0.0

	// Generate description
	snippet.Description = GenerateDescription(category, snippet.Content)

	return snippet, endLine
}

// parseDomainBlock extracts a complete domain block and parses its contents
func parseDomainBlock(lines []string, start int) (*CaddyEntry, int) {
	entry := &CaddyEntry{
		LineStart: start + 1, // 1-indexed
		Imports:   []string{},
		Domains:   []string{},
	}

	// Extract domain(s) from first line
	firstLine := strings.TrimSpace(lines[start])

	// More aggressive whitespace handling - handle domain.com{, domain.com  {, etc.
	// Find the opening brace and extract everything before it
	braceIndex := strings.Index(firstLine, "{")
	var domainPart string
	if braceIndex >= 0 {
		domainPart = strings.TrimSpace(firstLine[:braceIndex])
	} else {
		// Shouldn't happen if we got here, but handle it gracefully
		domainPart = strings.TrimSuffix(firstLine, "{")
		domainPart = strings.TrimSpace(domainPart)
	}

	// Check for marker comment on previous line
	if start > 0 {
		prevLine := strings.TrimSpace(lines[start-1])
		if strings.HasPrefix(prevLine, "# ===") && strings.HasSuffix(prevLine, "===") {
			entry.HasMarker = true
		}
	}

	// Handle multiple domains (comma-separated)
	// More aggressive whitespace trimming for each domain
	if strings.Contains(domainPart, ",") {
		domains := strings.Split(domainPart, ",")
		for _, d := range domains {
			d = strings.TrimSpace(d)
			if d != "" { // Skip empty strings from extra commas
				entry.Domains = append(entry.Domains, d)
			}
		}
		if len(entry.Domains) > 0 {
			entry.Domain = entry.Domains[0] // Primary is first
		}
	} else {
		entry.Domain = strings.TrimSpace(domainPart)
		entry.Domains = []string{entry.Domain}
	}

	// Extract complete block with brace counting
	blockLines := []string{lines[start]}
	braceCount := 1
	endLine := start

	for i := start + 1; i < len(lines); i++ {
		line := lines[i]
		blockLines = append(blockLines, line)

		open, close := countBracesOutsideQuotes(line)
		braceCount += open
		braceCount -= close

		if braceCount == 0 {
			endLine = i
			break
		}
	}

	entry.LineEnd = endLine + 1 // 1-indexed
	entry.RawBlock = strings.Join(blockLines, "\n")

	// Parse block contents
	parseBlockContents(entry, blockLines)

	// Set default port if not specified
	if entry.Port == 0 {
		if entry.SSL {
			entry.Port = 443
		} else {
			entry.Port = 80
		}
	}

	return entry, endLine
}

// parseBlockContents extracts configuration details from block lines
func parseBlockContents(entry *CaddyEntry, lines []string) {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Parse reverse_proxy directive
		if strings.HasPrefix(trimmed, "reverse_proxy ") {
			parseReverseProxy(entry, trimmed)
		}

		// Parse import statements
		if strings.HasPrefix(trimmed, "import ") {
			importName := strings.TrimPrefix(trimmed, "import ")
			importName = strings.TrimSpace(importName)
			entry.Imports = append(entry.Imports, importName)

			// Check for IP restriction import
			if importName == "ip_restricted" {
				entry.IPRestricted = true
			}
		}

		// Detect inline IP restriction matcher
		if strings.HasPrefix(trimmed, "@") && strings.Contains(trimmed, "not_allowed") {
			entry.IPRestricted = true
		}
		if strings.HasPrefix(trimmed, "@") && strings.Contains(trimmed, "external") {
			entry.IPRestricted = true
		}

		// Detect OAuth headers
		if strings.Contains(trimmed, "header_up X-Real-IP") {
			entry.OAuthHeaders = true
		}
		if strings.Contains(trimmed, "header_up X-Forwarded-For") {
			entry.OAuthHeaders = true
		}

		// Detect WebSocket support
		if strings.Contains(trimmed, "header_up Upgrade") {
			entry.WebSocket = true
		}
		if strings.Contains(trimmed, "header_up Connection") && strings.Contains(trimmed, "Upgrade") {
			entry.WebSocket = true
		}
	}
}

// parseReverseProxy extracts target, port, and SSL from reverse_proxy directive
func parseReverseProxy(entry *CaddyEntry, line string) {
	// Example: reverse_proxy https://10.0.28.9:32400
	// Example: reverse_proxy http://localhost:80
	// Example: reverse_proxy 10.0.28.3:4080

	// Remove "reverse_proxy " prefix
	line = strings.TrimPrefix(line, "reverse_proxy ")
	line = strings.TrimSpace(line)

	// Remove any trailing {
	line = strings.TrimSuffix(line, "{")
	line = strings.TrimSpace(line)

	// Check for SSL
	if strings.HasPrefix(line, "https://") {
		entry.SSL = true
		line = strings.TrimPrefix(line, "https://")
	} else if strings.HasPrefix(line, "http://") {
		entry.SSL = false
		line = strings.TrimPrefix(line, "http://")
	} else {
		// No scheme specified, assume http
		entry.SSL = false
	}

	// Extract host:port
	// Could be: 10.0.28.9:32400 or localhost:80
	parts := strings.Split(line, ":")
	if len(parts) >= 1 {
		entry.Target = parts[0]
	}
	if len(parts) >= 2 {
		// Extract port number (might have other params after)
		portStr := parts[1]
		// Port might be followed by space or other params
		portFields := strings.Fields(portStr)
		if len(portFields) > 0 {
			if port, err := strconv.Atoi(portFields[0]); err == nil {
				entry.Port = port
			}
		}
	}
}

