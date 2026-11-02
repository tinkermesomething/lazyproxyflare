package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"lazyproxyflare/internal/cloudflare"
	"lazyproxyflare/internal/diff"
)

// Tests for behavior that changed between Bubble Tea v1 and v2 and would
// otherwise fail silently.

func space() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "} }

// The key dispatcher switches on these strings. A library upgrade that renames
// one would make a key silently do nothing.
func TestKeyNamesMatchTheDispatcher(t *testing.T) {
	tests := []struct {
		msg  tea.KeyPressMsg
		want string
	}{
		{tea.KeyPressMsg{Code: tea.KeyEscape}, "esc"},
		{tea.KeyPressMsg{Code: tea.KeyEnter}, "enter"},
		{tea.KeyPressMsg{Code: tea.KeyTab}, "tab"},
		{tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}, "shift+tab"},
		{tea.KeyPressMsg{Code: tea.KeyBackspace}, "backspace"},
		{tea.KeyPressMsg{Code: tea.KeyUp}, "up"},
		{tea.KeyPressMsg{Code: tea.KeyDown}, "down"},
		{tea.KeyPressMsg{Code: tea.KeyLeft}, "left"},
		{tea.KeyPressMsg{Code: tea.KeyRight}, "right"},
		{tea.KeyPressMsg{Code: tea.KeyPgUp}, "pgup"},
		{tea.KeyPressMsg{Code: tea.KeyPgDown}, "pgdown"},
		{tea.KeyPressMsg{Code: tea.KeyHome}, "home"},
		{tea.KeyPressMsg{Code: tea.KeyEnd}, "end"},
		{space(), "space"},
		{tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}, "ctrl+c"},
		{tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl}, "ctrl+q"},
		{tea.KeyPressMsg{Code: 'w', Mod: tea.ModCtrl}, "ctrl+w"},
		{tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}, "ctrl+s"},
		{tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl}, "ctrl+b"},
		{tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl}, "ctrl+p"},
		{tea.KeyPressMsg{Code: '?', Text: "?"}, "?"},
		{tea.KeyPressMsg{Code: '@', Text: "@"}, "@"},
		{tea.KeyPressMsg{Code: 'E', Text: "E"}, "E"},
		{tea.KeyPressMsg{Code: 'x', Text: "x"}, "x"},
	}
	for _, tt := range tests {
		if got := tt.msg.String(); got != tt.want {
			t.Errorf("key %+v: String() = %q, dispatcher expects %q", tt.msg, got, tt.want)
		}
	}
}

// v1 reported the space bar as " "; v2 reports "space". Typing a space must
// still work in every hand-written text field.
func TestSpaceCanBeTypedIntoTextFields(t *testing.T) {
	typeKeys := func(m Model, keys ...tea.KeyPressMsg) Model {
		for _, k := range keys {
			m, _ = m.handleKeyMsg(k)
		}
		return m
	}
	a, b := tea.KeyPressMsg{Code: 'a', Text: "a"}, tea.KeyPressMsg{Code: 'b', Text: "b"}

	t.Run("list search", func(t *testing.T) {
		m := createTestModel()
		m.searching = true
		m = typeKeys(m, a, space(), b)
		if m.searchQuery != "a b" {
			t.Errorf("searchQuery = %q, want %q", m.searchQuery, "a b")
		}
	})
	t.Run("editor prompt", func(t *testing.T) {
		m := createTestModel()
		m.currentView = ViewSetEditor
		m = typeKeys(m, a, space(), b)
		if m.profile.EditorInput != "a b" {
			t.Errorf("EditorInput = %q, want %q", m.profile.EditorInput, "a b")
		}
	})
	t.Run("audit search", func(t *testing.T) {
		m := createTestModel()
		m.currentView = ViewAuditLog
		m.audit.SearchActive = true
		m = typeKeys(m, a, space(), b)
		if m.audit.SearchQuery != "a b" {
			t.Errorf("SearchQuery = %q, want %q", m.audit.SearchQuery, "a b")
		}
	})
	t.Run("import path", func(t *testing.T) {
		m := createTestModel()
		m.currentView = ViewConfirmImport
		m = typeKeys(m, a, space(), b)
		if m.profile.ImportPath != "a b" {
			t.Errorf("ImportPath = %q, want %q", m.profile.ImportPath, "a b")
		}
	})
}

