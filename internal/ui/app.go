package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"lazyproxyflare/internal/audit"
	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
	snippet_wizard "lazyproxyflare/internal/ui/snippet_wizard"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00D7FF")).
			MarginBottom(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true).
			Reverse(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	orphanedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))

	syncedIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	orphanedIconStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00"))
)

// Custom messages for async operations
type refreshStartMsg struct{}

type refreshCompleteMsg struct {
	entries  []diff.SyncedEntry
	snippets []caddy.Snippet
	err      error
}

type createEntryMsg struct {
	success     bool
	err         error
	dnsRecordID string // Track created DNS record for rollback
	backupPath  string // Track backup path
	errorStep   string // Which step failed
}

type deleteEntryMsg struct {
	success    bool
	err        error
	backupPath string // Track backup path
	errorStep  string // Which step failed
	domain     string // Domain that was deleted (for audit log)
	entityType string // "dns", "caddy", or "both"
}

type updateEntryMsg struct {
	success    bool
	err        error
	backupPath string // Track backup path
	errorStep  string // Which step failed
}

type syncEntryMsg struct {
	success     bool
	err         error
	backupPath  string // Track backup path (if syncing to Caddy)
	dnsRecordID string // Track DNS record ID (if syncing to DNS)
	errorStep   string // Which step failed
	domain      string // Domain that was synced (for audit log)
	syncType    string // "to_dns" or "to_caddy"
}





// NewModel creates a new Bubbletea model
func NewModel(entries []diff.SyncedEntry, snippets []caddy.Snippet, cfg *config.Config) Model {
	// Sort entries alphabetically by domain
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Domain < entries[j].Domain
	})

	// Initialize audit logger
	// Use ~/.config/lazyproxyflare as default config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")

	auditLogger, err := audit.NewLogger(configDir)
	if err != nil {
		// If we can't create logger, log error but continue (non-fatal)
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
	}

	return Model{
		entries:             entries,
		snippets:            snippets,
		config:              cfg,
		cursor:              0,
		currentView:         ViewList,
		selectedEntries:     make(map[string]bool),
		backup:              BackupState{RetentionDays: 30}, // Default: keep backups for 30 days
		panelFocus:          PanelFocusLeft,
		audit:               AuditState{Logger: auditLogger},
	}
}

// NewModelWithWizard creates a new model starting in wizard view (no profile, no data)
func NewModelWithWizard() Model {
	// Initialize audit logger
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")

	auditLogger, err := audit.NewLogger(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
	}

	// Initialize text input for wizard
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return Model{
		entries:             []diff.SyncedEntry{},
		snippets:            []caddy.Snippet{},
		config:              nil,
		cursor:              0,
		currentView:         ViewWizard,
		wizardStep:          WizardStepWelcome,
		wizardData:          WizardData{},
		wizardCursor:        0,
		wizardTextInput:     ti,
		profile:             ProfileState{Available: []string{}},
		selectedEntries:     make(map[string]bool),
		backup:              BackupState{RetentionDays: 30},
		panelFocus:          PanelFocusLeft,
		audit:               AuditState{Logger: auditLogger},
	}
}

// NewModelWithProfile creates a new model with a loaded profile and data
func NewModelWithProfile(entries []diff.SyncedEntry, snippets []caddy.Snippet, cfg *config.Config, profileName string, password string) Model {
	// Sort entries alphabetically by domain
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Domain < entries[j].Domain
	})

	// Initialize audit logger
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")

	auditLogger, err := audit.NewLogger(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
	}

	// Initialize text input for wizard
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return Model{
		entries:             entries,
		snippets:            snippets,
		config:              cfg,
		profile:             ProfileState{CurrentName: profileName},
		cursor:              0,
		currentView:         ViewList,
		selectedEntries:     make(map[string]bool),
		backup:              BackupState{RetentionDays: 30},
		panelFocus:          PanelFocusLeft,
		audit:               AuditState{Logger: auditLogger},
		wizardTextInput:     ti,
	}
}

