package caddy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveCaddyfile(t *testing.T) {
	// Create temp Caddyfile
	tempDir := t.TempDir()
	caddyfilePath := filepath.Join(tempDir, "Caddyfile")
	testContent := "test content\n"

	if err := os.WriteFile(caddyfilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test Caddyfile: %v", err)
	}

	// Archive it
	backupPath, err := ArchiveCaddyfile(caddyfilePath)
	if err != nil {
		t.Fatalf("ArchiveCaddyfile failed: %v", err)
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file not created: %s", backupPath)
	}

	// Verify backup content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	if string(backupContent) != testContent {
		t.Errorf("Backup content mismatch.\nExpected: %q\nGot: %q", testContent, string(backupContent))
	}

	// Verify backup filename format
	if !strings.Contains(backupPath, "Caddyfile.backup.") {
		t.Errorf("Backup path doesn't contain expected format: %s", backupPath)
	}

	fmt.Printf("✓ Archive created: %s\n", backupPath)
}

func TestGenerateFreshCaddyfile_ImportAll(t *testing.T) {
	// Create test migration content
	migrationContent := &MigrationContent{
		GlobalOptions: "\tadmin off\n\tauto_https off",
		Snippets: []Snippet{
			{
				Name:    "test_snippet",
				Content: "\theader X-Test \"value\"",
			},
		},
		Entries: []CaddyEntry{
			{
				Domain:   "test.example.com",
				RawBlock: "test.example.com {\n\treverse_proxy localhost:8080\n}",
			},
		},
	}

	options := MigrationOptions{
		ImportEntries:  true,
		ImportSnippets: true,
		ArchiveOld:     true,
		FreshTemplate:  false, // Preserve existing global options
	}

	result, err := GenerateFreshCaddyfile(migrationContent, options)
	if err != nil {
		t.Fatalf("GenerateFreshCaddyfile failed: %v", err)
	}

	// Verify structure
	requiredSections := []string{
		"admin off",
		"auto_https off",
		"SNIPPETS (Managed by LazyProxyFlare)",
		"(test_snippet)",
		"TUNNEL ENTRIES (Managed by LazyProxyFlare)",
		"test.example.com",
		"reverse_proxy localhost:8080",
	}

	for _, section := range requiredSections {
		if !strings.Contains(result, section) {
			t.Errorf("Generated Caddyfile missing section: %s", section)
		}
	}

	fmt.Printf("✓ Generated Caddyfile (import all):\n%s\n", result)
}

func TestGenerateFreshCaddyfile_FreshTemplate(t *testing.T) {
	migrationContent := &MigrationContent{
		GlobalOptions: "\tadmin 0.0.0.0:2019\n\temail old@example.com",
		Snippets:      []Snippet{},
		Entries:       []CaddyEntry{},
	}

	options := MigrationOptions{
		ImportEntries:  false,
		ImportSnippets: false,
		ArchiveOld:     false,
		FreshTemplate:  true, // Use fresh defaults, ignore existing
	}

	result, err := GenerateFreshCaddyfile(migrationContent, options)
	if err != nil {
		t.Fatalf("GenerateFreshCaddyfile failed: %v", err)
	}

	// Should use default options, not old ones
	if strings.Contains(result, "admin 0.0.0.0:2019") {
		t.Error("Should not contain old global options when FreshTemplate=true")
	}

	if !strings.Contains(result, "admin off") {
		t.Error("Should contain default global options")
	}

	// Should have placeholder text
	if !strings.Contains(result, "Snippets will be added here") {
		t.Error("Should contain snippet placeholder")
	}

	if !strings.Contains(result, "Domain entries will be added here") {
		t.Error("Should contain entries placeholder")
	}

	fmt.Printf("✓ Generated fresh Caddyfile:\n%s\n", result)
}

