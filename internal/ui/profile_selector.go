package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderProfileSelectorView renders the profile selector modal
func (m Model) renderProfileSelectorView() string {
	width := m.width
	height := m.height

	// Calculate modal dimensions
	modalWidth := int(float64(width) * 0.6)
	if modalWidth > 70 {
		modalWidth = 70
	}
	modalHeight := int(float64(height) * 0.6)
	if modalHeight > 25 {
		modalHeight = 25
	}

	// Modal styling
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorBlue).
		Padding(1, 2).
		Width(modalWidth)

	// Build content
	var b strings.Builder

	// Title
	b.WriteString(StyleInfo.Render("Select Profile"))
	b.WriteString("\n\n")

	// Profile list
	if len(m.profile.Available) == 0 {
		b.WriteString(StyleDim.Render("No profiles found."))
		b.WriteString("\n\n")
	} else {
		for i, profileName := range m.profile.Available {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}

			line := fmt.Sprintf("%s%d. %s", cursor, i+1, profileName)

			// Highlight current profile
			if profileName == m.profile.CurrentName {
				line = StyleSuccess.Render(line + " (active)")
			} else if i == m.cursor {
				line = StyleInfo.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Add new profile option
	cursor := "  "
	if m.cursor == len(m.profile.Available) {
		cursor = "> "
	}
	addNewLine := fmt.Sprintf("%s+ Add new profile (run wizard)", cursor)
	if m.cursor == len(m.profile.Available) {
		addNewLine = StyleInfo.Render(addNewLine)
	}
	b.WriteString(addNewLine)

	// Instructions
	b.WriteString("\n\n")
	b.WriteString(StyleDim.Render("j/k: navigate  Enter: select  e: edit  +/n: new  ESC: cancel"))

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(b.String()),
	)
}

// renderProfileEditView renders the profile edit modal
func (m Model) renderProfileEditView() string {
	width := m.width
	height := m.height

	// Calculate modal dimensions
	modalWidth := int(float64(width) * 0.7)
	if modalWidth > 80 {
		modalWidth = 80
	}
	modalHeight := int(float64(height) * 0.8)
	if modalHeight > 30 {
		modalHeight = 30
	}

	// Modal styling
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorBlue).
		Padding(1, 2).
		Width(modalWidth)

	// Build content
	var b strings.Builder

	// Title
	b.WriteString(StyleInfo.Render("Edit Profile: " + m.profile.EditData.OriginalName))
	b.WriteString("\n\n")

	// Error display
	if m.err != nil {
		b.WriteString(StyleError.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	// Define fields
	fields := []struct {
		label string
		value string
		isBool bool
		boolVal bool
	}{
		{"Profile Name", m.profile.EditData.Name, false, false},
		{"API Token", maskToken(m.profile.EditData.APIToken), false, false},
		{"Zone ID", m.profile.EditData.ZoneID, false, false},
		{"Domain", m.profile.EditData.Domain, false, false},
		{"Caddyfile Path", m.profile.EditData.CaddyfilePath, false, false},
		{"Container Path", m.profile.EditData.ContainerPath, false, false},
		{"Container Name", m.profile.EditData.ContainerName, false, false},
		{"CNAME Target", m.profile.EditData.CNAMETarget, false, false},
		{"Default Port", m.profile.EditData.Port, false, false},
		{"", "", false, false}, // Separator
		{"Default SSL", "", true, m.profile.EditData.SSL},
		{"Default Proxied", "", true, m.profile.EditData.Proxied},
	}

	for i, field := range fields {
		// Skip separator visual but keep in count
		if field.label == "" {
			b.WriteString("\n")
			continue
		}

		cursor := "  "
		if i == m.profile.EditCursor {
			cursor = "> "
		}

		var line string
		if field.isBool {
			checkbox := "[ ]"
			if field.boolVal {
				checkbox = "[✓]"
			}
			line = fmt.Sprintf("%s%-16s %s", cursor, field.label+":", checkbox)
		} else {
			line = fmt.Sprintf("%s%-16s %s", cursor, field.label+":", field.value)
		}

		if i == m.profile.EditCursor {
			line = StyleInfo.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Instructions
	b.WriteString("\n")
	b.WriteString(StyleDim.Render("Tab/j/k: navigate  Space: toggle  Enter: save  ESC: cancel"))

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(b.String()),
	)
}

// maskToken shows only first/last 4 chars of API token
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