// Non-printable keys carry no text, so they must not be typed as characters.
func TestNonPrintableKeysAreNotTyped(t *testing.T) {
	m := createTestModel()
	m.searching = true
	for _, k := range []tea.KeyPressMsg{
		{Code: tea.KeyUp}, {Code: tea.KeyF1}, {Code: 'a', Mod: tea.ModCtrl}, {Code: 'a', Mod: tea.ModAlt},
	} {
		m, _ = m.handleKeyMsg(k)
	}
	if m.searchQuery != "" {
		t.Errorf("non-printable keys were typed into the search: %q", m.searchQuery)
	}
}

// The space bar toggles selection in the list (the dispatcher case was " " in v1).
func TestSpaceTogglesSelectionInList(t *testing.T) {
	m := createTestModel()
	m.entries = []diff.SyncedEntry{{Domain: "a.example.com", Status: diff.StatusSynced, DNS: &cloudflare.DNSRecord{}}}
	m, _ = m.handleKeyMsg(space())
	if !m.selectedEntries["a.example.com"] {
		t.Errorf("space did not select the entry: %v", m.selectedEntries)
	}
	m, _ = m.handleKeyMsg(space())
	if m.selectedEntries["a.example.com"] {
		t.Errorf("second space did not deselect: %v", m.selectedEntries)
	}
}

// Paste is its own message in v2. The wizard's text input must still receive it.
func TestPasteReachesWizardTextInput(t *testing.T) {
	m := NewModelWithWizard()
	m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: tea.KeyEnter}) // welcome -> basic info
	if !m.isWizardTextInputStep() {
		t.Fatalf("expected a text input step, got %v", m.wizardStep)
	}

	m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: 'x', Text: "x"})
	updated, _ := m.Update(tea.PasteMsg{Content: "yz"})
	m = updated.(Model)
	if got := m.wizardTextInput.Value(); got != "xyz" {
		t.Errorf("wizard input = %q, want %q", got, "xyz")
	}

	// Typing a space into a bubbles input still works.
	m, _ = m.handleKeyMsg(space())
	m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: 'w', Text: "w"})
	if got := m.wizardTextInput.Value(); got != "xyz w" {
		t.Errorf("wizard input after space = %q, want %q", got, "xyz w")
	}
}

// Alt screen and mouse reporting are properties of the view in v2.
func TestViewRequestsAltScreenAndMouse(t *testing.T) {
	m := createTestModel()
	m.width, m.height = 100, 30
	v := m.View()
	if !v.AltScreen {
		t.Error("View must request the alternate screen")
	}
	if v.MouseMode != tea.MouseModeAllMotion {
		t.Errorf("MouseMode = %v, want all-motion (v1 parity)", v.MouseMode)
	}
}

// Wheel and click messages reach the mouse handlers through Update.
func TestMouseMessagesAreRouted(t *testing.T) {
	m := newBackupModel(t, 20, 179, 49)

	updated, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if got := updated.(Model).backup.Cursor; got != 1 {
		t.Errorf("wheel down: cursor = %d, want 1", got)
	}

	y := m.backupListTopY() + 3
	updated, _ = m.Update(tea.MouseClickMsg{Button: tea.MouseLeft, X: 90, Y: y})
	if got := updated.(Model).backup.Cursor; got != 3 {
		t.Errorf("click on row 3: cursor = %d, want 3", got)
	}

	// Right-click and release do nothing.
	updated, _ = m.Update(tea.MouseClickMsg{Button: tea.MouseRight, X: 90, Y: y})
	if got := updated.(Model).backup.Cursor; got != 0 {
		t.Errorf("right click changed the cursor to %d", got)
	}
}

// Space toggles the boolean fields in the profile editor.
func TestSpaceTogglesProfileEditBoolean(t *testing.T) {
	m := createTestModel()
	m.currentView = ViewProfileEdit
	m.profile.EditCursor = 13 // SSL
	before := m.profile.EditData.SSL
	m, _ = m.handleKeyMsg(space())
	if m.profile.EditData.SSL == before {
		t.Error("space did not toggle the SSL field")
	}
}

// Ctrl+J must insert a newline in the custom Caddy config field. The help
// screen promised Ctrl+M, which most terminals send as plain Enter.
func TestCustomConfigNewlineKey(t *testing.T) {
	m := createTestModel()
	m.currentView = ViewAdd
	m.addForm.FocusedField = 8 + len(m.snippets)

	m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: 'a', Text: "a"})
	m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: 'j', Mod: tea.ModCtrl})
	m, _ = m.handleKeyMsg(tea.KeyPressMsg{Code: 'b', Text: "b"})
	if got := m.addForm.CustomCaddyConfig; got != "a\nb" {
		t.Errorf("CustomCaddyConfig = %q, want %q", got, "a\nb")
	}
}
