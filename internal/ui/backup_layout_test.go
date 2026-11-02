package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/config"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

var backupBase = time.Date(2026, 1, 1, 12, 0, 0, 0, time.Local)

// backupStamp is the timestamp shown for backup i (0 = newest).
func backupStamp(i int) string {
	return backupBase.Add(-time.Duration(i) * time.Minute).Format("2006-01-02 15:04:05")
}

// newBackupModel creates a Caddyfile plus n backups with distinct mtimes in a temp dir.
func newBackupModel(t *testing.T, n, width, height int) Model {
	t.Helper()
	dir := t.TempDir()
	caddyfile := filepath.Join(dir, "Caddyfile")
	if err := os.WriteFile(caddyfile, []byte("current\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		p := fmt.Sprintf("%s.backup.%04d", caddyfile, i)
		if err := os.WriteFile(p, []byte("backup\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, backupBase, backupBase.Add(-time.Duration(i)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	m := createTestModel()
	m.config = &config.Config{Caddy: config.CaddyConfig{CaddyfilePath: caddyfile}}
	m.width, m.height = width, height
	m.currentView = ViewBackupManager
	return m
}

func renderBackupManagerModal(m Model) string {
	return stripANSI(RenderModalOverlayLarge("", "Backup Manager", m.renderBackupManagerContent(), m.width, m.height))
}

var screenSizes = []struct{ w, h int }{{179, 49}, {120, 30}, {80, 24}}

// The highlighted backup must always be on screen while moving through the list.
func TestBackupManagerCursorStaysVisible(t *testing.T) {
	for _, sz := range screenSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			const n = 60
			m := newBackupModel(t, n, sz.w, sz.h)

			check := func(dir string) {
				want := "→ " + backupStamp(m.backup.Cursor)
				if out := renderBackupManagerModal(m); !strings.Contains(out, want) {
					t.Fatalf("%s: cursor %d (offset %d) not visible, want %q in:\n%s", dir, m.backup.Cursor, m.backup.ScrollOffset, want, out)
				}
			}

			check("start")
			for i := 0; i < n-1; i++ {
				m, _ = m.handleNavigateDown()
				check("down")
			}
			if m.backup.Cursor != n-1 {
				t.Fatalf("cursor = %d, want %d", m.backup.Cursor, n-1)
			}
			for i := 0; i < n-1; i++ {
				m, _ = m.handleNavigateUp()
				check("up")
			}
			if m.backup.Cursor != 0 {
				t.Fatalf("cursor = %d, want 0", m.backup.Cursor)
			}
		})
	}
}

func TestBackupManagerMouseWheelKeepsCursorVisible(t *testing.T) {
	const n = 60
	m := newBackupModel(t, n, 179, 49)
	for i := 0; i < n-1; i++ {
		m, _ = m.handleMouseMsg(tea.MouseMsg{Type: tea.MouseWheelDown})
		want := "→ " + backupStamp(m.backup.Cursor)
		if out := renderBackupManagerModal(m); !strings.Contains(out, want) {
			t.Fatalf("wheel down: cursor %d not visible", m.backup.Cursor)
		}
	}
}

// A click must select the row that is drawn under the pointer.
func TestBackupManagerClickSelectsRowUnderPointer(t *testing.T) {
	for _, sz := range screenSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := newBackupModel(t, 60, sz.w, sz.h)
			lines := strings.Split(renderBackupManagerModal(m), "\n")
			for _, idx := range []int{0, 3, 7} {
				y := -1
				for i, l := range lines {
					if strings.Contains(l, backupStamp(idx)) {
						y = i
					}
				}
				if y < 0 {
					continue // row not on this screen
				}
				clicked, _ := m.handleMouseMsg(tea.MouseMsg{Type: tea.MouseLeft, X: sz.w / 2, Y: y})
				if clicked.backup.Cursor != idx {
					t.Errorf("click on screen row %d (backup %d) selected backup %d", y, idx, clicked.backup.Cursor)
				}
			}
		})
	}
}

