package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/cloudflare"
	"lazyproxyflare/internal/config"
	"lazyproxyflare/internal/diff"
)

// restoreBackupWithError wraps caddy.RestoreFromBackup with proper error handling
// Returns an error message indicating whether restore succeeded or failed
func restoreBackupWithError(caddyfilePath, backupPath string, originalErr error, operation string) error {
	if restoreErr := caddy.RestoreFromBackup(caddyfilePath, backupPath); restoreErr != nil {
		return fmt.Errorf("CRITICAL: %s failed AND backup restore failed: %w (original error: %v)", operation, restoreErr, originalErr)
	}
	return fmt.Errorf("%s failed (backup restored): %w", operation, originalErr)
}

// formatAndValidateCaddyfile formats and validates the Caddyfile using config values
func formatAndValidateCaddyfile(cfg *config.Config) error {
	return caddy.FormatAndValidateCaddyfile(
		cfg.Caddy.CaddyfilePath,
		cfg.Caddy.CaddyfileContainerPath,
		cfg.Caddy.ContainerName,
		cfg.Caddy.DockerMethod,
		cfg.Caddy.ComposeFilePath,
		cfg.Caddy.ValidationCommand,
	)
}

// deleteEntryCmd deletes a DNS record and Caddy entry with rollback on failure
func deleteEntryCmd(cfg *config.Config, entry diff.SyncedEntry, scope DeleteScope, apiToken string) tea.Cmd {
	return func() tea.Msg {
		var backupPath string
		var err error

		// Determine what to delete based on scope
		deleteDNS := false
		deleteCaddy := false

		switch scope {
		case DeleteAll:
			deleteDNS = entry.DNS != nil
			deleteCaddy = entry.Caddy != nil
		case DeleteDNSOnly:
			deleteDNS = entry.DNS != nil
		case DeleteCaddyOnly:
			deleteCaddy = entry.Caddy != nil
		}

		// Determine entity type for audit logging
		entityType := "both"
		if deleteDNS && !deleteCaddy {
			entityType = "dns"
		} else if !deleteDNS && deleteCaddy {
			entityType = "caddy"
		}

		// Step 1: Backup Caddyfile (if deleting Caddy entry)
		if deleteCaddy {
			backupPath, err = caddy.BackupCaddyfile(cfg.Caddy.CaddyfilePath)
			if err != nil {
				return deleteEntryMsg{
					success:    false,
					err:        err,
					errorStep:  "backup",
					domain:     entry.Domain,
					entityType: entityType,
				}
			}
		}

		// Step 2: Remove from Caddyfile (if deleting Caddy)
		if deleteCaddy {
			err = caddy.RemoveEntry(cfg.Caddy.CaddyfilePath, entry.Domain)
			if err != nil {
				return deleteEntryMsg{
					success:    false,
					err:        err,
					errorStep:  "caddy_remove",
					backupPath: backupPath,
					domain:     entry.Domain,
					entityType: entityType,
				}
			}

			// Step 3: Validate Caddyfile
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				// Rollback: Restore Caddyfile
				return deleteEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddyfile validation"),
					errorStep:  "caddy_validate",
					backupPath: backupPath,
					domain:     entry.Domain,
					entityType: entityType,
				}
			}

			// Step 4: Restart Caddy
			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				// Rollback: Restore Caddyfile
				return deleteEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddy restart"),
					errorStep:  "caddy_restart",
					backupPath: backupPath,
					domain:     entry.Domain,
					entityType: entityType,
				}
			}
		}

		// Step 5: Delete DNS record (if deleting DNS)
		if deleteDNS {
			cfClient := cloudflare.NewClient(apiToken)
			err = cfClient.DeleteDNSRecord(cfg.Cloudflare.ZoneID, entry.DNS.ID)
			if err != nil {
				// If Caddy was removed, we're in an inconsistent state
				// But we can't rollback Caddy at this point (already restarted)
				return deleteEntryMsg{
					success:    false,
					err:        err,
					errorStep:  "dns_delete",
					backupPath: backupPath,
					domain:     entry.Domain,
					entityType: entityType,
				}
			}
		}

		// Success!
		return deleteEntryMsg{
			success:    true,
			backupPath: backupPath,
			domain:     entry.Domain,
			entityType: entityType,
		}
	}
}

