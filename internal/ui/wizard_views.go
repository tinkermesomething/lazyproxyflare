package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"lazyproxyflare/internal/config"
)

// renderWizardView renders the current wizard step
func (m Model) renderWizardView() string {
	switch m.wizardStep {
	case WizardStepWelcome:
		return m.renderWizardWelcome()
	case WizardStepBasicInfo:
		return m.renderWizardBasicInfo()
	case WizardStepCloudflare:
		return m.renderWizardCloudflare()
	case WizardStepDockerConfig:
		return m.renderWizardDockerConfig()
	case WizardStepSummary:
		return m.renderWizardSummary()
	default:
		return "Unknown wizard step"
	}
}

// renderWizardModal renders wizard content in a modal
func (m Model) renderWizardModal(title string, content string, footer string) string {
	width := m.width
	height := m.height

	// Calculate modal dimensions (70% of screen, max 80 cols)
	modalWidth := int(float64(width) * 0.7)
	if modalWidth > 80 {
		modalWidth = 80
	}
	modalHeight := int(float64(height) * 0.8)
	if modalHeight > 35 {
		modalHeight = 35
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
	if title != "" {
		b.WriteString(StyleInfo.Render(title))
		b.WriteString("\n\n")
	}

	// Content
	b.WriteString(content)

	// Footer (instructions)
	if footer != "" {
		b.WriteString("\n\n")
		b.WriteString(StyleDim.Render(footer))
	}

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(b.String()),
	)
}

// renderFieldLabel renders a field label with focus indicator
func (m Model) renderFieldLabel(label string, focused bool, hasValue bool) string {
	prefix := "  "
	if focused {
		prefix = "> "
		return StyleInfo.Render(prefix + label)
	}
	if hasValue {
		return StyleSuccess.Render(prefix + label + " ✓")
	}
	return prefix + label
}

// renderTextInput renders the text input with focus styling
func (m Model) renderTextInput(focused bool) string {
	if focused {
		return "  " + m.wizardTextInput.View()
	}
	// Show current value when not focused
	val := m.wizardTextInput.Value()
	if val == "" {
		return StyleDim.Render("  (empty)")
	}
	return "  " + val
}

// Wizard Screen: Welcome
func (m Model) renderWizardWelcome() string {
	content := `Welcome to LazyProxyFlare Setup!

This wizard will set up your profile to manage:
  • Cloudflare DNS records
  • Caddy reverse proxy configuration

You'll need:
  • Cloudflare API token (Zone.DNS Edit permission)
  • Cloudflare Zone ID (from dashboard)
  • Docker container name for Caddy`

	footer := "Press Enter to continue, ESC to exit"
	return m.renderWizardModal("Setup Wizard", content, footer)
}

// Wizard Screen: Basic Info (Profile Name + Domain)
func (m Model) renderWizardBasicInfo() string {
	var b strings.Builder

	// Show validation error if present
	if m.err != nil {
		b.WriteString(StyleError.Render("⚠ " + m.err.Error()))
		b.WriteString("\n\n")
	}

	field := m.wizardData.CurrentField

	// Profile Name
	b.WriteString(m.renderFieldLabel("Profile Name", field == FieldProfileName, m.wizardData.ProfileName != ""))
	b.WriteString("\n")
	if field == FieldProfileName {
		b.WriteString("  ")
		b.WriteString(m.wizardTextInput.View())
	} else {
		if m.wizardData.ProfileName != "" {
			b.WriteString("  " + m.wizardData.ProfileName)
		} else {
			b.WriteString(StyleDim.Render("  (not set)"))
		}
	}
	b.WriteString("\n")
	if field == FieldProfileName {
		b.WriteString(StyleDim.Render("  Name for this profile (e.g., homelab, production)"))
	}
	b.WriteString("\n\n")

	// Domain
	b.WriteString(m.renderFieldLabel("Domain", field == FieldDomain, m.wizardData.Domain != ""))
	b.WriteString("\n")
	if field == FieldDomain {
		b.WriteString("  ")
		b.WriteString(m.wizardTextInput.View())
	} else {
		if m.wizardData.Domain != "" {
			b.WriteString("  " + m.wizardData.Domain)
		} else {
			b.WriteString(StyleDim.Render("  (not set)"))
		}
	}
	b.WriteString("\n")
	if field == FieldDomain {
		b.WriteString(StyleDim.Render("  Domain to manage (e.g., example.com)"))
	}

	footer := "Tab/↓: next field  Enter: continue  ESC: cancel"
	return m.renderWizardModal("Step 1: Basic Info", b.String(), footer)
}

