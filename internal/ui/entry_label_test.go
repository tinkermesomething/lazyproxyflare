package ui

import (
	"testing"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/diff"
)

// Entries that share one multi-domain Caddy block must still show their own hostname,
// otherwise the list shows the same label on several rows.
func TestEntryLabel(t *testing.T) {
	shared := &caddy.CaddyEntry{Domains: []string{"mail.example.com", "autodiscover.example.com", "autoconfig.example.com"}}
	single := &caddy.CaddyEntry{Domains: []string{"plex.example.com"}}

	tests := []struct {
		name  string
		entry diff.SyncedEntry
		want  string
	}{
		{"dns only", diff.SyncedEntry{Domain: "www.example.com"}, "www.example.com"},
		{"single-domain block", diff.SyncedEntry{Domain: "plex.example.com", Caddy: single}, "plex.example.com"},
		{"shared block, primary", diff.SyncedEntry{Domain: "mail.example.com", Caddy: shared}, "mail.example.com (shared x3)"},
		{"shared block, secondary", diff.SyncedEntry{Domain: "autodiscover.example.com", Caddy: shared}, "autodiscover.example.com (shared x3)"},
		{"legacy block without Domains list", diff.SyncedEntry{Domain: "old.example.com", Caddy: &caddy.CaddyEntry{Domain: "old.example.com"}}, "old.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := entryLabel(tt.entry); got != tt.want {
				t.Errorf("entryLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEntryLabelsAreUniquePerRow(t *testing.T) {
	shared := &caddy.CaddyEntry{Domains: []string{"a.example.com", "b.example.com"}}
	seen := map[string]bool{}
	for _, d := range []string{"a.example.com", "b.example.com"} {
		label := entryLabel(diff.SyncedEntry{Domain: d, Caddy: shared})
		if seen[label] {
			t.Errorf("duplicate label %q", label)
		}
		seen[label] = true
	}
}