// NewModelWithProfileSelector creates a new model starting in profile selector view
func NewModelWithProfileSelector(profiles []string, lastUsedProfile string) Model {
	// Initialize audit logger
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")

	auditLogger, err := audit.NewLogger(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize audit logger: %v\n", err)
	}

	// Set cursor to last used profile if found
	cursor := 0
	for i, profile := range profiles {
		if profile == lastUsedProfile {
			cursor = i
			break
		}
	}

	// Initialize text input for wizard
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	return Model{
		entries:             []diff.SyncedEntry{},
		snippets:            []caddy.Snippet{},
		config:              nil,
		profile:             ProfileState{CurrentName: lastUsedProfile, Available: profiles},
		cursor:              cursor,
		currentView:         ViewProfileSelector,
		selectedEntries:     make(map[string]bool),
		backup:              BackupState{RetentionDays: 30},
		panelFocus:          PanelFocusLeft,
		audit:               AuditState{Logger: auditLogger},
		wizardTextInput:     ti,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.EnableMouseAllMotion
}

// isValidIPAddress validates IPv4 address format
func isValidIPAddress(ip string) bool {
	if ip == "" {
		return false
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}

		// Check for leading zeros (except for "0" itself)
		if len(part) > 1 && part[0] == '0' {
			return false
		}

		// Parse as number
		num := 0
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return false
			}
			num = num*10 + int(ch-'0')
		}

		// Check range 0-255
		if num > 255 {
			return false
		}
	}

	return true
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.MouseMsg:
		return m.handleMouseMsg(msg)

	default:
		if m2, cmd, handled := m.handleAsyncMsg(msg); handled {
			return m2, cmd
		}
	}
	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Handle full-screen views that don't need the base layout
	switch m.currentView {
	case ViewWizard:
		return m.renderWizardView()
	case ViewProfileSelector:
		return m.renderProfileSelectorView()
	case ViewProfileEdit:
		return m.renderProfileEditView()
	case ViewAuditLog:
		return m.renderAuditLogView()
	}

	// Always render the main panel layout as base (for views that need it)
	base := m.renderPanelLayout()

	// Overlay modals on top of the main view
	switch m.currentView {
	case ViewHelp:
		return RenderModalOverlay(base, "Keyboard Shortcuts", m.renderHelpModalContent(), m.width, m.height)

	case ViewAdd:
		return RenderModalOverlay(base, "Add New Entry", m.renderAddFormContent(), m.width, m.height)

	case ViewEdit:
		return RenderModalOverlay(base, "Edit Entry", m.renderEditFormContent(), m.width, m.height)

	case ViewPreview:
		return RenderModalOverlay(base, "Preview Changes", m.renderPreviewContent(), m.width, m.height)

	case ViewConfirmDelete:
		return RenderModalOverlay(base, "Confirm Delete", m.renderConfirmDeleteContent(), m.width, m.height)

	case ViewConfirmSync:
		return RenderModalOverlay(base, "Confirm Sync", m.renderConfirmSyncContent(), m.width, m.height)

	case ViewBulkDeleteMenu:
		return RenderModalOverlay(base, "Bulk Delete", m.renderBulkDeleteMenuContent(), m.width, m.height)

	case ViewConfirmBulkDelete:
		return RenderModalOverlay(base, "Confirm Bulk Delete", m.renderConfirmBulkDeleteContent(), m.width, m.height)

	case ViewConfirmBatchDelete:
		return RenderModalOverlay(base, "Confirm Batch Delete", m.renderConfirmBatchDeleteContent(), m.width, m.height)

	case ViewConfirmBatchSync:
		return RenderModalOverlay(base, "Confirm Batch Sync", m.renderConfirmBatchSyncContent(), m.width, m.height)

	case ViewBackupManager:
		return RenderModalOverlay(base, "Backup Manager", m.renderBackupManagerContent(), m.width, m.height)

	case ViewBackupPreview:
		return RenderModalOverlay(base, "Backup Preview", m.renderBackupPreviewContent(), m.width, m.height)

	case ViewDeleteScope:
		return RenderModalOverlay(base, "Delete Scope", m.renderDeleteScopeContent(), m.width, m.height)

	case ViewRestoreScope:
		return RenderModalOverlay(base, "Restore Scope", m.renderRestoreScopeContent(), m.width, m.height)

	case ViewConfirmRestore:
		return RenderModalOverlay(base, "Confirm Restore", m.renderConfirmRestoreContent(), m.width, m.height)

	case ViewConfirmCleanup:
		return RenderModalOverlay(base, "Confirm Cleanup", m.renderConfirmCleanupContent(), m.width, m.height)

	case ViewSnippetDetail:
		return RenderModalOverlay(base, "Snippet Details", m.renderSnippetDetailView(), m.width, m.height)

	case ViewSnippetWizard:
		title := "Snippet Creation Wizard"
		var content string
		switch m.snippetWizardStep {
		case SnippetWizardWelcome:
			content = m.renderSnippetWizardWelcome()
		case SnippetWizardAutoDetect:
			content = m.renderSnippetWizardAutoDetect()
		case SnippetWizardIPRestriction:
			content = m.renderSnippetWizardIPRestriction()
		case SnippetWizardSecurityHeaders:
			content = m.renderSnippetWizardSecurityHeaders()
		case SnippetWizardPerformance:
			content = m.renderSnippetWizardPerformance()
		case SnippetWizardSummary:
			content = m.renderSnippetWizardSummary()
		case snippet_wizard.StepTemplateSelection:
			content = m.renderSnippetWizardTemplateSelection()
		case snippet_wizard.StepTemplateParams:
			content = m.renderSnippetWizardTemplateParams()
		case snippet_wizard.StepCustomSnippet:
			content = m.renderSnippetWizardCustom()
		default:
			content = "Wizard step not implemented"
		}
		return RenderModalOverlay(base, title, content, m.width, m.height)

	case ViewMigrationWizard:
		title := "Caddyfile Migration"
		content := m.renderMigrationWizard()
		return RenderModalOverlay(base, title, content, m.width, m.height)

	case ViewError:
		return RenderModalOverlay(base, "Error", m.renderErrorModalContent(), m.width, m.height)

	default:
		// Check if there's an error - show error modal overlay on top of current view
		if m.err != nil {
			return RenderModalOverlay(base, "Error", m.renderErrorModalContent(), m.width, m.height)
		}
		// Main view without modal
		return base
	}
}

// renderPanelLayout renders the main multi-panel layout

// getFilteredEntries returns entries filtered by search query and status filter

