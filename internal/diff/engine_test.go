package diff

import (
	"testing"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/cloudflare"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		dns      []cloudflare.DNSRecord
		caddy    []caddy.CaddyEntry
		expected map[string]SyncStatus
	}{
		{
			name:     "empty inputs",
			dns:      nil,
			caddy:    nil,
			expected: map[string]SyncStatus{},
		},
		{
			name: "all synced",
			dns: []cloudflare.DNSRecord{
				{Name: "app.example.com", Type: "CNAME"},
				{Name: "api.example.com", Type: "A"},
			},
			caddy: []caddy.CaddyEntry{
				{Domains: []string{"app.example.com"}},
				{Domains: []string{"api.example.com"}},
			},
			expected: map[string]SyncStatus{
				"app.example.com": StatusSynced,
				"api.example.com": StatusSynced,
			},
		},
		{
			name: "orphaned DNS only",
			dns: []cloudflare.DNSRecord{
				{Name: "orphan.example.com", Type: "CNAME"},
			},
			caddy:    nil,
			expected: map[string]SyncStatus{"orphan.example.com": StatusOrphanedDNS},
		},
		{
			name: "orphaned Caddy only",
			dns:  nil,
			caddy: []caddy.CaddyEntry{
				{Domains: []string{"local.example.com"}},
			},
			expected: map[string]SyncStatus{"local.example.com": StatusOrphanedCaddy},
		},
		{
			name: "mixed statuses",
			dns: []cloudflare.DNSRecord{
				{Name: "synced.example.com"},
				{Name: "dns-only.example.com"},
			},
			caddy: []caddy.CaddyEntry{
				{Domains: []string{"synced.example.com"}},
				{Domains: []string{"caddy-only.example.com"}},
			},
			expected: map[string]SyncStatus{
				"synced.example.com":     StatusSynced,
				"dns-only.example.com":   StatusOrphanedDNS,
				"caddy-only.example.com": StatusOrphanedCaddy,
			},
		},
		{
			name: "case insensitive matching",
			dns: []cloudflare.DNSRecord{
				{Name: "App.Example.COM"},
			},
			caddy: []caddy.CaddyEntry{
				{Domains: []string{"app.example.com"}},
			},
			expected: map[string]SyncStatus{"app.example.com": StatusSynced},
		},
		{
			name: "multi-domain caddy entry",
			dns: []cloudflare.DNSRecord{
				{Name: "primary.example.com"},
				{Name: "alias.example.com"},
			},
			caddy: []caddy.CaddyEntry{
				{Domains: []string{"primary.example.com", "alias.example.com"}},
			},
			expected: map[string]SyncStatus{
				"primary.example.com": StatusSynced,
				"alias.example.com":  StatusSynced,
			},
		},
		{
			name: "multi-domain caddy partially matched",
			dns: []cloudflare.DNSRecord{
				{Name: "primary.example.com"},
			},
			caddy: []caddy.CaddyEntry{
				{Domains: []string{"primary.example.com", "alias.example.com"}},
			},
			expected: map[string]SyncStatus{
				"primary.example.com": StatusSynced,
				"alias.example.com":  StatusOrphanedCaddy,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Compare(tt.dns, tt.caddy)

			if len(results) != len(tt.expected) {
				t.Fatalf("got %d results, want %d", len(results), len(tt.expected))
			}

			for _, entry := range results {
				want, ok := tt.expected[entry.Domain]
				if !ok {
					t.Errorf("unexpected domain in results: %s", entry.Domain)
					continue
				}
				if entry.Status != want {
					t.Errorf("domain %s: got status %v, want %v", entry.Domain, entry.Status, want)
				}

				// Verify pointer correctness
				switch want {
				case StatusSynced:
					if entry.DNS == nil || entry.Caddy == nil {
						t.Errorf("domain %s: synced entry should have both DNS and Caddy non-nil", entry.Domain)
					}
				case StatusOrphanedDNS:
					if entry.DNS == nil || entry.Caddy != nil {
						t.Errorf("domain %s: orphaned DNS should have DNS non-nil and Caddy nil", entry.Domain)
					}
				case StatusOrphanedCaddy:
					if entry.DNS != nil || entry.Caddy == nil {
						t.Errorf("domain %s: orphaned Caddy should have DNS nil and Caddy non-nil", entry.Domain)
					}
				}
			}
		})
	}
}

func TestSyncStatusString(t *testing.T) {
	tests := []struct {
		status SyncStatus
		want   string
	}{
		{StatusSynced, "Synced"},
		{StatusOrphanedDNS, "Orphaned (DNS)"},
		{StatusOrphanedCaddy, "Orphaned (Caddy)"},
		{SyncStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("SyncStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestSyncStatusIcon(t *testing.T) {
	if got := StatusSynced.Icon(); got != "✓" {
		t.Errorf("StatusSynced.Icon() = %q, want ✓", got)
	}
	if got := StatusOrphanedDNS.Icon(); got != "⚠" {
		t.Errorf("StatusOrphanedDNS.Icon() = %q, want ⚠", got)
	}
}
