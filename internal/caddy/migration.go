package caddy

import (
	"fmt"
	"os"
	"time"
)

// MigrationOptions specifies what to import during migration
type MigrationOptions struct {
	ImportEntries  bool // Import domain entries
	ImportSnippets bool // Import snippets
	ArchiveOld     bool // Archive existing Caddyfile
	FreshTemplate  bool // Start with fresh template (vs preserve global options)
}

// ArchiveCaddyfile creates a timestamped backup of the Caddyfile
// Returns the path to the archived file
func ArchiveCaddyfile(caddyfilePath string) (string, error) {
	// Get original file permissions
	fileInfo, err := os.Stat(caddyfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat Caddyfile: %w", err)
	}
	originalPerms := fileInfo.Mode().Perm()

	// Read current file
	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read Caddyfile for archiving: %w", err)
	}

	// Generate timestamped backup filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := fmt.Sprintf("%s.backup.%s", caddyfilePath, timestamp)

	// Write backup with original permissions
	if err := os.WriteFile(backupPath, content, originalPerms); err != nil {
		return "", fmt.Errorf("failed to write Caddyfile archive: %w", err)
	}

	return backupPath, nil
}

// GenerateFreshCaddyfile creates a new managed Caddyfile with imported content
// Returns the new Caddyfile content as a string
func GenerateFreshCaddyfile(migrationContent *MigrationContent, options MigrationOptions) (string, error) {
	var result string

	// 1. Add global options block
	if options.FreshTemplate {
		// Use default fresh template global options
		result = "{\n" + GetDefaultGlobalOptions() + "\n}\n\n"
	} else if migrationContent.HasGlobalOptions() {
		// Preserve existing global options
		result = "{\n" + migrationContent.GlobalOptions + "\n}\n\n"
	} else {
		// No global options, use defaults
		result = "{\n" + GetDefaultGlobalOptions() + "\n}\n\n"
	}

	// 2. Add snippets section header
	result += "#──────────────────────────────────────────────────────────────\n"
	result += "# SNIPPETS (Managed by LazyProxyFlare)\n"
	result += "#──────────────────────────────────────────────────────────────\n\n"

	// 3. Add snippets if importing
	if options.ImportSnippets && len(migrationContent.Snippets) > 0 {
		for _, snippet := range migrationContent.Snippets {
			result += fmt.Sprintf("(%s) {\n", snippet.Name)
			result += snippet.Content + "\n"
			result += "}\n\n"
		}
	} else {
		result += "# Snippets will be added here by the application\n\n"
	}

	// 4. Add entries section header
	result += "#──────────────────────────────────────────────────────────────\n"
	result += "# TUNNEL ENTRIES (Managed by LazyProxyFlare)\n"
	result += "#──────────────────────────────────────────────────────────────\n\n"

	// 5. Add entries if importing
	if options.ImportEntries && len(migrationContent.Entries) > 0 {
		for _, entry := range migrationContent.Entries {
			// Add entry marker comment
			result += fmt.Sprintf("# === %s ===\n", entry.Domain)

			// Add entry block (use raw block to preserve exact formatting)
			result += entry.RawBlock + "\n\n"
		}
	} else {
		result += "# Domain entries will be added here by the application\n"
	}

	return result, nil
}

// MigrateCaddyfile performs the full migration workflow
// Returns the backup path and any error
func MigrateCaddyfile(caddyfilePath string, options MigrationOptions) (backupPath string, err error) {
	// 1. Parse existing Caddyfile
	migrationContent, err := ParseCaddyfileForMigration(caddyfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse Caddyfile: %w", err)
	}

	// 2. Archive existing Caddyfile if requested
	if options.ArchiveOld {
		backupPath, err = ArchiveCaddyfile(caddyfilePath)
		if err != nil {
			return "", fmt.Errorf("failed to archive Caddyfile: %w", err)
		}
	}

	// 3. Generate fresh managed Caddyfile
	newContent, err := GenerateFreshCaddyfile(migrationContent, options)
	if err != nil {
		// If we created a backup, try to restore it
		if backupPath != "" {
			_ = RestoreFromBackup(caddyfilePath, backupPath)
		}
		return backupPath, fmt.Errorf("failed to generate new Caddyfile: %w", err)
	}

	// 4. Write new Caddyfile with preserved permissions
	var filePerms os.FileMode = 0644 // default permissions
	if backupPath != "" {
		// Get permissions from backup if it exists
		if backupInfo, statErr := os.Stat(backupPath); statErr == nil {
			filePerms = backupInfo.Mode().Perm()
		}
	}

	if err := os.WriteFile(caddyfilePath, []byte(newContent), filePerms); err != nil {
		// Try to restore backup on write failure
		if backupPath != "" {
			_ = RestoreFromBackup(caddyfilePath, backupPath)
		}
		return backupPath, fmt.Errorf("failed to write new Caddyfile: %w", err)
	}

	return backupPath, nil
}

// Note: RestoreFromBackup, ListBackups, and DeleteBackup are already
// implemented in manager.go and are reused here

// GetMigrationSummary returns a human-readable summary of what will be migrated
func GetMigrationSummary(migrationContent *MigrationContent, options MigrationOptions) string {
	var summary string

	entriesCount, snippetsCount := migrationContent.CountContent()

	summary += "Migration Summary:\n"
	summary += fmt.Sprintf("- Entries: %d (%s)\n", entriesCount,
		formatImportAction(options.ImportEntries))
	summary += fmt.Sprintf("- Snippets: %d (%s)\n", snippetsCount,
		formatImportAction(options.ImportSnippets))
	summary += fmt.Sprintf("- Global Options: %s\n",
		formatGlobalAction(migrationContent.HasGlobalOptions(), options.FreshTemplate))
	summary += fmt.Sprintf("- Archive Original: %s\n",
		formatBoolAction(options.ArchiveOld))

	return summary
}

func formatImportAction(willImport bool) string {
	if willImport {
		return "will import"
	}
	return "will discard"
}

func formatGlobalAction(hasGlobal, useFresh bool) string {
	if useFresh {
		return "use fresh defaults"
	}
	if hasGlobal {
		return "preserve existing"
	}
	return "use defaults (none found)"
}

func formatBoolAction(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
