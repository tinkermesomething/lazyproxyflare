package caddy

import (
	"fmt"
	"strings"
)

// Map of known proxy-level snippet names
var proxyLevelSnippetNames = map[string]bool{
	"performance":     true,
	"frame_embedding": true,
	"https_backend":   true,
	// Add other known proxy-level snippets here as needed
}

// isProxyLevelSnippet checks if a given snippet name should be imported inside a reverse_proxy block.
func isProxyLevelSnippet(snippetName string) bool {
	return proxyLevelSnippetNames[snippetName]
}

// GenerateBlockInput contains all parameters needed to generate a Caddy block
type GenerateBlockInput struct {
	FQDN              string   // DEPRECATED: Use Domains instead. Kept for backwards compatibility
	Domains           []string // Multiple FQDNs for this entry (e.g., ["app.example.com", "api.example.com"])
	Target            string   // Reverse proxy target (IP or hostname)
	Port              int      // Service port
	SSL               bool     // Use https:// vs http://
	LANOnly           bool     // Restrict to LAN subnet
	OAuth             bool     // Include OAuth headers
	WebSocket         bool     // Include WebSocket headers
	LANSubnet         string   // LAN subnet for IP restriction (e.g., "10.0.28.0/24")
	AllowedExtIP      string   // Allowed external IP (e.g., "166.1.123.74/32")
	AvailableSnippets []string // List of available snippet names from Caddyfile
	SelectedSnippets  []string // List of snippet names to import
	CustomCaddyConfig string   // Custom Caddy directives for one-off features
}

// GenerateCaddyBlock generates a Caddy configuration block from input parameters
// Uses snippet imports when selected, otherwise generates inline configuration
func GenerateCaddyBlock(input GenerateBlockInput) string {
	var b strings.Builder

	// Backwards compatibility: if Domains is empty, use FQDN
	domains := input.Domains
	if len(domains) == 0 && input.FQDN != "" {
		domains = []string{input.FQDN}
	}

	// Ensure we have at least one domain
	if len(domains) == 0 {
		// Fallback - should not happen in practice
		domains = []string{"unknown.example.com"}
	}

	// Primary domain for marker comment (first domain)
	primaryDomain := domains[0]

	// Generate comma-separated domain list for Caddy block
	domainList := strings.Join(domains, ", ")

	// Helper function to check if a snippet is selected
	hasSnippet := func(name string) bool {
		for _, s := range input.SelectedSnippets {
			if s == name {
				return true
			}
		}
		return false
	}

	// Header marker comment (uses primary domain)
	b.WriteString(fmt.Sprintf("# === %s ===\n", primaryDomain))

	// Domain block (comma-separated if multiple domains)
	b.WriteString(fmt.Sprintf("%s {\n", domainList))

	// Separate snippets into site-level and proxy-level
	var siteLevelSnippets []string
	var proxyLevelSnippets []string
	for _, snippetName := range input.SelectedSnippets {
		if isProxyLevelSnippet(snippetName) {
			proxyLevelSnippets = append(proxyLevelSnippets, snippetName)
		} else {
			siteLevelSnippets = append(siteLevelSnippets, snippetName)
		}
	}

	// Import site-level snippets
	for _, snippetName := range siteLevelSnippets {
		b.WriteString(fmt.Sprintf("\timport %s\n", snippetName))
	}
	// Add a newline if there were site-level snippets for better formatting
	if len(siteLevelSnippets) > 0 {
		b.WriteString("\n")
	}

	// Add custom Caddy directives if provided
	if input.CustomCaddyConfig != "" {
		// Add each line with proper indentation
		customLines := strings.Split(input.CustomCaddyConfig, "\n")
		for _, line := range customLines {
			if strings.TrimSpace(line) != "" {
				b.WriteString("\t" + line + "\n")
			}
		}
		b.WriteString("\n")
	}

	// LAN-only restriction (only if snippet not used)
	if input.LANOnly && !hasSnippet("ip_restricted") {
		b.WriteString("\t@external {\n")
		b.WriteString(fmt.Sprintf("\t\tnot remote_ip %s %s\n", input.LANSubnet, input.AllowedExtIP))
		b.WriteString("\t}\n")
		b.WriteString("\trespond @external 404\n")
		b.WriteString("\n")
	}

	// Reverse proxy directive - infer protocol from port
	// Port 443 is always HTTPS, port 80 is always HTTP
	// For non-standard ports, use the SSL setting
	var protocol string
	if input.Port == 443 {
		protocol = "https"
	} else if input.Port == 80 {
		protocol = "http"
	} else {
		// Non-standard port - use SSL checkbox
		if input.SSL {
			protocol = "https"
		} else {
			protocol = "http"
		}
	}

	// Determine if a reverse_proxy block is needed (for headers or proxy-level snippets)
	needsProxyBlock := input.OAuth || input.WebSocket || len(proxyLevelSnippets) > 0

	if needsProxyBlock {
		// Open reverse_proxy block
		b.WriteString(fmt.Sprintf("\treverse_proxy %s://%s:%d {\n", protocol, input.Target, input.Port))

		// Import proxy-level snippets
		for _, snippetName := range proxyLevelSnippets {
			b.WriteString(fmt.Sprintf("\t\timport %s\n", snippetName))
		}
		// Add a newline if there were proxy-level snippets for better formatting
		if len(proxyLevelSnippets) > 0 {
			b.WriteString("\n")
		}

		// OAuth headers (inside reverse_proxy block)
		if input.OAuth {
			b.WriteString("\t\theader_up X-Forwarded-User {http.request.header.X-Forwarded-User}\n")
			b.WriteString("\t\theader_up X-Forwarded-Groups {http.request.header.X-Forwarded-Groups}\n")
			b.WriteString("\t\theader_up X-Forwarded-Email {http.request.header.X-Forwarded-Email}\n")
			b.WriteString("\t\theader_up X-Forwarded-Preferred-Username {http.request.header.X-Forwarded-Preferred-Username}\n")
		}

		// WebSocket headers (inside reverse_proxy block)
		if input.WebSocket {
			b.WriteString("\t\theader_up Upgrade {http.request.header.Upgrade}\n")
			b.WriteString("\t\theader_up Connection {http.request.header.Connection}\n")
		}

		// Close reverse_proxy block
		b.WriteString("\t}\n")
	} else {
		// Simple one-line reverse_proxy (no headers or proxy-level snippets)
		b.WriteString(fmt.Sprintf("\treverse_proxy %s://%s:%d\n", protocol, input.Target, input.Port))
	}

	// Close block
	b.WriteString("}\n")

	return b.String()
}
