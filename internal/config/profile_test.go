package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateProfileName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"plain", "home", false},
		{"with dash and dot", "my-profile.v2", false},
		{"with spaces", "my profile", false},
		{"empty", "", true},
		{"parent traversal", "../evil", true},
		{"nested traversal", "a/../../evil", true},
		{"forward slash", "a/b", true},
		{"backslash", `a\b`, true},
		{"only dots", "..", true},
		{"leading dot", ".hidden", true},
		{"nul byte", "a\x00b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfileName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProfileName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// Path-unsafe names must be rejected before any file is touched.
func TestProfileFuncsRejectUnsafeNames(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	if _, err := LoadProfile("../evil"); err == nil {
		t.Error("LoadProfile accepted a traversal name")
	}
	if err := DeleteProfile("../evil"); err == nil {
		t.Error("DeleteProfile accepted a traversal name")
	}
	if err := SaveProfile("../evil", &ProfileConfig{}); err == nil {
		t.Error("SaveProfile accepted a traversal name")
	}

	// Nothing may have been created outside or inside the profiles dir.
	if _, err := os.Stat(filepath.Join(tmpDir, ".config", "lazyproxyflare", "evil.yaml")); !os.IsNotExist(err) {
		t.Errorf("traversal wrote a file, stat err = %v", err)
	}
}

// Saved profiles hold API tokens: files must be 0600 and the directory 0700.
func TestSaveProfilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	profile := &ProfileConfig{
		Cloudflare: CloudflareConfig{
			APIToken: "test-token",
			ZoneID:   "abcdef0123456789abcdef0123456789",
		},
		Domain: "example.com",
		Proxy: ProxyConfig{
			Type:       ProxyTypeCaddy,
			Deployment: DeploymentDocker,
			Caddy: CaddyProxyConfig{
				CaddyfilePath: "/etc/caddy/Caddyfile",
				ContainerName: "caddy",
			},
		},
	}
	if err := SaveProfile("perm-test", profile); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	profilesDir := filepath.Join(tmpDir, ".config", "lazyproxyflare", "profiles")
	dirInfo, err := os.Stat(profilesDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0700 {
		t.Errorf("profiles dir perms = %o, want 0700", got)
	}
	fileInfo, err := os.Stat(filepath.Join(profilesDir, "perm-test.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0600 {
		t.Errorf("profile file perms = %o, want 0600", got)
	}
}

func TestTightenPermissions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lazyproxyflare")
	profiles := filepath.Join(dir, "profiles")
	if err := os.MkdirAll(profiles, 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]os.FileMode{
		filepath.Join(profiles, "a.yaml"):      0644,
		filepath.Join(profiles, "a.yaml.bak"):  0644,
		filepath.Join(dir, "audit.log"):        0664,
		filepath.Join(dir, "last_profile.txt"): 0644,
		filepath.Join(dir, "already-tight"):    0400,
	}
	for p, m := range files {
		if err := os.WriteFile(p, []byte("x"), m); err != nil {
			t.Fatal(err)
		}
		os.Chmod(p, m)
	}
	// A symlink must not be followed: its target keeps its mode.
	outside := filepath.Join(t.TempDir(), "outside")
	os.WriteFile(outside, []byte("x"), 0644)
	os.Chmod(outside, 0644)
	os.Symlink(outside, filepath.Join(dir, "link"))
	os.Chmod(dir, 0755)
	os.Chmod(profiles, 0755)

	TightenPermissions(dir)

	for _, d := range []string{dir, profiles} {
		if got := mustPerm(t, d); got != 0700 {
			t.Errorf("%s = %o, want 700", d, got)
		}
	}
	for p := range files {
		want := os.FileMode(0600)
		if filepath.Base(p) == "already-tight" {
			want = 0400 // loosened bits are removed, tighter modes are kept
		}
		if got := mustPerm(t, p); got != want {
			t.Errorf("%s = %o, want %o", p, got, want)
		}
	}
	if got := mustPerm(t, outside); got != 0644 {
		t.Errorf("symlink target changed to %o", got)
	}

	// Missing directory is not an error.
	TightenPermissions(filepath.Join(t.TempDir(), "nope"))
}

func mustPerm(t *testing.T, p string) os.FileMode {
	t.Helper()
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}
