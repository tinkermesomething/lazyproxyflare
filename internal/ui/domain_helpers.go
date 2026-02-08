package ui

import (
	"fmt"
	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/diff"
	"strings"
)

// ParseSubdomains parses a multi-line textarea input into individual subdomains
// Filters out empty lines and trims whitespace
func ParseSubdomains(input string) []string {
	lines := strings.Split(input, "\n")
	subdomains := []string{}

	for _, line := range lines {
		subdomain := strings.TrimSpace(line)
		if subdomain != "" {
			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains
}

// BuildFQDNs converts subdomains to fully qualified domain names
// Example: ["app", "api"] + "example.com" -> ["app.example.com", "api.example.com"]
func BuildFQDNs(subdomains []string, baseDomain string) []string {
	fqdns := []string{}
	for _, subdomain := range subdomains {
		fqdn := subdomain + "." + baseDomain
		fqdns = append(fqdns, fqdn)
	}
	return fqdns
}

// ValidateSubdomains checks that subdomains are valid and non-duplicate
func ValidateSubdomains(subdomains []string) error {
	if len(subdomains) == 0 {
		return fmt.Errorf("at least one subdomain is required")
	}

	// Check for empty subdomains
	for i, subdomain := range subdomains {
		if strings.TrimSpace(subdomain) == "" {
			return fmt.Errorf("subdomain %d is empty", i+1)
		}
	}

	// Check for duplicates (case-insensitive)
	seen := make(map[string]bool)
	for _, subdomain := range subdomains {
		normalized := strings.TrimSpace(strings.ToLower(subdomain))
		if seen[normalized] {
			return fmt.Errorf("duplicate subdomain: %s", subdomain)
		}
		seen[normalized] = true
	}

	return nil
}

// CheckDomainConflicts checks if any of the FQDNs conflict with existing entries
// Returns error with conflicting domain name if found
func CheckDomainConflicts(fqdns []string, entries []diff.SyncedEntry, excludeDomain string) error {
	for _, fqdn := range fqdns {
		fqdnLower := strings.ToLower(fqdn)

		// Skip if this is the entry being edited
		if excludeDomain != "" && strings.ToLower(excludeDomain) == fqdnLower {
			continue
		}

		for _, entry := range entries {
			if entry.Caddy != nil {
				// Check all domains in the entry
				for _, domain := range entry.Caddy.Domains {
					if strings.ToLower(domain) == fqdnLower {
						primaryDomain := caddy.GetPrimaryDomain(entry.Caddy)
						if len(entry.Caddy.Domains) > 1 {
							return fmt.Errorf("domain %s already exists in entry '%s' (multi-domain entry)", fqdn, primaryDomain)
						}
						return fmt.Errorf("domain %s already exists in entry '%s'", fqdn, primaryDomain)
					}
				}

				// Check legacy Domain field for backwards compatibility
				if entry.Caddy.Domain != "" && strings.ToLower(entry.Caddy.Domain) == fqdnLower {
					return fmt.Errorf("domain %s already exists in entry '%s'", fqdn, entry.Caddy.Domain)
				}
			}
		}
	}

	return nil
}

// FormatSubdomainsForDisplay formats a subdomain string for display
// Shows "subdomain1" or "subdomain1, subdomain2, ..." truncated if too long
func FormatSubdomainsForDisplay(subdomains string, maxLength int) string {
	parsed := ParseSubdomains(subdomains)
	if len(parsed) == 0 {
		return ""
	}

	if len(parsed) == 1 {
		return parsed[0]
	}

	// Multiple subdomains - join with commas
	joined := strings.Join(parsed, ", ")
	if len(joined) <= maxLength {
		return joined
	}

	// Truncate and add ellipsis
	return joined[:maxLength-3] + "..."
}

// GetSubdomainsTextareaValue formats domains list back into textarea format (newline-separated)
// Used when editing an existing multi-domain entry
func GetSubdomainsTextareaValue(domains []string, baseDomain string) string {
	subdomains := []string{}

	for _, domain := range domains {
		// Remove base domain suffix to get subdomain
		domain = strings.TrimSpace(domain)
		suffix := "." + baseDomain
		if strings.HasSuffix(domain, suffix) {
			subdomain := strings.TrimSuffix(domain, suffix)
			subdomains = append(subdomains, subdomain)
		} else {
			// If it doesn't match base domain, include as-is (shouldn't happen)
			subdomains = append(subdomains, domain)
		}
	}

	return strings.Join(subdomains, "\n")
}
