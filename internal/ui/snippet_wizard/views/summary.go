package views

import (
	"fmt"
	"strings"
)

// RenderSummary renders the summary screen showing all snippets to be created
func RenderSummary(data SnippetWizardData) string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Summary"))
	b.WriteString("\n\n")

	snippetCount := len(data.GeneratedSnippets)

	if snippetCount == 0 {
		b.WriteString(StyleWarning.Render("No snippets selected."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Press ESC to go back and select snippets to create."))
		return b.String()
	}

	b.WriteString(StyleInfo.Render(fmt.Sprintf("Ready to create %d snippet(s):", snippetCount)))
	b.WriteString("\n\n")

	// List all generated snippets
	for _, gen := range data.GeneratedSnippets {
		b.WriteString(StyleKeybinding.Render("✓ " + gen.Name))
		b.WriteString("\n")
		b.WriteString(StyleDim.Render(fmt.Sprintf("  [%s] %s", gen.Category, gen.Description)))
		b.WriteString("\n\n")
	}

	b.WriteString(StyleInfo.Render("These snippets will be added to your Caddyfile."))
	b.WriteString("\n\n")

	b.WriteString(RenderNavigationHint("Enter: create snippets", "ESC: back", "Ctrl+Q: close"))

	return b.String()
}
