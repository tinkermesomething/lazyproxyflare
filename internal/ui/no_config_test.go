package ui

import (
	"strings"
	"testing"
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
