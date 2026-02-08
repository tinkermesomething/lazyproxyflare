package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// GetAPIToken retrieves the API token, expanding environment variables if necessary
func (c *Config) GetAPIToken() (string, error) {
	if c.Cloudflare.APIToken == "" {
		return "", fmt.Errorf("api_token is not set in the profile")
	}
	// expandEnvVars handles both plaintext tokens and ${VAR_NAME} style variables
	token := expandEnvVars(c.Cloudflare.APIToken)
	if token == "" {
		return "", fmt.Errorf("api_token is empty or the referenced environment variable is not set")
	}
	return token, nil
}

// ValidateStructure checks that required structural fields are present
func (c *Config) ValidateStructure() error {
	// Check required fields
	if c.Cloudflare.APIToken == "" {
		return fmt.Errorf("cloudflare.api_token is required")
	}
	if c.Cloudflare.ZoneID == "" {
		return fmt.Errorf("cloudflare.zone_id is required")
	}
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if c.Caddy.CaddyfilePath == "" {
		return fmt.Errorf("caddy.caddyfile_path is required")
	}
	// Either ContainerName (Docker) or CaddyBinaryPath (System) must be set
	if c.Caddy.ContainerName == "" && c.Caddy.CaddyBinaryPath == "" {
		return fmt.Errorf("either caddy.container_name (for Docker) or caddy.caddy_binary_path (for system) is required")
	}

	// Validate formats
	if !isValidZoneID(c.Cloudflare.ZoneID) {
		return fmt.Errorf("cloudflare.zone_id has invalid format (should be 32 hex characters)")
	}
	if !isValidDomain(c.Domain) {
		return fmt.Errorf("domain has invalid format (should be a valid FQDN)")
	}
	if c.Defaults.LANSubnet != "" && !isValidCIDR(c.Defaults.LANSubnet) {
		return fmt.Errorf("defaults.lan_subnet has invalid CIDR format")
	}
	if c.Defaults.AllowedExternalIP != "" && !isValidCIDR(c.Defaults.AllowedExternalIP) {
		return fmt.Errorf("defaults.allowed_external_ip has invalid CIDR format")
	}

	return nil
}

// isValidZoneID checks if the zone ID is a valid 32-character hex string
func isValidZoneID(zoneID string) bool {
	matched, _ := regexp.MatchString("^[a-f0-9]{32}$", zoneID)
	return matched
}

// isValidDomain checks if the domain is a valid FQDN
func isValidDomain(domain string) bool {
	// Basic domain validation: contains at least one dot and valid characters
	if !strings.Contains(domain, ".") {
		return false
	}
	// Check for valid domain characters
	matched, _ := regexp.MatchString("^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", domain)
	return matched
}

// IsValidCIDR checks if the string is a valid CIDR notation
// Exported for use in UI validation
func IsValidCIDR(cidr string) bool {
	// Empty string is valid (optional field)
	if cidr == "" {
		return true
	}

	// Basic CIDR validation: IP/prefix
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return false
	}
	// Check IP part (simple validation)
	ipParts := strings.Split(parts[0], ".")
	if len(ipParts) != 4 {
		return false
	}
	// Check prefix is numeric and in valid range (0-32)
	var prefix int
	if _, err := fmt.Sscanf(parts[1], "%d", &prefix); err != nil {
		return false
	}
	return prefix >= 0 && prefix <= 32
}

// isValidCIDR is internal wrapper for backward compatibility
func isValidCIDR(cidr string) bool {
	if cidr == "" {
		return true
	}
	return IsValidCIDR(cidr)
}

// expandEnvVars expands environment variables in format ${VAR_NAME}
// If the value is not in ${VAR_NAME} format, returns the value unchanged
// If the environment variable is not set, returns the original ${VAR_NAME} string
// (allowing validation to catch it with a helpful error message)
func expandEnvVars(value string) string {
	// Check if value is in ${VAR_NAME} format
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return value
	}

	// Extract variable name: "${VAR_NAME}" -> "VAR_NAME"
	varName := value[2 : len(value)-1]

	// Get from environment
	envValue := os.Getenv(varName)
	if envValue == "" {
		// Return original if env var not set (allows validation to catch it)
		return value
	}

	return envValue
}
