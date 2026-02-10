package ui

import (
	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/diff"
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

type editorFinishedMsg struct {
	err error
}

type exportProfileMsg struct {
	success bool
	err     error
	path    string
}

type importProfileMsg struct {
	success     bool
	err         error
	profileName string
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
