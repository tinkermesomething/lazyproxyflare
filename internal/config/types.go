package config

// Config represents the application configuration
type Config struct {
	Cloudflare CloudflareConfig `yaml:"cloudflare"`
	Domain     string           `yaml:"domain"`
	Caddy      CaddyConfig      `yaml:"caddy"`
	Defaults   DefaultsConfig   `yaml:"defaults"`
	UI         UIConfig         `yaml:"ui"`
	Backup     BackupConfig     `yaml:"backup,omitempty"`
}

// CloudflareConfig holds Cloudflare API credentials
type CloudflareConfig struct {
	// API Token for Cloudflare. Can be a plaintext token or an environment variable
	// reference in the format ${VAR_NAME}.
	APIToken string `yaml:"api_token"`

	ZoneID string `yaml:"zone_id"`
}

// CaddyConfig holds Caddy-related configuration
type CaddyConfig struct {
	CaddyfilePath          string `yaml:"caddyfile_path"`                     // Host path for reading/writing
	CaddyfileContainerPath string `yaml:"caddyfile_container_path,omitempty"` // Path inside Docker container (for validation)
	ContainerName          string `yaml:"container_name,omitempty"`           // Docker container name (docker deployment only)
	DockerMethod           string `yaml:"docker_method,omitempty"`            // "compose" or "plain"
	ComposeFilePath        string `yaml:"compose_file_path,omitempty"`        // Path to docker-compose.yml
	CaddyBinaryPath        string `yaml:"caddy_binary_path,omitempty"`        // Path to caddy binary (system deployment only)
	ValidationCommand      string `yaml:"validation_command,omitempty"`       // Command with placeholders
}

// DefaultsConfig holds default values for new entries
type DefaultsConfig struct {
	CNAMETarget       string `yaml:"cname_target"`
	Proxied           bool   `yaml:"proxied"`
	Port              int    `yaml:"port"`
	SSL               bool   `yaml:"ssl"`
	LANSubnet         string `yaml:"lan_subnet"`
	AllowedExternalIP string `yaml:"allowed_external_ip"`
}

// UIConfig holds UI preferences
type UIConfig struct {
	Theme  string `yaml:"theme"`
	Editor string `yaml:"editor,omitempty"`
}