// renderListView renders the main list view
func (m Model) renderListView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("LazyProxyFlare - %s", m.config.Domain))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Get filtered entries
	displayEntries := m.getFilteredEntries()

	// Summary stats
	synced, orphanedDNS, orphanedCaddy := 0, 0, 0
	for _, entry := range displayEntries {
		switch entry.Status {
		case diff.StatusSynced:
			synced++
		case diff.StatusOrphanedDNS:
			orphanedDNS++
		case diff.StatusOrphanedCaddy:
			orphanedCaddy++
		}
	}

	summary := fmt.Sprintf("%s %d synced  %s %d orphaned (DNS)  %s %d orphaned (Caddy)",
		syncedIconStyle.Render("✓"), synced,
		orphanedIconStyle.Render("⚠"), orphanedDNS,
		orphanedIconStyle.Render("⚠"), orphanedCaddy,
	)
	b.WriteString(summary)
	b.WriteString("\n")

	// Show filter and sort info
	filterParts := []string{}
	if m.statusFilter != FilterAll {
		filterParts = append(filterParts, fmt.Sprintf("Status: %s", m.statusFilter.String()))
	}
	if m.dnsTypeFilter != DNSTypeAll {
		filterParts = append(filterParts, fmt.Sprintf("DNS Type: %s", m.dnsTypeFilter.String()))
	}
	if m.searchQuery != "" {
		filterParts = append(filterParts, fmt.Sprintf("Search: \"%s\"", m.searchQuery))
	}

	// Build info line
	infoLine := ""
	if len(filterParts) > 0 {
		infoLine = fmt.Sprintf("Filter: %s | ", strings.Join(filterParts, ", "))
	}
	infoLine += fmt.Sprintf("Sort: %s | Showing %d of %d entries",
		m.sortMode.String(), len(displayEntries), len(m.entries))
	if len(m.selectedEntries) > 0 {
		infoLine += fmt.Sprintf(" | Selected: %d", len(m.selectedEntries))
	}
	b.WriteString(dimStyle.Render(infoLine))
	b.WriteString("\n")

	// Calculate visible range
	visibleHeight := m.height - 10 // Leave space for title, summary, status bar
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	start := m.scrollOffset
	end := start + visibleHeight
	if end > len(displayEntries) {
		end = len(displayEntries)
	}

	// Entry list
	for i := start; i < end; i++ {
		entry := displayEntries[i]

		// Selection checkbox
		checkbox := "[ ]"
		if m.selectedEntries[entry.Domain] {
			checkbox = "[✓]"
		}

		// Icon based on status
		icon := entry.Status.Icon()
		if entry.Status == diff.StatusSynced {
			icon = syncedIconStyle.Render(icon)
		} else {
			icon = orphanedIconStyle.Render(icon)
		}

		// Domain name
		domain := entry.Domain

		// Details based on what exists
		var details string
		if entry.DNS != nil && entry.Caddy != nil {
			// Both exist - show DNS type and target, plus Caddy target
			details = fmt.Sprintf("DNS:[%s]%s → Caddy:%s:%d",
				entry.DNS.Type, entry.DNS.Content, entry.Caddy.Target, entry.Caddy.Port)
		} else if entry.DNS != nil {
			// Only DNS - show type and target
			details = fmt.Sprintf("DNS:[%s]%s (no Caddy)", entry.DNS.Type, entry.DNS.Content)
		} else if entry.Caddy != nil {
			// Only Caddy
			details = fmt.Sprintf("Caddy:%s:%d (no DNS)", entry.Caddy.Target, entry.Caddy.Port)
		}

		// Render line
		line := fmt.Sprintf("%s %s %-40s %s", checkbox, icon, domain, details)

		// Apply cursor style
		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else if entry.Status != diff.StatusSynced {
			line = orphanedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(displayEntries) > visibleHeight {
		scrollInfo := fmt.Sprintf("\n[Showing %d-%d of %d]", start+1, end, len(displayEntries))
		b.WriteString(normalStyle.Render(scrollInfo))
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	var statusBar string
	if m.loading {
		// Show loading indicator
		statusBar = statusBarStyle.Render("⟳ Refreshing data from Cloudflare and Caddyfile...")
	} else if m.searching {
		// Show search prompt
		statusBar = statusBarStyle.Render(
			fmt.Sprintf("Search: %s_  (enter:accept  esc:cancel)", m.searchQuery),
		)
	} else if m.err != nil {
		// Show error
		statusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Padding(0, 1).
			Render(fmt.Sprintf("Error: %v (press r to retry)", m.err))
	} else {
		// Normal status bar - show batch operations if selections exist
		if len(m.selectedEntries) > 0 {
			statusBar = statusBarStyle.Render(
				"j/k:navigate  space:select  X:batch-delete  S:batch-sync  f:filter  t:type  o:sort  /:search  b:backups  r:refresh  ?:help  q:quit",
			)
		} else {
			statusBar = statusBarStyle.Render(
				"j/k:navigate  space:select  enter:details  a:add  d:delete  s:sync  f:filter  t:type  o:sort  /:search  b:backups  r:refresh  ?:help  q:quit",
			)
		}
	}
	b.WriteString(statusBar)

	return b.String()
}

// renderDetailsView renders detailed view of selected entry
func (m Model) renderDetailsView() string {
	if m.cursor >= len(m.entries) {
		return "No entry selected"
	}

	entry := m.entries[m.cursor]
	var b strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("Details - %s", entry.Domain))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Status
	statusLine := fmt.Sprintf("Status: %s %s\n\n",
		entry.Status.Icon(),
		entry.Status.String(),
	)
	if entry.Status == diff.StatusSynced {
		b.WriteString(syncedIconStyle.Render(statusLine))
	} else {
		b.WriteString(orphanedIconStyle.Render(statusLine))
	}

	// DNS Information
	if entry.DNS != nil {
		b.WriteString(titleStyle.Render("DNS Record (Cloudflare)"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Type:     %s\n", entry.DNS.Type))
		b.WriteString(fmt.Sprintf("  Name:     %s\n", entry.DNS.Name))
		b.WriteString(fmt.Sprintf("  Content:  %s\n", entry.DNS.Content))
		b.WriteString(fmt.Sprintf("  Proxied:  %v", entry.DNS.Proxied))
		if entry.DNS.Proxied {
			b.WriteString(" (Orange cloud enabled)")
		}
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  TTL:      %d", entry.DNS.TTL))
		if entry.DNS.TTL == 1 {
			b.WriteString(" (Auto)")
		}
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Zone ID:  %s\n", entry.DNS.ZoneID))
		b.WriteString(fmt.Sprintf("  Record ID: %s\n", entry.DNS.ID))
		b.WriteString("\n")
	} else {
		b.WriteString(orphanedIconStyle.Render("DNS Record: Not found in Cloudflare"))
		b.WriteString("\n\n")
	}

	// Caddy Information
	if entry.Caddy != nil {
		b.WriteString(titleStyle.Render("Caddy Configuration"))
		b.WriteString("\n")

		// Primary info
		b.WriteString(fmt.Sprintf("  Domain:   %s\n", entry.Caddy.Domain))
		if len(entry.Caddy.Domains) > 1 {
			b.WriteString(fmt.Sprintf("  Aliases:  %v\n", entry.Caddy.Domains[1:]))
		}
		b.WriteString(fmt.Sprintf("  Target:   %s\n", entry.Caddy.Target))
		b.WriteString(fmt.Sprintf("  Port:     %d\n", entry.Caddy.Port))
		b.WriteString(fmt.Sprintf("  SSL:      %v", entry.Caddy.SSL))
		if entry.Caddy.SSL {
			b.WriteString(" (HTTPS)")
		} else {
			b.WriteString(" (HTTP)")
		}
		b.WriteString("\n")

		// Features
		if entry.Caddy.IPRestricted || entry.Caddy.OAuthHeaders || entry.Caddy.WebSocket {
			b.WriteString("  Features: ")
			features := []string{}
			if entry.Caddy.IPRestricted {
				features = append(features, "IP Restricted")
			}
			if entry.Caddy.OAuthHeaders {
				features = append(features, "OAuth Headers")
			}
			if entry.Caddy.WebSocket {
				features = append(features, "WebSocket")
			}
			b.WriteString(strings.Join(features, ", "))
			b.WriteString("\n")
		}

		// Imports
		if len(entry.Caddy.Imports) > 0 {
			b.WriteString(fmt.Sprintf("  Imports:  %v\n", entry.Caddy.Imports))
		}

		// Location in file
		b.WriteString(fmt.Sprintf("  Location: Lines %d-%d", entry.Caddy.LineStart, entry.Caddy.LineEnd))
		if entry.Caddy.HasMarker {
			b.WriteString(" (has marker)")
		}
		b.WriteString("\n\n")

		// Raw block
		b.WriteString(titleStyle.Render("Raw Caddyfile Block"))
		b.WriteString("\n")
		b.WriteString(normalStyle.Render(entry.Caddy.RawBlock))
		b.WriteString("\n")
	} else {
		b.WriteString(orphanedIconStyle.Render("Caddy Configuration: Not found in Caddyfile"))
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	statusBar := statusBarStyle.Render("j/k:next/prev  e:edit  d:delete  s:sync  esc:back  q:quit")
	b.WriteString(statusBar)

	return b.String()
}

// renderHelpView renders the help screen
// renderHelpModalContent returns help content for modal display (two-column layout)

// renderAddForm renders the add entry form
func (m Model) renderAddForm() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Add New Entry")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Field labels and values
	fields := []struct {
		label       string
		value       string
		placeholder string
		focused     bool
	}{
		{
			label:       "Subdomain",
			value:       m.addForm.Subdomain,
			placeholder: "(without domain - will become subdomain." + m.config.Domain + ")",
			focused:     m.addForm.FocusedField == 0,
		},
		{
			label:       "CNAME Target",
			value:       m.addForm.DNSTarget,
			placeholder: "(DNS record target)",
			focused:     m.addForm.FocusedField == 1,
		},
		{
			label:       "Reverse Proxy Target",
			value:       m.addForm.ReverseProxyTarget,
			placeholder: "(internal IP or hostname for Caddy)",
			focused:     m.addForm.FocusedField == 4,
		},
		{
			label:       "Service Port",
			value:       m.addForm.ServicePort,
			placeholder: "",
			focused:     m.addForm.FocusedField == 5,
		},
	}

	// Render text input fields
	for _, field := range fields {
		// Label
		b.WriteString(normalStyle.Render(field.label + ":"))
		b.WriteString("\n  ")

		// Input field
		inputStyle := normalStyle.Copy()
		if field.focused {
			inputStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		displayValue := field.value
		if field.focused {
			displayValue += "_" // Cursor
		}
		if displayValue == "" || displayValue == "_" {
			displayValue = field.placeholder
		}

		// Pad to consistent width
		inputWidth := 60
		if len(displayValue) > inputWidth {
			displayValue = displayValue[:inputWidth]
		} else {
			displayValue = displayValue + strings.Repeat(" ", inputWidth-len(displayValue))
		}

		b.WriteString(inputStyle.Render("[" + displayValue + "]"))
		b.WriteString("\n\n")
	}

	// Checkboxes
	checkboxes := []struct {
		label   string
		checked bool
		index   int
	}{
		{"Proxy through Cloudflare (orange cloud)", m.addForm.Proxied, 4},
		{"LAN only (404 for external traffic)", m.addForm.LANOnly, 5},
		{"Enable SSL/TLS (https://)", m.addForm.SSL, 6},
		{"Include OAuth/OIDC headers", m.addForm.OAuth, 7},
		{"WebSocket support", m.addForm.WebSocket, 8},
	}

	for _, cb := range checkboxes {
		cbStyle := normalStyle.Copy()
		if m.addForm.FocusedField == cb.index {
			cbStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		checkbox := "[ ]"
		if cb.checked {
			checkbox = "[✓]"
		}

		line := fmt.Sprintf("  %s %s", checkbox, cb.label)
		b.WriteString(cbStyle.Render(line))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	statusBar := statusBarStyle.Render("↑/↓ or j/k:navigate  space:toggle  enter:preview  esc:cancel")
	b.WriteString(statusBar)

	return b.String()
}

// renderEditForm renders the edit entry form
func (m Model) renderEditForm() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Edit Entry")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Field labels and values
	fields := []struct {
		label       string
		value       string
		placeholder string
		focused     bool
	}{
		{
			label:       "Subdomain",
			value:       m.addForm.Subdomain,
			placeholder: "(without domain - will become subdomain." + m.config.Domain + ")",
			focused:     m.addForm.FocusedField == 0,
		},
		{
			label:       "CNAME Target",
			value:       m.addForm.DNSTarget,
			placeholder: "(DNS record target)",
			focused:     m.addForm.FocusedField == 1,
		},
		{
			label:       "Reverse Proxy Target",
			value:       m.addForm.ReverseProxyTarget,
			placeholder: "(internal IP or hostname for Caddy)",
			focused:     m.addForm.FocusedField == 4,
		},
		{
			label:       "Service Port",
			value:       m.addForm.ServicePort,
			placeholder: "",
			focused:     m.addForm.FocusedField == 5,
		},
	}

	// Render text input fields
	for _, field := range fields {
		// Label
		b.WriteString(normalStyle.Render(field.label + ":"))
		b.WriteString("\n  ")

		// Input field
		inputStyle := normalStyle.Copy()
		if field.focused {
			inputStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		displayValue := field.value
		if field.focused {
			displayValue += "_" // Cursor
		}
		if displayValue == "" || displayValue == "_" {
			displayValue = field.placeholder
		}

		// Pad to consistent width
		inputWidth := 60
		if len(displayValue) > inputWidth {
			displayValue = displayValue[:inputWidth]
		} else {
			displayValue = displayValue + strings.Repeat(" ", inputWidth-len(displayValue))
		}

		b.WriteString(inputStyle.Render("[" + displayValue + "]"))
		b.WriteString("\n\n")
	}

	// Checkboxes
	checkboxes := []struct {
		label   string
		checked bool
		index   int
	}{
		{"Proxy through Cloudflare (orange cloud)", m.addForm.Proxied, 4},
		{"LAN only (404 for external traffic)", m.addForm.LANOnly, 5},
		{"Enable SSL/TLS (https://)", m.addForm.SSL, 6},
		{"Include OAuth/OIDC headers", m.addForm.OAuth, 7},
		{"WebSocket support", m.addForm.WebSocket, 8},
	}

	for _, cb := range checkboxes {
		cbStyle := normalStyle.Copy()
		if m.addForm.FocusedField == cb.index {
			cbStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		checkbox := "[ ]"
		if cb.checked {
			checkbox = "[✓]"
		}

		line := fmt.Sprintf("  %s %s", checkbox, cb.label)
		b.WriteString(cbStyle.Render(line))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	statusBar := statusBarStyle.Render("↑/↓ or j/k:navigate  space:toggle  enter:preview  esc:cancel")
	b.WriteString(statusBar)

	return b.String()
}

// renderPreviewScreen renders the confirmation/preview screen
func (m Model) renderPreviewScreen() string {
	var b strings.Builder

	// Title - different for create vs update
	var title string
	if m.editingEntry != nil {
		title = titleStyle.Render("Confirm Update")
	} else {
		title = titleStyle.Render("Confirm Create")
	}
	b.WriteString(title)
	b.WriteString("\n\n")

	// Build FQDN
	fqdn := m.addForm.Subdomain + "." + m.config.Domain

	// Parse port
	port := 80
	if m.addForm.ServicePort != "" {
		fmt.Sscanf(m.addForm.ServicePort, "%d", &port)
	}

	// Different message for create vs update
	if m.editingEntry != nil {
		b.WriteString(normalStyle.Render("Will update the entry to:"))
	} else {
		b.WriteString(normalStyle.Render("Will create the following:"))
	}
	b.WriteString("\n\n")

	// DNS Record Preview
	dnsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D7FF")).
		Padding(0, 1).
		Width(70)

	dnsContent := strings.Builder{}
	dnsContent.WriteString(titleStyle.Render("Cloudflare DNS Record"))
	dnsContent.WriteString("\n")
	dnsContent.WriteString(fmt.Sprintf("  Type:     CNAME\n"))
	dnsContent.WriteString(fmt.Sprintf("  Name:     %s\n", fqdn))
	dnsContent.WriteString(fmt.Sprintf("  Target:   %s\n", m.addForm.DNSTarget))
	if m.addForm.Proxied {
		dnsContent.WriteString(fmt.Sprintf("  Proxied:  Yes (Orange cloud enabled)\n"))
	} else {
		dnsContent.WriteString(fmt.Sprintf("  Proxied:  No (DNS-only)\n"))
	}
	dnsContent.WriteString(fmt.Sprintf("  TTL:      Auto\n"))

	b.WriteString(dnsBox.Render(dnsContent.String()))
	b.WriteString("\n\n")

	// Caddy Block Preview
	caddyBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D7FF")).
		Padding(0, 1).
		Width(70)

	// Generate the actual Caddy block
	caddyBlock := caddy.GenerateCaddyBlock(caddy.GenerateBlockInput{
		FQDN:              fqdn,
		Target:            m.addForm.ReverseProxyTarget,
		Port:              port,
		SSL:               m.addForm.SSL,
		LANOnly:           m.addForm.LANOnly,
		OAuth:             m.addForm.OAuth,
		WebSocket:         m.addForm.WebSocket,
		LANSubnet:         m.config.Defaults.LANSubnet,
		AllowedExtIP:      m.config.Defaults.AllowedExternalIP,
		SelectedSnippets:  getSelectedSnippetNames(m.addForm.SelectedSnippets),
		CustomCaddyConfig: m.addForm.CustomCaddyConfig,
	})

	caddyContent := strings.Builder{}
	caddyContent.WriteString(titleStyle.Render("Caddyfile Entry"))
	caddyContent.WriteString("\n")
	caddyContent.WriteString(caddyBlock)

	b.WriteString(caddyBox.Render(caddyContent.String()))
	b.WriteString("\n\n")

	// Status/Error display
	b.WriteString("\n")
	if m.loading {
		if m.editingEntry != nil {
			b.WriteString(statusBarStyle.Render("  ⟳ Updating entry..."))
		} else {
			b.WriteString(statusBarStyle.Render("  ⟳ Creating entry..."))
		}
	} else if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		b.WriteString(errorStyle.Render(fmt.Sprintf("  ✗ Error: %v", m.err)))
	} else {
		if m.editingEntry != nil {
			b.WriteString(syncedIconStyle.Render("  ✓ Ready to update"))
		} else {
			b.WriteString(syncedIconStyle.Render("  ✓ Ready to create"))
		}
	}
	b.WriteString("\n\n")

	// Status bar
	var statusBar string
	if m.loading {
		statusBar = statusBarStyle.Render("Please wait...")
	} else {
		if m.editingEntry != nil {
			statusBar = statusBarStyle.Render("y:confirm and update  esc:back to edit")
		} else {
			statusBar = statusBarStyle.Render("y:confirm and create  esc:back to edit")
		}
	}
	b.WriteString(statusBar)

	return b.String()
}

// createSnippetsFromWizard creates snippets based on wizard selections
func (m Model) createSnippetsFromWizard() (Model, tea.Cmd) {
	// Use wizard's centralized generation logic
	wizard := snippet_wizard.Wizard{
		State: &snippet_wizard.WizardState{
			Data: &m.snippetWizardData,
		},
	}
	generatedSnippets := wizard.GenerateSnippets()

	if len(generatedSnippets) == 0 {
		m.currentView = ViewList
		m.err = fmt.Errorf("no snippets selected")
		return m, nil
	}

	// Convert GeneratedSnippets to snippet strings
	var snippetsToAdd []string
	var snippetNames []string
	for _, gen := range generatedSnippets {
		// Wrap content in snippet syntax
		snippetStr := fmt.Sprintf("(%s) {\n%s\n}", gen.Name, gen.Content)
		snippetsToAdd = append(snippetsToAdd, snippetStr)
		snippetNames = append(snippetNames, gen.Name)
	}

	// Read current Caddyfile
	content, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.currentView = ViewList
		m.err = fmt.Errorf("failed to read Caddyfile: %w", err)
		return m, nil
	}

	// Check for duplicate snippets and filter them out
	parsed := caddy.ParseCaddyfileWithSnippets(string(content))
	existingNames := make(map[string]bool)
	for _, s := range parsed.Snippets {
		existingNames[s.Name] = true
	}

	// Filter out duplicates - only create new snippets
	var snippetsToCreate []string
	var newSnippetNames []string
	var skippedDuplicates []string

	for i, name := range snippetNames {
		if existingNames[name] {
			skippedDuplicates = append(skippedDuplicates, name)
		} else {
			snippetsToCreate = append(snippetsToCreate, snippetsToAdd[i])
			newSnippetNames = append(newSnippetNames, name)
		}
	}

	// If all snippets are duplicates, inform user
	if len(snippetsToCreate) == 0 {
		m.currentView = ViewList
		m.err = fmt.Errorf("all selected snippets already exist: %s", strings.Join(skippedDuplicates, ", "))
		return m, nil
	}

	// Prepend snippets to Caddyfile
	var newContent strings.Builder
	newContent.WriteString("# === Snippets created by LazyProxyFlare ===\n")
	newContent.WriteString("# Generated: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")

	for _, snippet := range snippetsToCreate {
		newContent.WriteString(snippet)
		newContent.WriteString("\n\n")
	}

	newContent.WriteString("# === End of snippets ===\n\n")
	newContent.WriteString(string(content))

	// Backup current Caddyfile
	backupPath, err := caddy.BackupCaddyfile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.currentView = ViewList
		m.err = fmt.Errorf("failed to backup Caddyfile: %w", err)
		return m, nil
	}

	// Get permissions from backup to preserve them
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		m.currentView = ViewList
		m.err = fmt.Errorf("failed to stat backup: %w", err)
		return m, nil
	}
	backupPerms := backupInfo.Mode().Perm()

	// Write new content with original permissions
	if err := os.WriteFile(m.config.Caddy.CaddyfilePath, []byte(newContent.String()), backupPerms); err != nil {
		m.currentView = ViewList
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: write failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("failed to write Caddyfile (backup restored): %w", err)
		}
		return m, nil
	}

	// Validate Caddyfile
	if err := caddy.FormatAndValidateCaddyfile(
		m.config.Caddy.CaddyfilePath,
		m.config.Caddy.CaddyfileContainerPath,
		m.config.Caddy.ContainerName,
		m.config.Caddy.DockerMethod,
		m.config.Caddy.ComposeFilePath,
		m.config.Caddy.ValidationCommand,
	); err != nil {
		m.currentView = ViewList
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: validation failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("Caddyfile validation failed (backup restored): %w", err)
		}
		return m, nil
	}

	// Reload Caddy
	if err := caddy.RestartCaddy(m.config.Caddy.ContainerName); err != nil {
		m.currentView = ViewList
		m.err = fmt.Errorf("failed to restart Caddy: %w", err)
		return m, nil
	}

	// Reload snippets
	caddyContent, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err == nil {
		parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))
		m.snippets = parsed.Snippets
	}

	// Success! Return to list view with message
	m.currentView = ViewList
	m.snippetWizardData = SnippetWizardData{} // Reset wizard data

	// Build success message
	if len(skippedDuplicates) > 0 {
		// Some duplicates were skipped
		m.err = fmt.Errorf("✓ Created %d snippet(s): %s | Skipped %d duplicate(s): %s",
			len(newSnippetNames), strings.Join(newSnippetNames, ", "),
			len(skippedDuplicates), strings.Join(skippedDuplicates, ", "))
	} else {
		// All created successfully
		m.err = nil
	}

	// Log the operation (only log newly created snippets)
	if m.audit.Logger != nil && len(newSnippetNames) > 0 {
		logEntry := audit.LogEntry{
			Timestamp:  time.Now(),
			Operation:  audit.OperationCreate,
			EntityType: audit.EntityCaddy,
			Domain:     strings.Join(newSnippetNames, ", "),
			Details: map[string]interface{}{
				"snippets": newSnippetNames,
				"method":   "wizard",
				"skipped":  skippedDuplicates,
			},
			Result: audit.ResultSuccess,
		}
		_ = m.audit.Logger.Log(logEntry)
	}

	return m, nil
}

