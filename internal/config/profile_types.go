package config

import "time"

// ProfileConfig represents a complete profile configuration
type ProfileConfig struct {
	Profile    ProfileMetadata  `yaml:"profile"`
	Cloudflare CloudflareConfig `yaml:"cloudflare"`
	Domain     string           `yaml:"domain"`
	Proxy      ProxyConfig      `yaml:"proxy"`
	Defaults   DefaultsConfig   `yaml:"defaults"`
	UI         UIConfig         `yaml:"ui"`
	Backup     BackupConfig     `yaml:"backup,omitempty"`
}

// BackupConfig holds backup rotation settings
type BackupConfig struct {
	MaxBackups int `yaml:"max_backups,omitempty"` // Maximum number of backups to keep (0 = unlimited)
	MaxSizeMB  int `yaml:"max_size_mb,omitempty"` // Maximum total backup size in MB (0 = unlimited)
}

// ProfileMetadata contains profile metadata
type ProfileMetadata struct {
	Name      string    `yaml:"name"`
	CreatedAt time.Time `yaml:"created_at"`
}

// ProxyType represents the type of reverse proxy
type ProxyType string

const (
	ProxyTypeCaddy ProxyType = "caddy"
)

// DeploymentMethod represents how the proxy is deployed
type DeploymentMethod string

const (
	DeploymentDocker DeploymentMethod = "docker"
	DeploymentSystem DeploymentMethod = "system" // systemd, snap, or binary
)

// ProxyConfig holds reverse proxy configuration
type ProxyConfig struct {
	Type       ProxyType        `yaml:"type"`
	Deployment DeploymentMethod `yaml:"deployment"`
	Caddy      CaddyProxyConfig `yaml:"caddy,omitempty"`
	// Future: Nginx, Traefik configs
}

// CaddyProxyConfig holds Caddy-specific configuration
type CaddyProxyConfig struct {
	CaddyfilePath          string `yaml:"caddyfile_path"`                     // Host path for reading/writing
	CaddyfileContainerPath string `yaml:"caddyfile_container_path,omitempty"` // Path inside Docker container (for validation)
	ContainerName          string `yaml:"container_name,omitempty"`           // Only for docker deployment
	DockerMethod           string `yaml:"docker_method,omitempty"`            // "compose" or "plain" for docker exec commands
	ComposeFilePath        string `yaml:"compose_file_path,omitempty"`        // Path to docker-compose.yml (for compose method)
	CaddyBinaryPath        string `yaml:"caddy_binary_path,omitempty"`        // Path to caddy binary (for system deployment)
	ValidationCommand      string `yaml:"validation_command,omitempty"`       // Optional custom command
	RestartCommand         string `yaml:"restart_command,omitempty"`          // Optional custom command
}

// GetValidationCommand returns the validation command with placeholders
func (c *CaddyProxyConfig) GetValidationCommand(deployment DeploymentMethod) string {
	if c.ValidationCommand != "" {
		return c.ValidationCommand
	}

	switch deployment {
	case DeploymentDocker:
		if c.DockerMethod == "plain" {
			return "docker exec {container} caddy validate --config {path}"
		}
		if c.ComposeFilePath != "" {
			return "docker compose -f {compose_file} exec -T {container} caddy validate --config {path}"
		}
		return "docker compose exec -T {container} caddy validate --config {path}"
	case DeploymentSystem:
		// Use detected caddy binary path or fall back to "caddy"
		if c.CaddyBinaryPath != "" {
			return c.CaddyBinaryPath + " validate --config {path}"
		}
		return "caddy validate --config {path}"
	default:
		return "caddy validate --config {path}"
	}
}

// GetReloadCommand returns the reload command with placeholders
func (c *CaddyProxyConfig) GetReloadCommand(deployment DeploymentMethod) string {
	if c.RestartCommand != "" {
		return c.RestartCommand
	}

	switch deployment {
	case DeploymentDocker:
		if c.DockerMethod == "plain" {
			return "docker exec {container} caddy reload --config {path}"
		}
		if c.ComposeFilePath != "" {
			return "docker compose -f {compose_file} exec -T {container} caddy reload --config {path}"
		}
		return "docker compose exec -T {container} caddy reload --config {path}"
	case DeploymentSystem:
		// Use detected caddy binary path or fall back to "caddy"
		if c.CaddyBinaryPath != "" {
			return c.CaddyBinaryPath + " reload --config {path}"
		}
		return "caddy reload --config {path}"
	default:
		return ""
	}
}

// GetRestartCommand returns the restart command with placeholders
// Deprecated: Use GetReloadCommand instead for graceful reloads
func (c *CaddyProxyConfig) GetRestartCommand(deployment DeploymentMethod) string {
	return c.GetReloadCommand(deployment)
}
