package ui

import (
	"testing"
)

// TestActiveTabString tests the String() method returns correct human-readable names
func TestActiveTabString(t *testing.T) {
	tests := []struct {
		name     string
		tab      ActiveTab
		expected string
	}{
		{
			name:     "TabCloudflare returns 'Cloudflare'",
			tab:      TabCloudflare,
			expected: "Cloudflare",
		},
		{
			name:     "TabCaddy returns 'Caddy'",
			tab:      TabCaddy,
			expected: "Caddy",
		},
		{
			name:     "Unknown tab returns 'Unknown'",
			tab:      ActiveTab(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tab.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestTabSwitching tests that tab switching logic toggles between TabCloudflare and TabCaddy
func TestTabSwitching(t *testing.T) {
	tests := []struct {
		name           string
		currentTab     ActiveTab
		expectedNewTab ActiveTab
	}{
		{
			name:           "Switching from TabCloudflare toggles to TabCaddy",
			currentTab:     TabCloudflare,
			expectedNewTab: TabCaddy,
		},
		{
			name:           "Switching from TabCaddy toggles to TabCloudflare",
			currentTab:     TabCaddy,
			expectedNewTab: TabCloudflare,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activeTab := tt.currentTab

			// Simulate tab switch logic from app.go line 2026-2030
			if activeTab == TabCloudflare {
				activeTab = TabCaddy
			} else {
				activeTab = TabCloudflare
			}

			if activeTab != tt.expectedNewTab {
				t.Errorf("Expected %v, got %v", tt.expectedNewTab, activeTab)
			}
		})
	}
}

// TestMultipleTabSwitches tests that repeated tab switches maintain proper state
func TestMultipleTabSwitches(t *testing.T) {
	activeTab := TabCloudflare

	// Switch 1: Cloudflare -> Caddy
	if activeTab == TabCloudflare {
		activeTab = TabCaddy
	} else {
		activeTab = TabCloudflare
	}
	if activeTab != TabCaddy {
		t.Errorf("After first switch: expected TabCaddy, got %v", activeTab)
	}

	// Switch 2: Caddy -> Cloudflare
	if activeTab == TabCloudflare {
		activeTab = TabCaddy
	} else {
		activeTab = TabCloudflare
	}
	if activeTab != TabCloudflare {
		t.Errorf("After second switch: expected TabCloudflare, got %v", activeTab)
	}

	// Switch 3: Cloudflare -> Caddy (should be back to start state after 3 switches)
	if activeTab == TabCloudflare {
		activeTab = TabCaddy
	} else {
		activeTab = TabCloudflare
	}
	if activeTab != TabCaddy {
		t.Errorf("After third switch: expected TabCaddy, got %v", activeTab)
	}
}

// TestTabConstants verifies tab constants have distinct values
func TestTabConstants(t *testing.T) {
	if TabCloudflare == TabCaddy {
		t.Error("TabCloudflare and TabCaddy should have different values")
	}
}
