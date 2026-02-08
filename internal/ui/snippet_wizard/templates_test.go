package snippet_wizard

import (
	"strings"
	"testing"
)

func TestGenerateIPRestrictionSnippet(t *testing.T) {
	tests := []struct {
		name       string
		lanSubnet  string
		externalIP string
		wantLAN    bool   // Should contain LAN subnet
		wantExt    bool   // Should contain external IP
		wantString string // Specific string to check for
	}{
		{
			name:       "Both LAN and external IP provided",
			lanSubnet:  "10.0.0.0/8",
			externalIP: "127.0.0.1/32",
			wantLAN:    true,
			wantExt:    true,
			wantString: "not remote_ip 10.0.0.0/8 127.0.0.1/32",
		},
		{
			name:       "Only LAN subnet provided",
			lanSubnet:  "192.168.1.0/24",
			externalIP: "",
			wantLAN:    true,
			wantExt:    false,
			wantString: "not remote_ip 192.168.1.0/24",
		},
		{
			name:       "Only external IP provided (uses LAN in fallback)",
			lanSubnet:  "",
			externalIP: "203.0.113.5/32",
			wantLAN:    false,
			wantExt:    false,
			wantString: "not remote_ip 10.0.0.0/8", // fallback to default
		},
		{
			name:       "Neither provided (uses fallback)",
			lanSubnet:  "",
			externalIP: "",
			wantLAN:    false,
			wantExt:    false,
			wantString: "not remote_ip 10.0.0.0/8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateIPRestrictionSnippet(tt.lanSubnet, tt.externalIP)

			// Check for expected string
			if !strings.Contains(result, tt.wantString) {
				t.Errorf("Expected snippet to contain %q, got:\n%s", tt.wantString, result)
			}

			// Verify LAN subnet presence
			if tt.wantLAN && tt.lanSubnet != "" {
				if !strings.Contains(result, tt.lanSubnet) {
					t.Errorf("Expected snippet to contain LAN subnet %q, got:\n%s", tt.lanSubnet, result)
				}
			}

			// Verify external IP presence
			if tt.wantExt && tt.externalIP != "" {
				if !strings.Contains(result, tt.externalIP) {
					t.Errorf("Expected snippet to contain external IP %q, got:\n%s", tt.externalIP, result)
				}
			}

			// Verify snippet structure
			if !strings.Contains(result, "(ip_restricted)") {
				t.Errorf("Expected snippet to contain snippet name, got:\n%s", result)
			}
			if !strings.Contains(result, "@external") {
				t.Errorf("Expected snippet to contain @external matcher, got:\n%s", result)
			}
			if !strings.Contains(result, "respond @external 404") {
				t.Errorf("Expected snippet to contain respond directive, got:\n%s", result)
			}
		})
	}
}

func TestGenerateIPRestrictionSnippet_UpdateBehavior(t *testing.T) {
	// Simulate user typing in LAN field first
	snippet1 := GenerateIPRestrictionSnippet("10.0.0.0/8", "")
	if !strings.Contains(snippet1, "not remote_ip 10.0.0.0/8") {
		t.Errorf("Step 1: Expected LAN only, got:\n%s", snippet1)
	}
	if strings.Contains(snippet1, "127.0.0.1") {
		t.Errorf("Step 1: Should not contain external IP yet, got:\n%s", snippet1)
	}

	// Simulate user typing in external IP field (should update)
	snippet2 := GenerateIPRestrictionSnippet("10.0.0.0/8", "127.0.0.1/32")
	if !strings.Contains(snippet2, "not remote_ip 10.0.0.0/8 127.0.0.1/32") {
		t.Errorf("Step 2: Expected both IPs, got:\n%s", snippet2)
	}

	// Verify both parameters are present
	if !strings.Contains(snippet2, "10.0.0.0/8") {
		t.Errorf("Step 2: Lost LAN subnet, got:\n%s", snippet2)
	}
	if !strings.Contains(snippet2, "127.0.0.1/32") {
		t.Errorf("Step 2: External IP not added, got:\n%s", snippet2)
	}
}
