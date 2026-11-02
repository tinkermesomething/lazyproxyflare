package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func footerText(m Model, width int) string {
	return stripANSI(m.getStatusBarContentFor(width))
}

// The layout reserves one line for the footer: it must never be wider than the terminal.
func TestFooterFitsOneLineAtEveryWidth(t *testing.T) {
	m := createTestModel()
	contexts := map[string]func(Model) Model{
		"list dns":       func(m Model) Model { return m },
		"list caddy":     func(m Model) Model { m.activeTab = TabCaddy; return m },
		"details panel":  func(m Model) Model { m.panelFocus = PanelFocusDetails; return m },
		"snippets panel": func(m Model) Model { m.panelFocus = PanelFocusSnippets; return m },
		"with filter":    func(m Model) Model { m.searchQuery = "plex"; return m },
		"with selection": func(m Model) Model { m.selectedEntries = map[string]bool{"a": true}; return m },
	}
	for name, setup := range contexts {
		for width := 40; width <= 200; width += 3 {
			got := footerText(setup(m), width)
			if w := lipgloss.Width(got); w > width-2 {
				t.Errorf("%s @%d: footer is %d wide (max %d): %q", name, width, w, width-2, got)
			}
			if strings.Contains(got, "\n") {
				t.Errorf("%s @%d: footer wraps: %q", name, width, got)
			}
			if !strings.Contains(got, "?") || !strings.Contains(got, "quit") {
				t.Errorf("%s @%d: help/quit must always be shown: %q", name, width, got)
			}
		}
	}
}

// Wide terminals lose nothing compared to the old fixed footer.
func TestFooterShowsEverythingWhenWide(t *testing.T) {
	m := createTestModel()
	got := footerText(m, 200)
	for _, want := range []string{"nav", "view", "add", "edit", "del", "snippets", "profile", "search", "refresh", "backups", "help", "quit"} {
		if !strings.Contains(got, want) {
			t.Errorf("wide footer is missing %q: %q", want, got)
		}
	}
}

// Narrow terminals drop low-priority hints first, never the core ones.
func TestFooterDropsLowPriorityFirst(t *testing.T) {
	m := createTestModel()
	got := footerText(m, 60)
	for _, keep := range []string{"nav", "help", "quit"} {
		if !strings.Contains(got, keep) {
			t.Errorf("60-col footer lost %q: %q", keep, got)
		}
	}
	if strings.Contains(got, "profile") {
		t.Errorf("60-col footer should have dropped 'profile' first: %q", got)
	}
}

func TestFooterReflectsContext(t *testing.T) {
	m := createTestModel()

	m.panelFocus = PanelFocusSnippets
	if got := footerText(m, 200); !strings.Contains(got, "Snippets") {
		t.Errorf("snippets focus not shown: %q", got)
	}
	if got := footerText(m, 200); strings.Contains(got, "add") {
		t.Errorf("snippets panel must not advertise entry add: %q", got)
	}

	m = createTestModel()
	m.searchQuery = "plex"
	if got := footerText(m, 200); !strings.Contains(got, "clear") {
		t.Errorf("active filter should show how to clear it: %q", got)
	}

	m = createTestModel()
	m.activeTab = TabCaddy
	if got := footerText(m, 200); !strings.Contains(got, "editor") {
		t.Errorf("caddy tab should offer editor: %q", got)
	}
	m.activeTab = TabCloudflare
	if got := footerText(m, 200); strings.Contains(got, "editor") {
		t.Errorf("dns tab must not offer editor: %q", got)
	}
}

func TestFooterAdvertisesCommandLogAndHelpDocumentsIt(t *testing.T) {
	m := createTestModel()
	if got := footerText(m, 200); !strings.Contains(got, "@:log") {
		t.Errorf("footer should advertise the command log: %q", got)
	}
	for page := 0; page < 5; page++ {
		m.helpPage = page
		if strings.Contains(stripANSI(m.renderHelpModalContent()), "Command log") {
			return
		}
	}
	t.Error("no help page documents the command log")
}
