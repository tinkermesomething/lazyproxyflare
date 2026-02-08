package caddy

import (
	"fmt"
	"strings"
)

// GetPrimaryDomain returns the primary domain from a CaddyEntry
// Primary is the first domain in the Domains list
func GetPrimaryDomain(entry *CaddyEntry) string {
	if len(entry.Domains) > 0 {
		return entry.Domains[0]
	}
	// Fallback to old Domain field for backwards compatibility
	return entry.Domain
}

// GetAdditionalDomains returns all domains except the primary
func GetAdditionalDomains(entry *CaddyEntry) []string {
	if len(entry.Domains) <= 1 {
		return []string{}
	}
	return entry.Domains[1:]
}

// GetDomainCount returns the total number of domains for an entry
func GetDomainCount(entry *CaddyEntry) int {
	if len(entry.Domains) > 0 {
		return len(entry.Domains)
	}
	// Backwards compatibility
	if entry.Domain != "" {
		return 1
	}
	return 0
}

// GetAdditionalDomainCount returns the number of additional domains (excluding primary)
func GetAdditionalDomainCount(entry *CaddyEntry) int {
	count := GetDomainCount(entry)
	if count > 1 {
		return count - 1
	}
	return 0
}

// FormatDomainDisplay returns a string for displaying domains in the UI
// Shows "primary.example.com" or "primary.example.com (+2 more)"
func FormatDomainDisplay(entry *CaddyEntry) string {
	primary := GetPrimaryDomain(entry)
	additionalCount := GetAdditionalDomainCount(entry)

	if additionalCount == 0 {
		return primary
	}

	return fmt.Sprintf("%s (+%d more)", primary, additionalCount)
}

// HasDomain checks if an entry contains a specific domain
func HasDomain(entry *CaddyEntry, domain string) bool {
	domain = strings.TrimSpace(strings.ToLower(domain))

	for _, d := range entry.Domains {
		if strings.TrimSpace(strings.ToLower(d)) == domain {
			return true
		}
	}

	// Check old Domain field for backwards compatibility
	if strings.TrimSpace(strings.ToLower(entry.Domain)) == domain {
		return true
	}

	return false
}

// AddDomain adds a domain to an entry's domain list
// Returns error if domain already exists in the entry
func AddDomain(entry *CaddyEntry, domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Check for duplicates
	if HasDomain(entry, domain) {
		return fmt.Errorf("domain already exists in this entry: %s", domain)
	}

	// Initialize Domains slice if needed
	if len(entry.Domains) == 0 {
		// Migrate from old Domain field if present
		if entry.Domain != "" {
			entry.Domains = []string{entry.Domain}
		} else {
			entry.Domains = []string{}
		}
	}

	// Add new domain
	entry.Domains = append(entry.Domains, domain)

	// Update Domain field to maintain compatibility (always first domain)
	entry.Domain = entry.Domains[0]

	return nil
}

// RemoveDomain removes a domain from an entry's domain list
// Returns error if it's the last domain (entries must have at least one domain)
func RemoveDomain(entry *CaddyEntry, domain string) error {
	domain = strings.TrimSpace(strings.ToLower(domain))

	// Check if entry has multiple domains
	if GetDomainCount(entry) <= 1 {
		return fmt.Errorf("cannot remove last domain - entry must have at least one domain")
	}

	// Find and remove the domain
	newDomains := []string{}
	found := false

	for _, d := range entry.Domains {
		if strings.TrimSpace(strings.ToLower(d)) != domain {
			newDomains = append(newDomains, d)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("domain not found in entry: %s", domain)
	}

	entry.Domains = newDomains

	// Update Domain field to new primary (first domain)
	if len(entry.Domains) > 0 {
		entry.Domain = entry.Domains[0]
	}

	return nil
}

// SetPrimaryDomain sets a domain as the primary (first in list)
// Returns error if domain is not in the entry's domain list
func SetPrimaryDomain(entry *CaddyEntry, domain string) error {
	domain = strings.TrimSpace(domain)

	// Find the domain index
	domainIndex := -1
	for i, d := range entry.Domains {
		if strings.TrimSpace(strings.ToLower(d)) == strings.TrimSpace(strings.ToLower(domain)) {
			domainIndex = i
			break
		}
	}

	if domainIndex == -1 {
		return fmt.Errorf("domain not found in entry: %s", domain)
	}

	// Already primary
	if domainIndex == 0 {
		return nil
	}

	// Move domain to first position
	newDomains := []string{domain}
	for i, d := range entry.Domains {
		if i != domainIndex {
			newDomains = append(newDomains, d)
		}
	}

	entry.Domains = newDomains
	entry.Domain = entry.Domains[0]

	return nil
}

// MigrateToMultiDomain ensures an entry uses the new Domains field
// This provides backwards compatibility for old entries
func MigrateToMultiDomain(entry *CaddyEntry) {
	if len(entry.Domains) == 0 && entry.Domain != "" {
		// Old format: single Domain field
		entry.Domains = []string{entry.Domain}
	}

	// Ensure Domain field is set to primary
	if len(entry.Domains) > 0 && entry.Domain == "" {
		entry.Domain = entry.Domains[0]
	}
}

// ValidateDomains checks that an entry has valid domain configuration
func ValidateDomains(entry *CaddyEntry) error {
	// Ensure migration to new format
	MigrateToMultiDomain(entry)

	// Must have at least one domain
	if GetDomainCount(entry) == 0 {
		return fmt.Errorf("entry must have at least one domain")
	}

	// Check for empty domains
	for i, domain := range entry.Domains {
		if strings.TrimSpace(domain) == "" {
			return fmt.Errorf("domain %d is empty", i+1)
		}
	}

	// Check for duplicates within entry
	seen := make(map[string]bool)
	for _, domain := range entry.Domains {
		normalized := strings.TrimSpace(strings.ToLower(domain))
		if seen[normalized] {
			return fmt.Errorf("duplicate domain in entry: %s", domain)
		}
		seen[normalized] = true
	}

	return nil
}
