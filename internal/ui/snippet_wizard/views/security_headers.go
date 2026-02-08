package views

import (
	"strings"
)

// RenderSecurityHeaders renders the security headers setup screen
func RenderSecurityHeaders(data SnippetWizardData, cursor int, previewContent string) string {
	var b strings.Builder

	// Title
	b.WriteString(StyleInfo.Render("Security Headers Snippet"))
	b.WriteString("\n\n")

	// Description
	b.WriteString("Add security headers to protect against common web vulnerabilities.\n")
	b.WriteString("\n")

	// Enable checkbox
	checkbox := RenderCheckbox(data.CreateSecurityHeaders)
	b.WriteString(checkbox + " Create Security Headers snippet " + StyleDim.Render("(press space to toggle)"))
	b.WriteString("\n\n")

	if data.CreateSecurityHeaders {
		// Preset selection
		b.WriteString(StyleInfo.Render("Security Level:"))
		b.WriteString("\n")

		presets := []struct {
			name string
			desc string
		}{
			{"basic", "Basic protection (X-Content-Type-Options, X-Frame-Options)"},
			{"strict", "Strict protection (adds CSP, Referrer-Policy, Permissions-Policy)"},
			{"paranoid", "Maximum protection (strictest CSP, all security headers)"},
		}

		for i, preset := range presets {
			selected := data.SecurityPreset == preset.name
			cursorStr := ""
			if i == cursor {
				cursorStr = StyleHighlight.Render("→ ")
			} else {
				cursorStr = "  "
			}

			radio := RenderRadio(selected)

			line := cursorStr + radio + " " + StyleKeybinding.Render(preset.name) + " - " + StyleDim.Render(preset.desc)
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString("\n")

		// Preview
		if data.SecurityPreset != "" && previewContent != "" {
			b.WriteString(StyleInfo.Render("Preview:"))
			b.WriteString("\n")
			b.WriteString(PreviewStyle.Render(previewContent))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(RenderNavigationHint("Enter: next", "Space: toggle", "↑/↓: navigate", "ESC: back", "Ctrl+Q: close"))

	return b.String()
}