// saveSnippetEdit saves the edited snippet content back to Caddyfile
func (m Model) saveSnippetEdit() (Model, tea.Cmd) {
	if m.snippetPanel.EditingIndex < 0 || m.snippetPanel.EditingIndex >= len(m.snippets) {
		m.err = fmt.Errorf("invalid snippet index")
		return m, nil
	}

	snippet := m.snippets[m.snippetPanel.EditingIndex]
	newContent := m.snippetPanel.EditTextarea.Value()

	// Read current Caddyfile
	content, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.err = fmt.Errorf("failed to read Caddyfile: %w", err)
		return m, nil
	}

	// Replace the snippet content in the Caddyfile
	// We need to find the snippet block and replace it
	lines := strings.Split(string(content), "\n")

	// Build new Caddyfile content with updated snippet
	var newCaddyfile strings.Builder
	for i := 0; i < len(lines); i++ {
		// Check if this is the start of our snippet (LineStart is 1-indexed)
		if i+1 == snippet.LineStart {
			// Write the snippet name line
			newCaddyfile.WriteString(lines[i])
			newCaddyfile.WriteString("\n")

			// Write the new content (already includes proper indentation from textarea)
			newCaddyfile.WriteString(newContent)
			newCaddyfile.WriteString("\n")

			// Skip the old content lines (until LineEnd-1, as LineEnd is the closing brace)
			// Then write the closing brace
			i = snippet.LineEnd - 1 // -1 because LineEnd is 1-indexed and we're 0-indexed
			if i < len(lines) {
				newCaddyfile.WriteString(lines[i])
				newCaddyfile.WriteString("\n")
			}
		} else {
			newCaddyfile.WriteString(lines[i])
			newCaddyfile.WriteString("\n")
		}
	}

	// Backup current Caddyfile
	backupPath, err := caddy.BackupCaddyfile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.err = fmt.Errorf("failed to backup Caddyfile: %w", err)
		return m, nil
	}

	// Get permissions from backup to preserve them
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		m.err = fmt.Errorf("failed to stat backup: %w", err)
		return m, nil
	}
	backupPerms := backupInfo.Mode().Perm()

	// Write new content with original permissions
	if err := os.WriteFile(m.config.Caddy.CaddyfilePath, []byte(newCaddyfile.String()), backupPerms); err != nil {
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: write failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("failed to write Caddyfile (backup restored): %w", err)
		}
		return m, nil
	}

	// Validate Caddyfile
	if err := caddy.FormatAndValidateCaddyfile(
		m.config.Caddy.CaddyfilePath,
		m.config.Caddy.CaddyfileContainerPath,
		m.config.Caddy.ContainerName,
		m.config.Caddy.DockerMethod,
		m.config.Caddy.ComposeFilePath,
		m.config.Caddy.ValidationCommand,
	); err != nil {
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: validation failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("Caddyfile validation failed (backup restored): %w", err)
		}
		return m, nil
	}

	// Reload Caddy
	if err := caddy.RestartCaddy(m.config.Caddy.ContainerName); err != nil {
		m.err = fmt.Errorf("failed to restart Caddy: %w", err)
		return m, nil
	}

	// Reload snippets
	caddyContent, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err == nil {
		parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))
		m.snippets = parsed.Snippets
	}

	// Exit edit mode, stay in detail view
	m.snippetPanel.Editing = false
	m.err = nil

	// Log the operation
	if m.audit.Logger != nil {
		logEntry := audit.LogEntry{
			Timestamp:  time.Now(),
			Operation:  audit.OperationUpdate,
			EntityType: audit.EntityCaddy,
			Domain:     snippet.Name,
			Details: map[string]interface{}{
				"snippet": snippet.Name,
				"method":  "edit",
			},
			Result: audit.ResultSuccess,
		}
		_ = m.audit.Logger.Log(logEntry)
	}

	return m, nil
}

