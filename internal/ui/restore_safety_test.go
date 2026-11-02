package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lazyproxyflare/internal/config"
)

func TestRestoreDefaultsToCaddyfileOnly(t *testing.T) {
	m := mgrModel(t)
	m.currentView = ViewBackupManager
	p := m.config.Caddy.CaddyfilePath
	if err := os.WriteFile(p+".backup.20260101_000000", []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, _ = m.handleRestoreBackup()
	if m.currentView != ViewRestoreScope || m.backup.RestoreScope != RestoreCaddyOnly || m.backup.RestoreScopeCursor != int(RestoreCaddyOnly) {
		t.Errorf("view %v scope %v cursor %d", m.currentView, m.backup.RestoreScope, m.backup.RestoreScopeCursor)
	}
}

// A restore that fails validation must put the previous live file back.
func TestRestoreRollsBackWhenValidationFails(t *testing.T) {
	dir := t.TempDir()
	live := filepath.Join(dir, "Caddyfile")
	old := filepath.Join(dir, "Caddyfile.backup.hand-named")
	os.WriteFile(live, []byte("live content\n"), 0o644)
	os.WriteFile(old, []byte("restored content\n"), 0o644)

	cfg := &config.Config{}
	cfg.Caddy.CaddyfilePath = live
	cfg.Caddy.ValidationCommand = "false"

	err := restoreCaddyfileSafely(cfg, old)
	if err == nil || !strings.Contains(err.Error(), "backup restored") {
		t.Fatalf("want rollback error, got %v", err)
	}
	if got, _ := os.ReadFile(live); string(got) != "live content\n" {
		t.Errorf("live file = %q, want the pre-restore content", got)
	}
	if got, _ := os.ReadFile(old); string(got) != "restored content\n" {
		t.Error("the backup being restored was modified")
	}
}

func TestStaleRefreshFromPreviousProfileIsIgnored(t *testing.T) {
	m := mgrModel(t)
	m.config.Cloudflare.ZoneID = "zone-b"
	m.loading = true
	m.entries = nil

	stale := refreshCompleteMsg{source: "zone-a|" + m.config.Caddy.CaddyfilePath}
	stale.snippets = m.snippets
	updated, _ := m.Update(stale)
	got := updated.(Model)
	if len(got.snippets) != len(m.snippets) && len(got.snippets) != 0 {
		t.Fatal("unexpected snippet change")
	}
	if !got.loading {
		t.Error("a stale refresh must not end the current profile's loading state")
	}

	fresh := refreshCompleteMsg{source: refreshSourceOf(m.config), snippets: m.snippets}
	updated, _ = m.Update(fresh)
	if updated.(Model).loading {
		t.Error("the matching refresh must end loading")
	}
}

func TestRefreshDoesNotFallBackToTheWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	os.WriteFile(filepath.Join(dir, "Caddyfile"), []byte("wrong.example.com {\n}\n"), 0o644)

	cfg := &config.Config{}
	cfg.Caddy.CaddyfilePath = filepath.Join(dir, "does-not-exist", "Caddyfile")
	cfg.Cloudflare.APIToken = "t"
	msg := fetchData(cfg)
	if msg.err == nil || !strings.Contains(msg.err.Error(), "cannot read the Caddyfile") {
		t.Errorf("want a clear read error, got %v", msg.err)
	}
	if len(msg.entries) != 0 {
		t.Error("entries were loaded from the wrong file")
	}
}
