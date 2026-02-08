package ui

import (
	"fmt"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/audit"
	"lazyproxyflare/internal/caddy"
)

// handleAsyncMsg handles async operation result messages.
// Returns (model, cmd, handled) where handled indicates if the message was processed.
func (m Model) handleAsyncMsg(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case refreshCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.entries = msg.entries
			m.snippets = msg.snippets
			// Sort entries alphabetically by domain
			sort.Slice(m.entries, func(i, j int) bool {
				return m.entries[i].Domain < m.entries[j].Domain
			})
			m.cursor = 0
			m.scrollOffset = 0
			m.searchQuery = ""
			m.err = nil
		}
		return m, nil, true

	case migrationCompleteMsg:
		m2, cmd := m.handleMigrationComplete(msg)
		return m2, cmd, true

	case createEntryMsg:
		m.loading = false

		// Audit log the operation
		if m.audit.Logger != nil {
			fqdn := m.addForm.Subdomain + "." + m.config.Domain
			details := map[string]interface{}{
				"dns_type": m.addForm.DNSType,
				"target":   m.addForm.DNSTarget,
				"proxied":  m.addForm.Proxied,
				"dns_only": m.addForm.DNSOnly,
			}
			if !m.addForm.DNSOnly {
				details["reverse_proxy"] = m.addForm.ReverseProxyTarget
				details["port"] = m.addForm.ServicePort
			}

			result := audit.ResultSuccess
			errorMsg := ""
			if !msg.success {
				result = audit.ResultFailure
				errorMsg = fmt.Sprintf("%s: %v", msg.errorStep, msg.err)
			}

			entityType := audit.EntityBoth
			if m.addForm.DNSOnly {
				entityType = audit.EntityDNS
			}

			m.audit.Logger.Log(audit.LogEntry{
				Operation:  audit.OperationCreate,
				EntityType: entityType,
				Domain:     fqdn,
				Details:    details,
				Result:     result,
				Error:      errorMsg,
			})
		}

		if msg.success {
			// Success - return to list view, clear editing state, and refresh
			m.currentView = ViewList
			m.editingEntry = nil
			m.err = nil
			return m, refreshDataCmd(m.config), true
		} else {
			// Error - show error message, stay in preview
			m.err = fmt.Errorf("Failed at %s: %v", msg.errorStep, msg.err)
			return m, nil, true
		}

	case updateEntryMsg:
		m.loading = false

		// Audit log the operation
		if m.audit.Logger != nil && m.editingEntry != nil {
			fqdn := m.addForm.Subdomain + "." + m.config.Domain
			details := map[string]interface{}{
				"dns_type": m.addForm.DNSType,
				"target":   m.addForm.DNSTarget,
				"proxied":  m.addForm.Proxied,
				"dns_only": m.addForm.DNSOnly,
			}
			if !m.addForm.DNSOnly {
				details["reverse_proxy"] = m.addForm.ReverseProxyTarget
				details["port"] = m.addForm.ServicePort
			}

			result := audit.ResultSuccess
			errorMsg := ""
			if !msg.success {
				result = audit.ResultFailure
				errorMsg = fmt.Sprintf("%s: %v", msg.errorStep, msg.err)
			}

			// Determine entity type based on mode transitions
			entityType := audit.EntityBoth
			if m.addForm.DNSOnly && m.editingEntry.Caddy == nil {
				entityType = audit.EntityDNS // DNS-only to DNS-only
			} else if m.addForm.DNSOnly && m.editingEntry.Caddy != nil {
				entityType = audit.EntityBoth // Full to DNS-only (removed Caddy)
			} else if !m.addForm.DNSOnly && m.editingEntry.Caddy == nil {
				entityType = audit.EntityBoth // DNS-only to Full (added Caddy)
			}

			m.audit.Logger.Log(audit.LogEntry{
				Operation:  audit.OperationUpdate,
				EntityType: entityType,
				Domain:     fqdn,
				Details:    details,
				Result:     result,
				Error:      errorMsg,
			})
		}

		if msg.success {
			// Success - return to list view, clear editing state, and refresh
			m.currentView = ViewList
			m.editingEntry = nil
			m.err = nil
			return m, refreshDataCmd(m.config), true
		} else {
			// Error - show error message, stay in preview
			m.err = fmt.Errorf("Failed at %s: %v", msg.errorStep, msg.err)
			return m, nil, true
		}

	case deleteEntryMsg:
		m.loading = false

		// Audit log the operation
		if m.audit.Logger != nil && msg.domain != "" {
			result := audit.ResultSuccess
			errorMsg := ""
			if !msg.success {
				result = audit.ResultFailure
				errorMsg = fmt.Sprintf("%s: %v", msg.errorStep, msg.err)
			}

			var entityType audit.EntityType
			switch msg.entityType {
			case "dns":
				entityType = audit.EntityDNS
			case "caddy":
				entityType = audit.EntityCaddy
			default:
				entityType = audit.EntityBoth
			}

			m.audit.Logger.Log(audit.LogEntry{
				Operation:  audit.OperationDelete,
				EntityType: entityType,
				Domain:     msg.domain,
				Result:     result,
				Error:      errorMsg,
			})
		}

		if msg.success {
			// Success - return to list view and refresh
			m.currentView = ViewList
			m.err = nil
			return m, refreshDataCmd(m.config), true
		} else {
			// Error - show error message, stay in confirm delete
			m.err = fmt.Errorf("Failed at %s: %v", msg.errorStep, msg.err)
			return m, nil, true
		}

	case syncEntryMsg:
		m.loading = false

		// Audit log the operation
		if m.audit.Logger != nil && msg.domain != "" {
			result := audit.ResultSuccess
			errorMsg := ""
			if !msg.success {
				result = audit.ResultFailure
				errorMsg = fmt.Sprintf("%s: %v", msg.errorStep, msg.err)
			}

			// Determine entity type based on sync direction
			entityType := audit.EntityBoth
			if msg.syncType == "to_dns" {
				entityType = audit.EntityDNS
			} else if msg.syncType == "to_caddy" {
				entityType = audit.EntityCaddy
			}

			details := map[string]interface{}{
				"sync_direction": msg.syncType,
			}

			m.audit.Logger.Log(audit.LogEntry{
				Operation:  audit.OperationSync,
				EntityType: entityType,
				Domain:     msg.domain,
				Details:    details,
				Result:     result,
				Error:      errorMsg,
			})
		}

		if msg.success {
			// Success - return to list view and refresh
			m.currentView = ViewList
			m.err = nil
			return m, refreshDataCmd(m.config), true
		} else {
			// Error - show error message, stay in confirm sync
			m.err = fmt.Errorf("Failed at %s: %v", msg.errorStep, msg.err)
			return m, nil, true
		}

	case bulkDeleteMsg:
		m.loading = false

		// Audit log the operation
		if m.audit.Logger != nil && len(msg.deletedDomains) > 0 {
			result := audit.ResultSuccess
			errorMsg := ""
			if !msg.success {
				result = audit.ResultFailure
				errorMsg = fmt.Sprintf("%s: %v", msg.errorStep, msg.err)
			}

			var entityType audit.EntityType
			if msg.deleteType == "dns" {
				entityType = audit.EntityDNS
			} else if msg.deleteType == "caddy" {
				entityType = audit.EntityCaddy
			} else {
				entityType = audit.EntityBoth
			}

			// Determine operation type based on isSync flag
			operation := audit.OperationBatchDelete
			if msg.isSync {
				operation = audit.OperationBatchSync
			}

			// Log batch operation with first domain and batch count
			domain := msg.deletedDomains[0]
			if len(msg.deletedDomains) > 1 {
				domain = fmt.Sprintf("%s and %d others", domain, len(msg.deletedDomains)-1)
			}

			m.audit.Logger.Log(audit.LogEntry{
				Operation:  operation,
				EntityType: entityType,
				Domain:     domain,
				BatchCount: len(msg.deletedDomains),
				Result:     result,
				Error:      errorMsg,
			})
		}

		if msg.success {
			// Success - return to list view, clear bulk delete state and selections, and refresh
			m.currentView = ViewList
			m.bulkDelete.Entries = nil
			m.selectedEntries = make(map[string]bool) // Clear selections after batch operations
			m.err = nil
			// Show success message with count
			return m, refreshDataCmd(m.config), true
		} else {
			// Error - show error message with count of entries deleted before failure
			if msg.count > 0 {
				m.err = fmt.Errorf("Deleted %d entries before error at %s: %v", msg.count, msg.errorStep, msg.err)
			} else {
				m.err = fmt.Errorf("Failed at %s: %v", msg.errorStep, msg.err)
			}
			return m, nil, true
		}

	case restoreBackupMsg:
		m.loading = false

		// Audit log the restore operation
		if m.audit.Logger != nil {
			result := audit.ResultSuccess
			errorMsg := ""
			if !msg.success {
				result = audit.ResultFailure
				errorMsg = msg.err.Error()
			}

			var entityType audit.EntityType
			switch msg.scope {
			case RestoreDNSOnly:
				entityType = audit.EntityDNS
			case RestoreCaddyOnly:
				entityType = audit.EntityCaddy
			case RestoreAll:
				entityType = audit.EntityBoth
			default:
				entityType = audit.EntityBoth
			}

			// Extract filename from backup path for domain field
			backupFilename := filepath.Base(msg.backupPath)

			details := map[string]interface{}{
				"backup_file": backupFilename,
				"scope":       msg.scope.String(),
			}

			m.audit.Logger.Log(audit.LogEntry{
				Operation:  audit.OperationRestore,
				EntityType: entityType,
				Domain:     fmt.Sprintf("Backup: %s", backupFilename),
				Details:    details,
				Result:     result,
				Error:      errorMsg,
			})
		}

		if msg.success {
			// Success - return to list view and refresh data
			m.currentView = ViewList
			m.err = nil
			return m, refreshDataCmd(m.config), true
		} else {
			// Error - stay in confirm restore view with error
			m.err = msg.err
			return m, nil, true
		}

	case deleteBackupMsg:
		m.loading = false
		if msg.success {
			// Success - stay in backup manager, reset cursor if needed
			backups, err := caddy.ListBackups(m.config.Caddy.CaddyfilePath)
			if err == nil && m.backup.Cursor >= len(backups) && m.backup.Cursor > 0 {
				m.backup.Cursor--
			}
			m.err = nil
			return m, nil, true
		} else {
			// Error - show error message
			m.err = msg.err
			return m, nil, true
		}

	case cleanupBackupsMsg:
		m.loading = false
		if msg.success {
			// Success - return to backup manager, reset cursor
			m.currentView = ViewBackupManager
			m.backup.Cursor = 0
			m.backup.ScrollOffset = 0
			m.err = nil
			return m, nil, true
		} else {
			// Error - stay in confirm cleanup view with error
			m.err = msg.err
			return m, nil, true
		}

	default:
		return m, nil, false
	}
}