// createEntryCmd creates a new DNS record and Caddy entry with rollback on failure
func createEntryCmd(cfg *config.Config, form AddFormData, apiToken string) tea.Cmd {
	return func() tea.Msg {
		var dnsRecordIDs []string
		var backupPath string

		// Parse multiple subdomains (supports newline-separated input)
		subdomains := ParseSubdomains(form.Subdomain)

		// Validate subdomains
		if err := ValidateSubdomains(subdomains); err != nil {
			return createEntryMsg{
				success:   false,
				err:       fmt.Errorf("invalid subdomains: %w", err),
				errorStep: "validate_subdomains",
			}
		}

		// Build FQDNs
		fqdns := BuildFQDNs(subdomains, cfg.Domain)

		// Parse port
		port := 80
		if form.ServicePort != "" {
			fmt.Sscanf(form.ServicePort, "%d", &port)
		}

		// Step 0: Check for duplicate domains in Caddyfile (only if not DNS-only)
		var err error
		if !form.DNSOnly {
			caddyContent, err := os.ReadFile(cfg.Caddy.CaddyfilePath)
			if err != nil {
				return createEntryMsg{
					success:   false,
					err:       fmt.Errorf("failed to read Caddyfile: %w", err),
					errorStep: "duplicate_check",
				}
			}

			parsed := caddy.ParseCaddyfileWithSnippets(string(caddyContent))
			for _, fqdn := range fqdns {
				fqdnLower := strings.ToLower(fqdn)
				for _, entry := range parsed.Entries {
					if strings.ToLower(entry.Domain) == fqdnLower {
						return createEntryMsg{
							success:   false,
							err:       fmt.Errorf("domain %s already exists in Caddyfile - please delete it first or use edit", fqdn),
							errorStep: "duplicate_domain",
						}
					}
					// Also check alternate domains in case of multi-domain blocks
					for _, altDomain := range entry.Domains {
						if strings.ToLower(altDomain) == fqdnLower {
							return createEntryMsg{
								success:   false,
								err:       fmt.Errorf("domain %s already exists in Caddyfile (in multi-domain entry) - please delete it first or use edit", fqdn),
								errorStep: "duplicate_domain",
							}
						}
					}
				}
			}
		}

		// Step 1: Backup Caddyfile (skip if DNS-only mode)
		if !form.DNSOnly {
			backupPath, err = caddy.BackupCaddyfile(cfg.Caddy.CaddyfilePath)
			if err != nil {
				return createEntryMsg{
					success:   false,
					err:       err,
					errorStep: "backup",
				}
			}
		}

		// Step 2: Create DNS records in Cloudflare (one per domain)
		cfClient := cloudflare.NewClient(apiToken)
		dnsRecordIDs = []string{}

		for _, fqdn := range fqdns {
			dnsRecord := cloudflare.DNSRecord{
				Type:    form.DNSType,
				Name:    fqdn,
				Content: form.DNSTarget,
				Proxied: form.Proxied,
				TTL:     1, // Auto
			}

			createdRecord, err := cfClient.CreateDNSRecord(cfg.Cloudflare.ZoneID, dnsRecord)
			if err != nil {
				// Rollback: Delete any DNS records already created
				for _, recordID := range dnsRecordIDs {
					cfClient.DeleteDNSRecord(cfg.Cloudflare.ZoneID, recordID)
				}
				return createEntryMsg{
					success:    false,
					err:        fmt.Errorf("failed to create DNS record for %s: %w", fqdn, err),
					errorStep:  "dns_create",
					backupPath: backupPath,
				}
			}
			dnsRecordIDs = append(dnsRecordIDs, createdRecord.ID)
		}

		// Step 3: Generate and append Caddy block (skip if DNS-only mode)
		if !form.DNSOnly {
			caddyBlock := caddy.GenerateCaddyBlock(caddy.GenerateBlockInput{
				Domains:           fqdns, // Use new Domains field for multi-domain support
				Target:            form.ReverseProxyTarget,
				Port:              port,
				SSL:               form.SSL,
				LANOnly:           form.LANOnly,
				OAuth:             form.OAuth,
				WebSocket:         form.WebSocket,
				LANSubnet:         cfg.Defaults.LANSubnet,
				AllowedExtIP:      cfg.Defaults.AllowedExternalIP,
				SelectedSnippets:  getSelectedSnippetNames(form.SelectedSnippets),
				CustomCaddyConfig: form.CustomCaddyConfig,
			})

			err = caddy.AppendEntry(cfg.Caddy.CaddyfilePath, caddyBlock)
			if err != nil {
				// Rollback: Delete all DNS records
				for _, recordID := range dnsRecordIDs {
					cfClient.DeleteDNSRecord(cfg.Cloudflare.ZoneID, recordID)
				}
				return createEntryMsg{
					success:    false,
					err:        err,
					errorStep:  "caddy_append",
					backupPath: backupPath,
				}
			}

			// Step 4: Validate Caddyfile
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				// Rollback: Restore Caddyfile, delete all DNS records
				for _, recordID := range dnsRecordIDs {
					cfClient.DeleteDNSRecord(cfg.Cloudflare.ZoneID, recordID)
				}
				return createEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddyfile validation"),
					errorStep:  "caddy_validate",
					backupPath: backupPath,
				}
			}

			// Step 5: Restart Caddy container
			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				// Rollback: Restore Caddyfile, delete all DNS records
				for _, recordID := range dnsRecordIDs {
					cfClient.DeleteDNSRecord(cfg.Cloudflare.ZoneID, recordID)
				}
				return createEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddy restart"),
					errorStep:  "caddy_restart",
					backupPath: backupPath,
				}
			}
		}

		// Success!
		return createEntryMsg{
			success:    true,
			backupPath: backupPath,
			// Note: dnsRecordID field no longer used (we now have multiple)
		}
	}
}

