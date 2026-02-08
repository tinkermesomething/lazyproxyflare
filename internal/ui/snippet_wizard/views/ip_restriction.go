package views

import (
	"strings"
)

// SnippetWizardData is redeclared here to avoid import cycle
// This is kept in sync with the parent package's definition
type SnippetWizardData struct {
	CreateIPRestriction   bool
	LANSubnet             string
	AllowedExternalIP     string
	CreateSecurityHeaders bool
	SecurityPreset        string
	CreatePerformance     bool
	GeneratedSnippets     []GeneratedSnippet
}

type GeneratedSnippet struct {
	Name        string
	Category    string
	Content     string
	Description string
}

// RenderIPRestriction renders the IP restriction setup screen
func RenderIPRestriction(data SnippetWizardData, cursor int, previewContent string, lanValidationMsg string, externalValidationMsg string) string {
	var b strings.Builder

	// Title
	b.WriteString(StyleInfo.Render("IP Restriction Snippet"))
	b.WriteString("\n\n")

	// Description
	b.WriteString("Restrict access to your internal services to specific networks.\n")
	b.WriteString("Anyone outside these IPs will see a 404 error.\n")
	b.WriteString("\n")

	// Enable checkbox
	checkbox := RenderCheckbox(data.CreateIPRestriction)
	b.WriteString(checkbox + " Create IP Restriction snippet " + StyleDim.Render("(press space to toggle)"))
	b.WriteString("\n\n")

	if data.CreateIPRestriction {
		// LAN Subnet input
		b.WriteString(StyleInfo.Render("LAN Subnet (CIDR, required):"))
		b.WriteString("\n")

		lanValue := data.LANSubnet
		if lanValue == "" {
			lanValue = StyleDim.Render("e.g., 10.0.28.0/24")
		}

		if cursor == 0 {
			b.WriteString(StyleHighlight.Render("→ " + lanValue))
		} else {
			b.WriteString("  " + lanValue)
		}
		b.WriteString("\n")

		// Show validation error if present
		if lanValidationMsg != "" {
			b.WriteString(StyleWarning.Render("  ✗ " + lanValidationMsg))
			b.WriteString("\n")
		}
		b.WriteString("\n")

		// External IP input (optional)
		b.WriteString(StyleInfo.Render("External IP (optional):"))
		b.WriteString("\n")

		externalValue := data.AllowedExternalIP
		if externalValue == "" {
			externalValue = StyleDim.Render("e.g., 166.1.123.74/32 (leave empty to skip)")
		}

		if cursor == 1 {
			b.WriteString(StyleHighlight.Render("→ " + externalValue))
		} else {
			b.WriteString("  " + externalValue)
		}
		b.WriteString("\n")

		// Show validation error if present
		if externalValidationMsg != "" {
			b.WriteString(StyleWarning.Render("  ✗ " + externalValidationMsg))
			b.WriteString("\n")
		}
		b.WriteString("\n")

		// Preview
		b.WriteString(StyleInfo.Render("Preview:"))
		b.WriteString("\n")
		b.WriteString(PreviewStyle.Render(previewContent))
	}

	b.WriteString("\n\n")
	b.WriteString(RenderNavigationHint("Enter: next", "Space: toggle", "ESC: back", "Ctrl+Q: close"))

	return b.String()
}
