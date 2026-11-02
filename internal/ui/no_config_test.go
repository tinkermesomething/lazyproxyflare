package ui

import (
	"strings"
	"testing"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/diff"
)

// When several profiles exist the app opens on the profile selector with no
// config loaded. Modals opened from there (delete, export, import, errors) are
// drawn over a base view, which must not need a config.
func TestModalsOverProfileSelectorWithoutConfig(t *testing.T) {
	views := map[ViewMode]string{
		ViewConfirmDeleteProfile: "Delete Profile",
		ViewExportResult:         "Export Profile",
		ViewConfirmImport:        "Import Profile",
	}
	t.Setenv("HOME", t.TempDir()) // the model creates an audit logger under $HOME
	for view, title := range views {
		m := NewModelWithProfileSelector([]string{"a", "b"}, "a")
		m.width, m.height = 100, 30
		m.currentView = view
		m.profile.DeleteProfileName = "b"

		out := stripANSI(m.renderView())
		if !strings.Contains(out, title) {
			t.Errorf("view %v: modal %q not rendered", view, title)
		}
	}
}

// Editing an entry whose Caddy block is custom must warn that only DNS changes
// are saved, and must stay quiet for blocks the app can regenerate exactly.
func TestEditFormWarnsAboutCustomCaddyBlock(t *testing.T) {
	custom := &caddy.CaddyEntry{
		Domain: "mail.example.com", Domains: []string{"mail.example.com"},
		RawBlock: "mail.example.com {\n\treverse_proxy 10.0.0.9:80 {\n\t\theader_up X-Real-IP {remote_host}\n\t}\n}",
	}
	plain := &caddy.CaddyEntry{
		Domain: "app.example.com", Domains: []string{"app.example.com"},
		RawBlock: "app.example.com {\n\timport log_access\n\treverse_proxy 10.0.0.9:80\n}",
	}

	m := createTestModel()
	m.currentView = ViewEdit

	m.editingEntry = &diff.SyncedEntry{Domain: custom.Domain, Caddy: custom}
	out := stripANSI(m.renderEditFormContent())
	if !strings.Contains(out, "Custom Caddy config") || !strings.Contains(out, "X-Real-IP") {
		t.Errorf("no warning for a custom block:\n%.300s", out)
	}

	m.addForm.DNSOnly = true // switching to DNS-only removes the block on purpose
	if strings.Contains(stripANSI(m.renderEditFormContent()), "Custom Caddy config") {
		t.Error("warning shown while switching to DNS-only")
	}
	m.addForm.DNSOnly = false

	m.editingEntry = &diff.SyncedEntry{Domain: plain.Domain, Caddy: plain}
	if strings.Contains(stripANSI(m.renderEditFormContent()), "Custom Caddy config") {
		t.Error("warning shown for a block the app can regenerate")
	}
}