// updateEntryCmd updates an existing DNS record and Caddy entry with rollback on failure
func updateEntryCmd(cfg *config.Config, form AddFormData, oldEntry diff.SyncedEntry, apiToken string) tea.Cmd {
	return func() tea.Msg {
		var backupPath string

		// Parse subdomains (handles multi-line input)
		subdomains := ParseSubdomains(form.Subdomain)
		if len(subdomains) == 0 {
			return updateEntryMsg{
				success:   false,
				err:       fmt.Errorf("subdomain is required"),
				errorStep: "validate",
			}
		}

		// Build FQDNs - for update, use first subdomain for primary entry
		fqdns := BuildFQDNs(subdomains, cfg.Domain)
		fqdn := fqdns[0] // Primary FQDN for DNS record

		// Parse port
		port := 80
		if form.ServicePort != "" {
			fmt.Sscanf(form.ServicePort, "%d", &port)
		}

		// Step 1: Backup Caddyfile (if we're going to modify Caddy)
		var err error
		// Backup if: old entry had Caddy, OR we're adding Caddy (switching from DNS-only to full)
		if oldEntry.Caddy != nil || !form.DNSOnly {
			backupPath, err = caddy.BackupCaddyfile(cfg.Caddy.CaddyfilePath)
			if err != nil {
				return updateEntryMsg{
					success:   false,
					err:       err,
					errorStep: "backup",
				}
			}
		}

		// Step 2: Update DNS record in Cloudflare (only if DNS fields changed)
		cfClient := cloudflare.NewClient(apiToken)
		var oldDNSRecord cloudflare.DNSRecord
		dnsUpdated := false
		if oldEntry.DNS != nil {
			// Save old DNS values for rollback
			oldDNSRecord = *oldEntry.DNS

			// Check if DNS fields actually changed
			dnsChanged := oldEntry.DNS.Type != form.DNSType ||
				oldEntry.DNS.Content != form.DNSTarget ||
				oldEntry.DNS.Proxied != form.Proxied ||
				oldEntry.DNS.Name != fqdn

			if dnsChanged {
				// Prepare updated DNS record
				updatedDNS := cloudflare.DNSRecord{
					Type:    form.DNSType,
					Name:    fqdn,
					Content: form.DNSTarget,
					Proxied: form.Proxied,
					TTL:     1, // Auto
				}

				_, err = cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, updatedDNS)
				if err != nil {
					return updateEntryMsg{
						success:    false,
						err:        err,
						errorStep:  "dns_update",
						backupPath: backupPath,
					}
				}
				dnsUpdated = true
			}
		}

		// Step 3: Update Caddy configuration based on mode transitions
		// Case 1: Had Caddy before and switching to DNS-only -> Remove Caddy
		// Case 2: Had Caddy before and staying in full mode -> Update Caddy
		// Case 3: Didn't have Caddy and switching to full mode -> Add Caddy
		// Case 4: Didn't have Caddy and staying DNS-only -> Do nothing

		if oldEntry.Caddy != nil && form.DNSOnly {
			// Case 1: Switching to DNS-only mode - remove Caddy entry
			err = caddy.RemoveEntry(cfg.Caddy.CaddyfilePath, oldEntry.Domain)
			if err != nil {
				// Rollback DNS if we updated it
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        err,
					errorStep:  "caddy_remove",
					backupPath: backupPath,
				}
			}

			// Validate and restart after removal
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddyfile validation"),
					errorStep:  "caddy_validate",
					backupPath: backupPath,
				}
			}

			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddy restart"),
					errorStep:  "caddy_restart",
					backupPath: backupPath,
				}
			}
		} else if oldEntry.Caddy != nil && !form.DNSOnly {
			// Case 2: Had Caddy and staying in full mode - update Caddy entry
			// Remove old Caddy block
			err = caddy.RemoveEntry(cfg.Caddy.CaddyfilePath, oldEntry.Domain)
			if err != nil {
				// Rollback DNS if we updated it
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        err,
					errorStep:  "caddy_remove",
					backupPath: backupPath,
				}
			}

			// Generate and append new Caddy block
			caddyBlock := caddy.GenerateCaddyBlock(caddy.GenerateBlockInput{
				FQDN:              fqdn,
				Target:            form.ReverseProxyTarget,
				Port:              port,
				SSL:               form.SSL,
				LANOnly:           form.LANOnly,
				OAuth:             form.OAuth,
				WebSocket:         form.WebSocket,
				LANSubnet:         cfg.Defaults.LANSubnet,
				AllowedExtIP:      cfg.Defaults.AllowedExternalIP,
				SelectedSnippets:  getSelectedSnippetNames(form.SelectedSnippets),
				CustomCaddyConfig: form.CustomCaddyConfig,
			})

			err = caddy.AppendEntry(cfg.Caddy.CaddyfilePath, caddyBlock)
			if err != nil {
				// Rollback: Restore Caddyfile and DNS
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddyfile append"),
					errorStep:  "caddy_append",
					backupPath: backupPath,
				}
			}

			// Step 4: Validate Caddyfile
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				// Rollback: Restore Caddyfile and DNS
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddyfile validation"),
					errorStep:  "caddy_validate",
					backupPath: backupPath,
				}
			}

			// Step 5: Restart Caddy
			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				// Rollback: Restore Caddyfile and DNS
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddy restart"),
					errorStep:  "caddy_restart",
					backupPath: backupPath,
				}
			}
		} else if oldEntry.Caddy == nil && !form.DNSOnly {
			// Case 3: Didn't have Caddy, switching to full mode - add Caddy entry
			caddyBlock := caddy.GenerateCaddyBlock(caddy.GenerateBlockInput{
				FQDN:              fqdn,
				Target:            form.ReverseProxyTarget,
				Port:              port,
				SSL:               form.SSL,
				LANOnly:           form.LANOnly,
				OAuth:             form.OAuth,
				WebSocket:         form.WebSocket,
				LANSubnet:         cfg.Defaults.LANSubnet,
				AllowedExtIP:      cfg.Defaults.AllowedExternalIP,
				SelectedSnippets:  getSelectedSnippetNames(form.SelectedSnippets),
				CustomCaddyConfig: form.CustomCaddyConfig,
			})

			err = caddy.AppendEntry(cfg.Caddy.CaddyfilePath, caddyBlock)
			if err != nil {
				// Rollback: Restore Caddyfile and DNS
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddyfile append"),
					errorStep:  "caddy_append",
					backupPath: backupPath,
				}
			}

			// Validate Caddyfile
			err = formatAndValidateCaddyfile(cfg)
			if err != nil {
				// Rollback: Restore Caddyfile and DNS
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddyfile validation"),
					errorStep:  "caddy_validate",
					backupPath: backupPath,
				}
			}

			// Restart Caddy
			err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
			if err != nil {
				// Rollback: Restore Caddyfile and DNS
				if dnsUpdated {
					cfClient.UpdateDNSRecord(cfg.Cloudflare.ZoneID, oldEntry.DNS.ID, oldDNSRecord)
				}
				return updateEntryMsg{
					success:    false,
					err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddy restart"),
					errorStep:  "caddy_restart",
					backupPath: backupPath,
				}
			}
		}
		// Case 4: oldEntry.Caddy == nil && form.DNSOnly - do nothing with Caddy

		// Success!
		return updateEntryMsg{
			success:    true,
			backupPath: backupPath,
		}
	}
}

