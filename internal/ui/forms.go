package ui

import (
	"fmt"
	"strings"

	"lazyproxyflare/internal/caddy"

	"github.com/charmbracelet/lipgloss"
)

// wrapText wraps text to fit within maxWidth, preserving existing newlines
func wrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// Wrap long lines
		for len(line) > maxWidth {
			// Find last space before maxWidth
			breakPoint := maxWidth
			for j := maxWidth - 1; j > 0; j-- {
				if line[j] == ' ' {
					breakPoint = j
					break
				}
			}

			result.WriteString(line[:breakPoint])
			result.WriteString("\n")
			line = strings.TrimLeft(line[breakPoint:], " ")
		}
		result.WriteString(line)
	}

	return result.String()
}

// formatErrorForDisplay formats an error message for display in the UI
// Wraps text and adds visual structure for multi-line errors
func formatErrorForDisplay(err error, maxWidth int) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Wrap the error text
	wrapped := wrapText(errStr, maxWidth-4) // Leave room for prefix

	// Add prefix to first line, indent subsequent lines
	lines := strings.Split(wrapped, "\n")
	var result strings.Builder
	for i, line := range lines {
		if i == 0 {
			result.WriteString("✗ Error: ")
			result.WriteString(line)
		} else {
			result.WriteString("\n  ")
			result.WriteString(line)
		}
	}

	return result.String()
}