// deleteSnippet removes the selected snippet from the Caddyfile
func (m Model) deleteSnippet() (Model, tea.Cmd) {
	if m.snippetPanel.EditingIndex < 0 || m.snippetPanel.EditingIndex >= len(m.snippets) {
		m.err = fmt.Errorf("invalid snippet index")
		return m, nil
	}

	snippet := m.snippets[m.snippetPanel.EditingIndex]

	// Check if snippet is currently in use
	usage := m.calculateSnippetUsage()
	if usage[snippet.Name] > 0 {
		m.err = fmt.Errorf("cannot delete snippet '%s': currently used by %d entries", snippet.Name, usage[snippet.Name])
		return m, nil
	}

	// Read current Caddyfile
	content, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.err = fmt.Errorf("failed to read Caddyfile: %w", err)
		return m, nil
	}

	// Remove the snippet from the Caddyfile
	lines := strings.Split(string(content), "\n")

	// Build new Caddyfile content without this snippet
	var newCaddyfile strings.Builder
	skipUntilLine := -1

	for i := 0; i < len(lines); i++ {
		lineNum := i + 1 // Convert to 1-indexed

		// If we're at the snippet start, mark to skip until end
		if lineNum == snippet.LineStart {
			skipUntilLine = snippet.LineEnd
			continue
		}

		// Skip lines that are part of the snippet
		if skipUntilLine >= 0 && lineNum <= skipUntilLine {
			continue
		}

		// Write non-snippet lines
		newCaddyfile.WriteString(lines[i])
		newCaddyfile.WriteString("\n")
	}

	// Backup current Caddyfile
	backupPath, err := caddy.BackupCaddyfile(m.config.Caddy.CaddyfilePath)
	if err != nil {
		m.err = fmt.Errorf("failed to backup Caddyfile: %w", err)
		return m, nil
	}

	// Get permissions from backup to preserve them
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		m.err = fmt.Errorf("failed to stat backup: %w", err)
		return m, nil
	}
	backupPerms := backupInfo.Mode().Perm()

	// Write new content with original permissions
	if err := os.WriteFile(m.config.Caddy.CaddyfilePath, []byte(newCaddyfile.String()), backupPerms); err != nil {
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: write failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("failed to write Caddyfile (backup restored): %w", err)
		}
		return m, nil
	}

	// Validate Caddyfile
	if err := caddy.FormatAndValidateCaddyfile(
		m.config.Caddy.CaddyfilePath,
		m.config.Caddy.CaddyfileContainerPath,
		m.config.Caddy.ContainerName,
		m.config.Caddy.DockerMethod,
		m.config.Caddy.ComposeFilePath,
		m.config.Caddy.ValidationCommand,
	); err != nil {
		// Attempt to restore backup
		if restoreErr := caddy.RestoreFromBackup(m.config.Caddy.CaddyfilePath, backupPath); restoreErr != nil {
			m.err = fmt.Errorf("CRITICAL: validation failed AND backup restore failed: %w (original error: %v)", restoreErr, err)
		} else {
			m.err = fmt.Errorf("Caddyfile validation failed (backup restored): %w", err)
		}
		return m, nil
	}

	// Reload Caddy
	if err := caddy.RestartCaddy(m.config.Caddy.ContainerName); err != nil {
		m.err = fmt.Errorf("failed to restart Caddy: %w", err)
		return m, nil
	}

	// Reload snippets
	caddyContent, err := os.ReadFile(m.config.Caddy.CaddyfilePath)
	if err == nil {
		parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))
		m.snippets = parsed.Snippets
	}

	// Return to list view after deletion
	m.currentView = ViewList
	m.snippetPanel.Editing = false
	m.err = nil

	// Log the operation
	if m.audit.Logger != nil {
		logEntry := audit.LogEntry{
			Timestamp:  time.Now(),
			Operation:  audit.OperationDelete,
			EntityType: audit.EntityCaddy,
			Domain:     snippet.Name,
			Details: map[string]interface{}{
				"snippet": snippet.Name,
				"method":  "delete",
			},
			Result: audit.ResultSuccess,
		}
		_ = m.audit.Logger.Log(logEntry)
	}

	return m, nil
}

