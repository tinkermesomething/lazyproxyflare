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

// TestRenderSnippetsPanelTitle_TabCaddyWithSnippets tests snippets title on Caddy tab
func TestRenderSnippetsPanelTitle_TabCaddyWithSnippets(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCaddy
	m.entries = []diff.SyncedEntry{
		{
			Domain: "secure.example.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target:  "localhost",
				Port:    8080,
				Imports: []string{"oauth", "rate-limit"},
			},
		},
	}
	m.cursor = 0

	title := m.renderSnippetsPanelTitle()
	expected := "Applied Snippets (2)"
	if title != expected {
		t.Errorf("Expected '%s', got '%s'", expected, title)
	}
}

// TestRenderSnippetsPanelTitle_TabCaddyNoSnippets tests snippets title when entry has no snippets
func TestRenderSnippetsPanelTitle_TabCaddyNoSnippets(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCaddy
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target:  "localhost",
				Port:    8080,
				Imports: []string{},
			},
		},
	}
	m.cursor = 0

	title := m.renderSnippetsPanelTitle()
	expected := "Applied Snippets (0)"
	if title != expected {
		t.Errorf("Expected '%s', got '%s'", expected, title)
	}
}

// TestRenderSnippetsPanelTitle_TabCaddyNoCaddyConfig tests snippets title when no Caddy config
func TestRenderSnippetsPanelTitle_TabCaddyNoCaddyConfig(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCaddy
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusOrphanedDNS,
			Caddy: &caddy.CaddyEntry{
				Target:  "localhost",
				Port:    8080,
				Imports: []string{},
			},
		},
	}
	m.cursor = 0

	title := m.renderSnippetsPanelTitle()
	expected := "Applied Snippets (0)"
	if title != expected {
		t.Errorf("Expected '%s', got '%s'", expected, title)
	}
}

// TestRenderSnippetsPanelTitle_TabCloudflareWithSnippets tests title on Cloudflare tab
func TestRenderSnippetsPanelTitle_TabCloudflareWithSnippets(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCloudflare
	m.snippets = []caddy.Snippet{
		{
			Name:       "oauth",
			Category:   caddy.SnippetOAuthHeaders,
			Content:    "# OAuth config",
			LineStart:  10,
			LineEnd:    15,
		},
		{
			Name:       "rate-limit",
			Category:   caddy.SnippetPerformance,
			Content:    "# Rate limit",
			LineStart:  20,
			LineEnd:    25,
		},
		{
			Name:       "headers",
			Category:   caddy.SnippetSecurityHeaders,
			Content:    "# Security headers",
			LineStart:  30,
			LineEnd:    35,
		},
	}
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
		},
	}
	m.cursor = 0

	title := m.renderSnippetsPanelTitle()
	expected := "Available Snippets (3)"
	if title != expected {
		t.Errorf("Expected '%s', got '%s'", expected, title)
	}
}

// TestRenderSnippetsPanelTitle_SingleCategorySnippets tests title with single category
func TestRenderSnippetsPanelTitle_SingleCategorySnippets(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCloudflare
	m.snippets = []caddy.Snippet{
		{
			Name:       "oauth",
			Category:   caddy.SnippetOAuthHeaders,
			Content:    "# OAuth",
			LineStart:  1,
			LineEnd:    5,
		},
		{
			Name:       "jwt",
			Category:   caddy.SnippetOAuthHeaders,
			Content:    "# JWT",
			LineStart:  10,
			LineEnd:    15,
		},
	}
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
		},
	}
	m.cursor = 0

	title := m.renderSnippetsPanelTitle()
	expected := "Snippets (2 OAuth Headers)"
	if title != expected {
		t.Errorf("Expected '%s', got '%s'", expected, title)
	}
}

// TestRenderSnippetsPanelTitle_NoSnippets tests title with no snippets available
func TestRenderSnippetsPanelTitle_NoSnippets(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCloudflare
	m.snippets = []caddy.Snippet{}
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
		},
	}
	m.cursor = 0

	title := m.renderSnippetsPanelTitle()
	expected := "Snippets (0)"
	if title != expected {
		t.Errorf("Expected '%s', got '%s'", expected, title)
	}
}

