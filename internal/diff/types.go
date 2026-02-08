package diff

import (
	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/cloudflare"
)

// SyncStatus represents the sync state of an entry
type SyncStatus int

const (
	StatusSynced        SyncStatus = iota // Both exist
	StatusOrphanedDNS                     // Exists in DNS only
	StatusOrphanedCaddy                   // Exists in Caddy only
)

// String returns human-readable status
func (s SyncStatus) String() string {
	switch s {
	case StatusSynced:
		return "Synced"
	case StatusOrphanedDNS:
		return "Orphaned (DNS)"
	case StatusOrphanedCaddy:
		return "Orphaned (Caddy)"
	default:
		return "Unknown"
	}
}

// Icon returns emoji/symbol for TUI display
func (s SyncStatus) Icon() string {
	switch s {
	case StatusSynced:
		return "✓"
	case StatusOrphanedDNS:
		return "⚠"
	case StatusOrphanedCaddy:
		return "⚠"
	default:
		return "?"
	}
}

// SyncedEntry represents the result of comparing DNS and Caddy
type SyncedEntry struct {
	Domain string                   // Primary domain name
	DNS    *cloudflare.DNSRecord    // nil if not in DNS
	Caddy  *caddy.CaddyEntry        // nil if not in Caddy
	Status SyncStatus               // Sync status
}
