package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"lazyproxyflare/internal/audit"
	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
)

// The snippet manager is one screen for every snippet task: list the snippets
// in the Caddyfile, view, edit and delete them, and start the wizard for new
// ones. ViewSnippetManager is the list; ViewSnippetDetail shows one snippet.
// The row after the last snippet is "+ New snippet" (the wizard).

// snippetListRows is how many snippet rows (two lines each) the list shows.
func (m Model) snippetListRows() int {
	_, h := modalDimensions(m.width, m.height, true)
	// borders (2), header (1), blank (1), blank before hint (1), hint (1), notice (1)
	return max((h-7)/2, 1)
}

// snippetTextWidth is the columns available to modal text.
func (m Model) snippetTextWidth() int {
	w, _ := modalDimensions(m.width, m.height, true)
	return max(w-4, 10)
}

// snippetDetailRows is how many body lines the detail view shows.
func (m Model) snippetDetailRows() int {
	_, h := modalDimensions(m.width, m.height, true)
	// borders (2), name (1), comment (1), used-by (1), blank (1), blank (1), hint (1), notice (1)
	return max(h-9, 3)
}

// snippetUsers lists what imports a snippet: the primary domain of each entry
// and the names of other snippets. Deleting a snippet that is still imported
// would break the Caddyfile.
func (m Model) snippetUsers(name string) (entries, snippets []string) {
	seen := map[string]bool{}
	for _, e := range m.entries {
		if e.Caddy == nil {
			continue
		}
		for _, imp := range e.Caddy.Imports {
			primary := caddy.GetPrimaryDomain(e.Caddy)
			if imp == name && !seen[primary] {
				seen[primary] = true
				entries = append(entries, primary)
			}
		}
	}
	for _, s := range m.snippets {
		if s.Name == name {
			continue
		}
		for _, line := range strings.Split(s.Content, "\n") {
			if f := strings.Fields(line); len(f) == 2 && f[0] == "import" && f[1] == name {
				snippets = append(snippets, s.Name)
				break
			}
		}
	}
	return entries, snippets
}

// snippetByName returns the snippet currently shown in the detail view.
func (m Model) snippetByName(name string) (caddy.Snippet, bool) {
	for _, s := range m.snippets {
		if s.Name == name {
			return s, true
		}
	}
	return caddy.Snippet{}, false
}

// handleOpenSnippetManager opens the manager from the list view.
func (m Model) handleOpenSnippetManager() (Model, tea.Cmd) {
	if m.currentView != ViewList || m.loading || m.searching || m.config == nil {
		return m, nil
	}
	m.snippetPanel = SnippetPanelState{}
	m.currentView = ViewSnippetManager
	return m, nil
}

// startSnippetWizard launches the creation wizard from the manager.
func (m Model) startSnippetWizard() (Model, tea.Cmd) {
	m.currentView = ViewList // handleOpenSnippetWizard only starts from the list
	m, cmd := m.handleOpenSnippetWizard()
	if m.currentView == ViewSnippetWizard {
		m.snippetPanel.FromManager = true
	} else {
		m.currentView = ViewSnippetManager
	}
	return m, cmd
}

// returnFromSnippetWizard brings the user back to the manager when the wizard
// was started from it, showing the wizard's result as the notice.
func (m Model) returnFromSnippetWizard() Model {
	if !m.snippetPanel.FromManager {
		return m
	}
	m.snippetPanel.FromManager = false
	m.currentView = ViewSnippetManager
	if m.snippetPanel.Cursor > len(m.snippets) {
		m.snippetPanel.Cursor = len(m.snippets)
	}
	if m.err != nil {
		m.snippetPanel.Notice = m.err.Error()
		m.err = nil
	}
	return m
}

// handleSnippetKey handles every key press in the manager and detail views.
func (m Model) handleSnippetKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		m.quitting = true
		return m, tea.Quit
	}
	if m.snippetPanel.Busy {
		return m, nil
	}
	if m.currentView == ViewSnippetDetail {
		return m.handleSnippetDetailKey(msg)
	}
	return m.handleSnippetListKey(key)
}

