package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"lazyproxyflare/internal/caddy"
	"lazyproxyflare/internal/diff"
)

// renderListView renders the main list view
func (m Model) renderListView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("LazyProxyFlare - %s", m.config.Domain))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Get filtered entries
	displayEntries := m.getFilteredEntries()

	// Summary stats
	synced, orphanedDNS, orphanedCaddy := 0, 0, 0
	for _, entry := range displayEntries {
		switch entry.Status {
		case diff.StatusSynced:
			synced++
		case diff.StatusOrphanedDNS:
			orphanedDNS++
		case diff.StatusOrphanedCaddy:
			orphanedCaddy++
		}
	}

	summary := fmt.Sprintf("%s %d synced  %s %d orphaned (DNS)  %s %d orphaned (Caddy)",
		syncedIconStyle.Render("✓"), synced,
		orphanedIconStyle.Render("⚠"), orphanedDNS,
		orphanedIconStyle.Render("⚠"), orphanedCaddy,
	)
	b.WriteString(summary)
	b.WriteString("\n")

	// Show filter and sort info
	filterParts := []string{}
	if m.statusFilter != FilterAll {
		filterParts = append(filterParts, fmt.Sprintf("Status: %s", m.statusFilter.String()))
	}
	if m.dnsTypeFilter != DNSTypeAll {
		filterParts = append(filterParts, fmt.Sprintf("DNS Type: %s", m.dnsTypeFilter.String()))
	}
	if m.searchQuery != "" {
		filterParts = append(filterParts, fmt.Sprintf("Search: \"%s\"", m.searchQuery))
	}

	// Build info line
	infoLine := ""
	if len(filterParts) > 0 {
		infoLine = fmt.Sprintf("Filter: %s | ", strings.Join(filterParts, ", "))
	}
	infoLine += fmt.Sprintf("Sort: %s | Showing %d of %d entries",
		m.sortMode.String(), len(displayEntries), len(m.entries))
	if len(m.selectedEntries) > 0 {
		infoLine += fmt.Sprintf(" | Selected: %d", len(m.selectedEntries))
	}
	b.WriteString(dimStyle.Render(infoLine))
	b.WriteString("\n")

	// Calculate visible range
	visibleHeight := m.height - 10 // Leave space for title, summary, status bar
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	start := m.scrollOffset
	end := start + visibleHeight
	if end > len(displayEntries) {
		end = len(displayEntries)
	}

	// Entry list
	for i := start; i < end; i++ {
		entry := displayEntries[i]

		// Selection checkbox
		checkbox := "[ ]"
		if m.selectedEntries[entry.Domain] {
			checkbox = "[✓]"
		}

		// Icon based on status
		icon := entry.Status.Icon()
		if entry.Status == diff.StatusSynced {
			icon = syncedIconStyle.Render(icon)
		} else {
			icon = orphanedIconStyle.Render(icon)
		}

		// Domain name
		domain := entry.Domain

		// Details based on what exists
		var details string
		if entry.DNS != nil && entry.Caddy != nil {
			// Both exist - show DNS type and target, plus Caddy target
			details = fmt.Sprintf("DNS:[%s]%s → Caddy:%s:%d",
				entry.DNS.Type, entry.DNS.Content, entry.Caddy.Target, entry.Caddy.Port)
		} else if entry.DNS != nil {
			// Only DNS - show type and target
			details = fmt.Sprintf("DNS:[%s]%s (no Caddy)", entry.DNS.Type, entry.DNS.Content)
		} else if entry.Caddy != nil {
			// Only Caddy
			details = fmt.Sprintf("Caddy:%s:%d (no DNS)", entry.Caddy.Target, entry.Caddy.Port)
		}

		// Render line
		line := fmt.Sprintf("%s %s %-40s %s", checkbox, icon, domain, details)

		// Apply cursor style
		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else if entry.Status != diff.StatusSynced {
			line = orphanedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Show scroll indicator if needed
	if len(displayEntries) > visibleHeight {
		scrollInfo := fmt.Sprintf("\n[Showing %d-%d of %d]", start+1, end, len(displayEntries))
		b.WriteString(normalStyle.Render(scrollInfo))
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	var statusBar string
	if m.loading {
		// Show loading indicator
		statusBar = statusBarStyle.Render("⟳ Refreshing data from Cloudflare and Caddyfile...")
	} else if m.searching {
		// Show search prompt
		statusBar = statusBarStyle.Render(
			fmt.Sprintf("Search: %s_  (enter:accept  esc:cancel)", m.searchQuery),
		)
	} else if m.err != nil {
		// Show error
		statusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Padding(0, 1).
			Render(fmt.Sprintf("Error: %v (press r to retry)", m.err))
	} else {
		// Normal status bar - show batch operations if selections exist
		if len(m.selectedEntries) > 0 {
			statusBar = statusBarStyle.Render(
				"j/k:navigate  space:select  X:batch-delete  S:batch-sync  f:filter  t:type  o:sort  /:search  b:backups  r:refresh  ?:help  q:quit",
			)
		} else {
			statusBar = statusBarStyle.Render(
				"j/k:navigate  space:select  enter:details  a:add  d:delete  s:sync  f:filter  t:type  o:sort  /:search  b:backups  r:refresh  ?:help  q:quit",
			)
		}
	}
	b.WriteString(statusBar)

	return b.String()
}

// renderDetailsView renders detailed view of selected entry
func (m Model) renderDetailsView() string {
	if m.cursor >= len(m.entries) {
		return "No entry selected"
	}

	entry := m.entries[m.cursor]
	var b strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("Details - %s", entry.Domain))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Status
	statusLine := fmt.Sprintf("Status: %s %s\n\n",
		entry.Status.Icon(),
		entry.Status.String(),
	)
	if entry.Status == diff.StatusSynced {
		b.WriteString(syncedIconStyle.Render(statusLine))
	} else {
		b.WriteString(orphanedIconStyle.Render(statusLine))
	}

	// DNS Information
	if entry.DNS != nil {
		b.WriteString(titleStyle.Render("DNS Record (Cloudflare)"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Type:     %s\n", entry.DNS.Type))
		b.WriteString(fmt.Sprintf("  Name:     %s\n", entry.DNS.Name))
		b.WriteString(fmt.Sprintf("  Content:  %s\n", entry.DNS.Content))
		b.WriteString(fmt.Sprintf("  Proxied:  %v", entry.DNS.Proxied))
		if entry.DNS.Proxied {
			b.WriteString(" (Orange cloud enabled)")
		}
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  TTL:      %d", entry.DNS.TTL))
		if entry.DNS.TTL == 1 {
			b.WriteString(" (Auto)")
		}
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Zone ID:  %s\n", entry.DNS.ZoneID))
		b.WriteString(fmt.Sprintf("  Record ID: %s\n", entry.DNS.ID))
		b.WriteString("\n")
	} else {
		b.WriteString(orphanedIconStyle.Render("DNS Record: Not found in Cloudflare"))
		b.WriteString("\n\n")
	}

	// Caddy Information
	if entry.Caddy != nil {
		b.WriteString(titleStyle.Render("Caddy Configuration"))
		b.WriteString("\n")

		// Primary info
		b.WriteString(fmt.Sprintf("  Domain:   %s\n", entry.Caddy.Domain))
		if len(entry.Caddy.Domains) > 1 {
			b.WriteString(fmt.Sprintf("  Aliases:  %v\n", entry.Caddy.Domains[1:]))
		}
		b.WriteString(fmt.Sprintf("  Target:   %s\n", entry.Caddy.Target))
		b.WriteString(fmt.Sprintf("  Port:     %d\n", entry.Caddy.Port))
		b.WriteString(fmt.Sprintf("  SSL:      %v", entry.Caddy.SSL))
		if entry.Caddy.SSL {
			b.WriteString(" (HTTPS)")
		} else {
			b.WriteString(" (HTTP)")
		}
		b.WriteString("\n")

		// Features
		if entry.Caddy.IPRestricted || entry.Caddy.OAuthHeaders || entry.Caddy.WebSocket {
			b.WriteString("  Features: ")
			features := []string{}
			if entry.Caddy.IPRestricted {
				features = append(features, "IP Restricted")
			}
			if entry.Caddy.OAuthHeaders {
				features = append(features, "OAuth Headers")
			}
			if entry.Caddy.WebSocket {
				features = append(features, "WebSocket")
			}
			b.WriteString(strings.Join(features, ", "))
			b.WriteString("\n")
		}

		// Imports
		if len(entry.Caddy.Imports) > 0 {
			b.WriteString(fmt.Sprintf("  Imports:  %v\n", entry.Caddy.Imports))
		}

		// Location in file
		b.WriteString(fmt.Sprintf("  Location: Lines %d-%d", entry.Caddy.LineStart, entry.Caddy.LineEnd))
		if entry.Caddy.HasMarker {
			b.WriteString(" (has marker)")
		}
		b.WriteString("\n\n")

		// Raw block
		b.WriteString(titleStyle.Render("Raw Caddyfile Block"))
		b.WriteString("\n")
		b.WriteString(normalStyle.Render(entry.Caddy.RawBlock))
		b.WriteString("\n")
	} else {
		b.WriteString(orphanedIconStyle.Render("Caddy Configuration: Not found in Caddyfile"))
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	statusBar := statusBarStyle.Render("j/k:next/prev  e:edit  d:delete  s:sync  esc:back  q:quit")
	b.WriteString(statusBar)

	return b.String()
}

// renderHelpView renders the help screen
// renderHelpModalContent returns help content for modal display (two-column layout)

// renderAddForm renders the add entry form
func (m Model) renderAddForm() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Add New Entry")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Field labels and values
	fields := []struct {
		label       string
		value       string
		placeholder string
		focused     bool
	}{
		{
			label:       "Subdomain",
			value:       m.addForm.Subdomain,
			placeholder: "(without domain - will become subdomain." + m.config.Domain + ")",
			focused:     m.addForm.FocusedField == 0,
		},
		{
			label:       "CNAME Target",
			value:       m.addForm.DNSTarget,
			placeholder: "(DNS record target)",
			focused:     m.addForm.FocusedField == 1,
		},
		{
			label:       "Reverse Proxy Target",
			value:       m.addForm.ReverseProxyTarget,
			placeholder: "(internal IP or hostname for Caddy)",
			focused:     m.addForm.FocusedField == 4,
		},
		{
			label:       "Service Port",
			value:       m.addForm.ServicePort,
			placeholder: "",
			focused:     m.addForm.FocusedField == 5,
		},
	}

	// Render text input fields
	for _, field := range fields {
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

	// Checkboxes
	checkboxes := []struct {
		label   string
		checked bool
		index   int
	}{
		{"Proxy through Cloudflare (orange cloud)", m.addForm.Proxied, 4},
		{"LAN only (404 for external traffic)", m.addForm.LANOnly, 5},
		{"Enable SSL/TLS (https://)", m.addForm.SSL, 6},
		{"Include OAuth/OIDC headers", m.addForm.OAuth, 7},
		{"WebSocket support", m.addForm.WebSocket, 8},
	}

	for _, cb := range checkboxes {
		cbStyle := normalStyle.Copy()
		if m.addForm.FocusedField == cb.index {
			cbStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		checkbox := "[ ]"
		if cb.checked {
			checkbox = "[✓]"
		}

		line := fmt.Sprintf("  %s %s", checkbox, cb.label)
		b.WriteString(cbStyle.Render(line))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	statusBar := statusBarStyle.Render("↑/↓ or j/k:navigate  space:toggle  enter:preview  esc:cancel")
	b.WriteString(statusBar)

	return b.String()
}

// renderEditForm renders the edit entry form
func (m Model) renderEditForm() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Edit Entry")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Field labels and values
	fields := []struct {
		label       string
		value       string
		placeholder string
		focused     bool
	}{
		{
			label:       "Subdomain",
			value:       m.addForm.Subdomain,
			placeholder: "(without domain - will become subdomain." + m.config.Domain + ")",
			focused:     m.addForm.FocusedField == 0,
		},
		{
			label:       "CNAME Target",
			value:       m.addForm.DNSTarget,
			placeholder: "(DNS record target)",
			focused:     m.addForm.FocusedField == 1,
		},
		{
			label:       "Reverse Proxy Target",
			value:       m.addForm.ReverseProxyTarget,
			placeholder: "(internal IP or hostname for Caddy)",
			focused:     m.addForm.FocusedField == 4,
		},
		{
			label:       "Service Port",
			value:       m.addForm.ServicePort,
			placeholder: "",
			focused:     m.addForm.FocusedField == 5,
		},
	}

	// Render text input fields
	for _, field := range fields {
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

	// Checkboxes
	checkboxes := []struct {
		label   string
		checked bool
		index   int
	}{
		{"Proxy through Cloudflare (orange cloud)", m.addForm.Proxied, 4},
		{"LAN only (404 for external traffic)", m.addForm.LANOnly, 5},
		{"Enable SSL/TLS (https://)", m.addForm.SSL, 6},
		{"Include OAuth/OIDC headers", m.addForm.OAuth, 7},
		{"WebSocket support", m.addForm.WebSocket, 8},
	}

	for _, cb := range checkboxes {
		cbStyle := normalStyle.Copy()
		if m.addForm.FocusedField == cb.index {
			cbStyle = selectedStyle.Copy().Reverse(false).Bold(true).Foreground(lipgloss.Color("#00D7FF"))
		}

		checkbox := "[ ]"
		if cb.checked {
			checkbox = "[✓]"
		}

		line := fmt.Sprintf("  %s %s", checkbox, cb.label)
		b.WriteString(cbStyle.Render(line))
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	statusBar := statusBarStyle.Render("↑/↓ or j/k:navigate  space:toggle  enter:preview  esc:cancel")
	b.WriteString(statusBar)

	return b.String()
}

// renderPreviewScreen renders the confirmation/preview screen
func (m Model) renderPreviewScreen() string {
	var b strings.Builder

	// Title - different for create vs update
	var title string
	if m.editingEntry != nil {
		title = titleStyle.Render("Confirm Update")
	} else {
		title = titleStyle.Render("Confirm Create")
	}
	b.WriteString(title)
	b.WriteString("\n\n")

	// Build FQDN
	fqdn := m.addForm.Subdomain + "." + m.config.Domain

	// Parse port
	port := 80
	if m.addForm.ServicePort != "" {
		fmt.Sscanf(m.addForm.ServicePort, "%d", &port)
	}

	// Different message for create vs update
	if m.editingEntry != nil {
		b.WriteString(normalStyle.Render("Will update the entry to:"))
	} else {
		b.WriteString(normalStyle.Render("Will create the following:"))
	}
	b.WriteString("\n\n")

	// DNS Record Preview
	dnsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D7FF")).
		Padding(0, 1).
		Width(70)

	dnsContent := strings.Builder{}
	dnsContent.WriteString(titleStyle.Render("Cloudflare DNS Record"))
	dnsContent.WriteString("\n")
	dnsContent.WriteString(fmt.Sprintf("  Type:     CNAME\n"))
	dnsContent.WriteString(fmt.Sprintf("  Name:     %s\n", fqdn))
	dnsContent.WriteString(fmt.Sprintf("  Target:   %s\n", m.addForm.DNSTarget))
	if m.addForm.Proxied {
		dnsContent.WriteString(fmt.Sprintf("  Proxied:  Yes (Orange cloud enabled)\n"))
	} else {
		dnsContent.WriteString(fmt.Sprintf("  Proxied:  No (DNS-only)\n"))
	}
	dnsContent.WriteString(fmt.Sprintf("  TTL:      Auto\n"))

	b.WriteString(dnsBox.Render(dnsContent.String()))
	b.WriteString("\n\n")

	// Caddy Block Preview
	caddyBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D7FF")).
		Padding(0, 1).
		Width(70)

	// Generate the actual Caddy block
	caddyBlock := caddy.GenerateCaddyBlock(caddy.GenerateBlockInput{
		FQDN:              fqdn,
		Target:            m.addForm.ReverseProxyTarget,
		Port:              port,
		SSL:               m.addForm.SSL,
		LANOnly:           m.addForm.LANOnly,
		OAuth:             m.addForm.OAuth,
		WebSocket:         m.addForm.WebSocket,
		LANSubnet:         m.config.Defaults.LANSubnet,
		AllowedExtIP:      m.config.Defaults.AllowedExternalIP,
		SelectedSnippets:  getSelectedSnippetNames(m.addForm.SelectedSnippets),
		CustomCaddyConfig: m.addForm.CustomCaddyConfig,
	})

	caddyContent := strings.Builder{}
	caddyContent.WriteString(titleStyle.Render("Caddyfile Entry"))
	caddyContent.WriteString("\n")
	caddyContent.WriteString(caddyBlock)

	b.WriteString(caddyBox.Render(caddyContent.String()))
	b.WriteString("\n\n")

	// Status/Error display
	b.WriteString("\n")
	if m.loading {
		if m.editingEntry != nil {
			b.WriteString(statusBarStyle.Render("  ⟳ Updating entry..."))
		} else {
			b.WriteString(statusBarStyle.Render("  ⟳ Creating entry..."))
		}
	} else if m.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		b.WriteString(errorStyle.Render(fmt.Sprintf("  ✗ Error: %v", m.err)))
	} else {
		if m.editingEntry != nil {
			b.WriteString(syncedIconStyle.Render("  ✓ Ready to update"))
		} else {
			b.WriteString(syncedIconStyle.Render("  ✓ Ready to create"))
		}
	}
	b.WriteString("\n\n")

	// Status bar
	var statusBar string
	if m.loading {
		statusBar = statusBarStyle.Render("Please wait...")
	} else {
		if m.editingEntry != nil {
			statusBar = statusBarStyle.Render("y:confirm and update  esc:back to edit")
		} else {
			statusBar = statusBarStyle.Render("y:confirm and create  esc:back to edit")
		}
	}
	b.WriteString(statusBar)

	return b.String()
}

