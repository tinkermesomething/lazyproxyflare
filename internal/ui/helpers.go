package ui

import (
	"sort"
	"strings"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/diff"
)

// getFilteredEntries returns entries filtered by search query, status filter, and DNS type filter,
// then sorted according to the current sort mode. Also filters by active tab.
func (m Model) getFilteredEntries() []diff.SyncedEntry {
	filtered := m.entries

	// Apply tab filter first (filter by which data exists in the entry)
	switch m.activeTab {
	case TabCloudflare:
		// Show only entries with DNS records
		var dnsFiltered []diff.SyncedEntry
		for _, entry := range filtered {
			if entry.DNS != nil {
				dnsFiltered = append(dnsFiltered, entry)
			}
		}
		filtered = dnsFiltered
	case TabCaddy:
		// Show only entries with Caddy config
		var caddyFiltered []diff.SyncedEntry
		for _, entry := range filtered {
			if entry.Caddy != nil {
				caddyFiltered = append(caddyFiltered, entry)
			}
		}
		filtered = caddyFiltered
	}

	// Apply status filter
	if m.statusFilter != FilterAll {
		var statusFiltered []diff.SyncedEntry
		for _, entry := range filtered {
			switch m.statusFilter {
			case FilterSynced:
				if entry.Status == diff.StatusSynced {
					statusFiltered = append(statusFiltered, entry)
				}
			case FilterOrphanedDNS:
				if entry.Status == diff.StatusOrphanedDNS {
					statusFiltered = append(statusFiltered, entry)
				}
			case FilterOrphanedCaddy:
				if entry.Status == diff.StatusOrphanedCaddy {
					statusFiltered = append(statusFiltered, entry)
				}
			}
		}
		filtered = statusFiltered
	}

	// Apply DNS type filter
	if m.dnsTypeFilter != DNSTypeAll {
		var dnsTypeFiltered []diff.SyncedEntry
		for _, entry := range filtered {
			// Only filter entries that have DNS records
			if entry.DNS != nil {
				switch m.dnsTypeFilter {
				case DNSTypeCNAME:
					if entry.DNS.Type == "CNAME" {
						dnsTypeFiltered = append(dnsTypeFiltered, entry)
					}
				case DNSTypeA:
					if entry.DNS.Type == "A" {
						dnsTypeFiltered = append(dnsTypeFiltered, entry)
					}
				}
			}
		}
		filtered = dnsTypeFiltered
	}

	// Apply search query filter
	if m.searchQuery != "" {
		var searchFiltered []diff.SyncedEntry
		query := strings.ToLower(m.searchQuery)
		for _, entry := range filtered {
			if strings.Contains(strings.ToLower(entry.Domain), query) {
				searchFiltered = append(searchFiltered, entry)
			}
		}
		filtered = searchFiltered
	}

	// Apply sorting
	switch m.sortMode {
	case SortAlphabetical:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Domain < filtered[j].Domain
		})
	case SortByStatus:
		sort.Slice(filtered, func(i, j int) bool {
			// Sort order: Synced (0) < Orphaned DNS (1) < Orphaned Caddy (2)
			if filtered[i].Status != filtered[j].Status {
				return filtered[i].Status < filtered[j].Status
			}
			// Within same status, sort alphabetically
			return filtered[i].Domain < filtered[j].Domain
		})
	}

	return filtered
}

// getSnippetNames extracts just the names from a slice of Snippet structs
// This is used when passing to the Caddy generator which only needs snippet names
func getSnippetNames(snippets []caddy.Snippet) []string {
	names := make([]string, len(snippets))
	for i, snippet := range snippets {
		names[i] = snippet.Name
	}
	return names
}

// getSelectedSnippetNames converts the SelectedSnippets map to a slice of selected snippet names
// This is used when passing selected snippets to the Caddy generator
func getSelectedSnippetNames(selectedSnippets map[string]bool) []string {
	var selected []string
	for name, isSelected := range selectedSnippets {
		if isSelected {
			selected = append(selected, name)
		}
	}
	return selected
}