// renderAddFormContent renders the add entry form modal content
func (m Model) renderAddFormContent() string {
	var b strings.Builder

	// Subdomain field (multi-line textarea for multiple subdomains)
	subdomain := struct {
		label       string
		value       string
		placeholder string
		focused     bool
	}{
		label:       "Subdomain(s)",
		value:       m.addForm.Subdomain,
		placeholder: "(one subdomain per line - will become subdomain." + m.config.Domain + ")",
		focused:     m.addForm.FocusedField == 0,
	}

	b.WriteString(normalStyle.Render(subdomain.label + ":"))
	b.WriteString("\n")

	inputStyle := normalStyle.Copy()
	if subdomain.focused {
		inputStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
	}

	// Multi-line display for subdomain field
	displayValue := subdomain.value
	if displayValue == "" {
		displayValue = subdomain.placeholder
	}
	if subdomain.focused {
		displayValue += "_" // Cursor
	}

	// Split into lines and render each line
	lines := strings.Split(displayValue, "\n")
	inputWidth := 60
	for _, line := range lines {
		// Pad or truncate each line
		if len(line) > inputWidth {
			line = line[:inputWidth]
		} else {
			line = line + strings.Repeat(" ", inputWidth-len(line))
		}

		b.WriteString("  ")
		b.WriteString(inputStyle.Render("[" + line + "]"))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// DNS Type selector (field 1)
	b.WriteString(normalStyle.Render("DNS Type:"))
	b.WriteString("\n  ")

	dnsTypeStyle := normalStyle.Copy()
	if m.addForm.FocusedField == 1 {
		dnsTypeStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
	}

	dnsTypeDisplay := fmt.Sprintf("< %s >", m.addForm.DNSType)
	if m.addForm.FocusedField == 1 {
		dnsTypeDisplay = fmt.Sprintf("< %s > (space to toggle)", m.addForm.DNSType)
	}
	b.WriteString(dnsTypeStyle.Render(dnsTypeDisplay))
	b.WriteString("\n\n")

	// DNS Target field (field 2) - label changes based on DNS type
	dnsTargetLabel := "DNS Target"
	dnsTargetPlaceholder := "(CNAME target domain or A record IP address)"
	if m.addForm.DNSType == "CNAME" {
		dnsTargetPlaceholder = "(target domain for CNAME)"
	} else if m.addForm.DNSType == "A" {
		dnsTargetPlaceholder = "(IP address for A record)"
	}

	dnsTarget := struct {
		label       string
		value       string
		placeholder string
		focused     bool
	}{
		label:       dnsTargetLabel,
		value:       m.addForm.DNSTarget,
		placeholder: dnsTargetPlaceholder,
		focused:     m.addForm.FocusedField == 2,
	}

	b.WriteString(normalStyle.Render(dnsTarget.label + ":"))
	b.WriteString("\n  ")

	inputStyle = normalStyle.Copy()
	if dnsTarget.focused {
		inputStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
	}

	displayValue = dnsTarget.value
	if dnsTarget.focused {
		displayValue += "_"
	}
	if displayValue == "" || displayValue == "_" {
		displayValue = dnsTarget.placeholder
	}

	if len(displayValue) > inputWidth {
		displayValue = displayValue[:inputWidth]
	} else {
		displayValue = displayValue + strings.Repeat(" ", inputWidth-len(displayValue))
	}

	b.WriteString(inputStyle.Render("[" + displayValue + "]"))
	b.WriteString("\n\n")

	// DNS Only checkbox (field 3)
	dnsOnlyStyle := normalStyle.Copy()
	if m.addForm.FocusedField == 3 {
		dnsOnlyStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
	}
	dnsOnlyCheckmark := "[ ]"
	if m.addForm.DNSOnly {
		dnsOnlyCheckmark = "[✓]"
	}
	b.WriteString(dnsOnlyStyle.Render(dnsOnlyCheckmark + " DNS Only (skip Caddy configuration)"))
	b.WriteString("\n\n")

	// Proxied checkbox (field 6) - DNS setting, shown in DNS section
	proxiedStyle := normalStyle.Copy()
	if m.addForm.FocusedField == 6 {
		proxiedStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
	}
	proxiedCheckmark := "[ ]"
	if m.addForm.Proxied {
		proxiedCheckmark = "[✓]"
	}
	b.WriteString(proxiedStyle.Render(proxiedCheckmark + " Proxy through Cloudflare (orange cloud)"))
	b.WriteString("\n\n")

	// Separator for Caddy fields
	if !m.addForm.DNSOnly {
		b.WriteString(StyleDim.Render("--- Caddy Configuration ---"))
		b.WriteString("\n\n")
	}

	// Remaining text fields (greyed out if DNS Only is checked)
	fields := []struct {
		label       string
		value       string
		placeholder string
		focused     bool
		fieldIndex  int
	}{
		{
			label:       "Reverse Proxy Target",
			value:       m.addForm.ReverseProxyTarget,
			placeholder: "(internal IP or hostname for Caddy)",
			focused:     m.addForm.FocusedField == 4,
			fieldIndex:  4,
		},
		{
			label:       "Service Port",
			value:       m.addForm.ServicePort,
			placeholder: "",
			focused:     m.addForm.FocusedField == 5,
			fieldIndex:  5,
		},
	}

	// Render text input fields (Caddy-related, greyed out if DNS Only)
	for _, field := range fields {
		// Skip rendering if DNS Only
		if m.addForm.DNSOnly {
			continue
		}

		// Label
		b.WriteString(normalStyle.Render(field.label + ":"))
		b.WriteString("\n  ")

		// Input field
		inputStyle := normalStyle.Copy()
		if field.focused {
			inputStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		displayValue := field.value
		if field.focused {
			displayValue += "_" // Cursor
		}
		if displayValue == "" || displayValue == "_" {
			displayValue = field.placeholder
		}

		// Pad to consistent width
		inputWidth := 60
		if len(displayValue) > inputWidth {
			displayValue = displayValue[:inputWidth]
		} else {
			displayValue = displayValue + strings.Repeat(" ", inputWidth-len(displayValue))
		}

		b.WriteString(inputStyle.Render("[" + displayValue + "]"))
		b.WriteString("\n\n")
	}

	// Checkboxes (Caddy-related only, Proxied is rendered in DNS section above)
	checkboxes := []struct {
		label      string
		checked    bool
		index      int
		dnsRelated bool // If false, it's Caddy-related and should be hidden when DNS Only
	}{
		{"Backend uses HTTPS (upstream service SSL/TLS)", m.addForm.SSL, 7, false},
	}

	for _, cb := range checkboxes {
		// Skip Caddy-related checkboxes if DNS Only
		if m.addForm.DNSOnly && !cb.dnsRelated {
			continue
		}

		checkmark := "[ ]"
		if cb.checked {
			checkmark = "[✓]"
		}

		cbStyle := normalStyle.Copy()
		if m.addForm.FocusedField == cb.index {
			cbStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		b.WriteString(cbStyle.Render(checkmark + " " + cb.label))
		b.WriteString("\n")
	}

	// Snippet Selection Section (only show if not DNS-only mode)
	if !m.addForm.DNSOnly {
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("--- Apply Snippets ---"))
		b.WriteString("\n\n")

		// Check if snippets exist
		if len(m.snippets) == 0 {
			// No snippets available - show notice
			b.WriteString(StyleWarning.Render("⚠ No snippets found in Caddyfile"))
			b.WriteString("\n")
			b.WriteString(StyleDim.Render("Snippets are reusable configuration blocks for features like"))
			b.WriteString("\n")
			b.WriteString(StyleDim.Render("IP restrictions, security headers, and performance optimizations."))
			b.WriteString("\n\n")
			b.WriteString(StyleInfo.Render("Press 'w' from the main screen to create snippets using the wizard."))
			b.WriteString("\n")
		} else {
			// Snippets available - show selection
			// Start field index at 8 (after the 8 fields above: 0-7)
			snippetFieldStart := 8

			for i, snippet := range m.snippets {
				snippetFieldIndex := snippetFieldStart + i

				// Check if this snippet is selected
				isSelected := false
				if m.addForm.SelectedSnippets != nil {
					isSelected = m.addForm.SelectedSnippets[snippet.Name]
				}

				checkmark := "[ ]"
				if isSelected {
					checkmark = "[✓]"
				}

				snippetStyle := normalStyle.Copy()
				if m.addForm.FocusedField == snippetFieldIndex {
					snippetStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
				}

				// Show snippet with category badge
				categoryColor := snippet.Category.ColorCode()
				categoryBadge := lipgloss.NewStyle().
					Foreground(lipgloss.Color(categoryColor)).
					Bold(true).
					Render(fmt.Sprintf("[%s]", snippet.Category.String()))

				snippetLine := fmt.Sprintf("%s %s %s - %s",
					checkmark,
					categoryBadge,
					snippet.Name,
					snippet.Description)

				// Show suggestion hint for relevant snippets
				if m.shouldSuggestSnippet(snippet.Name) {
					snippetLine += " " + StyleInfo.Render("(suggested)")
				}

				// Consistent indentation with other form fields
				b.WriteString("  ")
				b.WriteString(snippetStyle.Render(snippetLine))
				b.WriteString("\n")
			}
		}

		// Custom Caddy Config Section (after snippets)
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("--- Custom Caddy Directives (Optional) ---"))
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render("Add custom Caddy directives for one-off features:"))
		b.WriteString("\n  ")

		// Field index for custom config is 8 + len(snippets)
		customConfigFieldIndex := 8 + len(m.snippets)
		customConfigStyle := normalStyle.Copy()
		if m.addForm.FocusedField == customConfigFieldIndex {
			customConfigStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		customConfigDisplay := m.addForm.CustomCaddyConfig
		if customConfigDisplay == "" {
			customConfigDisplay = "(e.g., header X-Custom-Header \"value\")"
		}
		if m.addForm.FocusedField == customConfigFieldIndex {
			customConfigDisplay += "_" // Cursor
		}

		// Multi-line display for custom config
		customConfigLines := strings.Split(customConfigDisplay, "\n")
		for i, line := range customConfigLines {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(customConfigStyle.Render("[" + line + "]"))
			if i < len(customConfigLines)-1 {
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("Navigate: ↑↓/jk  Toggle: space  Preview: enter  Cancel: esc"))

	return b.String()
}

// shouldSuggestSnippet returns true if a snippet should be suggested based on current form state
func (m Model) shouldSuggestSnippet(snippetName string) bool {
	// Suggest security_headers snippet if SSL is enabled
	if m.addForm.SSL && snippetName == "security_headers" {
		return true
	}

	// Suggest performance snippet for all entries (it's always beneficial)
	if snippetName == "performance" {
		return true
	}

	return false
}

// renderEditFormContent renders the edit entry form modal content
func (m Model) renderEditFormContent() string {
	// Edit form is identical to add form content
	return m.renderAddFormContent()
}

// renderPreviewContent renders the preview/confirmation modal content
func (m Model) renderPreviewContent() string {
	var b strings.Builder

	// Parse subdomains and build FQDNs
	subdomains := ParseSubdomains(m.addForm.Subdomain)
	fqdns := BuildFQDNs(subdomains, m.config.Domain)

	// Parse port
	port := 80
	if m.addForm.ServicePort != "" {
		fmt.Sscanf(m.addForm.ServicePort, "%d", &port)
	}

	// Different message for create vs update
	if m.editingEntry != nil {
		b.WriteString(normalStyle.Render("Will update the entry to:"))
	} else {
		if len(fqdns) == 1 {
			b.WriteString(normalStyle.Render("Will create the following:"))
		} else {
			b.WriteString(normalStyle.Render(fmt.Sprintf("Will create %d entries:", len(fqdns))))
		}
	}
	b.WriteString("\n\n")

	// DNS Record Preview
	dnsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBlue).
		Padding(0, 1).
		Width(70)

	dnsContent := strings.Builder{}
	if len(fqdns) == 1 {
		// Single domain - show traditional format
		dnsContent.WriteString(StyleInfo.Render("Cloudflare DNS Record"))
		dnsContent.WriteString("\n")
		dnsContent.WriteString(fmt.Sprintf("  Type:     %s\n", m.addForm.DNSType))
		dnsContent.WriteString(fmt.Sprintf("  Name:     %s\n", fqdns[0]))
		dnsContent.WriteString(fmt.Sprintf("  Target:   %s\n", m.addForm.DNSTarget))
		if m.addForm.Proxied {
			dnsContent.WriteString(fmt.Sprintf("  Proxied:  Yes (Orange cloud enabled)\n"))
		} else {
			dnsContent.WriteString(fmt.Sprintf("  Proxied:  No (DNS-only)\n"))
		}
		dnsContent.WriteString(fmt.Sprintf("  TTL:      Auto\n"))
	} else {
		// Multiple domains - show list
		dnsContent.WriteString(StyleInfo.Render(fmt.Sprintf("Cloudflare DNS Records (%d)", len(fqdns))))
		dnsContent.WriteString("\n")
		dnsContent.WriteString(fmt.Sprintf("  Type:     %s\n", m.addForm.DNSType))
		dnsContent.WriteString(fmt.Sprintf("  Target:   %s\n", m.addForm.DNSTarget))
		if m.addForm.Proxied {
			dnsContent.WriteString(fmt.Sprintf("  Proxied:  Yes (Orange cloud enabled)\n"))
		} else {
			dnsContent.WriteString(fmt.Sprintf("  Proxied:  No (DNS-only)\n"))
		}
		dnsContent.WriteString(fmt.Sprintf("  TTL:      Auto\n"))
		dnsContent.WriteString("\n")
		dnsContent.WriteString("  Domains:\n")
		for i, fqdn := range fqdns {
			if i == 0 {
				dnsContent.WriteString(fmt.Sprintf("    • %s (primary)\n", fqdn))
			} else {
				dnsContent.WriteString(fmt.Sprintf("    • %s\n", fqdn))
			}
		}
	}

	b.WriteString(dnsBox.Render(dnsContent.String()))
	b.WriteString("\n\n")

	// Caddy Block Preview (only if not DNS-only mode)
	if !m.addForm.DNSOnly {
		caddyBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBlue).
			Padding(0, 1).
			Width(70)

		// Convert selected snippets map to slice
		selectedSnippets := []string{}
		if m.addForm.SelectedSnippets != nil {
			for name, selected := range m.addForm.SelectedSnippets {
				if selected {
					selectedSnippets = append(selectedSnippets, name)
				}
			}
		}

		caddyContent := strings.Builder{}
		caddyContent.WriteString(StyleInfo.Render("Caddyfile Entry"))
		caddyContent.WriteString("\n")

		// Generate Caddy block (handles both single and multi-domain)
		caddyBlock := caddy.GenerateCaddyBlock(caddy.GenerateBlockInput{
			Domains:           fqdns, // Use Domains field for both single and multi-domain
			Target:            m.addForm.ReverseProxyTarget,
			Port:              port,
			SSL:               m.addForm.SSL,
			LANOnly:           m.addForm.LANOnly,
			OAuth:             m.addForm.OAuth,
			WebSocket:         m.addForm.WebSocket,
			LANSubnet:         m.config.Defaults.LANSubnet,
			AllowedExtIP:      m.config.Defaults.AllowedExternalIP,
			AvailableSnippets: getSnippetNames(m.snippets),
			SelectedSnippets:  selectedSnippets,
			CustomCaddyConfig: m.addForm.CustomCaddyConfig,
		})
		caddyContent.WriteString(caddyBlock)

		b.WriteString(caddyBox.Render(caddyContent.String()))
		b.WriteString("\n\n")
	} else {
		// DNS-only mode message
		b.WriteString(StyleDim.Render("(DNS-only mode: Caddy configuration will be skipped)"))
		b.WriteString("\n\n")
	}

	// Status display
	if m.loading {
		if m.editingEntry != nil {
			b.WriteString(StyleInfo.Render("⟳ Updating entry..."))
		} else {
			b.WriteString(StyleInfo.Render("⟳ Creating entry..."))
		}
	} else if m.err != nil {
		// Format error with word wrapping (modal is ~70 chars wide)
		b.WriteString(StyleError.Render(formatErrorForDisplay(m.err, 66)))
	} else {
		if m.editingEntry != nil {
			b.WriteString(StyleSuccess.Render("✓ Ready to update"))
		} else {
			b.WriteString(StyleSuccess.Render("✓ Ready to create"))
		}
	}
	b.WriteString("\n\n")

	// Instructions
	if !m.loading {
		b.WriteString(StyleDim.Render("Confirm: y  Back: esc"))
	}

	return b.String()
}
