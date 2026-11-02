package ui

import (
	"testing"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/cloudflare"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
)

// TestGetDetailsPanelTitle_TabCloudflare tests title on Cloudflare tab
func TestGetDetailsPanelTitle_TabCloudflare(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCloudflare
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
			DNS: &cloudflare.DNSRecord{
				Type:    "A",
				Content: "1.2.3.4",
			},
		},
	}
	m.cursor = 0

	title := m.getDetailsPanelTitle()
	if title != "DNS Details" {
		t.Errorf("Expected 'DNS Details', got '%s'", title)
	}
}

// TestGetDetailsPanelTitle_TabCaddy tests title on Caddy tab
func TestGetDetailsPanelTitle_TabCaddy(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCaddy
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target: "localhost",
				Port:   8080,
			},
		},
	}
	m.cursor = 0

	title := m.getDetailsPanelTitle()
	if title != "Caddy Config" {
		t.Errorf("Expected 'Caddy Config', got '%s'", title)
	}
}

// TestGetDetailsPanelTitle_NoSelection tests title when no entry is selected
func TestGetDetailsPanelTitle_NoSelection(t *testing.T) {
	m := createTestModel()
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
		},
	}
	m.cursor = 5 // Cursor beyond entries list

	title := m.getDetailsPanelTitle()
	if title != "Welcome" {
		t.Errorf("Expected 'Welcome', got '%s'", title)
	}
}

// TestGetDetailsPanelTitle_EmptyEntries tests title with no entries
func TestGetDetailsPanelTitle_EmptyEntries(t *testing.T) {
	m := createTestModel()
	m.entries = []diff.SyncedEntry{}
	m.cursor = 0

	title := m.getDetailsPanelTitle()
	if title != "Welcome" {
		t.Errorf("Expected 'Welcome', got '%s'", title)
	}
}

// TestGetDetailsPanelTitle_TabSwitchContext tests tab awareness
func TestGetDetailsPanelTitle_TabSwitchContext(t *testing.T) {
	m := createTestModel()
	m.entries = []diff.SyncedEntry{
		{
			Domain: "api.example.com",
			Status: diff.StatusSynced,
			DNS: &cloudflare.DNSRecord{
				Type:    "A",
				Content: "1.2.3.4",
			},
			Caddy: &caddy.CaddyEntry{
				Target: "10.0.0.1",
				Port:   3000,
			},
		},
	}
	m.cursor = 0

	// Start on Cloudflare tab
	m.activeTab = TabCloudflare
	title := m.getDetailsPanelTitle()
	if title != "DNS Details" {
		t.Errorf("TabCloudflare: expected 'DNS Details', got '%s'", title)
	}

	// Switch to Caddy tab
	m.activeTab = TabCaddy
	title = m.getDetailsPanelTitle()
	if title != "Caddy Config" {
		t.Errorf("TabCaddy: expected 'Caddy Config', got '%s'", title)
	}
}

// TestGetDetailsPanelTitle_FilteredView tests title respects filters
func TestGetDetailsPanelTitle_FilteredView(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCloudflare
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
			DNS: &cloudflare.DNSRecord{
				Type:    "A",
				Content: "1.2.3.4",
			},
		},
		{
			Domain: "orphaned.example.com",
			Status: diff.StatusOrphanedDNS,
		},
	}
	// Apply filter
	m.statusFilter = FilterSynced
	m.cursor = 0

	title := m.getDetailsPanelTitle()
	if title != "DNS Details" {
		t.Errorf("Expected 'DNS Details' in filtered view, got '%s'", title)
	}
}

// TestGetDetailsPanelTitle_CursorAtBoundary tests cursor at boundary conditions
func TestGetDetailsPanelTitle_CursorAtBoundary(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCaddy
	m.entries = []diff.SyncedEntry{
		{
			Domain: "first.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target: "localhost",
				Port:   8080,
			},
		},
		{
			Domain: "second.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target: "localhost",
				Port:   9000,
			},
		},
	}

	// Test cursor at last valid entry
	m.cursor = 1
	title := m.getDetailsPanelTitle()
	if title != "Caddy Config" {
		t.Errorf("Expected 'Caddy Config' at boundary, got '%s'", title)
	}

	// Test cursor at last entry + 1
	m.cursor = 2
	title = m.getDetailsPanelTitle()
	if title != "Welcome" {
		t.Errorf("Expected 'Welcome' at out-of-bounds, got '%s'", title)
	}
}

// createTestModel creates a minimal Model for testing
func createTestModel() Model {
	return Model{
		entries:         []diff.SyncedEntry{},
		snippets:        []caddy.Snippet{},
		config:          &config.Config{Domain: "example.com"},
		cursor:          0,
		snippetPanel:    SnippetPanelState{Cursor: 0},
		activeTab:       TabCloudflare,
		statusFilter:    FilterAll,
		dnsTypeFilter:   DNSTypeAll,
		sortMode:        SortAlphabetical,
		selectedEntries: make(map[string]bool),
		panelFocus:      PanelFocusLeft,
	}
}

// TestGetDetailsPanelTitle_ConsistencyWithTab ensures consistency when switching tabs
func TestGetDetailsPanelTitle_ConsistencyWithTab(t *testing.T) {
	m := createTestModel()
	m.entries = []diff.SyncedEntry{
		{
			Domain: "test.com",
			Status: diff.StatusSynced,
			DNS: &cloudflare.DNSRecord{
				Type:    "CNAME",
				Content: "target.example.com",
			},
			Caddy: &caddy.CaddyEntry{
				Target: "localhost",
				Port:   8080,
			},
		},
	}
	m.cursor = 0

	// Verify both tabs work with same entry
	m.activeTab = TabCloudflare
	cloudflareTitle := m.getDetailsPanelTitle()

	m.activeTab = TabCaddy
	caddyTitle := m.getDetailsPanelTitle()

	if cloudflareTitle == caddyTitle {
		t.Errorf("Titles should differ by tab: '%s' vs '%s'", cloudflareTitle, caddyTitle)
	}

	if cloudflareTitle != "DNS Details" || caddyTitle != "Caddy Config" {
		t.Errorf("Unexpected titles: cloudflare='%s', caddy='%s'", cloudflareTitle, caddyTitle)
	}
}
