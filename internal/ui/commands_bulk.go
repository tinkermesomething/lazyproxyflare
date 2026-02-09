package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/cloudflare"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
)

type bulkDeleteMsg struct {
	success        bool
	err            error
	count          int      // Number of entries deleted/synced
	errorStep      string   // Which step failed
	backupPath     string   // Track backup path (for Caddy bulk delete)
	deleteType     string   // "dns", "caddy", or "both" for audit logging
	deletedDomains []string // Domains that were deleted/synced for audit logging
	isSync         bool     // true for sync operations, false for delete operations
}

// bulkDeleteDNSCmd deletes all orphaned DNS records (DNS exists but no Caddy)
func bulkDeleteDNSCmd(cfg *config.Config, entries []diff.SyncedEntry, apiToken string) tea.Cmd {
	return func() tea.Msg {
		cfClient := cloudflare.NewClient(apiToken)
		deletedCount := 0
		deletedDomains := []string{}

		// Delete each orphaned DNS record
		for _, entry := range entries {
			if entry.Status != diff.StatusOrphanedDNS || entry.DNS == nil {
				continue
			}

			err := cfClient.DeleteDNSRecord(cfg.Cloudflare.ZoneID, entry.DNS.ID)
			if err != nil {
				return bulkDeleteMsg{
					success:        false,
					err:            err,
					count:          deletedCount,
					errorStep:      fmt.Sprintf("dns_delete_%s", entry.Domain),
					deleteType:     "dns",
					deletedDomains: deletedDomains,
				}
			}
			deletedCount++
			deletedDomains = append(deletedDomains, entry.Domain)
		}

		return bulkDeleteMsg{
			success:        true,
			count:          deletedCount,
			deleteType:     "dns",
			deletedDomains: deletedDomains,
		}
	}
}

// bulkDeleteCaddyCmd deletes all orphaned Caddy entries (Caddy exists but no DNS)
func bulkDeleteCaddyCmd(cfg *config.Config, entries []diff.SyncedEntry) tea.Cmd {
	return func() tea.Msg {
		var backupPath string
		var err error

		// Step 1: Backup Caddyfile
		backupPath, err = caddy.BackupCaddyfile(cfg.Caddy.CaddyfilePath)
		if err != nil {
			return bulkDeleteMsg{
				success:    false,
				err:        err,
				errorStep:  "backup",
				deleteType: "caddy",
			}
		}

		deletedCount := 0
		deletedDomains := []string{}

		// Step 2: Remove each orphaned Caddy entry
		for _, entry := range entries {
			if entry.Status != diff.StatusOrphanedCaddy || entry.Caddy == nil {
				continue
			}

			err = caddy.RemoveEntry(cfg.Caddy.CaddyfilePath, entry.Domain)
			if err != nil {
				// Rollback: Restore Caddyfile
				err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy remove")
				return bulkDeleteMsg{
					success:        false,
					err:            err,
					count:          deletedCount,
					errorStep:      fmt.Sprintf("caddy_remove_%s", entry.Domain),
					backupPath:     backupPath,
					deleteType:     "caddy",
					deletedDomains: deletedDomains,
				}
			}
			deletedCount++
			deletedDomains = append(deletedDomains, entry.Domain)
		}

		// Step 3: Validate Caddyfile
		err = formatAndValidateCaddyfile(cfg)
		if err != nil {
			// Rollback: Restore Caddyfile
			err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy validate")
			return bulkDeleteMsg{
				success:        false,
				err:            err,
				count:          deletedCount,
				errorStep:      "caddy_validate",
				backupPath:     backupPath,
				deleteType:     "caddy",
				deletedDomains: deletedDomains,
			}
		}

		// Step 4: Restart Caddy
		err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
		if err != nil {
			// Rollback: Restore Caddyfile
			err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy restart")
			return bulkDeleteMsg{
				success:        false,
				err:            err,
				count:          deletedCount,
				errorStep:      "caddy_restart",
				backupPath:     backupPath,
				deleteType:     "caddy",
				deletedDomains: deletedDomains,
			}
		}

		return bulkDeleteMsg{
			success:        true,
			count:          deletedCount,
			backupPath:     backupPath,
			deleteType:     "caddy",
			deletedDomains: deletedDomains,
		}
	}
}

