package ui

import (
	"os"
	"path/filepath"
	"testing"

	"lazyproxyflare/internal/config"
)

func TestValidateProfile(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()

	// Create a test Caddyfile
	testCaddyfile := filepath.Join(tmpDir, "Caddyfile")
	if err := os.WriteFile(testCaddyfile, []byte("# test caddyfile"), 0644); err != nil {
		t.Fatalf("Failed to create test Caddyfile: %v", err)
	}

	// Test cases
	tests := []struct {
		name        string
		wizardData  WizardData
		shouldError bool
		errorContains string
	}{
		{
			name: "Valid profile with existing Caddyfile",
			wizardData: WizardData{
				ProfileName:      "test-profile-unique",
				Domain:           "test-unique.example.com",
				ProxyType:        config.ProxyTypeCaddy,
				DeploymentMethod: config.DeploymentDocker,
				CaddyfilePath:    testCaddyfile,
				ContainerName:    "caddy",
			},
			shouldError: false,
		},
		{
			name: "Valid profile with non-existent Caddyfile but parent directory exists",
			wizardData: WizardData{
				ProfileName:      "test-profile-new-file",
				Domain:           "test-newfile.example.com",
				ProxyType:        config.ProxyTypeCaddy,
				DeploymentMethod: config.DeploymentDocker,
				CaddyfilePath:    filepath.Join(tmpDir, "NewCaddyfile"),
				ContainerName:    "caddy",
			},
			shouldError: false,
		},
		{
			name: "Empty Caddyfile path",
			wizardData: WizardData{
				ProfileName:      "test-empty-path",
				Domain:           "test-empty.example.com",
				ProxyType:        config.ProxyTypeCaddy,
				DeploymentMethod: config.DeploymentDocker,
				CaddyfilePath:    "",
				ContainerName:    "caddy",
			},
			shouldError: true,
			errorContains: "cannot be empty",
		},
		{
			name: "Caddyfile parent directory does not exist",
			wizardData: WizardData{
				ProfileName:      "test-bad-parent",
				Domain:           "test-badparent.example.com",
				ProxyType:        config.ProxyTypeCaddy,
				DeploymentMethod: config.DeploymentDocker,
				CaddyfilePath:    "/nonexistent/directory/Caddyfile",
				ContainerName:    "caddy",
			},
			shouldError: true,
			errorContains: "parent directory does not exist",
		},
		{
			name: "Empty container name for Docker deployment",
			wizardData: WizardData{
				ProfileName:      "test-empty-container",
				Domain:           "test-emptycontainer.example.com",
				ProxyType:        config.ProxyTypeCaddy,
				DeploymentMethod: config.DeploymentDocker,
				CaddyfilePath:    testCaddyfile,
				ContainerName:    "",
			},
			shouldError: true,
			errorContains: "container name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfile(&tt.wizardData)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
