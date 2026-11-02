package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
)

const mgrCaddyfile = `# LAN only.
(ip_restricted) {
	respond "404" 404
}

(log_access) {
	log
}

# Never imported
(orphan) {
	encode gzip
}

(chain) {
	import log_access
}

app.example.com {
	import ip_restricted
	reverse_proxy 10.0.0.2:80
}
`

func mgrModel(t *testing.T) Model {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Caddyfile")
	if err := os.WriteFile(path, []byte(mgrCaddyfile), 0o644); err != nil {
		t.Fatal(err)
	}
	parsed := caddy.ParseCaddyfileWithSnippets(mgrCaddyfile)
	m := createTestModel()
	m.config = &config.Config{Domain: "example.com"}
	m.config.Caddy.CaddyfilePath = path
	m.snippets = parsed.Snippets
	for _, e := range parsed.Entries {
		e := e
		m.entries = append(m.entries, diff.SyncedEntry{Domain: e.Domain, Caddy: &e})
	}
	m.width, m.height = 140, 40
	return m
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		var msg tea.KeyPressMsg
		switch k {
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "esc":
			msg = tea.KeyPressMsg{Code: tea.KeyEscape}
		case "down":
			msg = tea.KeyPressMsg{Code: tea.KeyDown}
		case "up":
			msg = tea.KeyPressMsg{Code: tea.KeyUp}
		case "ctrl+s":
			msg = tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
		default:
			msg = tea.KeyPressMsg{Code: rune(k[0]), Text: k}
		}
		m, _ = m.handleKeyMsg(msg)
	}
	return m
}

func TestWOpensTheManagerAndEscCloses(t *testing.T) {
	m := press(mgrModel(t), "w")
	if m.currentView != ViewSnippetManager {
		t.Fatalf("w opened %v", m.currentView)
	}
	if m = press(m, "esc"); m.currentView != ViewList {
		t.Errorf("esc went to %v", m.currentView)
	}
}

func TestManagerCursorStopsAtTheNewRow(t *testing.T) {
	m := press(mgrModel(t), "w")
	for i := 0; i < 20; i++ {
		m = press(m, "down")
	}
	if m.snippetPanel.Cursor != len(m.snippets) { // 4 snippets: rows 0-3, new row 4
		t.Errorf("cursor = %d, want %d (the new-snippet row)", m.snippetPanel.Cursor, len(m.snippets))
	}
	if m = press(m, "g"); m.snippetPanel.Cursor != 0 {
		t.Errorf("g: cursor = %d", m.snippetPanel.Cursor)
	}
	if m = press(m, "G"); m.snippetPanel.Cursor != len(m.snippets) {
		t.Errorf("G: cursor = %d", m.snippetPanel.Cursor)
	}
}

func TestDetailShowsUsersAndDeleteIsGuarded(t *testing.T) {
	m := press(mgrModel(t), "w", "enter") // ip_restricted, imported by app.example.com
	if m.currentView != ViewSnippetDetail || m.snippetPanel.DetailName != "ip_restricted" {
		t.Fatalf("view %v name %q", m.currentView, m.snippetPanel.DetailName)
	}
	out := ansi.Strip(m.renderSnippetDetailContent())
	for _, want := range []string{"ip_restricted", "LAN only.", "app.example.com", `respond "404" 404`} {
		if !strings.Contains(out, want) {
			t.Errorf("detail missing %q:\n%s", want, out)
		}
	}
	m = press(m, "d")
	if m.snippetPanel.ConfirmDelete || !strings.Contains(m.snippetPanel.Notice, "Cannot delete") {
		t.Errorf("a used snippet must not be deletable: confirm=%v notice=%q", m.snippetPanel.ConfirmDelete, m.snippetPanel.Notice)
	}

	// log_access is imported only by another snippet (chain): also protected.
	m = press(m, "esc", "down", "enter", "d")
	if m.snippetPanel.DetailName != "log_access" || m.snippetPanel.ConfirmDelete {
		t.Errorf("snippet imported by a snippet must be protected (name %q, confirm %v)", m.snippetPanel.DetailName, m.snippetPanel.ConfirmDelete)
	}

	// orphan is unused: d asks, any key but y cancels.
	m = press(m, "esc", "down", "enter", "d")
	if m.snippetPanel.DetailName != "orphan" || !m.snippetPanel.ConfirmDelete {
		t.Fatalf("expected a delete prompt for orphan (name %q)", m.snippetPanel.DetailName)
	}
	if m = press(m, "x"); m.snippetPanel.ConfirmDelete || m.snippetPanel.Busy {
		t.Error("a key other than y must cancel the delete")
	}
}

// The old editor swallowed y, d and Enter. Every key must reach the textarea.
func TestEditorReceivesEveryKey(t *testing.T) {
	m := press(mgrModel(t), "w", "down", "down", "e") // orphan
	if !m.snippetPanel.Editing {
		t.Fatal("e did not start editing")
	}
	before := m.snippetPanel.EditTextarea.Value()
	m = press(m, "y", "d", "enter", "q", "w")
	got := m.snippetPanel.EditTextarea.Value()
	if !strings.HasSuffix(got, "yd\nqw") || !strings.HasPrefix(got, before) {
		t.Errorf("typed keys lost: %q", got)
	}
	updated, _ := m.Update(tea.PasteMsg{Content: " pasted"})
	if v := updated.(Model).snippetPanel.EditTextarea.Value(); !strings.HasSuffix(v, "qw pasted") {
		t.Errorf("paste lost: %q", v)
	}
	m = press(m, "esc")
	if m.snippetPanel.Editing || m.currentView != ViewSnippetDetail {
		t.Error("esc must leave the editor but stay in the detail view")
	}
}

