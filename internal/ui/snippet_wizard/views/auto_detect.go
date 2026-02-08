package views

import (
	"fmt"
	"strings"
)

// DetectedPattern represents a pattern found in Caddyfile (local copy to avoid import cycle)
type DetectedPattern struct {
	Type          int
	Count         int
	ExampleDomain string
	RawDirective  string
	SuggestedName string
	Parameters    map[string]string
}

// RenderAutoDetect renders the auto-detection results screen
func RenderAutoDetect(patterns []DetectedPattern, selectedPatterns map[string]bool, cursor int) string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Auto-Detection Results"))
	b.WriteString("\n\n")

	if len(patterns) == 0 {
		b.WriteString(StyleInfo.Render("No unique patterns detected in your Caddyfile."))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Your existing snippets cover all common patterns."))
		b.WriteString("\n\n")
		b.WriteString(RenderNavigationHint("Enter: continue", "ESC: back", "Ctrl+Q: close"))
		return b.String()
	}

	// Explanation
	b.WriteString(StyleInfo.Render(fmt.Sprintf("Found %d unique pattern(s) in your Caddyfile:", len(patterns))))
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("Select patterns to extract as reusable snippets"))
	b.WriteString("\n\n")

	// List detected patterns
	for i, pattern := range patterns {
		// Cursor
		cursorStr := "  "
		if i == cursor {
			cursorStr = StyleHighlight.Render("→ ")
		}

		// Checkbox
		isSelected := selectedPatterns[pattern.SuggestedName]
		checkbox := RenderCheckbox(isSelected)

		// Pattern info
		patternLine := cursorStr + checkbox + " " + StyleKeybinding.Render(pattern.SuggestedName)

		// Add count if > 1
		if pattern.Count > 1 {
			patternLine += StyleDim.Render(fmt.Sprintf(" (used in %d domains)", pattern.Count))
		}

		b.WriteString(patternLine)
		b.WriteString("\n")

		// Details (indented)
		if pattern.ExampleDomain != "unknown" {
			b.WriteString(StyleDim.Render(fmt.Sprintf("    Example: %s", pattern.ExampleDomain)))
			b.WriteString("\n")
		}

		// Show key parameters
		if len(pattern.Parameters) > 0 {
			paramStrs := []string{}
			for key, value := range pattern.Parameters {
				paramStrs = append(paramStrs, fmt.Sprintf("%s=%s", key, value))
			}
			b.WriteString(StyleDim.Render(fmt.Sprintf("    Config: %s", strings.Join(paramStrs, ", "))))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	// Navigation
	b.WriteString(RenderNavigationHint("Space: toggle", "↑/↓: navigate", "Enter: continue", "ESC: back", "Ctrl+Q: close"))

	return b.String()
}