// Scrolling the preview must stop at the end of the diff, not the end of the file.
func TestBackupPreviewScrollStopsAtEndOfDiff(t *testing.T) {
	for _, sz := range screenSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			dir := t.TempDir()
			cur, bak := filepath.Join(dir, "Caddyfile"), filepath.Join(dir, "Caddyfile.backup.1")
			var curLines, bakLines []string
			for i := 0; i < 400; i++ {
				curLines = append(curLines, fmt.Sprintf("line %d", i))
				bakLines = append(bakLines, fmt.Sprintf("line %d", i))
			}
			for _, i := range []int{20, 200, 380} { // far apart: 3 hunks
				bakLines[i] = fmt.Sprintf("CHANGED %d", i)
			}
			os.WriteFile(cur, []byte(strings.Join(curLines, "\n")), 0600)
			os.WriteFile(bak, []byte(strings.Join(bakLines, "\n")), 0600)

			m := createTestModel()
			m.config = &config.Config{Caddy: config.CaddyConfig{CaddyfilePath: cur}}
			m.width, m.height = sz.w, sz.h
			m.currentView = ViewBackupPreview
			m.backup.PreviewPath = bak

			diff, err := GenerateDiff(cur, bak)
			if err != nil {
				t.Fatal(err)
			}
			rows := m.backupPreviewRows()
			wantMax := max(len(diff)-rows, 0)

			for i := 0; i < 1000; i++ {
				m, _ = m.handleNavigateDown()
			}
			if m.backup.PreviewScroll != wantMax {
				t.Errorf("after scrolling down PreviewScroll = %d, want %d (diff has %d lines, %d visible)", m.backup.PreviewScroll, wantMax, len(diff), rows)
			}

			out := stripANSI(RenderModalOverlayLarge("", "Backup Preview", m.renderBackupPreviewContent(), m.width, m.height))
			if !strings.Contains(out, "+ CHANGED 380") {
				t.Errorf("last change not visible after scrolling to the end:\n%s", out)
			}

			// Page down + top + bottom must agree with the same limits.
			m.backup.PreviewScroll = 0
			for i := 0; i < 100; i++ {
				m, _ = m.handleBackupPageDown()
			}
			if m.backup.PreviewScroll != wantMax {
				t.Errorf("after page down PreviewScroll = %d, want %d", m.backup.PreviewScroll, wantMax)
			}
			m.backup.PreviewScroll = 0
			m, _ = m.handleGoToBottom()
			if m.backup.PreviewScroll != wantMax {
				t.Errorf("after G PreviewScroll = %d, want %d", m.backup.PreviewScroll, wantMax)
			}
		})
	}
}

func TestFitLine(t *testing.T) {
	tests := []struct {
		name, in string
		width    int
		want     string
	}{
		{"short unchanged", "abc", 10, "abc"},
		{"exact fit unchanged", "abcde", 5, "abcde"},
		{"long is cut with ellipsis", "abcdefghij", 5, "abcd…"},
		{"tabs expand to spaces", "\tx", 10, "    x"},
		{"multibyte not split", "ééééé", 3, "éé…"},
		{"zero width", "abc", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fitLine(tt.in, tt.width); got != tt.want {
				t.Errorf("fitLine(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
			}
		})
	}
}

// A very long changed line must not push the diff box past the modal and hide its right border.
func TestBackupPreviewBoxStaysInsideModal(t *testing.T) {
	for _, sz := range screenSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			dir := t.TempDir()
			cur, bak := filepath.Join(dir, "Caddyfile"), filepath.Join(dir, "Caddyfile.backup.1")
			os.WriteFile(cur, []byte("a\nshort\nb\n"), 0600)
			os.WriteFile(bak, []byte("a\n"+strings.Repeat("x", 600)+"\nb\n"), 0600)

			m := createTestModel()
			m.config = &config.Config{Caddy: config.CaddyConfig{CaddyfilePath: cur}}
			m.width, m.height = sz.w, sz.h
			m.currentView = ViewBackupPreview
			m.backup.PreviewPath = bak

			out := stripANSI(RenderModalOverlayLarge("", "Backup Preview", m.renderBackupPreviewContent(), m.width, m.height))
			if !regexp.MustCompile(`╯\s*│`).MatchString(out) {
				t.Errorf("inner box right border is missing:\n%s", out)
			}
			if !strings.Contains(out, "…") {
				t.Errorf("long line was not shortened with an ellipsis:\n%s", out)
			}
		})
	}
}
