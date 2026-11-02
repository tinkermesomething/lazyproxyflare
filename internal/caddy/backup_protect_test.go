package caddy

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupNeverTouchesHandNamedBackups(t *testing.T) {
	dir := t.TempDir()
	cf := filepath.Join(dir, "Caddyfile")
	os.WriteFile(cf, []byte("x"), 0o644)

	hand := cf + ".backup.2026-11-01_pre-open"
	auto := []string{cf + ".backup.20250101_000000", cf + ".backup.20250102_000000", cf + ".backup.20250103_000000"}
	os.WriteFile(hand, []byte("keep me"), 0o644)
	old := time.Now().Add(-400 * 24 * time.Hour)
	os.Chtimes(hand, old, old)
	for i, p := range auto {
		os.WriteFile(p, []byte("auto"), 0o644)
		ts := old.Add(time.Duration(i) * time.Hour)
		os.Chtimes(p, ts, ts)
	}

	if got, _ := GetOldBackups(cf, time.Hour); len(got) != 3 {
		t.Errorf("age preview lists %d, want 3 (hand-named excluded)", len(got))
	}
	if n, _ := CleanupByCount(cf, 1); n != 2 {
		t.Errorf("count cleanup deleted %d, want 2", n)
	}
	if n, _ := CleanupBySize(cf, 1); n != 0 {
		t.Errorf("size cleanup deleted %d, want 0 (under limit)", n)
	}
	if err := CleanupOldBackups(cf, time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(hand); err != nil {
		t.Error("hand-named backup was deleted")
	}
	left, _ := ListBackups(cf)
	if len(left) != 1 || left[0].Path != hand || left[0].Managed {
		t.Errorf("left = %+v, want only the hand-named backup, unmanaged", left)
	}
}

func TestIsManagedBackup(t *testing.T) {
	for p, want := range map[string]bool{
		"/d/Caddyfile.backup.20260102_030405":      true,
		"/d/Caddyfile.backup.2026-11-01_pre-open":  false,
		"/d/Caddyfile.backup.20260102_030405.keep": false,
		"/d/Caddyfile.backup.pre-restore":          false,
	} {
		if IsManagedBackup(p) != want {
			t.Errorf("%s: want %v", p, want)
		}
	}
}