// TestRenderSnippetsPanelTitle_TabSwitchAwareness tests tab awareness
func TestRenderSnippetsPanelTitle_TabSwitchAwareness(t *testing.T) {
	m := createTestModel()
	m.snippets = []caddy.Snippet{
		{
			Name:       "oauth",
			Category:   caddy.SnippetOAuthHeaders,
			Content:    "# OAuth",
			LineStart:  1,
			LineEnd:    5,
		},
	}
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target:  "localhost",
				Port:    8080,
				Imports: []string{"oauth"},
			},
		},
	}
	m.cursor = 0

	// Start on Cloudflare tab - should show all snippets
	m.activeTab = TabCloudflare
	title := m.renderSnippetsPanelTitle()
	if title != "Snippets (1 OAuth Headers)" {
		t.Errorf("TabCloudflare: expected 'Snippets (1 OAuth Headers)', got '%s'", title)
	}

	// Switch to Caddy tab - should show applied snippets only
	m.activeTab = TabCaddy
	title = m.renderSnippetsPanelTitle()
	expected := "Applied Snippets (1)"
	if title != expected {
		t.Errorf("TabCaddy: expected '%s', got '%s'", expected, title)
	}
}

// TestRenderSnippetsPanelTitle_CursorBeyondList tests behavior when cursor is beyond entries
func TestRenderSnippetsPanelTitle_CursorBeyondList(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCaddy
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target:  "localhost",
				Port:    8080,
				Imports: []string{"oauth"},
			},
		},
	}
	m.cursor = 10 // Cursor beyond entries list

	title := m.renderSnippetsPanelTitle()
	// Should fall back to showing all snippets when no entry selected
	if title != "Snippets (0)" {
		t.Errorf("Expected 'Snippets (0)' when cursor beyond list, got '%s'", title)
	}
}

// TestRenderSnippetsPanelTitle_MultiCategorySnippets tests title with multiple categories
func TestRenderSnippetsPanelTitle_MultiCategorySnippets(t *testing.T) {
	m := createTestModel()
	m.activeTab = TabCloudflare
	m.snippets = []caddy.Snippet{
		{
			Name:       "oauth",
			Category:   caddy.SnippetOAuthHeaders,
			Content:    "# OAuth",
			LineStart:  1,
			LineEnd:    5,
		},
		{
			Name:       "rate-limit",
			Category:   caddy.SnippetPerformance,
			Content:    "# Rate limit",
			LineStart:  10,
			LineEnd:    15,
		},
	}
	m.entries = []diff.SyncedEntry{
		{
			Domain: "example.com",
			Status: diff.StatusSynced,
		},
	}
	m.cursor = 0

	title := m.renderSnippetsPanelTitle()
	// Multiple categories should use "Available Snippets"
	expected := "Available Snippets (2)"
	if title != expected {
		t.Errorf("Expected '%s', got '%s'", expected, title)
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

// TestRenderSnippetsPanelTitle_ComplexScenario tests a complex real-world scenario
func TestRenderSnippetsPanelTitle_ComplexScenario(t *testing.T) {
	m := createTestModel()

	// Set up multiple entries with different configurations
	m.entries = []diff.SyncedEntry{
		{
			Domain: "api.example.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target:  "10.0.0.1",
				Port:    3000,
				Imports: []string{"oauth", "rate-limit", "cors"},
			},
		},
		{
			Domain: "web.example.com",
			Status: diff.StatusSynced,
			Caddy: &caddy.CaddyEntry{
				Target:  "10.0.0.2",
				Port:    3001,
				Imports: []string{"security-headers"},
			},
		},
	}

	m.snippets = []caddy.Snippet{
		{Name: "oauth", Category: caddy.SnippetOAuthHeaders, Content: "# OAuth", LineStart: 1, LineEnd: 5},
		{Name: "rate-limit", Category: caddy.SnippetPerformance, Content: "# Rate", LineStart: 6, LineEnd: 10},
		{Name: "cors", Category: caddy.SnippetCORS, Content: "# CORS", LineStart: 11, LineEnd: 15},
		{Name: "security-headers", Category: caddy.SnippetSecurityHeaders, Content: "# Headers", LineStart: 16, LineEnd: 20},
	}

	// Test on Caddy tab with first entry selected
	m.activeTab = TabCaddy
	m.cursor = 0
	title := m.renderSnippetsPanelTitle()
	if title != "Applied Snippets (3)" {
		t.Errorf("Expected 'Applied Snippets (3)', got '%s'", title)
	}

	// Switch to second entry
	m.cursor = 1
	title = m.renderSnippetsPanelTitle()
	if title != "Applied Snippets (1)" {
		t.Errorf("Expected 'Applied Snippets (1)', got '%s'", title)
	}

	// Switch to Cloudflare tab
	m.activeTab = TabCloudflare
	title = m.renderSnippetsPanelTitle()
	expected := "Available Snippets (4)"
	if title != expected {
		t.Errorf("Expected '%s', got '%s'", expected, title)
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
