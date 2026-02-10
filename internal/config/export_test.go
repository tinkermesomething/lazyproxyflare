package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportImportProfile(t *testing.T) {
	// Create a temporary profiles directory
	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")
	os.MkdirAll(profilesDir, 0755)

	// Override home dir for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create the config dir structure
	configDir := filepath.Join(tmpDir, ".config", "lazyproxyflare", "profiles")
	os.MkdirAll(configDir, 0755)

	// Create a test profile
	profile := &ProfileConfig{
		Profile: ProfileMetadata{Name: "test-profile"},
		Domain:  "example.com",
		Cloudflare: CloudflareConfig{
			APIToken: "test-token",
			ZoneID:   "0123456789abcdef0123456789abcdef",
		},
		Proxy: ProxyConfig{
			Type:       ProxyTypeCaddy,
			Deployment: DeploymentDocker,
			Caddy: CaddyProxyConfig{
				CaddyfilePath: "/tmp/Caddyfile",
				ContainerName: "caddy",
			},
		},
	}
	err := SaveProfile("test-profile", profile)
	if err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}

	// Export it
	exportPath := filepath.Join(tmpDir, "export.tar.gz")
	err = ExportProfile("test-profile", exportPath)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Fatal("export file not created")
	}

	// Delete the profile
	DeleteProfile("test-profile")

	// Import it
	importedName, err := ImportProfile(exportPath, false)
	if err != nil {
		t.Fatalf("failed to import: %v", err)
	}
	if importedName != "test-profile" {
		t.Errorf("expected name 'test-profile', got '%s'", importedName)
	}

	// Verify imported profile loads
	imported, err := LoadProfile("test-profile")
	if err != nil {
		t.Fatalf("failed to load imported profile: %v", err)
	}
	if imported.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got '%s'", imported.Domain)
	}
}

func TestImportProfileNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tmpDir, ".config", "lazyproxyflare", "profiles")
	os.MkdirAll(configDir, 0755)

	// Create and export a profile
	profile := &ProfileConfig{
		Profile: ProfileMetadata{Name: "dupe"},
		Domain:  "example.com",
		Cloudflare: CloudflareConfig{
			APIToken: "test-token",
			ZoneID:   "0123456789abcdef0123456789abcdef",
		},
		Proxy: ProxyConfig{
			Type:       ProxyTypeCaddy,
			Deployment: DeploymentDocker,
			Caddy: CaddyProxyConfig{
				CaddyfilePath: "/tmp/Caddyfile",
				ContainerName: "caddy",
			},
		},
	}
	SaveProfile("dupe", profile)

	exportPath := filepath.Join(tmpDir, "dupe.tar.gz")
	ExportProfile("dupe", exportPath)

	// Try to import again without overwrite — should fail
	_, err := ImportProfile(exportPath, false)
	if err == nil {
		t.Fatal("expected error when importing duplicate without overwrite")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}

	// Import with overwrite — should succeed
	name, err := ImportProfile(exportPath, true)
	if err != nil {
		t.Fatalf("failed to import with overwrite: %v", err)
	}
	if name != "dupe" {
		t.Errorf("expected name 'dupe', got '%s'", name)
	}
}

func TestSanitizeProfileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal-name", "normal-name"},
		{"with spaces", "with_spaces"},
		{"with/slashes", "with_slashes"},
		{"", "imported"},
		{"a.b_c-d", "a.b_c-d"},
	}

	for _, tt := range tests {
		got := sanitizeProfileName(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeProfileName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGetDefaultExportDir(t *testing.T) {
	dir, err := GetDefaultExportDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(dir, "exports") {
		t.Errorf("expected path to contain 'exports', got %s", dir)
	}
}