// batchDeleteSelectedCmd deletes all selected entries
func batchDeleteSelectedCmd(cfg *config.Config, allEntries []diff.SyncedEntry, selectedDomains map[string]bool, apiToken string) tea.Cmd {
	return func() tea.Msg {
		var backupPath string
		var err error
		var dnsDeletedCount, caddyDeletedCount int
		deletedDomains := []string{}

		// Collect selected entries
		var selectedEntries []diff.SyncedEntry
		for _, entry := range allEntries {
			if selectedDomains[entry.Domain] {
				selectedEntries = append(selectedEntries, entry)
			}
		}

		// Step 1: Backup Caddyfile if any selected entries have Caddy configs
		needsBackup := false
		for _, entry := range selectedEntries {
			if entry.Caddy != nil {
				needsBackup = true
				break
			}
		}

		if needsBackup {
			backupPath, err = caddy.BackupCaddyfile(cfg.Caddy.CaddyfilePath)
			if err != nil {
				return bulkDeleteMsg{
					success:    false,
					err:        err,
					errorStep:  "backup",
					deleteType: "both",
				}
			}
		}

		// Step 2: Remove Caddy entries
		for _, entry := range selectedEntries {
			if entry.Caddy != nil {
				err = caddy.RemoveEntry(cfg.Caddy.CaddyfilePath, entry.Domain)
				if err != nil {
					if backupPath != "" {
						err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy remove")
					}
					return bulkDeleteMsg{
						success:        false,
						err:            err,
						count:          caddyDeletedCount,
						errorStep:      fmt.Sprintf("caddy_remove_%s", entry.Domain),
						backupPath:     backupPath,
						deleteType:     "both",
						deletedDomains: deletedDomains,
					}
				}
				caddyDeletedCount++
				deletedDomains = append(deletedDomains, entry.Domain)
			}
		}

		// Step 3: Validate and restart Caddy if we modified it
		if caddyDeletedCount > 0 {
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy validate")
				return bulkDeleteMsg{
					success:        false,
					err:            err,
					count:          caddyDeletedCount,
					errorStep:      "caddy_validate",
					backupPath:     backupPath,
					deleteType:     "both",
					deletedDomains: deletedDomains,
				}
			}

			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy restart")
				return bulkDeleteMsg{
					success:        false,
					err:            err,
					count:          caddyDeletedCount,
					errorStep:      "caddy_restart",
					backupPath:     backupPath,
					deleteType:     "both",
					deletedDomains: deletedDomains,
				}
			}
		}

		// Step 4: Delete DNS records
		cfClient := cloudflare.NewClient(apiToken)
		for _, entry := range selectedEntries {
			if entry.DNS != nil {
				err = cfClient.DeleteDNSRecord(cfg.Cloudflare.ZoneID, entry.DNS.ID)
				if err != nil {
					return bulkDeleteMsg{
						success:        false,
						err:            err,
						count:          dnsDeletedCount + caddyDeletedCount,
						errorStep:      fmt.Sprintf("dns_delete_%s", entry.Domain),
						deleteType:     "both",
						deletedDomains: deletedDomains,
					}
				}
				dnsDeletedCount++
				// Add domain if not already added (in case it has both DNS and Caddy)
				alreadyAdded := false
				for _, d := range deletedDomains {
					if d == entry.Domain {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					deletedDomains = append(deletedDomains, entry.Domain)
				}
			}
		}

		return bulkDeleteMsg{
			success:        true,
			count:          dnsDeletedCount + caddyDeletedCount,
			backupPath:     backupPath,
			deleteType:     "both",
			deletedDomains: deletedDomains,
		}
	}
}