// Wizard Screen: Cloudflare (API Token + Zone ID)
func (m Model) renderWizardCloudflare() string {
	var b strings.Builder

	// Show validation error if present
	if m.err != nil {
		b.WriteString(StyleError.Render("⚠ " + m.err.Error()))
		b.WriteString("\n\n")
	}

	field := m.wizardData.CurrentField

	// API Token
	b.WriteString(m.renderFieldLabel("API Token", field == FieldAPIToken, m.wizardData.APIToken != ""))
	b.WriteString("\n")
	if field == FieldAPIToken {
		b.WriteString("  ")
		b.WriteString(m.wizardTextInput.View())
	} else {
		if m.wizardData.APIToken != "" {
			// Mask token
			masked := "***" + m.wizardData.APIToken[max(0, len(m.wizardData.APIToken)-4):]
			b.WriteString("  " + masked)
		} else {
			b.WriteString(StyleDim.Render("  (not set)"))
		}
	}
	b.WriteString("\n")
	if field == FieldAPIToken {
		b.WriteString(StyleDim.Render("  Create at: dash.cloudflare.com/profile/api-tokens"))
		b.WriteString("\n")
		b.WriteString(StyleDim.Render("  Required permission: Zone.DNS (Edit)"))
	}
	b.WriteString("\n\n")

	// Zone ID
	b.WriteString(m.renderFieldLabel("Zone ID", field == FieldZoneID, m.wizardData.ZoneID != ""))
	b.WriteString("\n")
	if field == FieldZoneID {
		b.WriteString("  ")
		b.WriteString(m.wizardTextInput.View())
	} else {
		if m.wizardData.ZoneID != "" {
			b.WriteString("  " + m.wizardData.ZoneID)
		} else {
			b.WriteString(StyleDim.Render("  (not set)"))
		}
	}
	b.WriteString("\n")
	if field == FieldZoneID {
		b.WriteString(StyleDim.Render("  Find in: Cloudflare Dashboard → Domain → Overview"))
	}

	footer := "Tab/↓: next field  Shift+Tab/↑: prev field  Enter: continue  ESC: back"
	return m.renderWizardModal("Step 2: Cloudflare", b.String(), footer)
}