func (m Model) handleSnippetListKey(key string) (Model, tea.Cmd) {
	p := &m.snippetPanel
	total := len(m.snippets) + 1 // the "+ New snippet" row
	rows := m.snippetListRows()
	p.Notice = ""

	switch key {
	case "esc", "q", "w", "ctrl+w":
		m.currentView = ViewList
	case "down", "j":
		p.Cursor = min(p.Cursor+1, total-1)
	case "up", "k":
		p.Cursor = max(p.Cursor-1, 0)
	case "pgdown":
		p.Cursor = min(p.Cursor+rows, total-1)
	case "pgup":
		p.Cursor = max(p.Cursor-rows, 0)
	case "g", "home":
		p.Cursor = 0
	case "G", "end":
		p.Cursor = total - 1
	case "n":
		return m.startSnippetWizard()
	case "enter":
		if p.Cursor >= len(m.snippets) {
			return m.startSnippetWizard()
		}
		return m.openSnippetDetail(m.snippets[p.Cursor].Name, false), nil
	case "e":
		if p.Cursor < len(m.snippets) {
			return m.openSnippetDetail(m.snippets[p.Cursor].Name, true), nil
		}
	case "d":
		if p.Cursor < len(m.snippets) {
			m = m.openSnippetDetail(m.snippets[p.Cursor].Name, false)
			return m.askDeleteSnippet(), nil
		}
	}
	p.ScrollOffset = scrollIntoView(p.Cursor, p.ScrollOffset, rows)
	return m, nil
}

// openSnippetDetail shows one snippet, optionally straight in edit mode.
func (m Model) openSnippetDetail(name string, edit bool) Model {
	s, ok := m.snippetByName(name)
	if !ok {
		return m
	}
	m.snippetPanel.DetailName = name
	m.snippetPanel.DetailScroll = 0
	m.snippetPanel.ConfirmDelete = false
	m.snippetPanel.Notice = ""
	m.currentView = ViewSnippetDetail
	if edit {
		m.beginSnippetEdit(s)
	}
	return m
}

func (m *Model) beginSnippetEdit(s caddy.Snippet) {
	ta := textarea.New()
	ta.SetWidth(m.snippetTextWidth() - 2)
	ta.SetHeight(m.snippetDetailRows())
	ta.SetValue(s.Content)
	ta.Focus()
	m.snippetPanel.EditTextarea = ta
	m.snippetPanel.EditBase = s.Content
	m.snippetPanel.EditInitial = ta.Value()
	m.snippetPanel.Editing = true
}

// askDeleteSnippet starts the delete confirmation, or explains why the snippet
// cannot be deleted while something still imports it.
func (m Model) askDeleteSnippet() Model {
	entries, snippets := m.snippetUsers(m.snippetPanel.DetailName)
	if n := len(entries) + len(snippets); n > 0 {
		m.snippetPanel.Notice = fmt.Sprintf("Cannot delete: still imported by %d. Remove those imports first.", n)
		return m
	}
	m.snippetPanel.ConfirmDelete = true
	return m
}

