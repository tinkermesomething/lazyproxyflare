package caddy

import (
	"testing"
)

func TestGetPrimaryDomain(t *testing.T) {
	tests := []struct {
		name     string
		entry    CaddyEntry
		expected string
	}{
		{
			name:     "multi-domain entry",
			entry:    CaddyEntry{Domains: []string{"primary.com", "secondary.com", "tertiary.com"}},
			expected: "primary.com",
		},
		{
			name:     "single domain",
			entry:    CaddyEntry{Domains: []string{"single.com"}},
			expected: "single.com",
		},
		{
			name:     "backwards compatibility",
			entry:    CaddyEntry{Domain: "legacy.com"},
			expected: "legacy.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPrimaryDomain(&tt.entry)
			if got != tt.expected {
				t.Errorf("GetPrimaryDomain() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetAdditionalDomains(t *testing.T) {
	tests := []struct {
		name     string
		entry    CaddyEntry
		expected int
	}{
		{
			name:     "multi-domain entry",
			entry:    CaddyEntry{Domains: []string{"primary.com", "secondary.com", "tertiary.com"}},
			expected: 2, // secondary and tertiary
		},
		{
			name:     "single domain",
			entry:    CaddyEntry{Domains: []string{"single.com"}},
			expected: 0,
		},
		{
			name:     "empty domains",
			entry:    CaddyEntry{Domains: []string{}},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAdditionalDomains(&tt.entry)
			if len(got) != tt.expected {
				t.Errorf("GetAdditionalDomains() count = %v, want %v", len(got), tt.expected)
			}
		})
	}
}

func TestGetDomainCount(t *testing.T) {
	tests := []struct {
		name     string
		entry    CaddyEntry
		expected int
	}{
		{
			name:     "three domains",
			entry:    CaddyEntry{Domains: []string{"a.com", "b.com", "c.com"}},
			expected: 3,
		},
		{
			name:     "single domain",
			entry:    CaddyEntry{Domains: []string{"single.com"}},
			expected: 1,
		},
		{
			name:     "backwards compatibility",
			entry:    CaddyEntry{Domain: "legacy.com"},
			expected: 1,
		},
		{
			name:     "empty",
			entry:    CaddyEntry{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDomainCount(&tt.entry)
			if got != tt.expected {
				t.Errorf("GetDomainCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatDomainDisplay(t *testing.T) {
	tests := []struct {
		name     string
		entry    CaddyEntry
		expected string
	}{
		{
			name:     "multi-domain",
			entry:    CaddyEntry{Domains: []string{"mail.example.com", "webmail.example.com", "imap.example.com"}},
			expected: "mail.example.com (+2 more)",
		},
		{
			name:     "two domains",
			entry:    CaddyEntry{Domains: []string{"app.example.com", "api.example.com"}},
			expected: "app.example.com (+1 more)",
		},
		{
			name:     "single domain",
			entry:    CaddyEntry{Domains: []string{"single.example.com"}},
			expected: "single.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDomainDisplay(&tt.entry)
			if got != tt.expected {
				t.Errorf("FormatDomainDisplay() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHasDomain(t *testing.T) {
	entry := CaddyEntry{
		Domains: []string{"mail.example.com", "webmail.example.com"},
	}

	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "exact match",
			domain:   "mail.example.com",
			expected: true,
		},
		{
			name:     "case insensitive",
			domain:   "MAIL.EXAMPLE.COM",
			expected: true,
		},
		{
			name:     "with whitespace",
			domain:   "  mail.example.com  ",
			expected: true,
		},
		{
			name:     "second domain",
			domain:   "webmail.example.com",
			expected: true,
		},
		{
			name:     "not found",
			domain:   "notfound.example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasDomain(&entry, tt.domain)
			if got != tt.expected {
				t.Errorf("HasDomain(%q) = %v, want %v", tt.domain, got, tt.expected)
			}
		})
	}
}

func TestAddDomain(t *testing.T) {
	tests := []struct {
		name        string
		entry       CaddyEntry
		addDomain   string
		expectError bool
		expectCount int
	}{
		{
			name:        "add to empty",
			entry:       CaddyEntry{},
			addDomain:   "new.example.com",
			expectError: false,
			expectCount: 1,
		},
		{
			name:        "add to existing",
			entry:       CaddyEntry{Domains: []string{"first.com"}},
			addDomain:   "second.com",
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "add duplicate",
			entry:       CaddyEntry{Domains: []string{"existing.com"}},
			addDomain:   "existing.com",
			expectError: true,
			expectCount: 1,
		},
		{
			name:        "add empty",
			entry:       CaddyEntry{Domains: []string{"existing.com"}},
			addDomain:   "",
			expectError: true,
			expectCount: 1,
		},
		{
			name:        "migrate old Domain field",
			entry:       CaddyEntry{Domain: "old.com"},
			addDomain:   "new.com",
			expectError: false,
			expectCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddDomain(&tt.entry, tt.addDomain)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			got := GetDomainCount(&tt.entry)
			if got != tt.expectCount {
				t.Errorf("domain count = %v, want %v", got, tt.expectCount)
			}
		})
	}
}

func TestRemoveDomain(t *testing.T) {
	tests := []struct {
		name         string
		entry        CaddyEntry
		removeDomain string
		expectError  bool
		expectCount  int
	}{
		{
			name:         "remove from multi-domain",
			entry:        CaddyEntry{Domains: []string{"first.com", "second.com", "third.com"}},
			removeDomain: "second.com",
			expectError:  false,
			expectCount:  2,
		},
		{
			name:         "remove last domain",
			entry:        CaddyEntry{Domains: []string{"only.com"}},
			removeDomain: "only.com",
			expectError:  true, // Cannot remove last domain
			expectCount:  1,
		},
		{
			name:         "remove non-existent",
			entry:        CaddyEntry{Domains: []string{"first.com", "second.com"}},
			removeDomain: "notfound.com",
			expectError:  true,
			expectCount:  2,
		},
		{
			name:         "case insensitive removal",
			entry:        CaddyEntry{Domains: []string{"First.com", "Second.com"}},
			removeDomain: "FIRST.COM",
			expectError:  false,
			expectCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RemoveDomain(&tt.entry, tt.removeDomain)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			got := GetDomainCount(&tt.entry)
			if got != tt.expectCount {
				t.Errorf("domain count = %v, want %v", got, tt.expectCount)
			}
		})
	}
}

func TestSetPrimaryDomain(t *testing.T) {
	tests := []struct {
		name           string
		entry          CaddyEntry
		newPrimary     string
		expectError    bool
		expectedPrimary string
	}{
		{
			name:           "set third as primary",
			entry:          CaddyEntry{Domains: []string{"first.com", "second.com", "third.com"}},
			newPrimary:     "third.com",
			expectError:    false,
			expectedPrimary: "third.com",
		},
		{
			name:           "set already primary",
			entry:          CaddyEntry{Domains: []string{"first.com", "second.com"}},
			newPrimary:     "first.com",
			expectError:    false,
			expectedPrimary: "first.com",
		},
		{
			name:           "set non-existent as primary",
			entry:          CaddyEntry{Domains: []string{"first.com", "second.com"}},
			newPrimary:     "notfound.com",
			expectError:    true,
			expectedPrimary: "first.com", // Should remain unchanged
		},
		{
			name:           "case insensitive",
			entry:          CaddyEntry{Domains: []string{"first.com", "second.com", "third.com"}},
			newPrimary:     "SECOND.COM",
			expectError:    false,
			expectedPrimary: "SECOND.COM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetPrimaryDomain(&tt.entry, tt.newPrimary)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			got := GetPrimaryDomain(&tt.entry)
			if got != tt.expectedPrimary {
				t.Errorf("primary domain = %v, want %v", got, tt.expectedPrimary)
			}
		})
	}
}

func TestMigrateToMultiDomain(t *testing.T) {
	tests := []struct {
		name           string
		entry          CaddyEntry
		expectedDomains int
		expectedPrimary string
	}{
		{
			name:           "migrate old Domain field",
			entry:          CaddyEntry{Domain: "old.example.com"},
			expectedDomains: 1,
			expectedPrimary: "old.example.com",
		},
		{
			name:           "already migrated",
			entry:          CaddyEntry{Domains: []string{"new.example.com"}},
			expectedDomains: 1,
			expectedPrimary: "new.example.com",
		},
		{
			name:           "fix missing Domain field",
			entry:          CaddyEntry{Domains: []string{"fixed.example.com"}, Domain: ""},
			expectedDomains: 1,
			expectedPrimary: "fixed.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MigrateToMultiDomain(&tt.entry)

			if len(tt.entry.Domains) != tt.expectedDomains {
				t.Errorf("Domains count = %v, want %v", len(tt.entry.Domains), tt.expectedDomains)
			}

			if tt.entry.Domain != tt.expectedPrimary {
				t.Errorf("Domain = %v, want %v", tt.entry.Domain, tt.expectedPrimary)
			}
		})
	}
}

func TestValidateDomains(t *testing.T) {
	tests := []struct {
		name        string
		entry       CaddyEntry
		expectError bool
	}{
		{
			name:        "valid multi-domain",
			entry:       CaddyEntry{Domains: []string{"first.com", "second.com"}},
			expectError: false,
		},
		{
			name:        "valid single domain",
			entry:       CaddyEntry{Domains: []string{"single.com"}},
			expectError: false,
		},
		{
			name:        "empty entry",
			entry:       CaddyEntry{},
			expectError: true, // Must have at least one domain
		},
		{
			name:        "empty domain in list",
			entry:       CaddyEntry{Domains: []string{"valid.com", ""}},
			expectError: true,
		},
		{
			name:        "duplicate domains",
			entry:       CaddyEntry{Domains: []string{"dupe.com", "other.com", "dupe.com"}},
			expectError: true,
		},
		{
			name:        "duplicate domains case insensitive",
			entry:       CaddyEntry{Domains: []string{"Example.com", "example.com"}},
			expectError: true,
		},
		{
			name:        "old format migrates and validates",
			entry:       CaddyEntry{Domain: "old.com"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDomains(&tt.entry)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