// renderConfirmDeleteView renders the delete confirmation screen

// renderBackupManagerView renders the backup manager screen

// Modal content renderers (stub implementations to be filled in)


// getTemplateParamsCursorMax returns the max cursor value for the current template being configured
func (m Model) getTemplateParamsCursorMax() int {
	// Find the first selected template to configure (matches render logic)
	templateKeys := []string{
		"cors_headers", "rate_limiting", "large_uploads", "extended_timeouts",
		"static_caching", "compression_advanced", "https_backend", "auth_headers",
		"websocket_advanced", "custom_headers_inject", "rewrite_rules", "frame_embedding",
		"ip_restricted", "security_headers", "performance",
	}

	var currentTemplate string
	for _, key := range templateKeys {
		if m.snippetWizardData.SelectedTemplates[key] {
			currentTemplate = key
			break
		}
	}

	if currentTemplate == "" {
		return 0
	}

	// Return max cursor for each template type
	switch currentTemplate {
	case "cors_headers":
		return 2 // 3 fields: origins, methods, credentials
	case "rate_limiting":
		return 1 // 2 fields: req/s, burst
	case "large_uploads":
		return 0 // 1 field: max size
	case "extended_timeouts":
		return 2 // 3 fields: read, write, dial
	case "static_caching":
		return 1 // 2 fields: max age, etag
	case "ip_restricted":
		return 1 // 2 fields: lan subnet, external IP
	case "security_headers":
		return 2 // 3 presets
	case "compression_advanced":
		return 3 // 4 fields: level, gzip, zstd, brotli
	case "https_backend":
		return 1 // 2 fields: skip verify, keepalive
	case "performance":
		return 1 // 2 fields: gzip, zstd
	case "auth_headers":
		return 1 // 2 fields: forward IP, forward proto
	case "websocket_advanced":
		return 1 // 2 fields: upgrade timeout, ping interval
	case "custom_headers_inject":
		return 2 // 3 options: upstream, response, both
	case "rewrite_rules":
		return 1 // 2 fields: path pattern, rewrite to
	case "frame_embedding":
		return 0 // 1 field: allowed origins
	default:
		return 0
	}
}

