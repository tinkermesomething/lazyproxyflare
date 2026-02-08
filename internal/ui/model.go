package ui

import (
	"lazyproxyflare/internal/audit"
	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

// ViewMode represents the current view
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetails
	ViewHelp
	ViewAdd
	ViewEdit
	ViewPreview
	ViewDeleteScope
	ViewConfirmDelete
	ViewConfirmSync
	ViewBulkDeleteMenu
	ViewConfirmBulkDelete
	ViewConfirmBatchDelete
	ViewConfirmBatchSync
	ViewBackupManager
	ViewBackupPreview
	ViewRestoreScope
	ViewConfirmRestore
	ViewConfirmCleanup
	ViewAuditLog
	ViewSnippetDetail
	ViewProfileSelector
	ViewProfileEdit
	ViewWizard
	ViewSnippetWizard
	ViewMigrationWizard
	ViewError
)

// RestoreScope represents what to restore from backup
type RestoreScope int

const (
	RestoreAll RestoreScope = iota // Restore both Caddyfile and DNS
	RestoreDNSOnly                  // Restore DNS records only
	RestoreCaddyOnly                // Restore Caddyfile only
)

// String returns human-readable restore scope name
func (r RestoreScope) String() string {
	switch r {
	case RestoreAll:
		return "All (DNS + Caddyfile)"
	case RestoreDNSOnly:
		return "DNS Records Only"
	case RestoreCaddyOnly:
		return "Caddyfile Only"
	default:
		return "Unknown"
	}
}

// Description returns detailed description of restore scope
func (r RestoreScope) Description() string {
	switch r {
	case RestoreAll:
		return "Restore both Caddyfile configuration and DNS records from backup"
	case RestoreDNSOnly:
		return "Restore DNS records only, leave current Caddyfile unchanged"
	case RestoreCaddyOnly:
		return "Restore Caddyfile only, leave DNS records unchanged"
	default:
		return ""
	}
}

// DeleteScope represents what to delete from an entry
type DeleteScope int

const (
	DeleteAll DeleteScope = iota // Delete both DNS and Caddy
	DeleteDNSOnly                 // Delete DNS record only
	DeleteCaddyOnly               // Delete Caddyfile entry only
)

// String returns human-readable delete scope name
func (d DeleteScope) String() string {
	switch d {
	case DeleteAll:
		return "All (DNS + Caddyfile)"
	case DeleteDNSOnly:
		return "DNS Record Only"
	case DeleteCaddyOnly:
		return "Caddyfile Entry Only"
	default:
		return "Unknown"
	}
}

// Description returns detailed description of delete scope
func (d DeleteScope) Description() string {
	switch d {
	case DeleteAll:
		return "Delete both DNS record from Cloudflare and Caddy configuration"
	case DeleteDNSOnly:
		return "Delete DNS record from Cloudflare only, keep Caddyfile unchanged"
	case DeleteCaddyOnly:
		return "Delete Caddyfile entry only, keep DNS record in Cloudflare"
	default:
		return ""
	}
}

// FilterMode represents the current status filter
type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterSynced
	FilterOrphanedDNS
	FilterOrphanedCaddy
)

// String returns human-readable filter name
func (f FilterMode) String() string {
	switch f {
	case FilterAll:
		return "All"
	case FilterSynced:
		return "Synced"
	case FilterOrphanedDNS:
		return "Orphaned DNS"
	case FilterOrphanedCaddy:
		return "Orphaned Caddy"
	default:
		return "Unknown"
	}
}

// DNSTypeFilter represents the current DNS type filter
type DNSTypeFilter int

const (
	DNSTypeAll DNSTypeFilter = iota
	DNSTypeCNAME
	DNSTypeA
)

// String returns human-readable DNS type filter name
func (d DNSTypeFilter) String() string {
	switch d {
	case DNSTypeAll:
		return "All"
	case DNSTypeCNAME:
		return "CNAME"
	case DNSTypeA:
		return "A"
	default:
		return "Unknown"
	}
}

// SortMode represents the current sort order
type SortMode int

const (
	SortAlphabetical SortMode = iota
	SortByStatus
)

// ActiveTab represents which main view tab is active
type ActiveTab int

const (
	TabCloudflare ActiveTab = iota // DNS/Cloudflare information
	TabCaddy                        // Caddy/reverse proxy configuration
)

// String returns human-readable tab name
func (t ActiveTab) String() string {
	switch t {
	case TabCloudflare:
		return "Cloudflare"
	case TabCaddy:
		return "Caddy"
	default:
		return "Unknown"
	}
}

// String returns human-readable sort mode name
func (s SortMode) String() string {
	switch s {
	case SortAlphabetical:
		return "Alphabetical"
	case SortByStatus:
		return "By Status"
	default:
		return "Unknown"
	}
}

// DeleteState holds state for single-entry deletion
type DeleteState struct {
	EntryIndex  int         // Index of entry being deleted
	Scope       DeleteScope // What to delete (All/DNS/Caddy)
	ScopeCursor int         // Cursor for delete scope selection (0-2)
}

// SyncState holds state for single-entry sync
type SyncState struct {
	Entry *diff.SyncedEntry // Entry being synced (stored when 's' pressed)
}