func (m Model) handleSnippetDetailKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	key := msg.String()
	p := &m.snippetPanel
	name := p.DetailName

	if p.Editing {
		switch key {
		case "ctrl+s":
			body := p.EditTextarea.Value()
			if body == p.EditInitial {
				p.Notice = "No changes to save."
				return m, nil
			}
			p.Busy = true
			p.Notice = ""
			return m, snippetEditCmd(m.config, name, p.EditBase, retab(body))
		case "esc":
			p.Editing = false
			p.Notice = "Edit discarded."
			return m, nil
		}
		var cmd tea.Cmd
		p.EditTextarea, cmd = p.EditTextarea.Update(msg)
		return m, cmd
	}

	if p.ConfirmDelete {
		switch key {
		case "y":
			p.ConfirmDelete = false
			p.Busy = true
			return m, snippetDeleteCmd(m.config, name)
		default: // any other key cancels
			p.ConfirmDelete = false
			p.Notice = "Delete cancelled."
		}
		return m, nil
	}

	s, ok := m.snippetByName(name)
	if !ok {
		m.currentView = ViewSnippetManager
		return m, nil
	}
	p.Notice = ""
	maxScroll := max(len(strings.Split(s.Content, "\n"))-m.snippetDetailRows(), 0)
	switch key {
	case "esc", "q", "backspace", "left", "h":
		m.currentView = ViewSnippetManager
	case "e":
		m.beginSnippetEdit(s)
	case "d":
		return m.askDeleteSnippet(), nil
	case "down", "j":
		p.DetailScroll = min(p.DetailScroll+1, maxScroll)
	case "up", "k":
		p.DetailScroll = max(p.DetailScroll-1, 0)
	case "pgdown":
		p.DetailScroll = min(p.DetailScroll+m.snippetDetailRows(), maxScroll)
	case "pgup":
		p.DetailScroll = max(p.DetailScroll-m.snippetDetailRows(), 0)
	case "g", "home":
		p.DetailScroll = 0
	case "G", "end":
		p.DetailScroll = maxScroll
	}
	return m, nil
}

// handleSnippetPaste forwards pasted text to the snippet editor.
func (m Model) handleSnippetPaste(msg tea.PasteMsg) (Model, tea.Cmd) {
	if m.currentView != ViewSnippetDetail || !m.snippetPanel.Editing || m.snippetPanel.Busy {
		return m, nil
	}
	var cmd tea.Cmd
	m.snippetPanel.EditTextarea, cmd = m.snippetPanel.EditTextarea.Update(msg)
	return m, cmd
}

// handleSnippetMouse scrolls the list and selects rows.
func (m Model) handleSnippetMouse(msg mouseEvent) (Model, tea.Cmd) {
	p := &m.snippetPanel
	if p.Busy || p.Editing {
		return m, nil
	}
	if m.currentView == ViewSnippetDetail {
		switch msg.Action {
		case mouseWheelUp:
			p.DetailScroll = max(p.DetailScroll-1, 0)
		case mouseWheelDown:
			if s, ok := m.snippetByName(p.DetailName); ok {
				maxScroll := max(len(strings.Split(s.Content, "\n"))-m.snippetDetailRows(), 0)
				p.DetailScroll = min(p.DetailScroll+1, maxScroll)
			}
		}
		return m, nil
	}

	total := len(m.snippets) + 1
	switch msg.Action {
	case mouseWheelUp:
		p.Cursor = max(p.Cursor-1, 0)
	case mouseWheelDown:
		p.Cursor = min(p.Cursor+1, total-1)
	case mouseClick:
		// The list starts two lines below the modal's top border (header, blank).
		_, mh := modalDimensions(m.width, m.height, true)
		top := (m.height-mh)/2 + 1 + 2
		if row := (msg.Y - top) / 2; msg.Y >= top && row < m.snippetListRows() {
			if idx := p.ScrollOffset + row; idx < total {
				p.Cursor = idx
			}
		}
	}
	p.ScrollOffset = scrollIntoView(p.Cursor, p.ScrollOffset, m.snippetListRows())
	return m, nil
}

// --- file changes ---------------------------------------------------------

// snippetOpMsg is the result of an edit or delete.
type snippetOpMsg struct {
	op         string // "edit" or "delete"
	name       string
	err        error
	backupPath string
	snippets   []caddy.Snippet
}

