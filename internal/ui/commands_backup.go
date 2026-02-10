package ui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/cloudflare"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
)

type restoreBackupMsg struct {
	success    bool
	err        error
	scope      RestoreScope // What was restored (All/DNS/Caddy)
	backupPath string       // Backup file that was restored
}

type deleteBackupMsg struct {
	success bool
	err     error
}

type cleanupBackupsMsg struct {
	success      bool
	err          error
	deletedCount int
}

// restoreBackupCmd restores from backup based on the specified scope
func restoreBackupCmd(cfg *config.Config, backupPath string, scope RestoreScope, apiToken string) tea.Cmd {
	return func() tea.Msg {
		// Handle based on restore scope
		switch scope {
		case RestoreCaddyOnly:
			// Only restore Caddyfile
			err := caddy.RestoreFromBackup(cfg.Caddy.CaddyfilePath, backupPath)
			if err != nil {
				return restoreBackupMsg{
					success:    false,
					err:        fmt.Errorf("restore failed: %w", err),
					scope:      scope,
					backupPath: backupPath,
				}
			}

			// Validate restored Caddyfile
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				return restoreBackupMsg{
					success:    false,
					err:        fmt.Errorf("restored Caddyfile validation failed: %w", err),
					scope:      scope,
					backupPath: backupPath,
				}
			}

			// Restart Caddy
			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				return restoreBackupMsg{
					success:    false,
					err:        fmt.Errorf("caddy restart failed: %w", err),
					scope:      scope,
					backupPath: backupPath,
				}
			}

		case RestoreDNSOnly:
			// Only restore DNS records from backup
			err := restoreDNSFromBackup(cfg, backupPath, apiToken)
			if err != nil {
				return restoreBackupMsg{
					success:    false,
					err:        fmt.Errorf("DNS restore failed: %w", err),
					scope:      scope,
					backupPath: backupPath,
				}
			}

		case RestoreAll:
			// Restore Caddyfile
			err := caddy.RestoreFromBackup(cfg.Caddy.CaddyfilePath, backupPath)
			if err != nil {
				return restoreBackupMsg{
					success:    false,
					err:        fmt.Errorf("Caddyfile restore failed: %w", err),
					scope:      scope,
					backupPath: backupPath,
				}
			}

			// Validate restored Caddyfile
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				return restoreBackupMsg{
					success:    false,
					err:        fmt.Errorf("restored Caddyfile validation failed: %w", err),
					scope:      scope,
					backupPath: backupPath,
				}
			}

			// Restart Caddy
			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				return restoreBackupMsg{
					success:    false,
					err:        fmt.Errorf("caddy restart failed: %w", err),
					scope:      scope,
					backupPath: backupPath,
				}
			}

			// Restore DNS records
			err = restoreDNSFromBackup(cfg, backupPath, apiToken)
			if err != nil {
				return restoreBackupMsg{
					success:    false,
					err:        fmt.Errorf("DNS restore failed: %w", err),
					scope:      scope,
					backupPath: backupPath,
				}
			}
		}

		return restoreBackupMsg{
			success:    true,
			scope:      scope,
			backupPath: backupPath,
		}
	}
}

// restoreDNSFromBackup parses the backup Caddyfile and creates/updates DNS records
func restoreDNSFromBackup(cfg *config.Config, backupPath string, apiToken string) error {
	// Parse the backup Caddyfile to extract domains
	entries, err := caddy.ParseCaddyfile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to parse backup: %w", err)
	}

	// Initialize Cloudflare client
	cf := cloudflare.NewClient(apiToken)

	// Create/update DNS records for each domain in the backup
	for _, entry := range entries {
		for _, domain := range entry.Domains {
			// Create DNS record with default settings from config
			// Determine DNS type based on domain or use default
			dnsType := "CNAME"
			content := cfg.Defaults.CNAMETarget

			// Create the DNS record
			record := cloudflare.DNSRecord{
				Name:    domain,
				Type:    dnsType,
				Content: content,
				Proxied: cfg.Defaults.Proxied,
				TTL:     1, // Auto TTL
			}

			_, err := cf.CreateDNSRecord(cfg.Cloudflare.ZoneID, record)
			if err != nil {
				// If record already exists, try to update it
				// Note: This is a simplified approach - in production you might want to
				// fetch existing records first and update them
				return fmt.Errorf("failed to create DNS record for %s: %w", domain, err)
			}
		}
	}

	return nil
}

// deleteBackupCmd deletes a backup file
func deleteBackupCmd(backupPath string) tea.Cmd {
	return func() tea.Msg {
		err := os.Remove(backupPath)
		if err != nil {
			return deleteBackupMsg{
				success: false,
				err:     fmt.Errorf("delete failed: %w", err),
			}
		}
		return deleteBackupMsg{success: true}
	}
}

// cleanupBackupsCmd deletes backups older than the specified retention period
func cleanupBackupsCmd(caddyfilePath string, retentionDays int, maxBackups int, maxSizeMB int) tea.Cmd {
	return func() tea.Msg {
		totalDeleted := 0

		// Age-based cleanup
		maxAge := time.Duration(retentionDays) * 24 * time.Hour
		oldBackups, err := caddy.GetOldBackups(caddyfilePath, maxAge)
		if err != nil {
			return cleanupBackupsMsg{
				success: false,
				err:     fmt.Errorf("failed to get old backups: %w", err),
			}
		}
		totalDeleted += len(oldBackups)

		err = caddy.CleanupOldBackups(caddyfilePath, maxAge)
		if err != nil {
			return cleanupBackupsMsg{
				success: false,
				err:     err,
			}
		}

		// Count-based cleanup
		if maxBackups > 0 {
			n, err := caddy.CleanupByCount(caddyfilePath, maxBackups)
			if err == nil {
				totalDeleted += n
			}
		}

		// Size-based cleanup
		if maxSizeMB > 0 {
			n, err := caddy.CleanupBySize(caddyfilePath, maxSizeMB)
			if err == nil {
				totalDeleted += n
			}
		}

		return cleanupBackupsMsg{
			success:      true,
			deletedCount: totalDeleted,
		}
	}
}

// refreshDataCmd fetches fresh data from Cloudflare and Caddyfile
func refreshDataCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		// Get API token
		apiToken, err := cfg.GetAPIToken()
		if err != nil {
			return refreshCompleteMsg{err: fmt.Errorf("failed to get API token: %w", err)}
		}

		// Parse Caddyfile
		caddyContent, err := os.ReadFile(cfg.Caddy.CaddyfilePath)
		if err != nil {
			// Try local Caddyfile if configured path fails
			caddyContent, err = os.ReadFile("Caddyfile")
			if err != nil {
				return refreshCompleteMsg{err: err}
			}
		}

		// Parse with snippets
		parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))

		// Fetch DNS records from Cloudflare
		cfClient := cloudflare.NewClient(apiToken)

		cnameRecords, err := cfClient.ListDNSRecords(cfg.Cloudflare.ZoneID, "CNAME")
		if err != nil {
			return refreshCompleteMsg{err: err}
		}

		aRecords, err := cfClient.ListDNSRecords(cfg.Cloudflare.ZoneID, "A")
		if err != nil {
			return refreshCompleteMsg{err: err}
		}

		// Combine all DNS records
		allDNS := append(cnameRecords, aRecords...)

		// Run diff engine
		syncedEntries := diff.Compare(allDNS, parsed.Entries)

		return refreshCompleteMsg{entries: syncedEntries, snippets: parsed.Snippets, err: nil}
	}
}
