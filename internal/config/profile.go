package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// GetProfilesDir returns the directory where profiles are stored
func GetProfilesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	profilesDir := filepath.Join(homeDir, ".config", "lazyproxyflare", "profiles")
	return profilesDir, nil
}

// ListProfiles returns all available profile names
func ListProfiles() ([]string, error) {
	profilesDir, err := GetProfilesDir()
	if err != nil {
		return nil, err
	}

	// Check if profiles directory exists
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		return []string{}, nil // No profiles yet
	}

	// Read directory
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if it's a YAML file
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			// Remove extension to get profile name
			profileName := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
			profiles = append(profiles, profileName)
		}
	}

	return profiles, nil
}

// LoadProfile loads a profile configuration by name
func LoadProfile(name string) (*ProfileConfig, error) {
	profilesDir, err := GetProfilesDir()
	if err != nil {
		return nil, err
	}

	// Try .yaml first, then .yml
	profilePath := filepath.Join(profilesDir, name+".yaml")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		profilePath = filepath.Join(profilesDir, name+".yml")
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", name)
		}
	}

	// Read file
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	// Parse YAML
	var profile ProfileConfig
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile file: %w", err)
	}

	// Validate profile
	if err := validateProfile(&profile); err != nil {
		return nil, fmt.Errorf("invalid profile configuration: %w", err)
	}

	return &profile, nil
}

// SaveProfile saves a profile configuration
func SaveProfile(name string, profile *ProfileConfig) error {
	profilesDir, err := GetProfilesDir()
	if err != nil {
		return err
	}

	// Ensure profiles directory exists
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Set metadata
	if profile.Profile.Name == "" {
		profile.Profile.Name = name
	}
	if profile.Profile.CreatedAt.IsZero() {
		profile.Profile.CreatedAt = time.Now()
	}

	// Validate profile configuration before saving
	if err := validateProfile(profile); err != nil {
		return fmt.Errorf("invalid profile configuration: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Write to file
	profilePath := filepath.Join(profilesDir, name+".yaml")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// DeleteProfile removes a profile file
func DeleteProfile(name string) error {
	profilesDir, err := GetProfilesDir()
	if err != nil {
		return err
	}

	profilePath := filepath.Join(profilesDir, name+".yaml")
	if err := os.Remove(profilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// GetLastUsedProfile returns the name of the last used profile
func GetLastUsedProfile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	lastProfilePath := filepath.Join(homeDir, ".config", "lazyproxyflare", "last_profile.txt")

	// Check if file exists
	if _, err := os.Stat(lastProfilePath); os.IsNotExist(err) {
		return "", nil // No last profile
	}

	// Read file
	data, err := os.ReadFile(lastProfilePath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// SetLastUsedProfile saves the name of the last used profile
func SetLastUsedProfile(name string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "lazyproxyflare")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	lastProfilePath := filepath.Join(configDir, "last_profile.txt")
	return os.WriteFile(lastProfilePath, []byte(name), 0644)
}

// validateProfile validates a profile configuration
func validateProfile(p *ProfileConfig) error {
	// Check required fields
	if p.Cloudflare.APIToken == "" {
		return fmt.Errorf("cloudflare.api_token is required")
	}
	if p.Cloudflare.ZoneID == "" {
		return fmt.Errorf("cloudflare.zone_id is required")
	}
	if p.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if p.Proxy.Type == "" {
		return fmt.Errorf("proxy.type is required")
	}
	if p.Proxy.Deployment == "" {
		return fmt.Errorf("proxy.deployment is required")
	}

	// Validate proxy-specific config
	if p.Proxy.Type == ProxyTypeCaddy {
		if p.Proxy.Caddy.CaddyfilePath == "" {
			return fmt.Errorf("proxy.caddy.caddyfile_path is required for Caddy")
		}
		if p.Proxy.Deployment == DeploymentDocker && p.Proxy.Caddy.ContainerName == "" {
			return fmt.Errorf("proxy.caddy.container_name is required for Docker deployment")
		}
	}

	// Validate formats (reuse existing validators)
	if !isValidZoneID(p.Cloudflare.ZoneID) {
		return fmt.Errorf("cloudflare.zone_id has invalid format (should be 32 hex characters)")
	}
	if !isValidDomain(p.Domain) {
		return fmt.Errorf("domain has invalid format (should be a valid FQDN)")
	}
	if p.Defaults.LANSubnet != "" && !isValidCIDR(p.Defaults.LANSubnet) {
		return fmt.Errorf("defaults.lan_subnet has invalid CIDR format")
	}
	if p.Defaults.AllowedExternalIP != "" && !isValidCIDR(p.Defaults.AllowedExternalIP) {
		return fmt.Errorf("defaults.allowed_external_ip has invalid CIDR format")
	}

	return nil
}

// ProfileToLegacyConfig converts a ProfileConfig to the old Config format
// This is used to maintain compatibility with the rest of the codebase
func ProfileToLegacyConfig(profile *ProfileConfig) *Config {
	return &Config{
		Cloudflare: profile.Cloudflare,
		Domain:     profile.Domain,
		Caddy: CaddyConfig{
			CaddyfilePath:          profile.Proxy.Caddy.CaddyfilePath,
			CaddyfileContainerPath: profile.Proxy.Caddy.CaddyfileContainerPath,
			ContainerName:          profile.Proxy.Caddy.ContainerName,
			DockerMethod:           profile.Proxy.Caddy.DockerMethod,
			ComposeFilePath:        profile.Proxy.Caddy.ComposeFilePath,
			CaddyBinaryPath:        profile.Proxy.Caddy.CaddyBinaryPath,
			ValidationCommand:      profile.Proxy.Caddy.GetValidationCommand(profile.Proxy.Deployment),
		},
		Defaults: profile.Defaults,
		UI:       profile.UI,
	}
}