// syncEntryCmd syncs an orphaned entry by creating missing DNS or Caddy
func syncEntryCmd(cfg *config.Config, entry diff.SyncedEntry, apiToken string) tea.Cmd {
	return func() tea.Msg {
		// Determine what to create based on entry status
		if entry.Status == diff.StatusOrphanedDNS {
			// DNS exists but Caddy doesn't - create Caddy entry
			return syncToCaddyCmd(cfg, entry)()
		} else if entry.Status == diff.StatusOrphanedCaddy {
			// Caddy exists but DNS doesn't - create DNS record
			return syncToDNSCmd(cfg, entry, apiToken)()
		}
		return syncEntryMsg{
			success:   false,
			err:       fmt.Errorf("entry is not orphaned"),
			errorStep: "validation",
			domain:    entry.Domain,
			syncType:  "unknown",
		}
	}
}

// syncToCaddyCmd creates a Caddy entry for an orphaned DNS record
func syncToCaddyCmd(cfg *config.Config, entry diff.SyncedEntry) tea.Cmd {
	return func() tea.Msg {
		if entry.DNS == nil {
			return syncEntryMsg{
				success:   false,
				err:       fmt.Errorf("no DNS record found"),
				errorStep: "validation",
				domain:    entry.Domain,
				syncType:  "to_caddy",
			}
		}

		var backupPath string
		var err error

		// Step 1: Backup Caddyfile
		backupPath, err = caddy.BackupCaddyfile(cfg.Caddy.CaddyfilePath)
		if err != nil {
			return syncEntryMsg{
				success:   false,
				err:       err,
				errorStep: "backup",
				domain:    entry.Domain,
				syncType:  "to_caddy",
			}
		}

		// Step 2: Generate Caddy block using defaults from config
		caddyBlock := caddy.GenerateCaddyBlock(caddy.GenerateBlockInput{
			FQDN:              entry.Domain,
			Target:            "localhost", // Default target
			Port:              cfg.Defaults.Port,
			SSL:               cfg.Defaults.SSL,
			LANOnly:           false, // Default to not restricted
			OAuth:             false,
			WebSocket:         false,
			LANSubnet:         cfg.Defaults.LANSubnet,
			AllowedExtIP:      cfg.Defaults.AllowedExternalIP,
			AvailableSnippets: nil,
			CustomCaddyConfig: "", // No custom config for sync operation
		})

		// Step 3: Append to Caddyfile
		err = caddy.AppendEntry(cfg.Caddy.CaddyfilePath, caddyBlock)
		if err != nil {
			return syncEntryMsg{
				success:    false,
				err:        err,
				errorStep:  "caddy_append",
				backupPath: backupPath,
				domain:     entry.Domain,
				syncType:   "to_caddy",
			}
		}

		// Step 4: Validate Caddyfile
		err = formatAndValidateCaddyfile(cfg)
		if err != nil {
			// Rollback: Restore Caddyfile
			return syncEntryMsg{
				success:    false,
				err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddyfile validation"),
				errorStep:  "caddy_validate",
				backupPath: backupPath,
				domain:     entry.Domain,
				syncType:   "to_caddy",
			}
		}

		// Step 5: Restart Caddy
		err = caddy.RestartCaddy(cfg.Caddy.ContainerName)
		if err != nil {
			// Rollback: Restore Caddyfile
			return syncEntryMsg{
				success:    false,
				err:        restoreBackupWithError(cfg.Caddy.CaddyfilePath, backupPath, err, "Caddy restart"),
				errorStep:  "caddy_restart",
				backupPath: backupPath,
				domain:     entry.Domain,
				syncType:   "to_caddy",
			}
		}

		// Success!
		return syncEntryMsg{
			success:    true,
			backupPath: backupPath,
			domain:     entry.Domain,
			syncType:   "to_caddy",
		}
	}
}