func TestSavingWithoutChangesDoesNothing(t *testing.T) {
	m := press(mgrModel(t), "w", "down", "down", "e", "ctrl+s")
	if m.snippetPanel.Busy || m.snippetPanel.Notice != "No changes to save." {
		t.Errorf("busy=%v notice=%q", m.snippetPanel.Busy, m.snippetPanel.Notice)
	}
}

func TestNewSnippetRowStartsTheWizardAndEscComesBack(t *testing.T) {
	m := press(mgrModel(t), "w", "G", "enter")
	if m.currentView != ViewSnippetWizard {
		t.Fatalf("view = %v", m.currentView)
	}
	if m = press(m, "esc"); m.currentView != ViewSnippetManager {
		t.Errorf("esc from the wizard went to %v, want the manager", m.currentView)
	}
	if m = press(m, "n", "esc"); m.currentView != ViewSnippetManager {
		t.Errorf("n then esc ended on %v", m.currentView)
	}
	// Opened from the list, the wizard still returns to the list.
	m = mgrModel(t)
	m, _ = m.handleOpenSnippetWizard()
	if m = press(m, "esc"); m.currentView != ViewList {
		t.Errorf("wizard opened from the list returned to %v", m.currentView)
	}
}

func TestSnippetOpResults(t *testing.T) {
	m := press(mgrModel(t), "w", "down", "down", "e")
	fresh := caddy.ParseCaddyfileWithSnippets("(orphan) {\n\tencode zstd\n}\n").Snippets

	failed := m.handleSnippetOp(snippetOpMsg{op: "edit", name: "orphan", err: errors.New("validation failed (backup restored)")})
	if !failed.snippetPanel.Editing || !strings.Contains(failed.snippetPanel.Notice, "validation failed") {
		t.Error("a failed save must keep the editor open with the error")
	}

	saved := m.handleSnippetOp(snippetOpMsg{op: "edit", name: "orphan", backupPath: "/x/Caddyfile.backup.1", snippets: fresh})
	if saved.snippetPanel.Editing || !strings.HasPrefix(saved.snippetPanel.Notice, "✓ Saved orphan") || len(saved.snippets) != 1 {
		t.Errorf("save result: editing=%v notice=%q snippets=%d", saved.snippetPanel.Editing, saved.snippetPanel.Notice, len(saved.snippets))
	}

	m.snippetPanel.Editing = false
	m.snippetPanel.Cursor = 3
	deleted := m.handleSnippetOp(snippetOpMsg{op: "delete", name: "orphan", snippets: fresh})
	if deleted.currentView != ViewSnippetManager || deleted.snippetPanel.Cursor > len(fresh) {
		t.Errorf("delete result: view %v cursor %d", deleted.currentView, deleted.snippetPanel.Cursor)
	}
}

// A refused change must leave the Caddyfile byte-identical and take no backup.
func TestApplyCaddyfileChangeRefusalLeavesNoTrace(t *testing.T) {
	m := mgrModel(t)
	path := m.config.Caddy.CaddyfilePath
	_, err := applyCaddyfileChange(m.config, func(src string) (string, error) {
		return caddy.ReplaceSnippetBody(src, "orphan", "\tstale", "\tx\n")
	})
	if err == nil {
		t.Fatal("expected a refusal")
	}
	got, _ := os.ReadFile(path)
	if string(got) != mgrCaddyfile {
		t.Error("Caddyfile changed")
	}
	if matches, _ := filepath.Glob(path + ".backup.*"); len(matches) != 0 {
		t.Errorf("backups created: %v", matches)
	}
}

// Both views must fit their modal at every size, with many snippets.
func TestSnippetViewsFitTheModal(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {100, 30}, {179, 49}} {
		m := mgrModel(t)
		for i := 0; i < 40; i++ {
			m.snippets = append(m.snippets, caddy.Snippet{Name: fmt.Sprintf("extra_snippet_%02d", i), Comment: strings.Repeat("long comment ", 20), Content: strings.Repeat("\tdirective arg arg arg arg arg arg arg arg arg arg arg arg\n", 60)})
		}
		m.width, m.height = size[0], size[1]
		_, mh := modalDimensions(m.width, m.height, true)
		maxLines, maxWidth := mh-2, m.snippetTextWidth()

		m = press(m, "w", "G", "up", "up")
		check := func(what, out string) {
			lines := strings.Split(ansi.Strip(out), "\n")
			if len(lines) > maxLines {
				t.Errorf("%dx%d %s: %d lines, modal holds %d", size[0], size[1], what, len(lines), maxLines)
			}
			for _, l := range lines {
				if w := ansi.StringWidth(l); w > maxWidth {
					t.Errorf("%dx%d %s: line %d wide, max %d: %q", size[0], size[1], what, w, maxWidth, l)
				}
			}
		}
		check("list", m.renderSnippetManagerContent())

		m = press(m, "enter")
		check("detail", m.renderSnippetDetailContent())
		m = press(m, "G")
		check("detail scrolled", m.renderSnippetDetailContent())
		m = press(m, "e")
		check("editor", m.renderSnippetDetailContent())
	}
}

func TestSnippetManagerMouse(t *testing.T) {
	m := press(mgrModel(t), "w")
	updated, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if got := updated.(Model).snippetPanel.Cursor; got != 1 {
		t.Errorf("wheel down: cursor = %d", got)
	}
	updated, _ = updated.(Model).Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if got := updated.(Model).snippetPanel.Cursor; got != 0 {
		t.Errorf("wheel up: cursor = %d", got)
	}
}

func TestRetabRestoresLeadingTabsOnly(t *testing.T) {
	in := "    a\n        b x    y\nc\n  d"
	want := "\ta\n\t\tb x    y\nc\n  d"
	if got := retab(in); got != want {
		t.Errorf("retab = %q, want %q", got, want)
	}
}
