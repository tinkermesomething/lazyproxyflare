package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

// RenderCustomSnippet renders the custom snippet paste screen with textinput (name) and textarea (content)
func RenderCustomSnippet(nameInput textinput.Model, contentInput textarea.Model, cursor int) string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Custom Snippet - Paste Your Own"))
	b.WriteString("\n\n")

	// Instructions
	b.WriteString(StyleInfo.Render("Paste your Caddy snippet configuration below (Ctrl+V)"))
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("Note: Do NOT include the (name) { } wrapper - just the content inside"))
	b.WriteString("\n\n")

	// Snippet name field
	if cursor == 0 {
		b.WriteString(StyleKeybinding.Render("→ Snippet Name:"))
	} else {
		b.WriteString(StyleDim.Render("  Snippet Name:"))
	}
	b.WriteString("\n  ")
	b.WriteString(nameInput.View())
	b.WriteString("\n\n")

	// Snippet content field
	if cursor == 1 {
		b.WriteString(StyleKeybinding.Render("→ Snippet Content:"))
	} else {
		b.WriteString(StyleDim.Render("  Snippet Content:"))
	}
	b.WriteString("\n")
	// Indent each line of textarea output for consistent alignment
	textareaView := contentInput.View()
	for i, line := range strings.Split(textareaView, "\n") {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("  " + line)
	}
	b.WriteString("\n\n")

	// Preview
	snippetName := nameInput.Value()
	snippetContent := contentInput.Value()
	if snippetName != "" && snippetContent != "" {
		b.WriteString(StyleInfo.Render("Preview:"))
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("(" + snippetName + ") {"))
		b.WriteString("\n")
		for _, line := range strings.Split(snippetContent, "\n") {
			b.WriteString(StyleDim.Render("  " + line))
			b.WriteString("\n")
		}
		b.WriteString(StyleDim.Render("}"))
		b.WriteString("\n\n")
	}

	// Navigation
	b.WriteString(RenderNavigationHint("Tab: next field", "Ctrl+V: paste", "Enter: create", "ESC: back", "Ctrl+Q: close"))

	return b.String()
}