// syncToDNSCmd creates a DNS record for an orphaned Caddy entry
func syncToDNSCmd(cfg *config.Config, entry diff.SyncedEntry, apiToken string) tea.Cmd {
	return func() tea.Msg {
		if entry.Caddy == nil {
			return syncEntryMsg{
				success:   false,
				err:       fmt.Errorf("no Caddy entry found"),
				errorStep: "validation",
				domain:    entry.Domain,
				syncType:  "to_dns",
			}
		}

		// Create DNS record using defaults from config
		cfClient := cloudflare.NewClient(apiToken)
		dnsRecord := cloudflare.DNSRecord{
			Type:    "CNAME",
			Name:    entry.Domain,
			Content: cfg.Defaults.CNAMETarget,
			Proxied: cfg.Defaults.Proxied,
			TTL:     1, // Auto
		}

		createdRecord, err := cfClient.CreateDNSRecord(cfg.Cloudflare.ZoneID, dnsRecord)
		if err != nil {
			return syncEntryMsg{
				success:   false,
				err:       err,
				errorStep: "dns_create",
				domain:    entry.Domain,
				syncType:  "to_dns",
			}
		}

		// Success!
		return syncEntryMsg{
			success:     true,
			dnsRecordID: createdRecord.ID,
			domain:      entry.Domain,
			syncType:    "to_dns",
		}
	}
}