// AuditState holds state for the audit log viewer
type AuditState struct {
	Logger *audit.Logger    // Audit logger instance
	Logs   []audit.LogEntry // Loaded audit log entries
	Cursor int              // Currently selected log entry
	Scroll int              // Scroll offset for audit log view
}

// BackupState holds state for the backup manager
type BackupState struct {
	Cursor        int          // Currently selected backup
	ScrollOffset  int          // For scrolling backup list
	PreviewPath   string       // Path of backup being previewed/restored
	PreviewScroll int          // Scroll offset for backup preview content
	RetentionDays int          // Days to keep backups (for cleanup)
	RestoreScope      RestoreScope // What to restore (All/DNS/Caddy)
	RestoreScopeCursor int          // Cursor for restore scope selection (0-2)
}

// BulkDeleteState holds state for bulk deletion operations
type BulkDeleteState struct {
	Type       string             // "dns" or "caddy"
	MenuCursor int                // Menu navigation cursor
	Entries    []diff.SyncedEntry // Entries to be bulk deleted
}

// AddFormData represents the state of the add entry form
// ProfileEditData holds the form data for editing a profile
type ProfileEditData struct {
	// Original profile name (for rename detection)
	OriginalName string

	// Editable fields
	Name          string // Profile name
	APIToken      string // Cloudflare API token
	ZoneID        string // Cloudflare zone ID
	Domain        string // Domain name
	CaddyfilePath string // Host path to Caddyfile
	ContainerPath string // Container path to Caddyfile
	ContainerName string // Docker container name

	// Defaults
	CNAMETarget string // Default CNAME target
	Port        string // Default port
	SSL         bool   // Default SSL setting
	Proxied     bool   // Default proxied setting
}

type AddFormData struct {
	Subdomain          string
	DNSType            string // "CNAME" or "A"
	DNSTarget          string // CNAME target or A record IP
	DNSOnly            bool   // If true, create DNS record only (skip Caddy)
	ReverseProxyTarget string
	ServicePort        string
	Proxied            bool
	LANOnly            bool
	SSL                bool
	OAuth              bool
	WebSocket          bool
	SelectedSnippets   map[string]bool // Map of snippet name -> selected
	CustomCaddyConfig  string          // Custom Caddy directives (one-off features)
	FocusedField       int             // Which field is currently focused (0-10 + num snippets + custom config)
}

// Model represents the Bubbletea application state
type Model struct {
	// Data
	entries  []diff.SyncedEntry
	snippets []caddy.Snippet // Available Caddy snippets with metadata
	config   *config.Config

	// UI State
	currentView   ViewMode      // Current view mode
	cursor        int           // Currently selected item
	scrollOffset  int           // For scrolling long lists
	width         int           // Terminal width
	height        int           // Terminal height
	err           error         // Last error
	quitting      bool          // Whether we're quitting
	searchQuery   string        // Current search query
	searching     bool          // Whether in search mode
	loading       bool          // Whether data is being refreshed
	statusFilter  FilterMode    // Current status filter
	dnsTypeFilter DNSTypeFilter // Current DNS type filter
	sortMode      SortMode      // Current sort order
	activeTab     ActiveTab     // Current main view tab (Cloudflare/Caddy)

	// Multi-select state
	selectedEntries map[string]bool // Track selected entries by domain name

	// Form data
	addForm          AddFormData       // Add/Edit entry form state
	editingEntry     *diff.SyncedEntry // Entry being edited (nil if adding new)
	delete DeleteState // Single-entry deletion state
	sync   SyncState   // Single-entry sync state

	// Bulk delete state
	bulkDelete BulkDeleteState

	// Backup manager state
	backup BackupState

	// Audit log state
	audit AuditState

	// Panel state
	panelFocus PanelFocus // Which panel is focused (left or right)

	// Profile state
	currentProfileName string   // Name of currently loaded profile
	availableProfiles  []string // List of available profile names

	// Profile edit state
	profileEditData   ProfileEditData // Profile edit form data
	profileEditCursor int             // Current field in edit form

	// Wizard state
	wizardStep            WizardStep             // Current wizard step
	wizardData            WizardData             // Data collected during wizard
	wizardCursor          int                    // Cursor for selections in wizard
	wizardTextInput       textinput.Model        // Text input component for wizard (supports paste)
	wizardDockerContainers []caddy.DockerContainer // Detected Docker containers for wizard

	// Snippet panel state
	snippetCursor         int                     // Currently selected snippet
	snippetScrollOffset   int                     // For scrolling snippet list
	snippetCategoryFilter caddy.SnippetCategory   // Filter by category (SnippetUnknown = no filter)

	// Snippet editing state
	editingSnippet      bool            // Whether we're in snippet edit mode
	snippetEditTextarea textarea.Model  // Textarea for editing snippet content
	editingSnippetIndex int             // Index of snippet being edited

	// Snippet wizard state
	snippetWizardStep SnippetWizardStep // Current snippet wizard step
	snippetWizardData SnippetWizardData // Data collected during snippet wizard

	// Migration wizard state
	migrationWizardActive bool                 // True when migration wizard is active
	migrationWizardData   *MigrationWizardData // Migration wizard data

	// Error modal state
	previousView ViewMode // Previous view before showing error modal

	// Help modal pagination
	helpPage int // Current help page (0-4 for pages 1-5)
}
