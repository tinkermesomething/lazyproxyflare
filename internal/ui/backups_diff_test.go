package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func writeDiffFiles(t *testing.T, current, backup string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	cur := filepath.Join(dir, "Caddyfile")
	bak := filepath.Join(dir, "Caddyfile.backup.20260101_000000")
	if err := os.WriteFile(cur, []byte(current), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bak, []byte(backup), 0600); err != nil {
		t.Fatal(err)
	}
	return cur, bak
}

// The old set-based diff reported "identical" whenever both files had the same
// set of lines, even if a block was duplicated or moved. Restore previews must not lie.
func TestGenerateDiffDetectsReorderedAndRepeatedLines(t *testing.T) {
	tests := []struct {
		name            string
		current, backup string
	}{
		{"duplicated closing brace", "a\nb\n}\n", "a\n}\nb\n}\n"},
		{"moved line", "one\ntwo\nthree\n", "one\nthree\ntwo\n"},
		{"repeated import removed", "import x\nimport x\nimport y\n", "import x\nimport y\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cur, bak := writeDiffFiles(t, tt.current, tt.backup)
			lines, err := GenerateDiff(cur, bak)
			if err != nil {
				t.Fatal(err)
			}
			changed := 0
			for _, l := range lines {
				if l.Type == "added" || l.Type == "removed" {
					changed++
				}
			}
			if changed == 0 {
				t.Errorf("files differ but diff reports no changes: %+v", lines)
			}
		})
	}
}

func TestGenerateDiffIdenticalFiles(t *testing.T) {
	cur, bak := writeDiffFiles(t, "a\nb\n", "a\nb\n")
	lines, err := GenerateDiff(cur, bak)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0].Type != "same" {
		t.Errorf("expected single 'same' line, got %+v", lines)
	}
}
