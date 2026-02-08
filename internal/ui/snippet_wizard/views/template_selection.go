package views

import (
	"fmt"
	"strings"
)

// TemplateInfo represents a template option
type TemplateInfo struct {
	Name        string
	Description string
	Category    string
	Selected    bool
	InUse       bool
}

// RenderTemplateSelection renders the template selection screen with viewport scrolling
func RenderTemplateSelection(templates map[string]TemplateInfo, cursor int, maxHeight int) string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Select Snippets to Create"))
	b.WriteString("\n\n")

	// Instructions
	b.WriteString(StyleInfo.Render("Choose which snippets you want to configure:"))
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("Use Space to toggle selection, ↑/↓ to navigate, Enter to continue"))
	b.WriteString("\n\n")

	// Group templates by category
	categories := []struct {
		name      string
		templates []string
	}{
		{
			name: "Security",
			templates: []string{
				"cors_headers",
				"rate_limiting",
				"auth_headers",
				"ip_restricted",
				"security_headers",
			},
		},
		{
			name: "Performance",
			templates: []string{
				"static_caching",
				"compression_advanced",
				"performance",
			},
		},
		{
			name: "Backend Integration",
			templates: []string{
				"websocket_advanced",
				"extended_timeouts",
				"https_backend",
			},
		},
		{
			name: "Content Control",
			templates: []string{
				"large_uploads",
				"custom_headers_inject",
				"frame_embedding",
				"rewrite_rules",
			},
		},
	}

	// Build flat list of templates with category headers for viewport
	type ViewItem struct {
		IsHeader     bool
		CategoryName string
		TemplateKey  string
		Template     TemplateInfo
		Index        int
	}

	var items []ViewItem
	currentIndex := 0

	for _, cat := range categories {
		// Add category header
		items = append(items, ViewItem{
			IsHeader:     true,
			CategoryName: cat.name,
		})

		// Add templates
		for _, templateKey := range cat.templates {
			template, exists := templates[templateKey]
			if !exists {
				continue
			}
			items = append(items, ViewItem{
				IsHeader:    false,
				TemplateKey: templateKey,
				Template:    template,
				Index:       currentIndex,
			})
			currentIndex++
		}
	}

	// Calculate viewport: show items around cursor
	// Each template takes 3 lines (checkbox+name, description, blank)
	// Each category header takes 2 lines (header, blank)
	linesPerTemplate := 3
	linesPerHeader := 2
	headerLines := 4 // Title + instructions

	// Reserve space for footer (selection count + navigation)
	footerLines := 4
	availableLines := maxHeight - headerLines - footerLines

	// Find the item containing the cursor
	var visibleItems []ViewItem
	cursorItemIdx := -1

	// Find which item the cursor is on
	for i, item := range items {
		if !item.IsHeader && item.Index == cursor {
			cursorItemIdx = i
			break
		}
	}

	if cursorItemIdx >= 0 {
		// Calculate viewport around cursor
		startIdx := 0
		endIdx := len(items)

		// Try to center cursor in viewport
		linesUsed := 0
		itemsBeforeCursor := 0
		itemsAfterCursor := 0

		// Count lines needed for items before cursor
		for i := cursorItemIdx - 1; i >= 0; i-- {
			lineCount := linesPerHeader
			if !items[i].IsHeader {
				lineCount = linesPerTemplate
			}
			if linesUsed+lineCount > availableLines/2 {
				break
			}
			linesUsed += lineCount
			itemsBeforeCursor++
		}
		startIdx = cursorItemIdx - itemsBeforeCursor

		// Count lines for items after cursor (including cursor item)
		linesUsed = 0
		for i := cursorItemIdx; i < len(items); i++ {
			lineCount := linesPerHeader
			if !items[i].IsHeader {
				lineCount = linesPerTemplate
			}
			if linesUsed+lineCount > availableLines {
				break
			}
			linesUsed += lineCount
			itemsAfterCursor++
		}
		endIdx = cursorItemIdx + itemsAfterCursor

		visibleItems = items[startIdx:endIdx]
	} else {
		// Show from beginning
		visibleItems = items
	}

	// Render visible items
	for _, item := range visibleItems {
		if item.IsHeader {
			b.WriteString(StyleKeybinding.Render("━━ " + item.CategoryName + " ━━"))
			b.WriteString("\n\n")
		} else {
			// Cursor indicator and checkbox
			if item.Index == cursor {
				b.WriteString(StyleKeybinding.Render("→ "))
			} else {
				b.WriteString("  ")
			}

			if item.Template.Selected {
				b.WriteString(StyleKeybinding.Render("[✓] "))
			} else {
				b.WriteString(StyleDim.Render("[ ] "))
			}

			// Template name
			templateName := item.Template.Name
			if item.Template.InUse {
				templateName += " (in use)"
			}
			if item.Index == cursor {
				b.WriteString(StyleKeybinding.Render(templateName))
			} else {
				b.WriteString(templateName)
			}
			b.WriteString("\n")

			// Description
			if item.Index == cursor {
				b.WriteString("    " + StyleInfo.Render(item.Template.Description))
			} else {
				b.WriteString("    " + StyleDim.Render(item.Template.Description))
			}
			b.WriteString("\n\n")
		}
	}

	// Selection count and scroll indicator
	selectedCount := 0
	totalTemplates := 0
	for _, t := range templates {
		totalTemplates++
		if t.Selected {
			selectedCount++
		}
	}

	// Show scroll indicator if not all items visible
	scrollIndicator := ""
	if cursorItemIdx >= 0 {
		totalItems := len(items)
		firstVisibleIdx := 0
		lastVisibleIdx := len(items) - 1

		// Find actual visible range
		for i, item := range items {
			if len(visibleItems) > 0 {
				if !item.IsHeader && item.Index == visibleItems[0].Index {
					firstVisibleIdx = i
				}
				if !item.IsHeader && item.Index == visibleItems[len(visibleItems)-1].Index {
					lastVisibleIdx = i
				}
			}
		}

		if firstVisibleIdx > 0 || lastVisibleIdx < totalItems-1 {
			scrollIndicator = StyleDim.Render(fmt.Sprintf(" (scroll: showing %d-%d)", cursor+1, totalTemplates))
		}
	}

	b.WriteString(StyleInfo.Render(fmt.Sprintf("Selected: %d/%d snippets%s", selectedCount, totalTemplates, scrollIndicator)))
	b.WriteString("\n\n")

	// Navigation
	b.WriteString(RenderNavigationHint("Space: toggle", "Enter: continue", "ESC: back", "Ctrl+Q: close"))

	return b.String()
}