// applyCaddyfileChange rewrites the Caddyfile with transform, then validates it
// and restarts Caddy. Any failure puts the old file back. It returns the path
// of the backup it took first.
func applyCaddyfileChange(cfg *config.Config, transform func(src string) (string, error)) (string, error) {
	path := cfg.Caddy.CaddyfilePath
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat Caddyfile: %w", err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read Caddyfile: %w", err)
	}
	updated, err := transform(string(src))
	if err != nil {
		return "", err
	}
	backupPath, err := caddy.BackupCaddyfile(path)
	if err != nil {
		return "", fmt.Errorf("failed to back up Caddyfile: %w", err)
	}
	if err := os.WriteFile(path, []byte(updated), info.Mode().Perm()); err != nil {
		return backupPath, restoreBackupWithError(path, backupPath, err, "write")
	}
	if err := formatAndValidateCaddyfile(cfg); err != nil {
		return backupPath, restoreBackupWithError(path, backupPath, err, "validation")
	}
	if err := caddy.RestartCaddy(cfg.Caddy.ContainerName); err != nil {
		return backupPath, restoreBackupWithError(path, backupPath, err, "Caddy restart")
	}
	return backupPath, nil
}

func readSnippets(cfg *config.Config) []caddy.Snippet {
	data, err := os.ReadFile(cfg.Caddy.CaddyfilePath)
	if err != nil {
		return nil
	}
	return caddy.ParseCaddyfileWithSnippets(string(data)).Snippets
}

func snippetEditCmd(cfg *config.Config, name, oldBody, newBody string) tea.Cmd {
	return func() tea.Msg {
		backup, err := applyCaddyfileChange(cfg, func(src string) (string, error) {
			return caddy.ReplaceSnippetBody(src, name, oldBody, newBody)
		})
		return snippetOpMsg{op: "edit", name: name, err: err, backupPath: backup, snippets: readSnippets(cfg)}
	}
}

func snippetDeleteCmd(cfg *config.Config, name string) tea.Cmd {
	return func() tea.Msg {
		backup, err := applyCaddyfileChange(cfg, func(src string) (string, error) {
			return caddy.RemoveSnippetBlock(src, name)
		})
		return snippetOpMsg{op: "delete", name: name, err: err, backupPath: backup, snippets: readSnippets(cfg)}
	}
}

// handleSnippetOp applies the result of an edit or delete.
func (m Model) handleSnippetOp(msg snippetOpMsg) Model {
	p := &m.snippetPanel
	p.Busy = false
	if msg.snippets != nil {
		m.snippets = msg.snippets
	}
	if m.audit.Logger != nil {
		entry := audit.LogEntry{
			Timestamp:  time.Now(),
			Operation:  audit.OperationUpdate,
			EntityType: audit.EntityCaddy,
			Domain:     msg.name,
			Details:    map[string]interface{}{"snippet": msg.name, "method": msg.op},
			Result:     audit.ResultSuccess,
		}
		if msg.op == "delete" {
			entry.Operation = audit.OperationDelete
		}
		if msg.err != nil {
			entry.Result = audit.ResultFailure
			entry.Error = msg.err.Error()
		}
		_ = m.audit.Logger.Log(entry)
	}
	if msg.err != nil {
		p.Notice = msg.err.Error() // stay where we are: an edit keeps the user's text
		return m
	}
	backup := ""
	if msg.backupPath != "" {
		backup = " (backup: " + filepath.Base(msg.backupPath) + ")"
	}
	switch msg.op {
	case "edit":
		p.Editing = false
		p.Notice = "✓ Saved " + msg.name + backup
	case "delete":
		m.currentView = ViewSnippetManager
		p.Cursor = min(p.Cursor, len(m.snippets))
		p.ScrollOffset = scrollIntoView(p.Cursor, 0, m.snippetListRows())
		p.Notice = "✓ Deleted " + msg.name + backup
	}
	return m
}

// retab turns leading groups of four spaces back into tabs. The textarea
// expands tabs on load, and Caddy indents with tabs.
func retab(body string) string {
	lines := strings.Split(body, "\n")
	for i, l := range lines {
		n := 0
		for strings.HasPrefix(l[n*4:], "    ") {
			n++
		}
		lines[i] = strings.Repeat("\t", n) + l[n*4:]
	}
	return strings.Join(lines, "\n")
}
