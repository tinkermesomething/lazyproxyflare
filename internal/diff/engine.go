package diff

import (
	"strings"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/cloudflare"
)

// Compare compares DNS records with Caddy entries and returns sync status for each domain
func Compare(dnsRecords []cloudflare.DNSRecord, caddyEntries []caddy.CaddyEntry) []SyncedEntry {
	var results []SyncedEntry

	// Build map of DNS records by domain name (lowercase for case-insensitive matching)
	dnsMap := make(map[string]*cloudflare.DNSRecord)
	for i := range dnsRecords {
		record := &dnsRecords[i]
		domain := strings.ToLower(record.Name)
		dnsMap[domain] = record
	}

	// Build map of Caddy entries by domain name
	// Handle multi-domain blocks by creating entry for each domain
	caddyMap := make(map[string]*caddy.CaddyEntry)
	for i := range caddyEntries {
		entry := &caddyEntries[i]
		// Add entry for each domain in the Domains array
		for _, domain := range entry.Domains {
			domain = strings.ToLower(domain)
			caddyMap[domain] = entry
		}
	}

	// Build set of all unique domains from both sources
	allDomains := make(map[string]bool)
	for domain := range dnsMap {
		allDomains[domain] = true
	}
	for domain := range caddyMap {
		allDomains[domain] = true
	}

	// Compare each domain and determine status
	for domain := range allDomains {
		dnsRecord := dnsMap[domain]
		caddyEntry := caddyMap[domain]

		synced := SyncedEntry{
			Domain: domain,
			DNS:    dnsRecord,
			Caddy:  caddyEntry,
		}

		// Determine sync status
		if dnsRecord != nil && caddyEntry != nil {
			// Both exist - synced
			synced.Status = StatusSynced
		} else if dnsRecord != nil && caddyEntry == nil {
			// Only in DNS
			synced.Status = StatusOrphanedDNS
		} else if dnsRecord == nil && caddyEntry != nil {
			// Only in Caddy
			synced.Status = StatusOrphanedCaddy
		}

		results = append(results, synced)
	}

	return results
}
