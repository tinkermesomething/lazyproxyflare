package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// footerHint is one key hint in the status bar. Lower prio is more important:
// when the terminal is too narrow, the highest prio numbers are dropped first.
type footerHint struct {
	key, desc string
	space     bool // render "key desc" instead of "key:desc"
	prio      int
}

func (h footerHint) plain() string {
	if h.space {
		return h.key + " " + h.desc
	}
	return h.key + ":" + h.desc
}

func (h footerHint) render() string {
	if h.space {
		return StyleKeybinding.Render(h.key) + " " + h.desc
	}
	return formatKeybinding(h.key, h.desc)
}

// footerHints returns the hints for the current list context, in display order.
func (m Model) footerHints() []footerHint {
	help := footerHint{key: "?", desc: "help", prio: 1}
	quit := footerHint{key: "q", desc: "quit", prio: 1}
	nav := footerHint{key: "↑↓", desc: "nav", space: true, prio: 1}

	if len(m.selectedEntries) > 0 {
		return []footerHint{
			nav,
			{key: "space", desc: "select", prio: 2},
			{key: "X", desc: "delete", prio: 2},
			{key: "S", desc: "sync", prio: 2},
			{key: "esc", desc: "clear", prio: 2},
			help, quit,
		}
	}

	// Entry actions only apply to the entry list. The details and snippets
	// panels only move and switch focus.
	if m.panelFocus != PanelFocusLeft {
		return []footerHint{
			nav,
			{key: "shift+tab", desc: "panel", prio: 2},
			{key: "w", desc: "snippets", prio: 3},
			help, quit,
		}
	}

	hints := []footerHint{
		nav,
		{key: "tab", desc: "view", space: true, prio: 2},
		{key: "a", desc: "add", prio: 2},
		{key: "⏎", desc: "edit", prio: 2},
		{key: "d", desc: "del", prio: 2},
	}
	if m.searchQuery != "" {
		hints = append(hints, footerHint{key: "esc", desc: "clear filter", prio: 2})
	}
	if m.activeTab == TabCaddy {
		hints = append(hints, footerHint{key: "E", desc: "editor", prio: 5})
	}
	return append(hints,
		footerHint{key: "w", desc: "snippets", prio: 4},
		footerHint{key: "p", desc: "profile", prio: 5},
		footerHint{key: "/", desc: "search", prio: 3},
		footerHint{key: "r", desc: "refresh", prio: 4},
		footerHint{key: "b", desc: "backups", prio: 4},
		footerHint{key: "@", desc: "log", prio: 5},
		help, quit,
	)
}

// footerTag names where the hints apply.
func (m Model) footerTag() string {
	switch {
	case m.panelFocus == PanelFocusSnippets:
		return "[Snippets]"
	case m.panelFocus == PanelFocusDetails:
		return "[Details]"
	case m.activeTab == TabCaddy:
		return "[Caddy]"
	default:
		return "[DNS]"
	}
}

// fitFooter joins the tag and hints into one line of at most avail columns,
// dropping the least important hints first. Display order is kept.
func fitFooter(tag string, hints []footerHint, avail int) string {
	keep := make([]bool, len(hints))
	for i := range keep {
		keep[i] = true
	}
	width := func() int {
		w := lipgloss.Width(tag)
		for i, h := range hints {
			if keep[i] {
				w += 1 + lipgloss.Width(h.plain())
			}
		}
		return w
	}
	for width() > avail {
		drop := -1
		for i, h := range hints {
			if keep[i] && (drop < 0 || h.prio >= hints[drop].prio) {
				drop = i
			}
		}
		if drop < 0 || hints[drop].prio <= 1 {
			break // only core hints left: never drop those
		}
		keep[drop] = false
	}

	parts := []string{StyleDim.Render(tag)}
	for i, h := range hints {
		if keep[i] {
			parts = append(parts, h.render())
		}
	}
	return strings.Join(parts, " ")
}

// getStatusBarContentFor returns the status bar content for a terminal width.
func (m Model) getStatusBarContentFor(width int) string {
	if m.loading {
		return StyleInfo.Render("⟳ Loading...")
	}

	// Errors are shown in a modal, not in the status bar
	if m.err != nil {
		return StyleDim.Render("Error occurred - see modal above")
	}

	if m.searching {
		return fmt.Sprintf("Search: %s_ (%s to accept, %s to cancel)",
			m.searchQuery,
			StyleKeybinding.Render("enter"),
			StyleKeybinding.Render("esc"))
	}

	// RenderStatusBar pads one column on each side
	return fitFooter(m.footerTag(), m.footerHints(), width-2)
}