// toggleSnippetTemplateParamCheckbox toggles checkbox parameters in the current template
func (m *Model) toggleSnippetTemplateParamCheckbox() {
	// Find the first selected template to configure
	templateKeys := []string{
		"cors_headers", "rate_limiting", "large_uploads", "extended_timeouts",
		"static_caching", "compression_advanced", "https_backend", "auth_headers",
		"websocket_advanced", "custom_headers_inject", "rewrite_rules", "frame_embedding",
		"ip_restricted", "security_headers", "performance",
	}

	var currentTemplate string
	for _, key := range templateKeys {
		if m.snippetWizardData.SelectedTemplates[key] {
			currentTemplate = key
			break
		}
	}

	if currentTemplate == "" {
		return
	}

	config := m.snippetWizardData.SnippetConfigs[currentTemplate]
	if config.Parameters == nil {
		config.Parameters = make(map[string]interface{})
	}

	// Toggle checkbox parameters based on template and cursor position
	switch currentTemplate {
	case "cors_headers":
		if m.wizardCursor == 2 {
			current := config.Parameters["allow_credentials"]
			if current == "true" {
				config.Parameters["allow_credentials"] = "false"
			} else {
				config.Parameters["allow_credentials"] = "true"
			}
		}
	case "static_caching":
		if m.wizardCursor == 1 {
			current := config.Parameters["enable_etag"]
			if current == "true" {
				config.Parameters["enable_etag"] = "false"
			} else {
				config.Parameters["enable_etag"] = "true"
			}
		}
	case "compression_advanced":
		switch m.wizardCursor {
		case 1: // gzip
			current := config.Parameters["enable_gzip"]
			if current == "false" {
				config.Parameters["enable_gzip"] = "true"
			} else {
				config.Parameters["enable_gzip"] = "false"
			}
		case 2: // zstd
			current := config.Parameters["enable_zstd"]
			if current == "false" {
				config.Parameters["enable_zstd"] = "true"
			} else {
				config.Parameters["enable_zstd"] = "false"
			}
		case 3: // brotli
			current := config.Parameters["enable_brotli"]
			if current == "true" {
				config.Parameters["enable_brotli"] = "false"
			} else {
				config.Parameters["enable_brotli"] = "true"
			}
		}
	case "https_backend":
		if m.wizardCursor == 0 {
			current := config.Parameters["skip_verify"]
			if current == "true" {
				config.Parameters["skip_verify"] = "false"
			} else {
				config.Parameters["skip_verify"] = "true"
			}
		}
	case "performance":
		switch m.wizardCursor {
		case 0: // gzip
			current := config.Parameters["enable_gzip"]
			if current == "false" {
				config.Parameters["enable_gzip"] = "true"
			} else {
				config.Parameters["enable_gzip"] = "false"
			}
		case 1: // zstd
			current := config.Parameters["enable_zstd"]
			if current == "false" {
				config.Parameters["enable_zstd"] = "true"
			} else {
				config.Parameters["enable_zstd"] = "false"
			}
		}
	case "auth_headers":
		switch m.wizardCursor {
		case 0: // forward IP
			current := config.Parameters["forward_ip"]
			if current == "true" {
				config.Parameters["forward_ip"] = "false"
			} else {
				config.Parameters["forward_ip"] = "true"
			}
		case 1: // forward proto
			current := config.Parameters["forward_proto"]
			if current == "true" {
				config.Parameters["forward_proto"] = "false"
			} else {
				config.Parameters["forward_proto"] = "true"
			}
		}
	}

	m.snippetWizardData.SnippetConfigs[currentTemplate] = config
}

