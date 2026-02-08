package snippet_wizard

import (
	"testing"
)

// TestCIDRValidationBlocking verifies that Next() blocks progression when CIDR values are invalid
func TestCIDRValidationBlocking(t *testing.T) {
	tests := []struct {
		name              string
		createRestriction bool
		lanSubnet         string
		externalIP        string
		shouldProgress    bool
	}{
		{
			name:              "Valid LAN subnet, no external IP",
			createRestriction: true,
			lanSubnet:         "10.0.0.0/8",
			externalIP:        "",
			shouldProgress:    true,
		},
		{
			name:              "Valid LAN subnet and external IP",
			createRestriction: true,
			lanSubnet:         "10.0.0.0/8",
			externalIP:        "127.0.0.1/32",
			shouldProgress:    true,
		},
		{
			name:              "Empty LAN subnet (required)",
			createRestriction: true,
			lanSubnet:         "",
			externalIP:        "",
			shouldProgress:    false,
		},
		{
			name:              "Invalid LAN subnet format",
			createRestriction: true,
			lanSubnet:         "not-a-cidr",
			externalIP:        "",
			shouldProgress:    false,
		},
		{
			name:              "Valid LAN but invalid external IP",
			createRestriction: true,
			lanSubnet:         "10.0.0.0/8",
			externalIP:        "invalid",
			shouldProgress:    false,
		},
		{
			name:              "IP restriction disabled",
			createRestriction: false,
			lanSubnet:         "",
			externalIP:        "",
			shouldProgress:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create wizard with initial state
			wizard := NewWizard()
			wizard.State.CurrentStep = StepIPRestriction
			wizard.State.Data.CreateIPRestriction = tt.createRestriction
			wizard.State.Data.LANSubnet = tt.lanSubnet
			wizard.State.Data.AllowedExternalIP = tt.externalIP

			// Get initial step
			initialStep := wizard.State.CurrentStep

			// Try to progress
			wizard.Next()

			// Check if wizard progressed
			progressed := wizard.State.CurrentStep != initialStep

			if progressed != tt.shouldProgress {
				t.Errorf("Expected shouldProgress=%v, but got progressed=%v (currentStep=%v, initialStep=%v)",
					tt.shouldProgress, progressed, wizard.State.CurrentStep, initialStep)
			}
		})
	}
}

// TestValidateCIDR verifies the CIDR validation function
func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		cidr  string
		valid bool
	}{
		{"10.0.0.0/8", true},
		{"192.168.1.0/24", true},
		{"127.0.0.1/32", true},
		{"172.16.0.0/12", true},
		{"", true}, // Empty is valid (optional field)
		{"not-a-cidr", false},
		{"10.0.0.0", false}, // Missing /prefix
		{"10.0.0.0/33", false}, // Invalid prefix (>32)
		{"10.0.0/24", false}, // Incomplete IP
		{"256.0.0.0/8", false}, // Invalid IP octet
	}

	for _, tt := range tests {
		t.Run(tt.cidr, func(t *testing.T) {
			result := ValidateCIDR(tt.cidr)
			if result.Valid != tt.valid {
				t.Errorf("ValidateCIDR(%q) = %v, want %v (message: %s)",
					tt.cidr, result.Valid, tt.valid, result.Message)
			}
		})
	}
}