// Wizard Screen: Caddy Config (Deployment method + related fields)
func (m Model) renderWizardDockerConfig() string {
	var b strings.Builder

	// Show validation error if present
	if m.err != nil {
		b.WriteString(StyleError.Render("⚠ " + m.err.Error()))
		b.WriteString("\n\n")
	}

	field := m.wizardData.CurrentField

	// Deployment Method (radio selection)
	b.WriteString(m.renderFieldLabel("Deployment Method", field == FieldDeploymentMethod, m.wizardData.DeploymentMethod != ""))
	b.WriteString("\n")
	options := DeploymentOptions()
	for i, opt := range options {
		prefix := "    "
		radio := "○"
		// Check if this option is selected
		isSelected := false
		if opt.Value == "system" && m.wizardData.DeploymentMethod == config.DeploymentSystem {
			isSelected = true
		} else if opt.Value != "system" && m.wizardData.DeploymentMethod == config.DeploymentDocker && m.wizardData.DockerMethod == opt.Value {
			isSelected = true
		}
		if isSelected {
			radio = "●"
		}
		if field == FieldDeploymentMethod && i == m.wizardCursor {
			prefix = "  > "
			b.WriteString(StyleInfo.Render(fmt.Sprintf("%s%s %s", prefix, radio, opt.Label)))
		} else {
			b.WriteString(fmt.Sprintf("%s%s %s", prefix, radio, opt.Label))
		}
		b.WriteString("\n")
	}

	// Show Docker-specific fields
	if m.wizardData.DeploymentMethod == config.DeploymentDocker {
		// Compose File Path (only shown for compose method)
		if m.wizardData.DockerMethod == "compose" {
			b.WriteString("\n")
			b.WriteString(m.renderFieldLabel("Compose File Path", field == FieldComposeFilePath, m.wizardData.ComposeFilePath != ""))
			b.WriteString("\n")
			if field == FieldComposeFilePath {
				b.WriteString("  ")
				b.WriteString(m.wizardTextInput.View())
				b.WriteString("\n")
				b.WriteString(StyleDim.Render("  Path to docker-compose.yml (optional if in cwd)"))
			} else {
				if m.wizardData.ComposeFilePath != "" {
					b.WriteString("  " + m.wizardData.ComposeFilePath)
				} else {
					b.WriteString(StyleDim.Render("  (using default)"))
				}
			}
			b.WriteString("\n")
		}

		// Container Name
		b.WriteString("\n")
		b.WriteString(m.renderFieldLabel("Container Name", field == FieldContainerName, m.wizardData.ContainerName != ""))
		b.WriteString("\n")
		if field == FieldContainerName {
			// Show detected containers if available
			if len(m.wizardDockerContainers) > 0 {
				for i, container := range m.wizardDockerContainers {
					prefix := "    "
					radio := "○"
					if m.wizardData.ContainerName == container.Name {
						radio = "●"
					}
					if i == m.wizardCursor {
						prefix = "  > "
						b.WriteString(StyleInfo.Render(fmt.Sprintf("%s%s %s", prefix, radio, container.Name)))
						b.WriteString(StyleDim.Render(fmt.Sprintf(" (%s)", container.Image)))
					} else {
						b.WriteString(fmt.Sprintf("%s%s %s", prefix, radio, container.Name))
						b.WriteString(StyleDim.Render(fmt.Sprintf(" (%s)", container.Image)))
					}
					b.WriteString("\n")
				}
				// Manual entry option
				manualIdx := len(m.wizardDockerContainers)
				prefix := "    "
				radio := "○"
				if m.wizardCursor == manualIdx {
					prefix = "  > "
					b.WriteString(StyleInfo.Render(fmt.Sprintf("%s%s Enter manually", prefix, radio)))
					b.WriteString("\n")
					b.WriteString("  ")
					b.WriteString(m.wizardTextInput.View())
				} else {
					b.WriteString(fmt.Sprintf("%s%s Enter manually", prefix, radio))
				}
			} else {
				b.WriteString("  ")
				b.WriteString(m.wizardTextInput.View())
				b.WriteString("\n")
				b.WriteString(StyleDim.Render("  No Caddy containers detected. Enter name manually."))
			}
		} else {
			if m.wizardData.ContainerName != "" {
				b.WriteString("  " + m.wizardData.ContainerName)
			} else {
				b.WriteString(StyleDim.Render("  (not set)"))
			}
		}
		b.WriteString("\n")

		// Caddyfile Path (host)
		b.WriteString("\n")
		b.WriteString(m.renderFieldLabel("Caddyfile Path (host)", field == FieldCaddyfilePath, m.wizardData.CaddyfilePath != ""))
		b.WriteString("\n")
		if field == FieldCaddyfilePath {
			b.WriteString("  ")
			b.WriteString(m.wizardTextInput.View())
			b.WriteString("\n")
			b.WriteString(StyleDim.Render("  Full path on host (e.g., /home/user/caddy/Caddyfile)"))
		} else {
			if m.wizardData.CaddyfilePath != "" {
				b.WriteString("  " + m.wizardData.CaddyfilePath)
			} else {
				b.WriteString(StyleDim.Render("  (not set)"))
			}
		}
		b.WriteString("\n")

		// Caddyfile Container Path
		b.WriteString("\n")
		b.WriteString(m.renderFieldLabel("Caddyfile Path (container)", field == FieldCaddyfileContainerPath, m.wizardData.CaddyfileContainerPath != ""))
		b.WriteString("\n")
		if field == FieldCaddyfileContainerPath {
			b.WriteString("  ")
			b.WriteString(m.wizardTextInput.View())
			b.WriteString("\n")
			b.WriteString(StyleDim.Render("  Path inside container (default: /etc/caddy/Caddyfile)"))
		} else {
			if m.wizardData.CaddyfileContainerPath != "" {
				b.WriteString("  " + m.wizardData.CaddyfileContainerPath)
			} else {
				b.WriteString(StyleDim.Render("  /etc/caddy/Caddyfile (default)"))
			}
		}
	} else {
		// Show System-specific fields
		// Caddyfile Path
		b.WriteString("\n")
		b.WriteString(m.renderFieldLabel("Caddyfile Path", field == FieldCaddyfilePath, m.wizardData.CaddyfilePath != ""))
		b.WriteString("\n")
		if field == FieldCaddyfilePath {
			b.WriteString("  ")
			b.WriteString(m.wizardTextInput.View())
			b.WriteString("\n")
			b.WriteString(StyleDim.Render("  Usually /etc/caddy/Caddyfile"))
		} else {
			if m.wizardData.CaddyfilePath != "" {
				b.WriteString("  " + m.wizardData.CaddyfilePath)
			} else {
				b.WriteString(StyleDim.Render("  (not set)"))
			}
		}
		b.WriteString("\n")

		// Caddy Binary Path
		b.WriteString("\n")
		b.WriteString(m.renderFieldLabel("Caddy Binary Path", field == FieldCaddyBinaryPath, m.wizardData.CaddyBinaryPath != ""))
		b.WriteString("\n")
		if field == FieldCaddyBinaryPath {
			b.WriteString("  ")
			b.WriteString(m.wizardTextInput.View())
			b.WriteString("\n")
			if m.wizardData.CaddyBinaryPath != "" {
				b.WriteString(StyleSuccess.Render("  ✓ Auto-detected"))
			} else {
				b.WriteString(StyleDim.Render("  Could not auto-detect. Enter path manually."))
			}
		} else {
			if m.wizardData.CaddyBinaryPath != "" {
				b.WriteString("  " + m.wizardData.CaddyBinaryPath)
			} else {
				b.WriteString(StyleDim.Render("  (not set)"))
			}
		}
	}

	footer := "Tab/↓: next  Shift+Tab/↑: prev  Enter: continue  ESC: back"
	return m.renderWizardModal("Step 3: Caddy Configuration", b.String(), footer)
}