// batchSyncSelectedCmd syncs all selected entries by creating missing DNS or Caddy
func batchSyncSelectedCmd(cfg *config.Config, allEntries []diff.SyncedEntry, selectedDomains map[string]bool, apiToken string) tea.Cmd {
	return func() tea.Msg {
		var backupPath string
		var err error
		var syncedCount int
		syncedDomains := []string{}

		// Collect selected entries
		var selectedEntries []diff.SyncedEntry
		for _, entry := range allEntries {
			if selectedDomains[entry.Domain] {
				selectedEntries = append(selectedEntries, entry)
			}
		}

		// Step 1: Backup Caddyfile (we might add Caddy entries)
		backupPath, err = caddy.BackupCaddyfile(cfg.Caddy.CaddyfilePath)
		if err != nil {
			return bulkDeleteMsg{
				success:    false,
				err:        err,
				errorStep:  "backup",
				isSync:     true,
				deleteType: "both",
			}
		}

		caddyModified := false
		cfClient := cloudflare.NewClient(apiToken)

		// Step 2: Process each selected entry
		for _, entry := range selectedEntries {
			if entry.Status == diff.StatusOrphanedDNS && entry.DNS != nil {
				// Create Caddy entry
				caddyBlock := caddy.GenerateCaddyBlock(caddy.GenerateBlockInput{
					FQDN:              entry.Domain,
					Target:            "localhost",
					Port:              cfg.Defaults.Port,
					SSL:               cfg.Defaults.SSL,
					LANOnly:           false,
					OAuth:             false,
					WebSocket:         false,
					LANSubnet:         cfg.Defaults.LANSubnet,
					AllowedExtIP:      cfg.Defaults.AllowedExternalIP,
					AvailableSnippets: nil,
					CustomCaddyConfig: "", // No custom config for batch sync
				})

				err = caddy.AppendEntry(cfg.Caddy.CaddyfilePath, caddyBlock)
				if err != nil {
					err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy remove")
					return bulkDeleteMsg{
						success:        false,
						err:            err,
						count:          syncedCount,
						errorStep:      fmt.Sprintf("caddy_append_%s", entry.Domain),
						backupPath:     backupPath,
						isSync:         true,
						deleteType:     "both",
						deletedDomains: syncedDomains,
					}
				}
				caddyModified = true
				syncedCount++
				syncedDomains = append(syncedDomains, entry.Domain)

			} else if entry.Status == diff.StatusOrphanedCaddy && entry.Caddy != nil {
				// Create DNS record
				dnsRecord := cloudflare.DNSRecord{
					Type:    "CNAME",
					Name:    entry.Domain,
					Content: cfg.Defaults.CNAMETarget,
					Proxied: cfg.Defaults.Proxied,
					TTL:     1,
				}

				_, err = cfClient.CreateDNSRecord(cfg.Cloudflare.ZoneID, dnsRecord)
				if err != nil {
					if caddyModified {
						err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy remove")
					}
					return bulkDeleteMsg{
						success:        false,
						err:            err,
						count:          syncedCount,
						errorStep:      fmt.Sprintf("dns_create_%s", entry.Domain),
						isSync:         true,
						deleteType:     "both",
						deletedDomains: syncedDomains,
					}
				}
				syncedCount++
				syncedDomains = append(syncedDomains, entry.Domain)
			}
		}

		// Step 3: Validate and restart Caddy if modified
		if caddyModified {
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy validate")
				return bulkDeleteMsg{
					success:        false,
					err:            err,
					count:          syncedCount,
					errorStep:      "caddy_validate",
					backupPath:     backupPath,
					isSync:         true,
					deleteType:     "both",
					deletedDomains: syncedDomains,
				}
			}

			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				err = restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "caddy restart")
				return bulkDeleteMsg{
					success:        false,
					err:            err,
					count:          syncedCount,
					errorStep:      "caddy_restart",
					backupPath:     backupPath,
					isSync:         true,
					deleteType:     "both",
					deletedDomains: syncedDomains,
				}
			}
		}

		return bulkDeleteMsg{
			success:        true,
			count:          syncedCount,
			backupPath:     backupPath,
			isSync:         true,
			deleteType:     "both",
			deletedDomains: syncedDomains,
		}
	}
}
