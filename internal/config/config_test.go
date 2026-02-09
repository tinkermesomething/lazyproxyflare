package config

import (
	"os"
	"testing"
)

func TestIsValidZoneID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid 32 hex chars", "abcdef0123456789abcdef0123456789", true},
		{"valid all zeros", "00000000000000000000000000000000", true},
		{"too short", "abcdef01234567890123456789", false},
		{"too long", "abcdef0123456789abcdef01234567890", false},
		{"uppercase rejected", "ABCDEF0123456789ABCDEF0123456789", false},
		{"non-hex chars", "ghijkl0123456789abcdef0123456789", false},
		{"empty", "", false},
		{"with spaces", "abcdef0123456789 bcdef0123456789", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidZoneID(tt.input); got != tt.want {
				t.Errorf("isValidZoneID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple domain", "example.com", true},
		{"subdomain", "sub.example.com", true},
		{"with hyphens", "my-app.example.com", true},
		{"single label (no dot)", "localhost", false},
		{"empty", "", false},
		{"leading hyphen", "-example.com", false},
		{"trailing hyphen in label", "example-.com", false},
		{"numeric domain", "123.456", true},
		{"with underscore", "test_app.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidDomain(tt.input); got != tt.want {
				t.Errorf("isValidDomain(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidCIDR(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid /24", "192.168.1.0/24", true},
		{"valid /8", "10.0.0.0/8", true},
		{"valid /32", "127.0.0.1/32", true},
		{"valid /0", "0.0.0.0/0", true},
		{"empty is valid (optional)", "", true},
		{"missing prefix", "10.0.0.0", false},
		{"prefix too large", "10.0.0.0/33", false},
		{"bad IP (3 octets)", "10.0.0/24", false},
		{"not a CIDR", "not-a-cidr", false},
		{"negative prefix", "10.0.0.0/-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCIDR(tt.input); got != tt.want {
				t.Errorf("IsValidCIDR(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandEnvVars(t *testing.T) {
	t.Run("plaintext passthrough", func(t *testing.T) {
		if got := expandEnvVars("my-plain-token"); got != "my-plain-token" {
			t.Errorf("got %q, want plaintext passthrough", got)
		}
	})

	t.Run("env var expansion", func(t *testing.T) {
		os.Setenv("TEST_CF_TOKEN", "secret123")
		defer os.Unsetenv("TEST_CF_TOKEN")

		if got := expandEnvVars("${TEST_CF_TOKEN}"); got != "secret123" {
			t.Errorf("got %q, want %q", got, "secret123")
		}
	})

	t.Run("missing env var returns original", func(t *testing.T) {
		os.Unsetenv("NONEXISTENT_VAR_12345")
		if got := expandEnvVars("${NONEXISTENT_VAR_12345}"); got != "${NONEXISTENT_VAR_12345}" {
			t.Errorf("got %q, want original string", got)
		}
	})

	t.Run("partial format not expanded", func(t *testing.T) {
		if got := expandEnvVars("${INCOMPLETE"); got != "${INCOMPLETE" {
			t.Errorf("got %q, want original string", got)
		}
	})
}

func TestGetAPIToken(t *testing.T) {
	t.Run("empty token returns error", func(t *testing.T) {
		cfg := &Config{}
		_, err := cfg.GetAPIToken()
		if err == nil {
			t.Error("expected error for empty token")
		}
	})

	t.Run("plaintext token returned", func(t *testing.T) {
		cfg := &Config{Cloudflare: CloudflareConfig{APIToken: "plain-token"}}
		got, err := cfg.GetAPIToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "plain-token" {
			t.Errorf("got %q, want %q", got, "plain-token")
		}
	})

	t.Run("env var token expanded", func(t *testing.T) {
		os.Setenv("TEST_API_TOKEN", "env-secret")
		defer os.Unsetenv("TEST_API_TOKEN")

		cfg := &Config{Cloudflare: CloudflareConfig{APIToken: "${TEST_API_TOKEN}"}}
		got, err := cfg.GetAPIToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "env-secret" {
			t.Errorf("got %q, want %q", got, "env-secret")
		}
	})
}

func TestValidateStructure(t *testing.T) {
	validConfig := Config{
		Cloudflare: CloudflareConfig{
			APIToken: "test-token",
			ZoneID:   "abcdef0123456789abcdef0123456789",
		},
		Domain: "example.com",
		Caddy: CaddyConfig{
			CaddyfilePath: "/etc/caddy/Caddyfile",
			ContainerName: "caddy",
		},
	}

	t.Run("valid config passes", func(t *testing.T) {
		if err := validConfig.ValidateStructure(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing api_token", func(t *testing.T) {
		cfg := validConfig
		cfg.Cloudflare.APIToken = ""
		if err := cfg.ValidateStructure(); err == nil {
			t.Error("expected error for missing api_token")
		}
	})

	t.Run("missing zone_id", func(t *testing.T) {
		cfg := validConfig
		cfg.Cloudflare.ZoneID = ""
		if err := cfg.ValidateStructure(); err == nil {
			t.Error("expected error for missing zone_id")
		}
	})

	t.Run("invalid zone_id format", func(t *testing.T) {
		cfg := validConfig
		cfg.Cloudflare.ZoneID = "not-a-valid-zone-id"
		if err := cfg.ValidateStructure(); err == nil {
			t.Error("expected error for invalid zone_id format")
		}
	})

	t.Run("missing domain", func(t *testing.T) {
		cfg := validConfig
		cfg.Domain = ""
		if err := cfg.ValidateStructure(); err == nil {
			t.Error("expected error for missing domain")
		}
	})

	t.Run("invalid domain format", func(t *testing.T) {
		cfg := validConfig
		cfg.Domain = "not valid"
		if err := cfg.ValidateStructure(); err == nil {
			t.Error("expected error for invalid domain")
		}
	})

	t.Run("missing caddy path", func(t *testing.T) {
		cfg := validConfig
		cfg.Caddy.CaddyfilePath = ""
		if err := cfg.ValidateStructure(); err == nil {
			t.Error("expected error for missing caddyfile_path")
		}
	})

	t.Run("no container or binary path", func(t *testing.T) {
		cfg := validConfig
		cfg.Caddy.ContainerName = ""
		cfg.Caddy.CaddyBinaryPath = ""
		if err := cfg.ValidateStructure(); err == nil {
			t.Error("expected error when neither container_name nor caddy_binary_path set")
		}
	})

	t.Run("binary path instead of container", func(t *testing.T) {
		cfg := validConfig
		cfg.Caddy.ContainerName = ""
		cfg.Caddy.CaddyBinaryPath = "/usr/bin/caddy"
		if err := cfg.ValidateStructure(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid lan_subnet CIDR", func(t *testing.T) {
		cfg := validConfig
		cfg.Defaults.LANSubnet = "not-cidr"
		if err := cfg.ValidateStructure(); err == nil {
			t.Error("expected error for invalid lan_subnet")
		}
	})
}