// Wizard Screen: Summary
func (m Model) renderWizardSummary() string {
	var b strings.Builder

	b.WriteString("Review Your Configuration:\n\n")

	// Show validation error if present
	if m.err != nil {
		b.WriteString(StyleError.Render("Error: "+m.err.Error()) + "\n\n")
	}

	// Basic Info
	b.WriteString(StyleInfo.Render("Profile"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Name: %s\n", m.wizardData.ProfileName))
	b.WriteString(fmt.Sprintf("  Domain: %s\n", m.wizardData.Domain))
	b.WriteString("\n")

	// Cloudflare
	b.WriteString(StyleInfo.Render("Cloudflare"))
	b.WriteString("\n")
	maskedToken := "***"
	if len(m.wizardData.APIToken) > 8 {
		maskedToken = "***" + m.wizardData.APIToken[len(m.wizardData.APIToken)-8:]
	}
	b.WriteString(fmt.Sprintf("  API Token: %s\n", maskedToken))
	b.WriteString(fmt.Sprintf("  Zone ID: %s\n", m.wizardData.ZoneID))
	b.WriteString("\n")

	// Caddy Configuration
	b.WriteString(StyleInfo.Render("Caddy"))
	b.WriteString("\n")
	if m.wizardData.DeploymentMethod == config.DeploymentSystem {
		b.WriteString("  Deployment: System (systemd/snap/binary)\n")
		b.WriteString(fmt.Sprintf("  Caddy Binary: %s\n", m.wizardData.CaddyBinaryPath))
		b.WriteString(fmt.Sprintf("  Caddyfile: %s\n", m.wizardData.CaddyfilePath))
	} else {
		deployMethod := "Plain Docker"
		if m.wizardData.DockerMethod == "compose" {
			deployMethod = "Docker Compose"
		}
		b.WriteString(fmt.Sprintf("  Deployment: %s\n", deployMethod))
		if m.wizardData.DockerMethod == "compose" && m.wizardData.ComposeFilePath != "" {
			b.WriteString(fmt.Sprintf("  Compose File: %s\n", m.wizardData.ComposeFilePath))
		}
		b.WriteString(fmt.Sprintf("  Container: %s\n", m.wizardData.ContainerName))
		b.WriteString(fmt.Sprintf("  Caddyfile (host): %s\n", m.wizardData.CaddyfilePath))
		containerPath := m.wizardData.CaddyfileContainerPath
		if containerPath == "" {
			containerPath = "/etc/caddy/Caddyfile"
		}
		b.WriteString(fmt.Sprintf("  Caddyfile (container): %s\n", containerPath))
	}
	b.WriteString("\n")

	// Defaults
	b.WriteString(StyleInfo.Render("Defaults"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  CNAME Target: %s\n", m.wizardData.DefaultCNAMETarget))
	b.WriteString(fmt.Sprintf("  Port: %d\n", m.wizardData.DefaultPort))
	proxied := "No"
	if m.wizardData.DefaultProxied {
		proxied = "Yes"
	}
	b.WriteString(fmt.Sprintf("  Proxied: %s\n", proxied))

	footer := "y: Save profile  n: Cancel  b: Go back and edit"
	return m.renderWizardModal("Step 4: Review", b.String(), footer)
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
