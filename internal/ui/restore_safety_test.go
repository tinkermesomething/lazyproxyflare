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
