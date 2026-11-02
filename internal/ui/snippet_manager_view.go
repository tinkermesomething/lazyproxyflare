package ui

import (
	"fmt"
	"path/filepath"
	"strings"
)

// snippetDescription is what the list shows under a snippet's name: the user's
// own comment above it when there is one. The category guess is a fallback.
func snippetDescription(comment, guess string) string {
	if comment != "" {
		return comment
	}
	return guess
}

func usageLabel(entries, snippets []string) string {
	switch n := len(entries) + len(snippets); n {
	case 0:
		return StyleWarning.Render("unused")
	case 1:
		return StyleDim.Render("used by 1")
	default:
		return StyleDim.Render(fmt.Sprintf("used by %d", n))
	}
}

func (m Model) renderSnippetNotice() string {
	n := m.snippetPanel.Notice
	switch {
	case m.snippetPanel.Busy:
		return StyleInfo.Render("Applying: backup, validate, restart Caddy...")
	case n == "":
		return ""
	case strings.HasPrefix(n, "✓"):
		return StyleSuccess.Render(n)
	case n == "Delete cancelled." || n == "Edit discarded." || n == "No changes to save.":
		return StyleDim.Render(n)
	}
	return StyleError.Render(n)
}

// renderSnippetManagerContent renders the list of snippets.
func (m Model) renderSnippetManagerContent() string {
	var b strings.Builder
	p := m.snippetPanel
	width := m.snippetTextWidth()
	rows := m.snippetListRows()
	total := len(m.snippets) + 1

	b.WriteString(StyleDim.Render(fmt.Sprintf("%d snippets in %s", len(m.snippets), filepath.Base(m.config.Caddy.CaddyfilePath))))
	b.WriteString("\n\n")

	end := min(p.ScrollOffset+rows, total)
	for i := p.ScrollOffset; i < end; i++ {
		selected := i == p.Cursor
		marker := "  "
		if selected {
			marker = StyleHighlight.Render("→ ")
		}
		if i == len(m.snippets) {
			b.WriteString(marker + StyleKeybinding.Render("+ New snippet") + "\n")
			b.WriteString("    " + StyleDim.Render("Template, paste your own, or guided wizard") + "\n")
			continue
		}
		s := m.snippets[i]
		entries, snippets := m.snippetUsers(s.Name)
		name := s.Name
		if selected {
			name = StyleHighlight.Render(name)
		} else {
			name = StyleInfo.Render(name)
		}
		b.WriteString(marker + name + "  " + usageLabel(entries, snippets) + "\n")
		b.WriteString("    " + StyleDim.Render(fitLine(snippetDescription(s.Comment, s.Description), width-6)) + "\n")
	}
	if total > rows {
		b.WriteString(StyleDim.Render(fmt.Sprintf("[%d-%d of %d]", p.ScrollOffset+1, end, total)))
	}
	b.WriteString("\n")
	b.WriteString(m.renderSnippetNotice() + "\n")
	b.WriteString(StyleDim.Render("↑/↓ move  enter view  e edit  d delete  n new  esc close"))
	return b.String()
}

// renderSnippetDetailContent renders one snippet, its users and its body, or
// the editor.
func (m Model) renderSnippetDetailContent() string {
	p := m.snippetPanel
	s, ok := m.snippetByName(p.DetailName)
	if !ok {
		return StyleDim.Render("This snippet is no longer in the Caddyfile.") + "\n\n" + StyleDim.Render("esc: back")
	}
	width := m.snippetTextWidth()
	rows := m.snippetDetailRows()
	entries, snippets := m.snippetUsers(s.Name)

	var b strings.Builder
	b.WriteString(StyleInfo.Render(s.Name) + StyleDim.Render(fmt.Sprintf("  lines %d-%d", s.LineStart, s.LineEnd)) + "\n")
	b.WriteString(StyleDim.Render(fitLine(snippetDescription(s.Comment, s.Description), width)) + "\n")
	b.WriteString(StyleDim.Render(fitLine(usedByText(entries, snippets), width)) + "\n\n")

	hint := "e edit  d delete  ↑/↓ scroll  esc back"
	switch {
	case p.Editing:
		b.WriteString(p.EditTextarea.View())
		hint = "ctrl+s save (backup, validate, restart Caddy)  esc discard"
	case p.ConfirmDelete:
		b.WriteString(m.renderSnippetBody(s.Content, rows, width))
		hint = StyleWarning.Render("Delete "+s.Name+"?") + StyleDim.Render("  y confirm  any other key cancels")
		b.WriteString("\n" + m.renderSnippetNotice() + "\n" + hint)
		return b.String()
	default:
		b.WriteString(m.renderSnippetBody(s.Content, rows, width))
	}
	b.WriteString("\n" + m.renderSnippetNotice() + "\n" + StyleDim.Render(hint))
	return b.String()
}

// renderSnippetBody shows rows lines of the body starting at the scroll offset,
// padded so the hint line stays at the bottom.
func (m Model) renderSnippetBody(content string, rows, width int) string {
	lines := strings.Split(content, "\n")
	start := min(m.snippetPanel.DetailScroll, max(len(lines)-1, 0))
	var out []string
	for i := start; i < len(lines) && len(out) < rows; i++ {
		out = append(out, StyleDim.Render("│ ")+fitLine(lines[i], width-2))
	}
	for len(out) < rows {
		out = append(out, "")
	}
	return strings.Join(out, "\n")
}

func usedByText(entries, snippets []string) string {
	if len(entries)+len(snippets) == 0 {
		return "Not imported anywhere"
	}
	const show = 6
	var parts []string
	if len(entries) > 0 {
		list := entries
		more := ""
		if len(list) > show {
			more = fmt.Sprintf(" +%d more", len(list)-show)
			list = list[:show]
		}
		parts = append(parts, fmt.Sprintf("%d entries: %s%s", len(entries), strings.Join(list, ", "), more))
	}
	if len(snippets) > 0 {
		parts = append(parts, "snippets: "+strings.Join(snippets, ", "))
	}
	return "Used by " + strings.Join(parts, "; ")
}
