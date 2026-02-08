package views

import (
	"strings"
)

// RenderPerformance renders the performance optimization setup screen
func RenderPerformance(data SnippetWizardData, previewContent string) string {
	var b strings.Builder

	// Title
	b.WriteString(StyleInfo.Render("Performance Optimization Snippet"))
	b.WriteString("\n\n")

	// Description
	b.WriteString("Enable compression to reduce bandwidth and improve load times.\n")
	b.WriteString("\n")

	// Enable checkbox
	checkbox := RenderCheckbox(data.CreatePerformance)
	b.WriteString(checkbox + " Create Performance snippet " + StyleDim.Render("(press space to toggle)"))
	b.WriteString("\n\n")

	if data.CreatePerformance {
		b.WriteString(StyleInfo.Render("Features:"))
		b.WriteString("\n")
		b.WriteString("  • Gzip compression\n")
		b.WriteString("  • Zstd compression\n")
		b.WriteString("\n\n")

		// Preview
		b.WriteString(StyleInfo.Render("Preview:"))
		b.WriteString("\n")
		b.WriteString(PreviewStyle.Render(previewContent))
	}

	b.WriteString("\n\n")
	b.WriteString(RenderNavigationHint("Enter: next", "Space: toggle", "ESC: back", "Ctrl+Q: close"))

	return b.String()
}
