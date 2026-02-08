package views

import (
	"strings"
)

// RenderWelcome renders the welcome screen for snippet wizard with mode selection
func RenderWelcome(selectedMode int) string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Snippet Creation Wizard"))
	b.WriteString("\n\n")

	// Explanation
	b.WriteString(StyleInfo.Render("What are Snippets?"))
	b.WriteString("\n")
	b.WriteString("Snippets are reusable configuration blocks in Caddy that help you:\n")
	b.WriteString("  • Keep your Caddyfile DRY (Don't Repeat Yourself)\n")
	b.WriteString("  • Maintain consistent configuration across entries\n")
	b.WriteString("  • Easily update shared settings in one place\n")
	b.WriteString("\n\n")

	// Mode selection
	b.WriteString(StyleInfo.Render("Choose Creation Mode:"))
	b.WriteString("\n\n")

	modes := []struct {
		name        string
		description string
	}{
		{
			name:        "Templated - Advanced",
			description: "Select from 15+ templates and configure parameters",
		},
		{
			name:        "Custom - Paste Your Own",
			description: "Paste pre-written Caddy snippet configuration",
		},
		{
			name:        "Guided - Step by Step",
			description: "Answer questions to build common snippets",
		},
	}

	for i, mode := range modes {
		if i == selectedMode {
			b.WriteString(StyleKeybinding.Render("→ " + mode.name))
		} else {
			b.WriteString(StyleDim.Render("  " + mode.name))
		}
		b.WriteString("\n")
		if i == selectedMode {
			b.WriteString("  " + StyleInfo.Render(mode.description))
		} else {
			b.WriteString("  " + StyleDim.Render(mode.description))
		}
		b.WriteString("\n\n")
	}

	// Navigation
	b.WriteString(RenderNavigationHint("↑/↓: select mode", "Enter: continue", "ESC: cancel"))

	return b.String()
}