func TestMigrateCaddyfile(t *testing.T) {
	// Create temp directory with test Caddyfile
	tempDir := t.TempDir()
	caddyfilePath := filepath.Join(tempDir, "Caddyfile")

	testCaddyfile := `{
	admin off
}

(security_headers) {
	header X-Frame-Options "DENY"
}

test.example.com {
	import security_headers
	reverse_proxy localhost:8080
}
`

	if err := os.WriteFile(caddyfilePath, []byte(testCaddyfile), 0644); err != nil {
		t.Fatalf("Failed to create test Caddyfile: %v", err)
	}

	// Perform migration (import all, archive)
	options := MigrationOptions{
		ImportEntries:  true,
		ImportSnippets: true,
		ArchiveOld:     true,
		FreshTemplate:  false,
	}

	backupPath, err := MigrateCaddyfile(caddyfilePath, options)
	if err != nil {
		t.Fatalf("MigrateCaddyfile failed: %v", err)
	}

	// Verify backup was created
	if backupPath == "" {
		t.Error("Expected backup path to be returned")
	}

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file not created: %s", backupPath)
	}

	// Verify new Caddyfile was written
	newContent, err := os.ReadFile(caddyfilePath)
	if err != nil {
		t.Fatalf("Failed to read migrated Caddyfile: %v", err)
	}

	newContentStr := string(newContent)

	// Verify structure
	if !strings.Contains(newContentStr, "Managed by LazyProxyFlare") {
		t.Error("New Caddyfile should have management headers")
	}

	if !strings.Contains(newContentStr, "security_headers") {
		t.Error("New Caddyfile should contain imported snippet")
	}

	if !strings.Contains(newContentStr, "test.example.com") {
		t.Error("New Caddyfile should contain imported entry")
	}

	fmt.Printf("✓ Migration successful\n")
	fmt.Printf("  Backup: %s\n", backupPath)
	fmt.Printf("  New Caddyfile:\n%s\n", newContentStr)
}

func TestRestoreFromBackup(t *testing.T) {
	tempDir := t.TempDir()
	caddyfilePath := filepath.Join(tempDir, "Caddyfile")
	backupPath := filepath.Join(tempDir, "Caddyfile.backup.test")

	originalContent := "original content\n"
	newContent := "modified content\n"

	// Create original and backup
	if err := os.WriteFile(backupPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	if err := os.WriteFile(caddyfilePath, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to create Caddyfile: %v", err)
	}

	// Restore from backup
	if err := RestoreFromBackup(caddyfilePath, backupPath); err != nil {
		t.Fatalf("RestoreFromBackup failed: %v", err)
	}

	// Verify restoration
	restored, err := os.ReadFile(caddyfilePath)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(restored) != originalContent {
		t.Errorf("Restore failed.\nExpected: %q\nGot: %q", originalContent, string(restored))
	}

	fmt.Printf("✓ Restore from backup successful\n")
}

func TestListBackups(t *testing.T) {
	tempDir := t.TempDir()
	caddyfilePath := filepath.Join(tempDir, "Caddyfile")

	// Create main file
	if err := os.WriteFile(caddyfilePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create Caddyfile: %v", err)
	}

	// Create multiple backups
	backup1 := filepath.Join(tempDir, "Caddyfile.backup.2024-01-01_10-00-00")
	backup2 := filepath.Join(tempDir, "Caddyfile.backup.2024-01-02_10-00-00")
	nonBackup := filepath.Join(tempDir, "other.txt")

	for _, path := range []string{backup1, backup2, nonBackup} {
		if err := os.WriteFile(path, []byte("backup"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// List backups
	backups, err := ListBackups(caddyfilePath)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	// Verify count
	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}

	// Verify backups are in list
	foundBackup1 := false
	foundBackup2 := false

	for _, backup := range backups {
		if backup.Path == backup1 {
			foundBackup1 = true
		}
		if backup.Path == backup2 {
			foundBackup2 = true
		}
	}

	if !foundBackup1 || !foundBackup2 {
		t.Error("Not all backups found in list")
	}

	fmt.Printf("✓ Found %d backups\n", len(backups))
	for _, b := range backups {
		fmt.Printf("  - %s\n", filepath.Base(b.Path))
	}
}

func TestGetMigrationSummary(t *testing.T) {
	migrationContent := &MigrationContent{
		GlobalOptions: "\tadmin off",
		Snippets: []Snippet{
			{Name: "snippet1"},
			{Name: "snippet2"},
		},
		Entries: []CaddyEntry{
			{Domain: "test1.com"},
			{Domain: "test2.com"},
			{Domain: "test3.com"},
		},
	}

	options := MigrationOptions{
		ImportEntries:  true,
		ImportSnippets: false,
		ArchiveOld:     true,
		FreshTemplate:  false,
	}

	summary := GetMigrationSummary(migrationContent, options)

	// Verify summary contains key info
	if !strings.Contains(summary, "3") {
		t.Error("Summary should mention 3 entries")
	}

	if !strings.Contains(summary, "2") {
		t.Error("Summary should mention 2 snippets")
	}

	if !strings.Contains(summary, "will import") {
		t.Error("Summary should mention importing entries")
	}

	if !strings.Contains(summary, "will discard") {
		t.Error("Summary should mention discarding snippets")
	}

	fmt.Printf("✓ Migration Summary:\n%s\n", summary)
}
